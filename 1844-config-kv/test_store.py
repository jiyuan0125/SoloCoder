from config_store import ConfigStore
import json
import os
import shutil

shutil.rmtree("test_data", ignore_errors=True)
store = ConfigStore("test_data")

# 创建 base 分组
store.put_config("order-service", "base", {"timeout": 30, "debug": True})
print("Base config:", store.get_config("order-service", "base"))

# 创建 production 分组，继承 base
store.put_config("order-service", "production", {"debug": False, "max_pool_size": 10}, parent="base")
print("Production config (should include timeout from base):")
print(store.get_config("order-service", "production"))

# 检查数据文件
with open("test_data/projects.json", "r") as f:
    print("\nProjects data:")
    print(json.dumps(json.load(f), indent=2))

shutil.rmtree("test_data", ignore_errors=True)
