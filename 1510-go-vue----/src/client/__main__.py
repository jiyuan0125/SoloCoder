import argparse
import json
import sys
from datetime import date

from .api_client import ApiClient


def format_output(data):
    if isinstance(data, list) and data:
        for item in data:
            print(json.dumps(item, indent=2, default=str))
            print("---")
    elif data:
        print(json.dumps(data, indent=2, default=str))
    else:
        print("No data")


def cmd_plots(args, client):
    if args.action == "list":
        format_output(client.list_plots())
    elif args.action == "create":
        result = client.create_plot(args.name, args.variety, args.year)
        format_output(result)


def cmd_harvests(args, client):
    if args.action == "list":
        format_output(client.list_harvests(args.plot_id))
    elif args.action == "create":
        result = client.create_harvest(
            args.plot_id, args.harvest_date, args.quantity,
            args.brix, args.acidity
        )
        format_output(result)


def cmd_batches(args, client):
    if args.action == "list":
        format_output(client.list_batches())
    elif args.action == "create":
        harvest_ids = args.harvest_ids.split(",")
        quantities = [float(q) for q in args.quantities.split(",")]
        result = client.create_batch(
            args.name, args.description, harvest_ids,
            quantities, args.fermentation_start, args.fermentation_end
        )
        format_output(result)
    elif args.action == "complete-fermentation":
        result = client.complete_fermentation(args.batch_id, args.end_date)
        format_output(result)


def cmd_cellars(args, client):
    if args.action == "list":
        format_output(client.list_cellars())
    elif args.action == "create":
        result = client.create_cellar(args.name, args.location, args.total_slots)
        format_output(result)
    elif args.action == "usage":
        result = client.get_cellar_usage(args.cellar_id)
        format_output(result)


def cmd_storages(args, client):
    if args.action == "list":
        format_output(client.list_storages(args.cellar_id, args.batch_id, args.active_only))
    elif args.action == "create":
        result = client.create_storage(
            args.batch_id, args.cellar_id, args.shelf,
            args.position, args.aging_months, args.start_date
        )
        format_output(result)


def cmd_todos(args, client):
    if args.action == "pending-tastings":
        result = client.get_pending_tastings()
        if not result:
            print("No pending tastings")
        else:
            print(f"Pending tastings ({len(result)}):")
            for item in result:
                print(f"  - Batch: {item['batch_name']} (ID: {item['batch_id'][:8]}...)")
                print(f"    Due: {item['expected_end_date']}, {item['days_overdue']} days overdue")


def cmd_tastings(args, client):
    if args.action == "list":
        format_output(client.list_tastings(args.batch_id))
    elif args.action == "record":
        result = client.record_tasting(
            args.batch_id, args.tasting_date, args.decision,
            args.notes, args.additional_months
        )
        format_output(result)


def main():
    parser = argparse.ArgumentParser(
        description="Winery Management CLI Client",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    plots_parser = subparsers.add_parser("plots", help="Plot management")
    plots_sub = plots_parser.add_subparsers(dest="action", required=True)
    plots_sub.add_parser("list", help="List all plots")
    create_plot = plots_sub.add_parser("create", help="Create a new plot")
    create_plot.add_argument("--name", required=True, help="Plot name")
    create_plot.add_argument("--variety", required=True, help="Grape variety")
    create_plot.add_argument("--year", type=int, required=True, help="Planting year")

    harvests_parser = subparsers.add_parser("harvests", help="Harvest management")
    harvests_sub = harvests_parser.add_subparsers(dest="action", required=True)
    list_harvest = harvests_sub.add_parser("list", help="List harvests")
    list_harvest.add_argument("--plot-id", help="Filter by plot ID")
    create_harvest = harvests_sub.add_parser("create", help="Create harvest record")
    create_harvest.add_argument("--plot-id", required=True, help="Plot ID")
    create_harvest.add_argument("--harvest-date", required=True, help="Harvest date (YYYY-MM-DD)")
    create_harvest.add_argument("--quantity", type=float, required=True, help="Quantity (kg)")
    create_harvest.add_argument("--brix", type=float, required=True, help="Brix level")
    create_harvest.add_argument("--acidity", type=float, required=True, help="Acidity level")

    batches_parser = subparsers.add_parser("batches", help="Batch management")
    batches_sub = batches_parser.add_subparsers(dest="action", required=True)
    batches_sub.add_parser("list", help="List all batches")
    create_batch = batches_sub.add_parser("create", help="Create a new batch")
    create_batch.add_argument("--name", required=True, help="Batch name")
    create_batch.add_argument("--description", help="Batch description")
    create_batch.add_argument("--harvest-ids", required=True, help="Comma-separated harvest IDs")
    create_batch.add_argument("--quantities", required=True, help="Comma-separated quantities (kg)")
    create_batch.add_argument("--fermentation-start", required=True, help="Fermentation start date")
    create_batch.add_argument("--fermentation-end", help="Fermentation end date (optional)")
    complete_ferm = batches_sub.add_parser("complete-fermentation", help="Complete fermentation")
    complete_ferm.add_argument("--batch-id", required=True, help="Batch ID")
    complete_ferm.add_argument("--end-date", required=True, help="End date (YYYY-MM-DD)")

    cellars_parser = subparsers.add_parser("cellars", help="Cellar management")
    cellars_sub = cellars_parser.add_subparsers(dest="action", required=True)
    cellars_sub.add_parser("list", help="List all cellars")
    create_cellar = cellars_sub.add_parser("create", help="Create a new cellar")
    create_cellar.add_argument("--name", required=True, help="Cellar name")
    create_cellar.add_argument("--location", help="Cellar location")
    create_cellar.add_argument("--total-slots", type=int, required=True, help="Total storage slots")
    usage_cellar = cellars_sub.add_parser("usage", help="Check cellar usage")
    usage_cellar.add_argument("--cellar-id", required=True, help="Cellar ID")

    storages_parser = subparsers.add_parser("storages", help="Storage management")
    storages_sub = storages_parser.add_subparsers(dest="action", required=True)
    list_storage = storages_sub.add_parser("list", help="List storages")
    list_storage.add_argument("--cellar-id", help="Filter by cellar ID")
    list_storage.add_argument("--batch-id", help="Filter by batch ID")
    list_storage.add_argument("--active-only", action="store_true", help="Show only active")
    create_storage = storages_sub.add_parser("create", help="Create storage record")
    create_storage.add_argument("--batch-id", required=True, help="Batch ID")
    create_storage.add_argument("--cellar-id", required=True, help="Cellar ID")
    create_storage.add_argument("--shelf", type=int, required=True, help="Shelf number")
    create_storage.add_argument("--position", required=True, help="Position identifier")
    create_storage.add_argument("--aging-months", type=int, required=True, help="Expected aging months")
    create_storage.add_argument("--start-date", required=True, help="Storage start date")

    todos_parser = subparsers.add_parser("todos", help="Todo items")
    todos_sub = todos_parser.add_subparsers(dest="action", required=True)
    todos_sub.add_parser("pending-tastings", help="Show pending tastings")

    tastings_parser = subparsers.add_parser("tastings", help="Tasting management")
    tastings_sub = tastings_parser.add_subparsers(dest="action", required=True)
    list_tasting = tastings_sub.add_parser("list", help="List tastings")
    list_tasting.add_argument("--batch-id", help="Filter by batch ID")
    record_tasting = tastings_sub.add_parser("record", help="Record a tasting")
    record_tasting.add_argument("--batch-id", required=True, help="Batch ID")
    record_tasting.add_argument("--tasting-date", required=True, help="Tasting date (YYYY-MM-DD)")
    record_tasting.add_argument("--decision", required=True, choices=["continue_aging", "bottle"],
                                help="Decision: continue_aging or bottle")
    record_tasting.add_argument("--notes", help="Tasting notes")
    record_tasting.add_argument("--additional-months", type=int,
                                help="Additional months to age (required for continue_aging)")

    args = parser.parse_args()
    client = ApiClient()

    try:
        if args.command == "plots":
            cmd_plots(args, client)
        elif args.command == "harvests":
            cmd_harvests(args, client)
        elif args.command == "batches":
            cmd_batches(args, client)
        elif args.command == "cellars":
            cmd_cellars(args, client)
        elif args.command == "storages":
            cmd_storages(args, client)
        elif args.command == "todos":
            cmd_todos(args, client)
        elif args.command == "tastings":
            cmd_tastings(args, client)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
