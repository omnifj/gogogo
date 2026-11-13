//go:build ignore

// ============================================================
// 文件：07_context.go
// 阶段：中级
// 主题：context 包与取消传播
// 学习目标：
//   1. 理解 context.Context 的作用与使用约定
//   2. 掌握 WithCancel / WithTimeout / WithDeadline 的取消机制
//   3. 理解取消信号沿调用链传递的方式
//   4. 正确使用 WithValue 传递请求级元数据
// 运行方式：go run 02-intermediate/07_context.go
// 预计阅读时间：18 分钟
// ============================================================

package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	fmt.Println("========== 示例 1：WithCancel 手动取消 ==========")
	withCancelDemo()
	fmt.Println()

	fmt.Println("========== 示例 2：WithTimeout 超时控制 ==========")
	withTimeoutDemo()
	fmt.Println()

	fmt.Println("========== 示例 3：WithDeadline 截止时间 ==========")
	withDeadlineDemo()
	fmt.Println()

	fmt.Println("========== 示例 4：取消沿调用链传递 ==========")
	cancelPropagation()
	fmt.Println()

	fmt.Println("========== 示例 5：WithValue 传递请求级元数据 ==========")
	withValueDemo()
}

// ===== 示例 1：WithCancel 手动取消 =====
func withCancelDemo() {
	// ctx 是取消信号的载体；cancel 必须由调用方持有并负责调用
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go worker(ctx, 1, done)
	time.Sleep(30 * time.Millisecond)
	cancel() // 手动发出取消信号，worker 的 ctx.Done() 会立即被关闭
	<-done   // 等待 worker 真正退出，避免 goroutine 泄漏
	// ⚠️ 注意：cancel 必须调用（哪怕不打算提前取消），通常写成 defer cancel()，
	// 否则 context 关联的资源会泄漏到父 context 取消为止。
}

// worker 演示标准的"可取消"写法：select 同时监听工作与取消
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

// ===== 示例 2：WithTimeout 超时控制 =====
func withTimeoutDemo() {
	// 50ms 后自动取消，等价于 WithDeadline(ctx, time.Now().Add(50ms))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := slowJob(ctx, 200*time.Millisecond) // 任务需要 200ms，注定超时
	fmt.Println("慢任务返回:", err)
	// ⚠️ 注意：超时只负责"通知"，任务方必须主动监听 ctx.Done() 才能真正停下来。
}

func slowJob(ctx context.Context, cost time.Duration) error {
	select {
	case <-time.After(cost):
		return nil // 正常完成
	case <-ctx.Done():
		return ctx.Err() // 被取消/超时，返回 context.DeadlineExceeded 或 Canceled
	}
}

// ===== 示例 3：WithDeadline 截止时间 =====
func withDeadlineDemo() {
	deadline := time.Now().Add(40 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	// Deadline() 可以查询剩余时间，常用于给下游调用动态分配时间
	if d, ok := ctx.Deadline(); ok {
		fmt.Println("距离截止还有(约):", time.Until(d).Round(time.Millisecond))
	}
	<-ctx.Done() // 阻塞到截止时间到达
	fmt.Println("到点自动取消，Err =", ctx.Err())
}

// ===== 示例 4：取消沿调用链传递 =====
func cancelPropagation() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go serviceA(ctx, done) // A -> B -> C 三级调用，共享同一个 ctx
	time.Sleep(30 * time.Millisecond)
	cancel() // 在最外层取消一次，整条链路全部收到信号
	<-done
	fmt.Println("整条调用链已干净退出")
}

func serviceA(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	bDone := make(chan struct{})
	go serviceB(ctx, bDone)
	<-ctx.Done() // 不另建 context，直接沿用父级，信号自然传递
	<-bDone
	fmt.Println("serviceA 退出")
}

func serviceB(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	<-ctx.Done()
	fmt.Println("serviceB 退出（与 serviceA 收到的是同一个信号）")
	// ⚠️ 注意：函数首参数命名为 ctx 并透传，是 Go 社区强约定；
	// 不要把 context 藏进结构体字段（sdk 允许的少数例外除外）。
}

// ===== 示例 5：WithValue 传递请求级元数据 =====
// 自定义类型作 key，避免不同包之间的字符串 key 冲突
type ctxKey string

const requestIDKey ctxKey = "request-id"

func withValueDemo() {
	// ⚠️ 注意：WithValue 只适合放"请求范围元数据"（request-id、trace-id、
	// 认证身份等），不要塞业务结构体或可选参数——那会隐藏依赖、破坏类型安全。
	ctx := context.WithValue(context.Background(), requestIDKey, "req-42")
	handleRequest(ctx)
}

func handleRequest(ctx context.Context) {
	// 取不到时给默认值，保证中间件没注入也能工作
	id, _ := ctx.Value(requestIDKey).(string)
	if id == "" {
		id = "unknown"
	}
	fmt.Printf("处理请求，request-id = %s\n", id)
}
