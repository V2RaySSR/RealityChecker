package core

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"RealityChecker/internal/types"
)

// Coordinator 任务协调器
type Coordinator struct {
	pipeline        *Pipeline
	config          *types.Config
	workerPool      *WorkerPool
	concurrencyCtrl *ConcurrencyController
	stats           *CoordinatorStats
	mu              sync.RWMutex
	running         bool
}

// NewCoordinator 创建新的任务协调器
func NewCoordinator(pipeline *Pipeline, config *types.Config) *Coordinator {
	return &Coordinator{
		pipeline:        pipeline,
		config:          config,
		concurrencyCtrl: NewConcurrencyController(config),
		stats:           &CoordinatorStats{},
	}
}

// Start 启动协调器
func (c *Coordinator) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("协调器已在运行")
	}

	// 创建Worker Pool
	maxWorkers := int(c.config.Concurrency.MaxConcurrent)
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU() * 2
	}

	c.workerPool = NewWorkerPool(maxWorkers, c.pipeline)
	c.workerPool.Start()

	c.running = true
	return nil
}

// Stop 停止协调器
func (c *Coordinator) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	if c.workerPool != nil {
		c.workerPool.Stop()
	}

	c.running = false
	return nil
}

// CheckDomain 检测单个域名
func (c *Coordinator) CheckDomain(ctx context.Context, domain string) (*types.DetectionResult, error) {
	if !c.running {
		return nil, fmt.Errorf("协调器未运行")
	}

	startTime := time.Now()
	result, err := c.pipeline.Execute(ctx, domain)
	duration := time.Since(startTime)

	// 记录统计信息
	c.recordStats(duration, err == nil)

	// 记录性能指标
	if err == nil {
		c.concurrencyCtrl.RecordSuccess(duration)
	} else {
		c.concurrencyCtrl.RecordFailure(duration)
	}

	return result, err
}

// CheckDomains 批量检测域名
func (c *Coordinator) CheckDomains(ctx context.Context, domains []string) ([]*types.DetectionResult, error) {
	if !c.running {
		return nil, fmt.Errorf("协调器未运行")
	}

	if len(domains) == 0 {
		return []*types.DetectionResult{}, nil
	}

	// 使用Worker Pool处理批量任务
	results := make([]*types.DetectionResult, len(domains))
	errors := make([]error, len(domains))

	// 创建任务
	tasks := make([]*Task, len(domains))
	for i, domain := range domains {
		tasks[i] = &Task{
			Index:  i,
			Domain: domain,
			Result: make(chan *TaskResult, 1),
		}
	}

	// 提交任务到Worker Pool
	for _, task := range tasks {
		select {
		case c.workerPool.taskQueue <- task:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// 收集结果
	for i, task := range tasks {
		select {
		case result := <-task.Result:
			results[i] = result.Result
			errors[i] = result.Error
			c.recordStats(result.Duration, result.Error == nil)
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(c.config.Concurrency.CheckTimeout):
			results[i] = &types.DetectionResult{
				Domain: domains[i],
				Error:  fmt.Errorf("检测超时"),
			}
			errors[i] = results[i].Error
		}
	}

	return results, nil
}

// CheckDomainsStream 流式批量检测域名
func (c *Coordinator) CheckDomainsStream(ctx context.Context, domains []string) (<-chan *types.DetectionResult, error) {
	if !c.running {
		return nil, fmt.Errorf("协调器未运行")
	}

	resultChan := make(chan *types.DetectionResult, len(domains))

	go func() {
		defer close(resultChan)

		// 创建任务
		tasks := make([]*Task, len(domains))
		for i, domain := range domains {
			tasks[i] = &Task{
				Index:  i,
				Domain: domain,
				Result: make(chan *TaskResult, 1),
			}
		}

		// 提交任务到Worker Pool
		for _, task := range tasks {
			select {
			case c.workerPool.taskQueue <- task:
			case <-ctx.Done():
				return
			}
		}

		// 流式收集结果
		for _, task := range tasks {
			select {
			case result := <-task.Result:
				resultChan <- result.Result
				c.recordStats(result.Duration, result.Error == nil)
			case <-ctx.Done():
				return
			case <-time.After(c.config.Concurrency.CheckTimeout):
				resultChan <- &types.DetectionResult{
					Domain: task.Domain,
					Error:  fmt.Errorf("检测超时"),
				}
			}
		}
	}()

	return resultChan, nil
}

// recordStats 记录统计信息
func (c *Coordinator) recordStats(duration time.Duration, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stats.TotalTasks++
	if success {
		c.stats.CompletedTasks++
	} else {
		c.stats.FailedTasks++
	}
}

// GetStats 获取协调器统计信息
func (c *Coordinator) GetStats() *CoordinatorStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := *c.stats
	if c.workerPool != nil {
		stats.ActiveWorkers = c.workerPool.GetActiveWorkers()
	}

	return &stats
}

// Task 任务结构
type Task struct {
	Index  int
	Domain string
	Result chan *TaskResult
}

// TaskResult 任务结果
type TaskResult struct {
	Result   *types.DetectionResult
	Error    error
	Duration time.Duration
}

// WorkerPool Worker Pool
type WorkerPool struct {
	workers      int
	taskQueue    chan *Task
	pipeline     *Pipeline
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	activeWorkers int32
	mu           sync.RWMutex
}

// NewWorkerPool 创建新的Worker Pool
func NewWorkerPool(workers int, pipeline *Pipeline) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workers:   workers,
		taskQueue: make(chan *Task, workers*2),
		pipeline:  pipeline,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start 启动Worker Pool
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop 停止Worker Pool
func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()
}

// worker Worker函数
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case task := <-wp.taskQueue:
			wp.mu.Lock()
			wp.activeWorkers++
			wp.mu.Unlock()

			startTime := time.Now()
			var result *types.DetectionResult
			var err error
			
			// 安全执行pipeline
			if wp.pipeline == nil {
				err = fmt.Errorf("pipeline未初始化")
				result = &types.DetectionResult{
					Domain:   task.Domain,
					Error:    err,
					Suitable: false,
				}
			} else {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("pipeline执行panic: %v", r)
						result = &types.DetectionResult{
							Domain:   task.Domain,
							Error:    err,
							Suitable: false,
						}
					}
				}()
				result, err = wp.pipeline.Execute(wp.ctx, task.Domain)
			}
			duration := time.Since(startTime)

			// 安全发送结果
			select {
			case task.Result <- &TaskResult{
				Result:   result,
				Error:    err,
				Duration: duration,
			}:
			case <-wp.ctx.Done():
				return
			}

			wp.mu.Lock()
			wp.activeWorkers--
			wp.mu.Unlock()

		case <-wp.ctx.Done():
			return
		}
	}
}

// GetActiveWorkers 获取活跃Worker数量
func (wp *WorkerPool) GetActiveWorkers() int {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return int(wp.activeWorkers)
}

// ConcurrencyController 并发控制器
type ConcurrencyController struct {
	config *types.Config
	stats  *ConcurrencyStats
	mu     sync.RWMutex
}

// NewConcurrencyController 创建并发控制器
func NewConcurrencyController(config *types.Config) *ConcurrencyController {
	return &ConcurrencyController{
		config: config,
		stats:  &ConcurrencyStats{},
	}
}

// RecordSuccess 记录成功请求
func (cc *ConcurrencyController) RecordSuccess(duration time.Duration) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.stats.SuccessCount++
	cc.stats.TotalResponseTime += duration
}

// RecordFailure 记录失败请求
func (cc *ConcurrencyController) RecordFailure(duration time.Duration) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.stats.FailureCount++
	cc.stats.TotalResponseTime += duration
}

// GetStats 获取并发统计
func (cc *ConcurrencyController) GetStats() *ConcurrencyStats {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	stats := *cc.stats
	stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.SuccessCount+stats.FailureCount)
	if stats.SuccessCount+stats.FailureCount > 0 {
		stats.AverageResponseTime = stats.TotalResponseTime / time.Duration(stats.SuccessCount+stats.FailureCount)
	}

	return &stats
}

// ConcurrencyStats 并发统计
type ConcurrencyStats struct {
	SuccessCount        int64         `json:"success_count"`
	FailureCount        int64         `json:"failure_count"`
	TotalResponseTime   time.Duration `json:"total_response_time"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	SuccessRate         float64       `json:"success_rate"`
}
