"""命令行客户端"""
import typer
import json
from typing import Optional
from datetime import date

from .api_client import client

app = typer.Typer(help="宠物寄养中心管理系统命令行客户端")
zone_cmd = typer.Typer(help="区域管理")
cage_cmd = typer.Typer(help="笼位管理")
pet_cmd = typer.Typer(help="宠物管理")
boarding_cmd = typer.Typer(help="寄养管理")
feeding_cmd = typer.Typer(help="喂养管理")
todo_cmd = typer.Typer(help="待办事项")

app.add_typer(zone_cmd, name="zone")
app.add_typer(cage_cmd, name="cage")
app.add_typer(pet_cmd, name="pet")
app.add_typer(boarding_cmd, name="boarding")
app.add_typer(feeding_cmd, name="feeding")
app.add_typer(todo_cmd, name="todo")

def print_json(data):
    typer.echo(json.dumps(data, ensure_ascii=False, indent=2))

def print_table(headers, rows):
    if not rows:
        typer.echo("暂无数据")
        return
    col_widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            col_widths[i] = max(col_widths[i], len(str(cell) if cell is not None else ""))
    separator = "+" + "+".join("-" * (w + 2) for w in col_widths) + "+"
    typer.echo(separator)
    header_row = "|" + "|".join(f" {h.ljust(w)} " for h, w in zip(headers, col_widths)) + "|"
    typer.echo(header_row)
    typer.echo(separator)
    for row in rows:
        data_row = "|" + "|".join(f" {str(cell if cell is not None else '').ljust(w)} " for cell, w in zip(row, col_widths)) + "|"
        typer.echo(data_row)
    typer.echo(separator)

@app.command("status", help="检查服务端状态")
def server_status():
    result = client.get("/")
    print_json(result)

@zone_cmd.command("create", help="创建区域")
def zone_create(name: str = typer.Option(..., "--name", "-n", help="区域名称"),
                description: Optional[str] = typer.Option(None, "--desc", "-d", help="区域描述")):
    result = client.post("/zones", {"name": name, "description": description})
    typer.echo(f"✓ 区域创建成功，ID: {result['id']}")
    print_json(result)

@zone_cmd.command("list", help="列出所有区域")
def zone_list():
    zones = client.get("/zones")
    print_table(
        ["ID", "名称", "描述", "创建时间"],
        [(z["id"], z["name"], z["description"] or "-", z["created_at"] or "-") for z in zones]
    )

@zone_cmd.command("delete", help="删除区域")
def zone_delete(zone_id: int = typer.Argument(..., help="区域ID")):
    result = client.delete(f"/zones/{zone_id}")
    typer.echo(f"✓ {result['message']}")

@cage_cmd.command("create", help="创建笼位")
def cage_create(zone_id: int = typer.Option(..., "--zone-id", "-z", help="区域ID"),
                code: str = typer.Option(..., "--code", "-c", help="笼位编号"),
                daily_rate: float = typer.Option(50.0, "--rate", "-r", help="每日费用"),
                description: Optional[str] = typer.Option(None, "--desc", "-d", help="描述")):
    result = client.post("/cages", {
        "zone_id": zone_id,
        "code": code,
        "daily_rate": daily_rate,
        "description": description
    })
    typer.echo(f"✓ 笼位创建成功，ID: {result['id']}")
    print_json(result)

@cage_cmd.command("list", help="列出所有笼位")
def cage_list(zone_id: Optional[int] = typer.Option(None, "--zone-id", "-z", help="按区域过滤"),
              status: Optional[str] = typer.Option(None, "--status", "-s", 
                   help="状态过滤: 空闲/占用/清洁中/维修中")):
    params = {}
    if zone_id:
        params["zone_id"] = zone_id
    if status:
        params["status"] = status
    cages = client.get("/cages", params=params)
    print_table(
        ["ID", "编号", "区域", "状态", "日费率", "描述"],
        [(c["id"], c["code"], c["zone_name"] or "-", c["status"], c["daily_rate"], c["description"] or "-") 
         for c in cages]
    )

@cage_cmd.command("status", help="更新笼位状态")
def cage_status(cage_id: int = typer.Argument(..., help="笼位ID"),
                status: str = typer.Option(..., "--status", "-s", 
                         help="新状态: 空闲/占用/清洁中/维修中")):
    result = client.put(f"/cages/{cage_id}/status", {"status": status})
    typer.echo(f"✓ 笼位状态已更新为: {result['status']}")

@pet_cmd.command("create", help="创建宠物")
def pet_create(name: str = typer.Option(..., "--name", "-n", help="宠物名称"),
               pet_type: str = typer.Option(..., "--type", "-t", help="类型: 狗/猫/鸟/其他"),
               owner_name: str = typer.Option(..., "--owner", "-o", help="主人姓名"),
               owner_phone: str = typer.Option(..., "--phone", "-p", help="主人电话"),
               breed: Optional[str] = typer.Option(None, "--breed", "-b", help="品种"),
               age: Optional[int] = typer.Option(None, "--age", help="年龄"),
               notes: Optional[str] = typer.Option(None, "--notes", help="备注")):
    result = client.post("/pets", {
        "name": name,
        "type": pet_type,
        "owner_name": owner_name,
        "owner_phone": owner_phone,
        "breed": breed,
        "age": age,
        "notes": notes
    })
    typer.echo(f"✓ 宠物创建成功，ID: {result['id']}")
    print_json(result)

@pet_cmd.command("list", help="列出所有宠物")
def pet_list():
    pets = client.get("/pets")
    print_table(
        ["ID", "名称", "类型", "品种", "年龄", "主人", "电话"],
        [(p["id"], p["name"], p["type"], p["breed"] or "-", p["age"] or "-", 
          p["owner_name"], p["owner_phone"]) for p in pets]
    )

@pet_cmd.command("get", help="获取宠物详情")
def pet_get(pet_id: int = typer.Argument(..., help="宠物ID")):
    pet = client.get(f"/pets/{pet_id}")
    print_json(pet)

@pet_cmd.command("update", help="更新宠物信息")
def pet_update(pet_id: int = typer.Argument(..., help="宠物ID"),
               name: Optional[str] = typer.Option(None, "--name", "-n"),
               owner_name: Optional[str] = typer.Option(None, "--owner", "-o"),
               owner_phone: Optional[str] = typer.Option(None, "--phone", "-p"),
               notes: Optional[str] = typer.Option(None, "--notes")):
    data = {}
    if name:
        data["name"] = name
    if owner_name:
        data["owner_name"] = owner_name
    if owner_phone:
        data["owner_phone"] = owner_phone
    if notes is not None:
        data["notes"] = notes
    result = client.put(f"/pets/{pet_id}", data)
    typer.echo("✓ 宠物信息已更新")
    print_json(result)

@boarding_cmd.command("checkin", help="入住登记")
def boarding_checkin(pet_id: int = typer.Option(..., "--pet-id", "-p", help="宠物ID"),
                     cage_id: int = typer.Option(..., "--cage-id", "-c", help="笼位ID"),
                     start_date: str = typer.Option(..., "--start", "-s", help="入住日期 YYYY-MM-DD"),
                     end_date: str = typer.Option(..., "--end", "-e", help="预计离开日期 YYYY-MM-DD"),
                     service_fee: float = typer.Option(0.0, "--fee", "-f", help="服务费"),
                     notes: Optional[str] = typer.Option(None, "--notes", help="备注")):
    result = client.post("/boardings", {
        "pet_id": pet_id,
        "cage_id": cage_id,
        "start_date": start_date,
        "expected_end_date": end_date,
        "service_fee": service_fee,
        "notes": notes
    })
    typer.echo(f"✓ 入住登记成功，记录ID: {result['id']}")
    print_json(result)

@boarding_cmd.command("list", help="列出入住记录")
def boarding_list(status: Optional[str] = typer.Option(None, "--status", "-s", 
                      help="状态过滤: 入住中/逾期/已退房")):
    params = {}
    if status:
        params["status"] = status
    records = client.get("/boardings", params=params)
    print_table(
        ["ID", "宠物", "笼位", "入住日期", "预计离开", "实际离开", "状态", "总费用"],
        [(r["id"], r["pet_name"] or "-", r["cage_code"] or "-", 
          r["start_date"], r["expected_end_date"], 
          r["actual_end_date"] or "-", r["status"], 
          f"¥{r['total_amount']}" if r["total_amount"] is not None else "-") 
         for r in records]
    )

@boarding_cmd.command("get", help="获取入住记录详情")
def boarding_get(record_id: int = typer.Argument(..., help="入住记录ID")):
    record = client.get(f"/boardings/{record_id}")
    print_json(record)

@boarding_cmd.command("checkout", help="退房结账")
def boarding_checkout(record_id: int = typer.Argument(..., help="入住记录ID"),
                       actual_end_date: Optional[str] = typer.Option(None, "--end", "-e", 
                                             help="实际退房日期 YYYY-MM-DD"),
                       service_fee: Optional[float] = typer.Option(None, "--fee", "-f", 
                                               help="服务费")):
    data = {}
    if actual_end_date:
        data["actual_end_date"] = actual_end_date
    if service_fee is not None:
        data["service_fee"] = service_fee
    result = client.post(f"/boardings/{record_id}/checkout", data)
    typer.echo(f"✓ 退房成功，总费用: ¥{result['total_amount']}")
    print_json(result)

@boarding_cmd.command("check-overdue", help="检查逾期记录")
def check_overdue():
    result = client.post("/boardings/check-overdue")
    typer.echo(result["message"])

@feeding_cmd.command("log", help="记录喂养日志")
def feeding_log(boarding_id: int = typer.Option(..., "--boarding-id", "-b", help="入住记录ID"),
                mental_state: str = typer.Option(..., "--mental", "-m", help="精神状态"),
                defecation: str = typer.Option(..., "--defecation", "-d", help="排便情况"),
                log_date: Optional[str] = typer.Option(None, "--date", help="日期 YYYY-MM-DD"),
                feeding_time: Optional[str] = typer.Option(None, "--time", help="喂养时间"),
                food_type: Optional[str] = typer.Option(None, "--food", help="食物类型"),
                abnormal_symptoms: Optional[str] = typer.Option(None, "--symptoms", help="异常症状"),
                notes: Optional[str] = typer.Option(None, "--notes", help="备注")):
    result = client.post("/feeding-logs", {
        "boarding_id": boarding_id,
        "log_date": log_date,
        "feeding_time": feeding_time,
        "food_type": food_type,
        "mental_state": mental_state,
        "defecation": defecation,
        "abnormal_symptoms": abnormal_symptoms,
        "notes": notes
    })
    typer.echo(f"✓ 喂养日志记录成功，ID: {result['id']}")
    if abnormal_symptoms:
        typer.echo("⚠ 已检测到异常症状，自动生成兽医检查待办")
    print_json(result)

@feeding_cmd.command("list", help="列出喂养日志")
def feeding_list(boarding_id: Optional[int] = typer.Option(None, "--boarding-id", "-b", 
                                    help="按入住记录过滤")):
    params = {}
    if boarding_id:
        params["boarding_id"] = boarding_id
    logs = client.get("/feeding-logs", params=params)
    print_table(
        ["ID", "入住ID", "日期", "精神状态", "排便", "异常症状"],
        [(l["id"], l["boarding_id"], l["log_date"], l["mental_state"] or "-", 
          l["defecation"] or "-", l["abnormal_symptoms"] or "-") for l in logs]
    )

@feeding_cmd.command("health", help="记录健康数据")
def feeding_health(boarding_id: int = typer.Option(..., "--boarding-id", "-b", help="入住记录ID"),
                   temperature: Optional[float] = typer.Option(None, "--temp", help="体温"),
                   weight: Optional[float] = typer.Option(None, "--weight", help="体重"),
                   heart_rate: Optional[int] = typer.Option(None, "--heart", help="心率"),
                   respiratory_rate: Optional[int] = typer.Option(None, "--resp", help="呼吸频率"),
                   notes: Optional[str] = typer.Option(None, "--notes", help="备注")):
    result = client.post("/health-data", {
        "boarding_id": boarding_id,
        "temperature": temperature,
        "weight": weight,
        "heart_rate": heart_rate,
        "respiratory_rate": respiratory_rate,
        "notes": notes
    })
    typer.echo(f"✓ 健康数据记录成功，ID: {result['id']}")
    print_json(result)

@feeding_cmd.command("health-list", help="列出健康数据")
def health_list(boarding_id: Optional[int] = typer.Option(None, "--boarding-id", "-b", 
                                  help="按入住记录过滤")):
    params = {}
    if boarding_id:
        params["boarding_id"] = boarding_id
    data_list = client.get("/health-data", params=params)
    print_table(
        ["ID", "入住ID", "日期", "体温", "体重", "心率", "呼吸"],
        [(h["id"], h["boarding_id"], h["record_date"], 
          h["temperature"] or "-", h["weight"] or "-", 
          h["heart_rate"] or "-", h["respiratory_rate"] or "-") for h in data_list]
    )

@todo_cmd.command("create", help="创建待办")
def todo_create(title: str = typer.Option(..., "--title", "-t", help="标题"),
                description: Optional[str] = typer.Option(None, "--desc", "-d", help="描述"),
                boarding_id: Optional[int] = typer.Option(None, "--boarding-id", "-b", help="关联入住记录"),
                veterinary: bool = typer.Option(False, "--vet", help="兽医相关"),
                due_date: Optional[str] = typer.Option(None, "--due", help="截止日期")):
    result = client.post("/todos", {
        "title": title,
        "description": description,
        "boarding_id": boarding_id,
        "is_veterinary": veterinary,
        "due_date": due_date
    })
    typer.echo(f"✓ 待办创建成功，ID: {result['id']}")
    print_json(result)

@todo_cmd.command("list", help="列出待办")
def todo_list(status: Optional[str] = typer.Option(None, "--status", "-s", 
                   help="状态过滤: 待处理/处理中/已完成"),
              veterinary: Optional[bool] = typer.Option(None, "--vet", help="兽医相关")):
    params = {}
    if status:
        params["status"] = status
    if veterinary is not None:
        params["is_veterinary"] = veterinary
    todos = client.get("/todos", params=params)
    print_table(
        ["ID", "标题", "状态", "兽医", "截止日期"],
        [(t["id"], t["title"], t["status"], "是" if t["is_veterinary"] else "否", 
          t["due_date"] or "-") for t in todos]
    )

@todo_cmd.command("status", help="更新待办状态")
def todo_status(todo_id: int = typer.Argument(..., help="待办ID"),
                status: str = typer.Option(..., "--status", "-s", 
                         help="状态: 待处理/处理中/已完成")):
    result = client.put(f"/todos/{todo_id}/status", {"status": status})
    typer.echo(f"✓ 待办状态已更新为: {result['status']}")

@todo_cmd.command("delete", help="删除待办")
def todo_delete(todo_id: int = typer.Argument(..., help="待办ID")):
    result = client.delete(f"/todos/{todo_id}")
    typer.echo(f"✓ {result['message']}")

if __name__ == "__main__":
    app()
