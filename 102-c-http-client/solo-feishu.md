# Solo Check Report — 102-c-http-client

## 基本信息

| 字段 | 值 |
|------|-----|
| 项目 | 102-c-http-client |
| 技术栈 | C |
| 轮次 | R1 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 库/SDK |
| 修改范围 | 跨模块多文件 |

## 评测结论

| 字段 | 值 |
|------|-----|
| 任务是否完成 | **未完成** |
| 产物是否满意 | **不满意** |

## 编译与运行

- `make clean && make`：**通过**，gcc -Wall -Wextra 零警告
- `./http_alert_demo`：**运行成功**，POST 到 httpbin.org/post 返回 200，JSON body 正确
- `make ssl`：因环境缺少 libssl-dev 未能编译，但代码结构正确（`#ifdef WITH_SSL` 守卫）

## 需求逐项检查

| # | 需求 | 状态 | 说明 |
|---|------|------|------|
| 1 | 支持 GET、POST、PUT | ✅ | HttpMethod 枚举定义完整，API 层支持三种方法 |
| 2 | POST/PUT 能发 JSON body | ✅ | `http_alert_set_body` 支持任意 body，demo 构造了正确 JSON |
| 3 | 自定义 HTTP 头部 | ✅ | `http_alert_add_header` 可添加任意头部，demo 中添加了 Content-Type 和 Accept |
| 4 | 连接超时和读取超时分开设 | ✅ | TimeoutConfig 有 connect_timeout 和 read_timeout 两个独立字段 |
| 5 | 自动重试最多 3 次，间隔翻倍（1s,2s,4s） | ✅ | retry_scheduler 正确实现，实测 3 次重试间隔 1s→2s→4s |
| 6 | 整体超时上限（如 10 秒） | ⚠️ | 机制存在但不精确：仅在循环开头检查，sleep 期间不中断，实际耗时可能超出配置值一个完整尝试周期+延迟 |
| 7 | 支持 HTTPS，不验证证书 | ✅ | WITH_SSL 编译选项，SSL_VERIFY_NONE，未编译 SSL 时给出明确错误提示 |
| 8 | 拿到状态码和响应体 | ✅ | HttpResponse 包含 status_code、status_message、headers、body、body_len |
| 9 | 纯 C，不依赖第三方库，gcc Linux 编译通过 | ✅ | 无任何第三方依赖，零警告编译通过 |
| 10 | main 演示：POST 到 HTTP 测试服务，模拟告警 JSON | ✅ | main.c 完整实现，命令行参数支持，成功发送到 httpbin.org |
| 11 | 至少 4 个源文件按功能模块分 | ✅ | 4 个模块：http_alert、http_connection、http_request、retry_scheduler + main.c |
| 12 | 头文件和实现文件分开 | ✅ | 所有 4 个模块均有 .h/.c 文件对 |

## 发现的问题

### Bug 1（严重）：push_alert_json 对非 2xx 不可重试状态码返回成功

**文件**：`retry_scheduler.c` — `retry_scheduler_execute` + `push_alert_with_retry`

**现象**：当 HTTP 传输成功但服务器返回 4xx（非 429）状态码时，函数返回 `HTTP_ALERT_OK(0)`，调用方无法区分成功与失败。

**根因**：
```c
// retry_scheduler_execute 中：
result.last_error = ret;  // ret = HTTP_ALERT_OK（传输成功）

if (ret == HTTP_ALERT_OK) {
    if (is_success_response(result.response)) {
        result.success = 1;
        break;
    }
    // 非 2xx 且非 5xx/429 时，last_error 仍为 HTTP_ALERT_OK
}

// push_alert_with_retry 中：
return result.success ? HTTP_ALERT_OK : result.last_error;
// result.success = 0，但 result.last_error = HTTP_ALERT_OK → 返回 0
```

**复现**：`./http_alert_demo -u http://httpbin.org/get` → 服务器返回 405，程序打印 "Alert pushed successfully!"

**修复建议**：在 `retry_scheduler_execute` 中，当 HTTP 响应成功但状态码非 2xx 且不应重试时，设置 `result.last_error` 为一个新的错误码（如 `HTTP_ALERT_ERR_HTTP_ERROR`），或在返回值中携带 HTTP 状态码。

### Bug 2（轻微）：整体超时检查时机不精确

**文件**：`retry_scheduler.c` — `retry_scheduler_execute`

**现象**：配置 `overall_timeout=8s`，实际耗时 12 秒才触发超时退出。

**根因**：整体超时在循环开头检查，但 sleep（重试延迟）在检查之后执行。当检查通过后执行 sleep + 一次完整的连接尝试，实际耗时可能大幅超出配置值。

**修复建议**：在 sleep 之前再次检查整体超时剩余时间，将 sleep 时间限制为剩余时间与计划延迟的较小值。

### 问题 3（设计）：http_connection_send 用 read_timeout 控制 write 等待

**文件**：`http_connection.c` — `http_connection_send`

**现象**：非 SSL 的 send 路径使用 `conn->timeout.read_timeout_sec` 作为写等待超时。

**说明**：PROMPT 只要求连接超时和读取超时分开，没有要求发送超时。从实际效果看用 read_timeout 控制 send 等待是合理的，但语义上有些奇怪。

### 问题 4（轻微）：http_alert_parse_url 部分分配失败时内存泄漏

**文件**：`http_alert.c` — `http_alert_parse_url`

**说明**：如果在分配 host 后 path 分配失败，之前已分配的 url 和 host 不会被释放。实际中这些分配极小不太会失败，影响不大。

## 代码质量

| 维度 | 评分 | 说明 |
|------|------|------|
| 模块划分 | 优秀 | 4 个模块职责清晰：连接管理、请求构建、重试调度、核心类型 |
| 代码风格 | 良好 | 命名一致，缩进统一，注释适当 |
| 错误处理 | 良好 | 完善的错误码体系，大部分路径有正确的资源清理 |
| 内存管理 | 良好 | 整体良好，有少量错误路径泄漏 |
| 可测试性 | 良好 | 模块化设计便于单独测试各组件 |

## 不满意原因

**产物不满意**：`push_alert_json` 在 HTTP 传输成功但状态码为非 2xx（如 405、400、403）时，返回值仍为 `HTTP_ALERT_OK(0)`，调用方（main.c）据此打印"Alert pushed successfully!"，实际推送并未成功。这是告警推送模块的核心逻辑缺陷——运维系统会误判告警已送达。

**过程不满意**：核心的返回值判定逻辑有误，写完后仅用 httpbin.org/post（返回 200）做验证，没有用非 2xx 响应测试推送结果的判定逻辑。
