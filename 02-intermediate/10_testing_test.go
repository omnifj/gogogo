//go:build ignore

// ============================================================
// 文件：10_testing_test.go
// 阶段：中级
// 主题：表驱动测试、子测试、基准测试、示例测试
// 学习目标：
//   1. 掌握表驱动测试的组织方式
//   2. 用 t.Run 编写命名子测试
//   3. 编写 Benchmark 基准测试与 Example 示例测试
// 运行方式：go test -v 10_testing.go 10_testing_test.go
// 预计阅读时间：18 分钟
// ============================================================

package main

import (
	"fmt"
	"testing"
)

// ===== 示例 1：表驱动测试 =====
// 把多组输入/期望放进切片，一条循环全部跑完，是 Go 测试的主流写法
func TestAdd(t *testing.T) {
	cases := []struct {
		name string // 用例名，失败时便于定位
		a, b int
		want int
	}{
		{"正数相加", 2, 3, 5},
		{"负数相加", -1, -1, -2},
		{"正负相抵", 5, -5, 0},
		{"加零不变", 42, 0, 42},
	}
	for _, c := range cases {
		// t.Run 生成子测试，go test -run TestAdd/正数相加 可单独运行
		t.Run(c.name, func(t *testing.T) {
			if got := Add(c.a, c.b); got != c.want {
				// Errorf 记录失败但继续跑后续用例；Fatalf 会立即中止本测试
				t.Errorf("Add(%d, %d) = %d, 期望 %d", c.a, c.b, got, c.want)
			}
		})
	}
}

// ===== 示例 2：测试返回 error 的函数 =====
func TestDivide(t *testing.T) {
	t.Run("正常除法", func(t *testing.T) {
		got, err := Divide(10, 4)
		if err != nil {
			t.Fatalf("意外错误: %v", err)
		}
		if got != 2.5 {
			t.Errorf("Divide(10, 4) = %v, 期望 2.5", got)
		}
	})
	t.Run("除零报错", func(t *testing.T) {
		// ⚠️ 注意：测试错误路径时重点检查 err 是否为 nil，而不是错误文案
		if _, err := Divide(1, 0); err == nil {
			t.Error("Divide(1, 0) 应返回错误，却得到 nil")
		}
	})
}

// ===== 示例 3：回文判断的表驱动测试 =====
func TestIsPalindrome(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"level", true},
		{"上海海上", true}, // 多字节字符：验证 []rune 实现的正确性
		{"golang", false},
		{"", true},  // 空串是回文（边界用例）
		{"a", true}, // 单字符是回文（边界用例）
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("input=%q", c.input), func(t *testing.T) {
			if got := IsPalindrome(c.input); got != c.want {
				t.Errorf("IsPalindrome(%q) = %v, 期望 %v", c.input, got, c.want)
			}
		})
	}
}

// ===== 示例 4：单词计数测试 =====
func TestWordCount(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  int
	}{
		{"普通句子", "go is fun", 3},
		{"连续空白", "  a  b ", 2},
		{"空字符串", "", 0},
		{"含换行符", "one\ntwo", 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := WordCount(c.input); got != c.want {
				t.Errorf("WordCount(%q) = %d, 期望 %d", c.input, got, c.want)
			}
		})
	}
}

// ===== 示例 5：基准测试 =====
// Benchmark 前缀 + b.N 循环：框架会自动调整 N 以获得稳定的耗时数据
// 运行：go test -bench=. -benchtime=10x 10_testing.go 10_testing_test.go
func BenchmarkIsPalindrome(b *testing.B) {
	s := "上海海上"
	for i := 0; i < b.N; i++ {
		IsPalindrome(s)
	}
}

func BenchmarkWordCount(b *testing.B) {
	s := "the quick brown fox jumps over the lazy dog"
	for i := 0; i < b.N; i++ {
		WordCount(s)
	}
}

// ===== 示例 6：示例测试 =====
// Example 前缀的函数既是可以运行验证的测试，也是会出现在文档里的示例
// 末尾的 "// Output:" 注释与程序实际输出必须完全一致，否则测试失败
func ExampleAdd() {
	fmt.Println(Add(2, 3))
	// Output:
	// 5
}

func ExampleIsPalindrome() {
	fmt.Println(IsPalindrome("level"))
	fmt.Println(IsPalindrome("golang"))
	// Output:
	// true
	// false
}

// ⚠️ 注意：Example 没有 Output 注释时只编译不运行；
// 输出里有多余的空格或换行都会导致比较失败。
