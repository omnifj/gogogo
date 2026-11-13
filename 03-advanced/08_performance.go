//go:build ignore

// ============================================================
// 文件：08_performance.go
// 阶段：高级
// 主题：性能优化入门
// 学习目标：
//   1. 了解 Benchmark 与 pprof 的基本用法（注释讲解）
//   2. 理解逃逸分析 go build -gcflags="-m" 的作用
//   3. 掌握 sync.Pool 对象复用与切片容量预分配
// 运行方式：go run 03-advanced/08_performance.go
// 预计阅读时间：25 分钟
// ============================================================

package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

func main() {
	fmt.Println("========== 示例 1：简易耗时对比（Benchmark 思想） ==========")
	example1SimpleTiming()

	fmt.Println()
	fmt.Println("========== 示例 2：用 runtime/pprof 写一份 CPU profile ==========")
	example2CPUProfile()

	fmt.Println()
	fmt.Println("========== 示例 3：逃逸分析演示 ==========")
	example3EscapeAnalysis()

	fmt.Println()
	fmt.Println("========== 示例 4：sync.Pool 复用对象 ==========")
	example4SyncPool()

	fmt.Println()
	fmt.Println("========== 示例 5：预分配切片容量 vs 不预分配 ==========")
	example5Prealloc()
}

// ===== 示例 1：简易耗时对比（Benchmark 思想） =====
//
// 正式的基准测试写法（本文件是 main 程序，无法直接演示，注释讲解）：
//  1. 把被测函数放进 xxx_test.go，例如：
//     func BenchmarkConcat(b *testing.B) {
//     for i := 0; i < b.N; i++ { concat() }
//     }
//  2. 运行：go test -bench=. -benchmem
//     -bench=.    运行所有 Benchmark
//     -benchmem   额外报告每次操作的内存分配（B/op、allocs/op）
//  3. b.N 由测试框架自动调整，直到测量时间足够长、结果稳定。
//
// 这里用 time.Since 做一个"思想等价"的简易对比：
func example1SimpleTiming() {
	const n = 200000

	// 写法 A：字符串用 + 拼接（每次都产生新字符串，O(n²)）
	start := time.Now()
	s := ""
	for i := 0; i < 10000; i++ {
		s += "x"
	}
	_ = s
	fmt.Printf("字符串 + 拼接耗时: %v\n", time.Since(start))

	// 写法 B：bytes.Buffer 复用缓冲区（O(n)）
	start = time.Now()
	var buf bytes.Buffer
	for i := 0; i < 10000; i++ {
		buf.WriteString("x")
	}
	_ = buf.String()
	fmt.Printf("bytes.Buffer 拼接耗时: %v\n", time.Since(start))

	fmt.Println("(对比 n =", n, "规模的场景请改用 go test -bench 做精确测量)")
	// 提示：耗时数值每次运行都不同，且与环境有关，只作相对比较
}

// ===== 示例 2：用 runtime/pprof 写一份 CPU profile =====
//
// 测试场景下的 pprof 用法（注释讲解）：
//
//	go test -cpuprofile=cpu.out -bench=.
//	go tool pprof cpu.out        # 进入交互界面，用 top / web 查看热点
//
// Web 服务场景可 import _ "net/http/pprof" 自动挂载 /debug/pprof/。
//
// 下面演示在普通程序里用 runtime/pprof 采集 CPU profile：
func example2CPUProfile() {
	// 写到系统临时目录，演示完成后清理，避免留下垃圾文件
	tmpDir, err := os.MkdirTemp("", "golearning-pprof")
	if err != nil {
		fmt.Println("创建临时目录失败:", err)
		return
	}
	defer os.RemoveAll(tmpDir) // 为什么：临时文件必须清理

	f, err := os.CreateTemp(tmpDir, "cpu-*.prof")
	if err != nil {
		fmt.Println("创建 profile 文件失败:", err)
		return
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		fmt.Println("启动 CPU profile 失败:", err)
		return
	}

	// 制造一点 CPU 负载，让 profile 有内容可分析
	sum := 0
	for i := 0; i < 3000000; i++ {
		sum += i * i % 97
	}
	pprof.StopCPUProfile()

	info, _ := f.Stat()
	fmt.Printf("CPU profile 已写入临时文件（%d 字节），可用 go tool pprof 分析；sum=%d\n", info.Size(), sum)
	// ⚠️ 注意：CPU profile 采样需要足够的工作量，太快的程序可能采不到样本
}

// ===== 示例 3：逃逸分析演示 =====
//
// 逃逸分析查看方法（注释讲解）：
//
//	go build -gcflags="-m" ./...        # 打印每个变量的逃逸决策
//	go build -gcflags="-m -m" ./...     # -m 两次，输出更详细的理由
//
// 如果变量"逃逸到堆上（escapes to heap）"，会增加 GC 压力。
//
// 下面的函数供读者单独保存后用上面的命令观察：
func stackValue() int {
	x := 42 // 只在函数内使用，通常留在栈上（does not escape）
	return x
}

func heapValue() *int {
	x := 42 // 被指针返回，必然逃逸到堆上（escapes to heap）
	return &x
}

func example3EscapeAnalysis() {
	fmt.Println("栈分配结果:", stackValue())
	fmt.Println("堆分配结果:", *heapValue())
	fmt.Println("把本文件复制出去后运行 go build -gcflags=\"-m\" 可看到逃逸决策输出")
	// ⚠️ 注意：不要盲目追求零逃逸，先保证代码正确，再用 benchmark 验证优化收益
}

// ===== 示例 4：sync.Pool 复用对象 =====
func example4SyncPool() {
	// 为什么：频繁创建/销毁的大对象会造成 GC 压力，
	// sync.Pool 在 goroutine 之间缓存临时对象，减少分配次数。
	pool := sync.Pool{
		New: func() any {
			// New 在池为空时被调用，负责创建新对象
			return new(bytes.Buffer)
		},
	}

	// 模拟多次"借用-使用-归还"
	for i := 0; i < 5; i++ {
		buf := pool.Get().(*bytes.Buffer)
		buf.Reset() // ⚠️ 注意：归还前/取出后务必 Reset，否则会读到上次的残留数据
		buf.WriteString(fmt.Sprintf("第 %d 次复用", i+1))
		fmt.Println(buf.String())
		pool.Put(buf)
	}

	// ⚠️ 注意：sync.Pool 中的对象随时可能被 GC 清掉，
	// 它只适合缓存"丢了也不可惜"的临时对象，不能当连接池使用。
	runtime.GC() // 手动触发一次 GC，演示后保持环境干净
}

// ===== 示例 5：预分配切片容量 vs 不预分配 =====
func example5Prealloc() {
	const n = 500000

	// 不预分配：append 时容量不足会反复扩容 + 拷贝底层数组
	start := time.Now()
	var noPrealloc []int
	for i := 0; i < n; i++ {
		noPrealloc = append(noPrealloc, i)
	}
	fmt.Printf("不预分配 append %d 个元素耗时: %v\n", n, time.Since(start))

	// 预分配：make([]int, 0, n) 一次分配足够容量，append 不再扩容
	// 为什么：提前知道元素个数时，预分配能避免 O(log n) 次数组拷贝
	start = time.Now()
	prealloc := make([]int, 0, n)
	for i := 0; i < n; i++ {
		prealloc = append(prealloc, i)
	}
	fmt.Printf("预分配   append %d 个元素耗时: %v\n", n, time.Since(start))

	fmt.Printf("len 一致: %v，说明两种写法结果相同，差别只在性能\n", len(noPrealloc) == len(prealloc))
	// 提示：耗时数值每次运行都不同，只作相对比较；
	// 精确对比请写成 Benchmark 并用 go test -bench=. -benchmem 验证
}
