import asyncio
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from main import ConfigCreate, ConfigUpdate

def test_validation():
    print("=== 开始测试校验规则 ===")
    
    # 测试1: 正常配置
    print("\n1. 测试正常配置")
    try:
        config = ConfigCreate(
            app_name="test",
            key="valid_key",
            value="normal_value"
        )
        print(f"   ✓ 正常配置通过校验")
    except Exception as e:
        print(f"   ✗ 正常配置校验失败: {e}")
    
    # 测试2: key 为空
    print("\n2. 测试 key 为空")
    try:
        config = ConfigCreate(
            app_name="test",
            key="",
            value="value"
        )
        print(f"   ✗ 应该拒绝空 key")
    except Exception as e:
        print(f"   ✓ 正确拒绝空 key: {e}")
    
    # 测试3: key 包含空格
    print("\n3. 测试 key 包含空格")
    try:
        config = ConfigCreate(
            app_name="test",
            key="invalid key",
            value="value"
        )
        print(f"   ✗ 应该拒绝包含空格的 key")
    except Exception as e:
        print(f"   ✓ 正确拒绝空格 key: {e}")
    
    # 测试4: key 包含换行
    print("\n4. 测试 key 包含换行")
    try:
        config = ConfigCreate(
            app_name="test",
            key="invalid\nkey",
            value="value"
        )
        print(f"   ✗ 应该拒绝包含换行的 key")
    except Exception as e:
        print(f"   ✓ 正确拒绝换行 key: {e}")
    
    # 测试5: value 超过 10KB
    print("\n5. 测试 value 超过 10KB")
    large_value = "x" * (11 * 1024)  # 11KB
    try:
        config = ConfigCreate(
            app_name="test",
            key="valid_key",
            value=large_value
        )
        print(f"   ✗ 应该拒绝超过 10KB 的 value")
    except Exception as e:
        print(f"   ✓ 正确拒绝过大 value: {e}")
    
    # 测试6: value 刚好 10KB
    print("\n6. 测试 value 刚好 10KB")
    valid_value = "x" * (10 * 1024)  # 10KB
    try:
        config = ConfigCreate(
            app_name="test",
            key="valid_key",
            value=valid_value
        )
        print(f"   ✓ 10KB value 通过校验")
    except Exception as e:
        print(f"   ✗ 10KB value 校验失败: {e}")
    
    # 测试7: ConfigUpdate 校验
    print("\n7. 测试 ConfigUpdate 校验")
    try:
        update = ConfigUpdate(value="new_value")
        print(f"   ✓ ConfigUpdate 正常通过")
    except Exception as e:
        print(f"   ✗ ConfigUpdate 校验失败: {e}")
    
    print("\n=== 校验规则测试完成 ===")

if __name__ == "__main__":
    test_validation()
