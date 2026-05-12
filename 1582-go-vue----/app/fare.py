from .models import CardType
from .schemas import FareCalculation


def calculate_base_fare(distance: float) -> int:
    if distance <= 0:
        return 0

    fare = 200

    if distance <= 6:
        return fare

    remaining = distance - 6
    if remaining <= 0:
        return fare

    if distance <= 32:
        additional_units = (remaining + 3) // 4
        fare += int(additional_units * 100)
    else:
        fare_upto_32 = 200 + ((32 - 6 + 3) // 4) * 100
        over_32 = distance - 32
        additional_over_32 = (over_32 + 5) // 6
        fare = int(fare_upto_32 + additional_over_32 * 100)

    return fare


def calculate_fare(distance: float, card_type: CardType) -> FareCalculation:
    base_fare = calculate_base_fare(distance)
    final_fare = base_fare
    discount = "无"

    if card_type == CardType.SINGLE:
        discount = "无"
    elif card_type == CardType.STORED:
        final_fare = round(base_fare * 0.95)
        discount = "九五折"
    elif card_type == CardType.STUDENT:
        final_fare = round(base_fare * 0.5)
        final_fare = max(final_fare, 100)
        discount = "半价(最低1元)"
    elif card_type == CardType.ELDERLY:
        final_fare = 0
        discount = "免费"

    return FareCalculation(
        distance=distance,
        base_fare=base_fare,
        card_type=card_type,
        discount=discount,
        final_fare=final_fare
    )


def calculate_recharge_bonus(card_type: CardType, amount: int) -> int:
    if card_type == CardType.STUDENT:
        return round(amount * 0.1)
    return 0
