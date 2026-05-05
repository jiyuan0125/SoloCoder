# Solo Coder 填表数据

## 420-java-payroll-archive — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 薪资发放记录是敏感财务数据，需要长期保存且防篡改。做一个薪资归档API，每月结算后生成归档记录（员工ID、基本工资、扣款明细、实发金额等）。归档记录一旦生成不能修改或删除，只能追加备注。每条记录附带校验码——所有数值字段按固定顺序拼接后算哈希值。查询时同时返回校验码供核验。如果有人直接改了存储文件的数据，下次查询时重新算的校验码和存的不匹配，自动标记为数据异常。归档数据按年度归组，跨年度查询要分别从对应年度组中检索。支持生成年度薪资汇总报表——包含总支出、人均薪资、最高最低薪资、同比变化率。导出报表时敏感字段（员工姓名、身份证后四位）要脱敏处理。按员工和月份筛选查询。 Maven多模块（common/server/client），Spring Boot server 内存存储，纯Java CLI client HTTP调用。Java 17，禁止Lombok。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：maven-compiler-plugin缺少-parameters编译参数，Spring Boot 3.x无法解析@RequestParam导致所有GET端点（查询/报表/导出/对比）500错误。/export端点未调用MaskUtil脱敏，敏感字段明文输出（MaskUtil已写好但未使用）。无@Scheduled定时任务实现月度自动归档。无归档期间只读锁定机制。过程不满意：MaskUtil已编写却未调用说明脱敏需求被遗漏而非不懂实现，构建后未测试GET接口 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 420-java-payroll-archive |

---

## 420-java-payroll-archive — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 项目编译能过但启动后查询接口全是500，报表、导出、对比这些带请求参数的 GET 接口都挂了，应该是 maven-compiler-plugin 没加 -parameters 导致 Spring Boot 3 解析不了参数名。另外导出报表那个接口 /years/{year}/report/export 返回的数据跟普通报表一模一样，敏感字段（员工姓名、身份证后四位）全是明文没做脱敏，但我看代码里 MaskUtil 类是写好了的，就是没被调用。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：PROMPT 要求"归档操作每月执行一次由系统自动触发，财务确认后正式归档"，代码中无 @Scheduled 定时任务和 @EnableScheduling 注解。PROMPT 要求"归档期间当月薪资数据处于只读状态不允许修改"，代码中无任何锁定机制，confirm 后仍可对同一员工当月数据执行创建操作（会返回"已存在"而非"只读锁定"）。过程不满意：R1 反馈的两个具体 bug（-parameters、导出脱敏）已修复，说明理解能力没问题，但 PROMPT 中两个较复杂的业务需求（定时任务、只读锁定）始终未实现，可能是不熟悉 Spring Scheduling 或并发锁机制 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 420-java-payroll-archive |

---

