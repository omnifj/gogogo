//go:build ignore

// ============================================================
// 文件：02_errors.go
// 阶段：中级
// 主题：错误处理
// 学习目标：
//   1. 理解 error 接口本质与 errors.New / fmt.Errorf
//   2. 掌握 %w 包装错误与 errors.Is / errors.As 检查
//   3. 学会定义带字段的自定义错误类型
//   4. 明确 panic/recover 的适用边界
// 运行方式：go run 02-intermediate/02_errors.go
// 预计阅读时间：18 分钟
// ============================================================

package main

import (
	"errors"
	"fmt"
)

// ===== 示例 1：error 接口本质与 errors.New =====

// error 就是一个内置接口：type error interface { Error() string }
// 任何实现了 Error() string 的类型都是错误

// ErrNotFound 是哨兵错误（sentinel error），用于被 errors.Is 比较
var ErrNotFound = errors.New("记录不存在")

func findUser(id int) error {
	if id <= 0 {
		// errors.New 创建最简单的静态错误
		return errors.New("id 必须为正数")
	}
	if id > 100 {
		return ErrNotFound
	}
	return nil // nil 表示没有错误
}

func demoErrorBasics() {
	if err := findUser(-1); err != nil {
		// Go 的惯例：错误作为返回值显式处理，而不是抛异常
		fmt.Println("查询失败：", err)
	}
	if err := findUser(200); err != nil {
		fmt.Println("查询失败：", err)
	}
	fmt.Println("查询 id=50 结果 err =", findUser(50))
}

// ===== 示例 2：fmt.Errorf 与 %w 包装 =====

func openConfig(name string) error {
	// %w 把底层错误"包"进新错误，形成错误链，之后可以用 errors.Is 追溯
	err := fmt.Errorf("打开配置 %s 失败", name)
	return fmt.Errorf("初始化模块: %w", err)
}

func demoErrorf() {
	err := openConfig("app.yaml")
	fmt.Println("包装后的错误：", err)

	// fmt.Errorf 不带 %w 时只是格式化字符串，不形成错误链
	plain := fmt.Errorf("普通错误：%d", 42)
	fmt.Println("普通格式化错误：", plain)
}

// ===== 示例 3：errors.Is 与 errors.As =====

// 自定义错误类型：带字段，携带更多上下文
type QueryError struct {
	Query string
	Code  int
}

// 实现 Error() string，*QueryError 就是一个 error
func (e *QueryError) Error() string {
	return fmt.Sprintf("查询 %q 失败，错误码 %d", e.Query, e.Code)
}

func runQuery(q string) error {
	if q == "" {
		return &QueryError{Query: q, Code: 400}
	}
	// 包装哨兵错误，演示 errors.Is 能穿透包装层
	return fmt.Errorf("执行查询: %w", ErrNotFound)
}

func demoIsAs() {
	// errors.Is：沿错误链逐层比较，判断"是不是某个错误"
	err := runQuery("select *")
	if errors.Is(err, ErrNotFound) {
		fmt.Println("errors.Is 命中：底层原因是记录不存在")
	}

	// errors.As：沿错误链查找特定"类型"的错误，并取出字段
	err2 := runQuery("")
	var qe *QueryError
	if errors.As(err2, &qe) {
		fmt.Printf("errors.As 取出字段：Code=%d Query=%q\n", qe.Code, qe.Query)
	}
	// ⚠️ 注意：errors.As 的第二个参数必须是非 nil 指针，指向目标类型
}

// ===== 示例 4：error 与 nil 的常见坑 =====

func demoNilError() {
	var err error // 接口零值是 nil
	fmt.Println("未赋值的 error == nil ?", err == nil)

	// ⚠️ 注意：把非 nil 的具体类型装进 error 后，error 接口本身不再是 nil
	var qe *QueryError = nil
	err = qe // 接口里装着 (类型=*QueryError, 值=nil)，不等于 nil！
	fmt.Println("装了 nil 指针的 error == nil ?", err == nil)
	if err != nil {
		fmt.Println("误以为有错误：", err)
	}
}

// ===== 示例 5：panic / recover 的边界 =====

// ⚠️ 注意：panic 只应用于"不可恢复的程序错误"（如数组越界、程序员逻辑错误），
// 正常的业务错误（文件不存在、网络超时等）一律用 error 返回值处理。
// recover 只能在 defer 函数中生效，通常只在程序边界（如 HTTP 框架、main 顶层）使用。

func safeDivide(a, b int) (result int, err error) {
	// defer + recover：把可能发生的 panic 转换成 error 返回
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("发生 panic，已安全恢复：%v", r)
		}
	}()
	if b == 0 {
		// 这里仅用于教学演示：故意 panic 一次，由上面的 recover 捕获
		panic("除数为零（演示用 panic）")
	}
	return a / b, nil
}

func demoPanicRecover() {
	// 正常路径：用 error 处理
	if res, err := safeDivide(10, 2); err == nil {
		fmt.Println("10 / 2 =", res)
	}
	// 演示路径：panic 被 recover 捕获，程序不会崩溃
	if _, err := safeDivide(10, 0); err != nil {
		fmt.Println(err)
	}
	fmt.Println("程序在 panic 之后依然正常运行——这正是 recover 的意义")
}

func main() {
	fmt.Println("========== 示例 1：error 接口本质与 errors.New ==========")
	demoErrorBasics()
	fmt.Println()

	fmt.Println("========== 示例 2：fmt.Errorf 与 %w 包装 ==========")
	demoErrorf()
	fmt.Println()

	fmt.Println("========== 示例 3：errors.Is 与 errors.As ==========")
	demoIsAs()
	fmt.Println()

	fmt.Println("========== 示例 4：error 与 nil 的常见坑 ==========")
	demoNilError()
	fmt.Println()

	fmt.Println("========== 示例 5：panic / recover 的边界 ==========")
	demoPanicRecover()
}
