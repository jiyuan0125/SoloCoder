# 1549 第二轮修复指令

## Bug 描述

进出港登记接口 `POST /api/ports/{port_id}/records` 返回 500 错误。

## 根因

`server/routes/ports.py` 第 120 行 `create_port_record` 函数中：

```python
record = PortRecord(port_id=port_id, **record_data.dict())
```

`record_data`（类型 `PortRecordCreate`）的 schema 中包含 `port_id` 字段（required），而 path 参数也传入了同名 `port_id`。展开 `.dict()` 时导致 `PortRecord()` 构造函数收到两个 `port_id` 关键字参数，触发 `TypeError: PortRecord() got multiple values for keyword argument 'port_id'`。

如果不传 `port_id` 给 body，Pydantic 校验报 `Field required`。

## 修复方案

有两种方案（任选其一）：

**方案 A（推荐）**：从 `PortRecordCreate` schema 中移除 `port_id` 字段（它应该只从 path 参数获取），然后构造时正常展开：

```python
record = PortRecord(port_id=port_id, **record_data.dict())
```

**方案 B**：在展开 `.dict()` 时排除 `port_id`：

```python
data = record_data.dict(exclude={"port_id"})
record = PortRecord(port_id=port_id, **data)
```

## 要求

1. 只修复上述 bug，不要改动其他代码
2. 修复后确认 `POST /api/ports/{port_id}/records` 能正常创建进出港登记记录
3. 不要修改端口号（仍使用环境变量 PORT）
4. 不要添加新功能
