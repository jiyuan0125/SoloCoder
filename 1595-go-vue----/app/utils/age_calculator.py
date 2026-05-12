from datetime import datetime


def calculate_age_months(birth_date: datetime, reference_date: datetime = None) -> int:
    if reference_date is None:
        reference_date = datetime.utcnow()
    
    months = (reference_date.year - birth_date.year) * 12 + (reference_date.month - birth_date.month)
    
    if reference_date.day < birth_date.day:
        months -= 1
    
    return months


def validate_age_range(
    birth_date: datetime,
    min_age_months: int,
    max_age_months: int,
    reference_date: datetime = None
) -> dict:
    age_months = calculate_age_months(birth_date, reference_date)
    is_eligible = min_age_months <= age_months <= max_age_months
    
    message = ""
    if not is_eligible:
        if age_months < min_age_months:
            message = f"年龄不足，当前{age_months}个月，需要至少{min_age_months}个月"
        else:
            message = f"年龄超出范围，当前{age_months}个月，最大允许{max_age_months}个月"
    else:
        message = f"年龄符合要求，当前{age_months}个月"
    
    return {
        "age_months": age_months,
        "is_eligible": is_eligible,
        "message": message
    }
