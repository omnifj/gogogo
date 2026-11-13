//go:build ignore

// ============================================================
// 文件：07_embed_generate.go
// 阶段：高级
// 主题：go:embed、go:generate 与 build tags
// 学习目标：
//   1. 掌握 //go:embed 嵌入字符串、字节切片与 embed.FS
//   2. 理解 //go:generate 指令的作用与使用方式
//   3. 理解 build tags（//go:build）实现条件编译
// 运行方式：go run 03-advanced/07_embed_generate.go
// 预计阅读时间：15 分钟
// ============================================================

package main

import (
	"embed"
	"fmt"
)

// ===== 示例 1 用到的嵌入变量 =====

//go:embed assets/message.txt
var message string

//go:embed assets/message.txt
var messageBytes []byte

//go:embed assets
var assetsFS embed.FS

// ===== 示例 2 讲解 //go:generate =====
//
//go:generate gofmt -w $GOFILE
//
// 上面这一行是 //go:generate 指令：它告诉 `go generate` 命令要执行什么。
//   - $GOFILE 是当前文件名，此外还有 $GOPACKAGE、$DOLLAR 等内置变量。
//   - //go:generate 只是"登记"命令，不会在 go build / go run 时自动执行；
//     必须手动运行 `go generate ./...` 才会真正触发。
//   - 实际工程中它常用于代码生成，例如：
//       //go:generate stringer -type=Color        （生成 String() 方法）
//       //go:generate mockgen -source=foo.go      （生成测试 mock）
//   - 本课程登记的是一个无害命令（格式化本文件），并且不实际执行，
//     仅作为语法示例。
//
// ⚠️ 注意：//go:generate 后面不能有空行紧挨 package 声明，且必须顶格书写。

func main() {
	fmt.Println("========== 示例 1：//go:embed 嵌入资源 ==========")
	example1Embed()

	fmt.Println()
	fmt.Println("========== 示例 2：embed.FS 当作只读文件系统 ==========")
	example2EmbedFS()

	fmt.Println()
	fmt.Println("========== 示例 3：//go:generate 指令说明 ==========")
	example3Generate()

	fmt.Println()
	fmt.Println("========== 示例 4：build tags 条件编译 ==========")
	example4BuildTags()
}

// ===== 示例 1：//go:embed 嵌入资源 =====
func example1Embed() {
	// 为什么用 embed：资源在【编译期】被嵌入可执行文件，
	// 程序发布后无需携带 assets 目录，也不用担心相对路径问题。
	fmt.Print("嵌入的字符串内容:\n", message)
	fmt.Printf("嵌入的字节数: %d\n", len(messageBytes))

	// ⚠️ 注意：embed 路径规则
	//   - 不能包含 ..（不能跳出当前包目录）
	//   - 不能以 / 开头（不是绝对路径）
	//   - 只能嵌入当前目录及其子目录中的文件
	// ⚠️ 注意：嵌入发生在编译期，编译后修改 assets/message.txt 不会生效，
	//   需要重新编译。
}

// ===== 示例 2：embed.FS 当作只读文件系统 =====
func example2EmbedFS() {
	// embed.FS 实现了 fs.FS 接口，可以像操作文件系统一样读取嵌入的文件
	data, err := assetsFS.ReadFile("assets/message.txt")
	if err != nil {
		fmt.Println("读取失败:", err)
		return
	}
	fmt.Printf("通过 embed.FS 读取到 %d 字节\n", len(data))

	// 也可以遍历目录
	entries, err := assetsFS.ReadDir("assets")
	if err != nil {
		fmt.Println("读取目录失败:", err)
		return
	}
	for _, e := range entries {
		fmt.Println("assets 下的文件:", e.Name())
	}

	// embed.FS 常用于：嵌入静态网页、HTML 模板、配置文件等，
	// 与 http.FS / template.ParseFS 配合使用。
}

// ===== 示例 3：//go:generate 指令说明 =====
func example3Generate() {
	// 本文件顶部登记了 //go:generate gofmt -w $GOFILE。
	// 要真正执行它，需要手动运行：
	//   go generate ./03-advanced/...
	// go generate 会扫描源码中的 //go:generate 注释并逐条执行。
	//
	// 为什么：代码生成（如 protobuf、mock、stringer）产物需要随源码
	// 一起维护，把生成命令写在源码旁边可以保证两者不脱节。
	//
	// ⚠️ 注意：go generate 与 go build 完全独立，不会被自动触发；
	// 本课程仅做语法演示，不会真的执行生成命令。
	fmt.Println("查看本文件顶部的 //go:generate 注释即可，无需执行")
}

// ===== 示例 4：build tags 条件编译 =====
func example4BuildTags() {
	// build tag 写在文件第一行，用 //go:build 指定编译条件。
	//
	// 本文件自身的活例子：第一行的 `//go:build ignore`
	//   - 它让 go build ./... 跳过本文件（ignore 表示永不参与构建），
	//     但 `go run 文件名` 指定文件时不受影响，可以正常运行。
	//   - 这是"一个目录下放多个独立可运行 main"的常见技巧。
	//
	// 其他常见写法（仅注释演示，不要真的写进本文件）：
	//   //go:build linux           —— 只在 Linux 下编译
	//   //go:build windows && amd64 —— 多条件与
	//   //go:build linux || darwin —— 多条件或
	//   //go:build !windows        —— 取反
	//
	// 配合文件命名约定更常见：
	//   foo_linux.go / foo_windows.go —— 编译器按文件名自动识别目标平台
	//
	// ⚠️ 注意：//go:build 行与 package 声明之间必须有空行（本文件即是如此），
	//   否则不会被识别为编译约束。
	fmt.Printf("本文件通过 //go:build ignore 躲开了批量构建，却可以被 go run 直接执行\n")
}
