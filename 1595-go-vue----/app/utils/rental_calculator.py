from datetime import datetime


def calculate_rental_cost(
    venue,
    start_time: datetime,
    end_time: datetime
) -> dict:
    duration_hours = (end_time - start_time).total_seconds() / 3600
    
    start_hour = start_time.hour
    end_hour = end_time.hour
    
    crosses_noon = start_hour < 12 <= end_hour
    crosses_six_pm = start_hour < 18 <= end_hour
    
    starts_before_noon = start_hour < 12
    starts_after_noon = 12 <= start_hour < 18
    starts_after_six = start_hour >= 18
    
    ends_before_noon = end_hour <= 12
    ends_after_noon = 12 < end_hour <= 18
    ends_after_six = end_hour > 18
    
    use_half_day = False
    use_full_day = False
    
    if crosses_noon and crosses_six_pm:
        use_full_day = True
        total_amount = venue.full_day_rate
        billing_type = "全天计费"
    elif crosses_noon or crosses_six_pm:
        if starts_before_noon and ends_after_noon and not crosses_six_pm:
            use_half_day = True
            if starts_before_noon and ends_after_noon:
                total_amount = venue.half_day_rate
                billing_type = "半天计费（上午+下午）"
            elif starts_after_noon and crosses_six_pm:
                total_amount = venue.half_day_rate
                billing_type = "半天计费（下午+晚上）"
            else:
                total_amount = venue.half_day_rate
                billing_type = "半天计费"
        else:
            use_half_day = True
            total_amount = venue.half_day_rate
            billing_type = "半天计费"
    else:
        hourly_amount = int(duration_hours * venue.hourly_rate)
        half_day_amount = venue.half_day_rate
        full_day_amount = venue.full_day_rate
        
        total_amount = hourly_amount
        billing_type = "小时计费"
    
    if not use_half_day and not use_full_day:
        if total_amount > venue.half_day_rate:
            total_amount = venue.half_day_rate
            billing_type = "半天计费（小时费超过半天费）"
        if total_amount > venue.full_day_rate:
            total_amount = venue.full_day_rate
            billing_type = "全天计费（小时费超过全天费）"
    
    return {
        "total_amount": total_amount,
        "billing_type": billing_type,
        "duration_hours": duration_hours,
        "crosses_noon": crosses_noon,
        "crosses_six_pm": crosses_six_pm
    }
