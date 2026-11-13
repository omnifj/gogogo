//go:build ignore

// ============================================================
// 文件：09_pointers.go
// 阶段：初级
// 主题：指针
// 学习目标：
//   1. 掌握 & 取地址与 * 解引用
//   2. 学会用指针参数实现“按引用”修改
//   3. 了解 new 函数与 nil 指针的安全判断
//   4. 对比指针接收者与值接收者
// 运行方式：go run 01-basics/09_pointers.go
// 预计阅读时间：12 分钟
// ============================================================

package main

import "fmt"

func main() {
	fmt.Println("========== 示例 1：& 取地址与 * 解引用 ==========")
	addressAndDeref()
	fmt.Println()

	fmt.Println("========== 示例 2：指针参数实现按引用修改 ==========")
	passByReference()
	fmt.Println()

	fmt.Println("========== 示例 3：new 函数 ==========")
	newFunction()
	fmt.Println()

	fmt.Println("========== 示例 4：nil 指针的安全判断 ==========")
	nilPointerSafety()
	fmt.Println()

	fmt.Println("========== 示例 5：指针接收者 vs 值接收者 ==========")
	receiverCompare()
}

// ===== 示例 1：& 取地址与 * 解引用 =====
func addressAndDeref() {
	x := 42
	p := &x // & 取变量地址，p 的类型是 *int（指向 int 的指针）

	fmt.Println("x 的值:", x)
	fmt.Println("p 保存的地址:", p) // 地址本身每次运行可能不同，只看类型和指向关系
	fmt.Println("*p 解引用:", *p) // * 解引用：读取指针指向的值

	*p = 100 // 通过指针修改，x 本身被改变
	fmt.Println("修改 *p 之后 x:", x)

	// ⚠️ 注意：* 在不同位置含义不同——
	// 类型声明中 *int 表示“指针类型”；表达式中 *p 表示“解引用取值”。
	fmt.Printf("p 的类型: %T\n", p)
}

// ===== 示例 2：指针参数实现按引用修改 =====
func passByReference() {
	a, b := 1, 2
	fmt.Printf("交换前: a=%d, b=%d\n", a, b)

	// Go 只有一种传参方式：值传递。传指针 = 把地址的副本传过去，
	// 函数内部通过地址仍然能修改调用方的变量，效果等同“按引用”
	swap(&a, &b)
	fmt.Printf("交换后: a=%d, b=%d\n", a, b)

	n := 10
	doubleValue(n) // 传值，函数改的是副本
	fmt.Println("doubleValue 后 n 不变:", n)

	doublePointer(&n) // 传地址，函数改的是原变量
	fmt.Println("doublePointer 后 n 翻倍:", n)
}

func swap(a, b *int) {
	*a, *b = *b, *a // 通过两个指针交换它们指向的值
}

func doubleValue(n int) {
	n *= 2 // 只改了副本
}

func doublePointer(n *int) {
	*n *= 2 // 改了原变量
}

// ===== 示例 3：new 函数 =====
func newFunction() {
	// new(T) 分配一块 T 类型的零值内存，返回 *T 指针
	p := new(int)
	fmt.Println("new(int) 指向的值（零值）:", *p)

	*p = 7
	fmt.Println("赋值后:", *p)

	// new 同样适用于结构体，得到指向零值结构体的指针
	type Point struct{ X, Y int }
	pt := new(Point)
	pt.X = 3 // Go 语法糖：pt.X 等价于 (*pt).X，不需要手动解引用
	pt.Y = 4
	fmt.Println("new 创建的结构体:", *pt)

	// 实际开发中 &Point{...} 字面量更常用，new 多见于需要零值指针的场景
	pt2 := &Point{X: 1, Y: 2}
	fmt.Println("字面量取地址:", *pt2)
}

// ===== 示例 4：nil 指针的安全判断 =====
func nilPointerSafety() {
	var p *int // 只声明不初始化，p 是 nil 指针，不指向任何内存

	// ⚠️ 注意：对 nil 指针解引用会 panic（运行时崩溃）！
	// 危险写法（千万不要执行）：
	//     fmt.Println(*p) // panic: runtime error: invalid memory address
	// 正确做法：解引用前先判断是否为 nil
	if p == nil {
		fmt.Println("p 是 nil 指针，跳过解引用（安全）")
	}

	// 模拟“可能返回 nil”的函数，调用方务必判空
	user := findUser(2)
	if user == nil {
		fmt.Println("未找到 id=2 的用户，使用默认名字:", "匿名")
	} else {
		fmt.Println("找到用户:", *user)
	}

	user = findUser(1)
	if user != nil {
		fmt.Println("找到用户:", *user)
	}
}

func findUser(id int) *string {
	// 模拟查找：只有 id 为 1 时返回有效指针，否则返回 nil
	if id == 1 {
		name := "小明"
		return &name
	}
	return nil
}

// ===== 示例 5：指针接收者 vs 值接收者 =====

type Wallet struct {
	balance int
}

// 值接收者：操作副本，余额不会真正变化
func (w Wallet) DepositByValue(amount int) {
	w.balance += amount
}

// 指针接收者：操作原对象，余额真正变化
func (w *Wallet) DepositByPointer(amount int) {
	w.balance += amount
}

func (w Wallet) Balance() int {
	return w.balance
}

func receiverCompare() {
	w := Wallet{balance: 100}

	w.DepositByValue(50)
	fmt.Println("值接收者存入 50 后余额:", w.Balance()) // 仍是 100，修改丢失

	w.DepositByPointer(50)
	fmt.Println("指针接收者存入 50 后余额:", w.Balance()) // 变为 150

	// 选择建议：需要修改状态、结构体较大时用指针接收者；
	// 一个类型的方法应统一接收者风格，避免一半值一半指针。
}
