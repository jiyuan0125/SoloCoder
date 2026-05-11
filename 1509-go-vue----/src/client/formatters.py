from typing import List, Dict, Any


STAGE_NAMES = {
    "raw_material": "原料采购",
    "production": "生产加工",
    "processing": "深加工",
    "packaging": "包装",
    "warehouse": "仓储",
    "distribution": "配送",
    "retail": "零售上架"
}

INSPECTION_STATUS_NAMES = {
    "pending": "待检测",
    "passed": "合格",
    "failed": "不合格"
}

RECALL_STATUS_NAMES = {
    "active": "进行中",
    "completed": "已完成",
    "cancelled": "已取消"
}

TODO_STATUS_NAMES = {
    "pending": "待处理",
    "confirmed": "已确认",
    "overdue": "已逾期"
}


def format_supply_chain(record: Dict[str, Any]) -> str:
    lines = []
    lines.append(f"ID: {record.get('id', 'N/A')}")
    lines.append(f"批次: {record.get('batch_number', 'N/A')}")
    lines.append(f"环节: {STAGE_NAMES.get(record.get('stage'), record.get('stage', 'N/A'))}")
    lines.append(f"时间: {record.get('operation_time', 'N/A')}")
    lines.append(f"操作人: {record.get('operator', 'N/A')}")
    if record.get('location'):
        lines.append(f"地点: {record['location']}")
    if record.get('remarks'):
        lines.append(f"备注: {record['remarks']}")
    return "\n".join(lines)


def format_inspection(record: Dict[str, Any]) -> str:
    lines = []
    lines.append(f"ID: {record.get('id', 'N/A')}")
    lines.append(f"批次: {record.get('batch_number', 'N/A')}")
    lines.append(f"时间: {record.get('inspection_time', 'N/A')}")
    lines.append(f"检测人: {record.get('inspector', 'N/A')}")
    status = record.get('status')
    lines.append(f"结果: {INSPECTION_STATUS_NAMES.get(status, status)}")
    if record.get('items'):
        lines.append(f"项目: {', '.join(record['items'])}")
    if record.get('report'):
        lines.append(f"报告: {record['report']}")
    if record.get('remarks'):
        lines.append(f"备注: {record['remarks']}")
    return "\n".join(lines)


def format_recall(recall: Dict[str, Any]) -> str:
    lines = []
    lines.append(f"ID: {recall.get('id', 'N/A')}")
    lines.append(f"批次: {recall.get('batch_number', 'N/A')}")
    status = recall.get('status')
    lines.append(f"状态: {RECALL_STATUS_NAMES.get(status, status)}")
    lines.append(f"可追踪: {'是' if recall.get('can_track') else '否'}")
    lines.append(f"创建时间: {recall.get('created_at', 'N/A')}")
    lines.append(f"召回原因: {recall.get('reason', 'N/A')}")
    if recall.get('tracked_stages'):
        stage_names = [STAGE_NAMES.get(s, s) for s in recall['tracked_stages']]
        lines.append(f"追踪环节: {', '.join(stage_names)}")
    if recall.get('completed_at'):
        lines.append(f"完成时间: {recall['completed_at']}")
        lines.append(f"完成人: {recall.get('completed_by', 'N/A')}")
    return "\n".join(lines)


def format_todo(todo: Dict[str, Any]) -> str:
    lines = []
    lines.append(f"ID: {todo.get('id', 'N/A')}")
    lines.append(f"召回ID: {todo.get('recall_id', 'N/A')}")
    lines.append(f"批次: {todo.get('batch_number', 'N/A')}")
    stage = todo.get('stage')
    lines.append(f"环节: {STAGE_NAMES.get(stage, stage)}")
    lines.append(f"负责人: {todo.get('operator', 'N/A')}")
    status = todo.get('status')
    status_name = TODO_STATUS_NAMES.get(status, status)
    if status == 'overdue':
        status_name = f"[!] {status_name}"
    lines.append(f"状态: {status_name}")
    lines.append(f"创建时间: {todo.get('created_at', 'N/A')}")
    if todo.get('confirmed_at'):
        lines.append(f"确认时间: {todo['confirmed_at']}")
        lines.append(f"确认人: {todo.get('confirmed_by', 'N/A')}")
    if todo.get('remarks'):
        lines.append(f"备注: {todo['remarks']}")
    return "\n".join(lines)


def format_timeline(records: List[Dict[str, Any]]) -> str:
    if not records:
        return "  暂无记录"
    
    lines = []
    for i, record in enumerate(records, 1):
        stage = record.get('stage')
        stage_name = STAGE_NAMES.get(stage, stage)
        lines.append(f"\n[{i}] {stage_name}")
        lines.append(f"    时间: {record.get('operation_time', 'N/A')}")
        lines.append(f"    操作人: {record.get('operator', 'N/A')}")
        if record.get('location'):
            lines.append(f"    地点: {record['location']}")
        if record.get('remarks'):
            lines.append(f"    备注: {record['remarks']}")
    return "\n".join(lines)
