# Solo Coder 填表数据

## 414-java-training-platform — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 1 |
| User Prompt | 公司内部培训需要管理平台。做一个培训平台API，管理员创建课程（名称、描述、讲师、时长、名额上限）。员工查看列表并报名，人数满了就报不了。课程结束后讲师给学员打分（1到100），低于60分算不合格需要重修。课程开始前1小时停止报名。课程开始后已报名的不能再取消，除非提前24小时申请并获得讲师同意。课程结束超过7天但讲师还没给某些学员打分，每天自动提醒讲师直到全部评完。必修课按部门分配，相关员工必须报名且必须通过才能计入年度培训学分。选修课自愿报名，年度选修课学分满20分可获培训优秀证书。培训学分每年1月1日清零重新计算。提供课程列表、报名统计和成绩查询，支持按部门和时间段筛选。课程和报名 低于60分算不合格需要重修。课程开始后已报名的不能再取消，除非提前24小时申请并获得讲师同意。必修课按部门分配，相关员工必须报名且必须通过才能计入年度培训学分。选修课自愿报名，年度选修课学分满20分可获培训优秀证书。培训学分每年1月1日清零重新计算。讲师可以查看自己所有课程的学员评价。金额用 BigDecimal 处理，不能用 double 或 float。DTO 类手写 getter/setter，不用 Lombok。 项目采用 Maven 多模块布局。parent POM 负责依赖管理和版本控制。三个子模块如下：common 模块封装通信用的 DTO、请求/响应格式及错误码定义；server 模块是 Spring Boot Web 应用，监听端口提供 REST API，处理全部业务逻辑，运行时数据保存在内存中；client 模块为纯 Java 命令行程序（不依赖 Spring），通过 HTTP 调用 server 接口并把结果格式化到终端。各模块内按职责分包组织代码。禁止使用 Lombok。目标 Java 17，mvn clean package 构建成功。 本项目属于跨系统多模块架构。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：@PathVariable和@RequestParam注解缺少name属性，Maven编译器未启用-parameters参数，导致所有带路径参数和查询参数的REST端点运行时报错（GET /{id}、PUT /{id}、DELETE /{id}、带@RequestParam的查询全部500），只有无参的POST和GET列表接口能正常工作。server模块的spring-boot-maven-plugin未正确配置repackage目标，mvn clean package产出的JAR不是可执行的fat jar。课程结束超过7天自动提醒讲师未实现（无定时任务）。不及格学员重修机制未实现（报名时仅排除CANCELLED状态，FAILED状态学员无法重新报名）。过程不满意：-parameters和spring-boot repackage是Spring Boot项目的标准配置，缺少这些导致大部分接口无法使用，说明没有端到端实际运行测试 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 414-java-training-platform |

---
