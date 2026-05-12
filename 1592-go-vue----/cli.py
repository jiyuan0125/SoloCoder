import click
import sqlite3
import os
from datetime import datetime, date, timedelta
from tabulate import tabulate

DB_PATH = os.path.join(os.path.dirname(__file__), "exhibition.db")


def get_connection():
    return sqlite3.connect(DB_PATH)


@click.group()
def cli():
    pass


@cli.group()
def booths():
    pass


@booths.command(name="list")
@click.option("--hall-id", type=int, help="展厅ID")
@click.option("--status", help="状态: available/selected/occupied")
def list_booths(hall_id, status):
    conn = get_connection()
    cursor = conn.cursor()

    query = "SELECT id, hall_id, booth_number, area, is_special, base_price, status FROM booths"
    params = []
    conditions = []

    if hall_id:
        conditions.append("hall_id = ?")
        params.append(hall_id)
    if status:
        conditions.append("status = ?")
        params.append(status)

    if conditions:
        query += " WHERE " + " AND ".join(conditions)

    cursor.execute(query, params)
    rows = cursor.fetchall()

    data = []
    for row in rows:
        data.append([
            row[0], row[1], row[2], f"{row[3]}㎡",
            "是" if row[4] else "否", f"¥{row[5]}", row[6]
        ])

    click.echo(tabulate(data, headers=["ID", "展厅ID", "展位号", "面积", "特装", "价格", "状态"]))
    conn.close()


@booths.command(name="show")
@click.argument("booth_id", type=int)
def show_booth(booth_id):
    conn = get_connection()
    cursor = conn.cursor()

    cursor.execute("""
        SELECT b.id, b.hall_id, b.booth_number, b.area, b.is_special, 
               b.base_price, b.status, b.position, h.name as hall_name
        FROM booths b
        JOIN halls h ON b.hall_id = h.id
        WHERE b.id = ?
    """, (booth_id,))
    row = cursor.fetchone()

    if row:
        click.echo(f"展位 ID: {row[0]}")
        click.echo(f"展厅: {row[8]} (ID: {row[1]})")
        click.echo(f"展位号: {row[2]}")
        click.echo(f"面积: {row[3]}㎡")
        click.echo(f"特装: {'是' if row[4] else '否'}")
        click.echo(f"价格: ¥{row[5]}")
        click.echo(f"状态: {row[6]}")
        click.echo(f"位置: {row[7] or 'N/A'}")
    else:
        click.echo("展位未找到")

    conn.close()


@booths.command(name="available")
@click.argument("exhibition_id", type=int)
@click.option("--hall-id", type=int, help="展厅ID")
def list_available_booths(exhibition_id, hall_id):
    conn = get_connection()
    cursor = conn.cursor()

    cursor.execute("""
        SELECT booth_id FROM booth_selections 
        WHERE exhibition_id = ? AND status IN ('selected', 'confirmed')
    """, (exhibition_id,))
    occupied = set(r[0] for r in cursor.fetchall())

    query = """
        SELECT id, hall_id, booth_number, area, is_special, base_price 
        FROM booths WHERE status = 'available'
    """
    params = []
    if hall_id:
        query += " AND hall_id = ?"
        params.append(hall_id)

    cursor.execute(query, params)
    rows = cursor.fetchall()

    data = []
    for row in rows:
        if row[0] not in occupied:
            data.append([
                row[0], row[1], row[2], f"{row[3]}㎡",
                "是" if row[4] else "否", f"¥{row[5]}"
            ])

    click.echo(tabulate(data, headers=["ID", "展厅ID", "展位号", "面积", "特装", "价格"]))
    conn.close()


@cli.group()
def setup():
    pass


@setup.command(name="list")
@click.option("--hall-id", type=int, required=True, help="展厅ID")
@click.option("--date", help="日期 (YYYY-MM-DD)")
def list_setup(hall_id, date):
    conn = get_connection()
    cursor = conn.cursor()

    query = """
        SELECT ss.id, ss.setup_date, ss.start_time, ss.end_time, 
               ss.actual_hours, ss.billed_hours, bs.exhibitor_name,
               b.booth_number
        FROM setup_schedules ss
        JOIN booth_selections bs ON ss.selection_id = bs.id
        JOIN booths b ON bs.booth_id = b.id
        WHERE ss.hall_id = ?
    """
    params = [hall_id]

    if date:
        query += " AND ss.setup_date = ?"
        params.append(date)

    query += " ORDER BY ss.setup_date, ss.start_time"

    cursor.execute(query, params)
    rows = cursor.fetchall()

    data = []
    for row in rows:
        data.append([
            row[0], row[1], f"{row[2]}-{row[3]}",
            f"{row[4]:.2f}h" if row[4] else "N/A",
            f"{row[5]}h" if row[5] else "N/A",
            row[6], row[7]
        ])

    click.echo(tabulate(
        data,
        headers=["ID", "日期", "时段", "实际小时", "计费小时", "参展商", "展位"]
    ))
    conn.close()


@setup.command(name="show")
@click.argument("schedule_id", type=int)
def show_setup(schedule_id):
    conn = get_connection()
    cursor = conn.cursor()

    cursor.execute("""
        SELECT ss.id, ss.hall_id, ss.setup_date, ss.start_time, ss.end_time,
               ss.actual_hours, ss.billed_hours, bs.exhibitor_name,
               b.booth_number, b.area, b.is_special
        FROM setup_schedules ss
        JOIN booth_selections bs ON ss.selection_id = bs.id
        JOIN booths b ON bs.booth_id = b.id
        WHERE ss.id = ?
    """, (schedule_id,))
    row = cursor.fetchone()

    if row:
        click.echo(f"布展安排 ID: {row[0]}")
        click.echo(f"展厅 ID: {row[1]}")
        click.echo(f"日期: {row[2]}")
        click.echo(f"时段: {row[3]} - {row[4]}")
        click.echo(f"实际时长: {row[5]:.2f}小时" if row[5] else "实际时长: N/A")
        click.echo(f"计费时长: {row[6]}小时" if row[6] else "计费时长: N/A")
        click.echo(f"参展商: {row[7]}")
        click.echo(f"展位: {row[8]} ({row[9]}㎡)")
        click.echo(f"特装: {'是' if row[10] else '否'}")
    else:
        click.echo("布展安排未找到")

    conn.close()


@cli.group()
def exhibitions():
    pass


@exhibitions.command(name="list")
def list_exhibitions():
    conn = get_connection()
    cursor = conn.cursor()

    cursor.execute("""
        SELECT id, name, organizer, start_date, end_date
        FROM exhibitions ORDER BY start_date
    """)
    rows = cursor.fetchall()

    data = []
    for row in rows:
        data.append([row[0], row[1], row[2] or "N/A", row[3], row[4]])

    click.echo(tabulate(data, headers=["ID", "名称", "主办方", "开始", "结束"]))
    conn.close()


@exhibitions.command(name="halls")
@click.argument("exhibition_id", type=int)
def list_exhibition_halls(exhibition_id):
    conn = get_connection()
    cursor = conn.cursor()

    cursor.execute("""
        SELECT DISTINCT h.id, h.name, h.floor, h.total_area
        FROM exhibition_halls eh
        JOIN halls h ON eh.hall_id = h.id
        WHERE eh.exhibition_id = ?
    """, (exhibition_id,))
    rows = cursor.fetchall()

    data = []
    for row in rows:
        data.append([row[0], row[1], row[2] or "N/A", f"{row[3] or 0}㎡"])

    click.echo(tabulate(data, headers=["ID", "展厅名称", "楼层", "面积"]))
    conn.close()


@cli.command()
@click.option("--days", type=int, default=7, help="显示未来N天")
def overview(days):
    conn = get_connection()
    cursor = conn.cursor()

    click.echo("=" * 60)
    click.echo("会展中心概览")
    click.echo("=" * 60)

    cursor.execute("SELECT COUNT(*) FROM venues")
    venues_count = cursor.fetchone()[0]

    cursor.execute("SELECT COUNT(*) FROM halls")
    halls_count = cursor.fetchone()[0]

    cursor.execute("SELECT COUNT(*) FROM booths")
    booths_count = cursor.fetchone()[0]

    cursor.execute("SELECT COUNT(*) FROM booths WHERE status = 'available'")
    available_count = cursor.fetchone()[0]

    click.echo(f"\n展馆: {venues_count} 个")
    click.echo(f"展厅: {halls_count} 个")
    click.echo(f"展位: {booths_count} 个 (可用: {available_count})")

    today = date.today()
    end_date = today + timedelta(days=days)

    cursor.execute("""
        SELECT id, name, start_date, end_date FROM exhibitions
        WHERE start_date <= ? AND end_date >= ?
        ORDER BY start_date
    """, (end_date.isoformat(), today.isoformat()))

    active_exhibitions = cursor.fetchall()

    click.echo(f"\n进行中/即将开始的展览 ({days}天内):")
    if active_exhibitions:
        data = []
        for row in active_exhibitions:
            data.append([row[0], row[1], row[2], row[3]])
        click.echo(tabulate(data, headers=["ID", "名称", "开始", "结束"]))
    else:
        click.echo("  无")

    conn.close()


if __name__ == "__main__":
    cli()
