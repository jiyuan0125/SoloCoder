from typing import Optional
from uuid import UUID

import typer

from shared.models.enums import ReviewResult, SupplierStatus
from shared.protocols.supplier import (
    SupplierApproveRequest,
    SupplierCreateRequest,
    SupplierListResponse,
    SupplierResponse,
    SupplierReviewRequest,
    SupplierUpdateRequest,
)
from cli.client.http_client import get_default_client

app = typer.Typer(help="供应商管理命令")


@app.command(name="create")
def create_supplier(
    name: str = typer.Option(..., "--name", "-n", help="供应商名称"),
    contact: str = typer.Option(..., "--contact", "-c", help="联系人"),
    phone: str = typer.Option(..., "--phone", "-p", help="联系电话"),
    address: str = typer.Option(..., "--address", "-a", help="地址"),
) -> None:
    """创建新供应商（注册）"""
    client = get_default_client()
    request = SupplierCreateRequest(
        name=name,
        contact_person=contact,
        phone=phone,
        address=address,
    )
    try:
        response = client.post("/suppliers", request.model_dump(mode="json"))
        supplier = client.parse_response(response, SupplierResponse)
        typer.secho("✓ 供应商创建成功", fg=typer.colors.GREEN)
        _print_supplier(supplier)
    except Exception as e:
        typer.secho(f"✗ 创建失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="get")
def get_supplier(
    supplier_id: UUID = typer.Argument(..., help="供应商 ID"),
) -> None:
    """获取供应商详情"""
    client = get_default_client()
    try:
        response = client.get(f"/suppliers/{supplier_id}")
        supplier = client.parse_response(response, SupplierResponse)
        _print_supplier(supplier)
    except Exception as e:
        typer.secho(f"✗ 获取失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="list")
def list_suppliers(
    status: Optional[SupplierStatus] = typer.Option(
        None, "--status", "-s", help="按状态筛选 (pending/approved/rejected/suspended)"
    ),
) -> None:
    """列出所有供应商"""
    client = get_default_client()
    try:
        params = {}
        if status is not None:
            params["status"] = status.value
        response = client.get("/suppliers", params)
        result = client.parse_response(response, SupplierListResponse)

        if result.total == 0:
            typer.secho("暂无供应商", fg=typer.colors.YELLOW)
            return

        typer.secho(f"共 {result.total} 个供应商:\n", fg=typer.colors.BLUE)
        for supplier in result.suppliers:
            _print_supplier_short(supplier)
    except Exception as e:
        typer.secho(f"✗ 获取列表失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="update")
def update_supplier(
    supplier_id: UUID = typer.Argument(..., help="供应商 ID"),
    name: Optional[str] = typer.Option(None, "--name", "-n", help="供应商名称"),
    contact: Optional[str] = typer.Option(None, "--contact", "-c", help="联系人"),
    phone: Optional[str] = typer.Option(None, "--phone", "-p", help="联系电话"),
    address: Optional[str] = typer.Option(None, "--address", "-a", help="地址"),
) -> None:
    """更新供应商信息"""
    client = get_default_client()
    request = SupplierUpdateRequest(
        name=name,
        contact_person=contact,
        phone=phone,
        address=address,
    )
    try:
        response = client.put(f"/suppliers/{supplier_id}", request.model_dump(mode="json", exclude_none=True))
        supplier = client.parse_response(response, SupplierResponse)
        typer.secho("✓ 供应商更新成功", fg=typer.colors.GREEN)
        _print_supplier(supplier)
    except Exception as e:
        typer.secho(f"✗ 更新失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="approve")
def approve_supplier(
    supplier_id: UUID = typer.Argument(..., help="供应商 ID"),
    approved: bool = typer.Option(True, "--approved/--rejected", help="是否通过审核"),
    remarks: str = typer.Option("", "--remarks", "-r", help="备注"),
) -> None:
    """审核供应商（通过/拒绝）"""
    client = get_default_client()
    request = SupplierApproveRequest(approved=approved, remarks=remarks)
    try:
        response = client.post(f"/suppliers/{supplier_id}/approve", request.model_dump(mode="json"))
        supplier = client.parse_response(response, SupplierResponse)
        typer.secho(f"✓ 审核完成，状态已更新为: {supplier.status.value}", fg=typer.colors.GREEN)
        _print_supplier(supplier)
    except Exception as e:
        typer.secho(f"✗ 审核失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="suspend")
def suspend_supplier(
    supplier_id: UUID = typer.Argument(..., help="供应商 ID"),
) -> None:
    """暂停供应商资格"""
    client = get_default_client()
    try:
        response = client.post(f"/suppliers/{supplier_id}/suspend")
        supplier = client.parse_response(response, SupplierResponse)
        typer.secho("✓ 供应商已暂停", fg=typer.colors.GREEN)
        _print_supplier(supplier)
    except Exception as e:
        typer.secho(f"✗ 操作失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="reactivate")
def reactivate_supplier(
    supplier_id: UUID = typer.Argument(..., help="供应商 ID"),
) -> None:
    """恢复供应商资格"""
    client = get_default_client()
    try:
        response = client.post(f"/suppliers/{supplier_id}/reactivate")
        supplier = client.parse_response(response, SupplierResponse)
        typer.secho("✓ 供应商已恢复", fg=typer.colors.GREEN)
        _print_supplier(supplier)
    except Exception as e:
        typer.secho(f"✗ 操作失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


@app.command(name="review")
def review_qualification(
    supplier_id: UUID = typer.Argument(..., help="供应商 ID"),
    result: ReviewResult = typer.Option(
        ..., "--result", "-r", help="评审结果 (maintain/upgrade/downgrade)"
    ),
    reviewer: str = typer.Option(..., "--reviewer", "-p", help="评审人"),
    remarks: str = typer.Option("", "--notes", "-n", help="备注"),
) -> None:
    """资质等级评审（维持/升级/降级）"""
    client = get_default_client()
    request = SupplierReviewRequest(result=result, reviewer=reviewer, remarks=remarks)
    try:
        response = client.post(f"/suppliers/{supplier_id}/review", request.model_dump(mode="json"))
        supplier = client.parse_response(response, SupplierResponse)
        typer.secho("✓ 资质评审完成", fg=typer.colors.GREEN)
        _print_supplier(supplier)
    except Exception as e:
        typer.secho(f"✗ 评审失败: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


def _print_supplier(supplier: SupplierResponse) -> None:
    status_color = (
        typer.colors.GREEN
        if supplier.status == SupplierStatus.APPROVED
        else (
            typer.colors.YELLOW
            if supplier.status == SupplierStatus.PENDING
            else (
                typer.colors.RED
                if supplier.status == SupplierStatus.SUSPENDED
                else typer.colors.WHITE
            )
        )
    )
    typer.echo(f"  ID: {supplier.id}")
    typer.echo(f"  名称: {supplier.name}")
    typer.echo(f"  联系人: {supplier.contact_person}")
    typer.echo(f"  电话: {supplier.phone}")
    typer.echo(f"  地址: {supplier.address}")
    typer.echo(f"  资质等级: {supplier.qualification_level.value}")
    typer.secho(f"  状态: {supplier.status.value}", fg=status_color)
    if supplier.last_review_date:
        typer.echo(f"  上次评审: {supplier.last_review_date}")
    if supplier.next_review_date:
        typer.echo(f"  下次评审: {supplier.next_review_date}")


def _print_supplier_short(supplier: SupplierResponse) -> None:
    status_color = (
        typer.colors.GREEN
        if supplier.status == SupplierStatus.APPROVED
        else (
            typer.colors.YELLOW
            if supplier.status == SupplierStatus.PENDING
            else (
                typer.colors.RED
                if supplier.status == SupplierStatus.SUSPENDED
                else typer.colors.WHITE
            )
        )
    )
    typer.echo(
        f"  [{supplier.id}] {supplier.name} "
        f"(等级: {supplier.qualification_level.value}, "
    )
    typer.secho(f"    状态: {supplier.status.value}", fg=status_color)
