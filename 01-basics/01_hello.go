//go:build ignore

// ============================================================
// 文件：01_hello.go
// 阶段：初级
// 主题：Go 程序的基本结构
// 学习目标：
//   1. 理解 package / import / func main 的基本结构
//   2. 掌握 fmt.Println / Printf / Print 的区别
//   3. 学会常用格式化动词 %d %s %v %T %q
//   4. 理解 go run 与 go build 的区别
// 运行方式：go run 01-basics/01_hello.go
// 预计阅读时间：10 分钟
// ============================================================

// package 声明必须是文件的第一行有效代码（注释与构建标签除外）。
// main 包是特殊包：它告诉编译器"这是一个可执行程序"，而不是库。
package main

// import 导入本文件用到的包。
// fmt 是标准库中的格式化 I/O 包，提供打印函数。
// Go 零第三方依赖时，只用标准库就能完成绝大多数入门学习。
import "fmt"

// main 函数是程序的入口，必须出现在 main 包中。
// 它不能带参数、不能有返回值——这与 C/Java 不同。
func main() {
	fmt.Println("========== 示例 1：程序结构 ==========")
	example1ProgramStructure()
	fmt.Println()

	fmt.Println("========== 示例 2：三种打印函数的区别 ==========")
	example2PrintVariants()
	fmt.Println()

	fmt.Println("========== 示例 3：常用格式化动词 ==========")
	example3FormatVerbs()
	fmt.Println()

	fmt.Println("========== 示例 4：go run 与 go build ==========")
	example4RunVsBuild()
	fmt.Println()

	fmt.Println("========== 示例 5：打印的进阶用法 ==========")
	example5MorePrinting()
}

// ===== 示例 1：程序结构 =====
func example1ProgramStructure() {
	// 一个最小 Go 程序由三部分组成：
	//   1. package main    —— 声明这是可执行程序
	//   2. import "fmt"    —— 引入需要的包
	//   3. func main()     —— 程序入口
	// 三者缺一不可（import 只在用到时需要）。
	fmt.Println("package main + import + func main = 最小可执行程序")
	fmt.Println("Hello, Go! 我的第一个 Go 程序")
	// ⚠️ 注意：import 了却没使用的包会编译报错"imported and not used"。
	// 这是 Go 的刻意设计，目的是保持代码整洁、编译快速。
	// ⚠️ 注意：func main 必须没有参数也没有返回值。
	// 如果需要命令行参数，要用 os.Args；需要退出码，要用 os.Exit，
	// 这些都属于后续章节的内容，现在先记住入口的形态即可。
	// Go 的语句结尾不需要分号——编译器会自动在换行处补上分号，
	// 这也是"左花括号不能单独占一行"这一风格规定的由来。
}

// ===== 示例 2：三种打印函数的区别 =====
func example2PrintVariants() {
	// Println：参数之间自动加空格，结尾自动换行。
	// 适合快速、逐行地输出调试信息。
	fmt.Println("Println", "自动加空格", "并换行")

	// Print：不加空格、不换行，参数原样拼接。
	// ⚠️ 注意：Print 只会在两个相邻参数"都不是字符串"时才插入空格，
	// 因此输出纯字符串时很容易连成一片，初学者常在这里踩坑。
	fmt.Print("Print")
	fmt.Print("不自动换行")
	fmt.Print("，需要手动加 \\n\n")

	// Printf：按格式串输出，类似 C 语言的 printf。
	// 格式串中的 % 动词会被后面的参数依次替换。
	// 它不会自动换行，所以通常以 \n 结尾。
	name := "小明"
	age := 18
	fmt.Printf("Printf：我叫%s，今年%d岁。\n", name, age)
}

// ===== 示例 3：常用格式化动词 =====
func example3FormatVerbs() {
	// %d：十进制整数
	fmt.Printf("%%d  十进制整数   -> %d\n", 42)
	// %s：字符串
	fmt.Printf("%%s  字符串        -> %s\n", "你好")
	// %v：值的"默认格式"，几乎什么类型都能打印，是调试时的万能动词。
	fmt.Printf("%%v  默认格式      -> %v / %v / %v\n", 3.14, true, []int{1, 2, 3})
	// %T：打印值的类型，学习类型系统时非常有用。
	fmt.Printf("%%T  值的类型      -> %T / %T\n", 42, "hello")
	// %q：带双引号的字符串（或带引号的字符），可以直观看到转义字符。
	fmt.Printf("%%q  带引号字符串  -> %q\n", "有\n换行")
	// 提示：%% 是字面量百分号的写法，上面格式串里的 %%d 只是原样输出 "%d"。
	// 还有 %f（浮点）、%t（布尔）、%x（十六进制）等，后续会陆续遇到。
	fmt.Printf("%%f  浮点数        -> %.2f\n", 3.14159) // %.2f 保留两位小数
	fmt.Printf("%%t  布尔值        -> %t\n", 2 > 1)
}

// ===== 示例 4：go run 与 go build =====
func example4RunVsBuild() {
	// 这部分用文字讲解，因为 go run / go build 是命令行工具而非代码：
	//
	// go run 01-basics/01_hello.go
	//   - "编译 + 运行"一步完成，可执行文件放在临时目录，运行完即删除。
	//   - 适合学习、调试、快速验证，本系列示例都用这种方式运行。
	//
	// go build 01-basics/01_hello.go
	//   - 只编译不运行，在当前目录生成一个可执行文件（Windows 下是 01_hello.exe）。
	//   - 生成的 exe 可以脱离 Go 环境直接双击/拷贝到别的机器运行。
	//
	// ⚠️ 注意：本文件开头有一行 //go:build ignore（构建标签）。
	// 它的作用是让 `go build ./...` 这类整包构建命令忽略本文件，
	// 因为同目录下多个示例都有各自的 main 函数，一起编译会冲突。
	// 而 `go run 单个文件.go` 会忽略该标签，所以可以正常运行。
	fmt.Println("go run   = 编译到临时目录并立即运行（学习首选）")
	fmt.Println("go build = 生成独立的 .exe 可执行文件（发布用）")
	fmt.Println("文件头的 //go:build ignore 让整包构建时跳过本示例文件")
}

// ===== 示例 5：打印的进阶用法 =====
func example5MorePrinting() {
	// Sprint 家族：不直接输出，而是把格式化结果作为字符串返回。
	// 常用于拼接日志、构造错误信息等"先组合再使用"的场景。
	msg := fmt.Sprintf("Sprintf 拼出的字符串：%d + %d = %d", 1, 2, 3)
	fmt.Println(msg)

	// Println 的姊妹函数 Sprintln 同理，返回带空格和换行的字符串。
	line := fmt.Sprintln("这行", "来自 Sprintln")
	fmt.Print("Print 原样输出它：", line)

	// %x 十六进制、%b 二进制、%o 八进制：整数的不同进制视角。
	fmt.Printf("255 的十六进制 %%x=%x，二进制 %%b=%b，八进制 %%o=%o\n", 255, 255, 255)

	// 宽度与对齐：%5d 右对齐占 5 位，%-5d 左对齐，%05d 用 0 补齐。
	// 打印表格类输出时很有用。
	fmt.Printf("右对齐：[%5d] 左对齐：[%-5d] 补零：[%05d]\n", 42, 42, 42)

	// ⚠️ 注意：Printf 参数个数与动词个数不匹配时不会编译报错，
	// 而是输出 %!d(MISSING) 之类的怪字符串，只能靠肉眼检查发现。
	fmt.Printf("动词多于参数时的怪输出：%d %d\n", 1)
	// 学习建议：现阶段记住 %d %s %v %T %q 这五个动词就够用了，
	// 其中 %v 不知道用哪个时的兜底选择。
	// 如果担心写错格式串，可以让编辑器/IDE 帮你跑 go vet，
	// 它能静态检查出 Printf 动词与参数类型不匹配等常见问题。
	fmt.Println("小结：Println 最省事，Printf 最灵活，%v 最万能")
	// 至此，你已经掌握了 Go 程序的骨架与最常用的输出方式。
	// 下一个文件将学习变量与常量的声明。
}
