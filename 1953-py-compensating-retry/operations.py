from typing import Dict, Any, Callable, Tuple
import asyncio
import random


async def execute_deduct_inventory(params: Dict[str, Any]) -> Dict[str, Any]:
    await asyncio.sleep(0.1)
    product_id = params.get("product_id")
    quantity = params.get("quantity", 1)
    if random.random() < 0.1:
        raise Exception(f"库存服务暂时不可用，无法扣减产品 {product_id} 的库存")
    return {
        "message": f"成功扣减产品 {product_id} 的库存 {quantity} 件",
        "product_id": product_id,
        "quantity": quantity,
        "operation": "deduct_inventory"
    }


async def compensate_deduct_inventory(params: Dict[str, Any]) -> Dict[str, Any]:
    await asyncio.sleep(0.1)
    product_id = params.get("product_id")
    quantity = params.get("quantity", 1)
    if random.random() < 0.1:
        raise Exception(f"库存服务暂时不可用，无法回滚产品 {product_id} 的库存")
    return {
        "message": f"成功回滚产品 {product_id} 的库存 {quantity} 件",
        "product_id": product_id,
        "quantity": quantity,
        "operation": "compensate_deduct_inventory"
    }


async def execute_deduct_balance(params: Dict[str, Any]) -> Dict[str, Any]:
    await asyncio.sleep(0.1)
    user_id = params.get("user_id")
    amount = params.get("amount", 0)
    if random.random() < 0.1:
        raise Exception(f"余额服务暂时不可用，无法扣减用户 {user_id} 的余额")
    return {
        "message": f"成功扣减用户 {user_id} 的余额 ¥{amount}",
        "user_id": user_id,
        "amount": amount,
        "operation": "deduct_balance"
    }


async def compensate_deduct_balance(params: Dict[str, Any]) -> Dict[str, Any]:
    await asyncio.sleep(0.1)
    user_id = params.get("user_id")
    amount = params.get("amount", 0)
    if random.random() < 0.1:
        raise Exception(f"余额服务暂时不可用，无法回滚用户 {user_id} 的余额")
    return {
        "message": f"成功回滚用户 {user_id} 的余额 ¥{amount}",
        "user_id": user_id,
        "amount": amount,
        "operation": "compensate_deduct_balance"
    }


async def execute_send_coupon(params: Dict[str, Any]) -> Dict[str, Any]:
    await asyncio.sleep(0.1)
    user_id = params.get("user_id")
    coupon_id = params.get("coupon_id")
    if random.random() < 0.1:
        raise Exception(f"优惠券服务暂时不可用，无法给用户 {user_id} 发送优惠券")
    return {
        "message": f"成功给用户 {user_id} 发送优惠券 {coupon_id}",
        "user_id": user_id,
        "coupon_id": coupon_id,
        "operation": "send_coupon"
    }


async def compensate_send_coupon(params: Dict[str, Any]) -> Dict[str, Any]:
    await asyncio.sleep(0.1)
    user_id = params.get("user_id")
    coupon_id = params.get("coupon_id")
    if random.random() < 0.1:
        raise Exception(f"优惠券服务暂时不可用，无法撤销用户 {user_id} 的优惠券")
    return {
        "message": f"成功撤销用户 {user_id} 的优惠券 {coupon_id}",
        "user_id": user_id,
        "coupon_id": coupon_id,
        "operation": "compensate_send_coupon"
    }


async def execute_send_notification(params: Dict[str, Any]) -> Dict[str, Any]:
    await asyncio.sleep(0.1)
    user_id = params.get("user_id")
    message = params.get("message", "订单创建成功")
    if random.random() < 0.1:
        raise Exception(f"通知服务暂时不可用，无法给用户 {user_id} 发送通知")
    return {
        "message": f"成功给用户 {user_id} 发送通知: {message}",
        "user_id": user_id,
        "notification": message,
        "operation": "send_notification"
    }


async def compensate_send_notification(params: Dict[str, Any]) -> Dict[str, Any]:
    await asyncio.sleep(0.1)
    return {
        "message": "通知类操作无需补偿",
        "operation": "compensate_send_notification",
        "note": "通知是单向操作，不需要回滚"
    }


OPERATION_EXECUTORS: Dict[str, Callable] = {
    "deduct_inventory": execute_deduct_inventory,
    "deduct_balance": execute_deduct_balance,
    "send_coupon": execute_send_coupon,
    "send_notification": execute_send_notification,
}

OPERATION_COMPENSATORS: Dict[str, Callable] = {
    "deduct_inventory": compensate_deduct_inventory,
    "deduct_balance": compensate_deduct_balance,
    "send_coupon": compensate_send_coupon,
    "send_notification": compensate_send_notification,
}


def get_executor(operation_type: str) -> Callable:
    executor = OPERATION_EXECUTORS.get(operation_type)
    if not executor:
        raise ValueError(f"不支持的操作类型: {operation_type}")
    return executor


def get_compensator(operation_type: str) -> Callable:
    compensator = OPERATION_COMPENSATORS.get(operation_type)
    if not compensator:
        raise ValueError(f"不支持的补偿操作类型: {operation_type}")
    return compensator


def validate_operation_type(operation_type: str) -> bool:
    return operation_type in OPERATION_EXECUTORS
