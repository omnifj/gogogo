//go:build ignore

// ============================================================
// 文件：09_unsafe.go
// 阶段：高级
// 主题：unsafe 包入门
// 学习目标：
//   1. 理解 unsafe.Pointer 与 uintptr 的关系
//   2. 用 Sizeof/Alignof/Offsetof 观察结构体布局与内存对齐
//   3. 演示 string↔[]byte 零拷贝转换（只读使用）及其风险
// 运行方式：go run 03-advanced/09_unsafe.go
// 预计阅读时间：20 分钟
// ============================================================

package main

import (
	"fmt"
	"unsafe"
)

// 用于观察内存布局的结构体
type inner struct {
	a int8  // 1 字节
	b int64 // 8 字节
	c int16 // 2 字节
}

func main() {
	fmt.Println("========== 示例 1：unsafe.Pointer 与 uintptr 的关系 ==========")
	example1PointerAndUintptr()

	fmt.Println()
	fmt.Println("========== 示例 2：Sizeof / Alignof / Offsetof 观察内存布局 ==========")
	example2Layout()

	fmt.Println()
	fmt.Println("========== 示例 3：调整字段顺序节省内存 ==========")
	example3Reorder()

	fmt.Println()
	fmt.Println("========== 示例 4：string ↔ []byte 零拷贝转换（只读） ==========")
	example4ZeroCopy()

	fmt.Println()
	fmt.Println("========== 示例 5：unsafe 风险总结 ==========")
	example5RiskSummary()
}

// ===== 示例 1：unsafe.Pointer 与 uintptr 的关系 =====
func example1PointerAndUintptr() {
	x := int64(0x1122334455667788)

	// 普通指针 -> unsafe.Pointer -> 其他类型指针：可以绕过类型系统
	p := unsafe.Pointer(&x)

	// unsafe.Pointer -> uintptr：得到地址的数值表示
	addr := uintptr(p)
	fmt.Printf("变量地址（数值）: 0x%x\n", addr)

	// 为什么区分两者：
	//   unsafe.Pointer 是"指针"，GC 会跟踪它指向的对象；
	//   uintptr 只是"整数"，GC 不认为它持有对象。
	// ⚠️ 注意：不能把 unsafe.Pointer 转成 uintptr 存起来，
	//   隔几行再转回来使用——期间 GC 可能移动或回收该对象，
	//   uintptr 不会阻止回收，转回后就是悬空指针。
	//   正确做法：转换和使用必须在同一表达式/紧邻语句中完成。

	// 安全示例：同一行内完成"地址运算"，读取 x 的低 4 字节
	low := *(*int32)(unsafe.Pointer(uintptr(p))) // 紧邻使用，无中间存储
	fmt.Printf("低 4 字节（小端）: 0x%x\n", low)
}

// ===== 示例 2：Sizeof / Alignof / Offsetof 观察内存布局 =====
func example2Layout() {
	// Sizeof：类型占用的字节数（含对齐填充）
	fmt.Println("int8 大小:", unsafe.Sizeof(int8(0)))
	fmt.Println("int64 大小:", unsafe.Sizeof(int64(0)))

	// Alignof：类型的对齐要求（地址必须是该值的倍数）
	fmt.Println("int64 对齐:", unsafe.Alignof(int64(0)))

	var v inner
	// Offsetof：字段在结构体中的偏移量
	fmt.Println("inner 总大小:", unsafe.Sizeof(v))
	fmt.Println("字段 a 偏移:", unsafe.Offsetof(v.a))
	fmt.Println("字段 b 偏移:", unsafe.Offsetof(v.b))
	fmt.Println("字段 c 偏移:", unsafe.Offsetof(v.c))

	// 为什么 a(1B) 后面不是紧挨 b：
	// b 是 int64，需要 8 字节对齐，所以偏移 1~7 被填充，b 从偏移 8 开始。
	// 总大小 24 = a(1) + 填充(7) + b(8) + c(2) + 尾部填充(6)
	fmt.Println("结论：字段顺序影响内存占用，存在对齐填充")
}

// ===== 示例 3：调整字段顺序节省内存 =====
type compact struct {
	b int64 // 8 字节，先放大的
	c int16 // 2 字节
	a int8  // 1 字节
}

func example3Reorder() {
	fmt.Println("inner   大小:", unsafe.Sizeof(inner{}))
	fmt.Println("compact 大小:", unsafe.Sizeof(compact{}))
	// 为什么：把大字段放前面、按对齐要求从大到小排列，
	// 可以减少填充字节。compact 总大小 16 vs inner 的 24。
	// 在百万级对象切片中，这种优化能显著降低内存占用。
	fmt.Println("字段重排后节省了", unsafe.Sizeof(inner{})-unsafe.Sizeof(compact{}), "字节/对象")
}

// ===== 示例 4：string ↔ []byte 零拷贝转换（只读） =====
func example4ZeroCopy() {
	s := "hello unsafe"

	// 常规转换 []byte(s) 会复制一份数据（安全）；
	// 零拷贝写法直接复用 string 的底层数组：
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	fmt.Printf("零拷贝得到 []byte: %q，长度 %d\n", b, len(b))

	// 反向：[]byte -> string 零拷贝（同样需要保证原数组不被修改）
	raw := []byte("readonly")
	s2 := unsafe.String(unsafe.SliceData(raw), len(raw))
	fmt.Println("零拷贝得到 string:", s2)

	// ⚠️ 注意：零拷贝的 []byte 指向 string 的只读内存，
	//   绝对禁止修改（b[0] = 'X' 会直接崩溃或未定义行为）。
	// ⚠️ 注意：unsafe.String 的结果共享原 []byte 的内存，
	//   原切片被修改后 string 内容会"悄悄变化"，破坏 string 不可变约定。
	// ⚠️ 注意：普通业务代码应优先使用标准转换（[]byte(s) / string(b)），
	//   只有热点路径且经 benchmark 证明收益明显时才考虑零拷贝。

	_ = b  // 只读使用完毕
	_ = s2 // 只读使用完毕
}

// ===== 示例 5：unsafe 风险总结 =====
func example5RiskSummary() {
	fmt.Println("1) unsafe.Pointer 可绕过类型系统，编译器不再帮你检查类型安全")
	fmt.Println("2) uintptr 不阻止 GC 回收，长期持有再转回 unsafe.Pointer 是悬空指针")
	fmt.Println("3) 零拷贝共享底层内存，违反 string 不可变性会破坏依赖它的代码")
	fmt.Println("4) 不同平台对齐规则可能不同，手工内存布局假设会降低可移植性")
	fmt.Println("结论：业务代码避免 unsafe；仅限标准库级、经过充分测试的底层优化")
}
