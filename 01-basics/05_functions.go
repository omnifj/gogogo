//go:build ignore

// ============================================================
// 文件：05_functions.go
// 阶段：初级
// 主题：函数
// 学习目标：
//   1. 掌握多返回值与命名返回值
//   2. 掌握可变参数 ...int 的用法
//   3. 理解函数是一等公民：可赋值、作参数、作返回值
//   4. 理解 defer 的 LIFO 顺序及其对命名返回值的影响
// 运行方式：go run 01-basics/05_functions.go
// 预计阅读时间：15 分钟
// ============================================================

package main

import (
	"errors" // 标准库：创建错误值
	"fmt"
)

func main() {
	fmt.Println("========== 示例 1：多返回值 ==========")
	example1MultipleReturns()
	fmt.Println()

	fmt.Println("========== 示例 2：命名返回值与裸 return ==========")
	example2NamedReturns()
	fmt.Println()

	fmt.Println("========== 示例 3：可变参数 ==========")
	example3Variadic()
	fmt.Println()

	fmt.Println("========== 示例 4：函数是一等公民 ==========")
	example4FirstClass()
	fmt.Println()

	fmt.Println("========== 示例 5：defer 详解 ==========")
	example5Defer()
}

// divide 返回两个值：商和余数。多返回值是 Go 函数的天然能力，
// 不需要像其他语言那样用数组或 out 参数迂回。
func divide(a, b int) (int, int) {
	return a / b, a % b
}

// safeDivide 演示 Go 的标准模式：最后一个返回值是 error。
func safeDivide(a, b int) (int, error) {
	if b == 0 {
		// errors.New 创建一个错误值；返回 0 只是占位，不应被使用。
		return 0, errors.New("除数不能为 0")
	}
	return a / b, nil // nil 表示"没有错误"
}

// ===== 示例 1：多返回值 =====
func example1MultipleReturns() {
	quotient, remainder := divide(10, 3)
	fmt.Printf("10 / 3 = 商 %d，余 %d\n", quotient, remainder)

	// 调用方必须先检查 err 是否为 nil，再使用结果。
	result, err := safeDivide(10, 2)
	if err == nil {
		fmt.Println("10 / 2 =", result)
	}

	// 用 _ 丢弃不关心的返回值；这里只演示错误分支。
	_, err = safeDivide(10, 0)
	if err != nil {
		fmt.Println("safeDivide(10, 0) 出错：", err)
	}
	// ⚠️ 注意：声明了变量不用会编译报错，这逼着你不能"假装没看见"错误。
}

// splitSum 使用命名返回值 (total, avg)，
// 命名后返回值在函数体内就像已声明的局部变量一样可直接使用。
func splitSum(nums []int) (total int, avg float64) {
	for _, n := range nums {
		total += n
	}
	if len(nums) > 0 {
		avg = float64(total) / float64(len(nums))
	}
	return // 裸 return：自动返回 total 和 avg 的当前值
}

// ===== 示例 2：命名返回值与裸 return =====
func example2NamedReturns() {
	total, avg := splitSum([]int{80, 90, 100})
	fmt.Printf("total=%d, avg=%.2f\n", total, avg)

	// ⚠️ 注意：裸 return 存在争议。短函数里它很简洁，但长函数中
	// "到底返回了什么"要翻回签名处查看，降低可读性。
	// 社区共识：小函数可用，复杂函数请写 return total, avg 显式返回。
	fmt.Println("提示：命名返回值在小函数中简洁，大函数中慎用裸 return")
}

// sumAll 的参数 ...int 表示"接收任意多个 int"。
// 函数内部，nums 就是一个普通的 []int 切片。
func sumAll(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// ===== 示例 3：可变参数 =====
func example3Variadic() {
	// 调用时可以传 0 个、1 个或多个参数。
	fmt.Println("sumAll() =", sumAll())
	fmt.Println("sumAll(1) =", sumAll(1))
	fmt.Println("sumAll(1,2,3,4) =", sumAll(1, 2, 3, 4))

	// 已有切片时，用 切片... 展开传入。
	nums := []int{5, 6, 7}
	fmt.Println("sumAll(nums...) =", sumAll(nums...))
	// ⚠️ 注意：可变参数必须是参数列表的最后一个，且最多只能有一个。
}

// add 和 multiply 是普通函数，将被当作"值"传来传去。
func add(a, b int) int      { return a + b }
func multiply(a, b int) int { return a * b }

// apply 接收一个函数作为参数（类型 func(int, int) int），
// 体现"函数是一等公民"——函数和 int、string 一样可以被传递。
func apply(a, b int, op func(int, int) int) int {
	return op(a, b)
}

// getOperator 根据名字返回一个函数：函数也可以作为返回值。
func getOperator(name string) func(int, int) int {
	if name == "add" {
		return add
	}
	return multiply
}

// ===== 示例 4：函数是一等公民 =====
func example4FirstClass() {
	// 1. 函数可以赋值给变量，变量类型就是它的签名。
	op := add
	fmt.Println("op(3, 4) =", op(3, 4))

	// 2. 函数可以作为参数传给另一个函数（回调）。
	fmt.Println("apply(3, 4, multiply) =", apply(3, 4, multiply))

	// 3. 函数可以作为返回值返回，实现"工厂"效果。
	f := getOperator("add")
	fmt.Println("getOperator(\"add\")(2, 5) =", f(2, 5))

	// 4. 匿名函数：没有名字的函数字面量，可立即定义并使用。
	square := func(x int) int { return x * x }
	fmt.Println("匿名函数 square(6) =", square(6))
}

// deferredReturn 演示 defer 可以修改命名返回值。
// 返回 0 之后、函数真正退出之前，defer 里的 result++ 生效。
func deferredReturn() (result int) {
	defer func() {
		result++ // defer 在 return 赋值之后、函数退出之前执行
	}()
	return 0 // 先把 0 赋给 result，然后 defer 把它改成 1
}

// ===== 示例 5：defer 详解 =====
func example5Defer() {
	// defer 把函数调用推迟到"所在函数返回之前"执行。
	// 多个 defer 按 LIFO（后进先出）顺序执行。
	fmt.Println("--- defer 执行顺序（LIFO）---")
	defer fmt.Println("defer 1：最先注册，最后执行")
	defer fmt.Println("defer 2")
	defer fmt.Println("defer 3：最后注册，最先执行")
	fmt.Println("示例5 函数体的正常输出")

	// defer 修改命名返回值：返回前被 defer 拦截修改。
	fmt.Println("deferredReturn() =", deferredReturn(), "（defer 把 0 改成了 1）")

	// defer 最典型的用途：资源释放。
	// 打开资源后立刻写 defer 关闭，保证无论函数从哪个分支返回都会执行。
	fmt.Println("--- defer 模拟资源释放 ---")
	simulateResource()
	// ⚠️ 注意：defer 的参数在注册时就已求值，不是执行时才求值。
	x := 1
	defer fmt.Println("defer 捕获的 x =", x) // 注册时 x=1 已定
	x = 100
	fmt.Println("修改后 x =", x, "但 defer 打印的仍是 1")
}

// simulateResource 用打印模拟"打开资源 -> defer 关闭资源"的模式。
func simulateResource() {
	fmt.Println("打开资源（如文件/连接）")
	// 紧贴"打开"写 defer"关闭"，是 Go 的惯用写法，不容易忘记释放；
	// 即使有多个 return 分支，defer 也保证执行。
	defer fmt.Println("defer：关闭资源（函数返回前自动执行）")
	fmt.Println("使用资源……")
}
