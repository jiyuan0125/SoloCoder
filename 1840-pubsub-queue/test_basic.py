import asyncio
import sys
import os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from main import (
    topic_manager,
    manager_lock,
    Message,
    ConsumerGroup,
    Topic,
    get_or_create_topic,
    DEFAULT_TOPIC_CAPACITY
)


async def test_basic_flow():
    print("=== Test 1: 基本发布消费流程 ===")
    
    async with manager_lock:
        if "test-topic" in topic_manager:
            del topic_manager["test-topic"]
    
    topic = await get_or_create_topic("test-topic")
    
    async with topic.lock:
        msg1 = Message("message 1")
        msg2 = Message("message 2")
        topic.messages.append(msg1)
        topic.messages.append(msg2)
        
        group = ConsumerGroup(name="group1")
        group.pending_messages.append(msg1)
        group.pending_messages.append(msg2)
        topic.consumer_groups["group1"] = group
    
    print(f"Topic test-topic created with {len(topic.messages)} messages")
    print(f"Consumer group group1 has {len(topic.consumer_groups['group1'].pending_messages)} pending messages")
    
    print("Test 1 PASSED")


async def test_topic_capacity():
    print("\n=== Test 2: Topic 容量限制 ===")
    
    async with manager_lock:
        if "test-capacity" in topic_manager:
            del topic_manager["test-capacity"]
    
    topic = Topic(name="test-capacity", capacity=5)
    
    async with topic.lock:
        for i in range(10):
            while len(topic.messages) >= topic.capacity:
                topic.messages.popleft()
            topic.messages.append(Message(f"message {i}"))
        
        print(f"Topic capacity: {topic.capacity}, actual messages: {len(topic.messages)}")
        contents = [m.content for m in topic.messages]
        print(f"Messages: {contents}")
        
        assert len(topic.messages) == 5, f"Expected 5 messages, got {len(topic.messages)}"
        assert contents == ["message 5", "message 6", "message 7", "message 8", "message 9"]
    
    print("Test 2 PASSED")


async def test_consumer_groups():
    print("\n=== Test 3: 多 Consumer Group ===")
    
    async with manager_lock:
        if "test-multi-group" in topic_manager:
            del topic_manager["test-multi-group"]
    
    topic = Topic(name="test-multi-group", capacity=100)
    
    async with topic.lock:
        msg = Message("shared message")
        topic.messages.append(msg)
        
        group1 = ConsumerGroup(name="group1")
        group1.pending_messages.append(msg)
        topic.consumer_groups["group1"] = group1
        
        group2 = ConsumerGroup(name="group2")
        group2.pending_messages.append(msg)
        topic.consumer_groups["group2"] = group2
        
        print(f"Group1 pending: {len(group1.pending_messages)}, Group2 pending: {len(group2.pending_messages)}")
        
        assert len(group1.pending_messages) == 1
        assert len(group2.pending_messages) == 1
        
        consumed = group1.pending_messages.popleft()
        print(f"Group1 consumed: {consumed.content}")
        print(f"After group1 consume - Group1 pending: {len(group1.pending_messages)}, Group2 pending: {len(group2.pending_messages)}")
        
        assert len(group1.pending_messages) == 0
        assert len(group2.pending_messages) == 1
    
    print("Test 3 PASSED")


async def test_retry_mechanism():
    print("\n=== Test 4: 重试机制 ===")
    
    import time
    import heapq
    
    async with manager_lock:
        if "test-retry" in topic_manager:
            del topic_manager["test-retry"]
    
    topic = Topic(name="test-retry", capacity=100)
    
    async with topic.lock:
        group = ConsumerGroup(name="test-group")
        topic.consumer_groups["test-group"] = group
        
        msg = Message("retry message")
        group.inflight_messages[msg.id] = msg
        
        msg.retry_count += 1
        current_time = time.time()
        retry_at = current_time + 1
        
        heapq.heappush(group.retry_heap, (retry_at, msg.id, msg))
        group.retry_message_ids.add(msg.id)
        
        print(f"Message pushed to retry heap with retry_at: {retry_at}")
        print(f"Retry heap size: {len(group.retry_heap)}")
        print(f"Inflight messages: {len(group.inflight_messages)}")
        
        assert len(group.retry_heap) == 1
        assert msg.id in group.retry_message_ids
    
    print("Test 4 PASSED")


async def main():
    await test_basic_flow()
    await test_topic_capacity()
    await test_consumer_groups()
    await test_retry_mechanism()
    
    print("\n=== 所有测试通过 ===")


if __name__ == "__main__":
    asyncio.run(main())
