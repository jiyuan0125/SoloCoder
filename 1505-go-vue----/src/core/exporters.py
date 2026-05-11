from typing import List

from .models import BatchSummary


def format_batch_summaries_as_table(summaries: List[BatchSummary]) -> str:
    if not summaries:
        return "没有找到匹配的批次记录。"
    
    headers = [
        "批次ID",
        "配方名称",
        "产品类别",
        "计划数量",
        "实际数量",
        "生产日期",
        "质检结果",
        "待办状态",
    ]
    
    rows = []
    for s in summaries:
        actual_qty = f"{s.actual_quantity:.2f}" if s.actual_quantity is not None else "-"
        prod_date = s.production_date.strftime("%Y-%m-%d %H:%M:%S") if s.production_date else "-"
        inspection = s.inspection_result or "-"
        todo = s.todo_status or "-"
        
        rows.append([
            str(s.batch_id),
            s.recipe_name,
            s.product_category,
            f"{s.planned_quantity:.2f}",
            actual_qty,
            prod_date,
            inspection,
            todo,
        ])
    
    col_widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            if len(cell) > col_widths[i]:
                col_widths[i] = len(cell)
    
    def format_row(cells: List[str]) -> str:
        padded = [cells[i].ljust(col_widths[i]) for i in range(len(cells))]
        return " | ".join(padded)
    
    separator = "-+-".join("-" * w for w in col_widths)
    
    lines = []
    lines.append(format_row(headers))
    lines.append(separator)
    for row in rows:
        lines.append(format_row(row))
    
    return "\n".join(lines)
