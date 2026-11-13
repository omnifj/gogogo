//go:build ignore

// ============================================================
// 文件：08_generics.go
// 阶段：中级
// 主题：泛型（类型参数）
// 学习目标：
//   1. 掌握类型参数 [T any] 与约束接口的写法
//   2. 理解 comparable 与 ~ 近似约束
//   3. 会写泛型函数与泛型类型（如 Stack[T]）
//   4. 了解编译器的类型推断能力
// 运行方式：go run 02-intermediate/08_generics.go
// 预计阅读时间：18 分钟
// ============================================================

package main

import (
	"cmp"
	"fmt"
)

func main() {
	fmt.Println("========== 示例 1：泛型函数与类型推断 ==========")
	basicGeneric()
	fmt.Println()

	fmt.Println("========== 示例 2：自定义约束接口 ==========")
	customConstraint()
	fmt.Println()

	fmt.Println("========== 示例 3：comparable 与泛型 Map 键 ==========")
	comparableDemo()
	fmt.Println()

	fmt.Println("========== 示例 4：~ 近似约束 ==========")
	tildeConstraint()
	fmt.Println()

	fmt.Println("========== 示例 5：泛型类型 Stack[T] ==========")
	stackDemo()
}

// ===== 示例 1：泛型函数与类型推断 =====
// [T any] 声明类型参数 T，any 表示"任何类型"（最宽松的约束）
func First[T any](s []T) (T, bool) {
	var zero T // 零值技巧：切片为空时返回该类型的零值
	if len(s) == 0 {
		return zero, false
	}
	return s[0], true
}

func basicGeneric() {
	n, ok := First([]int{10, 20, 30}) // 编译器从实参推断出 T=int，无需手写 First[int]
	fmt.Println("First(int 切片):", n, ok)

	s, ok := First([]string{"go", "java"}) // 同一函数复用于 string
	fmt.Println("First(string 切片):", s, ok)

	_, ok = First([]float64{}) // 空切片返回零值 + false
	fmt.Println("First(空切片) ok =", ok)
	// ⚠️ 注意：推断不出类型时要显式写类型实参，如 First[int](nil)。
}

// ===== 示例 2：自定义约束接口 =====
// Number 是约束接口：列出允许的类型集合（并集用 | 分隔）
type Number interface {
	int | int64 | float64
}

// Sum 只能接受 Number 约束内的类型，编译期就拦住不合法调用
func Sum[T Number](s []T) T {
	var total T // 零值即 0，可安全累加
	for _, v := range s {
		total += v // 约束保证了 + 运算符可用
	}
	return total
}

func customConstraint() {
	fmt.Println("Sum(int):", Sum([]int{1, 2, 3}))
	fmt.Println("Sum(float64):", Sum([]float64{1.5, 2.5}))
	// ⚠️ 注意：Sum([]string{...}) 会在编译期报错——这正是约束的价值：
	// 把运行期错误提前到编译期发现。
}

// ===== 示例 3：comparable 与泛型 Map 键 =====
// cmp.Ordered 是标准库 cmp 包内置约束：所有可排序的内置类型
func MaxValue[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// IndexOf 用 comparable 约束：保证 T 支持 == 比较
func IndexOf[T comparable](s []T, target T) int {
	for i, v := range s {
		if v == target {
			return i
		}
	}
	return -1
}

func comparableDemo() {
	fmt.Println("MaxValue(3, 7):", MaxValue(3, 7))
	fmt.Println("MaxValue(\"go\", \"rust\"):", MaxValue("go", "rust"))
	fmt.Println("IndexOf 查找 \"b\":", IndexOf([]string{"a", "b", "c"}, "b"))
	// ⚠️ 注意：comparable 只允许 ==/!=；需要 <、> 时用 cmp.Ordered。
	// 切片、map、函数不可比较，不能作为 comparable 的实参。
}

// ===== 示例 4：~ 近似约束 =====
// MyInt 底层类型是 int，但它不是 int 本身
type MyInt int

// 约束写 int 时 MyInt 不满足；写成 ~int 表示"底层类型为 int 的类型均可"
func Double[T ~int | ~float64](v T) T {
	return v * 2
}

func tildeConstraint() {
	var m MyInt = 21
	fmt.Println("Double(MyInt):", Double(m))      // 没有 ~ 这里会编译失败
	fmt.Println("Double(int):", Double(21))       // 普通 int 当然也满足
	fmt.Println("Double(float64):", Double(1.25)) // 约束并集中的另一支
	// ⚠️ 注意：~ 只能用于约束接口，不能用于变量类型声明；
	// 自定义类型不会继承底层类型的方法，但会继承运算符能力。
}

// ===== 示例 5：泛型类型 Stack[T] =====
// 泛型不止能写函数，也能写类型——这是最经典的数据结构场景
type Stack[T any] struct {
	items []T
}

// 注意：泛型类型的方法，接收者要带上类型参数 [T]
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, true
}

func (s *Stack[T]) Len() int {
	return len(s.items)
}

func stackDemo() {
	st := Stack[int]{} // 泛型类型实例化时必须给出（或推断出）类型实参
	st.Push(1)
	st.Push(2)
	st.Push(3)
	fmt.Println("栈长度:", st.Len())
	for st.Len() > 0 {
		v, _ := st.Pop()
		fmt.Println("Pop:", v) // 后进先出：3, 2, 1
	}
	if _, ok := st.Pop(); !ok {
		fmt.Println("空栈 Pop 返回 ok=false")
	}
	// ⚠️ 注意：Go 泛型不支持方法的独立类型参数
	// （不能在方法上再声明 [U any]），需要时把 U 提到类型或函数层面。
}
