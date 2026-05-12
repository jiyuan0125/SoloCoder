#!/usr/bin/env python3
from decimal import Decimal
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from app import models, services, schemas
from app.database import Base

engine = create_engine("sqlite:///:memory:", connect_args={"check_same_thread": False})
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base.metadata.create_all(bind=engine)


def test_freight_calculation():
    print("\n" + "=" * 60)
    print("测试运费计算逻辑")
    print("=" * 60)
    
    from app.utils import calculate_freight, round_weight, calculate_basic_charge_fen, calculate_discount_fen
    
    print("\n1. 重量四舍五入测试:")
    print(f"   0.5吨 -> {round_weight(Decimal('0.5'))} 吨 (预期: 1)")
    print(f"   1.23吨 -> {round_weight(Decimal('1.23'))} 吨 (预期: 1.2)")
    print(f"   1.26吨 -> {round_weight(Decimal('1.26'))} 吨 (预期: 1.3)")
    print(f"   5吨 -> {round_weight(Decimal('5'))} 吨 (预期: 5)")
    
    print("\n2. 基础运费计算 (费率 10000分/吨 = 100元/吨):")
    rate = 10000
    print(f"   1吨: {calculate_basic_charge_fen(Decimal('1'), rate)} 分 = {calculate_basic_charge_fen(Decimal('1'), rate)/100} 元 (预期: 100元)")
    print(f"   5吨: {calculate_basic_charge_fen(Decimal('5'), rate)} 分 = {calculate_basic_charge_fen(Decimal('5'), rate)/100} 元 (预期: 500元)")
    print(f"   10.5吨: {calculate_basic_charge_fen(Decimal('10.5'), rate)} 分 = {calculate_basic_charge_fen(Decimal('10.5'), rate)/100} 元 (预期: 1050元)")
    
    print("\n3. 折扣计算 (基础运费):")
    print(f"   3000元 (300000分): 折扣 {calculate_discount_fen(300000)} 分 (预期: 0)")
    print(f"   6000元 (600000分): 折扣 {calculate_discount_fen(600000)} 分 = {calculate_discount_fen(600000)/100} 元 (预期: 100元 - 5000-6000部分九折)")
    print(f"   12000元 (1200000分): 折扣 {calculate_discount_fen(1200000)} 分 = {calculate_discount_fen(1200000)/100} 元 (预期: 900元 - 5000-10000九折+10000-12000八折)")
    
    print("\n4. 完整运费计算:")
    result = calculate_freight(Decimal('5'), rate, False)
    print(f"   普通货物 5吨: 基础={result['basic_charge_fen']/100}元, 折扣={result['discount_fen']/100}元, 危险品上浮={result['hazardous_surcharge_fen']/100}元, 总计={result['total_charge_fen']/100}元")
    
    result = calculate_freight(Decimal('5'), rate, True)
    print(f"   危险货物 5吨: 基础={result['basic_charge_fen']/100}元, 折扣={result['discount_fen']/100}元, 危险品上浮={result['hazardous_surcharge_fen']/100}元, 总计={result['total_charge_fen']/100}元")


def test_full_workflow():
    print("\n" + "=" * 60)
    print("测试完整业务流程")
    print("=" * 60)
    
    db = TestingSessionLocal()
    
    try:
        print("\n1. 创建车站 (北京-价区1, 上海-价区3, 广州-价区5):")
        s1 = services.create_station(db, schemas.StationCreate(name="北京", price_zone=1))
        s2 = services.create_station(db, schemas.StationCreate(name="上海", price_zone=3))
        s3 = services.create_station(db, schemas.StationCreate(name="广州", price_zone=5))
        print(f"   已创建: {s1.name}(价区{s1.price_zone}), {s2.name}(价区{s2.price_zone}), {s3.name}(价区{s3.price_zone})")
        
        print("\n2. 创建费率 (价区1->3: 100元/吨, 1->5: 150元/吨):")
        r1 = services.create_rate(db, schemas.RateCreate(from_zone=1, to_zone=3, rate_per_ton_fen=10000))
        r2 = services.create_rate(db, schemas.RateCreate(from_zone=1, to_zone=5, rate_per_ton_fen=15000))
        print(f"   已创建费率: 1->3={r1.rate_per_ton_fen/100}元/吨, 1->5={r2.rate_per_ton_fen/100}元/吨")
        
        print("\n3. 创建货票:")
        wb1 = services.create_waybill(db, schemas.WaybillCreate(
            from_station_name="北京",
            to_station_name="上海",
            weight_ton=Decimal("5"),
            is_hazardous=False,
            customer_code="CUST001",
        ))
        print(f"   货票1: {wb1.waybill_no} | 北京->上海 | 5吨 | 普通 | 客户CUST001 | 总运费 {wb1.total_charge_fen/100}元")
        
        wb2 = services.create_waybill(db, schemas.WaybillCreate(
            from_station_name="北京",
            to_station_name="广州",
            weight_ton=Decimal("8"),
            is_hazardous=True,
            customer_code="CUST001",
        ))
        print(f"   货票2: {wb2.waybill_no} | 北京->广州 | 8吨 | 危险品 | 客户CUST001 | 总运费 {wb2.total_charge_fen/100}元")
        
        wb3 = services.create_waybill(db, schemas.WaybillCreate(
            from_station_name="北京",
            to_station_name="上海",
            weight_ton=Decimal("0.5"),
            is_hazardous=False,
            customer_code=None,
        ))
        print(f"   货票3: {wb3.waybill_no} | 北京->上海 | 0.5吨(按1吨算) | 普通 | 无客户 | 总运费 {wb3.total_charge_fen/100}元")
        
        print("\n4. 查询货票列表 (客户CUST001):")
        waybills = services.list_waybills(db, customer_code="CUST001")
        for wb in waybills:
            print(f"   {wb.waybill_no} | {wb.from_station_name}->{wb.to_station_name} | {float(wb.weight_ton)}吨 | {wb.status} | {wb.total_charge_fen/100}元")
        
        print("\n5. 调整费率 (价区1->3: 100->120元/吨):")
        old_rate = services.get_current_rate(db, 1, 3)
        print(f"   调整前: {old_rate.rate_per_ton_fen/100}元/吨")
        new_rate = services.update_rate(db, old_rate.id, schemas.RateUpdate(rate_per_ton_fen=12000))
        print(f"   调整后: {new_rate.rate_per_ton_fen/100}元/吨")
        
        print("\n6. 结算货票 (测试费率调整差额计算):")
        print(f"   结算前货票1运费: {wb1.total_charge_fen/100}元")
        settlement = services.settle_waybill(db, wb1.id)
        print(f"   结算结果: 旧运费={settlement.old_total_charge_fen/100}元, 新运费={settlement.new_total_charge_fen/100}元, 差额={settlement.adjustment_fen/100}元, 是否调整={settlement.has_adjustment}")
        
        settlement2 = services.settle_waybill(db, wb2.id)
        print(f"   货票2结算: 旧运费={settlement2.old_total_charge_fen/100}元, 新运费={settlement2.new_total_charge_fen/100}元, 差额={settlement2.adjustment_fen/100}元")
        
        print("\n7. 月度对账报表:")
        from datetime import datetime
        now = datetime.now()
        report = services.get_monthly_report(db, now.year, now.month)
        print(f"   月份: {report.year_month}")
        print(f"   总票数: {report.total_count}")
        print(f"   总重量: {float(report.total_weight_ton)}吨")
        print(f"   总运费: {report.total_amount_fen/100}元")
        print(f"   已结算: {report.settled_amount_fen/100}元")
        print(f"   未结算: {report.unsettled_amount_fen/100}元")
        
        print("\n8. 客户月度统计:")
        stats = services.get_customer_monthly_stats(db, now.year, now.month)
        for stat in stats:
            print(f"   {stat.customer_code}: 票数={stat.total_count}, 总金额={stat.total_amount_fen/100}元, 已结={stat.settled_amount_fen/100}元")
        
        print("\n9. 导出CSV:")
        csv_content = services.export_waybills_to_csv(waybills + [wb3])
        lines = csv_content.split('\n')
        for line in lines[:5]:
            print(f"   {line}")
        
        print("\n" + "=" * 60)
        print("所有测试完成!")
        print("=" * 60)
        
    finally:
        db.close()


if __name__ == "__main__":
    import os
    if os.path.exists("./test_waybill.db"):
        os.remove("./test_waybill.db")
    
    test_freight_calculation()
    test_full_workflow()
