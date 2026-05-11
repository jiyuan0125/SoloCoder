import click
import json
from src.client.api_client import api_client

def print_json(data):
    click.echo(json.dumps(data, indent=2, ensure_ascii=False))

@click.group()
def cli():
    """养殖场数字化管理系统命令行客户端"""
    pass

@cli.group()
def batch():
    """批次管理"""
    pass

@batch.command("create")
@click.option("--breed", required=True, help="品种名称")
@click.option("--count", required=True, type=int, help="入栏数量")
@click.option("--gestation", default=114, type=int, help="妊娠周期天数 (默认: 114)")
def create_batch(breed, count, gestation):
    """创建新批次"""
    try:
        result = api_client.create_batch(breed, count, gestation)
        click.echo(f"批次创建成功，ID: {result['id']}")
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@batch.command("list")
def list_batches():
    """列出所有批次"""
    try:
        batches = api_client.list_batches()
        if not batches:
            click.echo("暂无批次")
            return
        for batch in batches:
            click.echo(f"ID: {batch['id']}, 品种: {batch['breed']}, 入栏: {batch['initial_count']}, "
                       f"存栏: {batch['current_count']}, 状态: {batch['status']}")
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@batch.command("get")
@click.argument("batch_id", type=int)
def get_batch(batch_id):
    """获取批次详情"""
    try:
        batch = api_client.get_batch(batch_id)
        print_json(batch)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@cli.group()
def breeding():
    """繁育管理"""
    pass

@breeding.command("create")
@click.option("--batch-id", required=True, type=int, help="批次ID")
@click.option("--female-id", required=True, help="母畜编号")
@click.option("--date", required=True, help="配种日期 (格式: YYYY-MM-DD)")
def create_breeding(batch_id, female_id, date):
    """记录配种信息"""
    try:
        result = api_client.create_breeding_record(batch_id, female_id, date)
        click.echo(f"配种记录创建成功，ID: {result['id']}")
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@breeding.command("list")
@click.option("--batch-id", type=int, help="按批次筛选")
def list_breeding(batch_id):
    """列出配种记录"""
    try:
        records = api_client.list_breeding_records(batch_id)
        if not records:
            click.echo("暂无配种记录")
            return
        for record in records:
            exp_date = record.get('expected_birth_date', 'N/A')
            click.echo(f"ID: {record['id']}, 批次: {record['batch_id']}, 母畜: {record['female_id']}, "
                       f"配种日期: {record['breeding_date']}, 预产期: {exp_date}, 状态: {record['status']}")
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@breeding.command("fail")
@click.argument("record_id", type=int)
@click.option("--reason", required=True, help="失败原因")
def breeding_fail(record_id, reason):
    """记录配种失败"""
    try:
        result = api_client.record_breeding_failure(record_id, reason)
        click.echo(f"配种失败记录已更新")
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@breeding.command("success")
@click.argument("record_id", type=int)
def breeding_success(record_id):
    """记录配种成功"""
    try:
        result = api_client.record_breeding_success(record_id)
        click.echo(f"配种成功记录已更新")
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@cli.group()
def vaccination():
    """防疫管理"""
    pass

@vaccination.command("create")
@click.option("--batch-id", required=True, type=int, help="批次ID")
@click.option("--vaccine", required=True, help="疫苗名称")
@click.option("--date", required=True, help="接种日期 (格式: YYYY-MM-DD)")
@click.option("--interval", type=int, help="下次接种间隔天数")
def create_vaccination(batch_id, vaccine, date, interval):
    """记录疫苗接种"""
    try:
        result = api_client.create_vaccination_record(batch_id, vaccine, date, interval)
        click.echo(f"疫苗接种记录创建成功，ID: {result['id']}")
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@vaccination.command("list")
@click.option("--batch-id", type=int, help="按批次筛选")
def list_vaccination(batch_id):
    """列出接种记录"""
    try:
        records = api_client.list_vaccination_records(batch_id)
        if not records:
            click.echo("暂无接种记录")
            return
        for record in records:
            next_due = record.get('next_due_date', 'N/A')
            click.echo(f"ID: {record['id']}, 批次: {record['batch_id']}, 疫苗: {record['vaccine_name']}, "
                       f"接种日期: {record['vaccination_date']}, 下次接种: {next_due}")
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@cli.group()
def slaughter():
    """出栏管理"""
    pass

@slaughter.command("create")
@click.option("--batch-id", required=True, type=int, help="批次ID")
@click.option("--count", required=True, type=int, help="出栏数量")
@click.option("--weight", required=True, type=float, help="平均体重(kg)")
@click.option("--price", required=True, type=float, help="单价(元/kg)")
@click.option("--date", required=True, help="出栏日期 (格式: YYYY-MM-DD)")
def create_slaughter(batch_id, count, weight, price, date):
    """记录出栏信息"""
    try:
        result = api_client.create_slaughter_record(batch_id, count, weight, price, date)
        click.echo(f"出栏记录创建成功，ID: {result['id']}")
        if price > 0:
            value = count * weight * price
            click.echo(f"本次产值: {value:.2f} 元")
        else:
            click.echo("单价为零或负数，不参与产值计算")
        print_json(result)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@slaughter.command("list")
@click.option("--batch-id", type=int, help="按批次筛选")
def list_slaughter(batch_id):
    """列出出栏记录"""
    try:
        records = api_client.list_slaughter_records(batch_id)
        if not records:
            click.echo("暂无出栏记录")
            return
        for record in records:
            click.echo(f"ID: {record['id']}, 批次: {record['batch_id']}, 数量: {record['count']}, "
                       f"体重: {record['avg_weight']}kg, 单价: {record['unit_price']}元/kg, "
                       f"日期: {record['slaughter_date']}")
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@cli.group()
def reminder():
    """提醒管理"""
    pass

@reminder.command("generate")
def generate_reminders():
    """生成待办提醒"""
    try:
        reminders = api_client.generate_reminders()
        if reminders:
            click.echo(f"生成了 {len(reminders)} 条新提醒:")
            for reminder in reminders:
                click.echo(f"  [{reminder['type']}] {reminder['message']}")
        else:
            click.echo("无新提醒")
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@reminder.command("list")
@click.option("--unread", is_flag=True, help="只显示未读提醒")
def list_reminders(unread):
    """列出提醒"""
    try:
        if unread:
            reminders = api_client.list_reminders(is_read=False)
        else:
            reminders = api_client.list_reminders()
        if not reminders:
            click.echo("暂无提醒")
            return
        for reminder in reminders:
            status = "已读" if reminder['is_read'] else "未读"
            click.echo(f"ID: {reminder['id']} [{status}] [{reminder['type']}] "
                       f"{reminder['message']} (截止: {reminder['due_date']})")
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@reminder.command("read")
@click.argument("reminder_id", type=int)
def mark_reminder_read(reminder_id):
    """标记提醒为已读"""
    try:
        result = api_client.mark_reminder_read(reminder_id)
        click.echo("提醒已标记为已读")
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

@cli.command()
def dashboard():
    """查看指标看板"""
    try:
        data = api_client.get_dashboard()
        click.echo("=" * 40)
        click.echo("养殖场数字化管理系统 - 指标看板")
        click.echo("=" * 40)
        click.echo(f"总存栏数: {data['total_stock']}")
        click.echo(f"本月入栏: {data['monthly_input_count']}")
        click.echo(f"本月出栏: {data['monthly_slaughter_count']}")
        click.echo(f"累计产值: {data['total_output_value']:.2f} 元")
        click.echo()
        click.echo("品种占比:")
        if data['breed_distribution']:
            total = data['total_stock']
            for breed, count in data['breed_distribution'].items():
                percentage = (count / total * 100) if total > 0 else 0
                click.echo(f"  {breed}: {count} ({percentage:.1f}%)")
        else:
            click.echo("  暂无数据")
        click.echo("=" * 40)
    except Exception as e:
        click.echo(f"错误: {e}", err=True)

if __name__ == "__main__":
    cli()
