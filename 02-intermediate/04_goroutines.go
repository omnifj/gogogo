//go:build ignore

// ============================================================
// 文件：04_goroutines.go
// 阶段：中级
// 主题：goroutine 并发
// 学习目标：
//   1. 掌握 go 语句启动 goroutine 与主 goroutine 的生命周期
//   2. 用 sync.WaitGroup 等待所有 goroutine 完成
//   3. 观察 runtime.GOMAXPROCS 与 goroutine 数量
//   4. 理解 goroutine 泄漏并用 runtime.NumGoroutine 演示
// 运行方式：go run 02-intermediate/04_goroutines.go
// 预计阅读时间：18 分钟
// ============================================================

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ===== 示例 1：go 语句与主 goroutine 的生命周期 =====

func worker(name string) {
	fmt.Println("  worker", name, "正在执行")
}

func demoGoStatement() {
	// go 语句：把函数调用放进一个新的 goroutine 异步执行
	go worker("A")
	go worker("B")

	// ⚠️ 注意：main 函数结束（主 goroutine 退出）时，整个程序立即终止，
	// 不管其他 goroutine 是否跑完。所以这里必须等待。
	time.Sleep(50 * time.Millisecond) // 简单等待，仅演示；生产代码请用 WaitGroup
	// 提示：输出顺序可能不同（goroutine 调度顺序不确定）
	fmt.Println("  main 等到 worker 执行完毕")
}

// ===== 示例 2：sync.WaitGroup 等待 =====

func demoWaitGroup() {
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1) // 每启动一个 goroutine，计数加一
		go func(id int) {
			defer wg.Done() // 完成时计数减一（defer 保证一定执行）
			fmt.Printf("  任务 %d 完成\n", id)
			// 提示：输出顺序可能不同
		}(i) // i 作为参数传入，语义清晰
	}
	wg.Wait() // 阻塞直到计数归零——程序确定退出，不依赖 sleep
	fmt.Println("  所有任务已完成（WaitGroup 保证）")
}

// ===== 示例 3：runtime.GOMAXPROCS 观察 =====

func demoGOMAXPROCS() {
	// GOMAXPROCS 控制可同时执行 Go 代码的逻辑 CPU 数，默认等于 CPU 核数
	fmt.Println("  逻辑 CPU 数：", runtime.NumCPU())
	fmt.Println("  当前 GOMAXPROCS：", runtime.GOMAXPROCS(0))

	old := runtime.GOMAXPROCS(2) // 临时改为 2，返回值是旧值
	fmt.Println("  改为 2 后：", runtime.GOMAXPROCS(0))
	runtime.GOMAXPROCS(old) // 恢复原值，避免影响后续示例

	// ⚠️ 注意：GOMAXPROCS 限制的是"并行"执行数量，goroutine 总数不受它限制，
	// 成千上万个 goroutine 依然可以在少量 P 上被调度。
}

// ===== 示例 4：goroutine 泄漏演示（错误示范，有对照） =====

// leak 里的 goroutine 永远阻塞在一个没人接收的 channel 上——这就是泄漏
func demoLeak() {
	before := runtime.NumGoroutine()
	ch := make(chan int)
	go func() {
		// 没有接收者，这个 goroutine 永远卡在这里，无法退出
		<-ch
	}()
	time.Sleep(20 * time.Millisecond) // 给调度器一点时间启动它
	after := runtime.NumGoroutine()
	fmt.Printf("  泄漏前 goroutine 数：%d，泄漏后：%d\n", before, after)
	fmt.Println("  ⚠️ 泄漏的 goroutine 永远无法回收，堆积多了会耗尽内存")
	// 本示例故意留下 1 个泄漏的 goroutine 用于对比；
	// 程序由 main 正常结束，进程退出时它随之销毁。
}

// ===== 示例 5：正确退出——goroutine 数回落 =====

func demoCleanExit() {
	before := runtime.NumGoroutine()
	done := make(chan struct{})

	for i := 0; i < 3; i++ {
		go func(id int) {
			defer fmt.Printf("  worker %d 已退出\n", id)
			// 提示：输出顺序可能不同
			select {
			case <-done: // 收到退出信号，主动结束，不再泄漏
				return
			case <-time.After(5 * time.Second): // 兜底超时，防止永远阻塞
				return
			}
		}(i)
	}

	time.Sleep(20 * time.Millisecond)
	fmt.Println("  运行中 goroutine 数：", runtime.NumGoroutine())

	close(done) // 广播退出信号：所有监听 done 的 goroutine 都被唤醒
	time.Sleep(20 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Printf("  收尾前 goroutine 数：%d，收尾后：%d（已回落）\n", before+3, after)
}

func main() {
	fmt.Println("========== 示例 1：go 语句与主 goroutine 生命周期 ==========")
	demoGoStatement()
	fmt.Println()

	fmt.Println("========== 示例 2：sync.WaitGroup 等待 ==========")
	demoWaitGroup()
	fmt.Println()

	fmt.Println("========== 示例 3：runtime.GOMAXPROCS 观察 ==========")
	demoGOMAXPROCS()
	fmt.Println()

	fmt.Println("========== 示例 4：goroutine 泄漏演示 ==========")
	demoLeak()
	fmt.Println()

	fmt.Println("========== 示例 5：正确退出，goroutine 数回落 ==========")
	demoCleanExit()
}
