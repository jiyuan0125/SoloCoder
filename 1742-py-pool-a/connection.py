from abc import ABC, abstractmethod
from typing import Any


class Connection(ABC):
    @abstractmethod
    async def close(self) -> None:
        pass

    @abstractmethod
    async def is_healthy(self) -> bool:
        pass

    @property
    @abstractmethod
    def raw(self) -> Any:
        pass


class MockConnection(Connection):
    def __init__(self):
        self._closed = False

    async def close(self) -> None:
        self._closed = True

    async def is_healthy(self) -> bool:
        return not self._closed

    @property
    def raw(self) -> Any:
        return self


class PostgreSQLConnection(Connection):
    def __init__(self, connection_params: dict):
        self._connection_params = connection_params
        self._conn = None

    async def connect(self) -> None:
        import psycopg2
        self._conn = psycopg2.connect(**self._connection_params)

    async def close(self) -> None:
        if self._conn:
            self._conn.close()
            self._conn = None

    async def is_healthy(self) -> bool:
        if not self._conn:
            return False
        try:
            cursor = self._conn.cursor()
            cursor.execute("SELECT 1")
            cursor.close()
            return True
        except Exception:
            return False

    @property
    def raw(self) -> Any:
        return self._conn


class MySQLConnection(Connection):
    def __init__(self, connection_params: dict):
        self._connection_params = connection_params
        self._conn = None

    async def connect(self) -> None:
        import pymysql
        self._conn = pymysql.connect(**self._connection_params)

    async def close(self) -> None:
        if self._conn:
            self._conn.close()
            self._conn = None

    async def is_healthy(self) -> bool:
        if not self._conn:
            return False
        try:
            self._conn.ping(reconnect=False)
            return True
        except Exception:
            return False

    @property
    def raw(self) -> Any:
        return self._conn
