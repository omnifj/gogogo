//go:build ignore

// ============================================================
// 文件：06_sync.go
// 阶段：中级
// 主题：sync 包与并发同步
// 学习目标：
//   1. 用 sync.Mutex / sync.RWMutex 保护共享状态
//   2. 用 sync.Once 实现单次初始化
//   3. 理解 sync/atomic 原子操作与锁的性能差异
//   4. 了解 sync.Map 的适用场景
// 运行方式：go run 02-intermediate/06_sync.go
// 预计阅读时间：18 分钟
// ============================================================

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	fmt.Println("========== 示例 1：sync.Mutex 保护共享计数器 ==========")
	mutexCounter()
	fmt.Println()

	fmt.Println("========== 示例 2：sync.RWMutex 读写锁 ==========")
	rwMutexCache()
	fmt.Println()

	fmt.Println("========== 示例 3：sync.Once 单次初始化 ==========")
	onceInit()
	fmt.Println()

	fmt.Println("========== 示例 4：sync/atomic 原子操作 ==========")
	atomicCounter()
	fmt.Println()

	fmt.Println("========== 示例 5：sync.Map 适用场景 ==========")
	syncMapDemo()
}

// ===== 示例 1：sync.Mutex 保护共享计数器 =====
func mutexCounter() {
	var (
		mu      sync.Mutex
		counter int
		wg      sync.WaitGroup
	)
	// 启动 100 个 goroutine，每个累加 1000 次，终值必须恰好是 100000
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				mu.Lock() // 临界区开始：同一时间只允许一个 goroutine 修改 counter
				counter++
				mu.Unlock() // 临界区结束；忘记 Unlock 会导致其他 goroutine 永久阻塞（死锁）
			}
		}()
	}
	wg.Wait()
	// 提示：输出顺序可能不同（goroutine 内部互斥交错），但计数终值必须确定为 100000
	fmt.Println("100 个 goroutine 累加后的计数器终值:", counter)

	// ⚠️ 注意：如果去掉 mu.Lock()/mu.Unlock()，counter++ 会被编译成
	// "读-加-写" 三步，多个 goroutine 交错执行会丢失更新（数据竞态）。
	// 用下面命令可以亲眼看 race detector 抓到它（本示例不会真实制造竞态）：
	//   go run -race 02-intermediate/06_sync.go
	// 故意注释掉 Lock/Unlock 再跑，race detector 会报告 WARNING: DATA RACE。
}

// ===== 示例 2：sync.RWMutex 读写锁 =====
func rwMutexCache() {
	var (
		rw    sync.RWMutex
		cache = make(map[string]string)
		wg    sync.WaitGroup
	)
	// 读操作远多于写操作时，RWMutex 让并发读互不阻塞，性能优于 Mutex
	read := func(key string) {
		defer wg.Done()
		rw.RLock() // 多个读锁可以共存
		_ = cache[key]
		rw.RUnlock()
	}
	write := func(key, val string) {
		defer wg.Done()
		rw.Lock() // 写锁独占：阻塞所有其他读与写
		cache[key] = val
		rw.Unlock()
	}
	wg.Add(1)
	go write("lang", "go")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go read("lang")
	}
	wg.Wait()
	fmt.Println("缓存写入后读取结果:", cache["lang"])
	// ⚠️ 注意：写多读少的场景用 RWMutex 反而更慢（锁开销更大），
	// 且读锁内不能再获取写锁，否则必然死锁。
}

// ===== 示例 3：sync.Once 单次初始化 =====
var (
	once     sync.Once
	configOK bool
)

func initConfig() {
	// 无论多少个 goroutine 调用，这个函数体只执行一次
	configOK = true
	fmt.Println("initConfig 实际执行了（只会看到一次）")
}

func onceInit() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			once.Do(initConfig) // Once 保证并发下也恰好执行一次
		}()
	}
	wg.Wait()
	fmt.Println("配置是否初始化成功:", configOK)
	// ⚠️ 注意：Once 不可重置；若初始化可能失败需要重试，应改用 errgroup 或手动锁。
}

// ===== 示例 4：sync/atomic 原子操作 =====
func atomicCounter() {
	var (
		counter atomic.Int64 // 零值即可用，无需构造
		wg      sync.WaitGroup
	)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				counter.Add(1) // 底层是 CPU 原子指令，比 Mutex 轻量得多
			}
		}()
	}
	wg.Wait()
	// 提示：输出顺序可能不同，但计数终值必须确定为 100000
	fmt.Println("atomic 计数器终值:", counter.Load())
	// ⚠️ 注意：atomic 只适合单个数值/指针的简单操作；
	// 涉及多个变量的一致性时仍需 Mutex（原子操作组合起来不是原子的）。
}

// ===== 示例 5：sync.Map 适用场景 =====
func syncMapDemo() {
	var m sync.Map // 免锁的并发安全 map，无需初始化
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", n%3) // 故意让多个 goroutine 写相同的 key
			m.Store(key, n)
			val, _ := m.Load(key)
			_ = val
		}(i)
	}
	wg.Wait()
	// LoadOrStore：存在则读，不存在则写，一步原子完成
	actual, loaded := m.LoadOrStore("hero", "gopher")
	fmt.Println("LoadOrStore 结果:", actual, "是否已存在:", loaded)
	m.Range(func(k, v any) bool {
		fmt.Printf("  %v => %v\n", k, v)
		return true // 返回 false 可提前终止遍历
	})
	// 提示：输出顺序可能不同，但每个 key 必定各出现一次。
	// ⚠️ 注意：sync.Map 适合"读多写少"或"不同 goroutine 操作不同 key"的场景；
	// 普通 map+Mutex 在写频繁时往往更快且类型安全（sync.Map 的键值都是 any）。
}
