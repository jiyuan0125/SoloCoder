# Solo Coder 填表数据

## 418-java-work-shift — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 工厂排班需要后端服务。做一个排班API，班次类型有早班（8-16点）、中班（16-24点）、夜班（0-8点）。给员工排班，每个记录包含员工ID、日期、班次。同一个员工不能连续排夜班超过3天。员工可以提交换班申请指定目标同事，对方确认后生效。换班生效后重新校验双方是否满足排班规则，如果导致任一方连续4天夜班就拒绝换班。相邻两天班次之间至少8小时间隔，夜班接早班属于违规。每周每人至少安排1天休息日，连续工作天数不超过6天。法定节假日排班按双倍工时计算。排班发布后员工可以在24小时内提出异议，超过24小时视为认可。历史排班数据支持按员工和日期范围查询，生成月度排班统计表。 Maven多模块（common/server/client），Spring Boot server 内存存储，纯Java CLI client HTTP调用。Java 17，禁止Lombok。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：maven-compiler-plugin缺少-parameters配置导致几乎所有GET/PUT/DELETE端点500。calculateInterval对AFTERNOON(16-24)→NIGHT(0-8)场景返回24h而非0h，漏检0小时间隔违规。simulateSwap将日期设为otherShift.getDate()而非currentShift.getDate()，换班规则校验可能不准确。过程不满意：编译配置遗漏加上间隔计算和换班模拟两处逻辑错误，说明排班规则核心逻辑未经实际测试 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 418-java-work-shift |

---
