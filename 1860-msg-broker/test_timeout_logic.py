import time
import threading
import sys
sys.path.insert(0, '/home/baru/Work/Private/work01/SoloCoder/1860-msg-broker')

from app.snowflake import generate_id

ACK_TIMEOUT_SECONDS = 2
MAX_RETRY = 3
RETRY_INTERVALS = [1, 2, 4]


class SimpleGroupState:
    def __init__(self, name):
        self.name = name
        self.last_ack_id = 0
        self.total_consumed = 0
        self.in_flight = {}
        self.group_retry_state = {}


class SimpleTopic:
    def __init__(self, name, capacity):
        self.name = name
        self.capacity = capacity
        self.messages = []
        self.dead_letters = []
        self.groups = {}
        self.lock = threading.RLock()
    
    def get_group(self, group):
        if group not in self.groups:
            self.groups[group] = SimpleGroupState(group)
        return self.groups[group]


class SimpleStore:
    def __init__(self):
        self.topics = {}
    
    def create_topic(self, name, capacity):
        if name in self.topics:
            raise ValueError(f"Topic '{name}' already exists")
        self.topics[name] = SimpleTopic(name, capacity)
        return self.topics[name]
    
    def produce(self, topic_name, content):
        topic = self.topics[topic_name]
        with topic.lock:
            msg_id = generate_id()
            message = {
                "id": msg_id,
                "topic": topic_name,
                "content": content,
                "created_at": time.time()
            }
            topic.messages.append(message)
            return msg_id
    
    def consume(self, topic_name, group, limit):
        topic = self.topics[topic_name]
        with topic.lock:
            group_state = topic.get_group(group)
            consumed = []
            now = time.time()
            in_flight_ids = set(group_state.in_flight.keys())
            
            for msg in topic.messages:
                if len(consumed) >= limit:
                    break
                
                msg_id = msg["id"]
                if msg_id in in_flight_ids:
                    continue
                
                retry_state = group_state.group_retry_state.get(msg_id)
                if retry_state and retry_state["next_retry_at"] and now < retry_state["next_retry_at"]:
                    continue
                
                consumed.append(msg)
                
                inflight_record = {
                    "message_id": msg_id,
                    "group": group,
                    "delivered_at": now,
                    "expires_at": now + ACK_TIMEOUT_SECONDS
                }
                group_state.in_flight[msg_id] = inflight_record
            
            return consumed
    
    def nack(self, topic_name, group, message_id, reason="nack"):
        topic = self.topics[topic_name]
        with topic.lock:
            group_state = topic.get_group(group)
            
            if message_id not in group_state.in_flight:
                return
            
            del group_state.in_flight[message_id]
            
            msg = None
            for m in topic.messages:
                if m["id"] == message_id:
                    msg = m
                    break
            
            if not msg:
                return
            
            if message_id not in group_state.group_retry_state:
                group_state.group_retry_state[message_id] = {
                    "retry_count": 0,
                    "next_retry_at": None
                }
            
            retry_state = group_state.group_retry_state[message_id]
            retry_state["retry_count"] += 1
            
            if retry_state["retry_count"] >= MAX_RETRY:
                dead_letter = {
                    "id": msg["id"],
                    "topic": topic_name,
                    "content": msg["content"],
                    "created_at": msg["created_at"],
                    "retry_count": retry_state["retry_count"],
                    "moved_to_deadletter_at": time.time(),
                    "reason": reason
                }
                topic.dead_letters.append(dead_letter)
                
                for other_group in topic.groups.values():
                    if message_id in other_group.in_flight:
                        del other_group.in_flight[message_id]
                    if message_id in other_group.group_retry_state:
                        del other_group.group_retry_state[message_id]
                
                for i, m in enumerate(topic.messages):
                    if m["id"] == message_id:
                        topic.messages.pop(i)
                        break
                
                print(f"   [DEAD LETTER] 消息 {message_id} 经过 {MAX_RETRY} 次重试后进入死信队列，原因: {reason}")
            else:
                retry_interval = RETRY_INTERVALS[min(retry_state["retry_count"] - 1, len(RETRY_INTERVALS) - 1)]
                retry_state["next_retry_at"] = time.time() + retry_interval
                print(f"   [NACK] 消息 {message_id} 第 {retry_state['retry_count']} 次失败，{retry_interval}秒后重试")
    
    def process_timeouts(self):
        for topic in list(self.topics.values()):
            with topic.lock:
                now = time.time()
                for group_name, group_state in list(topic.groups.items()):
                    expired = []
                    for msg_id, inflight in list(group_state.in_flight.items()):
                        if now >= inflight["expires_at"]:
                            expired.append(msg_id)
                    
                    for msg_id in expired:
                        self.nack(topic.name, group_name, msg_id, reason="ack_timeout")


def test_timeout_logic():
    print("=" * 60)
    print("测试 ACK 超时重投递逻辑 (超时时间 2 秒)")
    print("=" * 60)
    
    store = SimpleStore()
    store.create_topic("test-topic", 1000)
    
    print("\n1. 发送 1 条消息")
    msg_id = store.produce("test-topic", {"data": "hello"})
    print(f"   消息 ID: {msg_id}")
    
    print("\n2. Group-1 消费这条消息")
    messages = store.consume("test-topic", "group-1", 10)
    print(f"   消费到 {len(messages)} 条消息")
    
    group_state = store.topics["test-topic"].get_group("group-1")
    print(f"   in_flight: {list(group_state.in_flight.keys())}")
    print(f"   消息剩余: {len(store.topics['test-topic'].messages)} 条")
    
    print("\n3. 不 ACK，等待超时...")
    for i in range(5):
        time.sleep(1)
        store.process_timeouts()
        
        group_state = store.topics["test-topic"].get_group("group-1")
        in_flight_count = len(group_state.in_flight)
        dead_count = len(store.topics["test-topic"].dead_letters)
        msg_count = len(store.topics["test-topic"].messages)
        
        if in_flight_count == 0 and msg_count > 0:
            retry_state = group_state.group_retry_state.get(msg_id)
            if retry_state:
                next_retry = retry_state.get('next_retry_at')
                if next_retry:
                    wait_time = max(0, next_retry - time.time())
                    print(f"   第 {i+1} 秒: in_flight={in_flight_count}, 重试次数={retry_state['retry_count']}, 下次重试在 {wait_time:.1f}秒后")
                else:
                    print(f"   第 {i+1} 秒: in_flight={in_flight_count}, 消息已释放")
            else:
                print(f"   第 {i+1} 秒: in_flight={in_flight_count}")
        elif dead_count > 0:
            print(f"   第 {i+1} 秒: 已进入死信队列!")
            break
        else:
            print(f"   第 {i+1} 秒: in_flight={in_flight_count}, 等待超时...")
    
    print("\n4. 验证消息是否重新可消费 (或已进入死信)")
    dead = store.topics["test-topic"].dead_letters
    if dead:
        print(f"   死信队列: {len(dead)} 条消息")
        print(f"   死信内容: {dead[0]}")
    else:
        messages = store.consume("test-topic", "group-1", 10)
        print(f"   重新消费到 {len(messages)} 条消息")
        if messages:
            retry_state = group_state.group_retry_state.get(messages[0]["id"])
            if retry_state:
                print(f"   当前重试次数: {retry_state['retry_count']}")
    
    print("\n5. 验证 Group-2 独立消费进度")
    messages = store.consume("test-topic", "group-2", 10)
    print(f"   Group-2 消费到 {len(messages)} 条消息")
    
    print("\n" + "=" * 60)
    print("测试完成")
    print("=" * 60)
    print("\nACK 超时逻辑说明:")
    print("1. consume() 时设置 expires_at = now + ACK_TIMEOUT_SECONDS")
    print("2. process_timeouts() 定期检查，发现超时调用 nack()")
    print("3. nack() 增加 retry_count，设置 next_retry_at")
    print("4. 重试 3 次后进入死信队列")
    print("5. 每个 Group 的重试状态独立隔离")


if __name__ == "__main__":
    test_timeout_logic()
