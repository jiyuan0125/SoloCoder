# Solo Coder 填表数据

## 419-java-reimbursement-split — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 财务要求报销可以拆分到多个成本中心。做一个报销拆分API，一笔报销指定多个成本中心各分配百分比，所有百分比加起来必须正好100%，允许的最小分配单位是1%。比如差旅6000元，60%归属研发中心、40%归属市场中心。每个成本中心负责人都要审批通过。任一负责人驳回整笔回到待修改状态让员工重新调整，已经通过的审批全部清零重新走。审批过程中员工不能修改已提交的分配比例，必须等驳回后才能改。查询能看到每笔报销的分配明细和各中心审批状态。同一笔报销拆分的成本中心数量不超过5个，超过的话需要走特殊审批流程。已通过全部审批的报销单金额会自动汇总到各成本中心的月度支出报表中。 Maven多模块（common/server/client），Spring Boot server 内存存储，纯Java CLI client HTTP调用。Java 17，禁止Lombok。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：maven-compiler-plugin缺少-parameters配置，Spring Boot 3.x下@PathVariable未显式指定name导致submit/getDetail/monthlyReport/budgetWarning共5个端点全部500错误。驳回时未清零已通过审批——CC001已通过后CC002驳回，CC001仍显示APPROVED，需求要求"已通过的全部清零重新走"。特殊审批无实际流程——specialApprovalGranted直接取specialApprovalRequested的值，标记true即自动批准。过程不满意：手写JSON解析器覆盖不全且脆弱，核心审批流程的驳回清零逻辑与PROMPT矛盾 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 419-java-reimbursement-split |

---
