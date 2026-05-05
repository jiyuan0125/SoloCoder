# Solo Coder 填表数据

## 416-java-performance-review — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 绩效考核需要线上化。做一个绩效评估API，每季度HR发起评估，流程是员工自评（1到5分加评语）→上级评分（1到5分加评语）→HR确认。上级评分和员工自评分差异不能恰好为0，至少0.5分调整，防止上级偷懒直接抄员工自评分。评估完成后员工看到最终得分和部门内排名百分位。员工季度中间调岗了，绩效由新部门上级来评，原部门评语保留在记录中。绩效等级根据最终得分划分：4.5以上S级、3.5到4.5A级、2.5到3.5B级、1.5到2.5C级、1.5以下D级。连续两个季度C级进入绩效改进计划，连续三个季度C级或任一季度D级启动辞退流程。绩效得分影响年终奖系数：S级1.5、A级1.2、B级1.0、C级0.5、D级0。历史评估按部门和季度筛选查询。 BigDecimal处理金额，Optional不能直接.get()。 Maven多模块（common/server/client），Spring Boot server 内存存储，纯Java CLI client HTTP调用。Java 17，禁止Lombok。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：maven-compiler-plugin缺少-parameters配置，Spring Boot 3.x下@PathVariable未显式指定name导致路径变量接口返回500错误（需手动修复编译配置后才可正常工作）。上级评分差值校验仅检查≠0未检查≥0.5，PROMPT要求至少0.5分调整。调岗后无业务联动更新departmentId/managerId。过程不满意：编译配置遗漏导致原始交付物核心接口不可用，说明构建后未实际测试启动 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 416-java-performance-review |

---

## 416-java-performance-review — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 项目能编译通过但启动后很多接口返回500。我试了下，所有URL里带ID的接口（比如 GET /api/performance-reviews/{id}、GET /api/review-cycles/{id}）全部报错，日志里说 parameter name 找不到。好像是 Spring Boot 3 和 Java 17 的编译参数问题，@PathVariable 不加 name 属性的话需要 maven 编译器加个配置才行。还有上级评分的时候，我自评打了4分，上级也打了4分，系统居然接受了，不是说上下级评分差值不能为0吗？ |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：员工调岗（PUT /api/employees/{id}/transfer）只更新了Employee实体的departmentId和managerId，未同步更新已创建的PerformanceReview记录中的departmentId和managerId，导致调岗后绩效评估仍关联旧部门旧上级。连续C级计算countConsecutiveCGrads方法存在重复计数bug：当currentGrade为C时先初始化count=1，然后遍历所有review（含当前正在确认的review）又会将C级再次+1，导致连续C数多算1次。过程不满意：核心业务逻辑bug未全部修复，调岗联动是PROMPT明确要求的功能，修复后应实际测试调岗场景 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 416-java-performance-review |

---

