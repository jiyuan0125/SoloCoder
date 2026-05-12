import random
from datetime import datetime
from decimal import Decimal, ROUND_HALF_UP


def generate_waybill_no() -> str:
    timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
    random_num = random.randint(0, 9999)
    return f"HP{timestamp}{random_num:04d}"


def round_weight(weight: Decimal) -> Decimal:
    if weight < Decimal("1"):
        return Decimal("1")
    
    return weight.quantize(Decimal("0.1"), rounding=ROUND_HALF_UP)


def calculate_basic_charge_fen(weight: Decimal, rate_per_ton_fen: int) -> int:
    rounded_weight = round_weight(weight)
    weight_int = int(rounded_weight * Decimal("10"))
    return (weight_int * rate_per_ton_fen) // 10


def calculate_discount_fen(basic_charge_fen: int) -> int:
    total_yuan = basic_charge_fen / 100
    discount_fen = 0
    
    if total_yuan > 10000:
        above_10000 = total_yuan - 10000
        discount_fen += int(above_10000 * 100 * 0.2)
        total_yuan = 10000
    
    if total_yuan > 5000:
        above_5000 = total_yuan - 5000
        discount_fen += int(above_5000 * 100 * 0.1)
    
    return discount_fen


def calculate_freight(weight: Decimal, rate_per_ton_fen: int, is_hazardous: bool):
    basic_charge_fen = calculate_basic_charge_fen(weight, rate_per_ton_fen)
    discount_fen = calculate_discount_fen(basic_charge_fen)
    basic_after_discount = basic_charge_fen - discount_fen
    
    hazardous_surcharge_fen = 0
    if is_hazardous:
        hazardous_surcharge_fen = int(basic_charge_fen * 0.3)
    
    total_charge_fen = basic_after_discount + hazardous_surcharge_fen
    
    return {
        "basic_charge_fen": basic_charge_fen,
        "discount_fen": discount_fen,
        "hazardous_surcharge_fen": hazardous_surcharge_fen,
        "total_charge_fen": total_charge_fen,
    }
