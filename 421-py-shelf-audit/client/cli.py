import click
from typing import Any, Optional
from decimal import Decimal
from rich.console import Console
from rich.table import Table
from rich.panel import Panel

from client.api_client import ApiClient
from shared import LocationStatus, ErrorCode


console = Console()


def get_client(ctx: click.Context) -> ApiClient:
    client: ApiClient = ctx.obj["client"]
    return client


def print_response(response: dict[str, Any]) -> None:
    if response.get("success"):
        console.print(Panel("[green]Success[/green]", expand=False))
        data = response.get("data")
        if data:
            print_data(data)
    else:
        error_code = response.get("error_code", "unknown")
        error_message = response.get("error_message", "Unknown error")
        console.print(Panel(f"[red]Error: {error_code}[/red]\n{error_message}", expand=False))


def print_data(data: Any) -> None:
    if isinstance(data, list):
        if not data:
            console.print("[yellow]No results[/yellow]")
            return
        
        first_item = data[0]
        if isinstance(first_item, dict):
            print_table_from_dicts(data)
        else:
            console.print(data)
    elif isinstance(data, dict):
        print_dict_as_table(data)
    else:
        console.print(data)


def print_table_from_dicts(items: list[dict[str, Any]]) -> None:
    if not items:
        return
    
    first = items[0]
    table = Table()
    
    for key in first.keys():
        table.add_column(str(key), style="cyan")
    
    for item in items:
        row = [str(item.get(key, "")) for key in first.keys()]
        table.add_row(*row)
    
    console.print(table)


def print_dict_as_table(d: dict[str, Any]) -> None:
    table = Table(show_header=True)
    table.add_column("Key", style="magenta")
    table.add_column("Value", style="cyan")
    
    for key, value in d.items():
        if isinstance(value, dict):
            table.add_row(key, f"[dict with {len(value)} keys]")
        elif isinstance(value, list):
            table.add_row(key, f"[list with {len(value)} items]")
        else:
            table.add_row(key, str(value))
    
    console.print(table)


@click.group()
@click.option("--base-url", default="http://localhost:8000", help="API server base URL")
@click.pass_context
def cli(ctx: click.Context, base_url: str) -> None:
    """Shelf Audit System CLI"""
    ctx.ensure_object(dict)
    ctx.obj["client"] = ApiClient(base_url=base_url)


@cli.group()
def zone() -> None:
    """Zone management commands"""
    pass


@zone.command("create")
@click.argument("code")
@click.argument("name")
@click.option("--description", "-d", help="Zone description")
@click.pass_context
def zone_create(ctx: click.Context, code: str, name: str, description: Optional[str]) -> None:
    """Create a new zone"""
    client = get_client(ctx)
    result = client.create_zone(code, name, description)
    print_response(result)


@zone.command("list")
@click.pass_context
def zone_list(ctx: click.Context) -> None:
    """List all zones"""
    client = get_client(ctx)
    result = client.list_zones()
    print_response(result)


@zone.command("show")
@click.argument("code")
@click.pass_context
def zone_show(ctx: click.Context, code: str) -> None:
    """Show zone details"""
    client = get_client(ctx)
    result = client.get_zone_detail(code)
    print_response(result)


@cli.group()
def shelf() -> None:
    """Shelf management commands"""
    pass


@shelf.command("create")
@click.argument("zone_code")
@click.argument("shelf_number", type=int)
@click.option("--name", "-n", help="Shelf name")
@click.option("--layers", "-l", type=int, default=5, help="Number of layers")
@click.option("--columns", "-c", type=int, default=10, help="Number of columns")
@click.option("--capacity", type=int, default=100, help="Default capacity per location")
@click.pass_context
def shelf_create(
    ctx: click.Context,
    zone_code: str,
    shelf_number: int,
    name: Optional[str],
    layers: int,
    columns: int,
    capacity: int
) -> None:
    """Create a new shelf"""
    client = get_client(ctx)
    result = client.create_shelf(zone_code, shelf_number, name, layers, columns, capacity)
    print_response(result)


@shelf.command("list")
@click.option("--zone", "-z", help="Filter by zone code")
@click.pass_context
def shelf_list(ctx: click.Context, zone: Optional[str]) -> None:
    """List all shelves"""
    client = get_client(ctx)
    result = client.list_shelves(zone)
    print_response(result)


@shelf.command("show")
@click.argument("zone_code")
@click.argument("shelf_number", type=int)
@click.pass_context
def shelf_show(ctx: click.Context, zone_code: str, shelf_number: int) -> None:
    """Show shelf details"""
    client = get_client(ctx)
    result = client.get_shelf_detail(zone_code, shelf_number)
    print_response(result)


@cli.group()
def product() -> None:
    """Product management commands"""
    pass


@product.command("create")
@click.argument("sku")
@click.argument("name")
@click.option("--description", "-d", help="Product description")
@click.option("--price", "-p", type=str, required=True, help="Unit price (e.g., 19.99)")
@click.option("--volume", "-v", type=str, required=True, help="Volume per unit (e.g., 0.5)")
@click.pass_context
def product_create(
    ctx: click.Context,
    sku: str,
    name: str,
    description: Optional[str],
    price: str,
    volume: str
) -> None:
    """Create a new product"""
    client = get_client(ctx)
    result = client.create_product(sku, name, description, Decimal(price), Decimal(volume))
    print_response(result)


@product.command("list")
@click.pass_context
def product_list(ctx: click.Context) -> None:
    """List all products"""
    client = get_client(ctx)
    result = client.list_products()
    print_response(result)


@product.command("show")
@click.argument("sku")
@click.pass_context
def product_show(ctx: click.Context, sku: str) -> None:
    """Show product details"""
    client = get_client(ctx)
    result = client.get_product(sku)
    print_response(result)


@cli.group()
def location() -> None:
    """Location management commands"""
    pass


@location.command("list")
@click.option("--zone", "-z", help="Filter by zone code")
@click.option("--status", "-s", type=click.Choice([s.value for s in LocationStatus]), help="Filter by status")
@click.option("--product", "-p", help="Filter by product SKU")
@click.pass_context
def location_list(
    ctx: click.Context,
    zone: Optional[str],
    status: Optional[str],
    product: Optional[str]
) -> None:
    """List locations"""
    client = get_client(ctx)
    status_enum = LocationStatus(status) if status else None
    result = client.list_locations(zone, status_enum, product)
    print_response(result)


@location.command("show")
@click.argument("code")
@click.pass_context
def location_show(ctx: click.Context, code: str) -> None:
    """Show location details"""
    client = get_client(ctx)
    result = client.get_location(code)
    print_response(result)


@location.command("set-status")
@click.argument("code")
@click.argument("status", type=click.Choice([s.value for s in LocationStatus]))
@click.pass_context
def location_set_status(ctx: click.Context, code: str, status: str) -> None:
    """Set location status (idle, occupied, locked, under_inventory, under_maintenance)"""
    client = get_client(ctx)
    status_enum = LocationStatus(status)
    result = client.update_location_status(code, status_enum)
    print_response(result)


@location.command("set-capacity")
@click.argument("code")
@click.argument("max_capacity", type=int)
@click.pass_context
def location_set_capacity(ctx: click.Context, code: str, max_capacity: int) -> None:
    """Set location maximum capacity"""
    client = get_client(ctx)
    result = client.set_location_capacity(code, max_capacity)
    print_response(result)


@cli.group()
def stock() -> None:
    """Stock operations (in/out)"""
    pass


@stock.command("in")
@click.argument("product_sku")
@click.argument("quantity", type=int)
@click.option("--zone", "-z", help="Preferred zone for allocation")
@click.option("--location", "-l", help="Specific location code (skip auto-allocation)")
@click.pass_context
def stock_in_cmd(
    ctx: click.Context,
    product_sku: str,
    quantity: int,
    zone: Optional[str],
    location: Optional[str]
) -> None:
    """Stock in products to the warehouse"""
    client = get_client(ctx)
    result = client.stock_in(product_sku, quantity, zone, location)
    print_response(result)


@stock.command("out")
@click.argument("product_sku")
@click.argument("quantity", type=int)
@click.pass_context
def stock_out_cmd(ctx: click.Context, product_sku: str, quantity: int) -> None:
    """Stock out products from the warehouse (FIFO)"""
    client = get_client(ctx)
    result = client.stock_out(product_sku, quantity)
    print_response(result)


@cli.group()
def query() -> None:
    """Query operations"""
    pass


@query.command("location")
@click.argument("code")
@click.pass_context
def query_location_cmd(ctx: click.Context, code: str) -> None:
    """Query what products are in a location"""
    client = get_client(ctx)
    result = client.query_location(code)
    print_response(result)


@query.command("product")
@click.argument("sku")
@click.pass_context
def query_product_cmd(ctx: click.Context, sku: str) -> None:
    """Query which locations contain a product"""
    client = get_client(ctx)
    result = client.query_product_locations(sku)
    print_response(result)


@cli.group()
def inventory() -> None:
    """Inventory operations"""
    pass


@inventory.command("start")
@click.option("--zone", "-z", help="Start inventory for a specific zone")
@click.option("--locations", "-l", multiple=True, help="Start inventory for specific locations")
@click.option("--all", "-a", "include_all", is_flag=True, help="Start inventory for all locations")
@click.pass_context
def inventory_start(
    ctx: click.Context,
    zone: Optional[str],
    locations: tuple[str, ...],
    include_all: bool
) -> None:
    """Start an inventory session"""
    client = get_client(ctx)
    location_list = list(locations) if locations else None
    result = client.start_inventory(zone, location_list, include_all)
    print_response(result)


@inventory.command("complete")
@click.argument("session_id")
@click.option("--item", "-i", multiple=True, type=str, help="Inventory items: location_code=quantity (e.g., A-01-01-01=50)")
@click.pass_context
def inventory_complete(ctx: click.Context, session_id: str, item: tuple[str, ...]) -> None:
    """Complete an inventory session with actual counts"""
    client = get_client(ctx)
    items: list[tuple[str, int]] = []
    for i in item:
        parts = i.split("=")
        if len(parts) == 2:
            items.append((parts[0], int(parts[1])))
    
    result = client.complete_inventory(session_id, items)
    print_response(result)


@inventory.command("show")
@click.argument("session_id")
@click.pass_context
def inventory_show(ctx: click.Context, session_id: str) -> None:
    """Show inventory session details"""
    client = get_client(ctx)
    result = client.get_inventory_session(session_id)
    print_response(result)


@cli.group()
def alert() -> None:
    """Alert management commands"""
    pass


@alert.command("list")
@click.option("--resolved", "-r", is_flag=True, help="Show only resolved alerts")
@click.option("--unresolved", "-u", is_flag=True, help="Show only unresolved alerts")
@click.pass_context
def alert_list(ctx: click.Context, resolved: bool, unresolved: bool) -> None:
    """List alerts"""
    client = get_client(ctx)
    resolved_flag: Optional[bool] = None
    if resolved:
        resolved_flag = True
    elif unresolved:
        resolved_flag = False
    result = client.list_alerts(resolved_flag)
    print_response(result)


@alert.command("resolve")
@click.argument("alert_id")
@click.pass_context
def alert_resolve(ctx: click.Context, alert_id: str) -> None:
    """Resolve an alert"""
    client = get_client(ctx)
    result = client.resolve_alert(alert_id)
    print_response(result)


if __name__ == "__main__":
    cli()
