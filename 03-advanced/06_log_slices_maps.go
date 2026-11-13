//go:build ignore

// ============================================================
// 文件：06_log_slices_maps.go
// 阶段：高级
// 主题：结构化日志与泛型工具包
// 学习目标：
//   1. 掌握 log/slog 的 TextHandler / JSONHandler、日志级别与派生 logger
//   2. 熟练使用标准库泛型工具包 slices 与 maps
//   3. 了解 cmp 包的 Compare / Or 小工具
//   4. 学会"错误 + 日志"组合的最佳实践
// 运行方式：go run 03-advanced/06_log_slices_maps.go
// 预计阅读时间：20 分钟
// ============================================================

package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"slices"
)

func main() {
	fmt.Println("========== 示例 1：slog 的 TextHandler 与 JSONHandler ==========")
	example1SlogHandlers()

	fmt.Println()
	fmt.Println("========== 示例 2：With 派生 logger 与 Group 分组 ==========")
	example2DerivedLogger()

	fmt.Println()
	fmt.Println("========== 示例 3：slices 泛型工具包 ==========")
	example3Slices()

	fmt.Println()
	fmt.Println("========== 示例 4：maps 与 cmp ==========")
	example4MapsAndCmp()

	fmt.Println()
	fmt.Println("========== 示例 5：错误与日志组合的最佳实践 ==========")
	example5ErrorWithLog()
}

// ===== 示例 1：slog 的 TextHandler 与 JSONHandler =====
func example1SlogHandlers() {
	// 默认 logger 输出到标准错误，是 Text 风格，级别默认为 Info
	slog.Info("默认 logger", "module", "advanced")

	// TextHandler：人类可读的 key=value 格式，适合开发环境
	textLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // 为什么：显式设置最低级别，否则 Debug 会被过滤掉
	}))
	textLogger.Debug("调试信息", "key", "value")
	textLogger.Info("文本格式日志", "count", 3)

	// JSONHandler：机器可解析的 JSON 格式，适合生产环境接入日志系统
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	jsonLogger.Info("JSON 格式日志", "user", "alice", "age", 30)

	// slog.LevelDebug < LevelInfo < LevelWarn < LevelError，数值依次 +4
	// ⚠️ 注意：低于 Handler 级别的日志会被直接丢弃，不会输出
	jsonLogger.Debug("这条 Debug 不会输出，因为级别不够")
}

// ===== 示例 2：With 派生 logger 与 Group 分组 =====
func example2DerivedLogger() {
	base := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// With 返回一个新 logger，附带固定字段；原 logger 不受影响
	// 为什么：把请求 ID 等公共字段绑定到 logger 上，避免每行日志重复书写
	requestLogger := base.With("request_id", "req-1001", "service", "payment")
	requestLogger.Info("开始处理请求")

	// 可以继续派生，字段逐层累积
	childLogger := requestLogger.With("user_id", 42)
	childLogger.Info("用户操作完成")

	// Group 把一组字段放进一个嵌套分组，JSON 输出中体现为嵌套对象
	base.WithGroup("http").Info("请求详情", "method", "GET", "path", "/api/v1/users")

	// ⚠️ 注意：With 的键值必须成对出现，奇数个参数会导致键为 "!BADKEY"
}

// ===== 示例 3：slices 泛型工具包 =====
func example3Slices() {
	nums := []int{5, 2, 8, 1, 9, 2}

	// Sort 原地排序（升序），内部使用泛型，无需再写 sort.Ints
	slices.Sort(nums)
	fmt.Println("排序后:", nums)

	// Contains 判断元素是否存在，Index 返回首个下标（不存在返回 -1）
	fmt.Println("是否包含 8:", slices.Contains(nums, 8))
	fmt.Println("8 的下标:", slices.Index(nums, 8))

	// Clone 做一次浅拷贝，避免修改影响原切片
	// 为什么：切片是引用类型，直接赋值只拷贝切片头，底层数组仍共享
	copied := slices.Clone(nums)
	copied[0] = 100
	fmt.Println("原切片:", nums, "克隆后修改:", copied)

	// DeleteFunc 按条件删除元素，返回新切片（可能复用原底层数组）
	evens := slices.DeleteFunc(slices.Clone(nums), func(n int) bool {
		return n%2 != 0 // 删除所有奇数，只保留偶数
	})
	fmt.Println("只保留偶数:", evens)

	// Max / Min / IsSorted 等常用操作
	fmt.Println("最大值:", slices.Max(nums), "最小值:", slices.Min(nums))
	fmt.Println("是否已排序:", slices.IsSorted(nums))

	// Reverse 原地反转
	slices.Reverse(evens)
	fmt.Println("反转后:", evens)
}

// ===== 示例 4：maps 与 cmp =====
func example4MapsAndCmp() {
	scores := map[string]int{"alice": 90, "bob": 85, "carol": 92}

	// maps.Keys 返回键的迭代器（Go 1.23+ 的 range-over-func 风格）
	// 为什么：迭代器不保证顺序，所以先收集到切片再排序，保证输出稳定
	keys := slices.Collect(maps.Keys(scores))
	slices.Sort(keys)
	fmt.Println("所有键（排序后）:", keys)

	// maps.Equal 按键值对比较两个 map，顺序无关
	other := map[string]int{"alice": 90, "bob": 85, "carol": 92}
	fmt.Println("两个 map 相等:", maps.Equal(scores, other))

	// maps.Clone 浅拷贝一个 map
	cloned := maps.Clone(scores)
	cloned["bob"] = 100
	fmt.Println("原 map 的 bob:", scores["bob"], "克隆的 bob:", cloned["bob"])

	// cmp.Compare：返回 -1 / 0 / 1，常用于自定义排序的比较函数
	fmt.Println("cmp.Compare(3, 5):", cmp.Compare(3, 5))

	// cmp.Or：返回第一个非零值参数，常用于设置默认值
	// 为什么：比 if 链更简洁地表达"空则取默认"
	name := cmp.Or("", "默认名字")
	fmt.Println("cmp.Or 取默认值:", name)

	// ⚠️ 注意：maps.Clone 是浅拷贝，value 若是指针/切片则仍共享底层数据
}

// ===== 示例 5：错误与日志组合的最佳实践 =====
var errNotFound = errors.New("记录不存在")

func findRecord(id int) error {
	if id <= 0 {
		return fmt.Errorf("查询记录 %d: %w", id, errNotFound) // 用 %w 包装，保留错误链
	}
	return nil
}

func example5ErrorWithLog() {
	// 最佳实践 1：记录 error 时用 slog.Any("err", err) 并附上上下文字段
	err := findRecord(-1)
	if err != nil {
		// 最佳实践 2：先用 errors.Is 判断错误种类，再决定日志级别
		// 为什么：业务可预期的错误（如未找到）用 Warn，真正的异常才用 Error
		if errors.Is(err, errNotFound) {
			slog.Warn("业务异常：记录未找到", "id", -1, "err", err)
		} else {
			slog.Error("系统异常", "err", err)
		}
	}

	// 最佳实践 3：不要在每一层都"记录日志 + 返回错误"，会导致日志重复
	// 推荐：底层只返回错误（附上下文），由顶层统一记录
	// ⚠️ 注意：不要写 log.Fatal 后再返回错误，Fatal 会直接退出进程

	// 演示：把日志写到 io.Discard，可在压力场景下丢弃日志输出
	silent := slog.New(slog.NewTextHandler(io.Discard, nil))
	silent.Info("这条日志被丢弃，但调用是安全的")
	fmt.Println("错误与日志组合示例完成")
}
