//go:build ignore

// ============================================================
// 文件：02_concurrency_patterns.go
// 阶段：高级
// 主题：并发模式
// 学习目标：
//   1. 掌握 worker pool、pipeline、fan-out/fan-in 三种经典模式
//   2. 用 time.Ticker 做限流
//   3. 手写简易 errgroup：错误收集 + 失败即取消
//   4. 用 context 实现优雅退出，保证无 goroutine 泄漏
// 运行方式：go run 03-advanced/02_concurrency_patterns.go
// 预计阅读时间：25 分钟
// ============================================================

package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ===== 示例 1：worker pool（固定工人消费任务通道） =====
func demoWorkerPool() {
	fmt.Println("========== 示例 1：worker pool ==========")
	const workerCount = 3 // 工人数量固定，任务再多也不新增 goroutine
	const jobCount = 8
	jobs := make(chan int)    // 任务通道
	results := make(chan int) // 结果通道
	var wg sync.WaitGroup
	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// range 通道：通道关闭且取空后自动退出循环
			for j := range jobs {
				// 提示：输出顺序可能不同，但程序必须确定性退出、无泄漏
				fmt.Printf("worker %d 处理任务 %d\n", id, j)
				results <- j * j
			}
		}(w)
	}
	// 发送任务（在另一个 goroutine 中，避免阻塞 main）
	go func() {
		for j := 1; j <= jobCount; j++ {
			jobs <- j
		}
		close(jobs) // ⚠️ 注意：发完必须关闭，worker 才能退出循环
	}()
	// 等待所有 worker 结束后关闭结果通道
	go func() {
		wg.Wait()
		close(results)
	}()
	sum := 0
	for r := range results {
		sum += r
	}
	fmt.Println("所有结果平方和:", sum) // 1²+...+8² = 204，结果确定
	fmt.Println()
}

// ===== 示例 2：pipeline（生成 → 加工 → 汇出） =====
func demoPipeline() {
	fmt.Println("========== 示例 2：pipeline ==========")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 阶段 1：生成 1~5
	gen := func(ctx context.Context) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out) // 每阶段负责关闭自己的输出通道，下游才能感知结束
			for i := 1; i <= 5; i++ {
				select {
				case out <- i:
				case <-ctx.Done(): // 支持取消，防止上游泄漏
					return
				}
			}
		}()
		return out
	}
	// 阶段 2：加工（乘 10）
	square := func(ctx context.Context, in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for n := range in {
				select {
				case out <- n * 10:
				case <-ctx.Done():
					return
				}
			}
		}()
		return out
	}
	// 阶段 3：汇出（main 直接消费）
	for v := range square(ctx, gen(ctx)) {
		fmt.Print(v, " ")
	}
	fmt.Println("\npipeline 结束")
	fmt.Println()
}

// ===== 示例 3：fan-out / fan-in =====
func demoFanOutFanIn() {
	fmt.Println("========== 示例 3：fan-out / fan-in ==========")
	in := make(chan int)
	// fan-in：把多个输入通道合并成一个
	fanIn := func(chs ...<-chan int) <-chan int {
		out := make(chan int)
		var wg sync.WaitGroup
		wg.Add(len(chs))
		for _, ch := range chs {
			go func(c <-chan int) {
				defer wg.Done()
				for n := range c {
					out <- n
				}
			}(ch)
		}
		go func() {
			wg.Wait()
			close(out) // ⚠️ 注意：必须等所有输入排空才能关闭，否则发送会 panic
		}()
		return out
	}

	work := func(in <-chan int) <-chan int { // 单 goroutine 处理函数
		out := make(chan int)
		go func() {
			defer close(out)
			for n := range in {
				out <- n * 2
			}
		}()
		return out
	}
	// 数据源
	go func() {
		for i := 1; i <= 6; i++ {
			in <- i
		}
		close(in)
	}()
	// fan-out：同一输入分发给 3 个 worker；fan-in：合并结果
	merged := fanIn(work(in), work(in), work(in))
	total := 0
	for v := range merged {
		total += v // 提示：输出顺序可能不同，但程序必须确定性退出、无泄漏
	}
	fmt.Println("fan-in 汇总:", total) // 2*(1+..+6)=42，结果确定
	fmt.Println()
}

// ===== 示例 4：time.Ticker 限流 =====
func demoRateLimit() {
	fmt.Println("========== 示例 4：Ticker 限流 ==========")
	// 每 100ms 放行一个任务，模拟“每秒最多 10 个请求”
	limiter := time.NewTicker(100 * time.Millisecond)
	defer limiter.Stop() // ⚠️ 注意：Ticker 必须 Stop，否则底层定时器不释放
	start := time.Now()
	for i := 1; i <= 4; i++ {
		<-limiter.C // 阻塞等待下一次“放行令牌”
		fmt.Printf("任务 %d 在 %v 时执行\n", i, time.Since(start).Round(50*time.Millisecond))
	}
	fmt.Println("提示：耗时数值每次运行可能略有不同，但节奏约为每 100ms 一个")
	fmt.Println()
}

// group 是手写的简易 errgroup：WaitGroup + 错误收集 + 失败即取消
type group struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	once   sync.Once
	err    error
}

func newGroup(parent context.Context) *group {
	ctx, cancel := context.WithCancel(parent)
	return &group{ctx: ctx, cancel: cancel}
}

// Go 启动一个任务；任务返回非 nil 错误时取消整个组
func (g *group) Go(f func(context.Context) error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(g.ctx); err != nil {
			g.once.Do(func() { // 只记录第一个错误，避免并发写竞争
				g.err = err
				g.cancel() // 任意任务失败即取消，通知其他任务尽快退出
			})
		}
	}()
}

// Wait 等待全部任务结束并返回第一个错误
func (g *group) Wait() error {
	g.wg.Wait()
	g.cancel() // 释放 context 资源
	return g.err
}

// ===== 示例 5：简易 errgroup + context 优雅退出 =====
func demoErrgroup() {
	fmt.Println("========== 示例 5：errgroup 与优雅退出 ==========")
	g := newGroup(context.Background())
	// 任务 1：正常干活，但会被任务 2 的失败取消
	g.Go(func(ctx context.Context) error {
		for i := 1; ; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("任务 1 收到取消信号，清理后退出")
				return nil // 收到取消不算错误
			default:
				time.Sleep(30 * time.Millisecond)
			}
		}
	})
	// 任务 2：第 3 步时模拟失败
	g.Go(func(ctx context.Context) error {
		for i := 1; i <= 3; i++ {
			time.Sleep(50 * time.Millisecond)
			if i == 3 {
				return fmt.Errorf("任务 2 在第 %d 步失败", i)
			}
		}
		return nil
	})
	err := g.Wait()
	fmt.Println("errgroup 返回:", err)
	// 校验无 goroutine 泄漏：等一小会儿让收尾 goroutine 退出
	time.Sleep(50 * time.Millisecond)
	fmt.Println("存活 goroutine 数:", runtime.NumGoroutine()) // 应接近 1
}

func main() {
	demoWorkerPool()
	demoPipeline()
	demoFanOutFanIn()
	demoRateLimit()
	demoErrgroup()
}
