from collections import defaultdict
from decimal import Decimal
from typing import Optional
from uuid import UUID

from shared.constants.error_codes import ErrorCode
from shared.models.enums import QualificationLevel, QuoteStatus
from shared.models.quote import Quote
from shared.protocols.quote import QuoteComparisonItem, QuoteComparisonReport
from server.services.exceptions import BusinessException
from server.services.purchase_service import PurchaseService
from server.services.quote_service import QuoteService
from server.services.supplier_service import SupplierService


QUALIFICATION_WEIGHTS: dict[QualificationLevel, int] = {
    QualificationLevel.A: 3,
    QualificationLevel.B: 2,
    QualificationLevel.C: 1,
}


class ReportService:
    def __init__(self) -> None:
        self._purchase_service = PurchaseService()
        self._quote_service = QuoteService()
        self._supplier_service = SupplierService()

    def generate_comparison_report(self, purchase_id: UUID) -> QuoteComparisonReport:
        purchase = self._purchase_service.get_purchase(purchase_id)
        quotes = self._quote_service.list_quotes_by_purchase(purchase_id)

        if not purchase.items:
            raise BusinessException(ErrorCode.INVALID_INPUT, "采购需求无商品")

        product_name = purchase.items[0].product_name

        valid_quotes: list[Quote] = []
        for quote in quotes:
            if quote.latest_version is None:
                continue
            if quote.status in (QuoteStatus.SUBMITTED, QuoteStatus.LATE, QuoteStatus.AWARDED, QuoteStatus.LOST):
                valid_quotes.append(quote)

        if not valid_quotes:
            raise BusinessException(ErrorCode.INVALID_INPUT, "暂无有效报价")

        comparison_items: list[QuoteComparisonItem] = []
        prices: list[Decimal] = []

        for quote in valid_quotes:
            assert quote.latest_version is not None
            supplier = self._supplier_service.get_supplier(quote.supplier_id)

            item = QuoteComparisonItem(
                supplier_id=supplier.id,
                supplier_name=supplier.name,
                qualification_level=supplier.qualification_level,
                unit_price=quote.latest_version.unit_price,
                delivery_days=quote.latest_version.delivery_days,
                is_late=quote.latest_version.is_late,
                is_awarded=quote.status == QuoteStatus.AWARDED,
            )
            comparison_items.append(item)
            prices.append(quote.latest_version.unit_price)

        ranked_items = self._rank_quotes(comparison_items)

        price_distribution = self._calculate_price_distribution(prices)

        avg_price = sum(prices) / Decimal(len(prices)) if prices else Decimal("0")
        min_price = min(prices) if prices else Decimal("0")
        max_price = max(prices) if prices else Decimal("0")

        return QuoteComparisonReport(
            purchase_id=purchase.id,
            purchase_title=purchase.title,
            product_name=product_name,
            quotes=ranked_items,
            price_distribution=price_distribution,
            avg_price=avg_price.quantize(Decimal("0.0000")),
            min_price=min_price.quantize(Decimal("0.0000")),
            max_price=max_price.quantize(Decimal("0.0000")),
        )

    def _rank_quotes(self, items: list[QuoteComparisonItem]) -> list[QuoteComparisonItem]:
        def sort_key(item: QuoteComparisonItem) -> tuple[int, Decimal, int]:
            weight = QUALIFICATION_WEIGHTS[item.qualification_level]
            return (-weight, item.unit_price, item.delivery_days)

        sorted_items = sorted(items, key=sort_key)

        for i, item in enumerate(sorted_items):
            item.rank = i + 1

        return sorted_items

    def _calculate_price_distribution(self, prices: list[Decimal]) -> dict[str, int]:
        if not prices:
            return {}

        min_price = min(prices)
        max_price = max(prices)

        if min_price == max_price:
            return {f"{min_price}": len(prices)}

        range_size = (max_price - min_price) / Decimal("4")
        ranges: dict[str, int] = defaultdict(int)

        for price in prices:
            if price == max_price:
                bucket = 3
            else:
                bucket = int((price - min_price) / range_size)
                bucket = min(bucket, 3)

            range_start = min_price + range_size * Decimal(bucket)
            range_end = min_price + range_size * Decimal(bucket + 1)
            range_key = f"{range_start.quantize(Decimal('0.00'))} - {range_end.quantize(Decimal('0.00'))}"
            ranges[range_key] += 1

        return dict(sorted(ranges.items()))
