# Go 中级阶段：接口、错误处理与并发

> 适用项目：go-learning（go 渐进式学习项目）
> 阶段目录：`02-intermediate/`
> 前置要求：已完成初级阶段（`01-basics/`），熟悉基本语法、函数、切片、map 与结构体

中级阶段是从"会写 Go"走向"写好 Go"的关键一步。本阶段的代码不再停留在语法层面，而是围绕 Go 的三大核心设计哲学展开：**接口的隐式实现**、**显式的错误处理**、**基于 goroutine 与 channel 的并发模型**。此外还补充了闭包、泛型、JSON/文件 IO 与测试这四个工程实践中绕不开的主题。

---

## 一、本阶段学习目标

完成本阶段后，你应当能够：

1. **掌握接口的隐式实现与类型断言**：理解"实现了方法就是实现了接口"，熟练使用 comma-ok 断言与 type switch，会用小接口组合表达最小依赖。
2. **建立正确的错误处理习惯**：把 error 当作普通返回值处理，会用 `%w` 包装错误链，会用 `errors.Is` / `errors.As` 做错误判定，明确 panic/recover 的适用边界。
3. **能写安全的 goroutine / channel 并发代码**：会用 `sync.WaitGroup` 等待、用 `Mutex` / `RWMutex` / `atomic` 保护共享状态、用 channel 做同步与通信，能识别并避免 goroutine 泄漏与死锁。
4. **会用 context 控制取消与超时**：理解 `WithCancel` / `WithTimeout` / `WithDeadline` 的取消传播机制，知道 `WithValue` 的适用边界。
5. **会写表驱动测试**：掌握 `go test` 的基本用法，能编写表驱动测试、子测试、基准测试与 Example 示例测试。

---

## 二、环境准备与运行方式

### Go 版本要求

- 本项目按 **Go 1.27.1** 编写，要求 **Go 1.22+**（循环变量语义在 Go 1.22 发生了变化，`03_closures.go` 的示例 5 依赖这一新行为）。
- 零第三方依赖，全部使用标准库。
- 验证版本：

```powershell
go version
```

### 运行单个示例文件

每个 `.go` 文件都带有 `//go:build ignore` 构建标记，可脱离包管理独立运行。在**项目根目录**（`D:\Project\gogogo`）下执行：

```powershell
go run 02-intermediate/01_interfaces.go
```

其余文件同理，把文件名换成对应的即可。

### 运行测试（重要）

测试比较特殊：`10_testing.go` 与 `10_testing_test.go` 都带有 `//go:build ignore` 标记（为了让教学示例能独立 `go run`），所以不能用常规的 `go test ./...`，必须**进入 `02-intermediate` 目录**并用**显式文件列表**的方式运行：

```powershell
cd 02-intermediate

# 运行全部测试（-v 显示每个测试与子测试的明细）
go test -v 10_testing.go 10_testing_test.go

# 只跑某个测试（-run 后跟测试名，支持正则）
go test -run TestIsPalindrome -v 10_testing.go 10_testing_test.go

# 跑基准测试：-benchtime=10x 表示每个基准只迭代 10 次，快速看结果
go test "-bench=." "-benchtime=10x" 10_testing.go 10_testing_test.go
```

> ⚠️ PowerShell 注意：`-bench=.` 和 `-benchtime=10x` 在 PowerShell 中**必须加引号**（如 `"-bench=."`），否则 `.` 等字符可能被 shell 解释，导致参数解析失败。

### 竞态检测

Go 内置的 race detector 可以在运行时捕捉数据竞态，学习并发时非常有用：

```powershell
go run -race 02-intermediate/06_sync.go
```

关于 race detector 的更多说明见 `06_sync.go` 一节。

---

## 三、文件清单表

| 文件名 | 主题 | 核心知识点 | 建议阅读顺序 |
| --- | --- | --- | --- |
| 01_interfaces.go | 接口 | 隐式实现、空接口 any、comma-ok 类型断言、type switch、小接口组合、fmt.Stringer | 1 |
| 02_errors.go | 错误处理 | error 接口本质、errors.New、fmt.Errorf 与 %w 包装、errors.Is / errors.As、nil 接口坑、panic/recover 边界 | 2 |
| 03_closures.go | 闭包 | 按引用捕获、计数器与生成器、函数式选项模式、Go 1.22 循环变量新语义 | 3 |
| 04_goroutines.go | goroutine 并发 | go 语句、主 goroutine 生命周期、sync.WaitGroup、GOMAXPROCS、goroutine 泄漏与正确退出 | 4 |
| 05_channels.go | channel 通道 | 无缓冲/有缓冲阻塞行为、close 与 comma-ok 接收、for range 遍历、select 多路复用与超时、单向 channel、nil channel | 5 |
| 06_sync.go | sync 包与并发同步 | Mutex、RWMutex、Once、atomic 原子操作、sync.Map、race detector | 6 |
| 07_context.go | context 与取消传播 | WithCancel、WithTimeout、WithDeadline、取消沿调用链传递、WithValue 传元数据 | 7 |
| 08_generics.go | 泛型（类型参数） | [T any] 类型参数、自定义约束接口、comparable 与 cmp.Ordered、~ 近似约束、泛型类型 Stack[T] | 8 |
| 09_json_io.go | JSON 编解码与文件 IO | json.Marshal / Unmarshal、struct tag、os.WriteFile / os.ReadFile、io.Reader / io.Writer、bufio.Scanner | 9 |
| 10_testing.go + 10_testing_test.go | go test 测试入门 | 表驱动测试、t.Run 子测试、Errorf/Fatalf、Benchmark 基准测试、Example 示例测试 | 10 |

---

## 四、逐文件详解

### 01_interfaces.go：接口

**这个文件带你认识什么**

接口是 Go 实现多态的唯一机制，也是理解标准库（`fmt.Stringer`、`io.Reader`、`error`）的钥匙。Go 的接口是**隐式实现**的：一个类型只要具备接口要求的全部方法，就自动实现该接口，不需要 `implements` 之类的关键字。本文件从接口定义讲起，依次覆盖空接口 `any`、类型断言、type switch、接口组合，最后落到一个最实用的内置接口 `fmt.Stringer`。

**运行命令与预期输出解读**

```powershell
go run 02-intermediate/01_interfaces.go
```

输出分五个段落：

- **示例 1**：接口变量 `s` 先后装入 `Dog` 和 `Cat`，各自打印叫声；随后一个 `[]Speaker` 切片统一遍历，演示多态。
- **示例 2**：`printAny` 三次打印任意值的值与动态类型（`%v` 与 `%T`）；随后演示 comma-ok 断言——断言 string 成功，断言 int 失败但**不 panic**，得到零值并打印提示。
- **示例 3**：`describe` 用 type switch 分别处理 int、string、bool，3.14（float64）落入 default 分支。
- **示例 4**：`*Buffer` 作为 `ReadWriter` 写入两次后再读出拼接结果；随后把同一个值收窄为小接口 `Reader` 再读取。
- **示例 5**：实现了 `String()` 的 `Point` 被 `fmt.Println` 与 `%v` 打印成自定义格式 `Point(3, 4)`，与未实现 Stringer 的匿名结构体的默认字段输出形成对比。

本文件没有 goroutine，输出顺序完全确定。

**关键代码讲解**

1. 隐式实现——编译器替你检查：

```go
type Speaker interface {
	Speak() string
}

func (d Dog) Speak() string { return d.Name + "：汪汪！" }
```

`Dog` 上定义了 `Speak() string`，它就自动是 `Speaker`。为什么这样设计？因为接口的定义方和使用方可以互不知道对方的存在——标准库定义了 `fmt.Stringer`，你的类型不需要 import fmt 就能"碰巧"实现它，耦合度降到最低。

2. comma-ok 断言——永不 panic 的取回方式：

```go
var v any = "Go 语言"
if str, ok := v.(string); ok {
	fmt.Println("断言成功，是字符串：", str)
}
if n, ok := v.(int); !ok {
	fmt.Println("断言失败：v 不是 int，得到零值", n)
}
```

`v.(T)` 的单返回值形式在失败时会 panic；双返回值（comma-ok）形式失败时 `ok` 为 false、`str`/`n` 为该类型零值。**凡是不能 100% 确定动态类型的断言，都必须用 comma-ok。**

3. 指针接收者与接口满足：

```go
func (b *Buffer) Read() string   { return b.data }
func (b *Buffer) Write(s string) { b.data += s }

var rw ReadWriter = &Buffer{} // 必须是 *Buffer，不能是 Buffer
```

`Write` 要修改内部状态，所以定义在指针接收者上。方法集规则决定：**只有 `*Buffer` 的方法集同时包含指针接收者与值接收者方法**，因此这里必须传指针才能满足 `ReadWriter`。

4. 小接口组合表达最小依赖：

```go
type ReadWriter interface {
	Reader
	Writer
}
```

函数只声明自己需要的最小能力（只需要读就声明 `Reader` 参数），调用方就能传入任何满足条件的类型——这是 `io.ReadWriter` 等标准库接口的设计思想。

**常见坑/易错点**

- ⚠️ **直接类型断言失败会 panic**：`s := v.(string)` 在 v 不是 string 时直接崩溃，务必使用 comma-ok 形式（文件中已标注）。
- **指针接收者方法的类型，值形式不满足接口**：方法定义在 `*T` 上时，`T` 类型的变量不能赋给要求该方法的接口，新手常在这里遇到编译错误。
- **空接口 `any` 不是"没有类型"**：`any` 变量内部始终携带（动态类型, 动态值）二元组，`%T` 打印的是动态类型。
- **不要在 String() 里用 fmt.Sprintf("%v", p) 打印自己**：`%v` 会再次调用 `String()`，造成无限递归（要用 `Point(p)` 转换绕开）。

**动手练习**

1. 新增一个 `Robot` 类型实现 `Speaker`，把它也放进 `animals` 切片，观察多态输出。
2. 给 `describe` 增加 `case float64:` 分支，打印"浮点数，保留两位小数"。提示：在 type switch 的 case 里，`t` 已经是该分支的具体类型，可直接用 `fmt.Sprintf("%.2f", t)`。

---

### 02_errors.go：错误处理

**这个文件带你认识什么**

Go 没有异常机制，错误就是普通的接口返回值。本文件把错误处理讲透：`error` 接口的本质、哨兵错误、`fmt.Errorf` 的 `%w` 包装如何形成错误链、`errors.Is` 与 `errors.As` 如何沿链检查，以及"装了 nil 指针的 error 接口不等于 nil"这个经典陷阱，最后划清 panic/recover 的使用边界。

**运行命令与预期输出解读**

```powershell
go run 02-intermediate/02_errors.go
```

输出分五个段落：

- **示例 1**：`findUser(-1)` 打印"id 必须为正数"，`findUser(200)` 打印哨兵错误"记录不存在"，`findUser(50)` 返回 nil 并打印出来（`<nil>`）。
- **示例 2**：打印 `%w` 包装后的完整错误链文本，以及一个不带 `%w` 的普通格式化错误作对比。
- **示例 3**：`errors.Is` 穿透包装层命中 `ErrNotFound`；`errors.As` 从错误中取出 `*QueryError` 并打印其 `Code` 与 `Query` 字段。
- **示例 4**：两行布尔打印——未赋值的 error 是 nil，但装了 nil `*QueryError` 指针的 error 接口**不是 nil**，随后错误处理分支被误触发。
- **示例 5**：`safeDivide(10, 2)` 正常返回 5；`safeDivide(10, 0)` 内部 panic 被 defer+recover 捕获转为 error 返回；最后一行证明程序没有崩溃。

输出顺序完全确定。

**关键代码讲解**

1. `%w` 包装形成错误链：

```go
err := fmt.Errorf("打开配置 %s 失败", name)
return fmt.Errorf("初始化模块: %w", err)
```

不带 `%w` 的 `fmt.Errorf` 只是格式化出一个新字符串错误；加上 `%w` 后，底层错误被"包"进新错误，形成一条可以追溯的**错误链**，之后 `errors.Is` / `errors.As` 能沿链逐层检查。

2. errors.Is 与 errors.As 的分工：

```go
if errors.Is(err, ErrNotFound) {
	fmt.Println("errors.Is 命中：底层原因是记录不存在")
}

var qe *QueryError
if errors.As(err2, &qe) {
	fmt.Printf("errors.As 取出字段：Code=%d Query=%q\n", qe.Code, qe.Query)
}
```

- `errors.Is(err, target)`：判断"是不是某个错误"，用于哨兵错误比较，能穿透 `%w` 包装层。
- `errors.As(err, &target)`：判断"错误链里有没有某种**类型**的错误"，命中后把该错误填入 target，用于读取自定义错误的字段。

3. 用 defer+recover 把 panic 转成 error：

```go
func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("发生 panic，已安全恢复：%v", r)
		}
	}()
	if b == 0 {
		panic("除数为零（演示用 panic）")
	}
	return a / b, nil
}
```

`recover` 只在 defer 函数中生效，它拦截 panic 并让函数以命名返回值 `err` 正常返回。这是"在程序边界兜底"的标准写法。

**常见坑/易错点**

- ⚠️ **装了 nil 指针的 error 接口不等于 nil**：`var qe *QueryError = nil; err = qe` 后 `err != nil` 为 true，因为接口里装着 `(类型=*QueryError, 值=nil)`。函数返回错误时，成功路径应显式 `return nil`，不要返回一个可能为 nil 的具体类型指针。
- ⚠️ **errors.As 的第二个参数必须是非 nil 指针**，且指向目标错误类型（如 `&qe`，其中 `qe` 是 `*QueryError`），传错会 panic。
- **panic 不是异常机制的替代品**：文件明确约定——panic 只用于"不可恢复的程序错误"，正常业务错误（文件不存在、网络超时）一律走 error 返回值；recover 通常只在程序边界使用。
- **不要比较错误文案**：`err.Error() == "记录不存在"` 脆弱且无法穿透包装，应该用 `errors.Is`。

**动手练习**

1. 给 `QueryError` 增加一个 `Unwrap() error` 方法返回内部原因字段（需要先在结构体里加该字段），验证 `errors.Is` 能否穿透你的自定义错误。
2. 修改 `demoNilError`：写一个返回 `*QueryError` 的函数，在无错误时返回 nil 指针并赋给 error 变量，观察 `err != nil` 的结果，然后改成成功时显式 `return nil` 修复它。

---

### 03_closures.go：闭包

**这个文件带你认识什么**

闭包是"携带了自己环境的函数"。本文件用五个递进示例讲清闭包的核心机制：按引用捕获变量、用闭包保存私有状态（计数器、斐波那契生成器）、函数式选项模式这一重要工程实践，以及 Go 1.22 修复的循环变量闭包陷阱。

**运行命令与预期输出解读**

```powershell
go run 02-intermediate/03_closures.go
```

输出分五个段落：

- **示例 1**：`say()` 第一次打印"你好"，修改外层 `msg` 后第二次打印"世界"——证明闭包捕获的是变量本身而非当时的值。
- **示例 2**：`c1` 连续调用三次输出 1 2 3，独立的 `c2` 从 1 重新开始——每个计数器有自己的 `count`。
- **示例 3**：一行打印斐波那契前 10 项（0 1 1 2 3 5 8 13 21 34）。
- **示例 4**：打印默认配置 `{host:localhost port:8080}` 与自定义配置 `{host:0.0.0.0 port:9090}`。
- **示例 5**：Go 1.22+ 下三个闭包各自打印 0 1 2（而非旧版本的 3 3 3）；随后演示显式共享变量 `shared`，三个闭包调用后其值为 3。

输出顺序完全确定。

**关键代码讲解**

1. 按引用捕获——闭包看到的是"活的"变量：

```go
msg := "你好"
say := func() {
	fmt.Println("闭包读到：", msg)
}
say()      // 你好
msg = "世界"
say()      // 世界
```

闭包捕获的是变量本身（引用），不是定义那一刻的值副本。外层变量被闭包引用后会"逃逸"到堆上，外层函数返回后依然存活。

2. 计数器——闭包作为私有状态的载体：

```go
func newCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}
```

`count` 是自由变量：外层函数返回后它依然存活，且只有返回的闭包能访问它——这是 Go 里实现"私有字段"的一种轻量方式，每次调用 `newCounter` 都会得到一份独立的 `count`。

3. 函数式选项模式：

```go
type Option func(*Server)

func WithHost(h string) Option {
	return func(s *Server) { s.host = h }
}

func NewServer(opts ...Option) *Server {
	s := &Server{host: "localhost", port: 8080}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
```

`WithHost("0.0.0.0")` 并不是立刻设置 host，而是返回一个"把参数封存进函数体"的闭包，由 `NewServer` 统一应用。好处：默认值、可选参数、任意组合、向后兼容（新增选项不破坏旧调用），这是 Go 生态（如 gRPC）的主流配置风格。

4. Go 1.22 的循环变量新语义：

```go
for i := 0; i < 3; i++ {
	funcs = append(funcs, func() {
		fmt.Print(i, " ") // 1.22+：每个闭包捕获各自迭代的新 i
	})
}
```

Go 1.22 之前，循环变量 `i` 是同一个变量被反复赋值，闭包捕获引用后读到的都是最终值；1.22 起每次迭代的 `i` 都是新变量，每个闭包各得其所。

**常见坑/易错点**

- ⚠️ **阅读旧代码/旧教程时注意循环变量陷阱**：Go < 1.22 的惯用写法是 `i := i` 或把 i 作为参数传入闭包；新代码不再需要，但旧代码里的这些写法不是冗余。
- **按引用捕获意味着共享**：多个闭包捕获同一个变量时会相互影响（示例 5 的 `shared`），并发场景下捕获共享变量还要考虑加锁。
- **闭包会延长变量的生命周期**：大对象被闭包捕获后无法被 GC 回收，直到闭包本身被释放。
- **闭包不是并发安全的**：计数器示例在单 goroutine 下工作良好，多 goroutine 调用同一个计数器就是数据竞态（见 06_sync.go）。

**动手练习**

1. 写一个 `newAdder(base int) func(int) int`，返回的闭包每次调用把参数累加到自己的内部和（从 `base` 开始）并返回当前和。提示：参照 `newCounter`，把 `count` 初始化为 `base`。
2. 给选项模式增加 `WithTimeout(d int) Option` 选项（需要在 `Server` 里加 `timeout` 字段），构造一个只改超时、其余用默认值的服务器。

---

### 04_goroutines.go：goroutine 并发

**这个文件带你认识什么**

goroutine 是 Go 并发的执行单元——比线程轻量得多，可以轻松创建成千上万个。本文件讲四件事：`go` 语句的语义与"主 goroutine 退出即程序终止"的铁律、用 `sync.WaitGroup` 做确定性等待、用 `runtime.GOMAXPROCS` 观察并行度，以及 goroutine 泄漏的演示与正确的退出方式。

**运行命令与预期输出解读**

```powershell
go run 02-intermediate/04_goroutines.go
```

输出分五个段落：

- **示例 1**：worker A、B 各打印一行，main 最后打印等待结束。**worker A 与 B 的输出顺序可能不同**（调度顺序不确定）。
- **示例 2**：三个任务各打印"任务 N 完成"，**顺序可能不同**；最后由 WaitGroup 保证打印"所有任务已完成"。
- **示例 3**：打印逻辑 CPU 数、当前 GOMAXPROCS，临时改为 2 再打印，然后恢复原值。
- **示例 4**：用 `runtime.NumGoroutine` 对比泄漏前后的 goroutine 数量，后者比前者多 1——那个永远阻塞在 channel 接收上的 goroutine 就是泄漏。
- **示例 5**：先打印运行中的 goroutine 数，`close(done)` 广播退出信号后，三个 worker 各自打印退出信息（**顺序可能不同**），收尾后的 goroutine 数回落到接近初始值。

**关键代码讲解**

1. 主 goroutine 退出 = 程序终止：

```go
go worker("A")
go worker("B")

time.Sleep(50 * time.Millisecond) // 简单等待，仅演示；生产代码请用 WaitGroup
```

`main` 返回的瞬间，整个进程退出，**不管其他 goroutine 是否跑完**。所以启动 goroutine 后必须有等待机制。`time.Sleep` 只够教学演示，因为 sleep 时长靠猜——任务快了浪费时间，任务慢了丢失输出。

2. WaitGroup 确定性等待：

```go
var wg sync.WaitGroup
for i := 1; i <= 3; i++ {
	wg.Add(1)
	go func(id int) {
		defer wg.Done()
		fmt.Printf("  任务 %d 完成\n", id)
	}(i)
}
wg.Wait()
```

`Add(1)` 计数加一、`Done()` 计数减一、`Wait()` 阻塞到归零。`defer wg.Done()` 保证即使函数中途 panic（未恢复）前也执行。注意 `i` 作为参数传入闭包——语义清晰，任何 Go 版本都不会出错。

3. 用 close(channel) 广播退出信号：

```go
select {
case <-done: // 收到退出信号，主动结束，不再泄漏
	return
case <-time.After(5 * time.Second): // 兜底超时，防止永远阻塞
	return
}
...
close(done) // 广播：所有监听 done 的 goroutine 都被唤醒
```

已关闭 channel 上的接收操作**立即返回零值**，且对所有接收者都成立——一次 `close` 就能通知任意多个监听者，这是 Go 里广播取消信号的经典手法（07_context.go 的 `ctx.Done()` 正是这个模式）。

**常见坑/易错点**

- ⚠️ **main 结束程序就结束**：示例 1 的注释明确警告——不等待的话 worker 可能一行都来不及打印。
- ⚠️ **goroutine 泄漏无法回收**：泄漏的 goroutine 永远阻塞、占用栈内存，堆积多了会耗尽内存；示例 4 用数字把泄漏"可视化"了。排查工具：`runtime.NumGoroutine`、pprof。
- **WaitGroup 的 Add 要在启动 goroutine 之前调用**：放反了可能出现 `Wait()` 先于 `Add` 返回的竞态。
- **GOMAXPROCS 限制的是并行数，不是 goroutine 总数**：成千上万的 goroutine 依然可以在少量 P 上被调度（示例 3 的注释）。

**动手练习**

1. 把示例 1 的 `time.Sleep` 换成 `sync.WaitGroup`，删除 sleep 后程序行为应完全确定。
2. 修改示例 5：不调用 `close(done)`，仅靠 `time.After` 兜底，用 `runtime.NumGoroutine` 观察 5 秒后数量是否回落，体会"广播信号"比"各自超时"快多少。

---

### 05_channels.go：channel 通道

**这个文件带你认识什么**

channel 是 goroutine 之间的类型化管道，体现 Go 的座右铭"不要通过共享内存来通信，而要通过通信来共享内存"。本文件覆盖：无缓冲与有缓冲 channel 的阻塞语义、`close` 与接收端检测、`for range` 遍历、`select` 多路复用与 `time.After` 超时、单向 channel 参数与 nil channel 的坑。

**运行命令与预期输出解读**

```powershell
go run 02-intermediate/05_channels.go
```

输出分六个段落：

- **示例 1**：goroutine 发送、main 接收，打印收到的字符串——一次"同步握手"。
- **示例 2**：容量为 2 的缓冲 channel 先无阻塞地存入两个值，第三个值由 goroutine 发送（会阻塞到 main 开始接收），最后依次取出 1、2、3。
- **示例 3**：关闭 channel 后用 comma-ok 接收：第一次 `ok=true` 取到 10，取走 20 后第二次 `ok=false`、v 为零值 0。
- **示例 4**：`for range` 依次收到 1、2、3，channel 关闭后循环自动结束。
- **示例 5**：select 两轮各选中一个 channel（**输出顺序可能不同**，c2 只睡 10ms 通常先到）；随后演示超时——200ms 的慢操作被 50ms 的 `time.After` 截断，打印超时信息。
- **示例 6**：producer/consumer 通过单向 channel 参数协作打印消费 1、2、3；随后用 `select + default` 安全演示 nil channel 上"取不到值也不死锁"。

**关键代码讲解**

1. 无缓冲 channel 是同步握手：

```go
ch := make(chan string) // 无缓冲：发送方必须等到接收方准备好
go func() {
	ch <- "你好" // 发送会阻塞，直到 main 接收
}()
msg := <-ch
```

无缓冲 channel 上，发送和接收必须在时间上"碰头"，否则双方互相等待。这既是阻塞点，也是天然的同步点——接收完成意味着发送方一定已经执行到发送语句。

2. close 与 comma-ok 接收：

```go
close(ch) // 关闭：表示"不再发送"，已发送的值仍可取出

v, ok := <-ch // ok=false 表示 channel 已关闭且没有剩余值
```

`close` 的语义是"不再发送"，缓冲区里已存的值**仍然可以正常取出**；耗尽后再接收，`ok` 为 false、v 是元素零值。这正是 `for range` 遍历 channel 能自动退出的原理。

3. select + time.After 实现超时：

```go
select {
case msg := <-slow:
	fmt.Println("  收到：", msg)
case <-time.After(50 * time.Millisecond):
	fmt.Println("  超时：50ms 内没等到结果，放弃")
}
```

`select` 阻塞到任一 case 就绪；`time.After` 返回一个到点就有值送达的 channel，于是 50ms 后无论慢操作是否完成，超时分支必然就绪。

4. nil channel 的阻塞特性与正面用途：

```go
var nilCh chan int // nil channel：发送和接收都会永远阻塞
select {
case <-nilCh:
	fmt.Println("  永远不会执行到这里")
default:
	fmt.Println("  nil channel 上无值可取（default 兜底，没有死锁）")
}
```

nil channel 的收发永远阻塞，因此它在 `select` 里**永远不就绪**——反过来看，这是"临时禁用某个 select 分支"的标准技巧。

**常见坑/易错点**

- ⚠️ **向已关闭的 channel 发送会 panic**；关闭动作应由**发送方**负责（示例 3 注释）。
- ⚠️ **`for range` 遍历 channel 时忘记 close 会死锁**：循环会一直等新值，永远等不到（示例 4 注释）。
- ⚠️ **直接对 nil channel 收发会死锁**：要探测请用 `select + default`（示例 6 注释）。
- **缓冲满时发送阻塞、空时接收阻塞**：示例 2 演示了缓冲满后第三个发送必须等接收方腾位置。
- **超时分支返回后慢 goroutine 可能还在跑**：示例 5 里 200ms 的慢 goroutine 仍会完成并试图发送——无缓冲 channel 没人接收它就会泄漏，生产代码应配合缓冲 channel 或 context。

**动手练习**

1. 写一个 `fanIn(c1, c2 <-chan string) <-chan string`：启动一个 goroutine 把两个输入 channel 的值都转发到一个输出 channel，两个输入都关闭后关闭输出。提示：用两个 goroutine 各 range 一个输入，配合 WaitGroup。
2. 修改示例 5 的超时代码：把 `slow` 改成容量为 1 的缓冲 channel，说明为什么这样能避免 goroutine 泄漏。

---

### 06_sync.go：sync 包与并发同步

**这个文件带你认识什么**

channel 适合"通信"，而保护一小段共享状态用 `sync` 包的锁往往更直接。本文件覆盖并发同步五件套：`Mutex` 互斥锁、`RWMutex` 读写锁、`Once` 单次初始化、`atomic` 原子操作、`sync.Map` 并发安全 map，并在注释中给出了用 race detector 亲手抓数据竞态的方法。

**运行命令与预期输出解读**

```powershell
go run 02-intermediate/06_sync.go

# 竞态检测（学习并发必做）
go run -race 06_sync.go   # 在 02-intermediate 目录内；根目录下用 go run -race 02-intermediate/06_sync.go
```

输出分五个段落：

- **示例 1**：100 个 goroutine 各累加 1000 次，打印计数器终值——加锁后**恰好是 100000**（goroutine 内部执行顺序不确定，但终值确定）。
- **示例 2**：1 个写 goroutine 与 5 个读 goroutine 并发操作 RWMutex 保护的缓存，最后打印写入的值 `go`。
- **示例 3**：5 个 goroutine 并发调用 `once.Do(initConfig)`，"initConfig 实际执行了"只打印**一次**，随后打印初始化成功标志。
- **示例 4**：与示例 1 同样的累加规模，改用 `atomic.Int64`，终值同样确定为 100000。
- **示例 5**：10 个 goroutine 往 `sync.Map` 写入（故意写相同的 key），随后演示 `LoadOrStore` 并用 `Range` 遍历全部键值——**遍历输出的顺序可能不同**，但每个 key 各出现一次。

**关键代码讲解**

1. Mutex 划定临界区：

```go
mu.Lock() // 临界区开始：同一时间只允许一个 goroutine 修改 counter
counter++
mu.Unlock() // 临界区结束；忘记 Unlock 会导致其他 goroutine 永久阻塞（死锁）
```

`counter++` 实际是"读-加-写"三步，多 goroutine 交错执行会丢失更新（数据竞态）。锁把这三步变成不可分割的整体。文件注释还留了一个实验：故意注释掉 Lock/Unlock，再跑 `go run -race`，race detector 会报告 `WARNING: DATA RACE`。

2. RWMutex 读写分离：

```go
rw.RLock() // 多个读锁可以共存
_ = cache[key]
rw.RUnlock()

rw.Lock() // 写锁独占：阻塞所有其他读与写
cache[key] = val
rw.Unlock()
```

读多写少时 RWMutex 让并发读互不阻塞；但写锁是独占的，且**读锁内不能再获取写锁，否则必然死锁**（示例 2 注释）。

3. atomic.Int64 轻量计数：

```go
var counter atomic.Int64 // 零值即可用，无需构造
counter.Add(1)           // 底层是 CPU 原子指令，比 Mutex 轻量得多
fmt.Println("atomic 计数器终值:", counter.Load())
```

原子操作由 CPU 指令保证不可分割，适合单个数值/指针的简单操作；但**多个原子操作组合起来不是原子的**，涉及多个变量的一致性时仍需 Mutex（示例 4 注释）。

4. sync.Map 的适用场景：

```go
var m sync.Map // 免锁的并发安全 map，无需初始化
actual, loaded := m.LoadOrStore("hero", "gopher")
```

`sync.Map` 免初始化、无需锁，适合"读多写少"或"不同 goroutine 操作不同 key"的场景；但键值都是 `any`（失去类型安全），写频繁时普通 map+Mutex 往往更快（示例 5 注释）。

**常见坑/易错点**

- ⚠️ **忘记 Unlock 会死锁**：其他 goroutine 永久阻塞在 Lock 上；养成 `mu.Lock(); defer mu.Unlock()` 的习惯。
- ⚠️ **RWMutex 的读锁内不能再获取写锁**（锁升级），否则必然死锁；写多读少场景用 RWMutex 反而更慢。
- ⚠️ **atomic 只适合单个值的简单操作**：两个原子变量之间的一致性它管不了。
- ⚠️ **sync.Map 键值都是 any**：取出时要类型断言，且场景选错（写频繁）性能不如 map+Mutex。
- ⚠️ **Once 不可重置**：初始化可能失败需要重试的场景，应改用 errgroup 或手动锁（示例 3 注释）。

**动手练习**

1. 亲手做一次竞态实验：把示例 1 的 `mu.Lock()` / `mu.Unlock()` 注释掉，先 `go run` 多跑几次看终值是否小于 100000，再 `go run -race 02-intermediate/06_sync.go` 看 race detector 的 `WARNING: DATA RACE` 报告，最后恢复代码。
2. 用 `atomic.Int64` 重写 03_closures.go 的计数器，使其在 100 个 goroutine 并发调用下依然正确。提示：闭包内捕获 `*atomic.Int64`，用 `Add(1)` 代替 `count++`。

---

### 07_context.go：context 包与取消传播

**这个文件带你认识什么**

`context.Context` 是 Go 里跨 API 边界传递"取消信号、截止时间、请求级元数据"的标准载体。本文件覆盖三种取消源（`WithCancel` 手动、`WithTimeout` 超时、`WithDeadline` 截止时间）、取消信号沿调用链 A→B→C 的传播机制，以及 `WithValue` 的正确用法与边界。

**运行命令与预期输出解读**

```powershell
go run 02-intermediate/07_context.go
```

输出分五个段落：

- **示例 1**：main 启动 worker 后 30ms 调用 `cancel()`，worker 打印"收到取消信号"并注明退出原因（`context canceled`），随后 main 等待其真正退出。
- **示例 2**：50ms 超时的 context 交给需要 200ms 的慢任务，任务在 50ms 时被取消并返回 `context.DeadlineExceeded`。
- **示例 3**：先打印距离截止的剩余时间（约 40ms），阻塞到截止时间后打印自动取消的 `Err`。
- **示例 4**：最外层一次 `cancel()`，serviceB 与 serviceA 先后打印退出信息（**两者的打印顺序由代码结构决定为 B 先 A 后**，因为 A 要等 B 的 done），最后打印"整条调用链已干净退出"。
- **示例 5**：从 context 中取出 `request-id` 并打印。

**关键代码讲解**

1. 可取消函数的标准骨架：

```go
func worker(ctx context.Context, id int, done chan<- struct{}) {
	defer close(done)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("worker %d 收到取消信号，退出原因: %v\n", id, ctx.Err())
			return
		default:
			time.Sleep(10 * time.Millisecond) // 模拟一小段工作
		}
	}
}
```

`ctx.Done()` 返回的 channel 在取消时被关闭，于是所有监听者同时被唤醒（正是 04_goroutines.go 示例 5 的 close 广播模式）。`ctx.Err()` 告诉你原因是 `Canceled` 还是 `DeadlineExceeded`。

2. WithTimeout 只是把截止时间变成时长：

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()
err := slowJob(ctx, 200*time.Millisecond)
```

`WithTimeout(ctx, d)` 等价于 `WithDeadline(ctx, time.Now().Add(d))`。注意 `defer cancel()` 不可省（见常见坑）。

3. 取消沿调用链传播：

```go
go serviceA(ctx, done) // A -> B -> C 三级调用，共享同一个 ctx
...
cancel() // 在最外层取消一次，整条链路全部收到信号
```

`serviceA` 不新建 context，直接把父级 ctx 透传给 `serviceB`——取消信号沿 ctx 树自动传播，**一次 cancel 全链退出**。这就是 context 存在的意义：你不再需要手工在每一级之间穿一根 done channel。

**常见坑/易错点**

- ⚠️ **cancel 必须调用**（哪怕不打算提前取消），通常写成 `defer cancel()`，否则 context 关联的资源会泄漏到父 context 取消为止（示例 1 注释）。
- ⚠️ **超时只负责"通知"**：任务方必须主动监听 `ctx.Done()` 才能真正停下来（示例 2 注释）；一个不理会 ctx 的死循环任务，超时了也照样跑。
- ⚠️ **不要把 context 藏进结构体字段**：函数首参数命名为 `ctx` 并透传，是 Go 社区强约定（示例 4 注释）。
- ⚠️ **WithValue 只放请求范围元数据**（request-id、trace-id、认证身份等），不要塞业务结构体或可选参数——那会隐藏依赖、破坏类型安全；key 要用自定义类型（如文件中的 `ctxKey`）避免包间冲突（示例 5 注释）。

**动手练习**

1. 把示例 2 的 `slowJob` 时长改为 30ms（小于 50ms 超时），观察返回值从 `DeadlineExceeded` 变为 `nil`，理解"超时是上限而非必然"。
2. 写一个三层调用 `a -> b -> c`，每一层都在收到 `ctx.Done()` 时打印自己的名字，用 `WithTimeout(100ms)` 触发，验证三级都退出且无 goroutine 泄漏（可用 `runtime.NumGoroutine` 前后对比）。

---

### 08_generics.go：泛型（类型参数）

**这个文件带你认识什么**

Go 1.18 引入的泛型让"同一份逻辑服务多种类型"不再需要复制粘贴或牺牲类型安全的 `any` + 反射。本文件覆盖：泛型函数与类型推断、自定义约束接口（类型并集）、`comparable` 与 `cmp.Ordered` 内置约束、`~` 近似约束，以及泛型类型 `Stack[T]` 的完整实现。

**运行命令与预期输出解读**

```powershell
go run 02-intermediate/08_generics.go
```

输出分五个段落：

- **示例 1**：`First` 分别作用于 int 切片、string 切片并打印首元素，空切片打印 `ok = false`。
- **示例 2**：`Sum` 对 int 切片求和得 6，对 float64 切片求和得 4。
- **示例 3**：`MaxValue` 比较数字与字符串，`IndexOf` 在字符串切片中查找到索引 1。
- **示例 4**：自定义类型 `MyInt(21)`、普通 `int(21)`、`float64(1.25)` 都被 `Double` 翻倍（42、42、2.5）——`~` 约束让 `MyInt` 也能通过编译。
- **示例 5**：`Stack[int]` 压入 1、2、3，打印长度 3，然后按后进先出顺序弹出 3、2、1，空栈再 Pop 打印 `ok=false`。

输出顺序完全确定。

**关键代码讲解**

1. 泛型函数与零值技巧：

```go
func First[T any](s []T) (T, bool) {
	var zero T // 零值技巧：切片为空时返回该类型的零值
	if len(s) == 0 {
		return zero, false
	}
	return s[0], true
}
```

`[T any]` 声明类型参数；`var zero T` 得到 T 对应类型的零值，这是泛型代码里处理"没有值可返回"的惯用手法。调用端 `First([]int{10, 20, 30})` 由编译器推断出 `T=int`，通常无需手写 `First[int]`。

2. 约束接口——把运行期错误提前到编译期：

```go
type Number interface {
	int | int64 | float64
}

func Sum[T Number](s []T) T {
	var total T
	for _, v := range s {
		total += v // 约束保证了 + 运算符可用
	}
	return total
}
```

`|` 是类型并集：约束列出了允许的类型集合。因为集合内所有类型都支持 `+`，函数体内才能用加法；传 `[]string` 会在**编译期**报错——这正是约束的价值。

3. `~` 近似约束接纳自定义底层类型：

```go
type MyInt int

func Double[T ~int | ~float64](v T) T {
	return v * 2
}
```

`MyInt` 底层类型是 int 但它**不是 int**——约束写 `int` 时 `MyInt` 不满足；`~int` 表示"底层类型为 int 的类型均可"。这是约束里使用自定义具名类型的必备语法。

4. 泛型类型的方法接收者要带类型参数：

```go
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}
```

注意接收者写作 `*Stack[T]`（声明处的 `T`），而不是 `*Stack[any]`。实例化时给出类型实参：`Stack[int]{}`。

**常见坑/易错点**

- ⚠️ **推断不出类型时要显式写类型实参**，如 `First[int](nil)`（示例 1 注释）。
- ⚠️ **`comparable` 只允许 `==` / `!=`**；需要 `<`、`>` 时用 `cmp.Ordered`。切片、map、函数不可比较，不能作为 `comparable` 的实参（示例 3 注释）。
- ⚠️ **`~` 只能用于约束接口**，不能用于变量类型声明；自定义类型不会继承底层类型的方法，但会继承运算符能力（示例 4 注释）。
- ⚠️ **Go 泛型不支持方法的独立类型参数**——不能在已有 `[T]` 类型的方法上再声明 `[U any]`，需要时把 U 提到类型层面或改写成独立泛型函数（示例 5 注释）。

**动手练习**

1. 给 `Stack[T]` 增加 `Peek() (T, bool)` 方法：返回栈顶但不弹出，空栈返回零值与 false。提示：参照 `Pop`，只是不缩切片。
2. 写一个 `Keys[K comparable, V any](m map[K]V) []K` 泛型函数，返回 map 的所有键。提示：两个类型参数各自有自己的约束。

---

### 09_json_io.go：JSON 编解码与文件 IO

**这个文件带你认识什么**

这是偏工程实战的文件：`encoding/json` 的 Marshal/Unmarshal 与 struct tag 三件套（改名、`omitempty`、`"-"`）、`os.WriteFile` / `os.ReadFile` 一把梭的文件读写、`io.Reader` / `io.Writer` 接口如何让同一份代码处理不同数据来源，以及 `bufio.Scanner` 按行读取。

**运行命令与预期输出解读**

```powershell
go run 02-intermediate/09_json_io.go
```

输出分四个段落：

- **示例 1**：`User` 序列化成一行 JSON——`Age` 为零值被 `omitempty` 省略、`Email` 被 `"-"` 忽略、`Note` 无 tag 保持原名；随后 Unmarshal 回结构体并打印（`Email` 丢失、`Age` 回到零值）。
- **示例 2**：在系统临时目录创建子目录并写入 `note.txt`，打印文件路径（**每次运行的路径不同**）、读取的字节数与首行内容；`main` 返回时 `defer os.RemoveAll` 自动清理临时目录。
- **示例 3**：`io.Copy` 把 `strings.Reader` 的内容复制进 `bytes.Buffer` 并打印复制的字节数与内容；随后用 `json.NewEncoder` 直接把一个 User 编码进 buffer。
- **示例 4**：`bufio.Scanner` 按行打印 apple、banana、cherry 三行（带行号）。

**关键代码讲解**

1. struct tag 控制序列化行为：

```go
type User struct {
	Name  string `json:"name"`          // tag 改名：JSON 里叫 name 而不是 Name
	Age   int    `json:"age,omitempty"` // omitempty：零值时不输出该字段
	Email string `json:"-"`             // "-"：序列化时彻底忽略（敏感字段常用）
	Note  string // 无 tag：默认使用字段名 Note
}
```

三个要点：字段必须**导出**（大写开头）否则 `encoding/json` 看不到；`omitempty` 让零值字段不出现在输出里（注意"零值被省略"意味着 0、""、false 都无法与"未提供"区分）；`"-"` 彻底忽略，是密码等敏感字段的标配。

2. Unmarshal 必须传指针：

```go
var back User
err = json.Unmarshal(data, &back) // 必须传指针，否则无法写回
```

`json.Unmarshal` 要修改你提供的变量，传值的话它改的只是副本（且会报错）。

3. io.Reader / io.Writer 的接口价值：

```go
src := strings.NewReader("hello io interface")
var buf bytes.Buffer         // bytes.Buffer 实现 io.Writer
n, err := io.Copy(&buf, src) // Copy 不关心两端具体类型
```

`strings.NewReader`、`os.File`、网络连接都实现 `io.Reader`，所以同一个 `io.Copy` 能处理完全不同的数据来源——这是 01_interfaces.go"小接口"思想在标准库里的最佳示范。

4. bufio.Scanner 按行读取：

```go
sc := bufio.NewScanner(strings.NewReader(text))
for sc.Scan() { // 每读一行返回 true，读到 EOF 返回 false
	lineNo++
	fmt.Printf("  第 %d 行: %s\n", lineNo, sc.Text())
}
check(sc.Err())
```

`Scan()` 推进一行、`Text()` 取当前行；**循环结束后必须检查 `sc.Err()`**，因为 EOF 是正常结束，而 IO 错误藏在 `Err()` 里。

**常见坑/易错点**

- ⚠️ **JSON 中的未知字段默认被忽略**，多出的字段不会报错；需要严格校验时改用 `Decoder.DisallowUnknownFields()`（示例 1 注释）。
- ⚠️ **`os.WriteFile` 直接覆盖已存在文件**；需要追加或流式写入时改用 `os.OpenFile` + 标志位（示例 2 注释）。
- ⚠️ **Scanner 循环结束后必须检查 `Err()`**；且单行超过 64KB 会被默认缓冲区截断并报错，超长行要用 `sc.Buffer()` 扩容或改用 `bufio.Reader.ReadString`（示例 4 注释）。
- **字段不导出则 json 完全看不到**：`name` 小写字段会被静默跳过，序列化结果里查无此字段。
- 文件中的 `check` 是教学用的简化错误处理（打印后 `os.Exit(1)`）；真实项目应逐处判断并用 `%w` 包装错误。

**动手练习**

1. 把示例 1 的 `User` 改为带 `Password string \`json:"-"\`` 字段，序列化后确认输出中不含密码；再 Unmarshal 一段包含未知字段的 JSON，观察默认不报错的行为。
2. 修改 `scannerDemo`：把文本换成 `"a b c\nd e\n"`，用 `sc.Split(bufio.ScanWords)` 改成按单词切分，打印每个单词。

---

### 10_testing.go + 10_testing_test.go：go test 测试入门

**这个文件带你认识什么**

这是本阶段唯一的"双文件"组合：`10_testing.go` 是被测代码（`Add`、`Divide`、`IsPalindrome`、`WordCount` 四个函数加一个演示 main），`10_testing_test.go` 是对它的测试。合起来演示 Go 测试体系的四大件：**表驱动测试**、**t.Run 子测试**、**Benchmark 基准测试**、**Example 示例测试**。

**运行命令与预期输出解读**

先跑演示 main（在根目录）：

```powershell
go run 02-intermediate/10_testing.go
```

输出五个段落：Add 的两组加法结果；Divide 正常除法 2.5 与除零错误信息；`IsPalindrome` 对 `"level"`、`"上海海上"`（true）和 `"golang"`（false）的判断；`WordCount` 对普通句子、连续空白、空串、含换行符文本的计数；最后一个"手工模拟回归"演示为什么需要测试框架——3/3 通过后，提示把真正的回归交给 `go test`。

再跑测试（必须**进入 02-intermediate 目录**，PowerShell 中注意引号）：

```powershell
cd 02-intermediate
go test -v 10_testing.go 10_testing_test.go
```

`-v` 输出中可以看到：

- `TestAdd`、`TestIsPalindrome`、`TestWordCount` 各自展开为多个**命名子测试**（如 `TestAdd/正数相加`），全部 PASS。
- `TestDivide` 的两个子测试分别验证正常路径与除零错误路径。
- 两个 `Example` 函数通过 `// Output:` 注释与实际输出比对，一致则 PASS。
- 基准测试默认不运行，需要显式开启：

```powershell
go test "-bench=." "-benchtime=10x" 10_testing.go 10_testing_test.go
```

输出会列出 `BenchmarkIsPalindrome` 与 `BenchmarkWordCount` 各迭代 10 次及每次操作的纳秒数（ns/op）。

**关键代码讲解**

1. 表驱动 + 子测试——Go 测试的主流写法：

```go
cases := []struct {
	name string
	a, b int
	want int
}{
	{"正数相加", 2, 3, 5},
	{"负数相加", -1, -1, -2},
	{"正负相抵", 5, -5, 0},
	{"加零不变", 42, 0, 42},
}
for _, c := range cases {
	t.Run(c.name, func(t *testing.T) {
		if got := Add(c.a, c.b); got != c.want {
			t.Errorf("Add(%d, %d) = %d, 期望 %d", c.a, c.b, got, c.want)
		}
	})
}
```

把输入/期望放进匿名结构体切片，一条循环全部跑完；`t.Run` 生成命名子测试，失败时报告精确到用例名，还能用 `-run TestAdd/正数相加` 单独运行。新增用例只需加一行数据。

2. 测试错误路径——检查 err 而非文案：

```go
t.Run("除零报错", func(t *testing.T) {
	if _, err := Divide(1, 0); err == nil {
		t.Error("Divide(1, 0) 应返回错误，却得到 nil")
	}
})
```

⚠️ 重点检查 `err` 是否为 nil，而**不要断言错误文案**——文案是易变的实现细节（与 02_errors.go 的"用 errors.Is 而非比较文案"一脉相承）。

3. 基准测试的 b.N 约定：

```go
func BenchmarkIsPalindrome(b *testing.B) {
	s := "上海海上"
	for i := 0; i < b.N; i++ {
		IsPalindrome(s)
	}
}
```

`Benchmark` 前缀 + `b.N` 循环：框架会自动调整 N 以获得稳定的耗时数据（本教学命令用 `-benchtime=10x` 固定为 10 次以便快速看结果）。

4. Example 既是文档又是测试：

```go
func ExampleIsPalindrome() {
	fmt.Println(IsPalindrome("level"))
	fmt.Println(IsPalindrome("golang"))
	// Output:
	// true
	// false
}
```

末尾的 `// Output:` 注释与程序实际输出**必须完全一致**（含空格与换行），否则测试失败；没有 Output 注释的 Example 只编译不运行。

**常见坑/易错点**

- ⚠️ **`t.Errorf` 与 `t.Fatalf` 的区别**：`Errorf` 记录失败但**继续跑**后续用例；`Fatalf` **立即中止**本测试（`TestDivide` 中"意外错误"用 `Fatalf`，因为 err 非 nil 时后续断言没有意义）。
- ⚠️ **断言错误文案很脆弱**：重点查 `err == nil` 或用 `errors.Is`（示例 2 注释）。
- ⚠️ **Example 的输出比较是逐字节精确的**：多余空格、换行都会导致失败；没有 `// Output:` 的 Example 只编译不运行（文件末尾注释）。
- ⚠️ **本项目的测试文件带 `//go:build ignore`**：所以要用显式文件列表 `go test -v 10_testing.go 10_testing_test.go`；真实项目中测试文件不加 ignore，直接 `go test ./...` 即可（10_testing.go 末尾注释）。
- **PowerShell 中 `-bench=.` 要加引号**，见"环境准备与运行方式"一节。

**动手练习**

1. 为 `Divide` 表驱动化：把 `TestDivide` 改写成包含多组 `{a, b, want, wantErr bool}` 的表驱动测试，覆盖正常、除零、负数三组用例。
2. 给 `WordCount` 写一个 `ExampleWordCount`：打印 `WordCount("go is fun")`，并补全正确的 `// Output:` 注释。提示：先心算期望输出，再跑 `go test -run ExampleWordCount -v` 验证。

---

## 五、本阶段小结与进入下一阶段条件

**小结**

中级阶段的主线是 Go 的三大设计哲学：**接口**让你用隐式实现和小接口组合写出低耦合代码；**错误处理**用显式返回值 + 错误链（`%w` / `errors.Is` / `errors.As`）替代异常，让错误路径清晰可见；**并发**则提供了从 goroutine、channel、select 到 sync 锁、context 取消传播的一整套工具。配合闭包、泛型、JSON/IO 与测试这四块拼图，你已经具备了阅读和编写真实 Go 项目代码的完整基础。

**进入高级阶段（03-advanced）的条件**——请逐条自验：

1. 能不假思索地写出 comma-ok 类型断言（`if s, ok := v.(string); ok { ... }`），并说出单返回值断言失败会发生什么。
2. 能解释 nil channel 在 select 中的行为（永远不就绪），并举出一个它的正面用途。
3. 能说出 `errors.Is` 与 `errors.As` 的分工，并解释为什么 `var qe *QueryError = nil; err = qe` 之后 `err != nil`。
4. 能解释为什么 `for range` 遍历 channel 需要发送方 `close`，以及向已关闭 channel 发送数据的后果。
5. 能为任意一个纯函数写出带命名子测试的表驱动测试，并在 PowerShell 中正确运行它。

以上任何一条答不上来，建议回到对应文件重读并完成其中的动手练习。

---

## 六、自测题（5 题，附答案或提示）

**第 1 题（接口断言）**

下面代码的输出是什么？如果把断言改成单返回值形式 `s := v.(string)` 会怎样？

```go
var v any = 42
if s, ok := v.(string); ok {
	fmt.Println("字符串：", s)
} else {
	fmt.Println("不是 string，ok =", ok)
}
```

答案：打印"不是 string，ok = false"。comma-ok 形式断言失败不 panic，`s` 得到 string 零值 `""`。改成单返回值形式 `s := v.(string)` 时，因为 `v` 的动态类型是 int，程序会**直接 panic**（提示：回顾 01_interfaces.go 的 ⚠️ 注释）。

**第 2 题（errors.Is / errors.As）**

`err := fmt.Errorf("查询失败: %w", ErrNotFound)`，以下两种检查分别该用哪个函数？为什么？

1. 判断 err 的根源是不是 `ErrNotFound`；
2. 从 err 中取出 `*QueryError` 的 `Code` 字段。

答案：① 用 `errors.Is(err, ErrNotFound)`——它沿 `%w` 形成的错误链逐层比较"是不是某个错误"；② 用 `errors.As`，如 `var qe *QueryError; if errors.As(err, &qe) { ... }`——它沿链查找特定**类型**的错误并把值填出。注意第二个参数必须是非 nil 指针（提示：回顾 02_errors.go 示例 3）。

**第 3 题（闭包捕获）**

下面代码在 Go 1.22+（本项目为 1.27）下打印什么？在 Go 1.21 及更早版本下呢？

```go
funcs := make([]func(), 0, 3)
for i := 0; i < 3; i++ {
	funcs = append(funcs, func() { fmt.Print(i, " ") })
}
for _, f := range funcs {
	f()
}
```

答案：Go 1.22+ 打印 `0 1 2`——每次迭代的 `i` 都是新变量，每个闭包捕获各自的那一份；Go 1.21 及更早版本打印 `3 3 3`——循环变量是同一个变量被反复赋值，闭包按引用捕获，循环结束后读到的都是最终值（提示：回顾 03_closures.go 示例 5；旧代码的修复写法是 `i := i` 或把 i 作为参数传入闭包）。

**第 4 题（channel 关闭与 range）**

下面代码会正常结束、死锁还是 panic？为什么？

```go
ch := make(chan int, 2)
ch <- 1
ch <- 2
close(ch)
for v := range ch {
	fmt.Println(v)
}
```

答案：**正常结束**，依次打印 1 和 2。`close` 的语义是"不再发送"，缓冲区里已发送的值仍可取出；`for range` 在 channel 关闭且耗尽后自动退出循环。若去掉 `close(ch)`，range 循环在取完两个值后永远等待新值——所有 goroutine 都睡着，Go 运行时报 `fatal error: all goroutines are asleep - deadlock!`（提示：回顾 05_channels.go 示例 3、4）。

**第 5 题（context 超时）**

下面代码中 `slowJob` 返回什么错误？如果把 `WithTimeout` 的时长改成 300ms 呢？另外，`defer cancel()` 删掉会有什么后果？

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()
err := slowJob(ctx, 200*time.Millisecond) // slowJob 内部 select 监听 time.After(cost) 与 ctx.Done()
```

答案：50ms 超时时，ctx 先到期，`slowJob` 从 `ctx.Done()` 分支返回 `context.DeadlineExceeded`；改成 300ms 时任务先完成，返回 `nil`。删掉 `defer cancel()` 后，即使任务正常完成，context 关联的资源也要等到父 context 取消才会释放——这就是资源泄漏，所以 `cancel` 必须调用（提示：回顾 07_context.go 示例 1、2 的 ⚠️ 注释；超时只负责"通知"，任务方必须主动监听 `ctx.Done()`）。
