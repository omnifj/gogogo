//go:build ignore

// ============================================================
// 文件：01_reflection.go
// 阶段：高级
// 主题：反射
// 学习目标：
//   1. 掌握 reflect.TypeOf / reflect.ValueOf 的基本用法
//   2. 区分 Kind 与 Type 的概念差异
//   3. 遍历结构体字段与方法、读取 struct tag
//   4. 通过可寻址 Value 安全修改变量，了解反射的性能代价
// 运行方式：go run 03-advanced/01_reflection.go
// 预计阅读时间：20 分钟
// ============================================================

package main

import (
	"fmt"
	"reflect"
	"time"
)

// 用于演示的结构体，字段上带有 struct tag
type User struct {
	Name  string `json:"name" validate:"required"`
	Age   int    `json:"age"`
	email string // 小写字段不可导出，反射无法读写
}

// Greet 是一个导出方法，反射遍历方法时可以看到它
func (u User) Greet() string {
	return fmt.Sprintf("你好，我是 %s，%d 岁", u.Name, u.Age)
}

// ===== 示例 1：TypeOf 与 ValueOf =====
func demoTypeAndValue() {
	fmt.Println("========== 示例 1：TypeOf 与 ValueOf ==========")
	var x any = 3.14

	// TypeOf 拿到动态类型信息（静态结构：类型元数据）
	t := reflect.TypeOf(x)
	// ValueOf 拿到动态值信息（运行时的具体值）
	v := reflect.ValueOf(x)

	fmt.Println("Type:", t)          // 完整类型名
	fmt.Println("Kind:", t.Kind())   // 底层类别（预定义的枚举）
	fmt.Println("Value:", v)         // 反射值可直接打印
	fmt.Println("Float:", v.Float()) // 按具体种类取出数值

	// ⚠️ 注意：reflect.ValueOf(nil) 得到的是零 Value，调用其方法会 panic
	var p *int
	vp := reflect.ValueOf(p)
	fmt.Println("nil 指针的 Value 是否有效:", vp.IsValid(), "是否为 nil:", vp.IsNil())
	fmt.Println()
}

// ===== 示例 2：Kind 与 Type 的区别 =====
func demoKindVsType() {
	fmt.Println("========== 示例 2：Kind 与 Type 的区别 ==========")
	// 自定义类型 Celsius，底层类型是 float64
	type Celsius float64
	var c Celsius = 36.5
	var f float64 = 36.5

	tc, tf := reflect.TypeOf(c), reflect.TypeOf(f)
	// Type 是精确类型（含包路径与类型名），两者不同
	fmt.Println("Celsius 的 Type:", tc, "| float64 的 Type:", tf)
	fmt.Println("两者 Type 相同吗:", tc == tf)
	// Kind 是底层类别，两者都是 float64
	fmt.Println("两者 Kind 相同吗:", tc.Kind() == tf.Kind())

	// ⚠️ 注意：switch 类型判断时若想兼容自定义类型，应比较 Kind 而非 Type
	switch tc.Kind() {
	case reflect.Float64:
		fmt.Println("按 Kind 判断：Celsius 也算 float64 家族")
	}
	fmt.Println()
}

// ===== 示例 3：遍历结构体字段与方法 =====
func demoStructIterate() {
	fmt.Println("========== 示例 3：遍历结构体字段与方法 ==========")
	u := User{Name: "小明", Age: 18}
	t := reflect.TypeOf(u)
	v := reflect.ValueOf(u)

	// NumField 遍历字段（含不可导出字段）
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fmt.Printf("字段 %-5s 类型 %-7s 可导出: %-5v", f.Name, f.Type, f.IsExported())
		// ⚠️ 注意：不可导出字段调用 Interface() 会 panic，必须先判断
		if f.IsExported() {
			fmt.Printf(" 值: %v", v.Field(i).Interface())
		} else {
			fmt.Print(" 值: <不可读取>")
		}
		fmt.Println()
	}

	// NumMethod 遍历导出方法（值接收者的方法都能被值类型看到）
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		fmt.Println("方法:", m.Name, "签名:", m.Type)
	}
	// 通过反射调用方法：MethodByName 返回可调用 Value
	greet := v.MethodByName("Greet").Call(nil)
	fmt.Println("反射调用 Greet:", greet[0])
	fmt.Println()
}

// ===== 示例 4：通过可寻址 Value 安全修改变量 =====
func demoSetValue() {
	fmt.Println("========== 示例 4：反射修改变量 ==========")
	u := User{Name: "小明", Age: 18}

	// ⚠️ 注意：reflect.ValueOf(u) 得到的是副本，不可寻址，Set 会 panic
	// 必须传入指针再 Elem()，才能得到指向原变量的可寻址 Value
	v := reflect.ValueOf(&u).Elem()
	fmt.Println("CanSet（可修改）:", v.CanSet())

	age := v.FieldByName("Age")
	if age.IsValid() && age.CanSet() {
		age.SetInt(20) // 类型必须匹配，否则 panic
	}
	name := v.FieldByName("Name")
	if name.IsValid() && name.CanSet() {
		name.SetString("小红")
	}
	fmt.Println("修改后:", u.Greet())

	// ⚠️ 注意：小写字段 email 的 CanSet 为 false，强行 Set 会 panic
	fmt.Println("email 字段 CanSet:", v.FieldByName("email").CanSet())
	fmt.Println()
}

// ===== 示例 5：读取 struct tag =====
func demoStructTag() {
	fmt.Println("========== 示例 5：读取 struct tag ==========")
	t := reflect.TypeOf(User{})
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// StructTag.Get 按 key 取值，JSON 序列化、参数校验库都靠这个机制
		fmt.Printf("字段 %-6s json=%-8q validate=%q\n",
			f.Name, f.Tag.Get("json"), f.Tag.Get("validate"))
	}
	fmt.Println("提示：encoding/json 正是用反射读 tag 来决定序列化字段名")
	fmt.Println()
}

// 直接调用，用于与反射调用对比性能
func addDirect(a, b int) int { return a + b }

// ===== 示例 6：反射的性能代价 =====
func demoPerformance() {
	fmt.Println("========== 示例 6：反射调用的性能代价 ==========")
	const n = 1_000_000

	// 基线：直接调用
	start := time.Now()
	sum := 0
	for i := 0; i < n; i++ {
		sum += addDirect(1, 2)
	}
	directCost := time.Since(start)

	// 反射调用：Call 需要装箱参数、检查类型、再拆箱结果
	fn := reflect.ValueOf(addDirect)
	args := []reflect.Value{reflect.ValueOf(1), reflect.ValueOf(2)}
	start = time.Now()
	for i := 0; i < n; i++ {
		out := fn.Call(args)
		sum += int(out[0].Int())
	}
	reflectCost := time.Since(start)

	// 提示：耗时数值每次运行可能不同，取决于机器负载
	fmt.Printf("直接调用 %d 次耗时: %v\n", n, directCost)
	fmt.Printf("反射调用 %d 次耗时: %v\n", n, reflectCost)
	// ⚠️ 注意：直接调用太快时耗时可能为 0，不能拿来做除数
	if directCost > 0 {
		fmt.Printf("反射大约慢 %.1f 倍\n", float64(reflectCost)/float64(directCost))
	} else {
		fmt.Println("直接调用耗时低于计时精度（0s），但反射耗时仍可见，数量级差距明显")
	}
	fmt.Printf("校验和 %d（防止循环被编译器优化掉）\n", sum)
	fmt.Println("结论：热路径避免反射；配置解析、序列化等低频场景可放心使用")
}

func main() {
	demoTypeAndValue()
	demoKindVsType()
	demoStructIterate()
	demoSetValue()
	demoStructTag()
	demoPerformance()
}
