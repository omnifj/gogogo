//go:build ignore

// ============================================================
// 文件：06_arrays_slices.go
// 阶段：初级
// 主题：数组与切片
// 学习目标：
//   1. 理解数组是值类型，赋值和传参都会发生拷贝
//   2. 理解切片 = 指针 + 长度 + 容量 三要素
//   3. 掌握 append 的扩容行为与 copy 的用法
//   4. 认识共享底层数组带来的“坑”
// 运行方式：go run 01-basics/06_arrays_slices.go
// 预计阅读时间：12 分钟
// ============================================================

package main

import "fmt"

func main() {
	fmt.Println("========== 示例 1：数组是值类型 ==========")
	arrayValueSemantics()
	fmt.Println()

	fmt.Println("========== 示例 2：切片三要素 ==========")
	sliceThreeFactors()
	fmt.Println()

	fmt.Println("========== 示例 3：append 扩容行为 ==========")
	appendGrowth()
	fmt.Println()

	fmt.Println("========== 示例 4：copy 的使用 ==========")
	copyDemo()
	fmt.Println()

	fmt.Println("========== 示例 5：共享底层数组的坑 ==========")
	sharedArrayTrap()
}

// ===== 示例 1：数组是值类型 =====
func arrayValueSemantics() {
	a := [3]int{1, 2, 3}
	b := a // 数组赋值会整体拷贝，b 是一份独立副本
	b[0] = 100
	fmt.Println("a:", a) // a 不受影响，仍然是 {1, 2, 3}
	fmt.Println("b:", b)

	modifyArray(a)             // 传参同样发生拷贝
	fmt.Println("函数修改后 a:", a) // 原数组不变

	modifyArrayByPtr(&a) // 传指针才能真正修改原数组
	fmt.Println("指针修改后 a:", a)
}

func modifyArray(arr [3]int) {
	arr[0] = 999 // 只改了副本，调用方看不到
}

func modifyArrayByPtr(arr *[3]int) {
	arr[0] = 999 // 通过指针修改原数组
}

// ===== 示例 2：切片三要素 =====
func sliceThreeFactors() {
	// 切片内部结构：指向底层数组的指针 + 长度 len + 容量 cap
	s := []int{10, 20, 30, 40}
	fmt.Printf("s = %v, len = %d, cap = %d\n", s, len(s), cap(s))

	// 长度是“当前可见元素个数”，容量是“底层数组从起始位置起还能容纳多少”
	sub := s[1:3]
	fmt.Printf("sub = %v, len = %d, cap = %d\n", sub, len(sub), cap(sub))

	// 用 make 创建切片时可以指定长度和容量，避免频繁扩容
	buf := make([]int, 0, 8)
	fmt.Printf("buf = %v, len = %d, cap = %d\n", buf, len(buf), cap(buf))
}

// ===== 示例 3：append 扩容行为 =====
func appendGrowth() {
	s := make([]int, 0, 2) // 容量故意给得很小，便于观察扩容
	fmt.Printf("初始：len = %d, cap = %d\n", len(s), cap(s))

	oldCap := cap(s)
	for i := 0; i < 5; i++ {
		s = append(s, i)
		if cap(s) != oldCap {
			// 容量不足时 append 会申请一块更大的新底层数组，旧数组被丢弃
			fmt.Printf("追加 %d 后扩容：cap %d -> %d\n", i, oldCap, cap(s))
			oldCap = cap(s)
		}
	}
	fmt.Println("最终 s:", s)

	// ⚠️ 注意：append 返回的可能是新数组的切片，必须把返回值赋回变量
	// 错误写法：append(s, 6)  // 结果丢失，s 没有变化
}

// ===== 示例 4：copy 的使用 =====
func copyDemo() {
	src := []int{1, 2, 3, 4, 5}
	dst := make([]int, len(src))

	// copy 按“两个切片中较短的长度”复制元素，返回值是实际复制的个数
	n := copy(dst, src)
	fmt.Printf("复制了 %d 个元素，dst = %v\n", n, dst)

	dst[0] = 100 // copy 后是两块独立内存，互不影响
	fmt.Println("修改 dst 后 src:", src)

	// copy 也常用于切片元素插入/删除时移动数据，例如删除下标为 1 的元素
	s := []int{1, 2, 3, 4}
	copy(s[1:], s[2:]) // 把后面的元素整体前移一位
	s = s[:len(s)-1]   // 长度减一，丢弃尾部重复元素
	fmt.Println("删除下标 1 之后:", s)
}

// ===== 示例 5：共享底层数组的坑 =====
func sharedArrayTrap() {
	original := []int{1, 2, 3, 4, 5}
	part := original[1:3] // part 与 original 共享同一个底层数组

	part[0] = 200 // ⚠️ 注意：修改子切片，会影响原切片
	fmt.Println("original:", original)
	fmt.Println("part:", part)

	// ⚠️ 注意：更隐蔽的坑——part 容量是 4（底层数组从起始位置还有空间），
	// 前两次 append 不扩容，会直接覆盖 original 中下标 3、4 的元素！
	part = append(part, 999)
	part = append(part, 888)
	fmt.Println("容量内 append 后 original:", original) // 下标 3、4 被悄悄改写

	// 继续 append，当 len 超过容量 4 时才扩容换到新数组，此后与原数组彻底分离
	part = append(part, 777) // len=5 > cap=4，这一步触发扩容
	part = append(part, 666)
	part[0] = -1
	fmt.Println("扩容后 part:", part)
	fmt.Println("扩容后 original:", original) // part[0] 的修改不再影响 original

	// 防御手段一：copy 一份真正独立的子切片
	independent := make([]int, 2)
	copy(independent, original[1:3])
	independent[0] = 0
	fmt.Println("independent:", independent)

	// 防御手段二：三索引切片 original[1:3:3]，显式把容量限制为 2，
	// 这样第一次 append 就会扩容，永远不会写回 original
	safe := original[1:3:3]
	safe = append(safe, 555)
	fmt.Println("三索引切片 append 后 original:", original)
}
