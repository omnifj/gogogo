//go:build ignore

// ============================================================
// 文件：10_testing.go
// 阶段：中级
// 主题：go test 测试入门（被测代码 + main 演示）
// 学习目标：
//   1. 认识被测函数与测试文件的组织方式
//   2. 了解表驱动测试、子测试、基准测试、示例测试
//   3. 掌握配套测试文件的运行命令
// 运行方式：go run 02-intermediate/10_testing.go
// 预计阅读时间：15 分钟
// ============================================================

package main

import (
	"errors"
	"fmt"
	"strings"
)

// ============================================================
// 被测代码区：下面的函数是 10_testing_test.go 的测试对象
// ============================================================

// Add 返回两数之和，是最基础的被测函数
func Add(a, b int) int {
	return a + b
}

// Divide 演示"返回 error"的函数如何测试：除数为 0 时返回错误
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("除数不能为 0")
	}
	return a / b, nil
}

// IsPalindrome 判断字符串是否为回文（逐字符对称，区分大小写与空格）
// 用 []rune 处理，保证中文等多字节字符也能正确判断
func IsPalindrome(s string) bool {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		if r[i] != r[j] {
			return false
		}
	}
	return true
}

// WordCount 统计文本中的单词数（按空白分隔），空串返回 0
// strings.Fields 会自动处理连续空格、制表符和换行
func WordCount(s string) int {
	return len(strings.Fields(s))
}

// ============================================================
// main 演示区
// ============================================================

func main() {
	fmt.Println("========== 示例 1：Add 加法 ==========")
	fmt.Println("Add(2, 3) =", Add(2, 3))
	fmt.Println("Add(-1, 1) =", Add(-1, 1))
	fmt.Println()

	fmt.Println("========== 示例 2：Divide 带错误的除法 ==========")
	if q, err := Divide(10, 4); err == nil {
		fmt.Println("Divide(10, 4) =", q)
	}
	if _, err := Divide(1, 0); err != nil {
		fmt.Println("Divide(1, 0) 返回错误:", err)
	}
	fmt.Println()

	fmt.Println("========== 示例 3：IsPalindrome 回文判断 ==========")
	for _, s := range []string{"level", "上海海上", "golang"} {
		fmt.Printf("IsPalindrome(%q) = %v\n", s, IsPalindrome(s))
	}
	fmt.Println()

	fmt.Println("========== 示例 4：WordCount 单词计数 ==========")
	fmt.Println("WordCount(\"go is fun\") =", WordCount("go is fun"))
	fmt.Println("WordCount(\"  a  b \") =", WordCount("  a  b "))
	fmt.Println("WordCount(\"\") =", WordCount(""))
	fmt.Println("WordCount(\"one\\ntwo\") =", WordCount("one\ntwo"))
	fmt.Println()

	fmt.Println("========== 示例 5：为什么需要测试 ==========")
	// 以 Divide 为例：手写几组输入逐一验证很不划算，
	// 测试文件可以把这些验证固化下来，每次改动后一键回归
	checks := []struct {
		a, b float64
		want float64
	}{
		{10, 4, 2.5},
		{9, 3, 3},
		{1, 2, 0.5},
	}
	pass := 0
	for _, c := range checks {
		if got, err := Divide(c.a, c.b); err == nil && got == c.want {
			pass++
		}
	}
	fmt.Printf("手工模拟回归: %d/%d 通过（真正的回归请交给 go test）\n", pass, len(checks))
	fmt.Println()

	fmt.Println("========== 如何运行配套测试 ==========")
	// 配套测试文件是 10_testing_test.go（用 //go:build ignore 标记，
	// 因此要用显式文件列表的方式运行）：
	//   cd 02-intermediate
	//   go test -v 10_testing.go 10_testing_test.go
	// 只跑某个测试：        go test -run TestIsPalindrome -v 10_testing.go 10_testing_test.go
	// 跑基准测试：          go test -bench=. -benchtime=10x 10_testing.go 10_testing_test.go
	// ⚠️ 注意：真实的 Go 项目里测试文件不加 ignore 标记，直接 go test ./... 即可；
	// 本文件加 ignore 是为了让 go run 单文件教学示例能独立运行。
	fmt.Println("测试命令详见源码注释")
}
