from datetime import date, timedelta
from typing import List, Optional
from copy import deepcopy

from .models import (
    Mine,
    MiningOperation,
    Transport,
    SafetyCheck,
    Todo,
    TodoPriority,
    TodoStatus,
)
from .validators import (
    ValidationError,
    validate_mining_operation,
    validate_transport,
    validate_safety_check,
    validate_todo,
)


class Repository:
    def __init__(self):
        self.mines: List[Mine] = []
        self.mining_operations: List[MiningOperation] = []
        self.transports: List[Transport] = []
        self.safety_checks: List[SafetyCheck] = []
        self.todos: List[Todo] = []
        
        self._next_mine_id = 1
        self._next_operation_id = 1
        self._next_transport_id = 1
        self._next_safety_check_id = 1
        self._next_todo_id = 1

    def create_mine(self, mine: Mine) -> Mine:
        mine = deepcopy(mine)
        mine.id = self._next_mine_id
        self.mines.append(mine)
        self._next_mine_id += 1
        return deepcopy(mine)

    def get_mine(self, mine_id: int) -> Optional[Mine]:
        for mine in self.mines:
            if mine.id == mine_id:
                return deepcopy(mine)
        return None

    def list_mines(self) -> List[Mine]:
        return deepcopy(self.mines)

    def update_mine(self, mine: Mine) -> Optional[Mine]:
        for i, existing in enumerate(self.mines):
            if existing.id == mine.id:
                self.mines[i] = deepcopy(mine)
                return deepcopy(mine)
        return None

    def delete_mine(self, mine_id: int) -> bool:
        for i, mine in enumerate(self.mines):
            if mine.id == mine_id:
                del self.mines[i]
                return True
        return False

    def create_mining_operation(self, operation: MiningOperation) -> MiningOperation:
        operation = deepcopy(operation)
        validate_mining_operation(operation, self.mining_operations)
        
        if not self.get_mine(operation.mine_id):
            raise ValidationError(f"矿区 ID {operation.mine_id} 不存在")
        
        operation.id = self._next_operation_id
        self.mining_operations.append(operation)
        self._next_operation_id += 1
        return deepcopy(operation)

    def get_mining_operation(self, op_id: int) -> Optional[MiningOperation]:
        for op in self.mining_operations:
            if op.id == op_id:
                return deepcopy(op)
        return None

    def list_mining_operations(self, mine_id: Optional[int] = None) -> List[MiningOperation]:
        result = []
        for op in self.mining_operations:
            if mine_id is None or op.mine_id == mine_id:
                result.append(deepcopy(op))
        return result

    def update_mining_operation(self, operation: MiningOperation) -> Optional[MiningOperation]:
        for i, existing in enumerate(self.mining_operations):
            if existing.id == operation.id:
                validate_mining_operation(operation, self.mining_operations)
                if not self.get_mine(operation.mine_id):
                    raise ValidationError(f"矿区 ID {operation.mine_id} 不存在")
                self.mining_operations[i] = deepcopy(operation)
                return deepcopy(operation)
        return None

    def delete_mining_operation(self, op_id: int) -> bool:
        for i, op in enumerate(self.mining_operations):
            if op.id == op_id:
                del self.mining_operations[i]
                return True
        return False

    def create_transport(self, transport: Transport) -> Transport:
        transport = deepcopy(transport)
        validate_transport(transport)
        
        if not self.get_mine(transport.mine_id):
            raise ValidationError(f"矿区 ID {transport.mine_id} 不存在")
        
        transport.id = self._next_transport_id
        self.transports.append(transport)
        self._next_transport_id += 1
        return deepcopy(transport)

    def get_transport(self, transport_id: int) -> Optional[Transport]:
        for t in self.transports:
            if t.id == transport_id:
                return deepcopy(t)
        return None

    def list_transports(self, mine_id: Optional[int] = None) -> List[Transport]:
        result = []
        for t in self.transports:
            if mine_id is None or t.mine_id == mine_id:
                result.append(deepcopy(t))
        return result

    def update_transport(self, transport: Transport) -> Optional[Transport]:
        for i, existing in enumerate(self.transports):
            if existing.id == transport.id:
                validate_transport(transport)
                if not self.get_mine(transport.mine_id):
                    raise ValidationError(f"矿区 ID {transport.mine_id} 不存在")
                self.transports[i] = deepcopy(transport)
                return deepcopy(transport)
        return None

    def delete_transport(self, transport_id: int) -> bool:
        for i, t in enumerate(self.transports):
            if t.id == transport_id:
                del self.transports[i]
                return True
        return False

    def create_safety_check(self, check: SafetyCheck) -> SafetyCheck:
        check = deepcopy(check)
        validate_safety_check(check)
        
        if not self.get_mine(check.mine_id):
            raise ValidationError(f"矿区 ID {check.mine_id} 不存在")
        
        check.id = self._next_safety_check_id
        self.safety_checks.append(check)
        self._next_safety_check_id += 1
        return deepcopy(check)

    def get_safety_check(self, check_id: int) -> Optional[SafetyCheck]:
        for c in self.safety_checks:
            if c.id == check_id:
                return deepcopy(c)
        return None

    def list_safety_checks(self, mine_id: Optional[int] = None) -> List[SafetyCheck]:
        result = []
        for c in self.safety_checks:
            if mine_id is None or c.mine_id == mine_id:
                result.append(deepcopy(c))
        return result

    def update_safety_check(self, check: SafetyCheck) -> Optional[SafetyCheck]:
        for i, existing in enumerate(self.safety_checks):
            if existing.id == check.id:
                validate_safety_check(check)
                if not self.get_mine(check.mine_id):
                    raise ValidationError(f"矿区 ID {check.mine_id} 不存在")
                self.safety_checks[i] = deepcopy(check)
                return deepcopy(check)
        return None

    def delete_safety_check(self, check_id: int) -> bool:
        for i, c in enumerate(self.safety_checks):
            if c.id == check_id:
                del self.safety_checks[i]
                return True
        return False

    def create_todo(self, todo: Todo) -> Todo:
        todo = deepcopy(todo)
        validate_todo(todo)
        
        if not self.get_mine(todo.mine_id):
            raise ValidationError(f"矿区 ID {todo.mine_id} 不存在")
        
        if todo.safety_check_id and not self.get_safety_check(todo.safety_check_id):
            raise ValidationError(f"安全检查 ID {todo.safety_check_id} 不存在")
        
        todo.id = self._next_todo_id
        self.todos.append(todo)
        self._next_todo_id += 1
        return deepcopy(todo)

    def get_todo(self, todo_id: int) -> Optional[Todo]:
        for t in self.todos:
            if t.id == todo_id:
                return deepcopy(t)
        return None

    def list_todos(self, mine_id: Optional[int] = None, status: Optional[TodoStatus] = None) -> List[Todo]:
        result = []
        for t in self.todos:
            if mine_id is None or t.mine_id == mine_id:
                if status is None or t.status == status:
                    result.append(deepcopy(t))
        return result

    def update_todo(self, todo: Todo) -> Optional[Todo]:
        for i, existing in enumerate(self.todos):
            if existing.id == todo.id:
                validate_todo(todo)
                if not self.get_mine(todo.mine_id):
                    raise ValidationError(f"矿区 ID {todo.mine_id} 不存在")
                if todo.safety_check_id and not self.get_safety_check(todo.safety_check_id):
                    raise ValidationError(f"安全检查 ID {todo.safety_check_id} 不存在")
                self.todos[i] = deepcopy(todo)
                return deepcopy(todo)
        return None

    def delete_todo(self, todo_id: int) -> bool:
        for i, t in enumerate(self.todos):
            if t.id == todo_id:
                del self.todos[i]
                return True
        return False

    def upgrade_overdue_todos(self) -> int:
        today = date.today()
        upgraded = 0
        for i, todo in enumerate(self.todos):
            if todo.status != TodoStatus.COMPLETED and todo.deadline < today:
                if todo.priority == TodoPriority.LOW:
                    self.todos[i].priority = TodoPriority.MEDIUM
                elif todo.priority == TodoPriority.MEDIUM:
                    self.todos[i].priority = TodoPriority.HIGH
                elif todo.priority == TodoPriority.HIGH:
                    self.todos[i].priority = TodoPriority.URGENT
                upgraded += 1
        return upgraded
