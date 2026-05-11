import sys
sys.path.insert(0, 'src')

from datetime import date, datetime, timedelta
from core import (
    Mine,
    MiningOperation,
    Transport,
    SafetyCheck,
    SafetyCheckItem,
    Todo,
    Repository,
    ValidationError,
    get_monthly_statistics,
)


def test_mine_management():
    print("测试矿区管理...")
    repo = Repository()
    
    mine = Mine(name="一号矿区", mineral_type="金矿", annual_capacity=100000)
    created = repo.create_mine(mine)
    assert created.id == 1
    assert created.name == "一号矿区"
    
    mines = repo.list_mines()
    assert len(mines) == 1
    
    print("✓ 矿区管理测试通过")


def test_mining_operation_validation():
    print("测试采矿作业验证...")
    repo = Repository()
    
    mine = repo.create_mine(Mine(name="测试矿区", mineral_type="铁矿", annual_capacity=50000))
    
    op1 = MiningOperation(
        mine_id=mine.id,
        operation_date=date.today(),
        planned_output=100,
        actual_output=115
    )
    repo.create_mining_operation(op1)
    
    op2 = MiningOperation(
        mine_id=mine.id,
        operation_date=date.today(),
        planned_output=100,
        actual_output=130
    )
    try:
        repo.create_mining_operation(op2)
        assert False, "应该抛出异常"
    except ValidationError as e:
        assert "超过计划产量的120%" in str(e)
    
    op3 = MiningOperation(
        mine_id=mine.id,
        operation_date=date.today(),
        planned_output=100,
        actual_output=50
    )
    try:
        repo.create_mining_operation(op3)
        assert False, "应该抛出异常"
    except ValidationError as e:
        assert "已存在采矿作业记录" in str(e)
    
    print("✓ 采矿作业验证测试通过")


def test_transport_validation():
    print("测试运输验证...")
    repo = Repository()
    mine = repo.create_mine(Mine(name="测试矿区", mineral_type="铜矿", annual_capacity=30000))
    
    now = datetime.now()
    t1 = Transport(
        mine_id=mine.id,
        vehicle_number="A-123",
        transport_date=date.today(),
        departure_time=now,
        arrival_time=now - timedelta(hours=1),
        transport_volume=50
    )
    try:
        repo.create_transport(t1)
        assert False, "应该抛出异常"
    except ValidationError as e:
        assert "到达时间不能早于出发时间" in str(e)
    
    t2 = Transport(
        mine_id=mine.id,
        vehicle_number="A-123",
        transport_date=date.today(),
        departure_time=now - timedelta(hours=1),
        arrival_time=now,
        transport_volume=50
    )
    created = repo.create_transport(t2)
    assert created.id is not None
    
    print("✓ 运输验证测试通过")


def test_safety_check():
    print("测试安全检查...")
    repo = Repository()
    mine = repo.create_mine(Mine(name="测试矿区", mineral_type="铝矿", annual_capacity=40000))
    
    items1 = [
        SafetyCheckItem(name="通风系统", is_abnormal=False),
        SafetyCheckItem(name="排水系统", is_abnormal=False),
    ]
    check1 = SafetyCheck(mine_id=mine.id, check_date=date.today(), items=items1)
    created1 = repo.create_safety_check(check1)
    assert created1.status.value == "qualified"
    
    items2 = [
        SafetyCheckItem(name="通风系统", is_abnormal=False),
        SafetyCheckItem(name="排水系统", is_abnormal=True),
    ]
    check2 = SafetyCheck(mine_id=mine.id, check_date=date.today(), items=items2)
    created2 = repo.create_safety_check(check2)
    assert created2.status.value == "unqualified"
    
    print("✓ 安全检查测试通过")


def test_todo_validation():
    print("测试待办整改验证...")
    repo = Repository()
    mine = repo.create_mine(Mine(name="测试矿区", mineral_type="锌矿", annual_capacity=20000))
    
    valid_deadline = date.today() + timedelta(days=15)
    invalid_deadline = date.today() + timedelta(days=45)
    
    todo1 = Todo(
        mine_id=mine.id,
        description="测试整改项",
        responsible_person="张三",
        deadline=valid_deadline
    )
    created = repo.create_todo(todo1)
    assert created.id is not None
    
    todo2 = Todo(
        mine_id=mine.id,
        description="测试整改项2",
        responsible_person="李四",
        deadline=invalid_deadline
    )
    try:
        repo.create_todo(todo2)
        assert False, "应该抛出异常"
    except ValidationError as e:
        assert "不能超过30天" in str(e)
    
    print("✓ 待办整改验证测试通过")


def test_overdue_todo_upgrade():
    print("测试超期待办升级...")
    repo = Repository()
    mine = repo.create_mine(Mine(name="测试矿区", mineral_type="铅矿", annual_capacity=25000))
    
    overdue_deadline = date.today() - timedelta(days=5)
    todo = Todo(
        mine_id=mine.id,
        description="超期待办",
        responsible_person="王五",
        deadline=overdue_deadline
    )
    created = repo.create_todo(todo)
    
    upgraded = repo.upgrade_overdue_todos()
    assert upgraded >= 1
    
    updated = repo.get_todo(created.id)
    assert updated.priority.value != "medium"
    
    print("✓ 超期待办升级测试通过")


def test_statistics():
    print("测试统计功能...")
    repo = Repository()
    mine = repo.create_mine(Mine(name="测试矿区", mineral_type="镍矿", annual_capacity=35000))
    
    today = date.today()
    
    repo.create_mining_operation(MiningOperation(
        mine_id=mine.id,
        operation_date=today,
        planned_output=100,
        actual_output=90
    ))
    
    repo.create_mining_operation(MiningOperation(
        mine_id=mine.id,
        operation_date=today - timedelta(days=1),
        planned_output=100,
        actual_output=100
    ))
    
    now = datetime.now()
    repo.create_transport(Transport(
        mine_id=mine.id,
        vehicle_number="B-456",
        transport_date=today,
        departure_time=now - timedelta(hours=2),
        arrival_time=now,
        transport_volume=30
    ))
    
    repo.create_safety_check(SafetyCheck(
        mine_id=mine.id,
        check_date=today,
        items=[SafetyCheckItem(name="检查项1", is_abnormal=False)]
    ))
    
    stats = get_monthly_statistics(repo)
    
    assert stats.total_output == 190
    assert stats.completion_rate == 95
    assert stats.transport_count == 1
    assert stats.safety_check_count == 1
    
    print("✓ 统计功能测试通过")


if __name__ == "__main__":
    test_mine_management()
    test_mining_operation_validation()
    test_transport_validation()
    test_safety_check()
    test_todo_validation()
    test_overdue_todo_upgrade()
    test_statistics()
    print("\n🎉 所有核心测试通过！")
