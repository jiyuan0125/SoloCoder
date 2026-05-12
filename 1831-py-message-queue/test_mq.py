import time
import asyncio
import aiohttp

BASE_URL = "http://localhost:8000"


async def test_publish_consume_ack():
    print("=" * 60)
    print("Test 1: Basic publish, consume, and ack")
    print("=" * 60)

    async with aiohttp.ClientSession() as session:
        for i in range(5):
            async with session.post(
                f"{BASE_URL}/topics/test_topic_1/messages",
                json={"content": f"message_{i}"}
            ) as resp:
                result = await resp.json()
                print(f"Published message_{i}: {result['message_id']}")

        async with session.get(
            f"{BASE_URL}/topics/test_topic_1/consume",
            params={"group": "group1", "limit": 3}
        ) as resp:
            result = await resp.json()
            messages = result["messages"]
            print(f"Consumed {len(messages)} messages")
            for msg in messages:
                print(f"  - {msg['content']}")

            msg_ids = [msg["id"] for msg in messages]
            async with session.post(
                f"{BASE_URL}/topics/test_topic_1/ack",
                json={"message_ids": msg_ids}
            ) as resp:
                result = await resp.json()
                print(f"Acknowledged {result['count']} messages")

        async with session.get(
            f"{BASE_URL}/topics/test_topic_1/consume",
            params={"group": "group1", "limit": 10}
        ) as resp:
            result = await resp.json()
            messages = result["messages"]
            print(f"After ack, remaining messages: {len(messages)}")
            for msg in messages:
                print(f"  - {msg['content']}")


async def test_multiple_groups():
    print("\n" + "=" * 60)
    print("Test 2: Multiple consumer groups (independent consumption)")
    print("=" * 60)

    async with aiohttp.ClientSession() as session:
        for i in range(3):
            async with session.post(
                f"{BASE_URL}/topics/test_topic_2/messages",
                json={"content": f"multi_msg_{i}"}
            ) as resp:
                result = await resp.json()
                print(f"Published multi_msg_{i}")

        async with session.get(
            f"{BASE_URL}/topics/test_topic_2/consume",
            params={"group": "group_a", "limit": 10}
        ) as resp:
            result = await resp.json()
            print(f"Group A consumed: {len(result['messages'])} messages")

        async with session.get(
            f"{BASE_URL}/topics/test_topic_2/consume",
            params={"group": "group_b", "limit": 10}
        ) as resp:
            result = await resp.json()
            print(f"Group B consumed: {len(result['messages'])} messages")


async def test_no_repeat_delivery():
    print("\n" + "=" * 60)
    print("Test 3: No repeat delivery in same group")
    print("=" * 60)

    async with aiohttp.ClientSession() as session:
        async with session.post(
            f"{BASE_URL}/topics/test_topic_3/messages",
            json={"content": "single_msg"}
        ) as resp:
            result = await resp.json()
            print(f"Published: {result['message_id']}")

        async with session.get(
            f"{BASE_URL}/topics/test_topic_3/consume",
            params={"group": "same_group", "limit": 10}
        ) as resp:
            result = await resp.json()
            print(f"First consume: {len(result['messages'])} messages")

        async with session.get(
            f"{BASE_URL}/topics/test_topic_3/consume",
            params={"group": "same_group", "limit": 10}
        ) as resp:
            result = await resp.json()
            print(f"Second consume (before ack): {len(result['messages'])} messages")
            if len(result["messages"]) == 0:
                print("  ✓ Correct: Message not repeated while pending")


async def test_query():
    print("\n" + "=" * 60)
    print("Test 4: Query topics status")
    print("=" * 60)

    async with aiohttp.ClientSession() as session:
        async with session.get(f"{BASE_URL}/topics") as resp:
            result = await resp.json()
            print(f"Topics: {len(result['topics'])}")
            for topic in result["topics"]:
                print(f"  - {topic['name']}: {topic['total_messages']} messages")
                for group in topic["consumer_groups"]:
                    print(f"      {group['group']}: lag={group['lag']}, seq={group['last_acked_seq']}")


async def test_timeout():
    print("\n" + "=" * 60)
    print("Test 5: 30-second timeout (quick test - check pending)")
    print("=" * 60)

    async with aiohttp.ClientSession() as session:
        async with session.post(
            f"{BASE_URL}/topics/test_topic_timeout/messages",
            json={"content": "timeout_msg"}
        ) as resp:
            result = await resp.json()
            print(f"Published: {result['message_id']}")

        async with session.get(
            f"{BASE_URL}/topics/test_topic_timeout/consume",
            params={"group": "timeout_group", "limit": 10}
        ) as resp:
            result = await resp.json()
            print(f"Consumed: {len(result['messages'])} messages (should be 1)")

        async with session.get(f"{BASE_URL}/topics") as resp:
            result = await resp.json()
            for topic in result["topics"]:
                if topic["name"] == "test_topic_timeout":
                    print(f"Pending messages: {topic['pending_messages']} (should be 1)")


async def main():
    print("Starting Message Queue Tests...\n")
    
    await test_publish_consume_ack()
    await test_multiple_groups()
    await test_no_repeat_delivery()
    await test_query()
    await test_timeout()
    
    print("\n" + "=" * 60)
    print("All basic tests completed!")
    print("Note: Full 30-second timeout would require waiting...")
    print("=" * 60)


if __name__ == "__main__":
    asyncio.run(main())
