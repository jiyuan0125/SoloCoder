# Solo Coder 填表数据

## 417-java-employee-directory — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 员工通讯录需要集中管理。做一个通讯录API，员工信息包含姓名、工号、部门、职位、手机号、邮箱、办公地点。按姓名、部门、职位模糊搜索，搜索结果按相关性排序（完全匹配优先于模糊匹配）。员工本人可以更新自己的手机号和邮箱，其他信息只有HR能改。离职员工记录保留但标记为离职，默认搜索不显示离职人员，管理员可以切换查看含离职人员的完整列表。同一个手机号不能同时绑定两个在职员工。邮箱域名必须是公司域名，不允许填个人邮箱。提供部门人数统计和组织架构树形视图数据。通讯录变更记录操作日志。批量导入员工时遇到重复工号跳过并记录。 Maven多模块（common/server/client），Spring Boot server 内存存储，纯Java CLI client HTTP调用。Java 17，禁止Lombok。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：maven-compiler-plugin缺少-parameters编译参数，Spring Boot 3.x下@PathVariable/@RequestParam未显式指定name，导致GET列表、查看详情、更新员工、查看员工日志共4个接口全部返回500错误，约一半API端点不可用。过程不满意：Spring Boot 3.x下参数名解析是基本要求，构建后未实际测试带路径参数的接口 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 417-java-employee-directory |

---

## 417-java-employee-directory — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 启动之后试了几个接口，创建员工没问题，但只要URL里带参数的都500了。比如 GET /api/employees/E001 查详情、PUT /api/employees/E002 更新信息、GET /api/employees?includeResigned=true 查含离职人员的列表，全部报 parameter name 找不到。看了下好像 Spring Boot 3 + Java 17 需要编译器加个 -parameters 参数，@PathVariable 和 @RequestParam 才能正常工作。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：createEmployee()和batchImport()方法没有调用operationLogService.logChange()，创建和批量导入员工时不生成操作日志，PROMPT明确要求"通讯录变更记录操作日志——谁在什么时间改了谁的什么字段"。EmployeeClient.runWithArgs()方法仍只打印"命令行模式暂未实现完整功能"，PROMPT要求client是"纯Java的命令行客户端"。HttpClientWrapper使用.header("X-User-Name", currentUserName)直接设置中文header值，Java HttpClient默认用ISO-8859-1编码HTTP header，中文字符会被损坏。过程不满意：本轮只修复了-parameters编译参数这一个R1问题，其余3个R1 bug均未处理，说明没有全面对照R1评估结果逐项修复 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 417-java-employee-directory |

---
