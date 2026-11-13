//go:build ignore

// ============================================================
// 文件：05_channels.go
// 阶段：中级
// 主题：channel 通道
// 学习目标：
//   1. 掌握无缓冲/有缓冲 channel 的阻塞行为
//   2. 掌握 close、for range 遍历与 select 多路复用
//   3. 用 time.After 实现超时控制
//   4. 理解单向 channel 参数与 nil channel 的坑
// 运行方式：go run 02-intermediate/05_channels.go
// 预计阅读时间：18 分钟
// ============================================================

package main

import (
	"fmt"
	"time"
)

// ===== 示例 1：无缓冲 channel——同步握手 =====

func demoUnbuffered() {
	ch := make(chan string) // 无缓冲：发送方必须等到接收方准备好
	go func() {
		ch <- "你好" // 发送会阻塞，直到 main 接收——像两人当面交接
	}()
	msg := <-ch // 接收与发送在时间上"碰头"，完成一次同步握手
	fmt.Println("  收到：", msg)
}

// ===== 示例 2：有缓冲 channel 的阻塞行为 =====

func demoBuffered() {
	ch := make(chan int, 2) // 容量 2：缓冲没满时发送不阻塞
	ch <- 1
	ch <- 2
	fmt.Println("  缓冲已满前发送两个值，均未阻塞")

	// ⚠️ 注意：此时再 ch <- 3 会阻塞（缓冲已满），必须有接收者腾位置
	go func() { ch <- 3 }()
	time.Sleep(10 * time.Millisecond)

	fmt.Println("  依次取出：", <-ch, <-ch, <-ch)
	// ⚠️ 注意：从无缓冲或已空的 channel 接收也会阻塞，直到有新值到来
}

// ===== 示例 3：close 与接收端检测 =====

func demoClose() {
	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	close(ch) // 关闭：表示"不再发送"，已发送的值仍可取出

	v, ok := <-ch
	fmt.Printf("  取出 %d，ok=%v（channel 未耗尽）\n", v, ok)
	<-ch // 取出 20
	v, ok = <-ch
	// comma-ok 检测：ok=false 表示 channel 已关闭且没有剩余值
	fmt.Printf("  取出 %d，ok=%v（已关闭且耗尽，v 是零值）\n", v, ok)
	// ⚠️ 注意：向已关闭的 channel 发送会 panic；关闭动作应由发送方负责
}

// ===== 示例 4：for range 遍历 channel =====

func demoRange() {
	ch := make(chan int)
	go func() {
		for i := 1; i <= 3; i++ {
			ch <- i
		}
		close(ch) // ⚠️ 注意：不 close 的话，for range 会永远等下去（死锁）
	}()
	// for range 会一直接收，直到 channel 被关闭且耗尽后自动退出循环
	for v := range ch {
		fmt.Println("  range 收到：", v)
	}
	fmt.Println("  channel 已关闭，range 循环正常结束")
}

// ===== 示例 5：select 多路复用与超时 =====

func demoSelect() {
	c1 := make(chan string)
	c2 := make(chan string)
	go func() { time.Sleep(30 * time.Millisecond); c1 <- "来自 c1" }()
	go func() { time.Sleep(10 * time.Millisecond); c2 <- "来自 c2" }()

	for i := 0; i < 2; i++ {
		select {
		case m1 := <-c1:
			fmt.Println("  select 选中：", m1)
		case m2 := <-c2:
			fmt.Println("  select 选中：", m2)
			// 提示：输出顺序可能不同（取决于哪个 channel 先就绪）
		}
	}

	// time.After 实现超时：慢操作超过 50ms 就放弃等待
	slow := make(chan string)
	go func() { time.Sleep(200 * time.Millisecond); slow <- "太慢了" }()
	select {
	case msg := <-slow:
		fmt.Println("  收到：", msg)
	case <-time.After(50 * time.Millisecond):
		fmt.Println("  超时：50ms 内没等到结果，放弃")
	}
}

// ===== 示例 6：单向 channel 与 nil channel 的坑 =====

// producer 只能发送（chan<-），consumer 只能接收（<-chan）
// 单向类型在编译期防止误用方向，是良好的 API 设计
func producer(out chan<- int) {
	for i := 1; i <= 3; i++ {
		out <- i
	}
	close(out)
}

func consumer(in <-chan int) {
	for v := range in {
		fmt.Println("  消费：", v)
	}
}

func demoDirectional() {
	ch := make(chan int) // 双向 channel 可隐式转换为单向
	go producer(ch)
	consumer(ch)
}

func demoNilChannel() {
	var nilCh chan int // nil channel：发送和接收都会永远阻塞
	// ⚠️ 注意：直接 <-nilCh 会死锁！下面用 select 安全演示。
	select {
	case <-nilCh:
		fmt.Println("  永远不会执行到这里")
	default:
		// default 分支立即返回，避免死锁——这是探测 nil channel 的安全方式
		fmt.Println("  nil channel 上无值可取（default 兜底，没有死锁）")
	}
	// nil channel 的正面用途：在 select 里"禁用"某个分支
	// （nil case 永远不就绪，相当于把该分支临时关掉）
}

func main() {
	fmt.Println("========== 示例 1：无缓冲 channel（同步握手） ==========")
	demoUnbuffered()
	fmt.Println()

	fmt.Println("========== 示例 2：有缓冲 channel 的阻塞行为 ==========")
	demoBuffered()
	fmt.Println()

	fmt.Println("========== 示例 3：close 与接收端检测 ==========")
	demoClose()
	fmt.Println()

	fmt.Println("========== 示例 4：for range 遍历 channel ==========")
	demoRange()
	fmt.Println()

	fmt.Println("========== 示例 5：select 多路复用与超时 ==========")
	demoSelect()
	fmt.Println()

	fmt.Println("========== 示例 6：单向 channel 与 nil channel ==========")
	demoDirectional()
	demoNilChannel()
}
