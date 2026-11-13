//go:build ignore

// ============================================================
// 文件：04_control.go
// 阶段：初级
// 主题：流程控制
// 学习目标：
//   1. 掌握 if 及其初始化语句写法
//   2. 掌握 for 的三种写法与 range 遍历
//   3. 掌握 switch 的三种用法与 fallthrough
//   4. 会用 label 配合 break/continue 跳出多层循环
// 运行方式：go run 01-basics/04_control.go
// 预计阅读时间：15 分钟
// ============================================================

package main

import "fmt"

func main() {
	fmt.Println("========== 示例 1：if 与初始化语句 ==========")
	example1If()
	fmt.Println()

	fmt.Println("========== 示例 2：for 的三种写法 ==========")
	example2For()
	fmt.Println()

	fmt.Println("========== 示例 3：switch 三种用法 ==========")
	example3Switch()
	fmt.Println()

	fmt.Println("========== 示例 4：range 遍历 ==========")
	example4Range()
	fmt.Println()

	fmt.Println("========== 示例 5：label 跳出多层循环 ==========")
	example5Label()
}

// ===== 示例 1：if 与初始化语句 =====
func example1If() {
	score := 85
	// Go 的 if 不需要括号，但花括号必须写（即使只有一行）。
	if score >= 60 {
		fmt.Println("及格：", score)
	} else {
		fmt.Println("不及格：", score)
	}

	// if 支持"初始化语句"：在条件前声明一个只在 if/else 内有效的变量。
	// 好处是变量作用域被限制在分支里，不污染外层代码。
	if remainder := score % 10; remainder == 0 {
		fmt.Println("个位数是 0")
	} else {
		fmt.Printf("个位数是 %d\n", remainder)
	}
	// ⚠️ 注意：此处 remainder 已不可访问，它在 if 块结束时销毁。

	// ⚠️ 注意：Go 的 if 条件必须是 bool，不能用 0/非 0 整数当条件，
	// 这与 C 不同。例如写 if 1 {} 会直接编译报错。
}

// ===== 示例 2：for 的三种写法 =====
func example2For() {
	// Go 只有 for 一种循环关键字，没有 while / do-while，
	// 但 for 可以变化出三种写法覆盖所有循环场景。

	// 写法一：经典三段式（初始化；条件；后置语句）。
	fmt.Print("三段式：")
	for i := 0; i < 3; i++ {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// 写法二：条件式，相当于其他语言的 while。
	fmt.Print("条件式（类似 while）：")
	n := 1
	for n < 10 {
		fmt.Printf("%d ", n)
		n *= 3 // 1, 3, 9
	}
	fmt.Println()

	// 写法三：无限循环 for {}，用 break 主动退出。
	// 常用于"循环等待某个条件达成"的场景。
	fmt.Print("无限循环 + break：")
	count := 0
	for {
		count++
		if count >= 3 {
			break // break 立即结束整个循环
		}
		fmt.Printf("第%d轮 ", count)
	}
	fmt.Println("（第3轮时 break 退出）")
	// ⚠️ 注意：for {} 里忘记 break 就是死循环，程序永远不会退出。
}

// ===== 示例 3：switch 三种用法 =====
func example3Switch() {
	// 用法一：表达式 switch，与其他语言类似。
	day := 3
	switch day {
	case 1, 2, 3, 4, 5: // 一个 case 可以列多个值
		fmt.Println("工作日")
	case 6, 7:
		fmt.Println("周末")
	default: // 都不匹配时执行，可省略
		fmt.Println("非法的日期")
	}
	// ⚠️ 注意：Go 的 case 默认自动 break，不会像 C 那样"穿透"到下一个 case。

	// 用法二：无表达式 switch，case 里直接写布尔条件。
	// 等价于一串 if-else if，但可读性更好。
	score := 72
	switch {
	case score >= 90:
		fmt.Println("优秀")
	case score >= 60:
		fmt.Println("及格")
	default:
		fmt.Println("不及格")
	}

	// 用法三：fallthrough 强制穿透到下一个 case。
	// 它无条件执行下一个 case 的语句体（不看下一个 case 的条件）。
	switch v := 2; v {
	case 1:
		fmt.Println("一")
	case 2:
		fmt.Println("二（从这里 fallthrough）")
		fallthrough
	case 3:
		fmt.Println("三（被 fallthrough 强制执行）")
	}
	// ⚠️ 注意：fallthrough 很少用，见到时要知道它会跳过条件判断。
}

// ===== 示例 4：range 遍历 =====
func example4Range() {
	// range 用于遍历数组/切片/map/字符串/通道，返回"索引和值"。
	fruits := []string{"苹果", "香蕉", "橙子"}
	for i, fruit := range fruits {
		fmt.Printf("索引 %d -> %s\n", i, fruit)
	}

	// 只要值不要索引：用空白标识符 _ 丢弃索引。
	// ⚠️ 注意：声明了 i 却不用会编译报错，所以必须用 _ 显式丢弃。
	fmt.Print("只要值：")
	for _, fruit := range fruits {
		fmt.Print(fruit, " ")
	}
	fmt.Println()

	// 遍历 map：每次拿到 key 和 value。
	ages := map[string]int{"小明": 18, "小红": 20}
	for name, age := range ages {
		fmt.Printf("%s 今年 %d 岁\n", name, age)
	}
	// 提示：输出顺序可能不同——map 的遍历顺序在 Go 中是故意随机的，
	// 防止程序依赖遍历顺序。需要有序时请先把 key 排序。

	// 遍历字符串得到 rune（上节讲过，这里温习）。
	for i, r := range "Go" {
		fmt.Printf("字节下标 %d -> 字符 %c\n", i, r)
	}
}

// ===== 示例 5：label 跳出多层循环 =====
func example5Label() {
	// 普通 break 只能跳出"最内层"循环。
	// 想从嵌套循环里直接跳到最外层，需要给外层循环贴一个 label。
	fmt.Println("--- break 带 label：找到目标后整体退出 ---")
OuterLoop: // label 定义：名字 + 冒号，独占一行
	for i := 1; i <= 3; i++ {
		for j := 1; j <= 3; j++ {
			fmt.Printf("i=%d j=%d\n", i, j)
			if i*j >= 4 {
				break OuterLoop // 直接终止外层循环，不是内层
			}
		}
	}
	fmt.Println("已跳出外层循环")

	// continue 也支持 label：跳过外层循环的当前轮，进入下一轮。
	fmt.Println("--- continue 带 label：跳过外层当前轮 ---")
ContinueLoop:
	for i := 1; i <= 3; i++ {
		for j := 1; j <= 3; j++ {
			if j == 2 {
				continue ContinueLoop // 直接进入外层的下一轮 i
			}
			fmt.Printf("打印：i=%d j=%d\n", i, j)
		}
	}
	// ⚠️ 注意：label 虽强大，但滥用会让流程难以追踪（类似 goto），
	// 只在"多层循环确实需要整体退出"时使用。
}
