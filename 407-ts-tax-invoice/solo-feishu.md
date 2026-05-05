# Solo Coder 填表数据

## 407-ts-tax-invoice — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 1 |
| User Prompt | 财务管发票工作量太大。做一个发票管理API，支持录入（购买方名称税号、销售方名称税号、金额不含税、税额、价税合计、专票或普票）、作废和红冲。按税额区间筛选查询。已关联报销单的发票不能直接作废要先解除关联。红冲时红字发票金额和税额必须与原发票完全一致，价税合计自动反算校验。一张发票只能红冲一次，红冲后原发票和红字发票都要保留在记录中形成对应关系。作废和红冲是两种不同操作不能混用：作废是发票还没抵扣时直接作废，红冲是已经抵扣了需要开红字冲回。发票录入时税号格式要校验——统一社会信用代码为18位字母数字，税号和金额必须满足数学关系（金额不含税乘以税率等于税额，两者相加等于价税合计）。支持批量导入发票，单次最多100张，格式错误的跳过并在返回结果中列出失败原因。服务重启后所有发票数据要还在。税号格式校验——统一社会信用代码18位字母数字，税号和金额必须满足数学关系。作废是发票还没抵扣时直接作废，红冲是已经抵扣了需要开红字冲回。支持批量导入发票单次最多100张，格式错误的跳过并在返回中列出失败原因。发票查询结果按开票日期倒序。月度发票汇总自动统计专票和普票各自金额和税额。金额字段全部以整数分存储，税率计算使用整数运算保证精度。所有异步操作中的错误必须被正确捕获和处理，不允许静默吞掉异常。 以 TypeScript monorepo 形式搭建整个系统。packages/shared 承载公共的 TypeScript 类型、接口定义、消息格式和错误码枚举。packages/server 依靠 Node.js 自带的 http 模块搭建 HTTP 服务（拒绝使用任何第三方 Web 框架如 Express），处理全部业务逻辑，数据存于内存中。packages/cli 为命令行工具，通过子命令与参数向服务端发送请求，并将返回结果格式化展示在终端。项目根 package.json 通过 workspaces 关联三个子包，顶层 tsconfig.json 强制 strict 模式。各子包按职责将代码分散到多个文件。全程禁止 any 和 @ts-ignore。npm install && npm run build 能正常编译。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：router.ts 中 GET /api/invoices/monthly-summary 路由被排在它前面的正则路由 /^\/api\/invoices\/[^\/]+$/ 抢先匹配，"monthly-summary" 被当作发票 ID 查询，返回 404，月度汇总功能完全不可用。validateAmountRelation 使用浮点除法 taxAmount / amountExcludingTax 计算税率再 Math.round 回算，PROMPT 明确要求"税率计算使用整数运算保证精度"。过程不满意：路由定义顺序导致关键接口不可访问，写完后没有逐个接口测试验证 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 407-ts-tax-invoice |

---

## 407-ts-tax-invoice — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 2 |
| User Prompt | 我启动了服务跑了一下，其他接口都正常，但月度汇总 GET /api/invoices/monthly-summary 返回 404，错误信息是"发票不存在"。我看了一下 router.ts 的路由定义，monthly-summary 那个路由排在正则路由 `^\/api\/invoices\/[^\/]+$` 后面，这个正则把 "monthly-summary" 当成了发票 ID 匹配进去了，根本走不到 monthlySummaryHandler。另外 validateAmountRelation 里用 `taxAmount / amountExcludingTax` 算税率，这是浮点除法，PROMPT 要求"税率计算使用整数运算保证精度"。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 模块内多文件 |
| 任务是否完成 | 完成了任务 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 | |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 407-ts-tax-invoice |

---
