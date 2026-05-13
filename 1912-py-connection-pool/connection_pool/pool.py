import time
import uuid
import threading
from typing import Dict, Optional, List, Callable
from collections import deque
from .models import ConnectionStatus, DatasourceType, ConnectionStats
from .connection import Connection


class ConnectionPool:
    def __init__(self, name: str, type: DatasourceType, max_connections: int, idle_timeout: int):
        self.name = name
        self.type = type
        self.max_connections = max_connections
        self.idle_timeout = idle_timeout
        
        self._lock = threading.RLock()
        self._connections: Dict[str, Connection] = {}
        self._idle_ids: deque = deque()
        self._in_use_ids: set = set()
        self._on_connection_replaced: Optional[Callable] = None

    def _new_connection(self) -> Connection:
        conn_id = str(uuid.uuid4())
        conn = Connection(id=conn_id, datasource_name=self.name, type=self.type)
        conn.connect()
        return conn

    def acquire(self) -> Optional[Connection]:
        with self._lock:
            while self._idle_ids:
                conn_id = self._idle_ids.popleft()
                conn = self._connections.get(conn_id)
                if conn and conn.status == ConnectionStatus.IDLE:
                    conn.status = ConnectionStatus.IN_USE
                    conn.borrowed_time = time.time()
                    self._in_use_ids.add(conn_id)
                    return conn
            
            if self.current_size() < self.max_connections:
                conn = self._new_connection()
                self._connections[conn.id] = conn
                conn.status = ConnectionStatus.IN_USE
                conn.borrowed_time = time.time()
                self._in_use_ids.add(conn.id)
                return conn
            
            return None

    def release(self, conn_id: str) -> bool:
        with self._lock:
            conn = self._connections.get(conn_id)
            if not conn or conn.status != ConnectionStatus.IN_USE:
                return False
            
            self._in_use_ids.discard(conn_id)
            conn.last_used_time = time.time()
            
            if self.current_size() > self.max_connections:
                conn.close()
                del self._connections[conn_id]
                return True
            
            if conn.ping():
                conn.status = ConnectionStatus.IDLE
                conn.borrowed_time = None
                self._idle_ids.append(conn_id)
            else:
                conn.close()
                del self._connections[conn_id]
                
        return True

    def validate_idle_connections(self) -> int:
        removed_count = 0
        with self._lock:
            idle_connections = list(self._idle_ids)
            for conn_id in idle_connections:
                conn = self._connections.get(conn_id)
                if conn and conn.status == ConnectionStatus.IDLE:
                    conn.status = ConnectionStatus.VALIDATING
                    if conn.ping():
                        conn.last_validation_time = time.time()
                        conn.status = ConnectionStatus.IDLE
                    else:
                        conn.close()
                        del self._connections[conn_id]
                        self._idle_ids = deque(cid for cid in self._idle_ids if cid != conn_id)
                        removed_count += 1
        return removed_count

    def check_leaks(self) -> int:
        leak_count = 0
        now = time.time()
        with self._lock:
            in_use_copy = list(self._in_use_ids)
            for conn_id in in_use_copy:
                conn = self._connections.get(conn_id)
                if conn and conn.status == ConnectionStatus.IN_USE:
                    if conn.borrowed_time and (now - conn.borrowed_time) > 60:
                        conn.close()
                        del self._connections[conn_id]
                        self._in_use_ids.discard(conn_id)
                        leak_count += 1
        return leak_count

    def current_size(self) -> int:
        with self._lock:
            return len(self._connections)

    def get_stats(self) -> ConnectionStats:
        with self._lock:
            idle = 0
            in_use = 0
            validating = 0
            disconnected = 0
            for conn in self._connections.values():
                if conn.status == ConnectionStatus.IDLE:
                    idle += 1
                elif conn.status == ConnectionStatus.IN_USE:
                    in_use += 1
                elif conn.status == ConnectionStatus.VALIDATING:
                    validating += 1
                elif conn.status == ConnectionStatus.DISCONNECTED:
                    disconnected += 1
            return ConnectionStats(idle=idle, in_use=in_use, validating=validating, disconnected=disconnected)

    def update_max_connections(self, new_max: int):
        with self._lock:
            self.max_connections = new_max

    def get_all_connections(self) -> List[Connection]:
        with self._lock:
            return list(self._connections.values())
