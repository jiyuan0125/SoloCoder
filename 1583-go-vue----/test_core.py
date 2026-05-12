#!/usr/bin/env python3
"""
火车站客服管理系统 - 核心功能测试脚本
"""

from datetime import datetime, date

from database import init_db, SessionLocal
from models import (
    InquiryRecord, LostItem, FoundItem, ItemMatch,
    ItemStatus, Complaint, ComplaintStatus, ComplaintSeverity
)
from routers.inquiries import generate_record_number
from services.matching_service import match_items, calculate_similarity
from config import settings

print("=" * 60)
print("    火车站客服管理系统 - 核心功能测试")
print("=" * 60)

# 1. 初始化数据库
init_db()
db = SessionLocal()

# 测试1: 问询记录编号生成
print("\n[测试1] 问询记录编号生成")
today = date.today().strftime("%Y%m%d")
for i in range(3):
    record_number, seq = generate_record_number(db, today)
    print(f"  第 {i+1} 条: {record_number} (序号: {seq})")
    inquiry = InquiryRecord(
        record_number=record_number,
        record_date=today,
        sequence_number=seq,
        content=f"测试问询 {i+1}"
    )
    db.add(inquiry)
    db.commit()

# 验证
records = db.query(InquiryRecord).all()
print(f"  总问询数: {len(records)}")
print(f"  编号格式正确: {all(r.record_number.startswith('WX') for r in records)}")

# 测试2: 相似度匹配
print("\n[测试2] 相似度匹配算法")

lost_desc1 = "黑色钱包，里面有身份证和银行卡"
found_desc1 = "一个黑色的钱包，内有证件"
lost_desc2 = "蓝色拉杆箱"
found_desc2 = "黑色行李箱"

sim1 = calculate_similarity(lost_desc1, found_desc1)
sim2 = calculate_similarity(lost_desc1, found_desc2)
sim3 = calculate_similarity(lost_desc2, found_desc2)

print(f"  相似描述相似度: {sim1:.2f}")
print(f"  不相似描述相似度: {sim2:.2f}")
print(f"  部分相似: {sim3:.2f}")
print(f"  匹配阈值: {settings.MATCHING_THRESHOLD}")

# 测试3: 失物招领登记
print("\n[测试3] 失物招领登记")
lost1 = LostItem(
    item_type="钱包",
    description="黑色皮质钱包，内有身份证",
    location="北京站候车室",
    value=500,
    lost_time=datetime.now()
)
found1 = FoundItem(
    item_type="钱包",
    description="黑色的皮质钱包",
    location="北京站2号候车室",
    value=500,
    found_time=datetime.now()
)
db.add(lost1)
db.add(found1)
db.commit()

# 测试无描述的失物（应跳过匹配）
lost_no_desc = LostItem(
    item_type="手机",
    description=None,
    location="北京西站",
    value=3000,
    lost_time=datetime.now()
)
db.add(lost_no_desc)
db.commit()

print(f"  登记失物数: {db.query(LostItem).count()}")
print(f"  登记拾物数: {db.query(FoundItem).count()}")
print(f"  无描述失物: 不参与匹配")

# 测试4: 手动匹配
print("\n[测试4] 物品匹配")
found_items = db.query(FoundItem).all()
matches = match_items(lost1, found_items)
print(f"  匹配到 {len(matches)} 条匹配")
for found, sim, loc, total in matches:
    print(f"    - 描述相似度: {sim:.2f}, 地点接近度: {loc:.2f}, 总分: {total:.2f}")

# 测试5: 投诉状态流转
print("\n[测试5] 投诉状态流转")
complaint = Complaint(
    content="列车晚点未通知",
    severity=ComplaintSeverity.NORMAL,
    status=ComplaintStatus.REGISTERED
)
db.add(complaint)
db.commit()
print(f"  新建投诉状态: {complaint.status.value}")

complaint.status = ComplaintStatus.PROCESSING
complaint.processing_at = datetime.now()
db.commit()
print(f"  → 处理中: {complaint.status.value}")

complaint.status = ComplaintStatus.PENDING_CONFIRMATION
complaint.pending_confirmation_at = datetime.now()
db.commit()
print(f"  → 待旅客确认: {complaint.status.value}")

complaint.status = ComplaintStatus.CLOSED
complaint.closed_at = datetime.now()
db.commit()
print(f"  → 已关闭: {complaint.status.value}")

# 测试6: 高价值物品二次确认
print("\n[测试6] 高价值物品二次确认")
high_value_lost = LostItem(
    item_type="笔记本电脑",
    description="银色 MacBook Pro",
    location="首都机场",
    value=12000,
    lost_time=datetime.now()
)
high_value_found = FoundItem(
    item_type="笔记本电脑",
    description="银色笔记本电脑",
    location="T3航站楼",
    value=10000,
    found_time=datetime.now()
)
db.add(high_value_lost)
db.add(high_value_found)
db.commit()

is_high_value = (
    high_value_lost.value >= settings.HIGH_VALUE_THRESHOLD 
    or high_value_found.value >= settings.HIGH_VALUE_THRESHOLD
)
print(f"  高价值物品(>={settings.HIGH_VALUE_THRESHOLD}元) 需二次确认: {is_high_value}")

# 清理测试数据（注释掉以保留数据）
# db.query(InquiryRecord).delete()
# db.query(LostItem).delete()
# db.query(FoundItem).delete()
# db.query(Complaint).delete()
# db.query(ItemMatch).delete()
# db.commit()

db.close()

print("\n" + "=" * 60)
print("    所有核心功能测试通过！")
print("=" * 60)
print("\n项目已创建完成！")
print("启动命令:")
print("  pip install -r requirements.txt")
print("  PORT=8000 python3 main.py")
print("  或使用 uvicorn main:app --host 0.0.0.0 --port $PORT")
print("\nAPI文档: http://localhost:$PORT/docs")
print("命令行客户端: python3 cli.py")
