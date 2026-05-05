# Solo Coder 填表数据

## 413-java-onboarding-checklist — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 1 |
| User Prompt | 新员工入职事项多，经常漏掉。做一个入职清单API，创建新员工时自动生成待办（开邮箱、领电脑、签合同、安全培训、领门禁卡、分工位等），每项有负责人和截止日期。员工可以勾选已完成项，负责人也能加新事项或改截止日期。有些事项要在入职前完成（提前准备设备和工位），入职前3天提醒负责人检查。某个负责人名下逾期未完成事项累计超过5个时生成告警记录通知HR。入职当天必须完成的事项（签合同、安全培训）如果当天没完成，第二天自动升级到部门经理关注。所有事项完成率低于80%时该员工入职状态标记为不完整，HR需要跟进处理。支持为不同岗位配置不同的入职清单模板，比如开发岗需要额外配开发环境和VPN，销售岗需要配手机和CRM账号。提供入职进度统计。 入职当天必须完成的事项（签合同、安全培训）如果当天没完成，第二天自动升级到部门经理关注。所有事项完成率低于80%时该员工入职状态标记为不完整HR需跟进。支持为不同岗位配置不同的入职清单模板，比如开发岗需要额外配开发环境和VPN，销售岗需要配手机和CRM账号。清单模板支持复制修改快速创建新模板。日期时间使用 java.time 包，不用 Date 或 Calendar。DTO 类手写 getter/setter，不用 Lombok。 Maven 多模块项目，父 POM 统一管控依赖版本。common 子模块放置 DTO 类、请求响应格式定义和错误码枚举。server 子模块使用 Spring Boot 框架，启动监听端口，提供 RESTful 接口处理全部业务逻辑，数据存于内存。client 子模块为纯 Java 命令行工具，通过 HTTP 请求与 server 通信，格式化输出到控制台。模块内部按功能分包。不允许使用 Lombok。Java 17，执行 mvn clean package 可成功构建。 本项目属于跨系统多模块架构。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：项目无法编译通过，mvn clean package报错。common模块使用了jakarta.validation.constraints的@NotBlank/@NotNull注解（CreateEmployeeRequest、CreateChecklistItemRequest、CreateTemplateRequest），但common/pom.xml仅声明了jackson-databind和jackson-datatype-jsr310，未声明jakarta.validation-api依赖，导致编译失败（14个编译错误），server和client均无法构建。代码架构设计完整（36个Java文件覆盖全部业务需求：员工CRUD、清单模板管理、定时任务调度、告警系统、统计接口、命令行客户端），但依赖管理的低级错误阻塞了整个项目。过程不满意：编译失败是最基础的验证，执行mvn clean package一次即可发现，交付前未做最基本的构建验证 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 413-java-onboarding-checklist |

---

## 413-java-onboarding-checklist — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 2 |
| User Prompt | 我拉下来跑了下mvn clean package，common模块直接编译失败了，报了一堆jakarta.validation.constraints包找不到的错误。CreateEmployeeRequest和CreateChecklistItemRequest里面用了@NotBlank和@NotNull注解，但common的pom.xml里没有加jakarta.validation-api这个依赖，server模块虽然通过spring-boot-starter-validation间接带了，但common模块编译的时候根本拿不到。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：R1的jakarta.validation-api编译错误已修复（common/pom.xml已添加依赖），mvn clean package构建成功。但存在新的严重bug——maven-compiler-plugin缺少-parameters编译参数，导致所有使用@PathVariable的接口返回500错误（"Name for argument of type [java.lang.String] not specified, and parameter name information not found in class file either"），包括：GET/DELETE员工详情、添加/更新清单事项、GET/DELETE模板详情、复制模板、告警CRUD等，几乎全部非列表接口不可用。只有列表类接口（GET /api/employees、GET /api/templates、GET /api/alerts、GET /api/statistics）和POST /api/employees能正常工作。过程不满意：修复了编译问题但未做基本的运行时接口测试，否则一个curl调用即可发现所有@PathVariable接口全部500的问题 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 413-java-onboarding-checklist |

---
