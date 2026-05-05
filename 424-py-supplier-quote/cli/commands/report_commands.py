from uuid import UUID

import typer

from shared.protocols.quote import QuoteComparisonReport
from cli.client.http_client import get_default_client

app = typer.Typer(help="报表管理命令")


@app.command(name="comparison")
def comparison_report(
    purchase_id: UUID = typer.Argument(..., help="采购需求 ID"),
) -> None:
    """生成比价报表（按商品列出所有供应商报价的排名和价格分布）"""
    client = get_default_client()
    try:
        response = client.get(f"/reports/comparison/{purchase_id}")
        report = client.parse_response(response, QuoteComparisonReport)

        _print_comparison_report(report)
    except Exception as e:
        typer.secho(f"✗ 生成报表失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


def _print_comparison_report(report: QuoteComparisonReport) -> None:
    typer.secho("=" * 60, fg=typer.colors.BLUE)
    typer.secho("                   比价报表", fg=typer.colors.BLUE, bold=True)
    typer.secho("=" * 60, fg=typer.colors.BLUE)

    typer.echo(f"\n采购需求: {report.purchase_title}")
    typer.echo(f"商品名称: {report.product_name}")

    typer.echo(f"\n价格统计:")
    typer.echo(f"  平均价格: {report.avg_price}")
    typer.echo(f"  最低价格: {report.min_price}")
    typer.echo(f"  最高价格: {report.max_price}")

    typer.echo(f"\n价格分布:")
    for range_str, count in report.price_distribution.items():
        typer.echo(f"  {range_str}: {count} 个报价")

    typer.echo(f"\n供应商报价排名（按资质等级权重和价格综合排序）:")
    typer.echo("-" * 60)

    header = (
        f"{'排名':<4} {'供应商':<20} {'资质':<4} {'单价':<12} "
        f"{'交货天数':<8} {'迟到':<4} {'状态':<8}"
    )
    typer.secho(header, fg=typer.colors.CYAN)
    typer.echo("-" * 60)

    for quote in sorted(report.quotes, key=lambda q: q.rank):
        rank_mark = "★" if quote.is_awarded else ""
        late_mark = "是" if quote.is_late else "否"
        status = "已中标" if quote.is_awarded else ""

        line = (
            f"{quote.rank:<4} {quote.supplier_name:<20} "
            f"{quote.qualification_level.value:<4} {quote.unit_price:<12} "
            f"{quote.delivery_days:<8} {late_mark:<4} {status:<8}"
        )

        if quote.is_awarded:
            typer.secho(f"{line} {rank_mark}", fg=typer.colors.GREEN)
        elif quote.is_late:
            typer.secho(line, fg=typer.colors.YELLOW)
        else:
            typer.echo(line)

    typer.echo("-" * 60)
    typer.echo("\n说明:")
    typer.echo("  ★ 表示已中标")
    typer.echo("  资质等级权重: A(3分) > B(2分) > C(1分)")
    typer.echo("  排序规则: 先按资质等级降序，再按价格升序，最后按交货天数升序")
