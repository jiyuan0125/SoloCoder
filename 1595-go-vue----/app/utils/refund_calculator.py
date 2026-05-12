from datetime import datetime, timedelta


def calculate_refund_amount(
    ticket_price: int,
    performance_start_time: datetime,
    refund_request_time: datetime = None
) -> dict:
    if refund_request_time is None:
        refund_request_time = datetime.utcnow()
    
    time_to_performance = performance_start_time - refund_request_time
    hours_to_performance = time_to_performance.total_seconds() / 3600
    
    refund_percent = 0
    message = ""
    
    if hours_to_performance >= 48:
        refund_percent = 1.0
        message = "演出开始前48小时以上，全额退款"
    elif 24 <= hours_to_performance < 48:
        refund_percent = 0.8
        message = "演出开始前24-48小时，退款80%"
    elif 0 < hours_to_performance < 24:
        refund_percent = 0.5
        message = "演出开始前24小时内，退款50%"
    else:
        refund_percent = 0.0
        message = "演出当天或已结束，不予退款"
    
    refund_amount = int(ticket_price * refund_percent)
    
    return {
        "refund_amount": refund_amount,
        "refund_percent": refund_percent,
        "message": message
    }
