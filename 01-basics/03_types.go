//go:build ignore

// ============================================================
// 文件：03_types.go
// 阶段：初级
// 主题：基本类型与类型转换
// 学习目标：
//   1. 认识 Go 的基本类型：int 系列、uint、float、bool、string
//   2. 理解 Go 没有隐式类型转换，必须显式转换
//   3. 区分 rune 与 byte
//   4. 区分类型别名与自定义类型，会用 strconv 做字符串转换
// 运行方式：go run 01-basics/03_types.go
// 预计阅读时间：15 分钟
// ============================================================

package main

import (
	"fmt"
	"strconv" // 标准库：字符串与基本类型之间的相互转换
)

func main() {
	fmt.Println("========== 示例 1：基本类型速览 ==========")
	example1BasicTypes()
	fmt.Println()

	fmt.Println("========== 示例 2：显式类型转换 ==========")
	example2ExplicitConversion()
	fmt.Println()

	fmt.Println("========== 示例 3：rune 与 byte ==========")
	example3RuneAndByte()
	fmt.Println()

	fmt.Println("========== 示例 4：类型别名 vs 自定义类型 ==========")
	example4AliasVsDefined()
	fmt.Println()

	fmt.Println("========== 示例 5：strconv 与错误处理 ==========")
	example5Strconv()
}

// ===== 示例 1：基本类型速览 =====
func example1BasicTypes() {
	// int 系列：int8/16/32/64 指定位数，int 随平台（64 位机器上即 64 位）。
	// 无符号 uint 系列只能表示非负数，范围是 signed 的正数部分翻倍。
	var a int = -10
	var b uint = 10
	// 浮点只有 float32 和 float64 两种，默认推断是 float64。
	var f = 3.14
	var ok bool = true
	// string 是内置类型，用双引号；不可变（不能改其中的某个字节）。
	var s string = "你好，Go"

	fmt.Printf("int: %d (%T)\n", a, a)
	fmt.Printf("uint: %d (%T)\n", b, b)
	fmt.Printf("float: %v (%T)\n", f, f)
	fmt.Printf("bool: %t\n", ok)
	fmt.Printf("string: %s，字节长度 len=%d\n", s, len(s))
	// ⚠️ 注意：len("你好，Go") 返回的是字节数（UTF-8 下中文每字 3 字节），
	// 不是字符个数！所以这里输出 11（3个中文字符 9 字节 + Go 2 字节）。
	// 选择建议：整数没特殊需求就用 int，浮点没特殊需求就用 float64。
}

// ===== 示例 2：显式类型转换 =====
func example2ExplicitConversion() {
	var i int = 42
	var f float64 = 3.99

	// ⚠️ 注意：Go 没有隐式类型转换，哪怕 int + float64 这样"显然"的
	// 运算也会编译报错。必须显式写出转换：T(v)。
	sum := float64(i) + f
	fmt.Printf("float64(i) + f = %.2f\n", sum)

	// 浮点转整数会"截断"小数部分（不是四舍五入）。
	truncated := int(f)
	fmt.Printf("int(%v) = %d（直接截断，不是四舍五入）\n", f, truncated)

	// ⚠️ 注意：大范围值转小类型会发生溢出截断，Go 不会报错。
	var big int64 = 300
	small := int8(big)
	fmt.Printf("int64(%d) 转 int8 得到 %d（溢出截断）\n", big, small)

	// 数值与 string 之间不能直接用 T(v) 互转出可打印内容：
	// string(65) 得到的是字符 "A"（码点转换），而不是 "65"。
	fmt.Printf("string(65) = %q（是字符 A，不是字符串 \"65\"）\n", string(65))
}

// ===== 示例 3：rune 与 byte =====
func example3RuneAndByte() {
	// byte 是 uint8 的别名，表示一个字节（0~255）。
	// rune 是 int32 的别名，表示一个 Unicode 码点（一个"字符"）。
	// 字符串本质上是字节序列；用下标取到的是 byte，不是字符。
	str := "Go语言"
	fmt.Printf("len(%q) = %d 字节\n", str, len(str))
	fmt.Printf("str[0] = %d（byte 类型 %T，即字母 G 的 ASCII 码）\n", str[0], str[0])

	// 按 byte 遍历：下标逐个走，中文会被拆成 3 个字节，看不懂。
	// 按 rune 遍历（for range）：自动按 UTF-8 解码，每次得到一个完整字符。
	fmt.Print("按 rune 遍历：")
	for i, r := range str {
		fmt.Printf("[%d]%c ", i, r)
	}
	fmt.Println()
	// ⚠️ 注意：range 的索引 i 是字节下标（0,1,4,7），不是字符序号。

	// 单引号字面量是 rune；双引号是 string，两者不能混用。
	var ch rune = '语'
	var by byte = 'G'
	fmt.Printf("rune '语' 的码点 = %d，byte 'G' = %d\n", ch, by)
}

// ===== 示例 4：类型别名 vs 自定义类型 =====
func example4AliasVsDefined() {
	// 类型别名（type A = B，有等号）：A 和 B 是"同一个类型"的两个名字，
	// 编译器完全视为一体，可以互相赋值、混合运算。
	type MyInt = int
	var a MyInt = 10
	var b int = 20
	fmt.Printf("别名：MyInt(%d) + int(%d) = %d，无需转换\n", a, b, a+b)
	fmt.Printf("a 的真实类型：%T（就是 int）\n", a)

	// 自定义类型（type A B，无等号）：A 是基于 B 的"全新类型"，
	// 与 B 之间不能直接赋值或运算，必须显式转换。
	// 这样做的价值：给相同底层数据赋予不同语义，让编译器帮你防错。
	type Celsius float64    // 摄氏度
	type Fahrenheit float64 // 华氏度
	var c Celsius = 100
	var f2 Fahrenheit = 212
	// c + f2 会编译报错：都是温度但度量不同，混算是 bug，类型系统拦住了。
	// 确实需要混算时，必须显式转换回底层类型，把"我确认要这么算"写在代码里。
	sum := float64(c) + float64(f2)
	fmt.Printf("显式转换后才能相加：float64(%v) + float64(%v) = %v\n", c, f2, sum)
	// ⚠️ 注意：等号差一个字符，语义天壤之别。别名是"改名"，自定义是"新类型"。
}

// ===== 示例 5：strconv 与错误处理 =====
func example5Strconv() {
	// Itoa：int -> string（Itoa = Integer to ASCII）。
	s := strconv.Itoa(123)
	fmt.Printf("Itoa(123) = %q\n", s)

	// Atoi：string -> int，返回两个值：结果 + error。
	// Go 的错误处理风格：不抛异常，而是把 error 作为返回值显式检查。
	num, err := strconv.Atoi("456")
	if err != nil {
		fmt.Println("转换失败：", err)
	} else {
		fmt.Printf("Atoi(\"456\") = %d\n", num)
	}

	// ⚠️ 注意：忽略 err 是新手最常见的 bug 之一。
	// 转换失败时 num 是零值 0，如果不检查 err，0 会被当成正常结果用。
	bad, err := strconv.Atoi("不是数字")
	if err != nil {
		fmt.Printf("Atoi(\"不是数字\") 出错：%v（此时结果为 %d）\n", err, bad)
	}

	// ParseInt：更通用的解析，参数依次是：字符串、进制、位宽。
	// 下面按 16 进制解析 "ff"，位宽 64。
	hex, err := strconv.ParseInt("ff", 16, 64)
	if err == nil {
		fmt.Printf("ParseInt(\"ff\", 16, 64) = %d\n", hex)
	}

	// 反向：FormatInt 按指定进制把整数格式化为字符串。
	fmt.Printf("FormatInt(255, 16) = %q\n", strconv.FormatInt(255, 16))
	// 还有 ParseFloat / ParseBool / FormatFloat 等，用法相同。
	fv, err := strconv.ParseFloat("3.14", 64)
	if err == nil {
		fmt.Printf("ParseFloat(\"3.14\") = %v\n", fv)
	}
}
