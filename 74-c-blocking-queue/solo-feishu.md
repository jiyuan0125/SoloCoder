# Solo Coder 填表数据

## 74-c-blocking-queue — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | .335769888099319:b58c71a9a4d1459189c834de12840b41_69f623be57c8d33337f237b3.69f623ee57c8d33337f237ed.69f623ec5346bab035e8570a:Trae CN.T(2026/5/3 00:18:54) |
| 第一轮Session ID | .335769888099319:b58c71a9a4d1459189c834de12840b41_69f623be57c8d33337f237b3.69f623ee57c8d33337f237ed.69f623ec5346bab035e8570a:Trae CN.T(2026/5/3 00:18:54) |
| 轮次 | 1 |
| User Prompt | 用 C 写一个有界阻塞队列，支持多生产者和多消费者。基于 mutex + condition variable 实现。 核心 API： 1. bq_create(capacity) 创建队列，capacity 是最大容量 2. bq_put(queue, item) 入队。队列满时阻塞等待有空位。返回 0 成功，-1 队列已关闭 3. bq_take(queue, out_item) 出队。队列空时阻塞等待有数据。返回 0 成功，-1 队列已关闭且为空 4. bq_close(queue) 关闭队列。唤醒所有阻塞在 take 上的线程（让它们返回 -1）。关闭后 put 返回 -1 5. bq_size(queue) 返回当前队列中的元素数量（不阻塞） 批量操作： 6. bq_batch_put(queue, items, count) 批量入队。如果剩余空间 < count 则全部拒绝（不部分放入），返回实际入队的数量（0 或 count）。队列关闭时返回 -1 7. bq_batch_take(queue, items, max_count) 批量出队。取出当前队列中的所有元素（最多 max_count 个），返回实际取出的数量。队列关闭且为空时返回 -1 线程安全： 8. 多个线程可以同时调用 put 和 take，不会有数据竞争 9. close 和 put/take 之间不能死锁：close 在持锁期间发信号唤醒等待线程，然后释放锁 要求： 10. 所有 API 返回 int 类型，语义同上 11. 队列元素是 void* 类型（通用指针），队列本身不管理元素的生命周期 代码分 blocking_queue.c、blocking_queue.h 两个文件，gcc -lpthread 编译能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：bq_batch_put 在剩余空间不足时阻塞等待而非立即返回 0，PROMPT 明确要求"如果剩余空间 < count 则全部拒绝（不部分放入）"，当前实现会永久阻塞直到有足够空间。bq_batch_take 同理在队列空时阻塞等待，PROMPT 说"取出当前队列中的所有元素（最多 max_count 个）"，队列空时应立即返回 0 而非阻塞。过程不满意：batch 操作的阻塞行为与 PROMPT 矛盾，写完后没有实际测试 batch_put/batch_take 在空间/数据不足时的行为 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 74-c-blocking-queue |

---
