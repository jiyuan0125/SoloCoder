import click
from datetime import datetime

from app.database import SessionLocal, engine, Base
from app.models import Card, Transaction, Fault, CardType, CardStatus, FaultStatus, Priority


def init_db():
    Base.metadata.create_all(bind=engine)


def format_balance(cents: int) -> str:
    return f"¥{cents / 100:.2f}"


@click.group()
@click.version_option(version="1.0.0", prog_name="地铁AFC管理客户端")
def cli():
    init_db()


@cli.group()
def card():
    pass


@card.command("list")
@click.option("--limit", default=50, help="显示数量限制")
@click.option("--offset", default=0, help="偏移量")
def list_cards(limit, offset):
    db = SessionLocal()
    try:
        cards = db.query(Card).offset(offset).limit(limit).all()
        if not cards:
            click.echo("暂无票卡记录")
            return

        click.echo("-" * 80)
        click.echo(f"{'票卡号':<15} {'类型':<10} {'余额':<12} {'状态':<10} {'创建时间':<20}")
        click.echo("-" * 80)
        for c in cards:
            click.echo(
                f"{c.card_number:<15} {c.card_type.value:<10} {format_balance(c.balance):<12} "
                f"{c.status.value:<10} {c.created_at.strftime('%Y-%m-%d %H:%M'):<20}"
            )
        click.echo(f"\n共 {len(cards)} 张票卡")
    finally:
        db.close()


@card.command("show")
@click.argument("card_number")
def show_card(card_number):
    db = SessionLocal()
    try:
        c = db.query(Card).filter(Card.card_number == card_number).first()
        if not c:
            click.echo(f"票卡 {card_number} 不存在")
            return

        click.echo("=" * 60)
        click.echo("票卡详情")
        click.echo("=" * 60)
        click.echo(f"票卡号:    {c.card_number}")
        click.echo(f"类型:      {c.card_type.value}")
        click.echo(f"余额:      {format_balance(c.balance)}")
        click.echo(f"状态:      {c.status.value}")
        click.echo(f"创建时间:  {c.created_at.strftime('%Y-%m-%d %H:%M:%S')}")
        click.echo(f"更新时间:  {c.updated_at.strftime('%Y-%m-%d %H:%M:%S')}")
        click.echo("=" * 60)
    finally:
        db.close()


@card.command("create")
@click.argument("card_number")
@click.argument("card_type", type=click.Choice(["单程票", "储值票", "学生票", "老年票"]))
@click.option("--balance", default=0, type=int, help="初始余额(分)")
def create_card(card_number, card_type, balance):
    db = SessionLocal()
    try:
        existing = db.query(Card).filter(Card.card_number == card_number).first()
        if existing:
            click.echo(f"票卡 {card_number} 已存在")
            return

        type_map = {
            "单程票": CardType.SINGLE,
            "储值票": CardType.STORED,
            "学生票": CardType.STUDENT,
            "老年票": CardType.ELDERLY
        }

        c = Card(
            card_number=card_number,
            card_type=type_map[card_type],
            balance=balance,
            status=CardStatus.ACTIVE
        )
        db.add(c)
        db.commit()
        click.echo(f"票卡创建成功! 卡号: {card_number}, 类型: {card_type}")
    finally:
        db.close()


@card.command("recharge")
@click.argument("card_number")
@click.argument("amount", type=int)
def recharge_card(card_number, amount):
    db = SessionLocal()
    try:
        c = db.query(Card).filter(Card.card_number == card_number).first()
        if not c:
            click.echo(f"票卡 {card_number} 不存在")
            return

        if c.status != CardStatus.ACTIVE:
            click.echo(f"票卡状态为 {c.status.value}，无法充值")
            return

        amount_cents = amount * 100

        if amount_cents < 1000 or amount_cents > 50000:
            click.echo("充值金额必须在10-500元之间")
            return

        if amount_cents % 1000 != 0:
            click.echo("充值金额必须是10元的整数倍")
            return

        bonus = 0
        if c.card_type == CardType.STUDENT:
            bonus = round(amount_cents * 0.1)

        original = c.balance
        c.balance += (amount_cents + bonus)
        db.commit()

        click.echo(f"充值成功!")
        click.echo(f"原余额:  {format_balance(original)}")
        click.echo(f"充值:    ¥{amount:.2f}")
        if bonus > 0:
            click.echo(f"赠送:    {format_balance(bonus)}")
        click.echo(f"新余额:  {format_balance(c.balance)}")
    finally:
        db.close()


@cli.group()
def transaction():
    pass


@transaction.command("list")
@click.option("--card", help="按票卡号过滤")
@click.option("--limit", default=50, help="显示数量限制")
@click.option("--offset", default=0, help="偏移量")
def list_transactions(card, limit, offset):
    db = SessionLocal()
    try:
        query = db.query(Transaction)
        if card:
            query = query.filter(Transaction.card_number == card)

        txs = query.order_by(Transaction.id.desc()).offset(offset).limit(limit).all()

        if not txs:
            click.echo("暂无通行记录")
            return

        click.echo("-" * 120)
        click.echo(
            f"{'ID':<5} {'票卡号':<15} {'进站':<10} {'出站':<10} {'里程':<8} "
            f"{'基础票价':<10} {'折扣':<15} {'实际扣费':<10} {'状态':<10}"
        )
        click.echo("-" * 120)
        for tx in txs:
            status = "已完成" if tx.exit_time else "进行中"
            click.echo(
                f"{tx.id:<5} {tx.card_number:<15} {tx.entry_station or '-':<10} "
                f"{tx.exit_station or '-':<10} {tx.distance:<8.1f} "
                f"{format_balance(tx.base_fare):<10} {tx.discount:<15} "
                f"{format_balance(tx.final_fare):<10} {status:<10}"
            )
        click.echo(f"\n共 {len(txs)} 条记录")
    finally:
        db.close()


@transaction.command("show")
@click.argument("tx_id", type=int)
def show_transaction(tx_id):
    db = SessionLocal()
    try:
        tx = db.query(Transaction).filter(Transaction.id == tx_id).first()
        if not tx:
            click.echo(f"通行记录 {tx_id} 不存在")
            return

        click.echo("=" * 60)
        click.echo("通行记录详情")
        click.echo("=" * 60)
        click.echo(f"ID:         {tx.id}")
        click.echo(f"票卡号:     {tx.card_number}")
        click.echo(f"进站:       {tx.entry_station} ({tx.entry_time.strftime('%Y-%m-%d %H:%M:%S') if tx.entry_time else '-'})")
        click.echo(f"出站:       {tx.exit_station or '-'} ({tx.exit_time.strftime('%Y-%m-%d %H:%M:%S') if tx.exit_time else '-'})")
        click.echo(f"里程:       {tx.distance} 公里")
        click.echo(f"基础票价:   {format_balance(tx.base_fare)}")
        click.echo(f"折扣:       {tx.discount}")
        click.echo(f"实际扣费:   {format_balance(tx.final_fare)}")
        click.echo("=" * 60)
    finally:
        db.close()


@cli.group()
def fault():
    pass


@fault.command("list")
@click.option("--status", type=click.Choice(["待处理", "处理中", "已解决"]), help="按状态过滤")
@click.option("--limit", default=50, help="显示数量限制")
def list_faults(status, limit):
    db = SessionLocal()
    try:
        query = db.query(Fault)
        if status:
            status_map = {"待处理": FaultStatus.PENDING, "处理中": FaultStatus.PROCESSING, "已解决": FaultStatus.RESOLVED}
            query = query.filter(Fault.status == status_map[status])

        faults = query.order_by(Fault.id.desc()).limit(limit).all()

        if not faults:
            click.echo("暂无故障记录")
            return

        click.echo("-" * 100)
        click.echo(
            f"{'ID':<5} {'闸机':<10} {'车站':<12} {'优先级':<8} {'状态':<10} {'是否升级':<8} {'上报时间':<20}"
        )
        click.echo("-" * 100)
        for f in faults:
            escalated = "是" if f.escalated else "否"
            click.echo(
                f"{f.id:<5} {f.gate_id:<10} {f.station:<12} {f.priority.value:<8} "
                f"{f.status.value:<10} {escalated:<8} {f.reported_at.strftime('%Y-%m-%d %H:%M'):<20}"
            )
        click.echo(f"\n共 {len(faults)} 条故障记录")
    finally:
        db.close()


if __name__ == "__main__":
    cli()
