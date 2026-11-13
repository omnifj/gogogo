//go:build ignore

// ============================================================
// 文件：01_interfaces.go
// 阶段：中级
// 主题：接口
// 学习目标：
//   1. 理解接口的定义与 Go 的隐式实现机制
//   2. 掌握空接口 any、类型断言与 type switch
//   3. 学会用小接口组合大接口，并实践 fmt.Stringer
// 运行方式：go run 02-intermediate/01_interfaces.go
// 预计阅读时间：15 分钟
// ============================================================

package main

import "fmt"

// ===== 示例 1：接口定义与隐式实现 =====

// Speaker 是一个小接口：任何有 Speak() 方法的类型都自动实现它
type Speaker interface {
	Speak() string
}

type Dog struct{ Name string }

// Dog 实现了 Speak，无需 implements 关键字，编译器自动识别
func (d Dog) Speak() string { return d.Name + "：汪汪！" }

type Cat struct{ Name string }

func (c Cat) Speak() string { return c.Name + "：喵喵！" }

func demoImplicitImpl() {
	// 接口变量可以持有任何实现了该接口的具体值
	var s Speaker
	s = Dog{Name: "旺财"}
	fmt.Println(s.Speak())
	s = Cat{Name: "咪咪"} // 同一个接口变量，换一只猫也装得下
	fmt.Println(s.Speak())

	// 多态：切片里放接口，统一调用
	animals := []Speaker{Dog{"阿黄"}, Cat{"小白"}}
	for _, a := range animals {
		fmt.Println(a.Speak())
	}
}

// ===== 示例 2：空接口 any 与类型断言 =====

// any 是 interface{} 的别名，可以装任意类型的值
func printAny(v any) {
	fmt.Printf("值 = %v，动态类型 = %T\n", v, v)
}

func demoEmptyInterface() {
	printAny(42)
	printAny("你好")
	printAny([]int{1, 2})

	var v any = "Go 语言"
	// 类型断言：从 any 中取出具体类型
	// ⚠️ 注意：直接断言失败会 panic，务必使用 comma-ok 形式
	if str, ok := v.(string); ok {
		fmt.Println("断言成功，是字符串：", str)
	}
	if n, ok := v.(int); !ok {
		fmt.Println("断言失败：v 不是 int，得到零值", n)
	}
}

// ===== 示例 3：type switch 多类型分支 =====

func describe(v any) string {
	// type switch 是处理多种可能类型的标准写法
	switch t := v.(type) {
	case int:
		return fmt.Sprintf("整数 %d，平方是 %d", t, t*t)
	case string:
		return fmt.Sprintf("字符串，长度 %d", len(t))
	case bool:
		return fmt.Sprintf("布尔值：%v", t)
	default:
		return fmt.Sprintf("未知类型 %T", t)
	}
}

func demoTypeSwitch() {
	fmt.Println(describe(7))
	fmt.Println(describe("hello"))
	fmt.Println(describe(true))
	fmt.Println(describe(3.14))
}

// ===== 示例 4：接口组合 =====

// 模仿 io.ReadWriter 的思想：小接口组合成大接口
type Reader interface {
	Read() string
}

type Writer interface {
	Write(s string)
}

// ReadWriter 由两个小接口组合而成
// 好处：函数只声明自己需要的最小能力，调用方更自由
type ReadWriter interface {
	Reader
	Writer
}

type Buffer struct{ data string }

func (b *Buffer) Read() string   { return b.data }
func (b *Buffer) Write(s string) { b.data += s }
func (b *Buffer) String() string { return b.data }

func demoInterfaceCompose() {
	// 注意：Write 修改了内部状态，所以方法定义在指针接收者上，
	// 这里必须传 *Buffer 才能满足 ReadWriter 接口
	var rw ReadWriter = &Buffer{}
	rw.Write("Hello")
	rw.Write(", Go!")
	fmt.Println("组合接口读取：", rw.Read())

	// 只需要读取能力的函数，可以只依赖小接口 Reader
	var r Reader = rw
	fmt.Println("小接口读取：", r.Read())
}

// ===== 示例 5：fmt.Stringer 实践 =====

// fmt 包打印 %v/%s 时会自动调用 String() 方法
type Point struct{ X, Y int }

// 实现 Stringer 后，打印对象就能得到自定义格式
func (p Point) String() string {
	return fmt.Sprintf("Point(%d, %d)", p.X, p.Y)
}

func demoStringer() {
	p := Point{3, 4}
	// fmt.Println 发现 Point 实现了 fmt.Stringer，自动调用 String()
	fmt.Println("直接打印：", p)
	fmt.Printf("%%v 打印：%v\n", p)
	// 未实现 Stringer 的结构体默认打印字段列表
	fmt.Println("原始结构：", struct{ A int }{1})
}

func main() {
	fmt.Println("========== 示例 1：接口定义与隐式实现 ==========")
	demoImplicitImpl()
	fmt.Println()

	fmt.Println("========== 示例 2：空接口 any 与类型断言 ==========")
	demoEmptyInterface()
	fmt.Println()

	fmt.Println("========== 示例 3：type switch 多类型分支 ==========")
	demoTypeSwitch()
	fmt.Println()

	fmt.Println("========== 示例 4：接口组合（小接口 -> 大接口） ==========")
	demoInterfaceCompose()
	fmt.Println()

	fmt.Println("========== 示例 5：fmt.Stringer 实践 ==========")
	demoStringer()
}
