import asyncio
import httpx
import time
import json


async def test_group_isolation():
    base_url = "http://localhost:8000"
    
    async with httpx.AsyncClient(timeout=30.0) as client:
        print("=== 测试 Group 隔离 ===")
        print("1. 创建新 Topic")
        resp = await client.post(f"{base_url}/topics/group-test", json={"capacity": 1000})
        print(f"   创建: {resp.status_code}")
        
        print("\n2. 发送 5 条消息")
        msg_ids = []
        for i in range(5):
            resp = await client.post(
                f"{base_url}/topics/group-test/messages",
                json={"content": {"msg_index": i}}
            )
            msg_ids.append(resp.json()["id"])
            print(f"   消息 {i}: ID={msg_ids[-1]}")
        
        print("\n3. Group-1 消费消息")
        resp = await client.get(
            f"{base_url}/topics/group-test/consume?group=group-1&limit=10",
            headers={"X-Consumer-Id": "consumer-g1"}
        )
        data = resp.json()
        g1_msg_ids = [m["id"] for m in data["messages"]]
        print(f"   Group-1 拿到: {len(g1_msg_ids)} 条 - {g1_msg_ids}")
        
        print("\n4. Group-2 消费消息 (应该拿到同样的 5 条)")
        resp = await client.get(
            f"{base_url}/topics/group-test/consume?group=group-2&limit=10",
            headers={"X-Consumer-Id": "consumer-g2"}
        )
        data = resp.json()
        g2_msg_ids = [m["id"] for m in data["messages"]]
        print(f"   Group-2 拿到: {len(g2_msg_ids)} 条 - {g2_msg_ids}")
        
        if set(g1_msg_ids) == set(g2_msg_ids) == set(msg_ids):
            print("   ✓ Group 隔离正确：两个 Group 都能看到所有消息")
        else:
            print(f"   ✗ Group 隔离错误！预期都拿到 {set(msg_ids)}")
        
        print("\n5. Group-1 ACK 前 2 条")
        resp = await client.post(
            f"{base_url}/topics/group-test/ack",
            json={"message_ids": g1_msg_ids[:2]},
            headers={"X-Group": "group-1"}
        )
        print(f"   ACK 结果: {resp.json()}")
        
        print("\n6. 检查各 Group 进度")
        resp = await client.get(f"{base_url}/topics/group-test/groups")
        progress = resp.json()
        for g in progress["groups"]:
            print(f"   {g['group']}: last_ack={g['last_ack_id']}, consumed={g['total_consumed']}, in_flight={g['in_flight']}")
        
        print("\n7. Group-3 加入消费 (应该拿到全部 5 条)")
        resp = await client.get(
            f"{base_url}/topics/group-test/consume?group=group-3&limit=10",
            headers={"X-Consumer-Id": "consumer-g3"}
        )
        data = resp.json()
        g3_msg_ids = [m["id"] for m in data["messages"]]
        print(f"   Group-3 拿到: {len(g3_msg_ids)} 条 - {g3_msg_ids}")
        
        if len(g3_msg_ids) == 5:
            print("   ✓ 新 Group 能看到所有消息")
        else:
            print(f"   ✗ 新 Group 应该拿到 5 条消息")
        
        print("\n=== Group 隔离测试完成 ===")
        return True


async def test_ack_timeout():
    base_url = "http://localhost:8000"
    
    async with httpx.AsyncClient(timeout=30.0) as client:
        print("\n=== 测试 ACK 超时重投递 ===")
        print("注意：实际 ACK 超时是 30 秒，这里演示逻辑")
        
        print("\n1. 确认后台任务在运行")
        resp = await client.get(f"{base_url}/docs")
        print(f"   服务健康: {resp.status_code == 200}")
        
        print("\n2. 创建超时测试 Topic")
        resp = await client.post(f"{base_url}/topics/timeout-test", json={"capacity": 100})
        print(f"   创建: {resp.status_code}")
        
        print("\n3. 发送 1 条测试消息")
        resp = await client.post(
            f"{base_url}/topics/timeout-test/messages",
            json={"content": {"test": "timeout-test"}}
        )
        msg_id = resp.json()["id"]
        print(f"   消息 ID: {msg_id}")
        
        print("\n4. Group-A 消费这条消息")
        resp = await client.get(
            f"{base_url}/topics/timeout-test/consume?group=group-a&limit=10",
            headers={"X-Consumer-Id": "consumer-a"}
        )
        data = resp.json()
        print(f"   拿到 {len(data['messages'])} 条消息")
        
        print("\n5. 检查 in_flight 状态")
        resp = await client.get(f"{base_url}/topics/timeout-test/groups")
        progress = resp.json()
        for g in progress["groups"]:
            print(f"   {g['group']}: in_flight={g['in_flight']}")
        
        print("\n6. 不 ACK，消息将在 30 秒后超时...")
        print("   (实际应用中请等待 30 秒后再次消费验证)")
        print("   代码逻辑说明：")
        print("   - consume() 时设置 expires_at = now + 30")
        print("   - process_timeouts() 每秒检查一次")
        print("   - 发现超时调用 nack()，设置 retry_count 和 next_retry_at")
        print("   - 重试 3 次后进入死信队列")
        
        print("\n=== ACK 超时逻辑说明完成 ===")
        print("   请修改 app/store.py 中的 ACK_TIMEOUT_SECONDS = 3")
        print("   重启服务后可快速验证超时重投递")
        return True


async def test_multi_consumer_same_group():
    base_url = "http://localhost:8000"
    
    async with httpx.AsyncClient(timeout=30.0) as client:
        print("\n=== 测试同一 Group 多消费者 (不重复投递) ===")
        
        print("1. 创建 Topic 并发送 10 条消息")
        await client.post(f"{base_url}/topics/multi-consumer", json={"capacity": 100})
        for i in range(10):
            await client.post(
                f"{base_url}/topics/multi-consumer/messages",
                json={"content": {"i": i}}
            )
        
        print("2. 两个消费者同时消费")
        consumer_a_resp, consumer_b_resp = await asyncio.gather(
            client.get(
                f"{base_url}/topics/multi-consumer/consume?group=balanced-group&limit=10",
                headers={"X-Consumer-Id": "consumer-a"}
            ),
            client.get(
                f"{base_url}/topics/multi-consumer/consume?group=balanced-group&limit=10",
                headers={"X-Consumer-Id": "consumer-b"}
            )
        )
        
        a_msgs = {m["id"] for m in consumer_a_resp.json()["messages"]}
        b_msgs = {m["id"] for m in consumer_b_resp.json()["messages"]}
        
        print(f"   Consumer A: {len(a_msgs)} 条")
        print(f"   Consumer B: {len(b_msgs)} 条")
        print(f"   交集: {a_msgs & b_msgs}")
        
        if len(a_msgs & b_msgs) == 0:
            print("   ✓ 同一 Group 内消息不重复投递")
        else:
            print("   ✗ 发现重复投递！")
        
        print("\n=== 多消费者测试完成 ===")
        return True


async def main():
    print("=" * 60)
    print("消息代理功能测试")
    print("=" * 60)
    
    await test_group_isolation()
    await test_ack_timeout()
    await test_multi_consumer_same_group()
    
    print("\n" + "=" * 60)
    print("所有测试完成")
    print("=" * 60)


if __name__ == "__main__":
    asyncio.run(main())
