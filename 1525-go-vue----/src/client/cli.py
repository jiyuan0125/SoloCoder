import click
import json
from datetime import date
from typing import Optional

from .api_client import ApiClient


CERT_TYPES = ["mining_license", "safety_license", "env_approval", "other"]
STATUS_TYPES = ["active", "cancelled", "expired", "about_to_expire"]
TODO_TYPES = ["certificate_expiry", "annual_inspection", "compliance_rectification"]
TODO_STATUSES = ["pending", "in_progress", "completed"]


def get_client() -> ApiClient:
    return ApiClient()


def format_json(data) -> str:
    return json.dumps(data, ensure_ascii=False, indent=2)


def print_result(data) -> None:
    click.echo(format_json(data))


def status_color(status: str) -> str:
    color_map = {
        "active": "green",
        "cancelled": "yellow",
        "expired": "red",
        "about_to_expire": "bright_yellow",
        "pending": "yellow",
        "in_progress": "blue",
        "completed": "green",
    }
    return color_map.get(status, "white")


def type_label(cert_type: str) -> str:
    labels = {
        "mining_license": "采矿许可证",
        "safety_license": "安全生产许可证",
        "env_approval": "环评批复",
        "other": "其他",
    }
    return labels.get(cert_type, cert_type)


@click.group()
def main() -> None:
    """矿业行政审批管理系统命令行客户端"""
    pass


@main.group()
def certificate() -> None:
    """证照管理命令"""
    pass


@certificate.command("list")
@click.option("--status", "-s", type=click.Choice(STATUS_TYPES), help="按状态筛选")
@click.option("--type", "-t", "cert_type", type=click.Choice(CERT_TYPES), help="按类型筛选")
def certificate_list(status: Optional[str], cert_type: Optional[str]) -> None:
    """列出所有证照"""
    client = get_client()
    try:
        certs = client.list_certificates(status=status, cert_type=cert_type)
        if not certs:
            click.echo("没有找到证照记录")
            return
        
        click.echo(f"共找到 {len(certs)} 条证照记录:\n")
        for cert in certs:
            status_str = click.style(cert["status"], fg=status_color(cert["status"]))
            click.echo(
                f"ID: {cert['id']} | 名称: {cert['name']} | 编号: {cert['number']}\n"
                f"  类型: {type_label(cert['type'])} | 状态: {status_str}\n"
                f"  发证日期: {cert['issuing_date']} | 有效期至: {cert['expiry_date']}\n"
            )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@certificate.command("get")
@click.argument("cert_id", type=int)
def certificate_get(cert_id: int) -> None:
    """查看证照详情"""
    client = get_client()
    try:
        cert = client.get_certificate(cert_id)
        status_str = click.style(cert["status"], fg=status_color(cert["status"]))
        click.echo(
            f"证照详情:\n"
            f"  ID: {cert['id']}\n"
            f"  名称: {cert['name']}\n"
            f"  类型: {type_label(cert['type'])}\n"
            f"  编号: {cert['number']}\n"
            f"  状态: {status_str}\n"
            f"  发证日期: {cert['issuing_date']}\n"
            f"  有效期至: {cert['expiry_date']}\n"
        )
        if cert["remarks"]:
            click.echo(f"  备注: {cert['remarks']}")
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@certificate.command("create")
@click.option("--name", "-n", required=True, help="证照名称")
@click.option("--type", "-t", required=True, type=click.Choice(CERT_TYPES), help="证照类型")
@click.option("--number", "-num", required=True, help="证照编号")
@click.option("--issuing-date", "-id", required=True, type=click.DateTime(formats=["%Y-%m-%d"]), help="发证日期 YYYY-MM-DD")
@click.option("--expiry-date", "-ed", required=True, type=click.DateTime(formats=["%Y-%m-%d"]), help="有效期至 YYYY-MM-DD")
@click.option("--remarks", "-r", help="备注")
def certificate_create(name: str, type: str, number: str, issuing_date, expiry_date, remarks: Optional[str]) -> None:
    """创建新证照"""
    client = get_client()
    try:
        cert = client.create_certificate(
            name=name,
            cert_type=type,
            number=number,
            issuing_date=issuing_date.date(),
            expiry_date=expiry_date.date(),
            remarks=remarks,
        )
        click.echo(f"证照创建成功！ID: {cert['id']}")
        status_str = click.style(cert["status"], fg=status_color(cert["status"]))
        click.echo(f"名称: {cert['name']} | 状态: {status_str}")
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@certificate.command("cancel")
@click.argument("cert_id", type=int)
def certificate_cancel(cert_id: int) -> None:
    """注销证照"""
    client = get_client()
    try:
        cert = client.cancel_certificate(cert_id)
        click.echo(f"证照已注销！ID: {cert['id']}")
        status_str = click.style(cert["status"], fg=status_color(cert["status"]))
        click.echo(f"名称: {cert['name']} | 状态: {status_str}")
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@main.group()
def inspection() -> None:
    """年检管理命令"""
    pass


@inspection.command("list")
@click.option("--certificate-id", "-c", type=int, help="按证照ID筛选")
def inspection_list(certificate_id: Optional[int]) -> None:
    """列出年检记录"""
    client = get_client()
    try:
        inspections = client.list_inspections(certificate_id=certificate_id)
        if not inspections:
            click.echo("没有找到年检记录")
            return
        
        click.echo(f"共找到 {len(inspections)} 条年检记录:\n")
        for ins in inspections:
            result_str = click.style("通过", fg="green") if ins["result"] else click.style("未通过", fg="red")
            click.echo(
                f"ID: {ins['id']} | 证照ID: {ins['certificate_id']} | 年度: {ins['year']}\n"
                f"  检查日期: {ins['inspection_date']} | 结果: {result_str}\n"
            )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@inspection.command("create")
@click.option("--certificate-id", "-c", required=True, type=int, help="证照ID")
@click.option("--year", "-y", required=True, type=int, help="年度")
@click.option("--date", "-d", "inspection_date", required=True, type=click.DateTime(formats=["%Y-%m-%d"]), help="检查日期 YYYY-MM-DD")
@click.option("--result/--no-result", "-r/-nr", default=True, help="检查结果（通过/不通过）")
@click.option("--remarks", "-rm", help="备注")
def inspection_create(certificate_id: int, year: int, inspection_date, result: bool, remarks: Optional[str]) -> None:
    """创建年检记录"""
    client = get_client()
    try:
        ins = client.create_inspection(
            certificate_id=certificate_id,
            year=year,
            inspection_date=inspection_date.date(),
            result=result,
            remarks=remarks,
        )
        result_str = click.style("通过", fg="green") if ins["result"] else click.style("未通过", fg="red")
        click.echo(
            f"年检记录创建成功！ID: {ins['id']}\n"
            f"年度: {ins['year']} | 检查日期: {ins['inspection_date']} | 结果: {result_str}"
        )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@main.group()
def compliance() -> None:
    """合规检查命令"""
    pass


@compliance.command("list")
@click.option("--certificate-id", "-c", type=int, help="按证照ID筛选")
def compliance_list(certificate_id: Optional[int]) -> None:
    """列出合规检查记录"""
    client = get_client()
    try:
        checks = client.list_compliance_checks(certificate_id=certificate_id)
        if not checks:
            click.echo("没有找到合规检查记录")
            return
        
        click.echo(f"共找到 {len(checks)} 条合规检查记录:\n")
        for check in checks:
            compliant_str = click.style("合规", fg="green") if check["is_compliant"] else click.style("不合规", fg="red")
            safety_str = click.style("有安全隐患", fg="bright_red") if check["has_safety_issues"] else ""
            click.echo(
                f"ID: {check['id']} | 证照ID: {check['certificate_id']}\n"
                f"  检查日期: {check['check_date']} | 结果: {compliant_str} {safety_str}\n"
                f"  检查项目: {check['check_items']}\n"
            )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@compliance.command("create")
@click.option("--certificate-id", "-c", required=True, type=int, help="证照ID")
@click.option("--date", "-d", required=True, type=click.DateTime(formats=["%Y-%m-%d"]), help="检查日期 YYYY-MM-DD")
@click.option("--items", "-i", required=True, help="检查项目")
@click.option("--compliant/--no-compliant", "-co/-nco", default=True, help="是否合规")
@click.option("--safety-issues/--no-safety-issues", "-s/-ns", default=False, help="是否有安全隐患")
@click.option("--remarks", "-r", help="备注")
def compliance_create(certificate_id: int, date, items: str, compliant: bool, safety_issues: bool, remarks: Optional[str]) -> None:
    """创建合规检查记录"""
    client = get_client()
    try:
        check = client.create_compliance_check(
            certificate_id=certificate_id,
            check_date=date.date(),
            check_items=items,
            is_compliant=compliant,
            has_safety_issues=safety_issues,
            remarks=remarks,
        )
        compliant_str = click.style("合规", fg="green") if check["is_compliant"] else click.style("不合规", fg="red")
        safety_str = click.style("有安全隐患", fg="bright_red") if check["has_safety_issues"] else ""
        click.echo(
            f"合规检查记录创建成功！ID: {check['id']}\n"
            f"检查日期: {check['check_date']} | 结果: {compliant_str} {safety_str}"
        )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@main.group()
def todo() -> None:
    """待办事项命令"""
    pass


@todo.command("list")
@click.option("--status", "-s", type=click.Choice(TODO_STATUSES), help="按状态筛选")
@click.option("--type", "-t", "todo_type", type=click.Choice(TODO_TYPES), help="按类型筛选")
def todo_list(status: Optional[str], todo_type: Optional[str]) -> None:
    """列出待办事项"""
    client = get_client()
    try:
        todos = client.list_todos(status=status, todo_type=todo_type)
        if not todos:
            click.echo("没有找到待办事项")
            return
        
        click.echo(f"共找到 {len(todos)} 条待办事项:\n")
        for todo in todos:
            status_str = click.style(todo["status"], fg=status_color(todo["status"]))
            click.echo(
                f"ID: {todo['id']} | 类型: {todo['type']}\n"
                f"  截止日期: {todo['due_date']} | 状态: {status_str}\n"
                f"  描述: {todo['description']}\n"
            )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@todo.command("update")
@click.argument("todo_id", type=int)
@click.option("--status", "-s", required=True, type=click.Choice(TODO_STATUSES), help="新状态")
def todo_update(todo_id: int, status: str) -> None:
    """更新待办事项状态"""
    client = get_client()
    try:
        todo = client.update_todo_status(todo_id=todo_id, status=status)
        status_str = click.style(todo["status"], fg=status_color(todo["status"]))
        click.echo(
            f"待办事项更新成功！ID: {todo['id']}\n"
            f"新状态: {status_str}"
        )
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


@todo.command("process")
def todo_process() -> None:
    """处理到期提醒，自动生成待办"""
    client = get_client()
    try:
        result = client.process_reminders()
        click.echo(result["message"])
    except Exception as e:
        click.echo(f"错误: {e}", err=True)
    finally:
        client.close()


if __name__ == "__main__":
    main()
