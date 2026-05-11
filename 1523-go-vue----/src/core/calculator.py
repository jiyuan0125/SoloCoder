from .models import GradeCalculationInput, GradeCalculationResult


def calculate_recovery(
    feed_grade: float,
    concentrate_grade: float,
    tailings_grade: float,
) -> float:
    denominator = concentrate_grade - tailings_grade
    if denominator == 0:
        return 0.0
    numerator = concentrate_grade * (feed_grade - tailings_grade)
    denominator = feed_grade * (concentrate_grade - tailings_grade)
    if denominator == 0:
        return 0.0
    recovery = (numerator / denominator) * 100
    return recovery


def calculate_concentrate_yield(
    feed_grade: float,
    concentrate_grade: float,
    tailings_grade: float,
    feed_throughput: float,
) -> float:
    denominator = concentrate_grade - tailings_grade
    if denominator == 0:
        return 0.0
    yield_ratio = (feed_grade - tailings_grade) / denominator
    return yield_ratio * feed_throughput


class GradeCalculator:
    @staticmethod
    def calculate(
        input_data: GradeCalculationInput,
    ) -> GradeCalculationResult:
        recovery = calculate_recovery(
            input_data.feed_grade,
            input_data.concentrate_grade,
            input_data.tailings_grade,
        )
        result = GradeCalculationResult(theoretical_recovery=recovery)
        if input_data.feed_throughput is not None:
            yield_amount = calculate_concentrate_yield(
                input_data.feed_grade,
                input_data.concentrate_grade,
                input_data.tailings_grade,
                input_data.feed_throughput,
            )
            result.theoretical_concentrate_yield = yield_amount
        return result
