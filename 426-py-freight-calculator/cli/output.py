from typing import Any, Optional
from decimal import Decimal

from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich import print as rprint


console = Console()


def format_money(value: Any) -> str:
    if isinstance(value, str):
        try:
            value = Decimal(value)
        except ValueError:
            return str(value)
    if isinstance(value, Decimal):
        return f"¥{value:,.2f}"
    return str(value)


def print_health_check(result: dict[str, Any]) -> None:
    status = result.get("status", "unknown")
    service = result.get("service", "unknown")
    version = result.get("version", "unknown")
    
    color = "green" if status == "healthy" else "red"
    
    console.print(Panel(
        f"[{color} bold]Status: {status}[/]\n"
        f"Service: {service}\n"
        f"Version: {version}",
        title="Health Check",
        border_style="blue",
    ))


def print_calculation_result(result: dict[str, Any]) -> None:
    if not result.get("success", False):
        error = result.get("error", {})
        console.print(f"[bold red]错误:[/] {error.get('message', '未知错误')}")
        return
    
    data = result.get("data", {})
    is_cached = data.get("is_cached", False)
    request_id = data.get("request_id", "N/A")
    
    table = Table(title=f"运费计算结果 [{'yellow' if is_cached else 'green'}]{{'缓存' if is_cached else '实时'}}[/]", border_style="blue")
    table.add_column("项目", style="cyan", no_wrap=True)
    table.add_column("金额", style="magenta", justify="right")
    
    table.add_row("基础运费", format_money(data.get("total_base_freight", 0)))
    table.add_row("附加费", format_money(data.get("total_surcharge", 0)))
    table.add_row("保价费", format_money(data.get("total_insurance_fee", 0)))
    table.add_row("折扣", f"-{format_money(data.get('total_discount', 0))}")
    table.add_row("最终金额", format_money(data.get("total_final_amount", 0)), style="bold green")
    
    console.print(table)
    console.print(f"[dim]请求ID: {request_id}[/]")
    
    package_details = data.get("package_details", [])
    if package_details:
        pkg_table = Table(title="包裹明细", border_style="cyan")
        pkg_table.add_column("#", style="cyan")
        pkg_table.add_column("基础运费", style="white")
        pkg_table.add_column("附加费", style="yellow")
        pkg_table.add_column("保价费", style="blue")
        pkg_table.add_column("折扣", style="red")
        pkg_table.add_column("最终金额", style="green")
        pkg_table.add_column("最高赔付", style="magenta")
        
        for pkg in package_details:
            is_split = pkg.get("is_split", False)
            split_mark = " [yellow](拆分)[/]" if is_split else ""
            
            pkg_table.add_row(
                f"{pkg.get('package_index', 0) + 1}{split_mark}",
                format_money(pkg.get("base_freight", 0)),
                format_money(pkg.get("surcharge", 0)),
                format_money(pkg.get("insurance_fee", 0)),
                f"-{format_money(pkg.get('discount', 0))}",
                format_money(pkg.get("final_amount", 0)),
                format_money(pkg.get("max_compensation", 0)),
            )
            
            if is_split and pkg.get("split_packages"):
                for sp in pkg["split_packages"]:
                    pkg_table.add_row(
                        "  └─子包",
                        format_money(sp.get("base_freight", 0)),
                        "",
                        "",
                        "",
                        format_money(sp.get("total_freight", 0)),
                        "",
                        style="dim",
                    )
        
        console.print(pkg_table)


def print_statistics(result: dict[str, Any]) -> None:
    if not result.get("success", False):
        console.print(f"[bold red]错误:[/] {result.get('message', '获取统计数据失败')}")
        return
    
    message = result.get("message", "")
    if "可用月份" in message:
        console.print(Panel(message, title="可用统计月份", border_style="blue"))
        return
    
    data = result.get("data", {})
    year_month = data.get("year_month", "N/A")
    
    table = Table(title=f"月度统计 - {year_month}", border_style="green")
    table.add_column("指标", style="cyan")
    table.add_column("数值", style="magenta", justify="right")
    
    table.add_row("包裹总数", str(data.get("total_packages", 0)))
    table.add_row("基础运费合计", format_money(data.get("total_base_freight", 0)))
    table.add_row("附加费合计", format_money(data.get("total_surcharge", 0)))
    table.add_row("保价费合计", format_money(data.get("total_insurance_fee", 0)))
    table.add_row("折扣合计", f"-{format_money(data.get('total_discount', 0))}")
    table.add_row("最终金额合计", format_money(data.get("total_final_amount", 0)), style="bold green")
    
    console.print(table)
    
    breakdown = data.get("transport_mode_breakdown", {})
    if breakdown:
        mode_table = Table(title="运输方式分布", border_style="yellow")
        mode_table.add_column("运输方式", style="cyan")
        mode_table.add_column("订单数", style="magenta", justify="right")
        
        for mode, count in breakdown.items():
            mode_name = {
                "land": "陆运",
                "air": "空运",
                "sea": "海运",
            }.get(mode, mode)
            mode_table.add_row(mode_name, str(count))
        
        console.print(mode_table)


def print_config(result: dict[str, Any]) -> None:
    if not result.get("success", False):
        console.print(f"[bold red]错误:[/] {result.get('message', '获取配置失败')}")
        return
    
    data = result.get("data", {})
    pricing = data.get("pricing", {})
    remote = data.get("remote_areas", {})
    discounts = data.get("discounts", {})
    
    pricing_table = Table(title="定价配置", border_style="blue")
    pricing_table.add_column("项目", style="cyan")
    pricing_table.add_column("值", style="magenta")
    
    pricing_table.add_row("陆运首重价格", f"{pricing.get('land_first_weight_price')} 元/公斤")
    pricing_table.add_row("陆运续重价格", f"{pricing.get('land_continue_weight_price')} 元/公斤")
    pricing_table.add_row("海运20尺柜", f"{pricing.get('sea_20ft_price')} 元")
    pricing_table.add_row("海运40尺柜", f"{pricing.get('sea_40ft_price')} 元")
    pricing_table.add_row("体积重量除数", str(pricing.get('air_volumetric_divisor')))
    pricing_table.add_row("陆运尺寸上限", f"{pricing.get('land_max_perimeter_meters')} 米")
    pricing_table.add_row("空运尺寸上限", f"{pricing.get('air_max_perimeter_meters')} 米")
    pricing_table.add_row("保价费率", f"{float(pricing.get('insurance_rate', 0)) * 100}%")
    pricing_table.add_row("保价最低费用", f"{pricing.get('insurance_min_fee')} 元")
    pricing_table.add_row("未保价赔付倍数", f"{pricing.get('max_compensation_multiplier')} 倍")
    
    console.print(pricing_table)
    
    remote_table = Table(title="偏远地区配置", border_style="yellow")
    remote_table.add_column("项目", style="cyan")
    remote_table.add_column("值", style="magenta")
    
    areas = remote.get("areas", [])
    remote_table.add_row("偏远地区列表", ", ".join(areas) if areas else "无")
    remote_table.add_row("附加费率", f"{float(remote.get('surcharge_rate', 0)) * 100}%")
    
    console.print(remote_table)
    
    discount_table = Table(title="折扣配置", border_style="green")
    discount_table.add_column("累计金额阈值", style="cyan")
    discount_table.add_column("折扣率", style="magenta")
    
    tiers = discounts.get("tiers", [])
    if not tiers:
        discount_table.add_row("无折扣配置", "")
    else:
        for tier in tiers:
            discount_table.add_row(
                f"≥ {format_money(tier.get('threshold', 0))}",
                f"{float(tier.get('discount_rate', 0)) * 100}%",
            )
    
    console.print(discount_table)


def print_cache_stats(result: dict[str, Any]) -> None:
    if not result.get("success", False):
        console.print(f"[bold red]错误:[/] {result.get('message', '获取缓存统计失败')}")
        return
    
    data = result.get("data", {})
    console.print(Panel(
        f"缓存项数量: {data.get('total_items', 0)}",
        title="缓存统计",
        border_style="cyan",
    ))


def print_operation_result(result: dict[str, Any], operation_name: str) -> None:
    if result.get("success", False):
        message = result.get("message", f"{operation_name}成功")
        console.print(f"[bold green]✓[/] {message}")
        if "data" in result and isinstance(result["data"], dict):
            for key, value in result["data"].items():
                console.print(f"  {key}: {value}")
    else:
        error = result.get("error", {})
        message = error.get("message", result.get("message", f"{operation_name}失败"))
        console.print(f"[bold red]✗[/] {message}")
