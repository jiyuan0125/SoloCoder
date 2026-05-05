import click

from cli.commands.alert import alert_group
from cli.commands.inventory import inventory_group
from cli.commands.product import product_group
from cli.commands.replenishment import replenishment_group
from cli.commands.report import report_group


@click.group()
@click.version_option("0.1.0")
def main() -> None:
    """Inventory Alert System CLI

    A command-line interface for managing inventory, products, alerts, and reports.
    """
    pass


main.add_command(product_group)
main.add_command(inventory_group)
main.add_command(alert_group)
main.add_command(replenishment_group)
main.add_command(report_group)


if __name__ == "__main__":
    main()
