#!/usr/bin/env python3
"""索道运行管理系统测试脚本"""
from datetime import datetime
from app.database import Base, SessionLocal, engine
from app.models import RopewayStatus, Operator, InspectionType, InspectionStatus
from app.services.weather_service import WeatherMonitor, _lag_counters
from app.services.inspection_service import InspectionService
from app.services.passenger_service import PassengerService
from app.services.report_service import ReportService
from app.config import (
    WIND_SPEED_DECELERATE,
    WIND_SPEED_PAUSE,
    WIND_SPEED_STOP,
    LAG_PROTECTION_MINUTES
)

def reset_all():
    _lag_counters.clear()
    Base.metadata.drop_all(bind=engine)
    Base.metadata.create_all(bind=engine)

def test_weather_monitor():
    print("\n" + "="*60)
    print("测试 1: 天气监控与状态机")
    print("="*60)
    
    reset_all()
    db = SessionLocal()
    monitor = WeatherMonitor(db)
    
    print(f"\n当前状态: {monitor.current_status.value}")
    print(f"减速阈值: {WIND_SPEED_DECELERATE} m/s")
    print(f"暂停阈值: {WIND_SPEED_PAUSE} m/s")
    print(f"停止阈值: {WIND_SPEED_STOP} m/s")
    print(f"滞后保护: {LAG_PROTECTION_MINUTES} 分钟")
    
    print(f"\n--- 测试正常风况 (5 m/s) ---")
    weather, status = monitor.update_weather(5.0, False)
    print(f"风速 5.0 m/s -> 状态: {monitor.current_status.value}")
    assert monitor.current_status == RopewayStatus.NORMAL
    
    print(f"\n--- 测试滞后保护 - 连续3次超过减速阈值 ---")
    for i in range(LAG_PROTECTION_MINUTES):
        weather, status = monitor.update_weather(9.0, False)
        print(f"第 {i+1} 次更新 (9.0 m/s) -> 状态: {monitor.current_status.value}")
        if i < LAG_PROTECTION_MINUTES - 1:
            assert monitor.current_status == RopewayStatus.NORMAL
        else:
            assert monitor.current_status == RopewayStatus.DECELERATE
            print(f"✓ 连续 {LAG_PROTECTION_MINUTES} 次超过阈值，状态切换到 DECELERATE")
    
    print(f"\n--- 测试逐级降级 - 不能跳级 ---")
    weather, status = monitor.update_weather(22.0, False)
    print(f"尝试直接跳到 STOP (22.0 m/s) -> 状态: {monitor.current_status.value}")
    assert monitor.current_status == RopewayStatus.DECELERATE
    print("✓ 无法跳级，保持 DECELERATE")
    
    print(f"\n--- 测试继续降级到 PAUSE ---")
    for i in range(LAG_PROTECTION_MINUTES):
        weather, status = monitor.update_weather(15.0, False)
        print(f"第 {i+1} 次更新 (15.0 m/s) -> 状态: {monitor.current_status.value}")
    
    assert monitor.current_status == RopewayStatus.PAUSE
    print(f"✓ 连续 {LAG_PROTECTION_MINUTES} 次，状态切换到 PAUSE")
    
    print(f"\n--- 测试继续降级到 STOP ---")
    for i in range(LAG_PROTECTION_MINUTES):
        weather, status = monitor.update_weather(22.0, False)
        print(f"第 {i+1} 次更新 (22.0 m/s) -> 状态: {monitor.current_status.value}")
    
    assert monitor.current_status == RopewayStatus.STOP
    print(f"✓ 连续 {LAG_PROTECTION_MINUTES} 次，状态切换到 STOP")
    
    print(f"\n--- 测试天气恢复需操作员确认 ---")
    operator = Operator(username="test_operator", full_name="测试操作员")
    db.add(operator)
    db.commit()
    
    print(f"天气恢复正常 (5.0 m/s) 但状态保持: {monitor.current_status.value}")
    monitor.update_weather(5.0, False)
    assert monitor.current_status == RopewayStatus.STOP
    print("✓ 天气恢复后不自动恢复")
    
    record = monitor.confirm_recovery(operator.id, "确认天气安全")
    print(f"操作员确认后状态: {monitor.current_status.value}")
    assert monitor.current_status == RopewayStatus.NORMAL
    print("✓ 操作员确认后恢复正常")
    
    db.close()
    print("\n✓ 天气监控测试通过!")


def test_lightning():
    print("\n" + "="*60)
    print("测试 1b: 雷电检测 (直接停止)")
    print("="*60)
    
    reset_all()
    db = SessionLocal()
    monitor = WeatherMonitor(db)
    
    operator = Operator(username="admin", full_name="管理员")
    db.add(operator)
    db.commit()
    
    print(f"\n--- 测试雷电直接停止 ---")
    weather, status = monitor.update_weather(5.0, False)
    assert monitor.current_status == RopewayStatus.NORMAL
    print(f"初始状态: {monitor.current_status.value}")
    
    weather, status = monitor.update_weather(5.0, True)
    print(f"检测到雷电 -> 状态: {monitor.current_status.value}")
    assert monitor.current_status == RopewayStatus.STOP
    print("✓ 检测到雷电直接停止")
    
    db.close()
    print("\n✓ 雷电检测测试通过!")


def test_inspection_service():
    print("\n" + "="*60)
    print("测试 2: 检修管理服务")
    print("="*60)
    
    reset_all()
    db = SessionLocal()
    service = InspectionService(db)
    
    operator = Operator(username="inspector", full_name="检查员")
    db.add(operator)
    db.commit()
    
    print(f"\n--- 测试紧急问题次日不允许运营 ---")
    
    inspection = service.create_inspection(
        InspectionType.DAILY,
        datetime.now(),
        "日常检查"
    )
    
    service.update_inspection(
        inspection.id,
        issues_found="发现紧急安全问题，需要立即处理",
        is_urgent=True,
        resolved=False
    )
    
    can_operate = service.can_operate_tomorrow()
    print(f"存在未处理紧急问题 -> 次日允许运营: {can_operate}")
    assert not can_operate
    print("✓ 未处理紧急问题，次日不允许运营")
    
    service.update_inspection(
        inspection.id,
        resolved=True,
        completed_date=datetime.now(),
        operator_id=operator.id
    )
    
    can_operate = service.can_operate_tomorrow()
    print(f"问题已解决 -> 次日允许运营: {can_operate}")
    assert can_operate
    print("✓ 问题解决后次日允许运营")
    
    db.close()
    print("\n✓ 检修管理测试通过!")


def test_passenger_service():
    print("\n" + "="*60)
    print("测试 3: 载客量与排队管理")
    print("="*60)
    
    reset_all()
    db = SessionLocal()
    service = PassengerService(db)
    
    print(f"\n--- 测试设置吊厢容量 ---")
    service.set_capacity(100)
    capacity = service.get_active_capacity()
    print(f"吊厢总容量: {capacity}")
    assert capacity == 100
    print("✓ 容量设置成功")
    
    print(f"\n--- 测试乘客记录 (1.2米以下儿童不计入) ---")
    record = service.record_passengers(gondola_id=1, adult_count=5, child_count=3)
    print(f"成人: {record.adult_count}, 儿童: {record.child_count}")
    print(f"计费乘客: {record.total_counted} (儿童不计费)")
    assert record.total_counted == 5
    print("✓ 儿童不计入计费乘客")
    
    print(f"\n--- 测试排队限流 (超过容量3倍限流) ---")
    queue_data, is_limited = service.update_queue(150)
    print(f"排队人数: 150, 限流阈值: {100 * 3}, 是否限流: {is_limited}")
    assert not is_limited
    
    queue_data, is_limited = service.update_queue(350)
    print(f"排队人数: 350, 限流阈值: {100 * 3}, 是否限流: {is_limited}")
    assert is_limited
    print("✓ 超过容量3倍触发限流")
    
    db.close()
    print("\n✓ 载客与排队管理测试通过!")


def test_report_service():
    print("\n" + "="*60)
    print("测试 4: 运营报告与时间统计")
    print("="*60)
    
    reset_all()
    db = SessionLocal()
    service = ReportService(db)
    
    print(f"\n--- 测试生成日报 ---")
    report = service.generate_daily_report()
    print(f"报告日期: {report.report_date}")
    print(f"总运营时间: {report.total_operating_hours} 小时")
    print(f"有效运营时间: {report.effective_operating_hours} 小时")
    print(f"低于最低时长: {report.below_min_hours}")
    
    db.close()
    print("\n✓ 报告生成测试通过!")


def main():
    print("="*60)
    print("索道运行管理系统 - 功能测试")
    print("="*60)
    
    Base.metadata.create_all(bind=engine)
    
    try:
        test_weather_monitor()
        test_lightning()
        test_inspection_service()
        test_passenger_service()
        test_report_service()
        
        print("\n" + "="*60)
        print("✓ 所有测试通过!")
        print("="*60)
    except Exception as e:
        print(f"\n✗ 测试失败: {e}")
        import traceback
        traceback.print_exc()
        raise


if __name__ == "__main__":
    main()
