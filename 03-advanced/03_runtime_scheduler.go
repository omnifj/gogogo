//go:build ignore

// ============================================================
// 文件：03_runtime_scheduler.go
// 阶段：高级
// 主题：运行时与调度器
// 学习目标：
//   1. 理解 GMP 调度模型中 G/M/P 的角色
//   2. 掌握 runtime.Gosched / GOMAXPROCS / NumGoroutine
//   3. 学会用逃逸分析与 testing.AllocsPerRun 观察内存分配
//   4. 了解 GOGC 与 debug.SetGCPercent 对 GC 频率的影响
// 运行方式：go run 03-advanced/03_runtime_scheduler.go
// 预计阅读时间：25 分钟
// ============================================================

package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
)

// ===== 示例 1：GMP 模型速览（注释讲解 + 小实验） =====
//
// G（Goroutine）： goroutine 本身，保存栈、指令指针等执行现场。
// M（Machine）：  OS 线程的封装，真正在 CPU 上跑代码的实体。
// P（Processor）：逻辑处理器，持有一个可运行 G 的本地队列；
//
//	M 必须绑定一个 P 才能执行 G。P 的数量 = GOMAXPROCS。
//
// 调度要点：P 的本地队列空了会去“偷”（work stealing）其他 P 的 G，
// 所以少量线程就能跑海量 goroutine。
func demoGMP() {
	fmt.Println("========== 示例 1：GMP 模型 ==========")
	fmt.Println("CPU 核数（runtime.NumCPU）:", runtime.NumCPU())
	fmt.Println("当前 P 的数量（GOMAXPROCS）:", runtime.GOMAXPROCS(0))
	fmt.Println("main 也是一个 G，当前 G 数量:", runtime.NumGoroutine())

	// 小实验：启动 5 个 G，观察 G 数量变化
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); busy(1e6) }()
	}
	fmt.Println("启动 5 个 G 后，G 数量:", runtime.NumGoroutine())
	wg.Wait()
	fmt.Println("提示：G 再多，M/P 也只有几个，靠调度器分时复用")
	fmt.Println()
}

// busy 做一点无意义计算，防止 goroutine 瞬间结束看不到数量变化
func busy(n int) {
	x := 0
	for i := 0; i < n; i++ {
		x += i
	}
	_ = x
}

// ===== 示例 2：runtime.Gosched 主动让出 CPU =====
func demoGosched() {
	fmt.Println("========== 示例 2：Gosched 让出 ==========")
	done := make(chan struct{})
	// 这个 goroutine 每打印一次就主动让出，让对方有机会先跑
	go func() {
		for i := 0; i < 3; i++ {
			fmt.Println("协程 A 第", i+1, "次执行")
			runtime.Gosched() // 把自己放回队列尾部，让其他 G 先运行
		}
		close(done)
	}()
	for i := 0; i < 3; i++ {
		fmt.Println("协程 B(main) 第", i+1, "次执行")
		runtime.Gosched()
	}
	<-done
	fmt.Println("提示：输出顺序可能不同，Gosched 只是“建议”调度器换人")
	fmt.Println()
}

// ===== 示例 3：GOMAXPROCS 观察与调整 =====
func demoGomaxprocs() {
	fmt.Println("========== 示例 3：GOMAXPROCS ==========")
	old := runtime.GOMAXPROCS(0) // 传 0 只读取不修改
	fmt.Println("默认 P 数量:", old)

	// 调小为 1：所有 G 挤在一个 P 上串行调度
	runtime.GOMAXPROCS(1)
	fmt.Println("调整为 1 后:", runtime.GOMAXPROCS(0))

	// ⚠️ 注意：容器环境中 NumCPU 看到的是宿主机核数，
	// 生产环境可能需要按 cgroup 配额显式设置 GOMAXPROCS
	runtime.GOMAXPROCS(old) // 恢复，避免影响后面的示例
	fmt.Println("已恢复为:", runtime.GOMAXPROCS(0))
	fmt.Println()
}

// ===== 示例 4：NumGoroutine 统计与泄漏观察 =====
func demoNumGoroutine() {
	fmt.Println("========== 示例 4：NumGoroutine ==========")
	before := runtime.NumGoroutine()

	// 模拟泄漏：goroutine 阻塞在永不写入的通道上
	leak := make(chan struct{})
	for i := 0; i < 3; i++ {
		go func() { <-leak }()
	}
	during := runtime.NumGoroutine()
	fmt.Printf("泄漏前 %d → 泄漏中 %d（多了 %d 个）\n", before, during, during-before)

	// 关闭通道统一放行，goroutine 才能退出
	close(leak)
	// 给调度器一点时间回收
	for i := 0; i < 100 && runtime.NumGoroutine() > before; i++ {
		runtime.Gosched()
	}
	fmt.Println("修复后:", runtime.NumGoroutine())
	fmt.Println("⚠️ 注意：生产代码可用 pprof goroutine profile 定位类似泄漏")
	fmt.Println()
}

// stackAlloc 返回值本身，通常不发生堆分配
func stackAlloc() int {
	x := 42 // x 只在函数内使用，留在栈上，函数返回即释放
	return x
}

// heapAlloc 返回局部变量地址，x 会“逃逸”到堆上
//
//go:noinline
func heapAlloc() *int {
	x := 42
	return &x // 指针被带出了函数，栈帧销毁后仍需存活，只能分配在堆上
}

// ===== 示例 5：逃逸分析与分配次数 =====
//
// 读者可自行运行下面的命令，查看编译器的逃逸分析报告：
//
//	go build -gcflags="-m" 03_runtime_scheduler.go
//
// 说明：本文件带 //go:build ignore 标签，go build 默认会跳过它，
// 可把本文件复制到无标签的临时目录再执行上述命令。
// 报告里 "escapes to heap" 表示变量逃逸到堆，"does not escape" 表示留在栈上。
func demoEscapeAnalysis() {
	fmt.Println("========== 示例 5：逃逸分析 ==========")
	// testing.AllocsPerRun 测量每次调用的平均堆分配次数
	// ⚠️ 注意：它只统计堆分配，栈分配不计入
	stackAllocs := testing.AllocsPerRun(1000, func() { _ = stackAlloc() })
	heapAllocs := testing.AllocsPerRun(1000, func() { _ = heapAlloc() })

	fmt.Printf("stackAlloc 每次堆分配: %.0f 次（不逃逸）\n", stackAllocs)
	fmt.Printf("heapAlloc  每次堆分配: %.0f 次（逃逸到堆）\n", heapAllocs)
	fmt.Println("提示：逃逸越多 GC 压力越大，热路径应尽量减少逃逸")
	fmt.Println()
}

// ===== 示例 6：GOGC 与 GC 频率观察 =====
func demoGOGC() {
	fmt.Println("========== 示例 6：GOGC 简介 ==========")
	// GOGC=100（默认）表示堆增长到上次 GC 后的 2 倍时触发下次 GC
	old := debug.SetGCPercent(100)

	var ms runtime.MemStats
	triggerGC := func(label string) {
		// 分配一批临时对象制造垃圾
		sink := make([][]byte, 0, 2000)
		for i := 0; i < 2000; i++ {
			sink = append(sink, make([]byte, 1024))
		}
		_ = sink
		runtime.GC() // 手动触发一次，便于观察
		runtime.ReadMemStats(&ms)
		fmt.Printf("%s：累计 GC 次数 %d，当前堆 %d KB\n",
			label, ms.NumGC, ms.HeapAlloc/1024)
	}

	triggerGC("GOGC=100")

	// 调大 GOGC：堆允许长更大才 GC → GC 更少但内存占用更高
	debug.SetGCPercent(400)
	triggerGC("GOGC=400")

	// ⚠️ 注意：debug.SetGCPercent(-1) 会完全关闭 GC，生产慎用
	debug.SetGCPercent(old) // 恢复默认值
	fmt.Println("提示：耗时与内存数值每次运行可能不同；GOGC 是内存与 CPU 的权衡")
}

func main() {
	demoGMP()
	demoGosched()
	demoGomaxprocs()
	demoNumGoroutine()
	demoEscapeAnalysis()
	demoGOGC()
}
