# Solo Coder 填表数据

## 87-go-rate-limiter — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 用 Go 写一个 HTTP 限流中间件库，支持滑动窗口和令牌桶两种模式。实现 net/http.Handler 接口。 两种模式： 1. 滑动窗口模式（SlidingWindow）：统计最近 1 秒内的请求数（不是固定 1 秒窗口），超过 rate 则返回 429 2. 令牌桶模式（TokenBucket）：按固定速率（每秒 rate 个）往桶里放令牌，桶容量为 burst。请求消耗一个令牌，没有就返回 429 3. 通过 limiter.New(mode, rate, burst) 创建限流器，mode 为 "sliding_window" 或 "token_bucket" 按 Key 限流： 4. limiter.WithKey(key) 返回针对特定 key 的限流器。每个 key 独立计数，互不影响 5. 常见用法：按客户端 IP 限流，从请求的 RemoteAddr 提取 IP 6. 全局最大并发数：当所有 key 的总请求数超过 globalMax 时也返回 429 中间件行为： 7. 被限流的请求返回 HTTP 429 状态码，响应体为 JSON：{"error": "too many requests", "retry_after": 1}（retry_after 是建议客户端等待的秒数） 8. 限流计数和令牌桶的状态在内存中维护，进程重启后重置 统计： 9. limiter.Stats() 返回（总请求数, 通过数, 被限流数） 10. Stats() 是线程安全的，可以被多个 goroutine 并发调用 代码分 limiter.go、sliding_window.go、token_bucket.go 几个 package，go build 能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 库/SDK |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：token_bucket.go 第 65 行定义了自定义 min(a, b float64) float64 函数，但 go.mod 指定 go 1.21，该版本已内置 min built-in（支持 cmp.Ordered 类型包括 float64），导致编译报错 min redeclared in this block，项目无法通过 go build。KeyedLimiter.Allow() 未递增 parent.total 计数，通过 keyed 限流器处理请求时 Stats() 返回的 total 不等于 allowed 加 blocked。全局并发数检查存在 TOCTOU 竞态——checkGlobalLimit() 和 incGlobalActive() 分两次加锁执行，高并发下 globalActive 可能超过 globalMax。过程不满意：自定义 min 函数与 Go 1.21 内置 min 冲突，写完后没有执行 go build 验证编译 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 87-go-rate-limiter |

---

## 87-go-rate-limiter — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 我 go build 了下，token_bucket.go 报错了，说 min redeclared in this block，应该是 Go 1.21 内置的 min 跟你自定义的冲突了。另外我测了下 Stats() 返回的数字不太对，total 不等于 allowed 加 blocked，你看看 keyed limiter 那边是不是漏记了 total。还有 globalMax 那块，高并发压测的时候感觉超了也没拦住，check 和 inc 之间是不是有竞态问题 |
| 任务类型 | Bug修复 |
| 业务领域 | 库/SDK |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：RateLimiter 的 next handler 字段（limiter.go 第38行）是私有字段且没有公开的 setter 方法（如 Use 或 Middleware），无法将限流器作为标准 HTTP 中间件链式使用。ServeHTTP 中 rl.next != nil（第152行）永远为 false，被允许通过的请求无法转发到实际处理函数，客户端收到空 200 响应。PROMPT 明确要求"HTTP 限流中间件库"和"实现 net/http.Handler 接口"，当前无法完成中间件的核心职责（请求转发）。产物不满意：SlidingWindowLimiter 和 TokenBucketLimiter 的 Allow() 方法内部调用 baseLimiter 的 incTotal/incAllowed/incBlocked（额外加锁），同时 KeyedLimiter.Allow() 也对 parent 相同计数器递增，通过 KeyedLimiter 使用时每次请求触发两次计数器递增和两次锁操作，inner 的 baseLimiter 计数器从未被暴露使用（Stats 返回 parent 的），纯属不必要的开销 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 87-go-rate-limiter |

---
