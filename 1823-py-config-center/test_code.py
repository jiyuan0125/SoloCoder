import asyncio
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from main import ConfigCenter, ConfigItem, ConfigVersion, Subscriber
from datetime import datetime, timezone

async def test_config_center():
    print("=== 开始测试配置中心 ===")
    
    # 初始化配置中心
    center = ConfigCenter()
    
    # 测试1: 创建配置
    print("\n1. 测试创建配置")
    config = await center.create_config(
        app_name="test_app",
        key="db_connection",
        value="mysql://localhost:3306/test"
    )
    print(f"   创建成功: {config.app_name}/{config.key}, 版本: {config.current_version}")
    assert config.current_version == 1
    print("   ✓ 版本号从 1 开始，正确")
    
    # 测试2: 获取当前版本配置
    print("\n2. 测试获取配置")
    current = await center.get_config(app_name="test_app", key="db_connection")
    print(f"   当前版本: {current.version}, 值: {current.value}")
    assert current.version == 1
    print("   ✓ 获取成功")
    
    # 测试3: 更新配置
    print("\n3. 测试更新配置")
    new_version = await center.update_config(
        app_name="test_app",
        key="db_connection",
        value="postgresql://localhost:5432/test"
    )
    print(f"   更新成功，新版本: {new_version.version}")
    assert new_version.version == 2
    print("   ✓ 版本号自动递增")
    
    # 测试4: 获取指定版本
    print("\n4. 测试获取指定版本")
    v1 = await center.get_config(app_name="test_app", key="db_connection", version=1)
    print(f"   版本 1 的值: {v1.value}")
    assert v1.value == "mysql://localhost:3306/test"
    print("   ✓ 指定版本获取成功")
    
    # 测试5: 版本历史
    print("\n5. 测试版本历史")
    history = await center.get_version_history(app_name="test_app", key="db_connection")
    print(f"   历史版本数量: {len(history)}")
    assert len(history) == 2
    print("   ✓ 版本历史保留完整")
    
    # 测试6: 回滚
    print("\n6. 测试回滚")
    rollback_version = await center.rollback_config(
        app_name="test_app",
        key="db_connection",
        target_version=1
    )
    print(f"   回滚到版本 1，创建新版本: {rollback_version.version}")
    print(f"   新版本值: {rollback_version.value}")
    print(f"   是否是回滚: {rollback_version.is_rollback}")
    print(f"   回滚自版本: {rollback_version.rollback_from_version}")
    assert rollback_version.version == 3
    assert rollback_version.is_rollback == True
    assert rollback_version.rollback_from_version == 1
    print("   ✓ 回滚创建新版本，历史记录可追溯")
    
    # 测试7: 订阅者管理
    print("\n7. 测试订阅者管理")
    subscriber = await center.register_subscriber(
        app_name="test_app",
        callback_url="http://localhost:9000/webhook"
    )
    print(f"   订阅者注册成功: {subscriber.callback_url}")
    
    subscribers = await center.list_subscribers("test_app")
    print(f"   订阅者列表: {len(subscribers)} 个")
    assert len(subscribers) == 1
    print("   ✓ 订阅者管理正常")
    
    # 测试8: 列出配置
    print("\n8. 测试列出应用配置")
    configs = await center.list_configs("test_app")
    print(f"   test_app 配置项: {configs}")
    assert "db_connection" in configs
    print("   ✓ 配置列表正常")
    
    # 测试9: 删除配置
    print("\n9. 测试删除配置")
    await center.delete_config(app_name="test_app", key="db_connection")
    configs_after = await center.list_configs("test_app")
    print(f"   删除后配置数量: {len(configs_after)}")
    assert len(configs_after) == 0
    print("   ✓ 删除成功")
    
    print("\n=== 所有测试通过! ===")

if __name__ == "__main__":
    asyncio.run(test_config_center())
