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

## 418-java-work-shift — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 跑起来发现大部分查询接口都是500，URL带路径参数的全部挂了，应该是 Spring Boot 3 编译参数 -parameters 没加的问题。另外我看了一下排班间隔校验的逻辑，中班（16-24点）接夜班（0-8点）这种排班，中间间隔应该是0个小时，但代码算出来是24个小时，直接漏检了。还有换班模拟那边，simulateSwap 里日期赋值写成了 otherShift.getDate()，应该是当前排班的日期才对。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：PROMPT 要求"排班表发布后自动通知相关员工"，代码中完全没有通知机制的实现。PROMPT 要求"超过24小时视为认可"，ObjectionService 只在24小时内允许提交异议，但缺少超过24小时后自动将排班标记为已认可的逻辑。ShiftService.deleteShift 删除已发布排班时返回 SHIFT_NOT_FOUND 错误码而非语义准确的错误信息。MonthlyStatisticsDTO 中 normalWorkHours 和 holidayWorkHours 使用 double 类型，与 hourlyWage(BigDecimal) 相乘时先转 double 再转 BigDecimal 存在精度丢失风险。过程不满意：R1 反馈了3个技术 bug 和2个缺失功能，只修复了3个技术 bug，2个缺失功能（通知、自动认可）仍未实现 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 418-java-work-shift |

---

