from cli.commands.supplier_commands import app as supplier_app
from cli.commands.purchase_commands import app as purchase_app
from cli.commands.quote_commands import app as quote_app
from cli.commands.order_commands import app as order_app
from cli.commands.report_commands import app as report_app

__all__ = [
    "supplier_app",
    "purchase_app",
    "quote_app",
    "order_app",
    "report_app",
]
