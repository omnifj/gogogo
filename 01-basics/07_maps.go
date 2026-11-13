//go:build ignore

// ============================================================
// 文件：07_maps.go
// 阶段：初级
// 主题：map 映射
// 学习目标：
//   1. 掌握 map 的两种初始化方式（make 与字面量）
//   2. 掌握 comma-ok 判断键是否存在、delete 删除键
//   3. 理解 map 遍历无序，并学会行序无关的验证思路
//   4. 学会用 map 实现 set，并知道 map 并发不安全
// 运行方式：go run 01-basics/07_maps.go
// 预计阅读时间：12 分钟
// ============================================================

package main

import "fmt"

func main() {
	fmt.Println("========== 示例 1：map 的初始化 ==========")
	mapInit()
	fmt.Println()

	fmt.Println("========== 示例 2：comma-ok 与 delete ==========")
	commaOkAndDelete()
	fmt.Println()

	fmt.Println("========== 示例 3：遍历 map（无序） ==========")
	iterateMap()
	fmt.Println()

	fmt.Println("========== 示例 4：用 map 实现 set ==========")
	mapAsSet()
	fmt.Println()

	fmt.Println("========== 示例 5：map 并发不安全（预告） ==========")
	concurrencyPreview()
}

// ===== 示例 1：map 的初始化 =====
func mapInit() {
	// 方式一：make 创建空 map，map 使用前必须初始化
	scores := make(map[string]int)
	scores["Alice"] = 90
	scores["Bob"] = 85
	fmt.Println("make 创建:", scores)

	// 方式二：字面量，声明的同时写入初始数据
	prices := map[string]float64{
		"apple":  5.5,
		"banana": 3.2,
	}
	fmt.Println("字面量创建:", prices)

	// ⚠️ 注意：只声明不初始化得到 nil map，读可以、写会 panic
	var nilMap map[string]int
	fmt.Println("nil map 读取:", nilMap["x"]) // 读 nil map 返回零值，安全
	// nilMap["x"] = 1 // 这一行会 panic，务必先 make 再写入
	_ = nilMap
}

// ===== 示例 2：comma-ok 与 delete =====
func commaOkAndDelete() {
	ages := map[string]int{"Tom": 18, "Jerry": 20}

	// comma-ok 模式：第二个返回值表示键是否存在
	// 因为读取不存在的键会返回零值 0，必须用 ok 区分“不存在”和“值真的是 0”
	if age, ok := ages["Tom"]; ok {
		fmt.Println("Tom 的年龄:", age)
	}
	if _, ok := ages["Nobody"]; !ok {
		fmt.Println("Nobody 不存在于 map 中")
	}

	// delete 删除指定键；删除不存在的键不会报错，是安全操作
	delete(ages, "Tom")
	fmt.Println("删除 Tom 后:", ages)

	// 清空整个 map 的惯用法：逐个 delete，或者直接用内建 clear（Go 1.21+）
	clear(ages)
	fmt.Println("clear 后:", ages, "长度:", len(ages))
}

// ===== 示例 3：遍历 map（无序） =====
func iterateMap() {
	cities := map[string]string{
		"北京": "首都",
		"上海": "魔都",
		"广州": "花城",
		"深圳": "鹏城",
	}

	// 提示：输出顺序可能不同 —— Go 故意随机化 map 遍历顺序，
	// 防止程序依赖某种固定顺序而产生隐蔽 bug
	fmt.Println("遍历结果（每次运行顺序都可能不同）:")
	for city, nickname := range cities {
		fmt.Printf("  %s -> %s\n", city, nickname)
	}

	// 行序无关的设计建议：
	// 1. 测试断言时不要按顺序拼接字符串比较，改用 len + 逐键 comma-ok 检查；
	// 2. 需要稳定输出时，先把键取出放进切片，sort 排序后再遍历；
	// 3. 日志/序列化场景优先使用会排序键的库或自己排序，保证结果可复现。
	count := 0
	for _, nickname := range cities {
		if len(nickname) > 0 { // 只关心“每个键都有一个非空值”，与顺序无关
			count++
		}
	}
	fmt.Println("非空昵称数量（行序无关验证）:", count)
}

// ===== 示例 4：用 map 实现 set =====
func mapAsSet() {
	// Go 没有内置 set，惯用法是 map[T]struct{}
	// struct{} 是空结构体，占 0 字节，比 map[T]bool 更省内存
	set := make(map[string]struct{})

	members := []string{"Go", "Java", "Go", "Python", "Java"} // 含重复元素
	for _, m := range members {
		set[m] = struct{}{} // 重复写入同一键，自然去重
	}
	fmt.Println("去重后的元素个数:", len(set))

	// 判断元素是否存在同样用 comma-ok
	if _, ok := set["Go"]; ok {
		fmt.Println("Go 在集合中")
	}
	if _, ok := set["C++"]; !ok {
		fmt.Println("C++ 不在集合中")
	}
}

// ===== 示例 5：map 并发不安全（预告） =====
func concurrencyPreview() {
	// ⚠️ 注意：Go 的内建 map 不是并发安全的！
	// 多个 goroutine 同时读写同一个 map，会直接触发运行时错误
	// （fatal error: concurrent map read and map write），整个程序崩溃，
	// 而且这种错误无法用 recover 捕获。
	//
	// 本示例只做文字预告，不真正制造数据竞态（那会直接 crash）。
	// 正确的并发保护手段（sync.Mutex、sync.RWMutex、sync.Map）
	// 将在中级阶段的 06_sync.go 中详细讲解。

	m := map[string]int{"safe": 1} // 单 goroutine 下使用是完全安全的
	fmt.Println("单协程读写 map 没有问题:", m)
}
