import json
import click
from datetime import datetime
from .api import APIClient


def client():
    return APIClient()


def print_json(data):
    click.echo(json.dumps(data, ensure_ascii=False, indent=2, default=str))


@click.group()
def cli():
    pass


@cli.group()
def enterprise():
    pass


@enterprise.command("list")
@click.option("--name", help="按名称筛选")
def enterprise_list(name):
    result = client().get("/enterprises", {"name": name})
    print_json(result)


@enterprise.command("get")
@click.argument("id", type=int)
def enterprise_get(id):
    result = client().get(f"/enterprises/{id}")
    print_json(result)


@enterprise.command("create")
@click.option("--name", required=True, help="企业名称")
@click.option("--address", help="地址")
@click.option("--contact", help="联系人")
@click.option("--phone", help="联系电话")
@click.option("--license", help="许可证号")
@click.option("--user", help="操作人")
def enterprise_create(name, address, contact, phone, license, user):
    data = {"name": name}
    if address:
        data["address"] = address
    if contact:
        data["contact_person"] = contact
    if phone:
        data["contact_phone"] = phone
    if license:
        data["license_number"] = license
    result = client().post("/enterprises", data, {"user": user})
    print_json(result)


@enterprise.command("update")
@click.argument("id", type=int)
@click.option("--name", help="企业名称")
@click.option("--address", help="地址")
@click.option("--contact", help="联系人")
@click.option("--phone", help="联系电话")
@click.option("--license", help="许可证号")
@click.option("--user", help="操作人")
def enterprise_update(id, name, address, contact, phone, license, user):
    data = {}
    if name:
        data["name"] = name
    if address:
        data["address"] = address
    if contact:
        data["contact_person"] = contact
    if phone:
        data["contact_phone"] = phone
    if license:
        data["license_number"] = license
    result = client().put(f"/enterprises/{id}", data, {"user": user})
    print_json(result)


@enterprise.command("delete")
@click.argument("id", type=int)
@click.option("--user", help="操作人")
def enterprise_delete(id, user):
    result = client().delete(f"/enterprises/{id}", {"user": user})
    print_json(result)


@cli.group()
def chemical():
    pass


@chemical.command("list")
@click.option("--enterprise", type=int, help="企业ID")
@click.option("--category", help="类别")
@click.option("--name", help="名称")
def chemical_list(enterprise, category, name):
    result = client().get("/chemicals", {
        "enterprise_id": enterprise,
        "category": category,
        "name": name,
    })
    print_json(result)


@chemical.command("get")
@click.argument("id", type=int)
def chemical_get(id):
    result = client().get(f"/chemicals/{id}")
    print_json(result)


@chemical.command("create")
@click.option("--enterprise", type=int, required=True, help="企业ID")
@click.option("--name", required=True, help="危化品名称")
@click.option("--category", help="类别")
@click.option("--cas", help="CAS号")
@click.option("--hazard", help="危险等级")
@click.option("--desc", help="描述")
@click.option("--user", help="操作人")
def chemical_create(enterprise, name, category, cas, hazard, desc, user):
    data = {"name": name}
    if category:
        data["category"] = category
    if cas:
        data["cas_number"] = cas
    if hazard:
        data["hazard_level"] = hazard
    if desc:
        data["description"] = desc
    result = client().post(f"/enterprises/{enterprise}/chemicals", data, {"user": user})
    print_json(result)


@chemical.command("update")
@click.argument("id", type=int)
@click.option("--name", help="危化品名称")
@click.option("--category", help="类别")
@click.option("--cas", help="CAS号")
@click.option("--hazard", help="危险等级")
@click.option("--desc", help="描述")
@click.option("--user", help="操作人")
def chemical_update(id, name, category, cas, hazard, desc, user):
    data = {}
    if name:
        data["name"] = name
    if category:
        data["category"] = category
    if cas:
        data["cas_number"] = cas
    if hazard:
        data["hazard_level"] = hazard
    if desc:
        data["description"] = desc
    result = client().put(f"/chemicals/{id}", data, {"user": user})
    print_json(result)


@chemical.command("delete")
@click.argument("id", type=int)
@click.option("--user", help="操作人")
def chemical_delete(id, user):
    result = client().delete(f"/chemicals/{id}", {"user": user})
    print_json(result)


@cli.group()
def approval():
    pass


@approval.command("list")
@click.option("--enterprise", type=int, help="企业ID")
@click.option("--chemical", type=int, help="危化品ID")
def approval_list(enterprise, chemical):
    result = client().get("/approvals", {
        "enterprise_id": enterprise,
        "chemical_id": chemical,
    })
    print_json(result)


@approval.command("get")
@click.argument("id", type=int)
def approval_get(id):
    result = client().get(f"/approvals/{id}")
    print_json(result)


@approval.command("create")
@click.option("--enterprise", type=int, required=True, help="企业ID")
@click.option("--chemical", type=int, required=True, help="危化品ID")
@click.option("--level", required=True, type=click.Choice(["A", "B", "C"]), help="审批级别")
@click.option("--max-qty", type=float, required=True, help="最大许可量")
@click.option("--user", help="操作人")
def approval_create(enterprise, chemical, level, max_qty, user):
    data = {
        "enterprise_id": enterprise,
        "chemical_id": chemical,
        "approval_level": level,
        "max_quantity": max_qty,
    }
    result = client().post("/approvals", data, {"user": user})
    print_json(result)


@approval.command("approve")
@click.argument("id", type=int)
@click.option("--level", type=int, required=True, help="审批级别 1/2/3")
@click.option("--comment", help="审批意见")
@click.option("--user", help="操作人")
def approval_approve(id, level, comment, user):
    data = {"level": level}
    if comment:
        data["comment"] = comment
    if user:
        data["user"] = user
    result = client().post(f"/approvals/{id}/approve", data)
    print_json(result)


@approval.command("reject")
@click.argument("id", type=int)
@click.option("--level", type=int, required=True, help="审批级别 1/2/3")
@click.option("--comment", help="拒绝理由")
@click.option("--user", help="操作人")
def approval_reject(id, level, comment, user):
    data = {"level": level}
    if comment:
        data["comment"] = comment
    if user:
        data["user"] = user
    result = client().post(f"/approvals/{id}/reject", data)
    print_json(result)


@approval.command("renew")
@click.argument("id", type=int)
@click.option("--user", help="操作人")
def approval_renew(id, user):
    result = client().post(f"/approvals/{id}/renew", params={"user": user})
    print_json(result)


@approval.command("expiring")
def approval_expiring():
    result = client().get("/approvals/expiring/soon")
    print_json(result)


@approval.command("expired")
def approval_expired():
    result = client().get("/approvals/expired")
    print_json(result)


@cli.group()
def plan():
    pass


@plan.command("list")
@click.option("--enterprise", type=int, help="企业ID")
def plan_list(enterprise):
    result = client().get("/plans", {"enterprise_id": enterprise})
    print_json(result)


@plan.command("get")
@click.argument("id", type=int)
def plan_get(id):
    result = client().get(f"/plans/{id}")
    print_json(result)


@plan.command("create")
@click.option("--enterprise", type=int, required=True, help="企业ID")
@click.option("--title", required=True, help="预案标题")
@click.option("--version", help="版本")
@click.option("--content", required=True, help="预案内容")
@click.option("--user", help="操作人")
def plan_create(enterprise, title, version, content, user):
    data = {
        "enterprise_id": enterprise,
        "title": title,
        "content": content,
    }
    if version:
        data["version"] = version
    result = client().post("/plans", data, {"user": user})
    print_json(result)


@plan.command("update")
@click.argument("id", type=int)
@click.option("--title", help="预案标题")
@click.option("--version", help="版本")
@click.option("--content", help="预案内容")
@click.option("--user", help="操作人")
def plan_update(id, title, version, content, user):
    data = {}
    if title:
        data["title"] = title
    if version:
        data["version"] = version
    if content:
        data["content"] = content
    result = client().put(f"/plans/{id}", data, {"user": user})
    print_json(result)


@cli.group()
def accident():
    pass


@accident.command("list")
@click.option("--enterprise", type=int, help="企业ID")
@click.option("--status", help="状态")
def accident_list(enterprise, status):
    result = client().get("/accidents", {
        "enterprise_id": enterprise,
        "status": status,
    })
    print_json(result)


@accident.command("get")
@click.argument("id", type=int)
def accident_get(id):
    result = client().get(f"/accidents/{id}")
    print_json(result)


@accident.command("create")
@click.option("--enterprise", type=int, required=True, help="企业ID")
@click.option("--title", required=True, help="事故标题")
@click.option("--date", required=True, help="事故时间 ISO格式")
@click.option("--location", help="地点")
@click.option("--desc", required=True, help="描述")
@click.option("--severity", help="严重程度")
@click.option("--casualties", type=int, default=0, help="伤亡人数")
@click.option("--loss", type=float, default=0.0, help="经济损失")
@click.option("--status", default="processing", help="状态")
@click.option("--user", help="操作人")
def accident_create(enterprise, title, date, location, desc, severity, casualties, loss, status, user):
    data = {
        "enterprise_id": enterprise,
        "title": title,
        "accident_date": date,
        "description": desc,
        "casualties": casualties,
        "financial_loss": loss,
        "status": status,
    }
    if location:
        data["location"] = location
    if severity:
        data["severity"] = severity
    result = client().post("/accidents", data, {"user": user})
    print_json(result)


@accident.command("disposition")
@click.argument("id", type=int)
@click.option("--content", required=True, help="处置内容")
@click.option("--user", help="操作人")
def accident_disposition(id, content, user):
    data = {"disposition": content}
    if user:
        data["user"] = user
    result = client().post(f"/accidents/{id}/disposition", data)
    print_json(result)


@cli.group()
def ledger():
    pass


@ledger.command("list")
@click.option("--enterprise", type=int, help="企业ID")
@click.option("--chemical", type=int, help="危化品ID")
def ledger_list(enterprise, chemical):
    result = client().get("/ledgers", {
        "enterprise_id": enterprise,
        "chemical_id": chemical,
    })
    print_json(result)


@ledger.command("add")
@click.option("--enterprise", type=int, required=True, help="企业ID")
@click.option("--chemical", type=int, required=True, help="危化品ID")
@click.option("--type", "tx_type", required=True, type=click.Choice(["in", "out"]), help="in/out")
@click.option("--qty", type=float, required=True, help="数量")
@click.option("--unit", default="kg", help="单位")
@click.option("--date", help="交易日期")
@click.option("--desc", help="描述")
@click.option("--user", help="操作人")
def ledger_add(enterprise, chemical, tx_type, qty, unit, date, desc, user):
    data = {
        "enterprise_id": enterprise,
        "chemical_id": chemical,
        "transaction_type": tx_type,
        "quantity": qty,
        "unit": unit,
    }
    if date:
        data["transaction_date"] = date
    if desc:
        data["description"] = desc
    result = client().post("/ledgers/transactions", data, {"user": user})
    print_json(result)


@ledger.command("export")
@click.option("--enterprise", type=int, help="企业ID")
@click.option("--category", help="危化品类别")
@click.option("--output", "-o", help="输出文件路径")
def ledger_export(enterprise, category, output):
    csv_content = client().get_csv("/ledgers/export/csv", {
        "enterprise_id": enterprise,
        "category": category,
    })
    if output:
        with open(output, "w", encoding="utf-8") as f:
            f.write(csv_content)
        click.echo(f"已导出到: {output}")
    else:
        click.echo(csv_content)


@cli.group()
def drill():
    pass


@drill.command("list")
@click.option("--enterprise", type=int, help="企业ID")
def drill_list(enterprise):
    result = client().get("/drills", {"enterprise_id": enterprise})
    print_json(result)


@drill.command("get")
@click.argument("id", type=int)
def drill_get(id):
    result = client().get(f"/drills/{id}")
    print_json(result)


@drill.command("create")
@click.option("--enterprise", type=int, required=True, help="企业ID")
@click.option("--title", required=True, help="演练标题")
@click.option("--date", required=True, help="演练日期")
@click.option("--desc", help="描述")
@click.option("--participants", type=int, help="参与人数")
@click.option("--hours", type=float, help="时长(小时)")
@click.option("--result", "eval_result", help="评估结果")
@click.option("--eval-details", help="评估详情")
@click.option("--plan", type=int, help="关联预案ID")
@click.option("--user", help="操作人")
def drill_create(enterprise, title, date, desc, participants, hours, eval_result, eval_details, plan, user):
    data = {
        "enterprise_id": enterprise,
        "title": title,
        "drill_date": date,
    }
    if desc:
        data["description"] = desc
    if participants:
        data["participants"] = participants
    if hours:
        data["duration_hours"] = hours
    if eval_result:
        data["evaluation_result"] = eval_result
    if eval_details:
        data["evaluation_details"] = eval_details
    if plan:
        data["plan_id"] = plan
    result = client().post("/drills", data, {"user": user})
    print_json(result)


@cli.command("audit-logs")
@click.option("--entity-type", help="实体类型")
@click.option("--entity-id", type=int, help="实体ID")
def audit_logs(entity_type, entity_id):
    result = client().get("/audit-logs", {
        "entity_type": entity_type,
        "entity_id": entity_id,
    })
    print_json(result)


@cli.group()
def reminder():
    pass


@reminder.command("list")
@click.option("--all", is_flag=True, help="显示所有(包含已完成)")
def reminder_list(all):
    result = client().get("/reminders", {"only_active": not all})
    print_json(result)


@reminder.command("generate")
def reminder_generate():
    result = client().post("/reminders/generate")
    print_json(result)


@reminder.command("complete")
@click.argument("id", type=int)
def reminder_complete(id):
    result = client().post(f"/reminders/{id}/complete")
    print_json(result)


@cli.group()
def check():
    pass


@check.command("run")
def check_run():
    result = client().post("/ledger-checks")
    print_json(result)


@check.command("abnormal")
def check_abnormal():
    result = client().get("/ledger-checks/abnormal")
    print_json(result)


if __name__ == "__main__":
    cli()
