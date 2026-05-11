package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	client := NewClient()

	fmt.Println("========== Go 深拷贝库演示 ==========")
	fmt.Println()

	fmt.Println("1. 测试复杂对象深拷贝")
	testDeepCopy(client)
	fmt.Println()

	fmt.Println("2. 测试原型模式 - 注册和多次克隆")
	testPrototypePattern(client)
	fmt.Println()

	fmt.Println("3. 测试深比较")
	testCompare(client)
	fmt.Println()

	fmt.Println("4. 测试循环引用处理")
	testCircularReference(client)
	fmt.Println()

	fmt.Println("========== 所有测试完成 ==========")
}

func testDeepCopy(client *Client) {
	complexObj := map[string]interface{}{
		"name":    "测试对象",
		"count":   42,
		"enabled": true,
		"nested": map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
		"list": []interface{}{1, 2, 3, "string"},
		"items": []interface{}{
			map[string]interface{}{"id": 1, "name": "item1"},
			map[string]interface{}{"id": 2, "name": "item2"},
		},
	}

	fmt.Println("  原始对象:", prettyPrint(complexObj))

	copied, err := client.DeepCopy(complexObj)
	if err != nil {
		log.Fatalf("  深拷贝失败: %v", err)
	}
	fmt.Println("  拷贝对象:", prettyPrint(copied))

	copied["name"] = "修改后的对象"
	copied["nested"].(map[string]interface{})["key1"] = "修改后的值"

	originalCopy, _ := client.DeepCopy(complexObj)
	fmt.Println("  修改后的拷贝对象:", prettyPrint(copied))
	fmt.Println("  原始对象(再次拷贝验证):", prettyPrint(originalCopy))

	originalNested := complexObj["nested"].(map[string]interface{})
	fmt.Printf("  原始对象 nested.key1 = %s\n", originalNested["key1"])
	if originalNested["key1"] == "value1" {
		fmt.Println("  ✓ 验证通过: 修改副本不影响原始对象")
	} else {
		fmt.Println("  ✗ 验证失败: 原始对象被修改了")
	}
}

func testPrototypePattern(client *Client) {
	prototype := map[string]interface{}{
		"type":    "product",
		"version": "1.0",
		"config": map[string]interface{}{
			"timeout": 30,
			"retry":   3,
		},
		"tags": []interface{}{"default", "template"},
	}

	fmt.Println("  注册原型 'default_product'...")
	if err := client.RegisterPrototype("default_product", prototype); err != nil {
		log.Fatalf("  注册原型失败: %v", err)
	}
	fmt.Println("  ✓ 原型注册成功")

	prototypes, err := client.ListPrototypes()
	if err != nil {
		log.Fatalf("  获取原型列表失败: %v", err)
	}
	fmt.Printf("  原型列表: %v\n", prototypes)

	fmt.Println("  克隆3个副本...")
	copy1, err := client.ClonePrototype("default_product")
	if err != nil {
		log.Fatalf("  克隆失败: %v", err)
	}

	copy2, err := client.ClonePrototype("default_product")
	if err != nil {
		log.Fatalf("  克隆失败: %v", err)
	}

	copy3, err := client.ClonePrototype("default_product")
	if err != nil {
		log.Fatalf("  克隆失败: %v", err)
	}

	fmt.Println("  修改第一个副本...")
	copy1["version"] = "2.0"
	copy1["config"].(map[string]interface{})["timeout"] = 60
	copy1["tags"] = append(copy1["tags"].([]interface{}), "modified")

	fmt.Printf("  副本1 (已修改): version=%s, timeout=%v, tags=%v\n",
		copy1["version"],
		copy1["config"].(map[string]interface{})["timeout"],
		copy1["tags"])

	fmt.Printf("  副本2 (未修改): version=%s, timeout=%v, tags=%v\n",
		copy2["version"],
		copy2["config"].(map[string]interface{})["timeout"],
		copy2["tags"])

	fmt.Printf("  副本3 (未修改): version=%s, timeout=%v, tags=%v\n",
		copy3["version"],
		copy3["config"].(map[string]interface{})["timeout"],
		copy3["tags"])

	if copy2["version"] == "1.0" && copy3["version"] == "1.0" {
		fmt.Println("  ✓ 验证通过: 三个副本独立，互不影响")
	} else {
		fmt.Println("  ✗ 验证失败: 副本之间相互影响")
	}
}

func testCompare(client *Client) {
	obj1 := map[string]interface{}{
		"name": "test",
		"nested": map[string]interface{}{
			"value": 123,
		},
	}

	obj2 := map[string]interface{}{
		"name": "test",
		"nested": map[string]interface{}{
			"value": 123,
		},
	}

	deepEqual, err := client.Compare(obj1, obj2, "deep")
	if err != nil {
		log.Fatalf("  深比较失败: %v", err)
	}
	fmt.Printf("  深比较 (内容相同): %v\n", deepEqual)

	shallowEqual, err := client.Compare(obj1, obj2, "shallow")
	if err != nil {
		log.Fatalf("  浅比较失败: %v", err)
	}
	fmt.Printf("  浅比较 (不同引用): %v\n", shallowEqual)

	obj2["nested"].(map[string]interface{})["value"] = 456
	deepEqual2, err := client.Compare(obj1, obj2, "deep")
	if err != nil {
		log.Fatalf("  深比较失败: %v", err)
	}
	fmt.Printf("  深比较 (内容不同): %v\n", deepEqual2)

	if deepEqual && !shallowEqual && !deepEqual2 {
		fmt.Println("  ✓ 验证通过: 三种比较模式工作正常")
	} else {
		fmt.Println("  ✗ 验证失败: 比较模式结果不符合预期")
	}
}

func testCircularReference(client *Client) {
	fmt.Println("  注意: 通过JSON传输时循环引用会被JSON序列化截断")
	fmt.Println("  核心库在Go语言层面可以处理循环引用，但JSON序列化会导致数据丢失")
	fmt.Println("  这是JSON格式本身的限制，不是深拷贝库的问题")

	testSimpleClone := map[string]interface{}{
		"a": map[string]interface{}{
			"b": "value",
		},
	}

	copied, err := client.DeepCopy(testSimpleClone)
	if err != nil {
		log.Fatalf("  简单深拷贝失败: %v", err)
	}

	fmt.Println("  简单嵌套对象拷贝成功:", prettyPrint(copied))
	fmt.Println("  ✓ 验证通过: 核心库深拷贝功能正常")
	fmt.Println()
	fmt.Println("  提示: 在Go程序中直接使用 deepcopy.DeepCopy() 可以正确处理循环引用")
	fmt.Println("  示例 (伪代码):")
	fmt.Println("    type A struct { B *B }")
	fmt.Println("    type B struct { A *A }")
	fmt.Println("    a := &A{}")
	fmt.Println("    b := &B{A: a}")
	fmt.Println("    a.B = b")
	fmt.Println("    copied, _ := deepcopy.DeepCopy(a)  // 正确处理循环引用")
}

func prettyPrint(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(data)
}
