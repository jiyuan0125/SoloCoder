from decimal import Decimal, ROUND_HALF_UP, ROUND_CEILING
from typing import Optional, List, Tuple

from shared.models import (
    Package,
    TransportMode,
    ContainerType,
    SplitPackage,
    PackageDetail,
)
from server.config import app_config


def round_decimal(value: Decimal, decimals: int = 2) -> Decimal:
    return value.quantize(Decimal(f"0.{'0' * decimals}"), rounding=ROUND_HALF_UP)


def decimal_ceil(value: Decimal) -> int:
    if value == value.to_integral_value():
        return int(value)
    return int(value.to_integral_value(rounding=ROUND_CEILING))


def get_perimeter_cm(length: Decimal, width: Decimal, height: Decimal) -> Decimal:
    return length + width + height


def get_perimeter_m(length: Decimal, width: Decimal, height: Decimal) -> Decimal:
    cm_to_m = Decimal("0.01")
    return (length + width + height) * cm_to_m


def calculate_volumetric_weight(
    length: Decimal,
    width: Decimal,
    height: Decimal,
    divisor: Decimal,
) -> Decimal:
    volume = length * width * height
    return volume / divisor


def split_package_by_perimeter(
    pkg: Package,
    max_perimeter_m: Decimal,
) -> List[SplitPackage]:
    split_packages: List[SplitPackage] = []
    cm_to_m = Decimal("0.01")
    m_to_cm = Decimal("100")
    max_perimeter_cm = max_perimeter_m * m_to_cm
    
    current_perimeter_cm = get_perimeter_cm(pkg.length_cm, pkg.width_cm, pkg.height_cm)
    
    if current_perimeter_cm <= max_perimeter_cm:
        return split_packages
    
    dimensions = sorted([pkg.length_cm, pkg.width_cm, pkg.height_cm], reverse=True)
    largest = dimensions[0]
    middle = dimensions[1]
    smallest = dimensions[2]
    
    base_perimeter_cm = middle + smallest
    max_single_length_cm = max_perimeter_cm - base_perimeter_cm
    
    if max_single_length_cm <= Decimal("0"):
        return split_packages
    
    num_splits = decimal_ceil(largest / max_single_length_cm)
    weight_per_split = pkg.weight_kg / Decimal(str(num_splits))
    length_per_split = largest / Decimal(str(num_splits))
    
    for i in range(num_splits):
        split_pkg = SplitPackage(
            weight_kg=round_decimal(weight_per_split),
            length_cm=round_decimal(length_per_split),
            width_cm=middle,
            height_cm=smallest,
            base_freight=Decimal("0"),
            total_freight=Decimal("0"),
        )
        split_packages.append(split_pkg)
    
    return split_packages


class FreightCalculator:
    def __init__(self) -> None:
        self.config = app_config

    def _calculate_land_base(self, weight: Decimal) -> Decimal:
        pricing = self.config.pricing
        first_weight_price = pricing.land_first_weight_price
        continue_weight_price = pricing.land_continue_weight_price
        
        if weight <= Decimal("1"):
            return first_weight_price
        
        continue_weight = weight - Decimal("1")
        return first_weight_price + (continue_weight * continue_weight_price)

    def _calculate_air_base(
        self,
        weight: Decimal,
        length: Decimal,
        width: Decimal,
        height: Decimal,
    ) -> Decimal:
        pricing = self.config.pricing
        divisor = pricing.air_volumetric_divisor
        
        volumetric_weight = calculate_volumetric_weight(length, width, height, divisor)
        chargeable_weight = max(weight, volumetric_weight)
        
        base_price_per_kg = Decimal("15.00")
        return chargeable_weight * base_price_per_kg

    def _calculate_sea_base(self, container_type: Optional[ContainerType]) -> Decimal:
        pricing = self.config.pricing
        
        if container_type == ContainerType.CONTAINER_20FT:
            return pricing.sea_20ft_price
        elif container_type == ContainerType.CONTAINER_40FT:
            return pricing.sea_40ft_price
        
        return pricing.sea_20ft_price

    def _calculate_base_freight(
        self,
        transport_mode: TransportMode,
        weight: Decimal,
        length: Decimal,
        width: Decimal,
        height: Decimal,
        container_type: Optional[ContainerType],
    ) -> Decimal:
        if transport_mode == TransportMode.LAND:
            return self._calculate_land_base(weight)
        elif transport_mode == TransportMode.AIR:
            return self._calculate_air_base(weight, length, width, height)
        elif transport_mode == TransportMode.SEA:
            return self._calculate_sea_base(container_type)
        
        return Decimal("0")

    def _calculate_insurance(self, declared_value: Optional[Decimal]) -> Tuple[Decimal, Decimal]:
        pricing = self.config.pricing
        
        if declared_value is None or declared_value <= Decimal("0"):
            return Decimal("0"), Decimal("0")
        
        insurance_fee = declared_value * pricing.insurance_rate
        insurance_fee = max(insurance_fee, pricing.insurance_min_fee)
        
        max_compensation = declared_value
        
        return round_decimal(insurance_fee), max_compensation

    def _calculate_surcharge(
        self,
        base_freight: Decimal,
        origin: str,
        destination: str,
    ) -> Decimal:
        is_origin_remote = self.config.is_remote_area(origin)
        is_dest_remote = self.config.is_remote_area(destination)
        
        if not is_origin_remote and not is_dest_remote:
            return Decimal("0")
        
        surcharge_rate = self.config.remote_areas.surcharge_rate
        surcharge = base_freight * surcharge_rate
        
        return round_decimal(surcharge)

    def _get_max_perimeter(self, transport_mode: TransportMode) -> Optional[Decimal]:
        pricing = self.config.pricing
        
        if transport_mode == TransportMode.LAND:
            return pricing.land_max_perimeter_meters
        elif transport_mode == TransportMode.AIR:
            return pricing.air_max_perimeter_meters
        else:
            return None

    def _calculate_uninsured_compensation(self, base_freight: Decimal) -> Decimal:
        return base_freight * self.config.pricing.max_compensation_multiplier

    def calculate_package(
        self,
        pkg: Package,
        transport_mode: TransportMode,
        discount_rate: Decimal,
    ) -> PackageDetail:
        max_perimeter = self._get_max_perimeter(transport_mode)
        split_packages: Optional[List[SplitPackage]] = None
        total_base_freight = Decimal("0")
        
        if max_perimeter is not None:
            perimeter = get_perimeter_m(pkg.length_cm, pkg.width_cm, pkg.height_cm)
            if perimeter > max_perimeter:
                splits = split_package_by_perimeter(pkg, max_perimeter)
                if splits:
                    split_packages = []
                    for split in splits:
                        split_base = self._calculate_base_freight(
                            transport_mode,
                            split.weight_kg,
                            split.length_cm,
                            split.width_cm,
                            split.height_cm,
                            None,
                        )
                        split_surcharge = self._calculate_surcharge(
                            split_base, pkg.origin, pkg.destination
                        )
                        split_total = split_base + split_surcharge
                        
                        split.base_freight = round_decimal(split_base)
                        split.total_freight = round_decimal(split_total)
                        split_packages.append(split)
                        total_base_freight += split_base
                else:
                    total_base_freight = self._calculate_base_freight(
                        transport_mode,
                        pkg.weight_kg,
                        pkg.length_cm,
                        pkg.width_cm,
                        pkg.height_cm,
                        pkg.container_type,
                    )
            else:
                total_base_freight = self._calculate_base_freight(
                    transport_mode,
                    pkg.weight_kg,
                    pkg.length_cm,
                    pkg.width_cm,
                    pkg.height_cm,
                    pkg.container_type,
                )
        else:
            total_base_freight = self._calculate_base_freight(
                transport_mode,
                pkg.weight_kg,
                pkg.length_cm,
                pkg.width_cm,
                pkg.height_cm,
                pkg.container_type,
            )
        
        total_base_freight = round_decimal(total_base_freight)
        
        surcharge = self._calculate_surcharge(total_base_freight, pkg.origin, pkg.destination)
        insurance_fee, declared_compensation = self._calculate_insurance(pkg.declared_value)
        
        before_discount = total_base_freight + surcharge + insurance_fee
        discount = before_discount * discount_rate
        discount = round_decimal(discount)
        
        final_amount = before_discount - discount
        final_amount = round_decimal(max(final_amount, Decimal("0")))
        
        if pkg.declared_value is not None and pkg.declared_value > Decimal("0"):
            max_compensation = declared_compensation
        else:
            max_compensation = self._calculate_uninsured_compensation(total_base_freight)
        
        return PackageDetail(
            package_index=0,
            base_freight=total_base_freight,
            surcharge=surcharge,
            insurance_fee=insurance_fee,
            discount=discount,
            final_amount=final_amount,
            max_compensation=round_decimal(max_compensation),
            is_split=split_packages is not None and len(split_packages) > 0,
            split_packages=split_packages,
        )


freight_calculator: FreightCalculator = FreightCalculator()
