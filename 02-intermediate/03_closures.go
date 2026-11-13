//go:build ignore

// ============================================================
// 文件：03_closures.go
// 阶段：中级
// 主题：闭包
// 学习目标：
//   1. 理解闭包按引用捕获变量的机制
//   2. 用闭包实现计数器与生成器函数
//   3. 了解函数式选项模式雏形
//   4. 认清循环变量闭包陷阱（<1.22 历史问题）与 1.22+ 的新行为
// 运行方式：go run 02-intermediate/03_closures.go
// 预计阅读时间：15 分钟
// ============================================================

package main

import "fmt"

// ===== 示例 1：闭包按引用捕获变量 =====

func demoCapture() {
	msg := "你好"
	// 闭包捕获的是变量本身（引用），而不是定义那一刻的值副本
	say := func() {
		fmt.Println("闭包读到：", msg)
	}
	say()      // 此时 msg 还是 "你好"
	msg = "世界" // 修改外层变量
	say()      // 闭包读到的是新值，证明捕获的是引用
}

// ===== 示例 2：闭包实现计数器 =====

// newCounter 返回一个函数，每次调用让内部 count 加一
// count 是自由变量：外层函数返回后它依然存活，被闭包独占
func newCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func demoCounter() {
	c1 := newCounter()
	c2 := newCounter() // 每个计数器有自己独立的 count，互不干扰
	fmt.Println("c1:", c1(), c1(), c1())
	fmt.Println("c2:", c2(), c2())
}

// ===== 示例 3：斐波那契生成器 =====

// fibGen 每次调用返回下一个斐波那契数，状态 a、b 由闭包保存
func fibGen() func() int {
	a, b := 0, 1
	return func() int {
		next := a
		a, b = b, a+b
		return next
	}
}

func demoGenerator() {
	fib := fibGen()
	fmt.Print("斐波那契前 10 项：")
	for i := 0; i < 10; i++ {
		fmt.Print(fib(), " ")
	}
	fmt.Println()
}

// ===== 示例 4：函数式选项模式雏形 =====

type Server struct {
	host string
	port int
}

// Option 是"配置函数"：接收 Server 指针并修改它
type Option func(*Server)

// WithHost 返回一个闭包，把 host 参数"封存"进函数里
func WithHost(h string) Option {
	return func(s *Server) { s.host = h }
}

func WithPort(p int) Option {
	return func(s *Server) { s.port = p }
}

// NewServer 先给默认值，再依次应用各选项——调用方按需组合
func NewServer(opts ...Option) *Server {
	s := &Server{host: "localhost", port: 8080} // 默认值
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func demoOptions() {
	defaultSrv := NewServer()
	custom := NewServer(WithHost("0.0.0.0"), WithPort(9090))
	fmt.Printf("默认配置：%+v\n", *defaultSrv)
	fmt.Printf("自定义配置：%+v\n", *custom)
}

// ===== 示例 5：循环变量闭包陷阱（历史 vs 当前） =====

func demoLoopClosure() {
	// 【历史陷阱（Go < 1.22）】
	// 旧版本中循环变量 i 是"同一个变量"被反复赋值，
	// 闭包捕获引用后，循环结束时所有闭包读到的都是最终值 3。
	//
	// 【当前行为（Go 1.22+，本机 Go 1.27）】
	// 每次迭代 i 都是"新变量"，下面的输出是 0 1 2。
	funcs := make([]func(), 0, 3)
	for i := 0; i < 3; i++ {
		funcs = append(funcs, func() {
			fmt.Print(i, " ") // 1.22+：每个闭包捕获各自迭代的新 i
		})
	}
	fmt.Print("1.22+ 实际输出：")
	for _, f := range funcs {
		f()
	}
	fmt.Println()

	// 如果你确实想让所有闭包共享同一个变量，正确写法是显式声明在循环外
	shared := 0
	shareFuncs := make([]func(), 0, 3)
	for i := 0; i < 3; i++ {
		shareFuncs = append(shareFuncs, func() {
			shared++ // 三个闭包操作同一个 shared
		})
	}
	for _, f := range shareFuncs {
		f()
	}
	fmt.Println("显式共享变量：三个闭包调用后 shared =", shared)

	// ⚠️ 注意：在 Go 1.22 之前，若想在循环里保留每轮的值，
	// 惯用写法是 `i := i` 或把 i 作为参数传进闭包；新代码不再需要。
}

func main() {
	fmt.Println("========== 示例 1：闭包按引用捕获变量 ==========")
	demoCapture()
	fmt.Println()

	fmt.Println("========== 示例 2：闭包实现计数器 ==========")
	demoCounter()
	fmt.Println()

	fmt.Println("========== 示例 3：斐波那契生成器 ==========")
	demoGenerator()
	fmt.Println()

	fmt.Println("========== 示例 4：函数式选项模式雏形 ==========")
	demoOptions()
	fmt.Println()

	fmt.Println("========== 示例 5：循环变量闭包陷阱（历史 vs 1.22+） ==========")
	demoLoopClosure()
}
