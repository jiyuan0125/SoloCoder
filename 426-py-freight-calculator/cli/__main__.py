from typing import Optional, List
from decimal import Decimal
import os

import typer
from rich.console import Console

from shared.models import TransportMode, ContainerType, Package
from cli.client import ApiClient
from cli.output import (
    print_health_check,
    print_calculation_result,
    print_statistics,
    print_config,
    print_cache_stats,
    print_operation_result,
)


app = typer.Typer(
    name="freight",
    help="运费计算命令行工具",
    add_completion=False,
)

console = Console()


def get_client(base_url: Optional[str] = None) -> ApiClient:
    env_url = os.environ.get("FREIGHT_API_URL")
    url = base_url or env_url
    return ApiClient(base_url=url)


@app.command()
def health(
    base_url: Optional[str] = typer.Option(None, "--url", "-u", help="API 服务地址"),
) -> None:
    """检查服务健康状态"""
    try:
        client = get_client(base_url)
        result = client.health_check()
        print_health_check(result)
    except Exception as e:
        console.print(f"[bold red]连接失败:[/] {e}")
        raise typer.Exit(1)


@app.command()
def calculate(
    transport_mode: str = typer.Option(
        ...,
        "--mode",
        "-m",
        help="运输方式: land(陆运), air(空运), sea(海运)",
        case_sensitive=False,
    ),
    weight: Decimal = typer.Option(..., "--weight", "-w", help="包裹重量(公斤)"),
    length: Decimal = typer.Option(..., "--length", "-l", help="长度(厘米)"),
    width: Decimal = typer.Option(..., "--width", "-W", help="宽度(厘米)"),
    height: Decimal = typer.Option(..., "--height", "-h", help="高度(厘米)"),
    origin: str = typer.Option(..., "--from", "-f", help="发货地址"),
    destination: str = typer.Option(..., "--to", "-t", help="收货地址"),
    declared_value: Optional[Decimal] = typer.Option(
        None,
        "--value",
        "-v",
        help="申报货值(用于保价)",
    ),
    container_type: Optional[str] = typer.Option(
        None,
        "--container",
        "-c",
        help="集装箱类型: 20ft, 40ft (海运时必填)",
    ),
    customer_id: Optional[str] = typer.Option(
        None,
        "--customer",
        help="客户ID(用于累计运费折扣)",
    ),
    base_url: Optional[str] = typer.Option(None, "--url", "-u", help="API 服务地址"),
) -> None:
    """计算单个包裹运费"""
    try:
        mode_map = {
            "land": TransportMode.LAND,
            "air": TransportMode.AIR,
            "sea": TransportMode.SEA,
        }
        mode = mode_map.get(transport_mode.lower())
        if mode is None:
            console.print(f"[bold red]无效的运输方式:[/] {transport_mode}")
            console.print("可选值: land, air, sea")
            raise typer.Exit(1)
        
        container: Optional[ContainerType] = None
        if container_type:
            ct_map = {
                "20ft": ContainerType.CONTAINER_20FT,
                "40ft": ContainerType.CONTAINER_40FT,
            }
            container = ct_map.get(container_type.lower())
            if container is None:
                console.print(f"[bold red]无效的集装箱类型:[/] {container_type}")
                console.print("可选值: 20ft, 40ft")
                raise typer.Exit(1)
        
        pkg = Package(
            weight_kg=weight,
            length_cm=length,
            width_cm=width,
            height_cm=height,
            origin=origin,
            destination=destination,
            declared_value=declared_value,
            container_type=container,
        )
        
        client = get_client(base_url)
        result = client.calculate_freight(mode, [pkg], customer_id)
        print_calculation_result(result)
        
    except Exception as e:
        console.print(f"[bold red]计算失败:[/] {e}")
        raise typer.Exit(1)


@app.command("stats-months")
def list_statistics_months(
    base_url: Optional[str] = typer.Option(None, "--url", "-u", help="API 服务地址"),
) -> None:
    """列出可用的统计月份"""
    try:
        client = get_client(base_url)
        result = client.list_months()
        print_statistics(result)
    except Exception as e:
        console.print(f"[bold red]获取统计数据失败:[/] {e}")
        raise typer.Exit(1)


@app.command("stats")
def get_statistics(
    year_month: str = typer.Argument(..., help="年月，格式: YYYY-MM"),
    base_url: Optional[str] = typer.Option(None, "--url", "-u", help="API 服务地址"),
) -> None:
    """获取指定月份的统计数据"""
    try:
        client = get_client(base_url)
        result = client.get_monthly_statistics(year_month)
        print_statistics(result)
    except Exception as e:
        console.print(f"[bold red]获取统计数据失败:[/] {e}")
        raise typer.Exit(1)


@app.command("config-get")
def get_config(
    base_url: Optional[str] = typer.Option(None, "--url", "-u", help="API 服务地址"),
) -> None:
    """获取当前配置"""
    try:
        client = get_client(base_url)
        result = client.get_config()
        print_config(result)
    except Exception as e:
        console.print(f"[bold red]获取配置失败:[/] {e}")
        raise typer.Exit(1)


@app.command("config-pricing")
def update_pricing(
    land_first_weight: Optional[Decimal] = typer.Option(
        None,
        "--land-first",
        help="陆运首重价格(元/公斤)",
    ),
    land_continue_weight: Optional[Decimal] = typer.Option(
        None,
        "--land-continue",
        help="陆运续重价格(元/公斤)",
    ),
    sea_20ft: Optional[Decimal] = typer.Option(
        None,
        "--sea-20ft",
        help="20尺柜价格(元)",
    ),
    sea_40ft: Optional[Decimal] = typer.Option(
        None,
        "--sea-40ft",
        help="40尺柜价格(元)",
    ),
    base_url: Optional[str] = typer.Option(None, "--url", "-u", help="API 服务地址"),
) -> None:
    """更新定价配置"""
    try:
        client = get_client(base_url)
        result = client.update_pricing(
            land_first_weight_price=land_first_weight,
            land_continue_weight_price=land_continue_weight,
            sea_20ft_price=sea_20ft,
            sea_40ft_price=sea_40ft,
        )
        print_config(result)
    except Exception as e:
        console.print(f"[bold red]更新配置失败:[/] {e}")
        raise typer.Exit(1)


@app.command("config-remote")
def update_remote_areas(
    areas: Optional[List[str]] = typer.Option(
        None,
        "--area",
        "-a",
        help="偏远地区名称(可多次指定)",
    ),
    surcharge_rate: Optional[Decimal] = typer.Option(
        None,
        "--rate",
        "-r",
        help="附加费率(如 0.2 表示20%)",
    ),
    base_url: Optional[str] = typer.Option(None, "--url", "-u", help="API 服务地址"),
) -> None:
    """更新偏远地区配置"""
    try:
        client = get_client(base_url)
        result = client.update_remote_areas(
            areas=areas,
            surcharge_rate=surcharge_rate,
        )
        print_config(result)
    except Exception as e:
        console.print(f"[bold red]更新配置失败:[/] {e}")
        raise typer.Exit(1)


@app.command("config-discount")
def update_discounts(
    thresholds: List[Decimal] = typer.Option(
        ...,
        "--threshold",
        "-t",
        help="累计金额阈值(可多次指定，需与 --rate 一一对应)",
    ),
    rates: List[Decimal] = typer.Option(
        ...,
        "--rate",
        "-r",
        help="折扣率(可多次指定，需与 --threshold 一一对应)",
    ),
    base_url: Optional[str] = typer.Option(None, "--url", "-u", help="API 服务地址"),
) -> None:
    """更新折扣配置"""
    try:
        if len(thresholds) != len(rates):
            console.print("[bold red]错误:[/] 阈值数量和折扣率数量必须一致")
            raise typer.Exit(1)
        
        client = get_client(base_url)
        result = client.update_discounts(thresholds=thresholds, rates=rates)
        print_config(result)
    except Exception as e:
        console.print(f"[bold red]更新配置失败:[/] {e}")
        raise typer.Exit(1)


@app.command("cache-stats")
def cache_stats(
    base_url: Optional[str] = typer.Option(None, "--url", "-u", help="API 服务地址"),
) -> None:
    """获取缓存统计"""
    try:
        client = get_client(base_url)
        result = client.get_cache_stats()
        print_cache_stats(result)
    except Exception as e:
        console.print(f"[bold red]获取缓存统计失败:[/] {e}")
        raise typer.Exit(1)


@app.command("cache-cleanup")
def cache_cleanup(
    base_url: Optional[str] = typer.Option(None, "--url", "-u", help="API 服务地址"),
) -> None:
    """清理过期缓存"""
    try:
        client = get_client(base_url)
        result = client.cleanup_cache()
        print_operation_result(result, "清理过期缓存")
    except Exception as e:
        console.print(f"[bold red]清理缓存失败:[/] {e}")
        raise typer.Exit(1)


@app.command("cache-clear")
def cache_clear(
    base_url: Optional[str] = typer.Option(None, "--url", "-u", help="API 服务地址"),
) -> None:
    """清空所有缓存"""
    try:
        client = get_client(base_url)
        result = client.clear_cache()
        print_operation_result(result, "清空缓存")
    except Exception as e:
        console.print(f"[bold red]清空缓存失败:[/] {e}")
        raise typer.Exit(1)


def main() -> None:
    app()


if __name__ == "__main__":
    main()
