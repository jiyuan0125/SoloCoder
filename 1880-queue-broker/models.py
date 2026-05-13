import asyncio
from typing import List, Dict, Optional, Any
from dataclasses import dataclass, field


@dataclass
class Message:
    id: int
    content: str
    priority: int = 0
    retry_count: int = 0
    created_at: float = 0
    last_retry_at: float = 0


@dataclass
class ConsumerGroup:
    id: str
    offset: int = 0
    filter_expr: Optional[str] = None
    messages: List[Message] = field(default_factory=list)
    consumed: set = field(default_factory=set)


@dataclass
class Topic:
    name: str
    max_size: int
    creator: str
    messages: List[Message] = field(default_factory=list)
    dead_letter: List[Message] = field(default_factory=list)
    groups: Dict[str, ConsumerGroup] = field(default_factory=dict)
    next_message_id: int = 1
    lock: asyncio.Lock = field(default_factory=asyncio.Lock)
