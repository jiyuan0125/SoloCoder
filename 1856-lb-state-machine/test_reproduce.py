import asyncio
import time
from app import NodeState, Node, StateMachine, CallbackNotifier


class MockCallbackNotifier:
    def __init__(self):
        self.notifications = []

    async def start(self):
        pass

    async def stop(self):
        pass

    async def add_callback(self, url):
        pass

    async def notify(self, node_id, old_state, new_state, timestamp):
        self.notifications.append({
            'node_id': node_id,
            'old_state': old_state,
            'new_state': new_state,
            'timestamp': timestamp
        })
        print(f"[NOTIFY] Node {node_id[-8:]}: {old_state} -> {new_state}")


async def test_warming_weight():
    print("\n=== Test 1: Warming weight (预热权重应为配置的50%) ===")
    notifier = MockCallbackNotifier()
    sm = StateMachine(notifier)
    
    node_id = await sm.register_node("127.0.0.1:5001", 100, "/health")
    
    node = sm.nodes[node_id]
    print(f"初始状态: {node.state}, effective_weight: {node.get_effective_weight()}")
    
    for i in range(3):
        await sm.process_health_check_result(node_id, True)
    
    print(f"3次健康检查后, 状态: {node.state}, effective_weight: {node.get_effective_weight()}")
    
    if node.state == NodeState.WARMING:
        expected_effective = 100 // 2
        actual_effective = node.get_effective_weight()
        print(f"预期 effective_weight: {expected_effective}, 实际: {actual_effective}")
        assert actual_effective == expected_effective, f"权重不匹配! 预期 {expected_effective}, 实际 {actual_effective}"
        print("✓ 预热权重测试通过!")
        return True
    else:
        print(f"✗ 失败: 预期 WARMING 状态, 实际 {node.state}")
        return False


async def test_health_check_isolation():
    print("\n=== Test 2: 健康检查隔离性 (两个节点状态不应互相污染) ===")
    notifier = MockCallbackNotifier()
    sm = StateMachine(notifier)
    
    node1_id = await sm.register_node("127.0.0.1:9001", 100, "/health")
    node2_id = await sm.register_node("127.0.0.1:9002", 100, "/health")
    
    node1 = sm.nodes[node1_id]
    node2 = sm.nodes[node2_id]
    
    print(f"Node1 初始: state={node1.state}, success={node1.consecutive_success}")
    print(f"Node2 初始: state={node2.state}, success={node2.consecutive_success}")
    
    for i in range(3):
        await sm.process_health_check_result(node1_id, False)
        await sm.process_health_check_result(node2_id, True)
    
    print(f"\n3轮检查后 (Node1 失败, Node2 成功):")
    print(f"Node1: state={node1.state}, fail={node1.consecutive_failure}, success={node1.consecutive_success}")
    print(f"Node2: state={node2.state}, fail={node2.consecutive_failure}, success={node2.consecutive_success}")
    
    try:
        assert node1.consecutive_failure == 3, f"Node1 应该有 3 次失败, 实际 {node1.consecutive_failure}"
        assert node1.consecutive_success == 0, f"Node1 应该有 0 次成功, 实际 {node1.consecutive_success}"
        assert node2.state == NodeState.WARMING, f"Node2 应该是 WARMING, 实际 {node2.state}"
        print("✓ 健康检查隔离性测试通过!")
        return True
    except AssertionError as e:
        print(f"✗ 测试失败: {e}")
        return False


async def test_active_to_suspected():
    print("\n=== Test 3: 活跃 -> 疑似故障 ===")
    notifier = MockCallbackNotifier()
    sm = StateMachine(notifier)
    
    node_id = await sm.register_node("127.0.0.1:5001", 100, "/health")
    node = sm.nodes[node_id]
    
    for i in range(3):
        await sm.process_health_check_result(node_id, True)
    
    assert node.state == NodeState.WARMING, f"预期 WARMING, 实际 {node.state}"
    
    node.state = NodeState.ACTIVE
    node.warm_start_time = time.time() - 40
    
    print(f"节点设为 ACTIVE 状态")
    
    for i in range(2):
        await sm.process_health_check_result(node_id, False)
    
    print(f"2次失败后: state={node.state}, fail_count={node.consecutive_failure}")
    
    try:
        assert node.state == NodeState.SUSPECTED, f"预期 SUSPECTED, 实际 {node.state}"
        print("✓ 活跃到疑似故障测试通过!")
        return True
    except AssertionError as e:
        print(f"✗ 测试失败: {e}")
        return False


async def test_suspected_recovery():
    print("\n=== Test 4: 疑似故障恢复 ===")
    notifier = MockCallbackNotifier()
    sm = StateMachine(notifier)
    
    node_id = await sm.register_node("127.0.0.1:5001", 100, "/health")
    node = sm.nodes[node_id]
    
    node.state = NodeState.SUSPECTED
    
    for i in range(3):
        await sm.process_health_check_result(node_id, True)
    
    print(f"3次成功后: state={node.state}")
    
    try:
        assert node.state == NodeState.ACTIVE, f"预期 ACTIVE, 实际 {node.state}"
        print("✓ 疑似故障恢复测试通过!")
        return True
    except AssertionError as e:
        print(f"✗ 测试失败: {e}")
        return False


async def test_load_balancer_stability():
    print("\n=== Test 5: 负载均衡稳定性 (连续10次 pick 不应失败) ===")
    from app import LoadBalancer
    
    notifier = MockCallbackNotifier()
    sm = StateMachine(notifier)
    
    node1_id = await sm.register_node("127.0.0.1:9001", 100, "/health")
    node2_id = await sm.register_node("127.0.0.1:9002", 100, "/health")
    
    node1 = sm.nodes[node1_id]
    node2 = sm.nodes[node2_id]
    
    for i in range(3):
        await sm.process_health_check_result(node1_id, True)
        await sm.process_health_check_result(node2_id, True)
    
    node1.state = NodeState.ACTIVE
    node2.state = NodeState.ACTIVE
    
    lb = LoadBalancer(sm)
    
    results = []
    for i in range(10):
        node = lb.select_node()
        if node:
            results.append(node.id)
            print(f"Pick {i+1}: {node.id[-8:]} ({node.state.value})")
        else:
            print(f"Pick {i+1}: None (错误)")
    
    try:
        assert len(results) == 10, f"预期 10 次成功选择, 实际 {len(results)}"
        print("✓ 负载均衡稳定性测试通过!")
        return True
    except AssertionError as e:
        print(f"✗ 测试失败: {e}")
        return False


async def test_warming_to_active_timeout():
    print("\n=== Test 6: 预热超时自动转活跃 ===")
    notifier = MockCallbackNotifier()
    sm = StateMachine(notifier)
    
    node_id = await sm.register_node("127.0.0.1:5001", 100, "/health")
    node = sm.nodes[node_id]
    
    for i in range(3):
        await sm.process_health_check_result(node_id, True)
    
    assert node.state == NodeState.WARMING, f"预期 WARMING, 实际 {node.state}"
    print(f"进入预热状态, effective_weight: {node.get_effective_weight()}")
    
    node.warm_start_time = time.time() - 35
    
    await sm.check_warming_timeout()
    
    print(f"超时检查后: state={node.state}, effective_weight={node.get_effective_weight()}")
    
    try:
        assert node.state == NodeState.ACTIVE, f"预期 ACTIVE, 实际 {node.state}"
        assert node.get_effective_weight() == 100, f"有效权重应为 100, 实际 {node.get_effective_weight()}"
        print("✓ 预热超时自动转活跃测试通过!")
        return True
    except AssertionError as e:
        print(f"✗ 测试失败: {e}")
        return False


async def test_two_nodes_health_check_scenarios():
    print("\n=== Test 7: 双节点健康检查场景 (模拟9001宕机，9002正常) ===")
    notifier = MockCallbackNotifier()
    sm = StateMachine(notifier)
    
    node9001_id = await sm.register_node("127.0.0.1:9001", 100, "/health")
    node9002_id = await sm.register_node("127.0.0.1:9002", 100, "/health")
    
    node9001 = sm.nodes[node9001_id]
    node9002 = sm.nodes[node9002_id]
    
    print(f"阶段1: 两个节点都健康 - 预热到活跃")
    for round_num in range(4):
        await sm.process_health_check_result(node9001_id, True)
        await sm.process_health_check_result(node9002_id, True)
    
    if node9001.state == NodeState.WARMING:
        node9001.warm_start_time = time.time() - 35
        await sm.check_warming_timeout()
    
    if node9002.state == NodeState.WARMING:
        node9002.warm_start_time = time.time() - 35
        await sm.check_warming_timeout()
    
    print(f"  9001: state={node9001.state.value}, effective={node9001.get_effective_weight()}")
    print(f"  9002: state={node9002.state.value}, effective={node9002.get_effective_weight()}")
    
    print(f"\n阶段2: 模拟9001宕机, 9002继续健康 - 持续10轮检查")
    for round_num in range(10):
        await sm.process_health_check_result(node9001_id, False)
        await sm.process_health_check_result(node9002_id, True)
    
    print(f"  9001: state={node9001.state.value}, fail={node9001.consecutive_failure}, success={node9001.consecutive_success}")
    print(f"  9002: state={node9002.state.value}, fail={node9002.consecutive_failure}, success={node9002.consecutive_success}")
    
    try:
        assert node9001.state == NodeState.SUSPECTED, f"9001 应为 SUSPECTED, 实际 {node9001.state}"
        assert node9002.state == NodeState.ACTIVE, f"9002 应为 ACTIVE, 实际 {node9002.state}"
        print("✓ 双节点健康检查场景测试通过!")
        return True
    except AssertionError as e:
        print(f"✗ 测试失败: {e}")
        return False


async def main():
    print("=" * 70)
    print("Bug 修复验证测试")
    print("=" * 70)
    
    tests = [
        test_warming_weight,
        test_health_check_isolation,
        test_active_to_suspected,
        test_suspected_recovery,
        test_load_balancer_stability,
        test_warming_to_active_timeout,
        test_two_nodes_health_check_scenarios,
    ]
    
    passed = 0
    failed = 0
    
    for test in tests:
        try:
            if await test():
                passed += 1
            else:
                failed += 1
        except Exception as e:
            print(f"✗ 测试异常: {e}")
            import traceback
            traceback.print_exc()
            failed += 1
    
    print("\n" + "=" * 70)
    print(f"测试结果: {passed} 通过, {failed} 失败")
    print("=" * 70)
    
    return failed == 0


if __name__ == "__main__":
    success = asyncio.run(main())
    exit(0 if success else 1)
