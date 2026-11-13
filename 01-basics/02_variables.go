//go:build ignore

// ============================================================
// 文件：02_variables.go
// 阶段：初级
// 主题：变量与常量
// 学习目标：
//   1. 掌握 var 声明与类型推断
//   2. 掌握 := 短声明及其使用限制
//   3. 理解 const 与 iota 枚举
//   4. 理解零值与变量作用域
// 运行方式：go run 01-basics/02_variables.go
// 预计阅读时间：12 分钟
// ============================================================

package main

import "fmt"

// ===== 包级变量 =====
// 在所有函数之外声明的变量是"包级变量"，整个包的任何函数都能访问。
// 包级变量只能用 var 声明，不能使用 := 短声明。
var packageLevel = "我是包级变量（var 声明）"

// 包级变量同样支持类型推断，编译器根据初始值确定类型。
var packageCount = 100

func main() {
	fmt.Println("========== 示例 1：var 声明与类型推断 ==========")
	example1VarDeclaration()
	fmt.Println()

	fmt.Println("========== 示例 2：:= 短声明 ==========")
	example2ShortDeclaration()
	fmt.Println()

	fmt.Println("========== 示例 3：const 与 iota ==========")
	example3ConstAndIota()
	fmt.Println()

	fmt.Println("========== 示例 4：零值 ==========")
	example4ZeroValues()
	fmt.Println()

	fmt.Println("========== 示例 5：变量作用域 ==========")
	example5Scope()
}

// ===== 示例 1：var 声明与类型推断 =====
func example1VarDeclaration() {
	// 完整写法：var 变量名 类型 = 初始值
	var name string = "小明"
	// 类型推断：省略类型，编译器根据初始值推断，效果完全一样。
	// Go 鼓励这种写法，因为代码更简洁。
	var age = 18
	// 一次声明多个变量（批量声明）。
	var (
		city     = "北京"
		isActive = true
	)
	fmt.Printf("name=%s, age=%d, city=%s, isActive=%t\n", name, age, city, isActive)

	// ⚠️ 注意：推断出的类型是固定不变的。
	// age 被推断为 int，之后不能再给它赋字符串或小数。
	// 例如 age = 18.5 会编译报错：constant 18.5 truncated to integer。
	fmt.Printf("包级变量：%s, packageCount=%d\n", packageLevel, packageCount)
}

// ===== 示例 2：:= 短声明 =====
func example2ShortDeclaration() {
	// := 是"声明并初始化"的简写，只能在函数内部使用。
	// 它是 Go 代码中最常见的变量声明方式。
	score := 95.5
	fmt.Printf("score = %.1f（类型自动推断为 %T）\n", score, score)

	// := 可以同时声明多个变量（多重赋值）。
	x, y := 1, 2
	// 利用多重赋值可以一行交换两个变量，不需要临时变量。
	x, y = y, x
	fmt.Printf("交换后：x=%d, y=%d\n", x, y)

	// ⚠️ 注意 1：:= 不能用于包级变量，函数外只能写 var。
	// ⚠️ 注意 2：同一作用域内不能重复用 := 声明同名变量。
	// 例如再写 score := 100 会报 "no new variables on left side of :="。
	// 但只要左边有至少一个新变量，就允许混合声明：
	a := 1
	a, b := 2, 3 // b 是新变量，所以合法；a 是重新赋值
	fmt.Printf("混合声明：a=%d, b=%d\n", a, b)
}

// ===== 示例 3：const 与 iota =====
func example3ConstAndIota() {
	// const 声明常量，编译期就确定，运行期不可修改。
	const pi = 3.14159
	// 常量同样支持类型推断，也可以显式写类型。
	const appName string = "go-learning"
	fmt.Printf("pi=%v, appName=%s\n", pi, appName)

	// iota 是常量组里的"行号计数器"，从 0 开始逐行加一，
	// 常用于定义枚举。Go 没有 enum 关键字，iota 就是替代品。
	const (
		Monday    = iota // 0
		Tuesday          // 1，后续行省略表达式时自动沿用上一行
		Wednesday        // 2
		Thursday         // 3
		Friday           // 4
	)
	fmt.Printf("Monday=%d, Wednesday=%d, Friday=%d\n", Monday, Wednesday, Friday)

	// iota 还可以结合表达式，常用于定义字节单位。
	const (
		_  = 1 << (10 * iota) // 忽略第 0 行（用空白标识符 _ 丢弃）
		KB                    // 1 << 10 = 1024
		MB                    // 1 << 20
		GB                    // 1 << 30
	)
	fmt.Printf("KB=%d, MB=%d, GB=%d\n", KB, MB, GB)
	// ⚠️ 注意：常量一旦声明就不能赋值，pi = 3.14 会编译报错。
}

// ===== 示例 4：零值 =====
func example4ZeroValues() {
	// 只声明不初始化时，变量会被自动赋"零值"，Go 保证没有未初始化的内存。
	// 这与 C 语言"未初始化变量是随机垃圾值"完全不同，更安全。
	var i int
	var f float64
	var b bool
	var s string
	// %q 能直观看到空字符串是 "" 而不是乱码。
	fmt.Printf("int 零值: %d\n", i)
	fmt.Printf("float64 零值: %v\n", f)
	fmt.Printf("bool 零值: %t\n", b)
	fmt.Printf("string 零值: %q\n", s)
	// 常见零值速查：数值型是 0，布尔是 false，字符串是 ""，
	// 指针/切片/map 等引用类型是 nil（后续章节会讲）。
}

// ===== 示例 5：变量作用域 =====
func example5Scope() {
	// 作用域：变量在代码中"可见"的范围，由花括号 {} 决定。
	functionVar := "函数级变量"
	fmt.Println("函数内：", functionVar)
	fmt.Println("函数内也能访问包级变量：", packageLevel)

	{
		// 块级变量：只在当前 {} 块内有效，出了块就销毁。
		// 利用这一点可以限制变量的生命周期，避免污染外层。
		blockVar := "块级变量"
		fmt.Println("块内：", blockVar)
		// 内层可以访问外层的变量。
		fmt.Println("块内访问外层：", functionVar)
	}
	// ⚠️ 注意：这里如果打印 blockVar 会编译报错 undefined。
	// 注释掉的那行就是错误示范：
	// fmt.Println(blockVar) // 编译错误：undefined: blockVar

	// 变量遮蔽（shadowing）：内层声明同名变量会"遮住"外层变量。
	shadow := "外层"
	{
		shadow := "内层" // 这是一个新变量，不是修改外层
		fmt.Println("遮蔽演示 - 内层：", shadow)
	}
	fmt.Println("遮蔽演示 - 外层：", shadow)
	// ⚠️ 注意：遮蔽是常见 bug 来源，内层修改的并不是你以为的那个变量。
}
