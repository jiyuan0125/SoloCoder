# Solo Coder 填表数据

## 56-rust-mpmc-queue — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | .335769888099319:ee778b9e64b492736fdced86231e4215_69f58d5dd0eab67395271c9a.69f58d7ad0eab67395271c9e.69f58d7a5e5d9b2bb12eed43:Trae CN.T(2026/5/2 13:36:58) |
| 第一轮Session ID | .335769888099319:ee778b9e64b492736fdced86231e4215_69f58d5dd0eab67395271c9a.69f58d7ad0eab67395271c9e.69f58d7a5e5d9b2bb12eed43:Trae CN.T(2026/5/2 13:36:58) |
| 轮次 | 1 |
| User Prompt | 用 Rust 写一个多生产者多消费者（MPMC）的有界队列。支持阻塞的 push/pop 和非阻塞的 try_push/try_pop。队列容量在创建时指定。内部用固定大小的数组和两个原子索引（head/tail）实现，每个槽位用 cache line 对齐（#[repr(align(64))]）避免 false sharing。\n\n功能要求：\n1. push 如果队列满则阻塞等待直到有空位，pop 如果队列空则阻塞等待直到有数据\n2. try_push/try_pop 对应的非阻塞版本，满了/空了立即返回错误不阻塞\n3. 关闭队列时（调用 close），所有阻塞在 push/pop 上的线程被唤醒：push 返回 Closed 错误，pop 返回 Closed 错误，队列中还没被消费的数据丢弃\n4. close 后再调用 push 或 try_push 都返回 Closed 错误\n5. 提供一个 len() 方法返回当前队列中的元素数量\n6. 支持 Iterator：调用 into_iter() 消费队列中所有剩余元素，迭代过程中不能再 push\n\n代码分队列核心、原子操作封装几个模块，cargo build 能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 致命 bug：CachePadded\<T\> 使用 UnsafeCell 导致 Inner\<T\> 不满足 Sync，Queue/Sender/Receiver 均不满足 Send，无法跨线程使用。MPMC 队列完全无法多线程使用，违背核心设计目标。 |
| github地址 | |
| 分支/文件夹 | 56-rust-mpmc-queue |

---

## 56-rust-mpmc-queue — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | .335769888099319:3d7291b513a821768f284fa63c276b0e_69f58d5dd0eab67395271c9a.69f59b18d0eab67395271ec1.69f59b185e5d9b2bb12eed44:Trae CN.T(2026/5/2 14:35:04) |
| 第一轮Session ID | .335769888099319:ee778b9e64b492736fdced86231e4215_69f58d5dd0eab67395271c9a.69f58d7ad0eab67395271c9e.69f58d7a5e5d9b2bb12eed43:Trae CN.T(2026/5/2 13:36:58) |
| 轮次 | 2 |
| User Prompt | 队列无法跨线程使用，编译报错 `UnsafeCell cannot be shared between threads safely`。\n\n核心问题：`CachePadded<T>` 里用了 `UnsafeCell`，导致 `Inner<T>` 不满足 `Sync`，`Arc<Inner<T>>` 不满足 `Send`，所以 `Queue`/`Sender`/`Receiver` 都不能传到其他线程。\n\n需要给 `Inner<T>` 或 `CachePadded<T>` 加上 `unsafe impl Sync`（前提是线程安全确实由原子操作和 ready flag 保证了），修复后多线程的 push/pop/close 唤醒才能真正跑通。 |
| 任务类型 | Bug修复 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | R1 的 Send/Sync 问题已修复（unsafe impl），容量 off-by-one 也修好了。但 do_push 存在数据竞争：多生产者并发时读取相同 tail 值，写入同一槽位导致数据丢失。4p4c 压测 4000 条消息丢失 1429 条（64% 交付率）。需要用 compare_exchange CAS 原子操作抢占槽位后再写入。 |
| github地址 | |
| 分支/文件夹 | 56-rust-mpmc-queue |

---

## 56-rust-mpmc-queue — 第 3 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | .335769888099319:004a910252964eb4b25005b2a5f63bff_69f58d5dd0eab67395271c9a.69f5a0b9d0eab67395271f7f.69f5a0b95e5d9b2bb12eed45:Trae CN.T(2026/5/2 14:59:05) |
| 第一轮Session ID | .335769888099319:ee778b9e64b492736fdced86231e4215_69f58d5dd0eab67395271c9a.69f58d7ad0eab67395271c9e.69f58d7a5e5d9b2bb12eed43:Trae CN.T(2026/5/2 13:36:58) |
| 轮次 | 3 |
| User Prompt | R1 的 Send/Sync 和容量问题修好了，但多线程压测暴露了数据竞争 bug。\n\n用 4 个生产者线程并发 push，4 个消费者线程并发 pop，总共发 4000 条消息，实际只收到 2571 条，丢了 1429 条（35.7%）。\n\n根本原因：`do_push` 里多个生产者能读到相同的 tail 值，然后都往同一个 slot 写数据，后写的覆盖先写的，导致数据丢失。`do_pop` 也有类似问题。\n\n修复：`do_push` 和 `do_pop` 都要用 `compare_exchange`（CAS）原子操作来抢占 head/tail 槽位，CAS 成功后才能写入/读取对应 slot。具体来说：\n1. 循环读取当前 tail\n2. 计算 next_tail，检查是否 full\n3. 用 `compare_exchange_weak(tail, next_tail)` 尝试原子更新 tail\n4. CAS 成功 → 写入 slot 值 → 设 ready=true\n5. CAS 失败 → 重试\n\ndo_pop 同理用 CAS 抢占 head。 |
| 任务类型 | Bug修复 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | R2 的 35% 丢包修复为 0.12%，但仍有数据丢失。4p4c 20000 条压测：20000 条入队，仅 19977 条出队，23 条永久卡在队列中，1 个消费者线程永久挂起（spin on ready=false）。根因：do_pop 中 spin-wait on ready flag 在 buffer 多次 wrap-around 后可读到 stale 状态，导致消费者 CAS 了 head 但永远等不到 ready=true。 |
| github地址 | |
| 分支/文件夹 | 56-rust-mpmc-queue |

---

## 评测详情

### R3 修复了什么

1. ✅ **do_push CAS**：`compare_exchange_weak(tail, next_tail)` 原子抢占槽位后再写入
2. ✅ **do_pop CAS**：`compare_exchange_weak(head, next_head)` 原子抢占槽位后 spin-wait on ready
3. ✅ **CachePaddedAtomicUsize 新增 compare_exchange_weak**
4. ✅ **内置压测**：4p4c 4000 条 0 丢失 0 重复
5. ✅ **容量正确**：Queue::new(4) 真正能存 4 个元素

### 剩余 Bug：ready flag 在 buffer wrap-around 后 stale

`do_pop` CAS 抢占 head 后 spin-wait `slot.ready`。当 buffer 经过多次 wrap-around 后：
- 某个 slot 的 `ready` 在上一轮已被 consumer 设为 false
- 新 producer CAS tail 到该 slot 并写入新值
- 但在 producer 设 `ready=true` 之前，另一个 consumer 可能已 CAS head 到同一 slot 并开始 spin
- 如果该 consumer 看到的是旧的 `ready=false`（上一轮残留），而新 producer 的 `ready=true` 被另一个 consumer 读取并设回 false
- 则该 consumer 永远 spin 在 `ready=false`，永远不会返回

**压测证据**（带进度监控）：
```
pushed=20000 popped=19977 elapsed=101ms  (所有 producer 在 5ms 内完成)
pushed=20000 popped=19977 elapsed=10s    (消费者永久卡住，23 条无法出队)
```

### 6 项 PROMPT 需求验证

| # | 需求 | 状态 | 说明 |
|---|------|------|------|
| 1 | 阻塞 push/pop | ⚠️ | 低并发正确，高并发丢消息+挂线程 |
| 2 | 非阻塞 try_push/try_pop | ✅ | 正确 |
| 3 | close() 唤醒阻塞线程 | ✅ | 正确 |
| 4 | close 后 push/try_push 返回 Closed | ✅ | 正确 |
| 5 | len() 方法 | ✅ | 正确 |
| 6 | into_iter() | ✅ | 正确 |

### 三轮改进趋势

| 轮次 | 4p4c 4000msg 丢失率 | 主要问题 |
|------|----------------------|----------|
| R1 | 无法跨线程 | UnsafeCell → 不满足 Send/Sync |
| R2 | 64.7% (2571/4000) | do_push 无 CAS，同 slot 覆盖 |
| R3 | 0.12% (23/20000) | CAS 修复了大部分，但 ready flag stale |

---

## 56-rust-mpmc-queue — 第 4 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | .335769888099319:2796acd463617aad5d8e48a918791060_69f58d5dd0eab67395271c9a.69f5af92d0eab6739527201c.69f5af925e5d9b2bb12eed46:Trae CN.T(2026/5/2 16:02:26) |
| 第一轮Session ID | .335769888099319:ee778b9e64b492736fdced86231e4215_69f58d5dd0eab67395271c9a.69f58d7ad0eab67395271c9e.69f58d7a5e5d9b2bb12eed43:Trae CN.T(2026/5/2 13:36:58) |
| 轮次 | 4 |
| User Prompt | CAS 修复后丢包率从 64% 降到了 0.12%，进步很大，但还有问题。\n\n我做了个带进度监控的压测：4p4c 发 20000 条消息到 cap=100 的队列，所有 producer 在 5ms 内就完成了，但 10 秒后消费者只收到 19977 条，23 条卡在队列里出不来，1 个消费者线程永久挂起（在 spin-wait ready=false）。\n\n问题在 do_pop 里：consumer CAS 抢占了 head 后 spin-wait `slot.ready`。当 buffer 经历很多次 wrap-around，某个 slot 的 ready flag 可能处于 stale 状态。consumer CAS head 到这个 slot 后，看到的 ready=false 可能是上一轮 consumer 残留的 false，而不是当前 producer 还没来得及设的 true。producer 设了 true 但被另一个 consumer 读了并设回 false，那这个 spin 就永远等不到 true 了。\n\n修复思路：在 do_push 里，CAS tail 成功后先设 `ready = false`，再写 value，最后设 `ready = true`。这样 consumer 不会看到上一轮残留的 true。或者考虑换个同步机制，比如用 epoch/seqlock 来避免 ready flag 的 stale 问题。 |
| 任务类型 | Bug修复 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 | |
| github地址 | |
| 分支/文件夹 | 56-rust-mpmc-queue |

---

## 评测详情（R4 最终）

### R4 核心变更：ready flag → seqlock

R4 彻底重写了槽位同步机制，用 **sequence counter**（seqlock）替代了原来的 `ready: AtomicBool`：

- 每个 `Slot<T>` 新增 `seq: CachePaddedAtomicUsize`，初始值 0
- Producer CAS tail 成功后，等待 `seq == epoch * 2`（偶数=可写），写入后设 `seq = epoch * 2 + 1`（奇数=可读）
- Consumer CAS head 成功后，等待 `seq == epoch * 2 + 1`（奇数=可读），读取后设 `seq = (epoch + 1) * 2`（偶数=可写）
- seq 单调递增，每个 epoch 递增 2，天然防止 stale read

这是经典的 Dmitry Vyukov bounded MPMC queue 设计，经过学术界和工业界充分验证。

### 压测结果

| 测试 | 配置 | 结果 |
|------|------|------|
| 内置 4p4c | 20000 msg, cap=100 | ✅ 0 丢失 0 重复 |
| EXT1 4p4c | 20000 msg, cap=100 | ✅ 0 丢失 0 重复 |
| EXT2 4p4c | 100000 msg, cap=50 | ✅ 0 丢失 0 重复 |
| EXT3 close-while-push | 4p, cap=10, close mid-flight | ✅ 10 push 10 pop, 70µs |
| EXT4 close-empty | consumer waits on empty queue | ✅ Closed in 102µs |
| EXT5 8p8c | 50000 msg, cap=32 | ✅ 0 丢失 0 重复 |
| EXT6 2p2c | 10000 msg, cap=3 | ✅ 0 丢失 0 重复 |
| EXT7 into_iter | push then iterate | ✅ 正确 |
| EXT8 Send/Sync | compile-time bounds check | ✅ 全部通过 |

### 6 项 PROMPT 需求验证

| # | 需求 | 状态 | 说明 |
|---|------|------|------|
| 1 | 阻塞 push/pop | ✅ | 8p8c 5万条零丢失 |
| 2 | 非阻塞 try_push/try_pop | ✅ | 正确 |
| 3 | close() 唤醒阻塞线程 | ✅ | 空队列等待也能唤醒 |
| 4 | close 后 push/try_push 返回 Closed | ✅ | 正确 |
| 5 | len() 方法 | ✅ | 正确 |
| 6 | into_iter() | ✅ | 正确 |

### 四轮改进趋势

| 轮次 | 4p4c 20K 丢失率 | 主要问题 |
|------|------------------|----------|
| R1 | 无法跨线程 | UnsafeCell → 不满足 Send/Sync |
| R2 | 64.7% (2571/4000) | do_push 无 CAS，同 slot 覆盖 |
| R3 | 0.12% (23/20000) | CAS 修复了大部分，但 ready flag stale |
| R4 | 0% (0/200000+) | seqlock 彻底解决 |

### 其他问题

1. **dead_code 警告**：`CachePaddedAtomicBool` 整个类型未使用，`CachePaddedAtomicUsize` 的 `store`/`swap`/`fetch_add` 未使用
2. **代码重复**：Queue/Sender/Receiver 约 200 行重复的 push/pop 实现
