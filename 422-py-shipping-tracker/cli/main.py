from datetime import datetime, timezone
from typing import Optional

import click
import httpx

from cli.client import ShippingClient
from cli.formatters import (
    print_added_node,
    print_alerts,
    print_batch_query,
    print_created_waybill,
    print_error,
    print_split_response,
    print_success,
    print_waybill_query,
)
from shared.models import (
    CustomStatus,
    CustomStatusUpdateRequest,
    Location,
    NodeType,
    Operator,
    OperatorType,
    SignedBy,
    TrackingNodeCreate,
    WaybillCreate,
    WaybillSplitRequest,
)


def get_client(ctx: click.Context) -> ShippingClient:
    base_url: str = ctx.obj["base_url"]
    return ShippingClient(base_url)


def parse_operator(
    operator_type_str: str, operator_id: str, operator_name: Optional[str]
) -> Operator:
    try:
        operator_type = OperatorType(operator_type_str.lower())
    except ValueError:
        raise click.BadParameter(
            f"无效的操作人类型: {operator_type_str}，有效值: system, courier"
        )
    return Operator(
        operator_type=operator_type,
        operator_id=operator_id,
        operator_name=operator_name,
    )


def parse_node_type(node_type_str: str) -> NodeType:
    node_type_map: dict[str, NodeType] = {
        "picked_up": NodeType.PICKED_UP,
        "in_transit": NodeType.IN_TRANSIT,
        "arrived": NodeType.ARRIVED_AT_TRANSIT,
        "out_for_delivery": NodeType.OUT_FOR_DELIVERY,
        "signed": NodeType.SIGNED,
    }
    try:
        return node_type_map[node_type_str.lower()]
    except KeyError:
        valid_types = ", ".join(node_type_map.keys())
        raise click.BadParameter(
            f"无效的节点类型: {node_type_str}，有效值: {valid_types}"
        )


def parse_custom_status(status_str: str) -> CustomStatus:
    try:
        return CustomStatus(status_str)
    except ValueError:
        valid_statuses = ", ".join([s.value for s in CustomStatus])
        raise click.BadParameter(
            f"无效的海关状态: {status_str}，有效值: {valid_statuses}"
        )


@click.group()
@click.option(
    "--base-url",
    default="http://localhost:8000/api/v1",
    help="服务端 API 地址",
    envvar="SHIPPING_API_URL",
)
@click.pass_context
def app(ctx: click.Context, base_url: str) -> None:
    """物流跟踪系统命令行工具"""
    ctx.ensure_object(dict)
    ctx.obj["base_url"] = base_url


@app.command()
@click.option("--sender", required=True, help="发件人")
@click.option("--receiver", required=True, help="收件人")
@click.option("--address", required=True, help="收件地址")
@click.option("--international", is_flag=True, help="是否国际运单")
@click.option("--operator-type", type=click.Choice(["system", "courier"], case_sensitive=False),
              default="courier", help="操作人类型")
@click.option("--operator-id", required=True, help="操作人ID")
@click.option("--operator-name", help="操作人名称")
@click.pass_context
def create(
    ctx: click.Context,
    sender: str,
    receiver: str,
    address: str,
    international: bool,
    operator_type: str,
    operator_id: str,
    operator_name: Optional[str],
) -> None:
    """创建运单"""
    try:
        client = get_client(ctx)
        operator = parse_operator(operator_type, operator_id, operator_name)
        waybill_create = WaybillCreate(
            sender=sender,
            receiver=receiver,
            receiver_address=address,
            is_international=international,
            operator=operator,
        )
        waybill = client.create_waybill(waybill_create)
        print_created_waybill(waybill)
    except httpx.HTTPStatusError as e:
        error_detail = e.response.json().get("detail", {}) if e.response.content else {}
        error_code = error_detail.get("error_code")
        message = error_detail.get("message", str(e))
        print_error(message, error_code)
    except Exception as e:
        print_error(str(e))


@app.command()
@click.argument("waybill_number")
@click.pass_context
def query(ctx: click.Context, waybill_number: str) -> None:
    """查询运单及物流轨迹"""
    try:
        client = get_client(ctx)
        response = client.get_waybill(waybill_number)
        print_waybill_query(response)
    except httpx.HTTPStatusError as e:
        error_detail = e.response.json().get("detail", {}) if e.response.content else {}
        error_code = error_detail.get("error_code")
        message = error_detail.get("message", str(e))
        print_error(message, error_code)
    except Exception as e:
        print_error(str(e))


@app.command()
@click.argument("waybill_numbers", nargs=-1, required=True)
@click.pass_context
def batch(ctx: click.Context, waybill_numbers: tuple[str, ...]) -> None:
    """批量查询运单（最多50个）"""
    try:
        client = get_client(ctx)
        response = client.batch_query_waybills(list(waybill_numbers))
        print_batch_query(response)
    except httpx.HTTPStatusError as e:
        error_detail = e.response.json().get("detail", {}) if e.response.content else {}
        error_code = error_detail.get("error_code")
        message = error_detail.get("message", str(e))
        print_error(message, error_code)
    except Exception as e:
        print_error(str(e))


@app.command()
@click.argument("waybill_number")
@click.option("--node-type", required=True,
              type=click.Choice(["picked_up", "in_transit", "arrived", "out_for_delivery", "signed"],
                                case_sensitive=False),
              help="节点类型")
@click.option("--location", required=True, help="地点地址")
@click.option("--lat", type=float, help="纬度")
@click.option("--lon", type=float, help="经度")
@click.option("--timestamp", help="UTC 时间 (格式: YYYY-MM-DD HH:MM:SS)，默认当前时间")
@click.option("--operator-type", type=click.Choice(["system", "courier"], case_sensitive=False),
              default="courier", help="操作人类型")
@click.option("--operator-id", required=True, help="操作人ID")
@click.option("--operator-name", help="操作人名称")
@click.option("--signed-by", help="签收人姓名（签收节点必填）")
@click.option("--authorized", is_flag=True, help="是否代签收")
@click.option("--relation", help="与收件人关系（代签收时必填）")
@click.pass_context
def add_node(
    ctx: click.Context,
    waybill_number: str,
    node_type: str,
    location: str,
    lat: Optional[float],
    lon: Optional[float],
    timestamp: Optional[str],
    operator_type: str,
    operator_id: str,
    operator_name: Optional[str],
    signed_by: Optional[str],
    authorized: bool,
    relation: Optional[str],
) -> None:
    """添加物流跟踪节点"""
    try:
        client = get_client(ctx)
        parsed_node_type = parse_node_type(node_type)
        location_obj = Location(
            address=location,
            latitude=lat,
            longitude=lon,
        )
        if timestamp:
            try:
                ts = datetime.strptime(timestamp, "%Y-%m-%d %H:%M:%S").replace(tzinfo=timezone.utc)
            except ValueError:
                raise click.BadParameter("时间格式错误，应为: YYYY-MM-DD HH:MM:SS")
        else:
            ts = datetime.now(timezone.utc)
        operator = parse_operator(operator_type, operator_id, operator_name)
        signed_by_obj: Optional[SignedBy] = None
        if parsed_node_type == NodeType.SIGNED:
            if not signed_by:
                raise click.BadParameter("签收节点必须提供 --signed-by 参数")
            signed_by_obj = SignedBy(
                name=signed_by,
                is_authorized=authorized,
                authorized_relation=relation,
            )
        node_create = TrackingNodeCreate(
            node_type=parsed_node_type,
            location=location_obj,
            timestamp=ts,
            operator=operator,
            signed_by=signed_by_obj,
        )
        node = client.add_tracking_node(waybill_number, node_create)
        print_added_node(node)
    except httpx.HTTPStatusError as e:
        error_detail = e.response.json().get("detail", {}) if e.response.content else {}
        error_code = error_detail.get("error_code")
        message = error_detail.get("message", str(e))
        print_error(message, error_code)
    except Exception as e:
        print_error(str(e))


@app.command()
@click.option("--parent", required=True, help="父运单号")
@click.option("--sub-sender", multiple=True, help="子运单发件人（可多次使用）")
@click.option("--sub-receiver", multiple=True, help="子运单收件人（可多次使用）")
@click.option("--sub-address", multiple=True, help="子运单收件地址（可多次使用）")
@click.option("--sub-international", multiple=True, type=bool, default=[], help="子运单是否国际运单")
@click.option("--operator-type", type=click.Choice(["system", "courier"], case_sensitive=False),
              default="courier", help="操作人类型")
@click.option("--operator-id", required=True, help="操作人ID")
@click.option("--operator-name", help="操作人名称")
@click.pass_context
def split(
    ctx: click.Context,
    parent: str,
    sub_sender: tuple[str, ...],
    sub_receiver: tuple[str, ...],
    sub_address: tuple[str, ...],
    sub_international: tuple[bool, ...],
    operator_type: str,
    operator_id: str,
    operator_name: Optional[str],
) -> None:
    """拆分运单为多个子运单"""
    try:
        if len(sub_sender) != len(sub_receiver) or len(sub_sender) != len(sub_address):
            raise click.BadParameter("子运单的发件人、收件人、地址数量必须一致")
        client = get_client(ctx)
        operator = parse_operator(operator_type, operator_id, operator_name)
        sub_waybills: list[WaybillCreate] = []
        for i in range(len(sub_sender)):
            is_international = bool(sub_international[i]) if i < len(sub_international) else False
            sub_waybills.append(WaybillCreate(
                sender=sub_sender[i],
                receiver=sub_receiver[i],
                receiver_address=sub_address[i],
                is_international=is_international,
                operator=operator,
            ))
        split_request = WaybillSplitRequest(
            parent_waybill_number=parent,
            sub_waybills=sub_waybills,
        )
        response = client.split_waybill(split_request)
        print_split_response(response)
    except httpx.HTTPStatusError as e:
        error_detail = e.response.json().get("detail", {}) if e.response.content else {}
        error_code = error_detail.get("error_code")
        message = error_detail.get("message", str(e))
        print_error(message, error_code)
    except Exception as e:
        print_error(str(e))


@app.command()
@click.argument("waybill_number")
@click.option("--status", required=True, help="海关状态")
@click.option("--operator-type", type=click.Choice(["system", "courier"], case_sensitive=False),
              default="courier", help="操作人类型")
@click.option("--operator-id", required=True, help="操作人ID")
@click.option("--operator-name", help="操作人名称")
@click.pass_context
def custom_status(
    ctx: click.Context,
    waybill_number: str,
    status: str,
    operator_type: str,
    operator_id: str,
    operator_name: Optional[str],
) -> None:
    """更新国际运单的海关状态"""
    try:
        client = get_client(ctx)
        parsed_status = parse_custom_status(status)
        operator = parse_operator(operator_type, operator_id, operator_name)
        request = CustomStatusUpdateRequest(
            custom_status=parsed_status,
            operator=operator,
        )
        waybill = client.update_custom_status(waybill_number, request)
        print_created_waybill(waybill)
    except httpx.HTTPStatusError as e:
        error_detail = e.response.json().get("detail", {}) if e.response.content else {}
        error_code = error_detail.get("error_code")
        message = error_detail.get("message", str(e))
        print_error(message, error_code)
    except Exception as e:
        print_error(str(e))


@app.command()
@click.pass_context
def check_alerts(ctx: click.Context) -> None:
    """检查异常运单（超过72小时无更新）"""
    try:
        client = get_client(ctx)
        alerts = client.check_abnormal_waybills()
        print_alerts(alerts)
    except httpx.HTTPStatusError as e:
        error_detail = e.response.json().get("detail", {}) if e.response.content else {}
        error_code = error_detail.get("error_code")
        message = error_detail.get("message", str(e))
        print_error(message, error_code)
    except Exception as e:
        print_error(str(e))


@app.command()
@click.pass_context
def generate_alerts(ctx: click.Context) -> None:
    """为异常运单生成每日告警"""
    try:
        client = get_client(ctx)
        alerts = client.generate_alerts()
        print_alerts(alerts)
    except httpx.HTTPStatusError as e:
        error_detail = e.response.json().get("detail", {}) if e.response.content else {}
        error_code = error_detail.get("error_code")
        message = error_detail.get("message", str(e))
        print_error(message, error_code)
    except Exception as e:
        print_error(str(e))


@app.command()
@click.pass_context
def health(ctx: click.Context) -> None:
    """检查服务端健康状态"""
    try:
        client = get_client(ctx)
        status = client.health_check()
        print_success(f"服务端状态: {status.get('status', 'unknown')}")
    except httpx.HTTPStatusError as e:
        print_error(f"服务端异常: {e}")
    except Exception as e:
        print_error(f"连接失败: {e}")
