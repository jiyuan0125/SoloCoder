# Solo Coder 填表数据

## 95-go-crdt-set — 第 4 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | （待补充） |
| 第一轮Session ID | （待补充） |
| 轮次 | 4 |
| User Prompt | 用 Go 写一个 CRDT OR-Set（Observed-Remove Set）实现。支持 add、remove、merge 和持久化。 核心语义： 1. add(element)：添加元素。每次 add 生成一个唯一的 tag（用 uint64 计数器递增） 2. remove(element)：删除元素。记录被删除的 tag 集合（remove-tags） 3. 元素存在条件：元素的 add-tags 中至少有一个 tag 不在任何 remove-tags 中（即存在一次未被删除的 add） 4. 关键语义：节点 A 先 add("x", tag=1)，remove("x", tags={1})，再 add("x", tag=2)。对于没看到 tag=1 的节点 B，它只看到 remove(tags={1}) 和 add(tag=2)，所以 B 看不到 "x" Merge 规则： 5. merge(otherSet)：两个集合的元素取并集，add-tags 和 remove-tags 分别合并 6. 并发 add 同一个元素后 merge，集合中该元素只有一个实例 API： 7. set.Add(element) / set.Remove(element) / set.Contains(element) bool / set.Elements() []string 8. set.Merge(other *ORSet) / set.Clone() *ORSet 9. set.MarshalJSON() ([]byte, error) / set.UnmarshalJSON(data []byte) error 序列化： 10. JSON 格式：{"elements": [{"value": "x", "add_tags": [1,3], "remove_tags": [2]}]} 11. merge 时如果 value 相同则合并 add_tags 和 remove_tags 数组 测试： 12. 提供 MergeSimulator 函数：模拟 A、B 两个节点各自操作后再 merge，验证结果符合 CRDT OR-Set 语义 代码分 crdt.go 几个 package，go build 能过。 |
| 任务类型 | Bug修复 |
| 业务领域 | 库/SDK |
| 修改范围 | 模块内多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：TestORSet_KeySemantic_OutOfOrderMessages 测试在 R3 被删除而非修复，PROMPT 第 7 行明确标记的"关键语义"（乱序消息场景下 B 不应看到 "x"）缺少测试覆盖，OR-Set merge 对乱序消息的处理逻辑未经验证。过程不满意：连续 4 轮未通过，R3 选择删除测试来回避 bug 而非修复实现，违反基本的工程实践 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 95-go-crdt-set |

---
