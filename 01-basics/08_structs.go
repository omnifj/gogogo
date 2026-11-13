//go:build ignore

// ============================================================
// 文件：08_structs.go
// 阶段：初级
// 主题：结构体
// 学习目标：
//   1. 掌握结构体的定义与多种初始化方式
//   2. 理解值接收者与指针接收者的区别及选择标准
//   3. 掌握匿名字段嵌入与方法提升
//   4. 了解字段标签（struct tag）与结构体可比较性
// 运行方式：go run 01-basics/08_structs.go
// 预计阅读时间：14 分钟
// ============================================================

package main

import (
	"encoding/json"
	"fmt"
)

// ===== 示例 1 用的类型定义（必须在包级别） =====
type User struct {
	Name string
	Age  int
	City string
}

func main() {
	fmt.Println("========== 示例 1：结构体定义与初始化 ==========")
	structDefineAndInit()
	fmt.Println()

	fmt.Println("========== 示例 2：值接收者 vs 指针接收者 ==========")
	valueVsPointerReceiver()
	fmt.Println()

	fmt.Println("========== 示例 3：匿名字段嵌入与方法提升 ==========")
	embeddingDemo()
	fmt.Println()

	fmt.Println("========== 示例 4：字段标签与 JSON ==========")
	structTagDemo()
	fmt.Println()

	fmt.Println("========== 示例 5：结构体可比较性 ==========")
	comparabilityDemo()
}

// ===== 示例 1：结构体定义与初始化 =====
func structDefineAndInit() {
	// 方式一：按字段名初始化，推荐，可读性好且不受字段顺序影响
	u1 := User{Name: "小明", Age: 20, City: "北京"}
	fmt.Println("字段名初始化:", u1)

	// 方式二：按字段顺序初始化，必须写全所有字段，字段调整时容易出错
	u2 := User{"小红", 18, "上海"}
	fmt.Println("顺序初始化:", u2)

	// 方式三：只给部分字段赋值，其余为零值（int 是 0，string 是 ""）
	u3 := User{Name: "小刚"}
	fmt.Println("部分字段初始化:", u3)

	// 也可以先声明再逐个赋值
	var u4 User
	u4.Name = "小丽"
	u4.Age = 22
	fmt.Println("声明后赋值:", u4)
}

// ===== 示例 2 用的类型与方法 =====
type Counter struct {
	count int
}

// IncrementByValue 是值接收者：方法拿到的是结构体副本，修改不影响原对象
func (c Counter) IncrementByValue() {
	c.count++
}

// IncrementByPointer 是指针接收者：方法拿到原对象地址，修改会生效
func (c *Counter) IncrementByPointer() {
	c.count++
}

// Count 是只读方法，用值接收者即可
func (c Counter) Count() int {
	return c.count
}

// ===== 示例 2：值接收者 vs 指针接收者 =====
func valueVsPointerReceiver() {
	c := Counter{}
	c.IncrementByValue() // 改了副本，原对象不变
	fmt.Println("值接收者调用后:", c.Count())

	c.IncrementByPointer() // 改的是原对象
	fmt.Println("指针接收者调用后:", c.Count())

	// 何时用哪个（经验法则）：
	// 1. 方法需要修改接收者 -> 必须用指针接收者；
	// 2. 结构体很大，拷贝成本高 -> 用指针接收者；
	// 3. 同一个类型的方法最好统一用一种接收者，避免混用造成困惑；
	// 4. 只读小结构体用值接收者，语义上更安全（天然防止意外修改）。
}

// ===== 示例 3 用的类型与方法 =====
type Animal struct {
	Name string
}

func (a Animal) Speak() string {
	return a.Name + " 发出了叫声"
}

type Dog struct {
	Animal // 匿名字段嵌入：Dog 自动“继承” Animal 的字段和方法
	Breed  string
}

// ===== 示例 3：匿名字段嵌入与方法提升 =====
func embeddingDemo() {
	d := Dog{
		Animal: Animal{Name: "旺财"},
		Breed:  "柴犬",
	}

	// 方法提升：Dog 本身没有定义 Speak，但可以直接调用内嵌 Animal 的方法
	fmt.Println(d.Speak())

	// 字段也被提升，可以直接访问 d.Name；等价于 d.Animal.Name
	fmt.Println("名字:", d.Name, "品种:", d.Breed)

	// ⚠️ 注意：嵌入是“组合”而非真正的继承，Animal 不知道 Dog 的存在，
	// 不能用 Dog 替换需要 Animal 的地方，Go 没有继承体系。
}

// ===== 示例 4 用的类型 =====
type Product struct {
	ID    int     `json:"id"`              // 序列化时键名改为小写 id
	Name  string  `json:"name"`            // 序列化时键名改为 name
	Price float64 `json:"price"`           // 序列化时键名改为 price
	Note  string  `json:"-"`               // "-" 表示序列化时忽略该字段
	token string  `json:"token,omitempty"` // 小写字段不参与序列化，tag 无效但合法
}

// ===== 示例 4：字段标签与 JSON =====
func structTagDemo() {
	p := Product{ID: 1, Name: "机械键盘", Price: 299.5, Note: "内部备注"}

	// struct tag 是写在字段后的字符串元数据，encoding/json 会读取它
	data, err := json.Marshal(p)
	if err != nil {
		fmt.Println("序列化失败:", err)
		return
	}
	fmt.Println("JSON 输出:", string(data)) // 注意 Note 被忽略、token 未导出

	// 反序列化回结构体
	var p2 Product
	if err := json.Unmarshal([]byte(`{"id":2,"name":"鼠标","price":99}`), &p2); err != nil {
		fmt.Println("反序列化失败:", err)
		return
	}
	fmt.Printf("反序列化结果: %+v\n", p2)
}

// ===== 示例 5：结构体可比较性 =====
func comparabilityDemo() {
	// 所有字段都可比较（数值、字符串、布尔、指针、数组等）时，
	// 结构体可以用 == 比较，比较规则是逐字段相等
	a := User{Name: "小明", Age: 20}
	b := User{Name: "小明", Age: 20}
	fmt.Println("a == b:", a == b)

	// ⚠️ 注意：只要有一个字段不可比较（切片、map、函数），
	// 整个结构体就不能用 ==，下面这行会编译错误：
	// type Team struct { Members []string }
	// fmt.Println(Team{} == Team{}) // 编译失败：invalid operation

	// 此时可以用 reflect.DeepEqual（性能较差），或自己写比较逻辑
	type Team struct{ Members []string }
	t1 := Team{Members: []string{"a", "b"}}
	t2 := Team{Members: []string{"a", "b"}}
	fmt.Println("逐字段比较切片:", sliceEqual(t1.Members, t2.Members))
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
