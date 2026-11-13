//go:build ignore

// ============================================================
// 文件：10_strings_runes.go
// 阶段：初级
// 主题：字符串与 rune
// 学习目标：
//   1. 理解字符串不可变，以及 byte 与 rune 的区别
//   2. 理解 UTF-8 下 len 返回字节数而非字符数
//   3. 掌握按字节遍历与 range 按 rune 遍历的区别
//   4. 掌握 strings 包常用函数与 strings.Builder 高效拼接
// 运行方式：go run 01-basics/10_strings_runes.go
// 预计阅读时间：13 分钟
// ============================================================

package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("========== 示例 1：字符串不可变与 byte vs rune ==========")
	immutableAndRune()
	fmt.Println()

	fmt.Println("========== 示例 2：len 是字节数 ==========")
	lenIsBytes()
	fmt.Println()

	fmt.Println("========== 示例 3：按字节遍历 vs 按 rune 遍历 ==========")
	byteVsRuneIteration()
	fmt.Println()

	fmt.Println("========== 示例 4：strings 包常用函数 ==========")
	stringsPackage()
	fmt.Println()

	fmt.Println("========== 示例 5：strings.Builder 高效拼接 ==========")
	builderDemo()
	fmt.Println()

	fmt.Println("========== 示例 6：fmt 格式化动词复习 ==========")
	fmtVerbs()
}

// ===== 示例 1：字符串不可变与 byte vs rune =====
func immutableAndRune() {
	s := "hello"
	// ⚠️ 注意：Go 字符串不可变，下面这行无法编译：
	//     s[0] = 'H' // 编译错误：cannot assign to s[0]
	// 想“修改”字符串，要先转成 []rune（或 []byte），改完再转回 string
	runes := []rune(s)
	runes[0] = 'H' // 单引号是 rune 字面量（一个字符）
	s = string(runes)
	fmt.Println("转换后修改:", s)

	// byte 是 uint8 的别名，表示一个字节；rune 是 int32 的别名，表示一个 Unicode 码点
	var b byte = 'A'
	var r rune = '中'
	fmt.Printf("byte: %c（%d），rune: %c（%d）\n", b, b, r, r)

	// 处理含中文的字符串时优先用 []rune，否则一个汉字会被拆成多个字节
	text := []rune("你好")
	text[0] = '您'
	fmt.Println("rune 切片安全替换:", string(text))
}

// ===== 示例 2：len 是字节数 =====
func lenIsBytes() {
	en := "abc"
	cn := "你好"

	// UTF-8 编码下英文字母占 1 字节，一个汉字通常占 3 字节
	fmt.Printf("len(%q) = %d 字节\n", en, len(en))
	fmt.Printf("len(%q) = %d 字节（不是 2 个字符！）\n", cn, len(cn))

	// 想要“字符数”，用 utf8.RuneCountInString 或 len([]rune(s))
	fmt.Printf("%q 的字符数: %d\n", cn, len([]rune(cn)))

	// ⚠️ 注意：按字符数截断、限制输入长度等场景，
	// 直接用 len 会把多字节字符算错，务必换算成 rune 数量。
}

// ===== 示例 3：按字节遍历 vs 按 rune 遍历 =====
func byteVsRuneIteration() {
	s := "Go语言"

	fmt.Println("按字节遍历（for + 下标）:")
	for i := 0; i < len(s); i++ {
		// 汉字被拆成 3 个字节，单独看是无意义的数字
		fmt.Printf("  下标 %d: 字节值 %d\n", i, s[i])
	}

	fmt.Println("按 rune 遍历（range）:")
	for i, r := range s {
		// range 自动按 UTF-8 解码，r 是完整字符；i 是该字符的起始字节下标（会跳跃）
		fmt.Printf("  起始字节 %d: 字符 %c\n", i, r)
	}

	// ⚠️ 注意：range 遍历时下标不是连续的 0,1,2,3...，
	// 需要“第几个字符”时请自己计数，或先转成 []rune。
	n := 0
	for range s {
		n++
	}
	fmt.Println("用 range 计数得到字符数:", n)
}

// ===== 示例 4：strings 包常用函数 =====
func stringsPackage() {
	s := "  Go,Java,Go,Python  "

	fmt.Println("Trim 去首尾空白:", "["+strings.TrimSpace(s)+"]")

	trimmed := strings.TrimSpace(s)
	fmt.Println("Contains 包含子串:", strings.Contains(trimmed, "Java"))
	fmt.Println("HasPrefix 前缀:", strings.HasPrefix(trimmed, "Go"))
	fmt.Println("Count 统计出现次数:", strings.Count(trimmed, "Go"))

	parts := strings.Split(trimmed, ",") // 按逗号切分为切片
	fmt.Println("Split 结果:", parts)

	joined := strings.Join(parts, " | ") // 切片拼接回字符串
	fmt.Println("Join 结果:", joined)

	fmt.Println("Replace 替换:", strings.Replace(trimmed, "Go", "Golang", 1)) // 只替换第 1 个
	fmt.Println("ReplaceAll 全替换:", strings.ReplaceAll(trimmed, "Go", "Golang"))

	fields := strings.Fields("  a  b\tc\n d ") // 按任意空白切分，自动忽略连续空白
	fmt.Println("Fields 结果:", fields)

	fmt.Println("ToUpper/ToLower:", strings.ToUpper("abc"), strings.ToLower("ABC"))
}

// ===== 示例 5：strings.Builder 高效拼接 =====
func builderDemo() {
	words := []string{"Go", "是", "一门", "高效", "的", "语言"}

	// ⚠️ 注意：循环里用 s += x 拼接，每次都会产生新字符串（旧内容被拷贝），
	// 数据量大时非常浪费。strings.Builder 内部维护可增长缓冲区，零拷贝。

	var sb strings.Builder
	// 可选：预估容量，减少内部扩容次数
	sb.Grow(64)
	for _, w := range words {
		sb.WriteString(w)
	}
	fmt.Println("Builder 拼接:", sb.String())

	// Builder 也支持写入 rune、byte
	sb.Reset() // 重置后复用
	sb.WriteString("字符: ")
	sb.WriteRune('中')
	fmt.Println(sb.String())
}

// ===== 示例 6：fmt 格式化动词复习 =====
func fmtVerbs() {
	n := 42
	s := "你好"
	f := 3.14159

	fmt.Printf("%%d 十进制整数: %d\n", n)
	fmt.Printf("%%s 字符串: %s\n", s)
	fmt.Printf("%%v 任意值的默认格式: %v, %v, %v\n", n, f, []int{1, 2})
	fmt.Printf("%%T 值的类型: %T, %T\n", n, s)
	fmt.Printf("%%q 带引号的字符串: %q\n", s)
	fmt.Printf("%%x 十六进制: %x（数字）/ %x（字符串的字节）\n", n, s)

	// %q 与 %x 在调试日志中特别有用：%q 能暴露不可见字符，%x 能看原始字节
	fmt.Printf("含换行的字符串 %%q 效果: %q\n", "a\nb")
}
