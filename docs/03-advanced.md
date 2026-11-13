# Go 高级阶段：底层机制与工程实践

> 适用项目：go-learning（根目录 `D:\Project\gogogo`）
> 本阶段对应目录：`03-advanced/`
> 这是本课程的最后一个阶段，聚焦 Go 的底层机制（反射、调度器、GC、unsafe）与工程实践（并发模式、HTTP 服务与客户端、日志、性能优化、设计模式）。

---

## 一、本阶段学习目标

完成本阶段后，你应当能够：

1. **理解反射的能力边界与代价**：会用 `reflect.TypeOf` / `reflect.ValueOf` 读取类型与值、遍历结构体字段与方法、读取 struct tag、通过可寻址 Value 修改变量，并清楚反射在热路径上的性能代价。
2. **掌握常见并发模式**：能手写 worker pool、pipeline、fan-out/fan-in、Ticker 限流，以及一个带"错误收集 + 失败即取消"的简易 errgroup，并保证程序无 goroutine 泄漏。
3. **理解 GMP 调度模型与 GC 基本机制**：知道 G/M/P 各自的角色，会用 `GOMAXPROCS`、`NumGoroutine`、逃逸分析与 `debug.SetGCPercent` 观察和调优运行时行为。
4. **写出健壮的 HTTP 服务与客户端**：会配置 `http.Server` 的三个超时、组合中间件链、用 `Shutdown(ctx)` 优雅关闭；客户端侧会设置超时、用 context 取消请求、实现指数退避重试、复用连接。
5. **了解 unsafe 与性能优化手段**：理解内存对齐与结构体布局，知道 string↔[]byte 零拷贝的原理与风险，会用 `sync.Pool`、切片预分配、pprof 做基础的性能优化。
6. **掌握 Go 风格的设计模式**：函数式选项、`sync.Once` 单例、简单工厂、装饰器/中间件、依赖注入，以及"小接口、使用方定义、接受接口返回结构体"的接口设计原则。

---

## 二、环境准备与运行方式

### Go 版本要求

- 本项目按 **Go 1.27.1** 编写，**Go 1.22+** 即可运行全部示例。
- 本阶段用到的 `log/slog`、`slices`、`maps`、`cmp` 均为**标准库**（其中 `maps.Keys` 返回迭代器属于 Go 1.23+ 的 range-over-func 风格，建议 1.23+）。
- 项目零第三方依赖，无需 `go mod download`。

### 运行方式

每个 `.go` 文件都带 `//go:build ignore` 标签，因此 `go build ./...` 会跳过它们，但可以直接用 `go run` 指定文件独立运行。**在项目根目录下**执行：

```bash
go run 03-advanced/01_reflection.go
go run 03-advanced/02_concurrency_patterns.go
go run 03-advanced/03_runtime_scheduler.go
go run 03-advanced/04_http_server.go
go run 03-advanced/05_http_client.go
go run 03-advanced/06_log_slices_maps.go
go run 03-advanced/07_embed_generate.go
go run 03-advanced/08_performance.go
go run 03-advanced/09_unsafe.go
go run 03-advanced/10_design_patterns.go
```

### 特别说明

- **07_embed_generate.go** 依赖同目录下的 `assets/message.txt`（`//go:embed` 嵌入资源）。`go run` 单文件也支持 embed，无需额外参数，直接在项目根目录运行即可。
- **逃逸分析观察**：`03-advanced/03_runtime_scheduler.go` 文件内注释给出的方法是——由于本文件带 `//go:build ignore` 标签，`go build` 默认会跳过它，需要**把文件复制到无标签的临时目录**，再执行 `go build -gcflags="-m" 03_runtime_scheduler.go`，在编译输出中查看 `escapes to heap`（逃逸到堆）与 `does not escape`（留在栈上）的决策。以文件内注释为准。
- **04_http_server.go / 05_http_client.go 均为本地回环**：04 监听 `127.0.0.1` 的随机端口并用 `http.Client` 自测；05 使用 `httptest.NewServer`（只监听 127.0.0.1）。两者都**不需要外网**，演示完成后程序自行退出，不会常驻。
- **go:generate 说明**：07 文件顶部登记了 `//go:generate gofmt -w $GOFILE`，仅作语法演示，课程不要求也不建议实际执行。

---

## 三、文件清单表

| 文件名 | 主题 | 核心知识点 | 建议阅读顺序 |
| --- | --- | --- | --- |
| 01_reflection.go | 反射 | `TypeOf`/`ValueOf`、Kind 与 Type 的区别、遍历结构体字段与方法、可寻址 Value 修改、struct tag、反射性能代价 | 1 |
| 02_concurrency_patterns.go | 并发模式 | worker pool、pipeline、fan-out/fan-in、Ticker 限流、手写 errgroup + context 优雅退出 | 2 |
| 03_runtime_scheduler.go | 运行时与调度器 | GMP 模型、`Gosched`、`GOMAXPROCS`、`NumGoroutine` 泄漏观察、逃逸分析、`debug.SetGCPercent` | 3 |
| 04_http_server.go | HTTP 服务器 | `http.Server` 三超时、Handler/HandlerFunc、中间件链（logging/recover/timing）、ServeMux、`Shutdown(ctx)` 优雅关闭 | 4 |
| 05_http_client.go | HTTP 客户端 | Client Timeout、context 取消、指数退避重试、自定义请求头、JSON 解析、连接复用 | 5 |
| 06_log_slices_maps.go | 结构化日志与泛型工具包 | `slog` Text/JSON Handler、`With` 派生 logger、`slices`/`maps`/`cmp` 泛型工具、错误+日志最佳实践 | 6 |
| 07_embed_generate.go | embed、generate 与 build tags | `//go:embed` 嵌入 string/[]byte/embed.FS、`//go:generate` 指令、`//go:build` 条件编译 | 7 |
| 08_performance.go | 性能优化入门 | Benchmark 思想、`runtime/pprof` CPU profile、逃逸分析、`sync.Pool`、切片容量预分配 | 8 |
| 09_unsafe.go | unsafe 包入门 | `unsafe.Pointer` 与 `uintptr`、`Sizeof`/`Alignof`/`Offsetof`、字段重排省内存、string↔[]byte 零拷贝 | 9 |
| 10_design_patterns.go | Go 常见设计模式 | 函数式选项、`sync.Once` 单例、简单工厂、装饰器/中间件、依赖注入与接口设计原则 | 10 |

> 注：`03-advanced/assets/message.txt` 是 07 的嵌入资源文件，不是课程文件，故不列入表格（详见 07 的详解）。

---

## 四、逐文件详解

### 01_reflection.go：反射

**这个文件带你认识什么**

反射是 Go 在运行时"检视和操作自身类型"的能力。文件通过 6 个示例演示：`TypeOf`/`ValueOf` 的基本用法、`Kind` 与 `Type` 的概念差异、遍历结构体字段与方法、通过可寻址 Value 安全修改变量、读取 struct tag，以及反射调用与直接调用的性能对比。

**运行命令与预期输出解读**

```bash
go run 03-advanced/01_reflection.go
```

输出分为 6 个段落：

1. **示例 1**：打印 `3.14` 的 Type（`float64`）、Kind（`float64`）、Value，以及一个 nil `*int` 指针的 `IsValid()`/`IsNil()` 结果。
2. **示例 2**：对比自定义类型 `Celsius`（底层 `float64`）与 `float64`——两者 Type 不同但 Kind 相同。
3. **示例 3**：遍历 `User` 结构体的字段（`Name`、`Age` 可导出并显示值，小写 `email` 显示 `<不可读取>`）与方法（`Greet`），并反射调用 `Greet`。
4. **示例 4**：通过 `reflect.ValueOf(&u).Elem()` 把 `Age` 改为 20、`Name` 改为 "小红"，并显示 `email` 字段的 `CanSet` 为 false。
5. **示例 5**：逐字段打印 `json` 与 `validate` 两个 struct tag 的值。
6. **示例 6**：分别计时 100 万次直接调用与反射调用。耗时数值每次运行可能不同，取决于机器负载；直接调用太快时耗时可能为 0，此时不会输出倍数而是输出一条说明文字。

**关键代码讲解**

```go
t := reflect.TypeOf(x)   // 动态类型信息（类型元数据）
v := reflect.ValueOf(x)  // 动态值信息（运行时的具体值）
```

`TypeOf` 和 `ValueOf` 是反射的两个入口：前者回答"这是什么类型"，后者回答"值是什么"。`Kind()` 返回的是底层类别（如 `reflect.Float64`），而 `Type` 是精确类型——自定义类型 `Celsius float64` 与 `float64` 的 Type 不同、Kind 相同。因此**用 switch 做类型判断时若想兼容自定义类型，应比较 Kind 而非 Type**。

```go
v := reflect.ValueOf(&u).Elem()
age := v.FieldByName("Age")
if age.IsValid() && age.CanSet() {
    age.SetInt(20)
}
```

`reflect.ValueOf(u)` 得到的是结构体的**副本**，不可寻址，对它调用 `Set` 会 panic。必须传入指针再 `Elem()`，得到的 Value 才指向原变量、可寻址、可修改。修改前先检查 `IsValid() && CanSet()`，且 `SetInt`/`SetString` 的类型必须与字段真实类型匹配，否则同样 panic。

```go
fmt.Printf("字段 %-6s json=%-8q validate=%q\n",
    f.Name, f.Tag.Get("json"), f.Tag.Get("validate"))
```

`StructTag.Get(key)` 按 key 读取 tag 值。`encoding/json`、参数校验库正是靠这个机制在运行时读取 tag 来决定序列化字段名与校验规则——这是反射最经典的生产用途。

**常见坑/易错点**

1. **对零 Value 调用方法会 panic**：`reflect.ValueOf(nil)` 得到的是零 Value，文件用 `IsValid()` 先判断；nil 指针则可用 `IsNil()` 判断。
2. **不可导出字段不可读写**：小写字段（如 `email`）调用 `Interface()` 会 panic，必须先用 `f.IsExported()` 判断；其 `CanSet` 也是 false，强行 `Set` 会 panic。
3. **副本不可寻址**：直接 `reflect.ValueOf(u)` 后 `Set` 会 panic，必须 `&u` + `Elem()`。
4. **反射有性能代价**：`Call` 需要装箱参数、检查类型、再拆箱结果，比直接调用慢一个数量级以上。结论：热路径避免反射；配置解析、序列化等低频场景可放心使用。

**动手练习**

1. 给 `User` 增加一个导出字段 `City string`（带 `json:"city"` tag），运行程序观察示例 3、5 的输出变化，并用反射把它设为 "北京"。
   - 提示：复用示例 4 中 `FieldByName` + `CanSet` + `SetString` 的写法。
2. 修改示例 6，把 `n` 改为 `10_000_000`，观察直接调用与反射调用的耗时差距是否扩大。
   - 提示：若直接调用耗时为 0，代码已有分支处理，不会 panic。

---

### 02_concurrency_patterns.go：并发模式

**这个文件带你认识什么**

文件用 5 个示例串起 Go 工程中最常用的并发套路：固定工人数的 worker pool、生成→加工→汇出的 pipeline、fan-out/fan-in 的分发与合并、`time.Ticker` 限流，以及一个手写的简易 errgroup（WaitGroup + 错误收集 + 失败即取消），最后用 `runtime.NumGoroutine()` 校验无 goroutine 泄漏。

**运行命令与预期输出解读**

```bash
go run 03-advanced/02_concurrency_patterns.go
```

输出分 5 段：

1. **worker pool**：3 个 worker 处理 8 个任务，打印 "worker N 处理任务 M"。worker 与任务的对应关系、打印顺序每次运行可能不同，但最后的**平方和恒为 204**（1²+…+8²），程序确定性退出。
2. **pipeline**：生成 1~5、乘 10 后汇出，输出 `10 20 30 40 50`。
3. **fan-out/fan-in**：同一输入分发给 3 个 worker 各自乘 2 再合并，汇总结果恒为 42，输出顺序可能不同。
4. **Ticker 限流**：每 100ms 放行一个任务，打印 4 个任务的执行时刻（约 100ms、200ms、300ms、400ms），具体时间每次运行略有不同。
5. **errgroup**：任务 1 无限干活，任务 2 在第 3 步失败 → 整组被取消 → 任务 1 打印"收到取消信号，清理后退出"，最终打印第一个错误，且存活 goroutine 数应接近 1。

**关键代码讲解**

```go
for w := 1; w <= workerCount; w++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        for j := range jobs { // 通道关闭且取空后自动退出循环
            results <- j * j
        }
    }(w)
}
```

worker pool 的核心是**固定数量的 goroutine 消费无缓冲任务通道**。任务再多也不新增 goroutine，这就是"背压"：生产快于消费时，发送方自然阻塞在 `jobs <- j` 上，任务不会被无限堆积。发完任务必须 `close(jobs)`，worker 的 `range` 才能退出循环；另有 goroutine 在 `wg.Wait()` 后 `close(results)`，让 main 的 `range results` 也能结束——**两层关闭缺一不可**。

```go
go func() {
    wg.Wait()
    close(out) // 必须等所有输入排空才能关闭
}()
```

fan-in 把多个输入通道合并成一个输出通道：每个输入配一个 goroutine 转发到 `out`，全部转发完（`wg.Wait()`）才能关闭 `out`。**提前关闭会导致正在发送的 goroutine panic**。

```go
g.once.Do(func() { // 只记录第一个错误，避免并发写竞争
    g.err = err
    g.cancel() // 任意任务失败即取消
})
```

简易 errgroup 的三个要素：`WaitGroup` 等待全部任务、`sync.Once` 保证只记录第一个错误（并发写 `err` 会数据竞争）、`cancel()` 把失败信号通过 context 广播给其他任务。其他任务在 `select { case <-ctx.Done(): ... }` 中感知取消并清理退出。

**常见坑/易错点**

1. **发完任务忘记 `close(jobs)`**：worker 永远阻塞在 `range jobs` 上，程序死锁。文件注释明确提醒"发完必须关闭"。
2. **`time.Ticker` 必须 `Stop`**：否则底层定时器不释放，造成资源泄漏（文件用 `defer limiter.Stop()`）。
3. **fan-in 提前关闭输出通道**：`wg.Wait()` 之前就 `close(out)`，正在发送的 goroutine 会 panic。
4. **goroutine 泄漏**：示例 5 最后用 `runtime.NumGoroutine()` 验证——如果取消信号设计有漏洞，存活 goroutine 数会明显高于 1。泄漏是并发代码最隐蔽的 bug。

**动手练习**

1. 把 worker pool 的 `workerCount` 从 3 改为 1 再改为 8，观察输出顺序和总耗时的变化，确认平方和始终是 204。
   - 提示：结果与并发度无关，这正是"结果确定、顺序不确定"的好例子。
2. 给 pipeline 的 `gen` 阶段加上限流（参考示例 4 的 `time.Ticker`），让数字每 50ms 生成一个。
   - 提示：在 `out <- i` 之前 `<-limiter.C`。

---

### 03_runtime_scheduler.go：运行时与调度器

**这个文件带你认识什么**

文件用"注释讲解 + 小实验"的方式介绍 Go 运行时：GMP 调度模型中 G/M/P 的角色、`runtime.Gosched` 主动让出、`GOMAXPROCS` 的观察与调整、用 `NumGoroutine` 演示泄漏与修复、用 `testing.AllocsPerRun` 量化逃逸带来的堆分配，以及用 `debug.SetGCPercent` 观察 GOGC 对 GC 行为的影响。

**运行命令与预期输出解读**

```bash
go run 03-advanced/03_runtime_scheduler.go
```

输出分 6 段：

1. **GMP 模型**：打印 CPU 核数（`NumCPU`）、当前 P 数量（`GOMAXPROCS`）、启动 5 个 goroutine 前后的 G 数量。
2. **Gosched**：协程 A 与 main（协程 B）交替打印各 3 次。输出顺序每次运行可能不同——`Gosched` 只是"建议"调度器换人，不保证严格交替。
3. **GOMAXPROCS**：把 P 数量调为 1 再恢复，演示读取与修改。
4. **NumGoroutine**：故意让 3 个 goroutine 阻塞在永不写入的通道上（模拟泄漏），G 数量 +3；`close(leak)` 放行后回落。
5. **逃逸分析**：`stackAlloc` 每次堆分配 0 次（不逃逸），`heapAlloc` 每次 1 次（返回局部变量地址，逃逸到堆）。
6. **GOGC**：分别在 GOGC=100 与 GOGC=400 下制造垃圾并手动 `runtime.GC()`，打印累计 GC 次数与当前堆大小（KB）。数值每次运行可能不同。

**关键代码讲解**

```go
// G（Goroutine）：goroutine 本身，保存栈、指令指针等执行现场。
// M（Machine）：  OS 线程的封装，真正在 CPU 上跑代码的实体。
// P（Processor）：逻辑处理器，持有一个可运行 G 的本地队列；
//     M 必须绑定一个 P 才能执行 G。P 的数量 = GOMAXPROCS。
```

GMP 是理解 Go 并发的钥匙：**G 是任务，M 是工人（OS 线程），P 是工位**。M 必须绑定 P 才能执行 G；P 的本地队列空了会去"偷"（work stealing）其他 P 的 G。所以 G 可以有成千上万个，M/P 只有几个，靠调度器分时复用。

```go
stackAllocs := testing.AllocsPerRun(1000, func() { _ = stackAlloc() })
heapAllocs := testing.AllocsPerRun(1000, func() { _ = heapAlloc() })
```

`testing.AllocsPerRun` 测量每次调用的平均**堆分配**次数（栈分配不计入）。`stackAlloc` 返回值本身，变量留在栈上（0 次堆分配）；`heapAlloc` 返回局部变量的地址（`//go:noinline` 防止内联干扰），栈帧销毁后变量仍需存活，只能逃逸到堆上（1 次堆分配）。逃逸越多，GC 压力越大。

```go
old := debug.SetGCPercent(100) // 默认值：堆长到上次 GC 后的 2 倍时触发
// ...
debug.SetGCPercent(400)        // 堆允许长更大才 GC → GC 更少但内存更高
debug.SetGCPercent(old)        // 用完恢复，避免影响后续
```

GOGC 是**内存与 CPU 的权衡**：调大则 GC 次数减少（省 CPU）但堆占用更高；调小则相反。`debug.SetGCPercent` 返回旧值，演示完毕应当恢复。

**常见坑/易错点**

1. **容器中 `NumCPU` 看到的是宿主机核数**：文件注释提醒，生产容器环境可能需要按 cgroup 配额显式设置 `GOMAXPROCS`，否则 P 数量远超配额，调度反而低效。
2. **goroutine 泄漏很难自查**：示例 4 演示了用 `NumGoroutine` 观察，文件同时提醒生产环境应使用 pprof 的 goroutine profile 定位泄漏。
3. **`debug.SetGCPercent(-1)` 会完全关闭 GC**：生产环境慎用，堆会无限增长直到 OOM。
4. **逃逸分析命令受 build tag 影响**：本文件带 `//go:build ignore`，直接 `go build -gcflags="-m"` 会跳过它，需复制到无标签的临时目录再执行（以文件内注释为准）。

**动手练习**

1. 在示例 1 中把 goroutine 数量从 5 改为 100，观察启动后 G 数量与程序行为的变化。
   - 提示：G 再多，M/P 数量不变，这就是"少量线程跑海量 goroutine"。
2. 写一个新函数返回 `[]int` 切片头部指针（如 `&s[0]`），用 `testing.AllocsPerRun` 验证它是否逃逸。
   - 提示：参照 `heapAlloc` 的写法，记得加 `//go:noinline`。

---

### 04_http_server.go：HTTP 服务器

**这个文件带你认识什么**

文件完整演示了一个"生产姿势正确"的 HTTP 服务器：`Handler` 接口与 `HandlerFunc` 适配器两种注册方式、logging/recover/timing 三层中间件的组合、`http.Server` 结构体的三个超时参数、随机端口监听，以及用 `Shutdown(ctx)` 优雅关闭。程序启动后用 `http.Client` 自测 4 条路由，然后自行关闭退出。

**运行命令与预期输出解读**

```bash
go run 03-advanced/04_http_server.go
```

输出分三段：

1. **启动**：打印服务器监听的地址（`127.0.0.1` + 随机端口，每次运行端口不同，仅本机可访问）。
2. **客户端自测**：依次请求 `/hello`（200）、`/time`（200，返回当前时间）、`/panic`（500，被 recover 中间件兜底）、`/none`（404，ServeMux 默认 NotFound）。同时能看到中间件的输出：`[log]` 记录方法与路径，`[recover]` 打印捕获的 panic，`[time]` 打印每个请求的处理耗时（耗时数值每次不同）。
3. **优雅关闭**：打印关闭成功的说明，程序退出。

**关键代码讲解**

```go
srv := &http.Server{
    Addr:    "127.0.0.1:0", // 仅本机回环 + 随机端口
    Handler: chain(mux, loggingMiddleware, recoverMiddleware, timingMiddleware),
    ReadTimeout:  5 * time.Second,  // 读完整个请求（含 body）的时限
    WriteTimeout: 5 * time.Second,  // 写完整个响应的时限
    IdleTimeout:  60 * time.Second, // keep-alive 空闲连接的保留时长
}
```

**永远不要用 `http.ListenAndServe` 的默认零超时配置上生产**——没有超时的服务器会被慢速攻击（Slowloris）轻松拖垮。三个超时各司其职：Read 限制读请求、Write 限制写响应、Idle 限制空闲连接保留时长。`127.0.0.1:0` 表示只监听本机回环并让内核分配随机端口，是教学与测试环境最安全的写法。

```go
func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
    for i := len(mws) - 1; i >= 0; i-- {
        h = mws[i](h)
    }
    return h
}
```

中间件的本质是 `func(http.Handler) http.Handler`。`chain` 从后往前包裹，所以**先写的中间件在最外层**：`logging` 最先执行，`recover` 在中间兜底，`timing` 最贴近业务。panic 会被 `recover` 捕获并返回友好的 500，而不是让连接异常断开。

```go
ln, err := net.Listen("tcp", srv.Addr) // 先 Listen 拿到随机端口
go func() {
    if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
        fmt.Println("服务器异常退出:", err)
    }
}()
// ...
srv.Shutdown(ctx) // 停止接收新连接，等待进行中的请求完成
```

两个工程要点：一是**先 `net.Listen` 再 `Serve`**，否则拿不到随机分配的真实端口；二是 `Shutdown` 触发后 `Serve` 返回 `http.ErrServerClosed`，这属于**正常流程**而非错误，必须用 `errors.Is` 排除。`Shutdown(ctx)` 与 `Close()` 的区别：前者等待存量请求处理完，后者立即强杀所有连接、可能丢响应。

**常见坑/易错点**

1. **不关闭响应 Body 会导致连接无法复用**：自测代码中每个请求都 `defer resp.Body.Close()`（详见 05 的连接复用前提）。
2. **依赖 net/http 自带的 recover 不够**：文件注释说明，标准库会为每个连接单独 recover，但响应会异常断开——自己写 recover 中间件才能返回友好的 500。
3. **漏配超时**：零值 `http.Server` 没有任何超时，面对慢客户端会累积大量连接。
4. **把 `ErrServerClosed` 当错误处理**：`Shutdown` 后 `Serve` 返回它是预期行为。

**动手练习**

1. 新增一条路由 `/version`，返回固定文本 `v1.0`，重新运行确认自测输出。
   - 提示：用 `mux.HandleFunc` 注册，并在 `selfTest` 里加一个 `get("/version")`。
2. 把中间件顺序改为 `chain(mux, recoverMiddleware, loggingMiddleware, timingMiddleware)`，思考 `/panic` 请求的输出顺序会发生什么变化。
   - 提示：先写的在最外层；panic 发生时，尚未进入的内层中间件不会执行其"后置"逻辑。

---

### 05_http_client.go：HTTP 客户端

**这个文件带你认识什么**

文件基于 `httptest.NewServer` 搭建本地测试服务器（只监听 127.0.0.1，绝不触网），演示客户端侧的五个关键技能：给 `http.Client` 配超时、用 context 取消请求、指数退避重试、自定义请求头并用 `json.Decoder` 解析响应，以及连接复用（keep-alive）的前提条件。

**运行命令与预期输出解读**

```bash
go run 03-advanced/05_http_client.go
```

输出分 5 段：

1. **Client 与 Timeout**：`GET /json` 返回 200；`GET /slow`（服务端模拟 5 秒慢响应）在 2 秒超时时按预期报错。
2. **context 取消请求**：300ms 的 `context.WithTimeout` 取消对 `/slow` 的请求，打印实际耗时（约 300ms，每次运行略有不同）；服务端也会打印"感知到客户端取消，提前结束"。
3. **指数退避重试**：`/flaky` 路由前 2 次返回 500、第 3 次成功。输出依次打印"第 1 次失败，100ms 后重试""第 2 次失败，200ms 后重试""第 3 次尝试成功"。
4. **请求头与 JSON 解析**：服务端回显自定义头 `X-Request-Id: req-42`；`/json` 的响应被解析为 `user{Name:"gopher", Age:14}`。
5. **连接复用**：连续 3 次请求 `/json`，说明底层 TCP 连接被复用。

**关键代码讲解**

```go
client := &http.Client{Timeout: 2 * time.Second}
```

**`http.Get` 使用的默认 Client 没有超时，可能永久挂起**——这是 Go HTTP 客户端最常见的生产事故来源。任何对外请求都必须显式构造带 `Timeout` 的 Client（或用 context 控制）。

```go
ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
defer cancel()
req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/slow", nil)
_, err = client.Do(req)
if errors.Is(err, context.DeadlineExceeded) { ... }
```

更精细的控制交给 context：`NewRequestWithContext` 把 ctx 绑到请求上，超时/取消时 `Do` 返回的错误可以用 `errors.Is(err, context.DeadlineExceeded)` 判定。服务端也能通过 `r.Context().Done()` 感知取消、提前释放资源——这是 context 端到端取消的威力。

```go
backoff := time.Duration(100<<attempt) * time.Millisecond // 100ms、200ms、400ms……
fmt.Printf("第 %d 次失败，%v 后重试\n", attempt+1, backoff)
time.Sleep(backoff)
```

指数退避用**位移** `100<<attempt` 实现等比数列，让失败的服务端有喘息时间，避免重试风暴。同时注意：失败响应也要 `resp.Body.Close()`，否则连接无法回收。

```go
io.Copy(io.Discard, resp.Body)
resp.Body.Close()
```

连接复用的两个前提——**读完 Body 且 Close**。只 Close 不读完，连接会被丢弃重新建立。另外 `http.Client` 并发安全，应全局复用同一个实例，不要每次请求都新建（会丢失连接池）。

**常见坑/易错点**

1. **默认 Client 无超时**：`http.Get` / `http.Post` 便捷方法可能永久挂起，务必自建带 `Timeout` 的 Client。
2. **忘记关闭或没读完 Body**：两者都会破坏连接复用，高并发下表现为大量 TIME_WAIT / 新建连接。
3. **重试不退避**：固定间隔甚至立即重试会放大服务端故障；指数退避是基本要求。
4. **自定义请求头必须用 `NewRequest`**：`client.Get` 便捷方法无法加头。

**动手练习**

1. 给重试逻辑加上"最大退避上限"（如封顶 1 秒），并尝试把 `/flaky` 改为前 4 次失败，观察重试上限 5 次是否够用。
   - 提示：`if backoff > time.Second { backoff = time.Second }`。
2. 新增一个 `/echo-method` 路由返回请求方法名，用 `http.NewRequest(http.MethodPost, ...)` 验证。
   - 提示：服务端用 `r.Method` 获取方法名。

---

### 06_log_slices_maps.go：结构化日志与泛型工具包

**这个文件带你认识什么**

文件分两块：前半部分讲 Go 1.21+ 的标准库结构化日志 `log/slog`（Text/JSON 两种 Handler、级别、`With` 派生 logger、`WithGroup` 分组）；后半部分讲泛型时代的标准库工具包 `slices`、`maps`、`cmp`，最后给出"错误 + 日志"组合的三条最佳实践。

**运行命令与预期输出解读**

```bash
go run 03-advanced/06_log_slices_maps.go
```

输出分 5 段：

1. **slog 两种 Handler**：默认 logger 输出到 stderr；TextHandler 输出人类可读的 `key=value` 格式（含一条 Debug）；JSONHandler 输出机器可解析的 JSON——其中一条 Debug 因为级别不够（Handler 设为 Info）**不会输出**。
2. **派生 logger**：`With` 绑定的 `request_id`/`service` 字段出现在每条日志里，继续派生的 `user_id` 逐层累积；`WithGroup("http")` 把 method/path 放进嵌套对象。
3. **slices**：排序后 `[1 2 2 5 8 9]`、`Contains`/`Index`、Clone 浅拷贝、`DeleteFunc` 保留偶数、`Max`/`Min`/`IsSorted`、`Reverse`。
4. **maps 与 cmp**：收集并排序键 `[alice bob carol]`、`maps.Equal` 为 true、`maps.Clone` 浅拷贝互不影响；`cmp.Compare(3, 5)` 返回 -1，`cmp.Or("", "默认名字")` 返回默认值。
5. **错误与日志组合**：用 `errors.Is` 判定后按 Warn 级别记录业务异常，最后演示写入 `io.Discard` 的静默 logger。

**关键代码讲解**

```go
textLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug, // 显式设置最低级别，否则 Debug 会被过滤
}))
```

slog 的级别是 `Debug < Info < Warn < Error`（数值依次 +4），**低于 Handler 级别的日志会被直接丢弃**。默认级别是 Info，想看 Debug 必须显式设置。开发环境用 TextHandler（可读），生产环境用 JSONHandler（可被日志系统解析）。

```go
requestLogger := base.With("request_id", "req-1001", "service", "payment")
childLogger := requestLogger.With("user_id", 42)
```

`With` 返回**新** logger 并附带固定字段，原 logger 不受影响；字段逐层累积。把请求 ID 等公共字段绑定到 logger 上，避免每行日志重复书写——这是结构化日志相对 `log.Printf` 的核心优势。

```go
keys := slices.Collect(maps.Keys(scores)) // maps.Keys 返回迭代器
slices.Sort(keys)
```

`maps.Keys` 返回的是 Go 1.23+ range-over-func 风格的**迭代器**且不保证顺序，所以用 `slices.Collect` 收集成切片再排序，保证输出稳定。`slices.Clone`/`maps.Clone` 都是浅拷贝：顶层互不影响，但 value 若是指针/切片仍共享底层数据。

```go
if errors.Is(err, errNotFound) {
    slog.Warn("业务异常：记录未找到", "id", -1, "err", err)
} else {
    slog.Error("系统异常", "err", err)
}
```

最佳实践：先 `errors.Is` 判断错误种类，**业务可预期的错误用 Warn，真正的异常才用 Error**；底层函数用 `fmt.Errorf("...: %w", err)` 包装保留错误链，只在顶层统一记录日志。

**常见坑/易错点**

1. **`With` 的键值必须成对**：奇数个参数会导致键变成 `!BADKEY`。
2. **`maps.Clone`/`slices.Clone` 是浅拷贝**：value 是指针或切片时仍共享底层数据，改"克隆体"会影响原对象内部。
3. **不要每层都"记录日志 + 返回错误"**：会导致日志重复；底层只返回带上下文的错误，顶层统一记录。
4. **不要写 `log.Fatal` 后再返回错误**：`Fatal` 直接退出进程，后面的代码永远不会执行。

**动手练习**

1. 用 `slices.DeleteFunc` 把 `nums` 中小于 5 的元素全部删除，打印结果。
   - 提示：参照保留偶数的例子，把条件改为 `n < 5`。
2. 给 `example5ErrorWithLog` 增加一个"权限不足"错误 `errForbidden`，并在日志分支中为它使用 Warn 级别。
   - 提示：仿照 `errNotFound` 用 `errors.New` 声明，用 `%w` 包装，用 `errors.Is` 判断。

---

### 07_embed_generate.go：embed、generate 与 build tags

**这个文件带你认识什么**

文件讲三个与"构建"相关的特性：`//go:embed` 在编译期把文件嵌入可执行文件（嵌入为 `string`、`[]byte` 或 `embed.FS`）、`//go:generate` 指令登记代码生成命令、`//go:build` build tags 实现条件编译。本文件第一行的 `//go:build ignore` 本身就是一个活例子。

**运行命令与预期输出解读**

```bash
go run 03-advanced/07_embed_generate.go
```

输出分 4 段：

1. **embed 嵌入资源**：打印 `assets/message.txt` 被嵌入后的字符串内容（两行中文问候语）与字节数。
2. **embed.FS**：通过 `assetsFS.ReadFile` 再次读取该文件并打印字节数，再用 `ReadDir("assets")` 遍历目录、打印其中的文件名 `message.txt`。
3. **go:generate 说明**：打印提示"查看本文件顶部的 //go:generate 注释即可，无需执行"。
4. **build tags**：打印一行说明，解释本文件靠 `//go:build ignore` 躲开批量构建、却可以被 `go run` 直接执行。

`go run` 单文件同样支持 embed：编译器以该文件所在目录为包目录解析 `assets/message.txt`，因此**必须在项目根目录运行**（或保证相对路径能找到 `03-advanced/assets/`）。

**关键代码讲解**

```go
//go:embed assets/message.txt
var message string

//go:embed assets/message.txt
var messageBytes []byte

//go:embed assets
var assetsFS embed.FS
```

embed 的三种形态：单个文件嵌入为 `string` 或 `[]byte`；整个目录嵌入为 `embed.FS`（实现了 `fs.FS` 接口，可用 `ReadFile`/`ReadDir`，也可配合 `http.FS`、`template.ParseFS` 使用）。资源在**编译期**嵌入可执行文件，发布后无需携带 assets 目录，也不用担心相对路径问题。

```go
//go:generate gofmt -w $GOFILE
```

`//go:generate` 只是"登记"命令，**不会在 `go build` / `go run` 时自动执行**，必须手动运行 `go generate ./...` 才会触发。工程上常用于代码生成（如 `stringer -type=Color`、`mockgen -source=foo.go`），把生成命令写在源码旁边可保证两者不脱节。本课程登记的是无害的格式化命令，且不会实际执行。

**常见坑/易错点**

1. **embed 路径规则**：不能包含 `..`（不能跳出当前包目录）、不能以 `/` 开头（不是绝对路径）、只能嵌入当前目录及其子目录中的文件。
2. **嵌入发生在编译期**：编译完成后修改 `assets/message.txt` 不会生效，必须重新编译。
3. **`//go:generate` 必须顶格书写**，且后面不能有空行紧挨 package 声明。
4. **`//go:build` 行与 package 声明之间必须有空行**，否则不会被识别为编译约束（本文件即是正确示范）。

**动手练习**

1. 在 `03-advanced/assets/` 下新建一个 `hello.txt`（内容随意），重新运行程序，观察示例 2 的目录遍历输出。
   - 提示：`assetsFS` 嵌入的是整个 `assets` 目录，新增文件会在 `ReadDir` 中出现（注意需要重新编译）。
2. 新增一个 `//go:embed assets/hello.txt` 的 `string` 变量并打印。
   - 提示：仿照 `message` 的写法，注意变量声明与 `//go:embed` 指令必须相邻。

---

### 08_performance.go：性能优化入门

**这个文件带你认识什么**

文件以"可运行示例 + 注释讲解"的方式带你入门性能优化：用 `time.Since` 做简易耗时对比（并讲解正式 Benchmark 的写法）、在普通程序里用 `runtime/pprof` 采集一份 CPU profile、逃逸分析的查看方法、`sync.Pool` 对象复用，以及切片容量预分配的收益。

**运行命令与预期输出解读**

```bash
go run 03-advanced/08_performance.go
```

输出分 5 段（所有耗时数值每次运行都不同，只作相对比较）：

1. **简易耗时对比**：字符串 `+` 拼接 10000 次 vs `bytes.Buffer` 拼接 10000 次，前者明显更慢（每次拼接产生新字符串，O(n²)；后者复用缓冲区，O(n)）。
2. **CPU profile**：在系统临时目录生成一份 `cpu-*.prof` 文件并打印字节数，演示完自动清理临时目录。
3. **逃逸分析**：打印 `stackValue()` 与 `*heapValue()` 的结果（都是 42），并提示把文件复制出去后用 `go build -gcflags="-m"` 查看逃逸决策。
4. **sync.Pool**：5 次"借用-使用-归还"同一个 `bytes.Buffer`，打印 5 行"第 N 次复用"。
5. **预分配对比**：50 万个元素的 append，预分配版本明显更快；最后验证两种写法 `len` 一致（结果相同，差别只在性能）。

**关键代码讲解**

```go
// 正式的基准测试写法（注释讲解）：
//   func BenchmarkConcat(b *testing.B) {
//       for i := 0; i < b.N; i++ { concat() }
//   }
//   go test -bench=. -benchmem
```

文件用注释讲清了正式 Benchmark 的规则：被测函数放进 `xxx_test.go`，`b.N` 由测试框架自动调整直到结果稳定；`-benchmem` 额外报告每次操作的内存分配（B/op、allocs/op）。`time.Since` 只适合"思想等价"的粗对比，精确测量请用 `go test -bench`。

```go
pprof.StartCPUProfile(f)
// ... 制造 CPU 负载 ...
pprof.StopCPUProfile()
```

普通程序里采集 CPU profile 的三步：创建文件、`StartCPUProfile`、负载结束后 `StopCPUProfile`。产物可用 `go tool pprof cpu.out` 交互分析（top / web 查看热点）。注意**采样需要足够的工作量**，太快的程序可能采不到样本。Web 服务场景更常用 `import _ "net/http/pprof"` 自动挂载 `/debug/pprof/`。

```go
buf := pool.Get().(*bytes.Buffer)
buf.Reset() // 取出后务必 Reset，否则会读到上次的残留数据
buf.WriteString(...)
pool.Put(buf)
```

`sync.Pool` 在 goroutine 之间缓存临时对象，减少分配次数、降低 GC 压力。两个纪律：池为空时由 `New` 创建新对象；**取出后必须 `Reset`** 再使用。

**常见坑/易错点**

1. **`sync.Pool` 取出不 `Reset`**：会读到上次使用的残留数据，造成难以排查的逻辑错误。
2. **把 `sync.Pool` 当连接池用**：池中的对象随时可能被 GC 清掉，只适合缓存"丢了也不可惜"的临时对象。
3. **盲目追求零逃逸**：先保证代码正确，再用 benchmark 验证优化收益，不要为优化而优化。
4. **CPU profile 采样不足**：负载太轻时 profile 可能是空的或没有参考价值。

**动手练习**

1. 把示例 5 的 `n` 从 500000 改为 5000，观察两种写法的耗时差距是否还明显，思考为什么。
   - 提示：扩容次数随 n 增长，n 小时差距被计时噪声淹没。
2. 把示例 1 的字符串拼接改写成一个 `BenchmarkConcat` 放进临时目录的 `xxx_test.go`，用 `go test -bench=. -benchmem` 跑一次。
   - 提示：参照文件顶部注释中的写法，注意测试文件不能带 `//go:build ignore`。

---

### 09_unsafe.go：unsafe 包入门

**这个文件带你认识什么**

文件演示 `unsafe` 包的两个典型用途：一是用 `Sizeof`/`Alignof`/`Offsetof` 观察结构体内存布局与对齐填充，并通过调整字段顺序节省内存；二是 string↔[]byte 的零拷贝转换及其风险。同时讲清 `unsafe.Pointer` 与 `uintptr` 的关键区别。

**运行命令与预期输出解读**

```bash
go run 03-advanced/09_unsafe.go
```

输出分 5 段：

1. **Pointer 与 uintptr**：打印一个 `int64` 变量的地址（十六进制，每次运行不同），并用紧邻的地址运算读取其低 4 字节（小端机器上为 `0x55667788`）。
2. **内存布局**：`inner{a int8; b int64; c int16}` 的总大小为 24 字节，字段 a/b/c 的偏移分别是 0/8/16——a 后面的 1~7 偏移是对齐填充。
3. **字段重排**：把大字段放前面的 `compact{b int64; c int16; a int8}` 只有 16 字节，每个对象节省 8 字节。
4. **零拷贝**：`unsafe.Slice(unsafe.StringData(s), len(s))` 得到共享 string 底层数组的 `[]byte`；反向用 `unsafe.String` 从 `[]byte` 得到 string。
5. **风险总结**：打印 4 条 unsafe 风险与"业务代码避免 unsafe"的结论。

**关键代码讲解**

```go
p := unsafe.Pointer(&x)
addr := uintptr(p)                      // 地址的数值表示
low := *(*int32)(unsafe.Pointer(uintptr(p))) // 同一行内完成转换与使用
```

`unsafe.Pointer` 是**指针**，GC 会跟踪它指向的对象；`uintptr` 只是**整数**，GC 不认为它持有对象。因此不能把 `unsafe.Pointer` 转成 `uintptr` 存起来隔几行再用——期间 GC 可能移动或回收该对象，转回后就是悬空指针。**转换和使用必须在同一表达式/紧邻语句中完成**。

```go
type inner struct {
    a int8  // 1 字节
    b int64 // 8 字节，需要 8 字节对齐 → 偏移 1~7 被填充
    c int16 // 2 字节
}
// inner 总大小 24 = a(1) + 填充(7) + b(8) + c(2) + 尾部填充(6)
```

对齐规则：字段地址必须是其对齐值的倍数。`b` 是 `int64` 需 8 字节对齐，所以 a 后面填 7 字节。把字段按对齐要求**从大到小排列**（`compact`），总大小从 24 降到 16。在百万级对象的切片中，这种重排能显著降低内存占用。

```go
b := unsafe.Slice(unsafe.StringData(s), len(s)) // string → []byte 零拷贝
s2 := unsafe.String(unsafe.SliceData(raw), len(raw)) // []byte → string 零拷贝
```

零拷贝直接复用底层数组，省掉一次内存分配与拷贝。但它**共享内存**：string 的底层是只读的，修改零拷贝得到的 `[]byte` 会崩溃或未定义行为；反向转换后修改原切片，string 内容会"悄悄变化"，破坏 string 不可变约定。

**常见坑/易错点**

1. **`uintptr` 不阻止 GC 回收**：长期持有再转回 `unsafe.Pointer` 是悬空指针，必须紧邻使用。
2. **禁止修改零拷贝得到的 `[]byte`**：`b[0] = 'X'` 会直接崩溃或未定义行为（string 底层是只读内存）。
3. **零拷贝破坏 string 不可变性**：`unsafe.String` 的结果共享原切片内存，原切片被改后 string 内容会变化。
4. **可移植性差**：不同平台对齐规则可能不同，手工内存布局假设会降低可移植性。结论：业务代码避免 unsafe，仅限标准库级、经过充分测试的底层优化。

**动手练习**

1. 定义结构体 `{ a int8; b int32; c int8 }`，用 `Sizeof` 打印它的大小，再重排字段使总大小最小。
   - 提示：`int32` 需要 4 字节对齐；最优排列是 `b, a, c`，大小为 8。
2. 用 `Offsetof` 验证你对上题每个字段偏移的预测。
   - 提示：参照示例 2 打印三个字段的偏移量。

---

### 10_design_patterns.go：Go 常见设计模式

**这个文件带你认识什么**

文件用 5 个示例讲完 Go 工程中最常用的模式与原则：函数式选项模式、`sync.Once` 单例、简单工厂、装饰器/中间件，以及依赖注入配合三条接口设计原则（小接口、使用方定义接口、接受接口返回结构体）。Go 的模式比面向对象语言更轻量，大多只是"函数 + 接口"的组合。

**运行命令与预期输出解读**

```bash
go run 03-advanced/10_design_patterns.go
```

输出分 5 段：

1. **函数式选项**：`NewServer()` 全用默认值（`localhost:8080`，5s 超时）；`NewServer(WithPort(9090), WithTimeout(10*time.Second))` 只覆盖关心的两项。
2. **单例**：3 个 goroutine 并发调用 `GetConfig()` 都拿到 "全局配置"（打印顺序每次运行可能不同，但初始化只发生一次），且两次调用返回同一对象（`true`）。
3. **简单工厂**：`NewNotifier("email")`、`NewNotifier("sms")` 分别返回带 `[邮件]` / `[短信]` 前缀的实现；`NewNotifier("pigeon")` 返回预期中的错误。
4. **装饰器**：`WithLogging(WithUpper(greet))("gopher")` 输出 `[日志] 处理 gopher，耗时 ...`（耗时每次不同）和大写化的最终结果。
5. **依赖注入**：`UserService` 注入 `DBUserSaver` 时打印 `[DB] 保存用户: alice`；注入 `FakeUserSaver` 时不碰真实资源，测试替身记录下调用 `bob`。

**关键代码讲解**

```go
type Option func(*Server)

func NewServer(opts ...Option) *Server {
    s := &Server{host: "localhost", port: 8080, timeout: 5 * time.Second}
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

函数式选项模式的核心是"修改配置的函数"。先填默认值、再依次应用选项——**新增配置项时不用改构造函数签名，向后兼容**。这是 `grpc.Dial`、众多 Go 库的标准做法。注意要给出合理默认值，避免零值导致行为异常。

```go
func GetConfig() *config {
    once.Do(func() {
        instance = &config{name: "全局配置"}
    })
    return instance
}
```

`sync.Once` 保证初始化函数**即使并发调用也只执行一次**，天然线程安全，无需手动加锁——这是 Go 实现单例的惯用法。

```go
type UserSaver interface {
    Save(name string) error // 小接口：只声明真正需要的方法
}

func NewUserService(s UserSaver) *UserService { // 接受接口、返回结构体
    return &UserService{saver: s}
}
```

三条接口设计原则的集中体现：**小接口**（只有 `Save` 一个方法）；**接口在使用方定义**（`UserService` 需要什么就声明什么，而不是实现方提供"大而全"接口）；**接受接口、返回结构体**（参数用接口让调用方自由传入任意实现，返回值用具体结构体让调用方用到全部字段和方法）。注入接口后，测试时可轻松换成 `FakeUserSaver`——这就是易测试性。

**常见坑/易错点**

1. **选项模式忘了给默认值**：调用方一个选项都不传时，零值（如 `port: 0`）可能导致程序行为异常。
2. **单例手写双重检查锁**：Go 里直接用 `sync.Once` 即可，自己加锁容易写出有数据竞争的代码。
3. **装饰器顺序搞反**：`WithLogging(WithUpper(greet))` 像洋葱一样层层包裹，先写的在最外层——日志记录的是最外层看到的行为。
4. **接口定义在实现方**：会导致接口臃肿、实现方与使用方耦合；Go 的惯例是"谁用谁定义"。

**动手练习**

1. 给 `Server` 增加一个 `maxConns int` 配置项（默认 100），并实现对应的 `WithMaxConns` 选项。
   - 提示：仿照 `WithPort` 写一个返回 `Option` 的函数，并在 `NewServer` 的默认值里加上 `maxConns: 100`。
2. 给工厂增加一个 `webhookNotifier`（前缀 `[Webhook]`），让 `NewNotifier("webhook")` 可用。
   - 提示：实现 `Notify` 方法并在 `switch` 中加一个 `case`。

---

## 五、本阶段小结与进入下一阶段条件

### 小结

高级阶段把视线从"怎么写"转向"为什么这样运行"：反射让你理解 `encoding/json` 等库的工作原理，并发模式让你能搭建可靠的多 goroutine 系统，GMP 与 GC 让你看懂程序的运行时行为，HTTP 两章把超时、中间件、优雅关闭、重试、连接复用这些生产要素一一落实，slog 与泛型工具包展示了现代 Go 标准库的面貌，embed/build tags 解决了构建与分发问题，性能与 unsafe 两章给了优化的工具与边界意识，最后的设计模式则把前面的一切收拢为可维护的工程结构。

### 结课后的进阶方向

本阶段是 go-learning 的最后一个阶段。之后你可以：

1. **阅读标准库源码**：从 `net/http`、`log/slog`、`sync` 开始，对照本阶段学到的概念验证理解。
2. **写一个真实项目**：如一个带中间件、结构化日志、优雅关闭的 REST API 服务，把 02/04/05/06/10 的知识串起来。
3. **pprof 实战**：给真实程序做 CPU/内存/goroutine profile，用 `go tool pprof` 定位热点与泄漏。
4. **学习 `go test` 生态**：表驱动测试、Benchmark、fuzzing（`go test -fuzz`）。
5. **了解 Go 团队博客与 proposals**：跟踪 generics、range-over-func 等新特性的演进。

### 可检验的结课标准

- [ ] 能用反射遍历任意结构体的字段与 tag，并解释为什么修改字段必须 `&v.Elem()` 且 `CanSet()` 为 true。
- [ ] 能手写一个 worker pool，说清"发完任务为什么要 close 通道"以及 fan-in 为什么必须等 `wg.Wait()` 再关闭输出通道。
- [ ] 能解释 G/M/P 各自的角色，并说明容器环境为什么要关注 `GOMAXPROCS`。
- [ ] 能写出一个带三个超时、recover 中间件和 `Shutdown(ctx)` 优雅关闭的 HTTP 服务器，以及一个带超时、重试、连接复用的客户端。
- [ ] 能说出 unsafe 零拷贝 `[]byte` 至少两条风险，并给出"什么场景才值得用 unsafe"的判断标准。

---

## 六、自测题（5 题，附答案或提示）

**1.（reflect 可寻址修改）** 下面的代码会 panic，为什么？如何修改才能成功把 `u.Age` 改为 20？

```go
u := User{Name: "小明", Age: 18}
v := reflect.ValueOf(u)
v.FieldByName("Age").SetInt(20)
```

答案：`reflect.ValueOf(u)` 得到的是结构体的**副本**，不可寻址，`CanSet()` 为 false，调用 `SetInt` 会 panic。必须传入指针再 `Elem()` 得到指向原变量的可寻址 Value：

```go
v := reflect.ValueOf(&u).Elem()
age := v.FieldByName("Age")
if age.IsValid() && age.CanSet() {
    age.SetInt(20) // 类型必须匹配
}
```

---

**2.（worker pool 背压）** 在 02 的 worker pool 中，`jobs` 是无缓冲通道且只有 3 个 worker，如果一次性要派发 100 万个任务，内存会被任务队列撑爆吗？为什么？

答案：不会。无缓冲通道没有队列容量，发送方在 `jobs <- j` 上**阻塞**，直到有 worker 取走任务才能发送下一个——这就是背压（backpressure）：生产速度被消费速度自然限流，任意时刻"在飞"的任务最多只有 worker 数量级，不会堆积。这也正是 worker pool 用固定 goroutine 数控制资源占用的意义。

---

**3.（GOMAXPROCS）** `runtime.GOMAXPROCS(0)` 与 `runtime.GOMAXPROCS(1)` 分别做了什么？在容器环境里只依赖默认值有什么风险？

答案：传 0 是**只读取**当前 P 的数量、不修改（返回当前值）；传 1 是**设置** P 数量为 1，所有 goroutine 挤在一个 P 上串行调度，返回值是旧值。容器环境中 `runtime.NumCPU()` 看到的是**宿主机**核数，默认 `GOMAXPROCS` 也随之按宿主机核数设置，可能远超 cgroup 的 CPU 配额，导致频繁抢占与限流；生产环境应按配额显式设置 `GOMAXPROCS`（或使用自动适配方案）。

---

**4.（http.Server 超时/优雅关闭）** 04 中的 `srv.Shutdown(ctx)` 与 `srv.Close()` 有什么区别？为什么 `Serve` 返回 `http.ErrServerClosed` 不算错误？

答案：`Shutdown(ctx)` 是**优雅关闭**：停止接收新连接，并等待进行中的请求处理完成（受 ctx 超时约束，示例给了 2 秒）；`Close()` 是**立即强杀**所有连接，进行中的请求可能丢失响应。`Shutdown` 被调用后，阻塞中的 `Serve` 返回 `http.ErrServerClosed`，这是通知"服务器是按计划关闭的"，属于正常流程，所以要用 `errors.Is(err, http.ErrServerClosed)` 排除，不应当作异常打印或告警。

---

**5.（unsafe 零拷贝风险）** 09 中用 `unsafe.Slice(unsafe.StringData(s), len(s))` 把 string 零拷贝转成 `[]byte`。请说出至少两条这种写法的风险。

答案：（任意两条即可）

1. **禁止修改**：零拷贝得到的 `[]byte` 指向 string 的只读底层内存，`b[0] = 'X'` 会直接崩溃或产生未定义行为。
2. **GC/生命周期风险（uintptr 相关）**：涉及 `unsafe.Pointer` 与 `uintptr` 互转时，`uintptr` 不阻止 GC 回收，分开存放再转回是悬空指针，转换与使用必须紧邻。
3. **破坏 string 不可变约定（反向转换）**：`unsafe.String` 的结果共享原 `[]byte` 内存，原切片被修改后 string 内容会"悄悄变化"，依赖 string 不可变性的代码（如 map key）会被破坏。
4. 结论：普通业务代码应优先用标准转换 `[]byte(s)` / `string(b)`，只有热点路径且经 benchmark 证明收益明显时才考虑零拷贝。

---

> 至此，go-learning 全部三个阶段（基础、进阶、高级）的内容就完整了。祝你在真实的 Go 项目中玩得开心！
