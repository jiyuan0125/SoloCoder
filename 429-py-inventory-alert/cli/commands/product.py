from typing import Any, Optional

import click
import httpx

from cli.client import APIClient
from cli.output import (
    console,
    print_error,
    print_json,
    print_product_detail,
    print_product_table,
    print_success,
)
from shared.models import Product, ProductCreate, ProductUpdate


@click.group(name="product")
def product_group() -> None:
    """Product management commands"""
    pass


@product_group.command(name="create")
@click.option("--sku", required=True, help="Product SKU")
@click.option("--name", required=True, help="Product name")
@click.option("--category", required=True, help="Product category")
@click.option("--stock", default=0, type=int, help="Initial stock")
@click.option("--min-stock", default=0, type=int, help="Minimum safety stock")
@click.option("--max-stock", default=1000, type=int, help="Maximum safety stock")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def create_product(
    sku: str,
    name: str,
    category: str,
    stock: int,
    min_stock: int,
    max_stock: int,
    output_json: bool,
) -> None:
    """Create a new product"""
    from shared.models import SafetyStock

    with APIClient() as client:
        product_data = ProductCreate(
            sku=sku,
            name=name,
            category=category,
            initial_stock=stock,
            safety_stock=SafetyStock(min_stock=min_stock, max_stock=max_stock),
        )

        try:
            response = client.post(
                "/products",
                json=product_data.model_dump(mode="json"),
            )
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                product = Product(**result["data"])
                print_success(f"Product created: {product.sku}")
                print_product_detail(product)
            else:
                print_error(result.get("error", "Failed to create product"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@product_group.command(name="get")
@click.argument("sku")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def get_product(sku: str, output_json: bool) -> None:
    """Get product details by SKU"""
    with APIClient() as client:
        try:
            response = client.get(f"/products/{sku}")
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                product = Product(**result["data"])
                print_product_detail(product)
            else:
                print_error(result.get("error", "Product not found"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@product_group.command(name="list")
@click.option("--category", default=None, help="Filter by category")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def list_products(category: Optional[str], output_json: bool) -> None:
    """List all products, optionally filtered by category"""
    with APIClient() as client:
        try:
            params = {}
            if category:
                params["category"] = category

            response = client.get("/products", params=params)
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                products = [Product(**p) for p in result["data"]]
                print_product_table(products)
            else:
                print_error(result.get("error", "Failed to list products"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@product_group.command(name="update")
@click.argument("sku")
@click.option("--name", default=None, help="New product name")
@click.option("--category", default=None, help="New product category")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def update_product(
    sku: str,
    name: Optional[str],
    category: Optional[str],
    output_json: bool,
) -> None:
    """Update product information"""
    with APIClient() as client:
        update_data = ProductUpdate(name=name, category=category)

        try:
            response = client.put(
                f"/products/{sku}",
                json=update_data.model_dump(mode="json", exclude_unset=True),
            )
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                product = Product(**result["data"])
                print_success(f"Product updated: {product.sku}")
                print_product_detail(product)
            else:
                print_error(result.get("error", "Failed to update product"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@product_group.command(name="delete")
@click.argument("sku")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def delete_product(sku: str, output_json: bool) -> None:
    """Delete a product by SKU"""
    with APIClient() as client:
        try:
            response = client.delete(f"/products/{sku}")
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success"):
                print_success(f"Product deleted: {sku}")
            else:
                print_error(result.get("error", "Failed to delete product"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@product_group.command(name="set-safety-stock")
@click.argument("sku")
@click.option("--min-stock", required=True, type=int, help="Minimum safety stock")
@click.option("--max-stock", required=True, type=int, help="Maximum safety stock")
@click.option("--seasonal", is_flag=True, help="Enable seasonal safety stock")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def set_safety_stock(
    sku: str,
    min_stock: int,
    max_stock: int,
    seasonal: bool,
    output_json: bool,
) -> None:
    """Set safety stock for a product"""
    from shared.models import SafetyStock

    with APIClient() as client:
        safety_stock = SafetyStock(
            is_seasonal=seasonal,
            min_stock=min_stock,
            max_stock=max_stock,
        )

        try:
            response = client.put(
                f"/products/{sku}/safety-stock",
                json=safety_stock.model_dump(mode="json"),
            )
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success") and result.get("data"):
                product = Product(**result["data"])
                print_success(f"Safety stock updated for: {product.sku}")
                print_product_detail(product)
            else:
                print_error(result.get("error", "Failed to update safety stock"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")


@product_group.command(name="bulk-set-safety-stock")
@click.option("--category", required=True, help="Product category to update")
@click.option("--min-stock", required=True, type=int, help="Minimum safety stock")
@click.option("--max-stock", required=True, type=int, help="Maximum safety stock")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
def bulk_set_safety_stock(
    category: str,
    min_stock: int,
    max_stock: int,
    output_json: bool,
) -> None:
    """Bulk set safety stock for all products in a category"""
    from shared.models import BulkSafetyStockUpdate, SafetyStock

    with APIClient() as client:
        bulk_update = BulkSafetyStockUpdate(
            category=category,
            safety_stock=SafetyStock(min_stock=min_stock, max_stock=max_stock),
        )

        try:
            response = client.post(
                "/products/bulk-safety-stock",
                json=bulk_update.model_dump(mode="json"),
            )
            result = response.json()

            if output_json:
                print_json(result)
                return

            if result.get("success"):
                count = result.get("data", 0)
                print_success(f"Updated safety stock for {count} products in category: {category}")
            else:
                print_error(result.get("error", "Failed to bulk update safety stock"))
        except httpx.HTTPError as e:
            print_error(f"Connection error: {e}")
