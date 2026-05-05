# Solo Coder 填表数据

## 411-java-leave-approver — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 1 |
| User Prompt | 公司请假全靠纸质单太慢。做一个请假审批API，员工填请假类型（年假/病假/事假/婚假/产假/陪产假）、日期范围和事由。系统自动校验剩余假期天数够不够。年假根据工龄定：1到3年5天、3到10年10天、10年以上15天，每年1月1日按最新工龄重新计算。上级可以批准或驳回。上一年剩的年假可以延续到次年3月31日，过期自动清零。病假必须填附件名（病假条或就诊证明照片），不填直接拒绝。婚假和产假也要提供证明附件名。同一时间段不能同时请两种假，比如年假和事假不能重叠。请假期间如果遇到法定节假日，节假日不计入请假天数。请假跨周末的，周末是否计入要看请假类型——病假和事假计入，年假不计入。查询团队成员请假日历方便交接，日历视图显示每人每天的请假状态。 每年1月1日按最新工龄重新计算年假天数。同一时间段不能同时请两种假。请假期间遇到法定节假日不计入请假天数。请假跨周末的，病假和事假计入周末天数，年假不计入。婚假按法定天数给：3天，晚婚加10天。产假按国家规定执行，陪产假15天。请假审批记录永久保存供年度考勤汇总。日期时间使用 java.time 包，不用 Date 或 Calendar。字符串比较用 equals 不用 ==。 做成 Maven 多模块项目。父 POM 管理公共依赖和版本号。三个子模块：common 模块定义通信协议的 DTO 类、请求响应格式和错误码枚举；server 模块是 Spring Boot 应用，监听端口提供 REST 接口，处理所有业务请求，数据保存在内存中；client 模块是命令行工具（纯 Java，不用 Spring），通过 HTTP 调用 server 的接口，结果格式化输出到终端。各模块内部按职责分包。不使用 Lombok。Java 17，mvn clean package 能构建成功。 本项目属于跨系统多模块架构。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：5个严重Bug——(1)@PathVariable导致所有带路径参数的GET接口返回400（父POM未使用spring-boot-starter-parent，maven-compiler-plugin缺少-parameters编译参数，Spring Boot 3.x无法解析参数名）；(2)LeaveType.ANNUAL的countWeekends=true应为false，PROMPT明确说"年假不计入周末"，实测5/8-5/11申请年假返回4天应为2天；(3)LeaveType.PERSONAL的requiresAttachment=true应为false，PROMPT只要求病假/婚假/产假要附件，实测申请事假不传附件被拒绝；(4)CLI客户端JAR缺少依赖无法运行（maven-jar-plugin打thin JAR，java -jar报NoClassDefFoundError）；(5)审批接口managerId为null时NPE（LeaveService.approveLeave第130行managerId.equals()空指针）。3个中等问题——EmployeeDTO.yearsOfService始终为0（toDTO从未赋值）、年假自动重算和结转过期的两个定时任务方法存在但从未调度（无@EnableScheduling和@Scheduled）。过程不满意：LeaveType枚举中ANNUAL的countWeekends和PERSONAL的requiresAttachment两个布尔值与需求完全相反，说明实现时未仔细对照需求原文；构建后没有实际运行测试GET接口和客户端JAR |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 411-java-leave-approver |

---

## 411-java-leave-approver — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 2 |
| User Prompt | 我刚跑了一下项目，发现所有带路径参数的 GET 接口全返回 400，报错说参数名找不到，比如 GET /api/employees/1、GET /api/leaves/1 这些全挂了，只有 POST 能用。 CLI 客户端也跑不起来，java -jar 直接 NoClassDefFoundError 找不到 jackson 的类。 然后我测了一下年假，申请 5/8（周五）到 5/11（周一），按需求年假不算周末应该只算 2 天工作日，结果返回了 4 天。 还有事假不需要附件的吧？申请事假不传附件就报错说"该请假类型必须提供附件"。员工信息里的 yearsOfService 也全是 0。 需求里说的"每年 1 月 1 日按工龄重算年假"和"3 月 31 日清零结转年假"这两个定时任务代码里有方法但没人调。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：R1 的 7 个 bug 中修复了 6 个，但审批接口 managerId=null 时的 NPE 仍未修复。LeaveService.approveLeave 第 130 行仍为 `if (!managerId.equals(employee.getManagerId()))`，当 managerId 为 null 时直接抛出 NullPointerException，返回 500。应改为 `if (managerId == null || !managerId.equals(employee.getManagerId()))` 或使用 Objects.equals。过程不满意：R1 已明确指出 NPE 位于第 130 行并给出了具体原因，修复时漏掉了这个 bug |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 411-java-leave-approver |

---
