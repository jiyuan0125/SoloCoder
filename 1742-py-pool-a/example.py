import asyncio
from pool_manager import PoolManager
from pool_config import PoolConfig
from connection import MockConnection


async def create_mock_connection():
    conn = MockConnection()
    return conn


async def main():
    manager = PoolManager()
    
    await manager.create_pool(
        name="user_db",
        config=PoolConfig(
            max_connections=5,
            idle_timeout=60,
            wait_timeout=10,
            min_usage_count=1,
            min_connections=2,
            idle_recycle_interval=60,
            create_connection=create_mock_connection
        )
    )
    
    await manager.create_pool(
        name="order_db",
        config=PoolConfig(
            max_connections=10,
            idle_timeout=120,
            wait_timeout=30,
            min_usage_count=2,
            min_connections=3,
            idle_recycle_interval=60,
            create_connection=create_mock_connection
        )
    )
    
    print("Initial stats:")
    stats = manager.get_all_stats()
    for name, s in stats.items():
        print(f"  {name}: active={s.active_connections}, idle={s.idle_connections}")
    
    async with manager.connection("user_db") as conn1:
        async with manager.connection("user_db") as conn2:
            async with manager.connection("order_db") as conn3:
                print("\nAfter acquiring 3 connections:")
                stats = manager.get_all_stats()
                for name, s in stats.items():
                    print(f"  {name}: active={s.active_connections}, idle={s.idle_connections}")
    
    print("\nAfter releasing all connections:")
    stats = manager.get_all_stats()
    for name, s in stats.items():
        print(f"  {name}: active={s.active_connections}, idle={s.idle_connections}")
    
    await manager.close_all()


if __name__ == "__main__":
    asyncio.run(main())
