//go:build ignore

// ============================================================
// 文件：04_http_server.go
// 阶段：高级
// 主题：HTTP 服务器
// 学习目标：
//   1. 掌握 http.Server 结构体配置与三个超时参数
//   2. 区分 Handler 接口与 HandlerFunc 适配器
//   3. 组合中间件链（logging / recover / 计时）
//   4. 用 http.ServeMux 路由并用 Shutdown(ctx) 优雅关闭
// 运行方式：go run 03-advanced/04_http_server.go
// 预计阅读时间：25 分钟
// ============================================================

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// helloHandler 演示 Handler 接口：任何实现 ServeHTTP 的类型都是 Handler
type helloHandler struct{}

func (helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "你好，来自 Handler 接口实现")
}

// ===== 中间件：本质都是 func(http.Handler) http.Handler =====

// loggingMiddleware 记录每次请求的方法与路径
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[log] %s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r) // 调用链上的下一个处理器
	})
}

// recoverMiddleware 兜底 panic，避免单个 handler 拖垮整个服务器
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// ⚠️ 注意：net/http 会为每个连接单独 recover，但响应会异常断开，
				// 自己 recover 才能返回友好的 500
				http.Error(w, "内部错误", http.StatusInternalServerError)
				fmt.Println("[recover] 捕获 panic:", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// timingMiddleware 统计请求处理耗时
func timingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// 提示：耗时数值每次运行可能不同
		fmt.Printf("[time] %s 耗时 %v\n", r.URL.Path, time.Since(start))
	})
}

// chain 把中间件按顺序包到最终 handler 上（先写的最外层）
func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// ===== 示例 1：搭建并配置 http.Server =====
func buildServer() *http.Server {
	mux := http.NewServeMux()

	// Handler 接口方式注册
	mux.Handle("/hello", helloHandler{})

	// HandlerFunc 方式：普通函数一行适配，最常用
	mux.HandleFunc("/time", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, time.Now().Format("15:04:05"))
	})

	// 故意 panic 的路由，用于验证 recover 中间件
	mux.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("模拟 handler 崩溃")
	})

	srv := &http.Server{
		// 监听 127.0.0.1:0：仅本机回环 + 随机端口，教学环境最安全
		Addr: "127.0.0.1:0",
		// 中间件链：logging 在最外层，recover 兜底，timing 最贴近业务
		Handler: chain(mux, loggingMiddleware, recoverMiddleware, timingMiddleware),
		// 三个超时是防慢速攻击（Slowloris）的基本配置
		ReadTimeout:  5 * time.Second,  // 读完整个请求（含 body）的时限
		WriteTimeout: 5 * time.Second,  // 写完整个响应的时限
		IdleTimeout:  60 * time.Second, // keep-alive 空闲连接的保留时长
	}
	return srv
}

// ===== 示例 2：启动服务器（随机端口） =====
func startServer(srv *http.Server) (addr string, err error) {
	// 先 Listen 再 Serve，这样才能拿到随机分配的真实端口
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return "", err
	}
	go func() {
		// Shutdown 触发后 Serve 返回 ErrServerClosed，属于正常流程
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("服务器异常退出:", err)
		}
	}()
	return ln.Addr().String(), nil
}

// ===== 示例 3：用 http.Client 自测路由与中间件 =====
func selfTest(addr string) {
	base := "http://" + addr
	client := &http.Client{Timeout: 3 * time.Second}

	get := func(path string) {
		resp, err := client.Get(base + path)
		if err != nil {
			fmt.Println("请求失败:", err)
			return
		}
		defer resp.Body.Close() // ⚠️ 注意：不关闭 Body 会导致连接无法复用
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("GET %-6s -> %d %s", path, resp.StatusCode, string(body))
	}

	get("/hello") // 期望 200
	get("/time")  // 期望 200
	get("/panic") // 期望 500（被 recover 中间件兜底）
	get("/none")  // 期望 404（ServeMux 默认 NotFound）
}

// ===== 示例 4：优雅关闭 =====
func gracefulShutdown(srv *http.Server) {
	// Shutdown 会停止接收新连接，并等待进行中的请求完成
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println("优雅关闭失败:", err)
		return
	}
	fmt.Println("服务器已优雅关闭：新连接被拒，存量请求已处理完")
	fmt.Println("对比：srv.Close() 会立即强杀所有连接，可能丢响应")
}

func main() {
	fmt.Println("========== 示例 1-2：配置并启动服务器 ==========")
	srv := buildServer()
	addr, err := startServer(srv)
	if err != nil {
		fmt.Println("启动失败:", err)
		return
	}
	fmt.Println("服务器监听于", addr, "（随机端口，仅本机）")
	fmt.Println()

	fmt.Println("========== 示例 3：客户端自测 ==========")
	selfTest(addr)
	fmt.Println()

	fmt.Println("========== 示例 4：优雅关闭 ==========")
	gracefulShutdown(srv)
}
