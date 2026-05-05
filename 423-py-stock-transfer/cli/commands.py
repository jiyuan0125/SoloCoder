from typing import Optional
from uuid import UUID

import httpx
import typer

from cli.client import ApiClient
from cli.printer import (
    print_error,
    print_inventory_overview,
    print_loss_list,
    print_loss_summary,
    print_products,
    print_print_document,
    print_success,
    print_transaction_list,
    print_transfer_detail,
    print_transfer_list,
    print_warehouses,
)
from shared.enums import (
    ApprovalLevel,
    ApprovalStatus,
    TransferStatus,
    TransferType,
)

app = typer.Typer(
    name="stock-transfer",
    help="库存调拨系统命令行工具",
    no_args_is_help=True,
)


def get_client(base_url: str) -> ApiClient:
    return ApiClient(base_url=base_url)


@app.command()
def health(
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """检查服务端健康状态"""
    try:
        with get_client(base_url) as client:
            result = client.health_check()
            status = result.get("status", "unknown")
            if status == "ok":
                print_success("服务端运行正常")
            else:
                print_error(f"服务端状态异常: {status}")
    except httpx.ConnectError:
        print_error("无法连接到服务端，请确认服务已启动")
    except Exception as e:
        print_error(str(e))


@app.command("warehouse-list")
def list_warehouses(
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """列出所有仓库"""
    try:
        with get_client(base_url) as client:
            warehouses = client.list_warehouses()
            print_warehouses(warehouses)
    except Exception as e:
        print_error(str(e))


@app.command("product-list")
def list_products(
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """列出所有商品"""
    try:
        with get_client(base_url) as client:
            products = client.list_products()
            print_products(products)
    except Exception as e:
        print_error(str(e))


@app.command("inventory-list")
def list_inventory(
    warehouse_id: Optional[str] = typer.Option(
        None,
        "--warehouse-id",
        "-w",
        help="仓库ID过滤",
    ),
    product_id: Optional[str] = typer.Option(
        None,
        "--product-id",
        "-p",
        help="商品ID过滤",
    ),
    page: int = typer.Option(1, "--page", help="页码"),
    page_size: int = typer.Option(100, "--page-size", help="每页数量"),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """列出库存总览"""
    try:
        wh_uuid = UUID(warehouse_id) if warehouse_id else None
        prod_uuid = UUID(product_id) if product_id else None

        with get_client(base_url) as client:
            result = client.list_inventory(
                warehouse_id=wh_uuid,
                product_id=prod_uuid,
                page=page,
                page_size=page_size,
            )
            print_inventory_overview(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-create")
def create_transfer(
    source_warehouse_id: str = typer.Option(
        ...,
        "--source",
        "-s",
        help="源仓库ID",
    ),
    target_warehouse_id: str = typer.Option(
        ...,
        "--target",
        "-t",
        help="目标仓库ID",
    ),
    items: list[str] = typer.Option(
        ...,
        "--item",
        "-i",
        help="调拨商品项，格式: product_id:quantity",
    ),
    transfer_type: TransferType = typer.Option(
        TransferType.INTRA_COMPANY,
        "--type",
        help="调拨类型",
    ),
    created_by: Optional[str] = typer.Option(
        None,
        "--created-by",
        help="创建人ID",
    ),
    remark: Optional[str] = typer.Option(
        None,
        "--remark",
        help="备注",
    ),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """创建调拨单

    示例:
      stock-transfer transfer-create \\
        -s <source_warehouse_id> \\
        -t <target_warehouse_id> \\
        -i <product_id1>:10 \\
        -i <product_id2>:5
    """
    try:
        source_uuid = UUID(source_warehouse_id)
        target_uuid = UUID(target_warehouse_id)
        created_by_uuid = UUID(created_by) if created_by else None

        item_list: list[dict[str, str | int]] = []
        for item_str in items:
            if ":" not in item_str:
                raise ValueError(f"无效的商品项格式: {item_str}，应为 product_id:quantity")
            parts = item_str.split(":")
            if len(parts) != 2:
                raise ValueError(f"无效的商品项格式: {item_str}")
            product_id = parts[0].strip()
            quantity = int(parts[1].strip())
            item_list.append({
                "product_id": product_id,
                "requested_quantity": quantity,
            })

        with get_client(base_url) as client:
            result = client.create_transfer(
                source_warehouse_id=source_uuid,
                target_warehouse_id=target_uuid,
                items=item_list,
                transfer_type=transfer_type,
                created_by=created_by_uuid,
                remark=remark,
            )
            print_success("调拨单创建成功")
            print_transfer_detail(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-get")
def get_transfer(
    transfer_id: str = typer.Argument(..., help="调拨单ID"),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """获取调拨单详情"""
    try:
        transfer_uuid = UUID(transfer_id)
        with get_client(base_url) as client:
            result = client.get_transfer(transfer_uuid)
            print_transfer_detail(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-list")
def list_transfers(
    status: Optional[TransferStatus] = typer.Option(
        None,
        "--status",
        help="状态过滤",
    ),
    source_warehouse_id: Optional[str] = typer.Option(
        None,
        "--source",
        "-s",
        help="源仓库ID过滤",
    ),
    target_warehouse_id: Optional[str] = typer.Option(
        None,
        "--target",
        "-t",
        help="目标仓库ID过滤",
    ),
    transfer_type: Optional[TransferType] = typer.Option(
        None,
        "--type",
        help="调拨类型过滤",
    ),
    page: int = typer.Option(1, "--page", help="页码"),
    page_size: int = typer.Option(20, "--page-size", help="每页数量"),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """列出调拨单"""
    try:
        source_uuid = UUID(source_warehouse_id) if source_warehouse_id else None
        target_uuid = UUID(target_warehouse_id) if target_warehouse_id else None

        with get_client(base_url) as client:
            result = client.list_transfers(
                status=status,
                source_warehouse_id=source_uuid,
                target_warehouse_id=target_uuid,
                transfer_type=transfer_type,
                page=page,
                page_size=page_size,
            )
            print_transfer_list(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-confirm")
def confirm_transfer(
    transfer_id: str = typer.Argument(..., help="调拨单ID"),
    operator_id: Optional[str] = typer.Option(
        None,
        "--operator",
        help="操作人ID",
    ),
    remark: Optional[str] = typer.Option(
        None,
        "--remark",
        help="备注",
    ),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """确认调拨单（冻结库存）"""
    try:
        transfer_uuid = UUID(transfer_id)
        operator_uuid = UUID(operator_id) if operator_id else None

        with get_client(base_url) as client:
            result = client.confirm_transfer(
                transfer_id=transfer_uuid,
                operator_id=operator_uuid,
                remark=remark,
            )
            print_success("调拨单已确认，库存已冻结")
            print_transfer_detail(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-approve")
def approve_transfer(
    transfer_id: str = typer.Argument(..., help="调拨单ID"),
    approval_level: ApprovalLevel = typer.Option(
        ApprovalLevel.WAREHOUSE_MANAGER,
        "--level",
        "-l",
        help="审批级别",
    ),
    approve: bool = typer.Option(
        True,
        "--approve/--reject",
        help="批准或拒绝",
    ),
    approver_id: str = typer.Option(
        ...,
        "--approver-id",
        help="审批人ID",
    ),
    approver_name: str = typer.Option(
        ...,
        "--approver-name",
        help="审批人姓名",
    ),
    comment: Optional[str] = typer.Option(
        None,
        "--comment",
        help="审批意见",
    ),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """审批调拨单"""
    try:
        transfer_uuid = UUID(transfer_id)
        approver_uuid = UUID(approver_id)

        approval_status = ApprovalStatus.APPROVED if approve else ApprovalStatus.REJECTED

        with get_client(base_url) as client:
            result = client.approve_transfer(
                transfer_id=transfer_uuid,
                approval_level=approval_level,
                approval_status=approval_status,
                approver_id=approver_uuid,
                approver_name=approver_name,
                comment=comment,
            )
            if approve:
                print_success("调拨单已批准")
            else:
                print_success("调拨单已拒绝")
            print_transfer_detail(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-approve-inter-company")
def approve_inter_company(
    transfer_id: str = typer.Argument(..., help="调拨单ID"),
    operator_id: str = typer.Option(
        ...,
        "--operator-id",
        help="操作人ID",
    ),
    operator_name: str = typer.Option(
        ...,
        "--operator-name",
        help="操作人姓名",
    ),
    approve: bool = typer.Option(
        True,
        "--approve/--reject",
        help="批准或拒绝",
    ),
    comment: Optional[str] = typer.Option(
        None,
        "--comment",
        help="审批意见",
    ),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """跨公司调拨审批"""
    try:
        transfer_uuid = UUID(transfer_id)
        operator_uuid = UUID(operator_id)

        with get_client(base_url) as client:
            result = client.approve_inter_company(
                transfer_id=transfer_uuid,
                operator_id=operator_uuid,
                operator_name=operator_name,
                approved=approve,
                comment=comment,
            )
            if approve:
                print_success("跨公司调拨已批准")
            else:
                print_success("跨公司调拨已拒绝")
            print_transfer_detail(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-ship")
def ship_transfer(
    transfer_id: str = typer.Argument(..., help="调拨单ID"),
    operator_id: Optional[str] = typer.Option(
        None,
        "--operator",
        help="操作人ID",
    ),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """发货（状态变为运输中）"""
    try:
        transfer_uuid = UUID(transfer_id)
        operator_uuid = UUID(operator_id) if operator_id else None

        with get_client(base_url) as client:
            result = client.ship_transfer(
                transfer_id=transfer_uuid,
                operator_id=operator_uuid,
            )
            print_success("调拨单已发货，状态变为运输中")
            print_transfer_detail(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-arrive")
def arrive_transfer(
    transfer_id: str = typer.Argument(..., help="调拨单ID"),
    items: list[str] = typer.Option(
        ...,
        "--item",
        "-i",
        help="实际到达数量，格式: item_id:arrived_quantity",
    ),
    operator_id: Optional[str] = typer.Option(
        None,
        "--operator",
        help="操作人ID",
    ),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """确认到达（可指定实际到达数量，有损耗）

    示例:
      stock-transfer transfer-arrive <transfer_id> \\
        -i <item_id1>:95 \\
        -i <item_id2>:50
    """
    try:
        transfer_uuid = UUID(transfer_id)
        operator_uuid = UUID(operator_id) if operator_id else None

        arrival_items: list[dict[str, str | int]] = []
        for item_str in items:
            if ":" not in item_str:
                raise ValueError(f"无效的商品项格式: {item_str}，应为 item_id:arrived_quantity")
            parts = item_str.split(":")
            if len(parts) != 2:
                raise ValueError(f"无效的商品项格式: {item_str}")
            item_id = parts[0].strip()
            arrived_quantity = int(parts[1].strip())
            arrival_items.append({
                "item_id": item_id,
                "arrived_quantity": arrived_quantity,
            })

        with get_client(base_url) as client:
            result = client.arrive_transfer(
                transfer_id=transfer_uuid,
                arrival_items=arrival_items,
                operator_id=operator_uuid,
            )
            print_success("已确认到达，损耗已记录")
            print_transfer_detail(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-stock")
def stock_transfer(
    transfer_id: str = typer.Argument(..., help="调拨单ID"),
    operator_id: Optional[str] = typer.Option(
        None,
        "--operator",
        help="操作人ID",
    ),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """完成入库（正式扣减源仓、增加目标仓）"""
    try:
        transfer_uuid = UUID(transfer_id)
        operator_uuid = UUID(operator_id) if operator_id else None

        with get_client(base_url) as client:
            result = client.stock_transfer(
                transfer_id=transfer_uuid,
                operator_id=operator_uuid,
            )
            print_success("已完成入库，库存已更新")
            print_transfer_detail(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-cancel")
def cancel_transfer(
    transfer_id: str = typer.Argument(..., help="调拨单ID"),
    reason: str = typer.Option(..., "--reason", "-r", help="取消原因"),
    operator_id: Optional[str] = typer.Option(
        None,
        "--operator",
        help="操作人ID",
    ),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """取消调拨单（解冻库存）"""
    try:
        transfer_uuid = UUID(transfer_id)
        operator_uuid = UUID(operator_id) if operator_id else None

        with get_client(base_url) as client:
            result = client.cancel_transfer(
                transfer_id=transfer_uuid,
                reason=reason,
                operator_id=operator_uuid,
            )
            print_success("调拨单已取消，库存已解冻")
            print_transfer_detail(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-print-outbound")
def print_outbound(
    transfer_id: str = typer.Argument(..., help="调拨单ID"),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """打印调拨出库单"""
    try:
        transfer_uuid = UUID(transfer_id)
        with get_client(base_url) as client:
            result = client.print_outbound(transfer_uuid)
            print_print_document(result)
    except Exception as e:
        print_error(str(e))


@app.command("transfer-print-inbound")
def print_inbound(
    transfer_id: str = typer.Argument(..., help="调拨单ID"),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """打印调拨入库单"""
    try:
        transfer_uuid = UUID(transfer_id)
        with get_client(base_url) as client:
            result = client.print_inbound(transfer_uuid)
            print_print_document(result)
    except Exception as e:
        print_error(str(e))


@app.command("loss-list")
def list_losses(
    warehouse_id: Optional[str] = typer.Option(
        None,
        "--warehouse-id",
        "-w",
        help="仓库ID过滤",
    ),
    transfer_id: Optional[str] = typer.Option(
        None,
        "--transfer-id",
        "-t",
        help="调拨单ID过滤",
    ),
    page: int = typer.Option(1, "--page", help="页码"),
    page_size: int = typer.Option(20, "--page-size", help="每页数量"),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """列出损耗记录"""
    try:
        wh_uuid = UUID(warehouse_id) if warehouse_id else None
        trans_uuid = UUID(transfer_id) if transfer_id else None

        with get_client(base_url) as client:
            result = client.list_losses(
                warehouse_id=wh_uuid,
                transfer_id=trans_uuid,
                page=page,
                page_size=page_size,
            )
            print_loss_list(result)
    except Exception as e:
        print_error(str(e))


@app.command("loss-summary")
def get_loss_summary(
    warehouse_id: Optional[str] = typer.Option(
        None,
        "--warehouse-id",
        "-w",
        help="仓库ID过滤",
    ),
    year: Optional[int] = typer.Option(
        None,
        "--year",
        help="年份",
    ),
    month: Optional[int] = typer.Option(
        None,
        "--month",
        help="月份",
    ),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """查看月度损耗统计"""
    try:
        wh_uuid = UUID(warehouse_id) if warehouse_id else None

        with get_client(base_url) as client:
            result = client.get_loss_summary(
                warehouse_id=wh_uuid,
                year=year,
                month=month,
            )
            print_loss_summary(result)
    except Exception as e:
        print_error(str(e))


@app.command("transaction-list")
def list_transactions(
    warehouse_id: Optional[str] = typer.Option(
        None,
        "--warehouse-id",
        "-w",
        help="仓库ID过滤",
    ),
    product_id: Optional[str] = typer.Option(
        None,
        "--product-id",
        "-p",
        help="商品ID过滤",
    ),
    reference_id: Optional[str] = typer.Option(
        None,
        "--reference-id",
        "-r",
        help="关联ID过滤（如调拨单ID）",
    ),
    page: int = typer.Option(1, "--page", help="页码"),
    page_size: int = typer.Option(20, "--page-size", help="每页数量"),
    base_url: str = typer.Option(
        "http://localhost:8000",
        "--base-url",
        "-u",
        help="服务端地址",
    ),
) -> None:
    """列出库存变动流水"""
    try:
        wh_uuid = UUID(warehouse_id) if warehouse_id else None
        prod_uuid = UUID(product_id) if product_id else None
        ref_uuid = UUID(reference_id) if reference_id else None

        with get_client(base_url) as client:
            result = client.list_transactions(
                warehouse_id=wh_uuid,
                product_id=prod_uuid,
                reference_id=ref_uuid,
                page=page,
                page_size=page_size,
            )
            print_transaction_list(result)
    except Exception as e:
        print_error(str(e))


if __name__ == "__main__":
    app()
