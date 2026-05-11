#!/usr/bin/env python3
import sys
import os
from datetime import date, timedelta

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from src.core.database import SessionLocal, init_db
from src.core.models import (
    Certificate,
    CertificateType,
    CertificateStatus,
    AnnualInspection,
    ComplianceCheck,
    Todo,
    TodoType,
    TodoStatus,
)
from src.core.services import (
    CertificateService,
    InspectionService,
    ComplianceService,
    TodoService,
    ValidationError,
)


def test_certificate_validation():
    print("=== 测试证照有效性验证 ===")
    db = SessionLocal()
    
    try:
        cert = CertificateService.create_certificate(
            db,
            name="测试采矿许可证",
            cert_type=CertificateType.MINING_LICENSE,
            number="TEST-2026-001",
            issuing_date=date(2026, 1, 1),
            expiry_date=date(2025, 12, 31),
        )
        print("✗ 应该抛出异常：有效期早于发证日期")
    except ValidationError as e:
        print(f"✓ 正确验证：{e}")
    finally:
        db.close()


def test_certificate_status():
    print("\n=== 测试证照状态 ===")
    db = SessionLocal()
    
    today = date.today()
    
    cert1 = CertificateService.create_certificate(
        db,
        name="正常有效期证照",
        cert_type=CertificateType.MINING_LICENSE,
        number="TEST-STATUS-001",
        issuing_date=today,
        expiry_date=today + timedelta(days=180),
    )
    print(f"✓ 正常证照状态: {cert1.status} (期望: active)")
    assert cert1.status == CertificateStatus.ACTIVE
    
    cert2 = CertificateService.create_certificate(
        db,
        name="即将到期证照",
        cert_type=CertificateType.SAFETY_LICENSE,
        number="TEST-STATUS-002",
        issuing_date=today,
        expiry_date=today + timedelta(days=60),
    )
    print(f"✓ 即将到期证照状态: {cert2.status} (期望: about_to_expire)")
    assert cert2.status == CertificateStatus.ABOUT_TO_EXPIRE
    
    cert3 = CertificateService.create_certificate(
        db,
        name="已过期证照",
        cert_type=CertificateType.ENV_APPROVAL,
        number="TEST-STATUS-003",
        issuing_date=today - timedelta(days=365),
        expiry_date=today - timedelta(days=30),
    )
    print(f"✓ 已过期证照状态: {cert3.status} (期望: expired)")
    assert cert3.status == CertificateStatus.EXPIRED
    
    db.close()


def test_annual_inspection_unique():
    print("\n=== 测试年检唯一性 ===")
    db = SessionLocal()
    
    cert = CertificateService.create_certificate(
        db,
        name="年检测试证照",
        cert_type=CertificateType.MINING_LICENSE,
        number="TEST-INSPECTION-001",
        issuing_date=date.today(),
        expiry_date=date.today() + timedelta(days=365),
    )
    
    InspectionService.create_inspection(
        db,
        certificate_id=cert.id,
        year=2026,
        inspection_date=date.today(),
        result=True,
    )
    print("✓ 第一条年检记录创建成功")
    
    try:
        InspectionService.create_inspection(
            db,
            certificate_id=cert.id,
            year=2026,
            inspection_date=date.today(),
            result=False,
        )
        print("✗ 应该抛出异常：同一年度多条年检记录")
    except ValidationError as e:
        print(f"✓ 正确验证：{e}")
    
    db.close()


def test_safety_compliance():
    print("\n=== 测试安全合规性 ===")
    db = SessionLocal()
    
    cert = CertificateService.create_certificate(
        db,
        name="安全合规测试证照",
        cert_type=CertificateType.SAFETY_LICENSE,
        number="TEST-COMPLIANCE-001",
        issuing_date=date.today(),
        expiry_date=date.today() + timedelta(days=365),
    )
    
    check = ComplianceService.create_check(
        db,
        certificate_id=cert.id,
        check_date=date.today(),
        check_items="安全设备检查",
        is_compliant=True,
        has_safety_issues=True,
    )
    print(f"✓ 有安全隐患但标记合规的检查结果: {check.is_compliant} (期望: False)")
    assert check.is_compliant == False
    
    db.close()


def test_cancelled_certificate():
    print("\n=== 测试注销证照 ===")
    db = SessionLocal()
    
    cert = CertificateService.create_certificate(
        db,
        name="注销测试证照",
        cert_type=CertificateType.MINING_LICENSE,
        number="TEST-CANCEL-001",
        issuing_date=date.today(),
        expiry_date=date.today() + timedelta(days=365),
    )
    
    cancelled_cert = CertificateService.cancel_certificate(db, cert.id)
    print(f"✓ 注销后证照状态: {cancelled_cert.status} (期望: cancelled)")
    assert cancelled_cert.status == CertificateStatus.CANCELLED
    
    db.close()


def test_todo_generation():
    print("\n=== 测试待办事项生成 ===")
    db = SessionLocal()
    
    today = date.today()
    
    cert = CertificateService.create_certificate(
        db,
        name="待办测试证照",
        cert_type=CertificateType.MINING_LICENSE,
        number="TEST-TODO-001",
        issuing_date=today,
        expiry_date=today + timedelta(days=60),
    )
    
    todos = TodoService.list_todos(db, todo_type=TodoType.CERTIFICATE_EXPIRY)
    cert_todos = [t for t in todos if t.certificate_id == cert.id]
    print(f"✓ 即将到期证照自动生成待办数: {len(cert_todos)} (期望: 1)")
    assert len(cert_todos) >= 1
    
    check = ComplianceService.create_check(
        db,
        certificate_id=cert.id,
        check_date=today,
        check_items="设备检查",
        is_compliant=False,
        has_safety_issues=False,
    )
    
    todos = TodoService.list_todos(db, todo_type=TodoType.COMPLIANCE_RECTIFICATION)
    rectification_todos = [t for t in todos if t.compliance_check_id == check.id]
    print(f"✓ 不合规检查自动生成整改待办数: {len(rectification_todos)} (期望: 1)")
    assert len(rectification_todos) == 1
    
    db.close()


def main():
    print("开始测试矿业行政审批管理系统核心功能...\n")
    
    test_certificate_validation()
    test_certificate_status()
    test_annual_inspection_unique()
    test_safety_compliance()
    test_cancelled_certificate()
    test_todo_generation()
    
    print("\n=== 所有核心测试通过！ ===")
    
    if os.path.exists("./mining_admin.db"):
        os.remove("./mining_admin.db")
        print("\n清理测试数据库完成")


if __name__ == "__main__":
    init_db()
    main()
