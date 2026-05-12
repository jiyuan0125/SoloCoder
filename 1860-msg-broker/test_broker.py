import asyncio
import httpx
import time
import json


async def test_broker():
    base_url = "http://localhost:8000"
    
    async with httpx.AsyncClient(timeout=30.0) as client:
        print("=== 1. 创建 Topic ===")
        resp = await client.post(f"{base_url}/topics/order-events", json={"capacity": 50000})
        print(f"创建 Topic 响应: {resp.status_code} - {resp.text}")
        
        print("\n=== 2. 获取 Topic 列表 ===")
        resp = await client.get(f"{base_url}/topics")
        print(f"Topic 列表: {resp.json()}")
        
        print("\n=== 3. 发送消息 ===")
        for i in range(5):
            resp = await client.post(
                f"{base_url}/topics/order-events/messages",
                json={"content": {"order_id": f"order-{i}", "status": "paid"}},
                timeout=30.0
            )
            if resp.status_code == 200:
                print(f"发送消息 {i}: {resp.json()}")
            else:
                print(f"发送消息 {i} 失败: {resp.status_code} - {resp.text}")
        
        print("\n=== 4. Group-1 消费消息 ===")
        resp = await client.get(
            f"{base_url}/topics/order-events/consume?group=group-1&limit=10",
            headers={"X-Consumer-Id": "consumer-a"}
        )
        data = resp.json()
        print(f"Group-1 消费到 {len(data['messages'])} 条消息")
        msg_ids = [m["id"] for m in data["messages"]]
        print(f"消息 IDs: {msg_ids}")
        
        print("\n=== 5. ACK 消息 ===")
        if msg_ids:
            resp = await client.post(
                f"{base_url}/topics/order-events/ack",
                json={"message_ids": msg_ids[:2]},
                headers={"X-Group": "group-1"}
            )
            print(f"ACK 响应: {resp.json()}")
        
        print("\n=== 6. 检查 Group 消费进度 ===")
        resp = await client.get(f"{base_url}/topics/order-events/groups")
        print(f"消费进度: {resp.json()}")
        
        print("\n=== 7. 检查死信队列 ===")
        resp = await client.get(f"{base_url}/topics/order-events/deadletter")
        print(f"死信队列: {resp.json()}")
        
        print("\n=== 8. 测试非法 JSON 消息 ===")
        resp = await client.post(
            f"{base_url}/topics/order-events/messages",
            content="this is not json",
            headers={"Content-Type": "application/json"}
        )
        print(f"非法 JSON 响应: {resp.status_code} - {resp.text}")
        
        print("\n=== 9. 测试消息体大小 (5MB 模拟) ===")
        large_content = "x" * (1024 * 1024 * 2)
        resp = await client.post(
            f"{base_url}/topics/order-events/messages",
            json={"content": {"large_data": large_content}}
        )
        print(f"大消息响应: {resp.status_code}")
        if resp.status_code == 200:
            print(f"大消息 ID: {resp.json()}")
        
        print("\n=== 10. 多消费者同一 Group (不重复投递) ===")
        for i in range(10):
            await client.post(
                f"{base_url}/topics/order-events/messages",
                json={"content": {"test": i}}
            )
        
        consumer_a = await client.get(
            f"{base_url}/topics/order-events/consume?group=group-2&limit=10",
            headers={"X-Consumer-Id": "consumer-a"}
        )
        consumer_b = await client.get(
            f"{base_url}/topics/order-events/consume?group=group-2&limit=10",
            headers={"X-Consumer-Id": "consumer-b"}
        )
        
        a_msgs = {m["id"] for m in consumer_a.json()["messages"]}
        b_msgs = {m["id"] for m in consumer_b.json()["messages"]}
        print(f"Consumer A 拿到: {len(a_msgs)} 条")
        print(f"Consumer B 拿到: {len(b_msgs)} 条")
        print(f"交集 (应该为空): {a_msgs & b_msgs}")
        
        print("\n=== 所有测试完成 ===")


if __name__ == "__main__":
    asyncio.run(test_broker())
