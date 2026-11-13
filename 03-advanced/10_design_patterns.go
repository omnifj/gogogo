//go:build ignore

// ============================================================
// 文件：10_design_patterns.go
// 阶段：高级
// 主题：Go 常见设计模式
// 学习目标：
//   1. 掌握函数式选项模式与构造函数注入（依赖注入）
//   2. 掌握单例（sync.Once）、简单工厂、装饰器/中间件
//   3. 理解 Go 接口设计原则：小接口、使用方定义、接受接口返回结构体
// 运行方式：go run 03-advanced/10_design_patterns.go
// 预计阅读时间：25 分钟
// ============================================================

package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("========== 示例 1：函数式选项模式 ==========")
	example1FunctionalOptions()

	fmt.Println()
	fmt.Println("========== 示例 2：单例（sync.Once） ==========")
	example2Singleton()

	fmt.Println()
	fmt.Println("========== 示例 3：简单工厂 ==========")
	example3Factory()

	fmt.Println()
	fmt.Println("========== 示例 4：装饰器 / 中间件 ==========")
	example4Decorator()

	fmt.Println()
	fmt.Println("========== 示例 5：依赖注入与接口设计原则 ==========")
	example5DependencyInjection()
}

// ===== 示例 1：函数式选项模式 =====
// Server 的配置都做成可选的，调用方只传关心的选项
type Server struct {
	host    string
	port    int
	timeout time.Duration
}

// Option 是"修改 Server 配置的函数"，这是函数式选项模式的核心
type Option func(*Server)

// WithHost / WithPort / WithTimeout 分别构造一个选项
// 为什么：新增配置项时不用改构造函数签名，向后兼容
func WithHost(host string) Option {
	return func(s *Server) { s.host = host }
}

func WithPort(port int) Option {
	return func(s *Server) { s.port = port }
}

func WithTimeout(d time.Duration) Option {
	return func(s *Server) { s.timeout = d }
}

// NewServer 先填默认值，再依次应用选项
func NewServer(opts ...Option) *Server {
	s := &Server{host: "localhost", port: 8080, timeout: 5 * time.Second}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func example1FunctionalOptions() {
	defaultServer := NewServer() // 一个选项都不传，全用默认值
	custom := NewServer(WithPort(9090), WithTimeout(10*time.Second))
	fmt.Printf("默认配置: %+v\n", *defaultServer)
	fmt.Printf("自定义配置: %+v\n", *custom)
	// ⚠️ 注意：选项模式下要给出合理的默认值，避免零值导致程序行为异常
}

// ===== 示例 2：单例（sync.Once） =====
type config struct {
	name string
}

var (
	instance *config
	once     sync.Once
)

// GetConfig 保证全局只初始化一次，且并发安全
func GetConfig() *config {
	once.Do(func() {
		// 为什么用 sync.Once：即使多个 goroutine 同时调用，
		// 这个函数也只执行一次，天然线程安全，无需手动加锁
		instance = &config{name: "全局配置"}
	})
	return instance
}

func example2Singleton() {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c := GetConfig()
			fmt.Printf("goroutine %d 拿到配置: %s\n", id, c.name)
		}(i)
	}
	wg.Wait()
	// 提示：输出顺序可能不同（goroutine 调度不确定），但配置只初始化一次
	fmt.Println("两次调用是同一对象:", GetConfig() == GetConfig())
}

// ===== 示例 3：简单工厂 =====
// Notifier 是工厂产出的统一接口
type Notifier interface {
	Notify(msg string) string
}

type emailNotifier struct{}

func (emailNotifier) Notify(msg string) string { return "[邮件] " + msg }

type smsNotifier struct{}

func (smsNotifier) Notify(msg string) string { return "[短信] " + msg }

// NewNotifier 按类型名创建对应实现
// 为什么：调用方只依赖接口和类型名，不感知具体实现，便于扩展
func NewNotifier(kind string) (Notifier, error) {
	switch strings.ToLower(kind) {
	case "email":
		return emailNotifier{}, nil
	case "sms":
		return smsNotifier{}, nil
	default:
		return nil, fmt.Errorf("未知的通知类型: %s", kind)
	}
}

func example3Factory() {
	for _, kind := range []string{"email", "sms"} {
		n, err := NewNotifier(kind)
		if err != nil {
			fmt.Println("创建失败:", err)
			continue
		}
		fmt.Println(n.Notify("订单已发货"))
	}
	if _, err := NewNotifier("pigeon"); err != nil {
		fmt.Println("预期中的错误:", err)
	}
}

// ===== 示例 4：装饰器 / 中间件 =====
// Handler 是一个普通的业务函数类型
type Handler func(name string) string

// WithLogging 是一个装饰器：包装原 Handler，在前后附加日志与计时
// 为什么：日志、计时、鉴权等横切逻辑与业务逻辑解耦，可自由组合叠加
func WithLogging(h Handler) Handler {
	return func(name string) string {
		start := time.Now()
		result := h(name)
		fmt.Printf("[日志] 处理 %s，耗时 %v\n", name, time.Since(start))
		return result
	}
}

// WithUpper 再叠一层：把返回值转为大写
func WithUpper(h Handler) Handler {
	return func(name string) string {
		return strings.ToUpper(h(name))
	}
}

func greet(name string) string { return "你好, " + name }

func example4Decorator() {
	// 像洋葱一样层层包装：先 greet，再 Upper，最外层 Logging
	h := WithLogging(WithUpper(greet))
	fmt.Println("最终结果:", h("gopher"))
	// 提示：耗时数值每次运行可能不同
}

// ===== 示例 5：依赖注入与接口设计原则 =====
// 原则 1：小接口——只声明真正需要的方法（这里只有 Save）
// 原则 2：接口在【使用方】定义——UserService 需要什么就声明什么，而不是由实现方提供一个"大而全"的接口
type UserSaver interface {
	Save(name string) error
}

// UserService 通过构造函数注入依赖（依赖注入）
type UserService struct {
	saver UserSaver
}

func NewUserService(s UserSaver) *UserService {
	return &UserService{saver: s}
}

func (u *UserService) Register(name string) string {
	if err := u.saver.Save(name); err != nil {
		return "保存失败: " + err.Error()
	}
	return "注册成功: " + name
}

// 生产实现：保存到"数据库"（这里用打印模拟）
type DBUserSaver struct{}

func (DBUserSaver) Save(name string) error {
	fmt.Println("[DB] 保存用户:", name)
	return nil
}

// 测试实现：只记录调用，不碰任何真实资源；为什么：注入接口后测试可换 fake，即易测试性
type FakeUserSaver struct {
	saved []string
}

func (f *FakeUserSaver) Save(name string) error {
	f.saved = append(f.saved, name)
	return nil
}

func example5DependencyInjection() {
	prod := NewUserService(DBUserSaver{})
	fmt.Println(prod.Register("alice"))

	fake := &FakeUserSaver{}
	testSvc := NewUserService(fake)
	fmt.Println(testSvc.Register("bob"))
	fmt.Println("测试替身记录到的调用:", fake.saved)

	// 原则 3：接受接口、返回结构体
	//   函数参数用接口（UserSaver）——调用方自由传入任意实现；
	//   返回值用具体结构体（*UserService）——调用方能用到全部字段和方法。
	fmt.Println("设计原则：小接口、使用方定义接口、接受接口返回结构体")
}
