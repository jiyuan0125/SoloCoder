#!/usr/bin/env python3
import sys
sys.path.insert(0, 'src')

from datetime import date, timedelta
from src.core import (
    FarmManager, BatchCreate, BreedingCreate, BreedingUpdate,
    VaccinationCreate, SlaughterCreate,
    BatchStatus, BreedingStatus,
    SlaughterExceedsStockError, DuplicateBreedingError, BatchEmptyError
)


def test_core_functionality():
    print("=" * 60)
    print("养殖场数字化管理系统 - 核心功能测试")
    print("=" * 60)
    
    manager = FarmManager()
    
    print("\n1. 测试品种管理")
    breeds = manager.get_breeds()
    print(f"  已加载 {len(breeds)} 个默认品种")
    for breed in breeds[:3]:
        print(f"    - {breed.name}: 妊娠周期 {breed.pregnancy_days} 天")
    
    print("\n2. 测试批次管理")
    batch = manager.create_batch(BatchCreate(
        breed_id="breed_001",
        entry_date=date.today(),
        entry_quantity=100,
        notes="测试批次"
    ))
    print(f"  创建批次成功: ID={batch.id[:8]}..., 品种={batch.breed_id}, 存栏={batch.current_stock}")
    
    batches = manager.get_batches()
    print(f"  当前批次数量: {len(batches)}")
    
    print("\n3. 测试繁育管理")
    today = date.today()
    breeding = manager.create_breeding(BreedingCreate(
        batch_id=batch.id,
        female_id="F001",
        breeding_date=today,
        sire_id="M001"
    ))
    print(f"  创建配种记录成功: ID={breeding.id[:8]}...")
    print(f"    母畜: {breeding.female_id}, 配种日期: {breeding.breeding_date}")
    print(f"    预计产仔日期: {breeding.expected_birth_date}")
    
    try:
        manager.create_breeding(BreedingCreate(
            batch_id=batch.id,
            female_id="F001",
            breeding_date=today
        ))
        print("  ❌ 错误: 同一母畜同一天重复配种应该被拒绝")
    except DuplicateBreedingError as e:
        print(f"  ✅ 正确拒绝重复配种: {e.message}")
    
    print("\n4. 测试防疫管理")
    vaccination = manager.create_vaccination(VaccinationCreate(
        batch_id=batch.id,
        vaccine_name="猪瘟疫苗",
        vaccination_date=today,
        interval_days=30,
        type="routine"
    ))
    print(f"  创建防疫记录成功: ID={vaccination.id[:8]}...")
    print(f"    疫苗: {vaccination.vaccine_name}")
    print(f"    下次接种日期: {vaccination.next_vaccination_date}")
    
    print("\n5. 测试出栏管理")
    slaughter = manager.create_slaughter(SlaughterCreate(
        batch_id=batch.id,
        slaughter_date=today,
        quantity=50,
        average_weight=100.5,
        unit_price=15.0
    ))
    print(f"  创建出栏记录成功: ID={slaughter.id[:8]}...")
    print(f"    数量: {slaughter.quantity}, 平均体重: {slaughter.average_weight}kg")
    print(f"    单价: {slaughter.unit_price}元/kg, 总产值: {slaughter.total_value}元")
    
    updated_batch = manager.get_batch(batch.id)
    print(f"  批次存栏更新: {batch.current_stock} -> {updated_batch.current_stock}")
    
    try:
        manager.create_slaughter(SlaughterCreate(
            batch_id=batch.id,
            slaughter_date=today,
            quantity=100,
            average_weight=100.0,
            unit_price=15.0
        ))
        print("  ❌ 错误: 出栏数量超过存栏应该被拒绝")
    except SlaughterExceedsStockError as e:
        print(f"  ✅ 正确拒绝超额出栏: {e.message}")
    
    print("\n6. 测试负价格处理")
    slaughter_zero_price = manager.create_slaughter(SlaughterCreate(
        batch_id=batch.id,
        slaughter_date=today,
        quantity=10,
        average_weight=100.0,
        unit_price=0.0
    ))
    print(f"  零单价出栏记录: 总产值={slaughter_zero_price.total_value}")
    
    print("\n7. 测试批次清空")
    final_batch = manager.get_batch(batch.id)
    remaining = final_batch.current_stock
    if remaining > 0:
        manager.create_slaughter(SlaughterCreate(
            batch_id=batch.id,
            slaughter_date=today,
            quantity=remaining,
            average_weight=100.0,
            unit_price=15.0
        ))
    
    emptied_batch = manager.get_batch(batch.id)
    print(f"  批次状态: {emptied_batch.status}, 存栏: {emptied_batch.current_stock}")
    
    try:
        manager.create_breeding(BreedingCreate(
            batch_id=batch.id,
            female_id="F002",
            breeding_date=today
        ))
        print("  ❌ 错误: 清空批次应该不能新增配种记录")
    except BatchEmptyError as e:
        print(f"  ✅ 正确拒绝清空批次操作: {e.message}")
    
    print("\n8. 测试提醒生成")
    manager.create_breeding(BreedingCreate(
        batch_id=manager.create_batch(BatchCreate(
            breed_id="breed_001",
            entry_date=date.today(),
            entry_quantity=10
        )).id,
        female_id="F003",
        breeding_date=date.today() - timedelta(days=114 - 5)
    ))
    
    reminders = manager.generate_reminders()
    print(f"  生成的提醒数量: {len(reminders)}")
    for reminder in reminders[:2]:
        print(f"    - {reminder.type}: {reminder.message[:50]}...")
    
    print("\n9. 测试指标看板")
    metrics = manager.get_dashboard_metrics()
    print(f"  总存栏: {metrics.total_stock}")
    print(f"  本月入栏: {metrics.monthly_entry}")
    print(f"  本月出栏: {metrics.monthly_slaughter}")
    print(f"  活跃批次: {metrics.active_batches}")
    print(f"  品种分布: {metrics.breed_distribution}")
    print(f"  待处理提醒: {metrics.pending_reminders}")
    
    print("\n" + "=" * 60)
    print("所有核心功能测试完成！")
    print("=" * 60)


if __name__ == "__main__":
    test_core_functionality()
