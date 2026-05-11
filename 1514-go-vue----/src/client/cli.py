from datetime import date, time
from typing import List, Optional

import typer
from rich.console import Console
from rich.table import Table

from src.client.api_client import PetClinicClient

app = typer.Typer(help="宠物医院门诊管理系统 CLI")
console = Console()
client = PetClinicClient()


@app.command("create-owner")
def create_owner(
    name: str = typer.Option(..., help="主人姓名"),
    phone: str = typer.Option(..., help="联系电话"),
    address: Optional[str] = typer.Option(None, help="地址"),
):
    result = client.create_owner(name=name, phone=phone, address=address)
    console.print(f"[green]创建主人成功[/green]")
    console.print(result)


@app.command("list-owners")
def list_owners():
    owners = client.list_owners()
    table = Table(title="主人列表")
    table.add_column("ID", style="cyan")
    table.add_column("姓名", style="magenta")
    table.add_column("电话", style="yellow")
    table.add_column("地址")
    for owner in owners:
        table.add_row(
            owner.get("id", ""),
            owner.get("name", ""),
            owner.get("phone", ""),
            owner.get("address") or "",
        )
    console.print(table)


@app.command("create-pet")
def create_pet(
    name: str = typer.Option(..., help="宠物名称"),
    species: str = typer.Option(..., help="物种（如：狗、猫）"),
    owner_id: str = typer.Option(..., help="主人ID"),
    breed: Optional[str] = typer.Option(None, help="品种"),
    gender: Optional[str] = typer.Option(None, help="性别"),
    birth_date: Optional[str] = typer.Option(None, help="出生日期 YYYY-MM-DD"),
):
    birth = date.fromisoformat(birth_date) if birth_date else None
    result = client.create_pet(
        name=name,
        species=species,
        owner_id=owner_id,
        breed=breed,
        gender=gender,
        birth_date=birth,
    )
    console.print(f"[green]创建宠物档案成功[/green]")
    console.print(result)


@app.command("list-pets")
def list_pets():
    pets = client.list_pets()
    table = Table(title="宠物列表")
    table.add_column("ID", style="cyan")
    table.add_column("名称", style="magenta")
    table.add_column("物种", style="yellow")
    table.add_column("品种")
    table.add_column("主人ID")
    for pet in pets:
        table.add_row(
            pet.get("id", ""),
            pet.get("name", ""),
            pet.get("species", ""),
            pet.get("breed") or "",
            pet.get("owner_id", ""),
        )
    console.print(table)


@app.command("create-doctor")
def create_doctor(
    name: str = typer.Option(..., help="医生姓名"),
    specialty: Optional[str] = typer.Option(None, help="专长"),
    phone: Optional[str] = typer.Option(None, help="电话"),
):
    result = client.create_doctor(name=name, specialty=specialty, phone=phone)
    console.print(f"[green]创建医生成功[/green]")
    console.print(result)


@app.command("list-doctors")
def list_doctors():
    doctors = client.list_doctors()
    table = Table(title="医生列表")
    table.add_column("ID", style="cyan")
    table.add_column("姓名", style="magenta")
    table.add_column("专长")
    table.add_column("电话")
    for doctor in doctors:
        table.add_row(
            doctor.get("id", ""),
            doctor.get("name", ""),
            doctor.get("specialty") or "",
            doctor.get("phone") or "",
        )
    console.print(table)


@app.command("register")
def register(
    pet_id: str = typer.Option(..., help="宠物ID"),
    owner_id: str = typer.Option(..., help="主人ID"),
    symptoms: Optional[str] = typer.Option(None, help="症状描述"),
    doctor_id: Optional[str] = typer.Option(None, help="指定医生ID"),
):
    result = client.create_registration(
        pet_id=pet_id, owner_id=owner_id, symptoms=symptoms, doctor_id=doctor_id
    )
    console.print(f"[green]挂号成功[/green]")
    console.print(result)


@app.command("waiting-queue")
def waiting_queue():
    registrations = client.get_waiting_queue()
    table = Table(title="候诊队列")
    table.add_column("挂号ID", style="cyan")
    table.add_column("宠物ID", style="magenta")
    table.add_column("主人ID")
    table.add_column("症状")
    table.add_column("状态")
    for reg in registrations:
        table.add_row(
            reg.get("id", ""),
            reg.get("pet_id", ""),
            reg.get("owner_id", ""),
            reg.get("symptoms") or "",
            reg.get("status", ""),
        )
    console.print(table)


@app.command("start-treatment")
def start_treatment(
    registration_id: str = typer.Option(..., help="挂号ID"),
    doctor_id: str = typer.Option(..., help="医生ID"),
):
    result = client.start_treatment(
        registration_id=registration_id, doctor_id=doctor_id
    )
    console.print(f"[green]开始接诊[/green]")
    console.print(result)


@app.command("complete-registration")
def complete_registration(
    registration_id: str = typer.Option(..., help="挂号ID"),
):
    result = client.complete_registration(registration_id=registration_id)
    console.print(f"[green]完成诊疗[/green]")
    console.print(result)


@app.command("diagnose")
def diagnose(
    registration_id: str = typer.Option(..., help="挂号ID"),
    doctor_id: str = typer.Option(..., help="医生ID"),
    diagnosis_text: str = typer.Option(..., help="诊断结果"),
    remarks: Optional[str] = typer.Option(None, help="备注"),
):
    result = client.create_diagnosis(
        registration_id=registration_id,
        doctor_id=doctor_id,
        diagnosis=diagnosis_text,
        remarks=remarks,
    )
    console.print(f"[green]记录诊断成功[/green]")
    console.print(result)


@app.command("get-fee")
def get_fee(
    registration_id: str = typer.Option(..., help="挂号ID"),
):
    fee = client.get_fee_record(registration_id=registration_id)
    console.print(f"[bold]费用明细[/bold]")
    console.print(f"总金额: {fee.get('total_amount', 0)}")
    console.print(f"已支付: {'是' if fee.get('paid') else '否'}")
    table = Table(title="费用项目")
    table.add_column("项目名称", style="cyan")
    table.add_column("类型", style="magenta")
    table.add_column("单价")
    table.add_column("数量")
    table.add_column("小计")
    table.add_column("计费")
    for item in fee.get("items", []):
        table.add_row(
            item.get("item_name", ""),
            item.get("item_type", ""),
            str(item.get("unit_price", 0)),
            str(item.get("quantity", 0)),
            str(item.get("subtotal", 0)),
            "是" if item.get("is_charged") else "否",
        )
    console.print(table)


@app.command("list-unpaid-fees")
def list_unpaid_fees():
    fees = client.list_unpaid_fees()
    table = Table(title="未支付费用")
    table.add_column("费用ID", style="cyan")
    table.add_column("挂号ID", style="magenta")
    table.add_column("总金额")
    for fee in fees:
        table.add_row(
            fee.get("id", ""),
            fee.get("registration_id", ""),
            str(fee.get("total_amount", 0)),
        )
    console.print(table)


@app.command("pay-fee")
def pay_fee(
    fee_id: str = typer.Option(..., help="费用ID"),
):
    result = client.pay_fee(fee_id=fee_id)
    console.print(f"[green]支付成功[/green]")
    console.print(result)


@app.command("create-vaccine")
def create_vaccine(
    name: str = typer.Option(..., help="疫苗名称"),
    manufacturer: Optional[str] = typer.Option(None, help="生产商"),
    interval_days: Optional[int] = typer.Option(None, help="推荐间隔天数"),
):
    result = client.create_vaccine(
        name=name,
        manufacturer=manufacturer,
        recommended_interval_days=interval_days,
    )
    console.print(f"[green]创建疫苗成功[/green]")
    console.print(result)


@app.command("list-vaccines")
def list_vaccines():
    vaccines = client.list_vaccines()
    table = Table(title="疫苗列表")
    table.add_column("ID", style="cyan")
    table.add_column("名称", style="magenta")
    table.add_column("生产商")
    table.add_column("推荐间隔(天)")
    for vaccine in vaccines:
        table.add_row(
            vaccine.get("id", ""),
            vaccine.get("name", ""),
            vaccine.get("manufacturer") or "",
            str(vaccine.get("recommended_interval_days") or ""),
        )
    console.print(table)


@app.command("record-vaccination")
def record_vaccination(
    pet_id: str = typer.Option(..., help="宠物ID"),
    vaccine_id: str = typer.Option(..., help="疫苗ID"),
    inoculation_date: str = typer.Option(..., help="接种日期 YYYY-MM-DD"),
    doctor_id: Optional[str] = typer.Option(None, help="医生ID"),
    batch_number: Optional[str] = typer.Option(None, help="批号"),
    next_date: Optional[str] = typer.Option(None, help="下次接种日期 YYYY-MM-DD"),
    remarks: Optional[str] = typer.Option(None, help="备注"),
):
    inoc_date = date.fromisoformat(inoculation_date)
    next_dt = date.fromisoformat(next_date) if next_date else None
    result = client.record_vaccination(
        pet_id=pet_id,
        vaccine_id=vaccine_id,
        inoculation_date=inoc_date,
        doctor_id=doctor_id,
        batch_number=batch_number,
        next_inoculation_date=next_dt,
        remarks=remarks,
    )
    console.print(f"[green]记录接种成功[/green]")
    console.print(result)


@app.command("vaccine-history")
def vaccine_history(
    pet_id: str = typer.Option(..., help="宠物ID"),
):
    records = client.get_vaccine_history(pet_id=pet_id)
    table = Table(title=f"宠物 {pet_id} 疫苗接种历史")
    table.add_column("疫苗名称", style="cyan")
    table.add_column("接种日期", style="magenta")
    table.add_column("下次接种")
    table.add_column("批号")
    for record in records:
        table.add_row(
            record.get("vaccine_name", ""),
            record.get("inoculation_date", ""),
            record.get("next_inoculation_date") or "",
            record.get("batch_number") or "",
        )
    console.print(table)


@app.command("create-work-slot")
def create_work_slot(
    doctor_id: str = typer.Option(..., help="医生ID"),
    slot_date: str = typer.Option(..., help="日期 YYYY-MM-DD"),
    start_time: str = typer.Option(..., help="开始时间 HH:MM"),
    end_time: str = typer.Option(..., help="结束时间 HH:MM"),
    max_appointments: int = typer.Option(1, help="最大预约数"),
):
    slot_dt = date.fromisoformat(slot_date)
    start = time.fromisoformat(start_time)
    end = time.fromisoformat(end_time)
    result = client.create_work_slot(
        doctor_id=doctor_id,
        slot_date=slot_dt,
        start_time=start,
        end_time=end,
        max_appointments=max_appointments,
    )
    console.print(f"[green]创建工作时段成功[/green]")
    console.print(result)


@app.command("list-work-slots")
def list_work_slots(
    doctor_id: Optional[str] = typer.Option(None, help="医生ID"),
    slot_date: Optional[str] = typer.Option(None, help="日期 YYYY-MM-DD"),
):
    slot_dt = date.fromisoformat(slot_date) if slot_date else None
    slots = client.list_work_slots(doctor_id=doctor_id, slot_date=slot_dt)
    table = Table(title="工作时段")
    table.add_column("ID", style="cyan")
    table.add_column("医生ID", style="magenta")
    table.add_column("日期")
    table.add_column("开始")
    table.add_column("结束")
    table.add_column("当前/最大")
    table.add_column("可用")
    for slot in slots:
        table.add_row(
            slot.get("id", ""),
            slot.get("doctor_id", ""),
            slot.get("date", ""),
            slot.get("start_time", ""),
            slot.get("end_time", ""),
            f"{slot.get('current_appointments', 0)}/{slot.get('max_appointments', 0)}",
            "是" if slot.get("is_available") else "否",
        )
    console.print(table)


@app.command("book-appointment")
def book_appointment(
    owner_id: str = typer.Option(..., help="主人ID"),
    pet_id: str = typer.Option(..., help="宠物ID"),
    work_slot_id: str = typer.Option(..., help="工作时段ID"),
    remarks: Optional[str] = typer.Option(None, help="备注"),
):
    result = client.create_appointment(
        owner_id=owner_id,
        pet_id=pet_id,
        work_slot_id=work_slot_id,
        remarks=remarks,
    )
    status = result.get("status", "")
    if status == "waitlist":
        console.print(
            f"[yellow]该时段已满，已加入候补队列，位置：{result.get('waitlist_position')}[/yellow]"
        )
    else:
        console.print(f"[green]预约成功[/green]")
    console.print(result)


@app.command("list-appointments")
def list_appointments(
    owner_id: Optional[str] = typer.Option(None, help="主人ID"),
    appointment_date: Optional[str] = typer.Option(None, help="日期 YYYY-MM-DD"),
):
    appt_date = date.fromisoformat(appointment_date) if appointment_date else None
    appointments = client.list_appointments(
        owner_id=owner_id, appointment_date=appt_date
    )
    table = Table(title="预约列表")
    table.add_column("ID", style="cyan")
    table.add_column("宠物ID", style="magenta")
    table.add_column("日期")
    table.add_column("时间")
    table.add_column("状态")
    table.add_column("候补位置")
    for appt in appointments:
        table.add_row(
            appt.get("id", ""),
            appt.get("pet_id", ""),
            appt.get("appointment_date", ""),
            f"{appt.get('start_time', '')} - {appt.get('end_time', '')}",
            appt.get("status", ""),
            str(appt.get("waitlist_position") or ""),
        )
    console.print(table)


@app.command("cancel-appointment")
def cancel_appointment(
    appointment_id: str = typer.Option(..., help="预约ID"),
):
    result = client.cancel_appointment(appointment_id=appointment_id)
    console.print(f"[green]取消预约成功[/green]")
    console.print(result)


@app.command("my-todos")
def my_todos(
    owner_id: str = typer.Option(..., help="主人ID"),
):
    todos = client.get_pending_todos(owner_id=owner_id)
    if not todos:
        console.print("[green]没有待处理的用药提醒[/green]")
        return
    table = Table(title="用药待办")
    table.add_column("ID", style="cyan")
    table.add_column("药品", style="magenta")
    table.add_column("剂量")
    table.add_column("截止日期")
    table.add_column("状态")
    for todo in todos:
        table.add_row(
            todo.get("id", ""),
            todo.get("medicine_name", ""),
            todo.get("dosage", ""),
            todo.get("due_date", ""),
            todo.get("status", ""),
        )
    console.print(table)


@app.command("complete-todo")
def complete_todo(
    todo_id: str = typer.Option(..., help="待办ID"),
):
    result = client.complete_todo(todo_id=todo_id)
    console.print(f"[green]待办已完成[/green]")
    console.print(result)


@app.command("todos-by-date")
def todos_by_date(
    target_date: str = typer.Option(..., help="日期 YYYY-MM-DD"),
):
    dt = date.fromisoformat(target_date)
    todos = client.get_todos_by_date(target_date=dt)
    table = Table(title=f"{target_date} 用药待办")
    table.add_column("ID", style="cyan")
    table.add_column("宠物ID", style="magenta")
    table.add_column("药品")
    table.add_column("剂量")
    table.add_column("状态")
    for todo in todos:
        table.add_row(
            todo.get("id", ""),
            todo.get("pet_id", ""),
            todo.get("medicine_name", ""),
            todo.get("dosage", ""),
            todo.get("status", ""),
        )
    console.print(table)


def main():
    app()


if __name__ == "__main__":
    main()
