# Solo Coder 填表数据

## 448-go-realtime-dashboard — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 运营看板需要实时数据推送。做一个实时看板API，客户端订阅感兴趣的数据指标，服务端有新数据时主动推送。客户端可设刷新频率最小5秒防止请求过于频繁。客户端断开后5分钟内重连要补发断连期间错过的数据，超过5分钟的只推送最新值不补发历史。同时在线最多100个客户端，超过拒绝新连接并返回当前排队位置。查询当前在线客户端列表和各指标最新值。某指标超过1小时没新数据推送时要标注可能已过期提醒使用者注意数据时效性。指标数据支持设置告警阈值——超过阈值自动推送告警消息给所有订阅该指标的客户端。历史数据保留最近7天，超过7天的自动清理。提供数据趋势预览——最近1小时每5分钟一个数据点。订阅和指标超过5分钟的只推送最新值不补发历史。指标数据支持设置告警阈值——超过阈值自动推送告警消息给所有订阅该指标的客户端。历史数据保留最近7天，超过7天的自动清理。提供数据趋势预览——最近1小时每5分钟一个数据点。客户端断开连接时自动取消订阅释放资源指标数据支持同比环比对比分析。看板布局可配置——用户可以自定义看板上显示哪些指标和排列顺序。数据推送延迟监控——超过阈值自动告警。指标元数据管理——每个指标包含名称、单位、描述和计算公式。。 分两个独立程序。服务端启动后监听端口对外提供HTTP服务，负责所有业务逻辑，数据存在内存里。客户端是命令行工具，用户通过命令行发请求给服务端，服务端处理完返回结果，客户端展示给用户。两个程序各自独立main包，通信格式定义放在公共包。各程序内部按职责分文件。先go mod init再开发，go build ./...能编译通过。 本项目属于跨系统多模块架构。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：客户端 history 命令不传 --start/--end 参数时默认用 time.Now() 生成时间范围，RFC3339 格式包含 +08:00 时区偏移，拼接到 URL query string 时 + 号没有被 URL encode，服务端解析 start 参数返回 400，导致 history 命令直接报错。传了显式时间参数（如 Z 结尾的 UTC 时间）则正常。过程不满意：Go 的 url.Values 或 net/url.QueryEscape 是处理 query 参数的标准做法，写完后没有测试客户端不传参数时的默认行为 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 448-go-realtime-dashboard |

---

## 448-go-realtime-dashboard — 第二轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 我试了下客户端的 history 命令，直接 ./solo-client history cpu_usage 不带任何时间参数就报错了，返回 400。但加上 --start 和 --end 就正常。应该是默认时间拼 URL 的时候 +08:00 没处理好吧 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 已完成任务 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 |  |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 448-go-realtime-dashboard |

### R2 评估详情

**Bug 修复分析**：
- R1 bug 根因：客户端 `GetMetricHistory` 使用 `fmt.Sprintf` 手动拼接 URL，`time.Now().Format(time.RFC3339)` 产生的 `+08:00` 时区偏移中 `+` 号未被 URL encode，服务端将 `+` 解析为空格导致时间解析失败返回 400
- R2 修复方式：改用 `url.Values{}` 构建 query 参数，调用 `.Encode()` 自动处理 `+` → `%2B` 的编码。当不传时间参数时 `startTime`/`endTime` 为 nil，不添加 query 参数，服务端使用默认值（最近24小时到现在）
- 修复范围：`client/api_client.go` 的 `GetMetricHistory` 方法，改动精准且最小化

**测试结果**：
| 测试项 | 结果 |
|--------|------|
| history 不带时间参数（R1 bug） | ✅ 正常返回数据 |
| history 带 UTC 时间参数 | ✅ 正常 |
| history 带 +08:00 时区参数 | ✅ 正常 |
| create-metric | ✅ |
| update-metric | ✅ |
| get-metric | ✅ |
| metrics 列表 | ✅ |
| set-threshold | ✅ |
| 告警触发（超阈值推送） | ✅ alerts 返回告警记录 |
| trend 趋势 | ✅ |
| comparison 同比环比 | ✅ |
| delay-stats 延迟统计 | ✅ |
| save-layout / get-layout | ✅（需同一 client-id） |
| status / status --details | ✅ |
| cleanup 清理 | ✅ |
| 重复创建指标（409） | ✅ |
| 不存在的指标（404） | ✅ |
| go build ./... | ✅ 编译通过 |

**代码质量**：修复干净利落，使用标准库 `net/url.Values` 处理 query 参数编码，是 Go 的最佳实践。未引入新问题。

---

