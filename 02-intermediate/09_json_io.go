//go:build ignore

// ============================================================
// 文件：09_json_io.go
// 阶段：中级
// 主题：JSON 编解码与文件 IO
// 学习目标：
//   1. 掌握 json.Marshal / Unmarshal 与 struct tag 的用法
//   2. 用 os.WriteFile / os.ReadFile 读写文件
//   3. 理解 io.Reader / io.Writer 接口思想
//   4. 用 bufio.Scanner 按行读取文本
// 运行方式：go run 02-intermediate/09_json_io.go
// 预计阅读时间：16 分钟
// ============================================================

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// ===== 示例 1：json.Marshal / Unmarshal 基础 =====

// User 的字段必须导出（大写开头），否则 encoding/json 看不到
type User struct {
	Name  string `json:"name"`          // tag 改名：JSON 里叫 name 而不是 Name
	Age   int    `json:"age,omitempty"` // omitempty：零值时不输出该字段
	Email string `json:"-"`             // "-"：序列化时彻底忽略（敏感字段常用）
	Note  string // 无 tag：默认使用字段名 Note
}

func marshalDemo() {
	u := User{Name: "gopher", Age: 0, Email: "secret@x.io", Note: "hi"}
	data, err := json.Marshal(u)
	check(err)
	fmt.Println("Marshal 结果:", string(data))
	// Age 是零值被 omitempty 省略；Email 被 "-" 忽略

	var back User
	err = json.Unmarshal(data, &back) // 必须传指针，否则无法写回
	check(err)
	fmt.Printf("Unmarshal 回来: %+v\n", back)
	// ⚠️ 注意：JSON 中的未知字段默认被忽略，多出的字段不会报错；
	// 需要严格校验时改用 Decoder.DisallowUnknownFields()。
}

// ===== 示例 2：os.WriteFile / os.ReadFile 读写文件 =====
func fileIODemo() string {
	dir, err := os.MkdirTemp("", "golearn") // 系统临时目录下建独立子目录
	check(err)

	path := dir + string(os.PathSeparator) + "note.txt"
	content := "第一行\n第二行\n第三行\n"
	err = os.WriteFile(path, []byte(content), 0o644) // 0644：所有者可读写，他人只读
	check(err)
	fmt.Println("已写入临时文件:", path)

	raw, err := os.ReadFile(path)
	check(err)
	fmt.Println("读取字节数:", len(raw), "首行:", strings.Split(string(raw), "\n")[0])
	return dir // 返回目录路径，交给 main 统一 defer 清理
	// ⚠️ 注意：os.WriteFile 会直接覆盖已存在文件；
	// 需要追加或流式写入时请改用 os.OpenFile + 标志位。
}

// ===== 示例 3：io.Reader / io.Writer 接口思想 =====
func readerWriterDemo() {
	// strings.NewReader、os.File、网络连接都实现 io.Reader，
	// 因此同一个函数可以处理完全不同的数据来源——这就是接口的价值
	src := strings.NewReader("hello io interface")
	var buf bytes.Buffer         // bytes.Buffer 实现 io.Writer
	n, err := io.Copy(&buf, src) // Copy 不关心两端具体类型
	check(err)
	fmt.Printf("io.Copy 复制了 %d 字节: %q\n", n, buf.String())

	// json.NewEncoder 直接往任何 io.Writer 写，不必先 Marshal 成 []byte
	var out bytes.Buffer
	err = json.NewEncoder(&out).Encode(User{Name: "encoder", Age: 3})
	check(err)
	fmt.Println("Encoder 直写 Writer:", strings.TrimSpace(out.String()))
}

// ===== 示例 4：bufio.Scanner 按行读取 =====
func scannerDemo() {
	text := "apple\nbanana\ncherry\n"
	// Scanner 默认按行切分，也可换成 ScanWords 等策略
	sc := bufio.NewScanner(strings.NewReader(text))
	lineNo := 0
	for sc.Scan() { // 每读一行返回 true，读到 EOF 返回 false
		lineNo++
		fmt.Printf("  第 %d 行: %s\n", lineNo, sc.Text())
	}
	check(sc.Err()) // ⚠️ 注意：循环结束后必须检查 Err，EOF 之外的错误藏在这里
	// ⚠️ 注意：单行超过 64KB 会被默认缓冲区截断并报错，
	// 超长行要用 sc.Buffer() 扩容或改用 bufio.Reader.ReadString。
}

// check 是教学用的简化错误处理；真实项目应逐处判断并包装错误
func check(err error) {
	if err != nil {
		fmt.Println("出错:", err)
		os.Exit(1)
	}
}

func main() {
	fmt.Println("========== 示例 1：json.Marshal / Unmarshal 与 struct tag ==========")
	marshalDemo()
	fmt.Println()

	fmt.Println("========== 示例 2：os.WriteFile / os.ReadFile ==========")
	tmpDir := fileIODemo()
	defer os.RemoveAll(tmpDir) // 程序结束前清理临时目录
	fmt.Println("（临时目录将在 main 返回时自动清理）")
	fmt.Println()

	fmt.Println("========== 示例 3：io.Reader / io.Writer 接口思想 ==========")
	readerWriterDemo()
	fmt.Println()

	fmt.Println("========== 示例 4：bufio.Scanner 按行读取 ==========")
	scannerDemo()
}
