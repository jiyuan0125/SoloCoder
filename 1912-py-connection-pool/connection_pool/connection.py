import time
import socket
import httpx
from typing import Optional, Union
from dataclasses import dataclass, field
from .models import ConnectionStatus, DatasourceType


@dataclass
class Connection:
    id: str
    datasource_name: str
    type: DatasourceType
    status: ConnectionStatus = ConnectionStatus.IDLE
    last_used_time: float = field(default_factory=time.time)
    borrowed_time: Optional[float] = None
    last_validation_time: Optional[float] = None
    _conn: Union[httpx.Client, socket.socket, None] = None

    def _create_connection(self):
        if self.type == DatasourceType.HTTP:
            self._conn = httpx.Client(timeout=5.0)
        elif self.type == DatasourceType.TCP:
            self._conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self._conn.settimeout(5.0)

    def connect(self):
        self._create_connection()
        self.status = ConnectionStatus.IDLE
        self.last_used_time = time.time()
        self.last_validation_time = time.time()

    def ping(self) -> bool:
        if self._conn is None:
            return False
        try:
            if self.type == DatasourceType.HTTP:
                response = self._conn.head("http://example.com", timeout=5.0)
                return response.status_code < 500
            elif self.type == DatasourceType.TCP:
                self._conn.getsockopt(socket.SOL_SOCKET, socket.SO_KEEPALIVE)
                return True
        except Exception:
            return False
        return True

    def close(self):
        try:
            if self._conn:
                if self.type == DatasourceType.HTTP:
                    self._conn.close()
                elif self.type == DatasourceType.TCP:
                    self._conn.close()
        except Exception:
            pass
        finally:
            self._conn = None
            self.status = ConnectionStatus.DISCONNECTED

    def __del__(self):
        self.close()
