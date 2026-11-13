# go-learning：Go 语言渐进式学习项目

一个**零第三方依赖、开箱即跑**的 Go 语言自学项目。从「Hello World」到并发模式与底层优化，每个 `.go` 文件都是一堂可以独立运行的小课：配有详细中文注释、示例分隔、陷阱标注，并附带逐文件讲解的文档。

- **零依赖**：只使用 Go 标准库，不需要 `go get` 任何东西，没有 `go.sum`。
- **开箱即跑**：每个文件顶部带 `//go:build ignore`，用 `go run 文件路径` 即可独立运行，互不冲突。
- **渐进式**：初级（语法基础）→ 中级（接口/并发/测试）→ 高级（反射/性能/unsafe/设计模式）。

## 环境要求

- Go 1.27 或更高版本（本项目基于 go1.27 编写；Go 1.21+ 一般均可运行，个别示例讲解的是 Go 1.22 的循环变量新语义）。
- 验证安装：`go version`

## 学习路径

```
第一阶段 01-basics/      第二阶段 02-intermediate/      第三阶段 03-advanced/
语法与数据结构      →     接口、错误、并发、测试     →     反射、性能、网络、底层
（约 2~3 天）            （约 3~5 天）                  （约 5~7 天）
```

建议的学习方法：

1. **先读文件头注释**：每个文件开头写明主题、学习目标、运行方式、预计阅读时间。
2. **再读文档**：`docs/` 目录下有对应阶段的逐文件详解、常见坑和动手练习。
3. **边读边跑**：每读完一个示例就 `go run` 一次，对照输出理解。
4. **动手改**：完成文档中每个文件的「动手练习」，把代码改坏再修好，是最快的学法。

每个阶段的建议阅读顺序、自测题和「进入下一阶段条件」见对应文档：

- [docs/01-basics.md](docs/01-basics.md)
- [docs/02-intermediate.md](docs/02-intermediate.md)
- [docs/03-advanced.md](docs/03-advanced.md)

## 目录结构

```
gogogo/
├── go.mod                          # module go-learning，零第三方依赖
├── README.md
├── .gitignore
├── 01-basics/                      # 初级：语法基础（10 个文件）
│   ├── 01_hello.go                 # 程序结构、fmt 打印
│   ├── 02_variables.go             # 变量、常量、iota、作用域
│   ├── 03_types.go                 # 基本类型、类型转换、rune/byte、strconv
│   ├── 04_control.go               # if/for/switch/range/label
│   ├── 05_functions.go             # 函数、多返回值、defer
│   ├── 06_arrays_slices.go         # 数组、切片、append 的坑
│   ├── 07_maps.go                  # map、set、无序遍历
│   ├── 08_structs.go               # 结构体、方法、嵌入
│   ├── 09_pointers.go              # 指针
│   └── 10_strings_runes.go         # 字符串、rune、UTF-8
├── 02-intermediate/                # 中级：接口与并发（10+1 个文件）
│   ├── 01_interfaces.go            # 接口、类型断言、type switch
│   ├── 02_errors.go                # error、errors.Is/As、panic/recover 边界
│   ├── 03_closures.go              # 闭包、循环变量陷阱
│   ├── 04_goroutines.go            # goroutine、WaitGroup、泄漏
│   ├── 05_channels.go              # channel、select、超时
│   ├── 06_sync.go                  # Mutex、Once、atomic、-race
│   ├── 07_context.go               # Context 取消与超时
│   ├── 08_generics.go              # 泛型
│   ├── 09_json_io.go               # JSON、文件与 IO
│   ├── 10_testing.go               # 被测代码 + 演示
│   └── 10_testing_test.go          # 表驱动测试、基准、示例测试
├── 03-advanced/                    # 高级：底层与工程实践（10 个文件）
│   ├── 01_reflection.go            # 反射
│   ├── 02_concurrency_patterns.go  # worker pool、pipeline、errgroup
│   ├── 03_runtime_scheduler.go     # GMP、调度器、逃逸分析、GOGC
│   ├── 04_http_server.go           # HTTP 服务端、中间件、优雅关闭
│   ├── 05_http_client.go           # HTTP 客户端、超时、重试
│   ├── 06_log_slices_maps.go       # slog、slices/maps/cmp
│   ├── 07_embed_generate.go        # go:embed、go:generate、build tags
│   ├── assets/message.txt          # 07 的 embed 资源
│   ├── 08_performance.go           # 性能分析、sync.Pool、预分配
│   ├── 09_unsafe.go                # unsafe 包
│   └── 10_design_patterns.go       # Go 风格设计模式
└── docs/                           # 与三个阶段一一对应的详解文档
    ├── 01-basics.md
    ├── 02-intermediate.md
    └── 03-advanced.md
```

## 运行命令速查表

| 目的 | 命令 |
| --- | --- |
| 运行任意一个示例文件（项目根目录下） | `go run 01-basics/01_hello.go` |
| 运行测试（表驱动/子测试） | `cd 02-intermediate; go test -v 10_testing.go 10_testing_test.go` |
| 只跑某个测试 | `go test -run TestIsPalindrome -v 10_testing.go 10_testing_test.go` |
| 跑基准测试（PowerShell 中 `-bench=.` 必须加引号） | `go test "-bench=." -benchmem "-benchtime=100x" 10_testing.go 10_testing_test.go` |
| 竞态检测 | `go run -race 02-intermediate/06_sync.go` |
| 检查格式（无输出即通过） | `gofmt -l .` |
| 一键格式化全部文件 | `gofmt -w .` |
| 构建检查（所有文件带 ignore 标签，提示 matched no packages 属正常） | `go build ./...` |
| 观察逃逸分析 | `go run -gcflags="-m" 03-advanced/03_runtime_scheduler.go` |

## 常用工具说明

### gofmt：代码格式化

Go 官方强制风格工具，本项目所有文件均已通过 gofmt。

```powershell
gofmt -l .        # 列出格式不合规的文件；无输出 = 全部合规
gofmt -w <文件>   # 就地格式化某个文件
```

### go run 与 go build

- `go run 文件路径`：编译到临时目录并立即运行，适合学习时快速验证。
- 本项目每个文件都带 `//go:build ignore`，所以 `go build ./...` 不会把任何示例编译成包（会提示 `matched no packages`，这是**预期行为**，退出码为 0）。这样做的好处是：31 个 `package main` 共存于同一模块也互不冲突，且每个文件都能用 `go run` 单独跑。

### go test

`02-intermediate/10_testing_test.go` 演示了表驱动测试、子测试 `t.Run`、基准测试、示例测试。因为示例文件都带 ignore 标签，请用**显式文件列表**的方式运行测试（见速查表），而不是 `go test ./...`。

### go run -race：竞态检测器

```powershell
go run -race 02-intermediate/06_sync.go
```

加上 `-race` 后，运行期一旦检测到数据竞态会立刻报告。学习并发阶段（04~06）时，可以试试故意删掉锁再用 `-race` 跑，观察检测器如何抓出问题。

> ⚠️ 注意：`-race` 依赖 cgo。Windows 上需要 `CGO_ENABLED=1` 并安装 gcc（如 MinGW-w64），否则会报 `go: -race requires cgo`。没有 C 编译器时可以先跳过，竞态检测在 Linux/macOS 或装好 gcc 的 Windows 上均可使用。

## 常见问题

**Q：为什么 `go build ./...` 什么都不产出？**
A：每个示例文件都带 `//go:build ignore`，是有意设计：保证 31 个 `func main` 互不冲突、都能 `go run` 独立运行。要产出可执行文件，可用 `go build -o hello.exe 01-basics/01_hello.go`。

**Q：示例输出里并发打印的顺序和注释不一样？**
A：凡涉及 goroutine 调度或 map 遍历的地方，代码中都有 `// 提示：输出顺序可能不同` 标注，顺序不同是正常现象；但最终结果（如计数终值）是确定的。

**Q：可以照搬示例到生产代码吗？**
A：示例为教学做了简化（例如集中在一个 main 里）。工程实践请重点参考高级阶段的 02（并发模式）、04/05（HTTP）、06（日志）与 10（设计模式）。
