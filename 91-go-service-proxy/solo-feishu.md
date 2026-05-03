# Solo Coder 填表数据

## 91-go-service-proxy — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 用 Go 写一个反向代理服务。监听本地端口，根据路径前缀转发到不同后端服务。 路由规则： 1. 配置文件（JSON）：定义路径前缀到后端地址列表的映射。如 {"/api": ["localhost:8081", "localhost:8082"], "/static": ["localhost:8083"]} 2. 请求路径匹配最长的前缀规则（/api/users 匹配 /api 规则，不匹配 /a 开头的规则） 3. 未匹配任何规则的请求返回 404 负载均衡： 4. 同一路由下多个后端实例用加权轮询（weighted round-robin）分发 5. weight 越大分到的流量越多。权重比例：weight=3 的实例分到的请求数是 weight=1 的 3 倍 6. 加权轮询算法：每个后端维护 currentWeight，选中后 currentWeight += effectiveWeight，选中 totalWeight 最大的。选中的后端 currentWeight -= totalWeight 熔断： 7. 某个后端连续 3 次请求失败（连接失败或返回 5xx）开启熔断，30 秒后进入 half-open 状态 8. half-open 状态放 1 个请求探测，成功则关闭熔断，失败则重新开始 30 秒等待 9. 熔断期间该后端的请求直接返回 503，不尝试连接 超时和取消： 10. 请求有 30 秒超时（可配置），超时后取消转发，向客户端返回 504 11. 客户端断开连接时（ctx.Done()）取消转发，后端不应继续处理 代码分 proxy.go、router.go、balancer.go、breaker.go 几个 package，go build 能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：breaker.go 的 Allow() 在 HalfOpen 状态下直接返回 true，没有限制只放 1 个探测请求，PROMPT 明确要求"half-open 状态放 1 个请求探测"，当前实现允许 half-open 期间无限请求通过。过程不满意：加权轮询 Select() 中 currentWeight 的修改在每个 backend 独立锁内完成但不是原子操作，高并发下权重分配不准确，写完后没有分析并发场景 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 91-go-service-proxy |

---

## 91-go-service-proxy — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 我测了下熔断这块，让某个后端连续报错触发熔断后等 30 秒进入 half-open，这时候如果同时来多个请求，它们全都能打到这个后端上去，PROMPT 说的是 half-open 只放 1 个请求探测。另外 balancer.go 里 Select 的加权轮询在高并发下好像不太对，每个 backend 的 currentWeight 是各自加锁改的，但 totalWeight 的计算和选中的判断不是原子的，并发多了权重分配会乱。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：Select() 遍历后端时对每个后端调用 Allow() 会触发 Open→HalfOpen 状态转换并设置 halfOpenPending=true，但加权轮询可能选中另一个后端，导致 HalfOpen 后端的 halfOpenPending 无法被重置，探测请求永远不会真正发出去，熔断无法恢复。proxy.go 第 60-62 行有一个无意义的 goroutine，只做 `<-ctx.Done()` 然后退出，是死代码。过程不满意：R1 的两个 bug（half-open 并发限制、balancer 并发安全）都已修复，但修复 Allow() 时引入了新的状态管理问题，没有分析 Allow() 和 Select() 的交互副作用 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 91-go-service-proxy |

---

