import os
import time
import uuid
from typing import List

import aiosqlite
from fastapi import FastAPI, Query
from pydantic import BaseModel

DB_PATH = os.getenv("SQLITE_DB", "mq.db")
PORT = int(os.getenv("PORT", 8000))
ACK_TIMEOUT = 30

app = FastAPI()


class MessageRequest(BaseModel):
    content: str


class AckRequest(BaseModel):
    message_ids: List[str]


async def init_db():
    async with aiosqlite.connect(DB_PATH) as db:
        await db.execute(
            """
            CREATE TABLE IF NOT EXISTS topics (
                name TEXT PRIMARY KEY,
                created_at REAL
            )
            """
        )
        await db.execute(
            """
            CREATE TABLE IF NOT EXISTS messages (
                id TEXT PRIMARY KEY,
                topic TEXT,
                content TEXT,
                created_at REAL,
                seq INTEGER NOT NULL,
                FOREIGN KEY (topic) REFERENCES topics(name),
                UNIQUE(topic, seq)
            )
            """
        )
        await db.execute(
            """
            CREATE TABLE IF NOT EXISTS consumer_groups (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                topic TEXT,
                name TEXT,
                last_acked_seq INTEGER DEFAULT 0,
                UNIQUE(topic, name),
                FOREIGN KEY (topic) REFERENCES topics(name)
            )
            """
        )
        await db.execute(
            """
            CREATE TABLE IF NOT EXISTS pending_messages (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                topic TEXT,
                message_id TEXT,
                group_name TEXT,
                seq INTEGER,
                acquired_at REAL,
                UNIQUE(topic, message_id, group_name),
                FOREIGN KEY (topic) REFERENCES topics(name),
                FOREIGN KEY (message_id) REFERENCES messages(id)
            )
            """
        )
        await db.commit()


async def get_or_create_topic(db, name: str):
    cursor = await db.execute("SELECT name FROM topics WHERE name = ?", (name,))
    row = await cursor.fetchone()
    if not row:
        await db.execute(
            "INSERT INTO topics (name, created_at) VALUES (?, ?)",
            (name, time.time())
        )
    return name


async def get_or_create_group(db, topic: str, group_name: str):
    cursor = await db.execute(
        "SELECT id, last_acked_seq FROM consumer_groups WHERE topic = ? AND name = ?",
        (topic, group_name)
    )
    row = await cursor.fetchone()
    if not row:
        await db.execute(
            "INSERT INTO consumer_groups (topic, name, last_acked_seq) VALUES (?, ?, 0)",
            (topic, group_name)
        )
    return group_name


@app.on_event("startup")
async def startup():
    await init_db()


@app.post("/topics/{name}/messages")
async def publish_message(name: str, request: MessageRequest):
    async with aiosqlite.connect(DB_PATH) as db:
        await get_or_create_topic(db, name)

        cursor = await db.execute(
            "SELECT COALESCE(MAX(seq), 0) FROM messages WHERE topic = ?",
            (name,)
        )
        row = await cursor.fetchone()
        next_seq = (row[0] or 0) + 1

        message_id = str(uuid.uuid4())
        await db.execute(
            "INSERT INTO messages (id, topic, content, created_at, seq) VALUES (?, ?, ?, ?, ?)",
            (message_id, name, request.content, time.time(), next_seq)
        )
        await db.commit()
        return {"message_id": message_id, "status": "published"}


@app.get("/topics/{name}/consume")
async def consume_messages(
    name: str,
    group: str = Query(...),
    limit: int = Query(1, ge=1, le=100)
):
    async with aiosqlite.connect(DB_PATH) as db:
        await get_or_create_topic(db, name)
        await get_or_create_group(db, name, group)

        cursor = await db.execute(
            "SELECT last_acked_seq FROM consumer_groups WHERE topic = ? AND name = ?",
            (name, group)
        )
        row = await cursor.fetchone()
        last_acked_seq = row[0] if row else 0

        now = time.time()
        await db.execute(
            """
            DELETE FROM pending_messages
            WHERE topic = ? AND group_name = ? AND acquired_at <= ?
            """,
            (name, group, now - ACK_TIMEOUT)
        )

        cursor = await db.execute(
            """
            SELECT m.id, m.content, m.created_at, m.seq
            FROM messages m
            WHERE m.topic = ?
              AND m.seq > ?
              AND m.id NOT IN (
                  SELECT message_id FROM pending_messages
                  WHERE topic = ? AND group_name = ?
              )
            ORDER BY m.seq ASC
            LIMIT ?
            """,
            (name, last_acked_seq, name, group, limit)
        )
        rows = await cursor.fetchall()
        messages = []
        for row in rows:
            message_id, content, created_at, seq = row
            await db.execute(
                """
                INSERT OR REPLACE INTO pending_messages
                (topic, message_id, group_name, seq, acquired_at)
                VALUES (?, ?, ?, ?, ?)
                """,
                (name, message_id, group, seq, now)
            )
            messages.append({
                "id": message_id,
                "content": content,
                "created_at": created_at
            })
        await db.commit()
        return {"messages": messages}


@app.post("/topics/{name}/ack")
async def acknowledge_messages(name: str, request: AckRequest):
    if not request.message_ids:
        return {"status": "acknowledged", "count": 0}

    async with aiosqlite.connect(DB_PATH) as db:
        await get_or_create_topic(db, name)

        placeholders = ",".join("?" for _ in request.message_ids)
        cursor = await db.execute(
            f"""
            SELECT message_id, group_name, seq FROM pending_messages
            WHERE topic = ? AND message_id IN ({placeholders})
            """,
            (name, *request.message_ids)
        )
        valid_rows = await cursor.fetchall()
        valid_ids = [row[0] for row in valid_rows]

        if valid_ids:
            group_max_seqs = {}
            for row in valid_rows:
                _, group_name, seq = row
                if group_name not in group_max_seqs or seq > group_max_seqs[group_name]:
                    group_max_seqs[group_name] = seq

            valid_placeholders = ",".join("?" for _ in valid_ids)
            await db.execute(
                f"""
                DELETE FROM pending_messages
                WHERE topic = ? AND message_id IN ({valid_placeholders})
                """,
                (name, *valid_ids)
            )

            for group_name, max_seq in group_max_seqs.items():
                cursor = await db.execute(
                    "SELECT last_acked_seq FROM consumer_groups WHERE topic = ? AND name = ?",
                    (name, group_name)
                )
                current_row = await cursor.fetchone()
                current_max = current_row[0] if current_row else 0
                if max_seq > current_max:
                    await db.execute(
                        """
                        UPDATE consumer_groups
                        SET last_acked_seq = ?
                        WHERE topic = ? AND name = ?
                        """,
                        (max_seq, name, group_name)
                    )

        await db.commit()
        return {"status": "acknowledged", "count": len(valid_ids)}


@app.get("/topics")
async def list_topics():
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT name, created_at FROM topics ORDER BY created_at ASC")
        rows = await cursor.fetchall()
        topics = []
        for row in rows:
            name, created_at = row

            cursor = await db.execute(
                "SELECT COALESCE(MAX(seq), 0) FROM messages WHERE topic = ?",
                (name,)
            )
            max_seq = (await cursor.fetchone())[0]

            cursor = await db.execute(
                """
                SELECT COUNT(*) FROM pending_messages WHERE topic = ?
                """,
                (name,)
            )
            pending = (await cursor.fetchone())[0]

            cursor = await db.execute(
                """
                SELECT name, last_acked_seq
                FROM consumer_groups
                WHERE topic = ?
                """,
                (name,)
            )
            groups = []
            for g_row in await cursor.fetchall():
                g_name, last_seq = g_row
                lag = max_seq - last_seq
                groups.append({
                    "group": g_name,
                    "last_acked_seq": last_seq,
                    "lag": lag
                })

            topics.append({
                "name": name,
                "created_at": created_at,
                "total_messages": max_seq,
                "pending_messages": pending,
                "consumer_groups": groups
            })

        return {"topics": topics}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=PORT)
