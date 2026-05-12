# 铁路客运调度管理系统

基于 FastAPI 的高速铁路客运调度管理系统，支持列车信息管理、交路维护、乘务排班、晚点处理和交路优化。

## 功能特性

### 1. 列车信息管理
- 动车组信息维护（车型、定员、乘务定额）
- 列车状态管理（空闲、运营中、维护中）

### 2. 交路管理
- 交路计划创建和维护
- **约束规则**：
  - 一个动车组同时只能跑一个交路
  - 交路间至少 30 分钟折返时间

### 3. 乘务管理
- 乘务人员信息管理（司机、列车长、列车员）
- 乘务组管理
- **工时规则**：
  - 司机连续值乘不超过 4 小时，需休息 30 分钟
  - 日工作时间不超过 8 小时
  - 月度工作时间不超过 160 小时

### 4. 晚点处理
- 晚点超过 30 分钟自动变更状态
- **乘务调整方案生成**：
  - 若乘务组当日无后续交路，方案仅包含晚点信息
  - 若有后续交路，生成完整的人员调度方案
- 晚点恢复后自动重新核算月度工时

### 5. 交路优化建议
- 动车组日均运用低于 8 小时生成优化建议
- 折返时间过长生成优化建议

### 6. 工时预警
- 司机值乘记录完成后自动累加到月度工时
- 月度工时超过 160 小时自动推送预警给调度主任

### 7. 命令行客户端
- 查询交路信息
- 查看乘务排班
- 查看预警和优化建议

## 项目结构

```
.
├── main.py                  # FastAPI 入口
├── cli.py                   # 命令行客户端
├── requirements.txt         # 依赖列表
├── init_sample.py           # 示例数据初始化
├── check_syntax.py          # 语法检查脚本
├── railway.db               # SQLite 数据库（自动生成）
└── app/
    ├── __init__.py
    ├── database.py          # 数据库连接配置
    ├── models.py            # SQLAlchemy 数据模型
    ├── schemas/
    │   ├── __init__.py
    │   └── models.py        # Pydantic 数据模型
    ├── services/
    │   ├── __init__.py
    │   ├── train_service.py # 列车和交路服务
    │   ├── crew_service.py  # 乘务和工时规则服务
    │   ├── delay_service.py # 晚点处理服务
    │   └── optimization_service.py  # 优化和预警服务
    └── routers/
        ├── __init__.py
        ├── trains.py        # 列车管理 API
        ├── routes.py        # 交路管理 API
        ├── crews.py         # 乘务组管理 API
        ├── schedules.py     # 排班管理 API
        └── operations.py    # 运营管理 API
```

## 安装与运行

### 1. 安装依赖

```bash
# 创建虚拟环境（推荐）
python3 -m venv venv
source venv/bin/activate

# 安装依赖
pip install -r requirements.txt
```

### 2. 初始化示例数据

```bash
python init_sample.py
```

### 3. 运行 Web API 服务

```bash
uvicorn main:app --reload
```

访问 http://localhost:8000/docs 查看 Swagger API 文档。

### 4. 使用命令行客户端

```bash
# 查看列车列表
python cli.py train list

# 查看交路列表
python cli.py route list
python cli.py route list --today
python cli.py route list --status scheduled

# 查看乘务人员
python cli.py crew list
python cli.py crew list --type driver

# 查看乘务组
python cli.py crew groups

# 查看月度工时统计
python cli.py crew schedule
python cli.py crew schedule --year 2024 --month 5

# 查看预警
python cli.py alert list

# 查看优化建议
python cli.py optimization list
python cli.py optimization detail <opt_id>

# 查看晚点调整方案
python cli.py delay list
python cli.py delay detail <adjustment_id>
```

## API 接口

### 列车管理
- `POST /api/trains/` - 创建列车
- `GET /api/trains/` - 列车列表
- `GET /api/trains/{train_id}` - 列车详情
- `PATCH /api/trains/{train_id}` - 更新列车
- `DELETE /api/trains/{train_id}` - 删除列车

### 交路管理
- `POST /api/routes/` - 创建交路
- `GET /api/routes/` - 交路列表
- `GET /api/routes/{route_id}` - 交路详情
- `POST /api/routes/{route_id}/start` - 开始交路
- `POST /api/routes/{route_id}/complete` - 完成交路
- `POST /api/routes/{route_id}/cancel` - 取消交路

### 乘务管理
- `POST /api/crews/members` - 创建乘务员
- `GET /api/crews/members` - 乘务员列表
- `POST /api/crews/groups` - 创建乘务组
- `GET /api/crews/groups` - 乘务组列表

### 排班管理
- `POST /api/schedules/duty-records` - 创建值乘记录
- `POST /api/schedules/duty-records/{id}/complete` - 完成值乘
- `GET /api/schedules/monthly-report/{year}/{month}` - 月度工时统计

### 运营管理
- `POST /api/operations/routes/{id}/check-delay` - 检查晚点
- `POST /api/operations/routes/{id}/recover-delay` - 恢复晚点
- `GET /api/operations/train-utilization` - 列车利用率
- `POST /api/operations/generate-optimizations` - 生成优化建议
- `GET /api/operations/alerts` - 预警列表
- `POST /api/operations/alerts/{id}/acknowledge` - 确认预警

## 核心业务规则

### 交路约束
1. 同一动车组在同一时间只能分配一个交路
2. 交路间折返时间至少 30 分钟

### 司机工时规则
| 规则 | 限制 |
|------|------|
| 连续值乘 | ≤ 4 小时，之后需休息 30 分钟 |
| 日工作时间 | ≤ 8 小时 |
| 月度工作时间 | ≤ 160 小时 |

### 晚点处理
1. 晚点 ≥ 30 分钟：自动标记为晚点状态
2. 生成晚点调整方案发送给值班调度
3. 若有后续交路：提供完整的人员调度方案
4. 若无后续交路：仅发送晚点信息通知

### 交路优化触发条件
1. 动车组日均运用时间 < 8 小时
2. 平均折返时间 > 120 分钟

### 预警机制
- 司机月度工时超过 160 小时
- 自动发送预警给调度主任
