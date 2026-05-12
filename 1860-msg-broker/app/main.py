import os
import json
import uuid
import asyncio
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException, Request, Query, Header
from fastapi.responses import JSONResponse
from typing import List, Optional

from .store import store
from .models import (
    CreateTopicRequest,
    ProduceMessageRequest,
    AckRequest,
    TopicListResponse,
    TopicInfo,
    GroupsResponse,
    ConsumerProgress,
    DeadLetterResponse,
    ConsumeResponse,
    ProduceResponse
)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

BACKGROUND_TASK_INTERVAL = 1
background_task = None


async def background_timeout_processor():
    while True:
        try:
            store.process_timeouts()
        except Exception as e:
            logger.error(f"Error in background processor: {e}")
        await asyncio.sleep(BACKGROUND_TASK_INTERVAL)


@asynccontextmanager
async def lifespan(app: FastAPI):
    global background_task
    background_task = asyncio.create_task(background_timeout_processor())
    yield
    if background_task:
        background_task.cancel()


app = FastAPI(title="Message Broker", lifespan=lifespan)


@app.post("/topics/{name}")
async def create_topic(name: str, request: CreateTopicRequest):
    try:
        store.create_topic(name, request.capacity)
        return {"message": f"Topic '{name}' created with capacity {request.capacity}"}
    except ValueError as e:
        raise HTTPException(status_code=409, detail=str(e))


@app.get("/topics", response_model=TopicListResponse)
async def list_topics():
    topics = store.list_topics()
    return TopicListResponse(topics=[TopicInfo(**t) for t in topics])


@app.post("/topics/{name}/messages", response_model=ProduceResponse)
async def produce_message(name: str, request: Request):
    topic = store.get_topic(name)
    if not topic:
        raise HTTPException(status_code=404, detail=f"Topic '{name}' does not exist")
    
    try:
        body_bytes = await request.body()
        if not body_bytes:
            raise HTTPException(status_code=400, detail="Empty request body")
        
        body_str = body_bytes.decode('utf-8')
        try:
            data = json.loads(body_str)
        except json.JSONDecodeError as e:
            logger.warning(f"Invalid JSON message received for topic '{name}': {e}")
            raise HTTPException(status_code=400, detail=f"Invalid JSON: {str(e)}")
        
        if "content" not in data:
            raise HTTPException(status_code=400, detail="Missing 'content' field")
        
        msg_id = store.produce(name, data["content"])
        return ProduceResponse(id=msg_id)
    
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error producing message: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")


@app.get("/topics/{name}/consume", response_model=ConsumeResponse)
async def consume_messages(
    name: str,
    group: str,
    limit: int = Query(10, ge=1, le=100),
    x_consumer_id: Optional[str] = Header(None)
):
    topic = store.get_topic(name)
    if not topic:
        raise HTTPException(status_code=404, detail=f"Topic '{name}' does not exist")
    
    consumer_id = x_consumer_id or f"consumer-{uuid.uuid4().hex[:8]}"
    
    try:
        messages = store.consume(name, group, consumer_id, limit)
        return ConsumeResponse(messages=messages)
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        logger.error(f"Error consuming messages: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")


@app.post("/topics/{name}/ack")
async def ack_messages(
    name: str,
    request: AckRequest,
    x_group: Optional[str] = Header(None)
):
    topic = store.get_topic(name)
    if not topic:
        raise HTTPException(status_code=404, detail=f"Topic '{name}' does not exist")
    
    if not request.message_ids:
        return {"acked": 0, "not_found": 0}
    
    try:
        acked, not_found = store.ack(name, x_group or "default", request.message_ids)
        return {"acked": acked, "not_found": not_found}
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        logger.error(f"Error acking messages: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")


@app.get("/topics/{name}/groups", response_model=GroupsResponse)
async def get_group_progress(name: str):
    topic = store.get_topic(name)
    if not topic:
        raise HTTPException(status_code=404, detail=f"Topic '{name}' does not exist")
    
    try:
        progress = store.get_group_progress(name)
        return GroupsResponse(groups=[ConsumerProgress(**p) for p in progress])
    except Exception as e:
        logger.error(f"Error getting group progress: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")


@app.get("/topics/{name}/deadletter", response_model=DeadLetterResponse)
async def get_dead_letters(name: str):
    topic = store.get_topic(name)
    if not topic:
        raise HTTPException(status_code=404, detail=f"Topic '{name}' does not exist")
    
    try:
        dead_letters = store.get_dead_letters(name)
        return DeadLetterResponse(messages=dead_letters)
    except Exception as e:
        logger.error(f"Error getting dead letters: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")


@app.post("/topics/{name}/deadletter/{message_id}/resend")
async def resend_dead_letter(name: str, message_id: int):
    topic = store.get_topic(name)
    if not topic:
        raise HTTPException(status_code=404, detail=f"Topic '{name}' does not exist")
    
    try:
        success = store.resend_dead_letter(name, message_id)
        if success:
            return {"message": f"Message {message_id} resent successfully"}
        else:
            raise HTTPException(status_code=404, detail=f"Dead letter message {message_id} not found")
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error resending dead letter: {e}")
        raise HTTPException(status_code=500, detail="Internal server error")


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("app.main:app", host="0.0.0.0", port=port, reload=True)
