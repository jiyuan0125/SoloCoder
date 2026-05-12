#!/usr/bin/env python3
"""修复测试：风速突然跳到高危级别时的逐级降级逻辑"""
from datetime import datetime
from app.database import Base, SessionLocal, engine
from app.models import RopewayStatus, Operator
from app.services.weather_service import WeatherMonitor, _lag_counters, _pending_confirmation
from app.config import (
    WIND_SPEED_DECELERATE,
    WIND_SPEED_PAUSE,
    WIND_SPEED_STOP,
    LAG_PROTECTION_MINUTES
)

def reset_state():
    _lag_counters.clear()
    global _pending_confirmation
    _pending_confirmation = False
    Base.metadata.drop_all(bind=engine)
    Base.metadata.create_all(bind=engine)

def test_sudden_high_wind():
    print("\n" + "="*70)
    print("测试场景1: 风速突然从正常跳到25m/s（高危级别）")
    print("="*70)
    print(f"减速阈值: {WIND_SPEED_DECELERATE} m/s")
    print(f"暂停阈值: {WIND_SPEED_PAUSE} m/s")
    print(f"停止阈值: {WIND_SPEED_STOP} m/s")
    print(f"滞后保护: 连续 {LAG_PROTECTION_MINUTES} 次")
    print()
    
    reset_state()
    db = SessionLocal()
    monitor = WeatherMonitor(db)
    
    print(f"初始状态: {monitor.current_status.value}")
    assert monitor.current_status == RopewayStatus.NORMAL
    
    print(f"\n--- 发送10次天气更新，风速=25 m/s ---")
    print(f"预期行为: NORMAL → DECELERATE → PAUSE → STOP (逐级降级)")
    print()
    
    for i in range(1, 11):
        weather, status = monitor.update_weather(25.0, False)
        print(f"第 {i:2d} 次更新 -> 状态: {monitor.current_status.value:10s}  (滞后计数: {_lag_counters.get('global', 0)})")
        
        if i == 3:
            assert monitor.current_status == RopewayStatus.DECELERATE, f"第3次应切到DECELERATE，但当前是 {monitor.current_status.value}"
            print("  ✓ 第3次 -> DECELERATE (超过8m/s，3次滞后保护)")
        elif i == 6:
            assert monitor.current_status == RopewayStatus.PAUSE, f"第6次应切到PAUSE，但当前是 {monitor.current_status.value}"
            print("  ✓ 第6次 -> PAUSE (超过14m/s，3次滞后保护)")
        elif i == 9:
            assert monitor.current_status == RopewayStatus.STOP, f"第9次应切到STOP，但当前是 {monitor.current_status.value}"
            print("  ✓ 第9次 -> STOP (超过20m/s，3次滞后保护)")
    
    print()
    assert monitor.current_status == RopewayStatus.STOP
    print("✓ 测试通过: 风速25m/s逐级降级成功！")
    db.close()

def test_pending_confirmation():
    print("\n" + "="*70)
    print("测试场景2: pending_confirmation 标记")
    print("="*70)
    
    reset_state()
    db = SessionLocal()
    monitor = WeatherMonitor(db)
    
    operator = Operator(username="test_op", full_name="测试员")
    db.add(operator)
    db.commit()
    
    print(f"\n--- 让系统进入 STOP 状态 ---")
    for i in range(12):
        monitor.update_weather(25.0, False)
    
    assert monitor.current_status == RopewayStatus.STOP
    print(f"当前状态: {monitor.current_status.value}")
    print(f"等待确认: {monitor.pending_confirmation}")
    assert monitor.pending_confirmation == False
    
    print(f"\n--- 风速恢复正常 (5m/s) ---")
    weather, status = monitor.update_weather(5.0, False)
    print(f"当前状态: {monitor.current_status.value} (应保持STOP)")
    print(f"等待确认: {monitor.pending_confirmation}")
    
    assert monitor.current_status == RopewayStatus.STOP
    assert monitor.pending_confirmation == True
    print("✓ 天气恢复后状态保持STOP，pending_confirmation=True")
    
    print(f"\n--- 操作员确认恢复 ---")
    record = monitor.confirm_recovery(operator.id, "确认安全")
    assert record is not None
    assert monitor.current_status == RopewayStatus.NORMAL
    assert monitor.pending_confirmation == False
    print(f"当前状态: {monitor.current_status.value}")
    print(f"等待确认: {monitor.pending_confirmation}")
    print("✓ 操作员确认后恢复NORMAL，pending_confirmation=False")
    
    db.close()
    print("\n✓ 测试通过: pending_confirmation 标记正确！")

def test_step_by_step_downgrade():
    print("\n" + "="*70)
    print("测试场景3: 逐步增加风速验证逐级降级")
    print("="*70)
    
    reset_state()
    db = SessionLocal()
    monitor = WeatherMonitor(db)
    
    print(f"\n阶段1: 风速 = 10 m/s (超过8m/s，应到DECELERATE)")
    for i in range(5):
        monitor.update_weather(10.0, False)
        print(f"  第{i+1}次 -> {monitor.current_status.value}")
    
    assert monitor.current_status == RopewayStatus.DECELERATE
    print("✓ 阶段1完成: DECELERATE")
    
    print(f"\n阶段2: 风速 = 16 m/s (超过14m/s，应到PAUSE)")
    for i in range(5):
        monitor.update_weather(16.0, False)
        print(f"  第{i+1}次 -> {monitor.current_status.value}")
    
    assert monitor.current_status == RopewayStatus.PAUSE
    print("✓ 阶段2完成: PAUSE")
    
    print(f"\n阶段3: 风速 = 22 m/s (超过20m/s，应到STOP)")
    for i in range(5):
        monitor.update_weather(22.0, False)
        print(f"  第{i+1}次 -> {monitor.current_status.value}")
    
    assert monitor.current_status == RopewayStatus.STOP
    print("✓ 阶段3完成: STOP")
    
    db.close()
    print("\n✓ 测试通过: 逐级降级逻辑正确！")

def main():
    print("="*70)
    print("天气监控状态切换逻辑 - 修复验证测试")
    print("="*70)
    
    Base.metadata.create_all(bind=engine)
    
    try:
        test_sudden_high_wind()
        test_pending_confirmation()
        test_step_by_step_downgrade()
        
        print("\n" + "="*70)
        print("✓ 所有修复验证测试通过!")
        print("="*70)
    except Exception as e:
        print(f"\n✗ 测试失败: {e}")
        import traceback
        traceback.print_exc()
        raise

if __name__ == "__main__":
    main()
