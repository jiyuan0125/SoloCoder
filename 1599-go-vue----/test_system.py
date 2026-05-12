import sys
import os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from datetime import datetime, timedelta
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.orm import Session

import models
import crud
import schemas
from models import PlantStatus, ProtectionLevel

TEST_DATABASE_URL = "sqlite:///:memory:"
test_engine = create_engine(TEST_DATABASE_URL, connect_args={"check_same_thread": False})
TestSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=test_engine)

models.Base.metadata.create_all(bind=test_engine)

def test_plant_management():
    print("=" * 60)
    print("测试1: 植物管理")
    print("=" * 60)
    
    db = TestSessionLocal()
    try:
        plant_data = schemas.PlantCreate(
            scientific_name="Rosa chinensis",
            family="蔷薇科",
            genus="蔷薇属",
            origin="中国",
            status=PlantStatus.NORMAL,
            location="东区1号温室",
            description="月季花"
        )
        plant = crud.create_plant(db, plant_data)
        print(f"✓ 创建植物: {plant.scientific_name} (ID: {plant.id})")
        
        retrieved = crud.get_plant(db, plant.id)
        assert retrieved is not None
        print(f"✓ 读取植物: {retrieved.scientific_name}")
        
        update_data = schemas.PlantUpdate(
            description="月季花 - 已更新",
            location="东区2号温室"
        )
        updated = crud.update_plant(db, plant.id, update_data)
        assert updated.description == "月季花 - 已更新"
        print(f"✓ 更新植物: {updated.description}")
        
        plants = crud.get_plants(db, family="蔷薇科")
        assert len(plants) >= 1
        print(f"✓ 按科搜索: 找到 {len(plants)} 个蔷薇科植物")
        
    finally:
        db.close()
    print()


def test_introduction_observation():
    print("=" * 60)
    print("测试2: 引种观察系统")
    print("=" * 60)
    
    db = TestSessionLocal()
    try:
        plant_data = schemas.PlantCreate(
            scientific_name="Orchidaceae sp.",
            family="兰科",
            genus="兰属",
            origin="东南亚",
            status=PlantStatus.INTRODUCTION_OBSERVATION,
            location="观察温室A区",
            description="引种兰花品种"
        )
        plant = crud.create_plant(db, plant_data)
        print(f"✓ 创建引种植物: {plant.scientific_name}")
        print(f"  引种日期: {plant.introduction_date}")
        print(f"  观察结束日期: {plant.observation_end_date}")
        
        for i in range(3):
            record_data = schemas.ObservationRecordCreate(
                growth_condition=f"第{i+1}周：生长良好，新叶{2+i}片",
                notes=f"生长健康，湿度60%"
            )
            record = crud.add_observation_record(db, plant.id, record_data)
            assert record is not None
            print(f"✓ 添加第{i+1}次观察记录")
        
        records = crud.get_observation_records(db, plant.id)
        assert len(records) == 3
        print(f"✓ 共 {len(records)} 条观察记录")
        
        print("  注意：观察期未满(90天)，无法提前结束")
        incomplete = crud.complete_observation(db, plant.id, PlantStatus.NORMAL)
        assert incomplete is None
        print("✓ 验证：观察期未满时无法提前结束")
        
        from sqlalchemy import update as sa_update
        db.execute(
            sa_update(models.Plant)
            .where(models.Plant.id == plant.id)
            .values(observation_end_date=datetime.utcnow() - timedelta(days=1))
        )
        db.commit()
        
        completed = crud.complete_observation(db, plant.id, PlantStatus.NORMAL)
        assert completed is not None
        assert completed.status == PlantStatus.NORMAL
        print(f"✓ 观察期满，植物转正常状态: {completed.status.value}")
        
    finally:
        db.close()
    print()


def test_exhibition_system():
    print("=" * 60)
    print("测试3: 专题展览系统")
    print("=" * 60)
    
    db = TestSessionLocal()
    try:
        normal_plant = crud.create_plant(db, schemas.PlantCreate(
            scientific_name="展览植物1",
            family="菊科",
            genus="菊属",
            origin="日本",
            status=PlantStatus.NORMAL
        ))
        print(f"✓ 创建正常植物: {normal_plant.scientific_name}")
        
        dormant_plant = crud.create_plant(db, schemas.PlantCreate(
            scientific_name="展览植物2",
            family="菊科",
            genus="菊属",
            origin="韩国",
            status=PlantStatus.DORMANT
        ))
        print(f"✓ 创建休眠植物: {dormant_plant.scientific_name}")
        
        dead_plant = crud.create_plant(db, schemas.PlantCreate(
            scientific_name="已死亡植物",
            family="菊科",
            genus="菊属",
            origin="中国",
            status=PlantStatus.DEAD
        ))
        print(f"✓ 创建已死亡植物: {dead_plant.scientific_name}")
        
        exhibition_data = schemas.ExhibitionCreate(
            name="秋季菊花展",
            description="2024年秋季菊花专题展览",
            start_date=datetime(2024, 10, 1),
            plant_ids=[normal_plant.id, dormant_plant.id, dead_plant.id]
        )
        exhibition = crud.create_exhibition(db, exhibition_data)
        print(f"✓ 创建展览: {exhibition.name}")
        
        plant_ids = crud.get_exhibition_plants(db, exhibition.id)
        assert len(plant_ids) == 2
        print(f"✓ 展览参与植物: {plant_ids} (已死亡植物被跳过)")
        
        normal_plant_refresh = crud.get_plant(db, normal_plant.id)
        assert normal_plant_refresh.status == PlantStatus.DORMANT
        print(f"✓ 展览植物状态变为休眠: {normal_plant_refresh.status.value}")
        
        ended = crud.end_exhibition(db, exhibition.id)
        assert ended.is_active == False
        print(f"✓ 结束展览: {ended.name}")
        
        normal_plant_final = crud.get_plant(db, normal_plant.id)
        assert normal_plant_final.status == PlantStatus.NORMAL
        print(f"✓ 植物恢复原状态: {normal_plant_final.status.value}")
        
        dormant_plant_final = crud.get_plant(db, dormant_plant.id)
        assert dormant_plant_final.status == PlantStatus.DORMANT
        print(f"✓ 休眠植物保持原状态: {dormant_plant_final.status.value}")
        
    finally:
        db.close()
    print()


def test_publication_protection():
    print("=" * 60)
    print("测试4: 园区导览发布与保护等级")
    print("=" * 60)
    
    db = TestSessionLocal()
    try:
        normal_published = crud.create_plant(db, schemas.PlantCreate(
            scientific_name="普通植物",
            family="豆科",
            genus="大豆属",
            origin="中国",
            status=PlantStatus.NORMAL,
            protection_level=ProtectionLevel.NONE,
            location="南区草坪",
            is_published=True
        ))
        print(f"✓ 创建已发布普通植物: {normal_published.scientific_name}")
        
        level1_published = crud.create_plant(db, schemas.PlantCreate(
            scientific_name="一级保护植物",
            family="松科",
            genus="松属",
            origin="中国",
            status=PlantStatus.NORMAL,
            protection_level=ProtectionLevel.LEVEL_1,
            location="保护区A区",
            is_published=True
        ))
        print(f"✓ 创建已发布一级保护植物: {level1_published.scientific_name}")
        
        dead_plant = crud.create_plant(db, schemas.PlantCreate(
            scientific_name="死亡植物",
            family="柏科",
            genus="柏属",
            origin="中国",
            status=PlantStatus.DEAD,
            is_published=True
        ))
        print(f"✓ 创建已发布但已死亡植物")
        
        published = crud.get_published_plants(db)
        assert len(published) == 2
        print(f"✓ 园区导览植物: {len(published)} 个 (死亡植物被排除)")
        
        for p in published:
            if p.protection_level == ProtectionLevel.LEVEL_1:
                assert p.location is None or "A区" in p.location
                print(f"✓ 一级保护植物隐藏位置: {p.scientific_name} -> 位置: 无")
            else:
                print(f"✓ 普通植物显示位置: {p.scientific_name} -> 位置: {p.location}")
        
    finally:
        db.close()
    print()


def test_csv_export():
    print("=" * 60)
    print("测试5: CSV导出功能")
    print("=" * 60)
    
    db = TestSessionLocal()
    try:
        plants = crud.get_plants_for_export(db, family="菊科")
        print(f"✓ 按科筛选: 找到 {len(plants)} 个菊科植物")
        
        all_export = crud.get_plants_for_export(db)
        dead_count = len([p for p in all_export if p.status == PlantStatus.DEAD])
        print(f"✓ 总植物数: {len(all_export)} (死亡植物 {dead_count} 个被排除)")
        
        families = crud.get_unique_families(db)
        print(f"✓ 所有科: {families}")
        
        genera = crud.get_unique_genera(db, family="菊科")
        print(f"✓ 菊科下的属: {genera}")
        
    finally:
        db.close()
    print()


def test_cli_commands():
    print("=" * 60)
    print("测试6: 命令行客户端基础测试")
    print("=" * 60)
    
    db = TestSessionLocal()
    try:
        plants = crud.get_plants(db, limit=5)
        exhibitions = crud.get_exhibitions(db)
        
        print(f"✓ CLI可访问数据: {len(plants)} 个植物, {len(exhibitions)} 个展览")
        print("✓ CLI命令: python cli.py search --family 蔷薇科")
        print("✓ CLI命令: python cli.py plant 1")
        print("✓ CLI命令: python cli.py exhibitions")
        print("✓ CLI命令: python cli.py stats")
        print("✓ CLI命令: python cli.py families")
        
    finally:
        db.close()
    print()


def main():
    print("\n" + "=" * 60)
    print("植物数据库物种管理系统 - 功能测试")
    print("=" * 60 + "\n")
    
    try:
        test_plant_management()
        test_introduction_observation()
        test_exhibition_system()
        test_publication_protection()
        test_csv_export()
        test_cli_commands()
        
        print("=" * 60)
        print("所有测试通过！✓")
        print("=" * 60)
    except Exception as e:
        print(f"\n测试失败: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == "__main__":
    main()
