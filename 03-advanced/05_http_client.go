//go:build ignore

// ============================================================
// 文件：05_http_client.go
// 阶段：高级
// 主题：HTTP 客户端
// 学习目标：
//   1. 掌握 http.Client 与 Timeout 的正确配置
//   2. 用 context 取消请求、实现指数退避重试
//   3. 自定义请求头并解析 JSON 响应
//   4. 理解连接复用及其前提条件
// 运行方式：go run 03-advanced/05_http_client.go
// 预计阅读时间：20 分钟
// ============================================================

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

// newTestServer 搭建本地测试服务器（httptest 只监听 127.0.0.1，绝不触网）
func newTestServer() *httptest.Server {
	var failCount atomic.Int32 // /flaky 路由前 2 次失败，第 3 次成功
	mux := http.NewServeMux()

	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name":"gopher","age":14}`)
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(5 * time.Second): // 模拟慢响应
		case <-r.Context().Done(): // 客户端取消后服务端也能感知
			fmt.Println("[server] /slow 感知到客户端取消，提前结束")
			return
		}
		fmt.Fprintln(w, "终于完成了")
	})
	mux.HandleFunc("/flaky", func(w http.ResponseWriter, r *http.Request) {
		if failCount.Add(1) < 3 {
			http.Error(w, "暂时失败", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "第 3 次终于成功")
	})
	mux.HandleFunc("/echo-header", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, r.Header.Get("X-Request-Id"))
	})
	return httptest.NewServer(mux)
}

// ===== 示例 1：http.Client 与 Timeout =====
func demoClientTimeout(base string) {
	fmt.Println("========== 示例 1：Client 与 Timeout ==========")
	// ⚠️ 注意：http.Get 用的默认 Client 没有超时，可能永久挂起
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(base + "/json")
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("GET /json 状态码:", resp.StatusCode)

	// 请求 /slow（5 秒）必然触发 2 秒超时
	_, err = client.Get(base + "/slow")
	if err != nil {
		fmt.Println("GET /slow 按预期超时:", err)
	}
	fmt.Println()
}

// ===== 示例 2：带 context 的请求取消 =====
func demoContextCancel(base string) {
	fmt.Println("========== 示例 2：context 取消请求 ==========")
	client := &http.Client{} // 不设 Timeout，完全交给 context 控制

	// 300ms 后自动取消
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/slow", nil)
	if err != nil {
		fmt.Println("构造请求失败:", err)
		return
	}
	start := time.Now()
	_, err = client.Do(req)
	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Printf("请求被 context 取消，耗时 %v\n", time.Since(start).Round(50*time.Millisecond))
	}
	fmt.Println("提示：耗时数值每次运行可能略有不同")
	fmt.Println()
}

// ===== 示例 3：指数退避重试 =====
func demoRetry(base string) {
	fmt.Println("========== 示例 3：指数退避重试 ==========")
	client := &http.Client{Timeout: time.Second}

	var body string
	for attempt := 0; attempt < 5; attempt++ {
		resp, err := client.Get(base + "/flaky")
		if err == nil && resp.StatusCode == http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			body = string(b)
			fmt.Printf("第 %d 次尝试成功\n", attempt+1)
			break
		}
		if err == nil {
			resp.Body.Close() // 失败响应也要关闭 Body
		}
		// 退避：100ms、200ms、400ms……让服务端有喘息时间
		backoff := time.Duration(100<<attempt) * time.Millisecond
		fmt.Printf("第 %d 次失败，%v 后重试\n", attempt+1, backoff)
		time.Sleep(backoff)
	}
	fmt.Println("最终结果:", body)
	fmt.Println()
}

// ===== 示例 4：自定义请求头 + 解析 JSON =====
type user struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func demoHeaderAndJSON(base string) {
	fmt.Println("========== 示例 4：请求头与 JSON 解析 ==========")
	client := &http.Client{Timeout: 2 * time.Second}

	// 自定义请求头必须用 NewRequest，Get 便捷方法无法加头
	req, _ := http.NewRequest(http.MethodGet, base+"/echo-header", nil)
	req.Header.Set("X-Request-Id", "req-42")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	echo, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Print("服务端回显的 X-Request-Id: ", string(echo)) // []byte 需转 string 才能按文本打印

	// 解析 JSON 响应：json.Decoder 直接读流，比 ReadAll+Unmarshal 更省内存
	resp2, err := client.Get(base + "/json")
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp2.Body.Close()
	var u user
	if err := json.NewDecoder(resp2.Body).Decode(&u); err != nil {
		fmt.Println("JSON 解析失败:", err)
		return
	}
	fmt.Printf("解析结果: %+v\n", u)
	fmt.Println()
}

// ===== 示例 5：连接复用（keep-alive） =====
func demoConnReuse(base string) {
	fmt.Println("========== 示例 5：连接复用说明 ==========")
	client := &http.Client{Timeout: 2 * time.Second}

	// 连续请求同一路径；底层 Transport 会把连接放回池子复用
	for i := 1; i <= 3; i++ {
		resp, err := client.Get(base + "/json")
		if err != nil {
			fmt.Println("请求失败:", err)
			return
		}
		// ⚠️ 注意：复用连接的两个前提——读完 Body 且 Close；
		// 只 Close 不读完，连接会被丢弃重新建立
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		fmt.Printf("第 %d 次请求完成，复用了底层 TCP 连接\n", i)
	}
	fmt.Println("说明：http.Client 并发安全，应全局复用同一个实例，")
	fmt.Println("      不要每次请求都新建（会丢失连接池）")
}

func main() {
	// 本地测试服务器：进程退出前关闭
	server := newTestServer()
	defer server.Close()
	fmt.Println("测试服务器:", server.URL, "（仅本机回环，不触网）")
	fmt.Println()

	demoClientTimeout(server.URL)
	demoContextCancel(server.URL)
	demoRetry(server.URL)
	demoHeaderAndJSON(server.URL)
	demoConnReuse(server.URL)
}
