## 122-c-atomic-counter — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 帮我用 C 写一个高性能的请求计数器模块，用于网关层的限流。网关要对每个 IP 地址（或每个API 路径）的请求频率做限制，超过阈值的请求直接拒绝。基本计数：每个 key 维护一个计数器，支持原子加 1 和读取当前值。滑动窗口限流：固定窗口和滑动窗口两种模式。key 自动清理：超过 5 分钟没操作的 key 自动移除。统计查询：活跃 key 数、总请求量、被拒绝请求量、某 key 近 1 分钟请求次数和被拒绝次数。性能要求：用 CPU 原子指令实现无锁计数。隐藏坑：滑动窗口小窗口切换时间点要精确对齐。用纯 C 写，gcc 在 Linux 上编译通过。代码要拆成至少 4 个源文件。 |
| 任务类型 | 新功能开发 |
| 业务领域 | 库/SDK |
| 修改范围 | 全新项目 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：1) global_total_requests 只计数允许的请求而非总请求量，语义错误；2) 固定窗口模式缺少 rejected 计数器，get_key_stats 的 rejected_last_minute 始终为 0；3) collect_expired_callback 死代码未清理；4) counter_get_or_create 竞态路径 free 前未调 pthread_spin_destroy。过程不满意：编译有 -Wunused-function warning 但未处理。 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 122-c-atomic-counter |

---
