import sys
import os
from sqlalchemy.orm import Session
from typing import Optional
import argparse

from database import SessionLocal, engine
import models
import crud
from models import PlantStatus, ProtectionLevel

models.Base.metadata.create_all(bind=engine)


def format_plant(plant):
    return f"""
ID: {plant.id}
学名: {plant.scientific_name}
科属: {plant.family} {plant.genus}
原产地: {plant.origin}
状态: {plant.status.value}
保护等级: {plant.protection_level.value}
位置: {plant.location or 'N/A'}
已发布: {'是' if plant.is_published else '否'}
引种日期: {plant.introduction_date.strftime('%Y-%m-%d') if plant.introduction_date else 'N/A'}
观察结束日期: {plant.observation_end_date.strftime('%Y-%m-%d') if plant.observation_end_date else 'N/A'}
"""


def format_exhibition(exhibition):
    return f"""
ID: {exhibition.id}
名称: {exhibition.name}
描述: {exhibition.description or 'N/A'}
开始日期: {exhibition.start_date.strftime('%Y-%m-%d')}
结束日期: {exhibition.end_date.strftime('%Y-%m-%d') if exhibition.end_date else '进行中'}
状态: {'进行中' if exhibition.is_active else '已结束'}
"""


def cmd_search_plants(args):
    db = SessionLocal()
    try:
        plants = crud.get_plants(
            db=db,
            family=args.family,
            genus=args.genus,
            status=args.status,
            scientific_name=args.name,
            skip=0,
            limit=100,
        )
        if not plants:
            print("未找到匹配的植物")
            return
        print(f"找到 {len(plants)} 个植物：\n")
        for plant in plants:
            print("-" * 50)
            print(format_plant(plant))
    finally:
        db.close()


def cmd_get_plant(args):
    db = SessionLocal()
    try:
        plant = crud.get_plant(db=db, plant_id=args.id)
        if not plant:
            print(f"未找到ID为 {args.id} 的植物")
            return
        print(format_plant(plant))
        records = crud.get_observation_records(db=db, plant_id=args.id)
        if records:
            print(f"\n观察记录（共 {len(records)} 条）：")
            for record in records:
                print(f"  [{record.record_date.strftime('%Y-%m-%d')}] {record.growth_condition}")
                if record.notes:
                    print(f"    备注: {record.notes}")
    finally:
        db.close()


def cmd_list_exhibitions(args):
    db = SessionLocal()
    try:
        is_active = None
        if args.active:
            is_active = True
        elif args.inactive:
            is_active = False
        exhibitions = crud.get_exhibitions(db=db, skip=0, limit=100, is_active=is_active)
        if not exhibitions:
            print("未找到展览")
            return
        print(f"找到 {len(exhibitions)} 个展览：\n")
        for exhibition in exhibitions:
            print("-" * 50)
            print(format_exhibition(exhibition))
            plant_ids = crud.get_exhibition_plants(db=db, exhibition_id=exhibition.id)
            print(f"参与植物ID: {', '.join(map(str, plant_ids)) if plant_ids else '无'}")
    finally:
        db.close()


def cmd_get_exhibition(args):
    db = SessionLocal()
    try:
        exhibition = crud.get_exhibition(db=db, exhibition_id=args.id)
        if not exhibition:
            print(f"未找到ID为 {args.id} 的展览")
            return
        print(format_exhibition(exhibition))
        plant_ids = crud.get_exhibition_plants(db=db, exhibition_id=args.id)
        if plant_ids:
            print(f"\n参与植物（共 {len(plant_ids)} 个）：")
            for plant_id in plant_ids:
                plant = crud.get_plant(db=db, plant_id=plant_id)
                if plant:
                    print(f"  ID: {plant.id}, 学名: {plant.scientific_name}, 原状态: {plant.status.value}")
    finally:
        db.close()


def cmd_list_families(args):
    db = SessionLocal()
    try:
        families = crud.get_unique_families(db=db)
        if not families:
            print("数据库中没有植物数据")
            return
        print("所有科：\n")
        for family in families:
            print(f"  - {family}")
        print(f"\n共 {len(families)} 个科")
    finally:
        db.close()


def cmd_list_genera(args):
    db = SessionLocal()
    try:
        genera = crud.get_unique_genera(db=db, family=args.family)
        if not genera:
            print("数据库中没有匹配的属")
            return
        if args.family:
            print(f"科 {args.family} 下的所有属：\n")
        else:
            print("所有属：\n")
        for genus in genera:
            print(f"  - {genus}")
        print(f"\n共 {len(genera)} 个属")
    finally:
        db.close()


def cmd_stats(args):
    db = SessionLocal()
    try:
        total_plants = db.query(models.Plant).count()
        normal_plants = db.query(models.Plant).filter(models.Plant.status == PlantStatus.NORMAL).count()
        dormant_plants = db.query(models.Plant).filter(models.Plant.status == PlantStatus.DORMANT).count()
        observing_plants = db.query(models.Plant).filter(models.Plant.status == PlantStatus.INTRODUCTION_OBSERVATION).count()
        dead_plants = db.query(models.Plant).filter(models.Plant.status == PlantStatus.DEAD).count()
        published_plants = db.query(models.Plant).filter(models.Plant.is_published == True).count()
        active_exhibitions = db.query(models.Exhibition).filter(models.Exhibition.is_active == True).count()
        
        print("=" * 50)
        print("植物数据库统计")
        print("=" * 50)
        print(f"总植物数: {total_plants}")
        print(f"正常状态: {normal_plants}")
        print(f"休眠状态: {dormant_plants}")
        print(f"引种观察: {observing_plants}")
        print(f"已死亡: {dead_plants}")
        print(f"已发布到导览: {published_plants}")
        print(f"进行中的展览: {active_exhibitions}")
        print("=" * 50)
    finally:
        db.close()


def main():
    parser = argparse.ArgumentParser(description="植物数据库物种管理系统 - 命令行客户端")
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    search_parser = subparsers.add_parser("search", help="搜索植物")
    search_parser.add_argument("--family", type=str, help="按科搜索")
    search_parser.add_argument("--genus", type=str, help="按属搜索")
    search_parser.add_argument("--status", type=str, choices=["正常", "休眠", "引种观察", "已死亡"], help="按状态搜索")
    search_parser.add_argument("--name", type=str, help="按学名关键字搜索")
    search_parser.set_defaults(func=cmd_search_plants)
    
    plant_parser = subparsers.add_parser("plant", help="查看植物详情")
    plant_parser.add_argument("id", type=int, help="植物ID")
    plant_parser.set_defaults(func=cmd_get_plant)
    
    list_exh_parser = subparsers.add_parser("exhibitions", help="列出展览")
    list_exh_parser.add_argument("--active", action="store_true", help="仅显示进行中的展览")
    list_exh_parser.add_argument("--inactive", action="store_true", help="仅显示已结束的展览")
    list_exh_parser.set_defaults(func=cmd_list_exhibitions)
    
    exh_parser = subparsers.add_parser("exhibition", help="查看展览详情")
    exh_parser.add_argument("id", type=int, help="展览ID")
    exh_parser.set_defaults(func=cmd_get_exhibition)
    
    subparsers.add_parser("families", help="列出所有科").set_defaults(func=cmd_list_families)
    
    genera_parser = subparsers.add_parser("genera", help="列出所有属")
    genera_parser.add_argument("--family", type=str, help="指定科")
    genera_parser.set_defaults(func=cmd_list_genera)
    
    subparsers.add_parser("stats", help="查看统计信息").set_defaults(func=cmd_stats)
    
    args = parser.parse_args()
    
    if args.command is None:
        parser.print_help()
        return
    
    if hasattr(args, 'status') and args.status:
        status_map = {
            "正常": PlantStatus.NORMAL,
            "休眠": PlantStatus.DORMANT,
            "引种观察": PlantStatus.INTRODUCTION_OBSERVATION,
            "已死亡": PlantStatus.DEAD,
        }
        args.status = status_map[args.status]
    
    args.func(args)


if __name__ == "__main__":
    main()
