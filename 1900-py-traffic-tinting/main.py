import os
import random
import asyncio
import uuid
from typing import Dict, List, Optional, Any
from enum import Enum

from fastapi import FastAPI, HTTPException, Request, Response
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field
import httpx


class MatchType(str, Enum):
    HEADER = "header"
    QUERY = "query"
    USER_ID = "user_id"


class RuleCreate(BaseModel):
    match_type: MatchType
    key: Optional[str] = None
    value: Optional[str] = None
    user_ids: Optional[List[str]] = None
    tag: str


class Rule(RuleCreate):
    id: str


class BackendCreate(BaseModel):
    address: str
    tags: List[str] = Field(default_factory=list)


class Backend(BaseModel):
    id: str
    address: str
    tags: List[str]


class Storage:
    def __init__(self):
        self.rules: List[Rule] = []
        self.backends: Dict[str, Backend] = {}
        self.tag_counter: Dict[str, int] = {}
        self._lock = asyncio.Lock()

    async def add_rule(self, rule_create: RuleCreate) -> Rule:
        async with self._lock:
            rule = Rule(id=str(uuid.uuid4()), **rule_create.model_dump())
            self.rules.append(rule)
            return rule

    async def update_rule(self, rule_id: str, rule_create: RuleCreate) -> Rule:
        async with self._lock:
            for i, rule in enumerate(self.rules):
                if rule.id == rule_id:
                    self.rules[i] = Rule(id=rule_id, **rule_create.model_dump())
                    return self.rules[i]
            raise HTTPException(status_code=404, detail="Rule not found")

    async def delete_rule(self, rule_id: str) -> None:
        async with self._lock:
            for i, rule in enumerate(self.rules):
                if rule.id == rule_id:
                    self.rules.pop(i)
                    return
            raise HTTPException(status_code=404, detail="Rule not found")

    async def get_rules(self) -> List[Rule]:
        async with self._lock:
            return list(self.rules)

    async def add_backend(self, backend_create: BackendCreate) -> Backend:
        async with self._lock:
            backend = Backend(id=str(uuid.uuid4()), **backend_create.model_dump())
            self.backends[backend.id] = backend
            tags = backend.tags if backend.tags else [""]
            for tag in tags:
                if tag not in self.tag_counter:
                    self.tag_counter[tag] = 0
            return backend

    async def delete_backend(self, backend_id: str) -> None:
        async with self._lock:
            if backend_id not in self.backends:
                raise HTTPException(status_code=404, detail="Backend not found")
            del self.backends[backend_id]

    async def get_backends(self) -> List[Backend]:
        async with self._lock:
            return list(self.backends.values())

    async def get_backend_by_tag(self, tag: str) -> Optional[Backend]:
        async with self._lock:
            candidates = []
            for backend in self.backends.values():
                backend_tags = backend.tags if backend.tags else [""]
                if tag in backend_tags:
                    candidates.append(backend)
            if candidates:
                return random.choice(candidates)
            return None

    async def increment_counter(self, tag: str) -> None:
        async with self._lock:
            self.tag_counter[tag] = self.tag_counter.get(tag, 0) + 1

    async def get_counter(self, tag: str) -> int:
        async with self._lock:
            return self.tag_counter.get(tag, 0)

    async def get_all_counters(self) -> Dict[str, int]:
        async with self._lock:
            return dict(self.tag_counter)


app = FastAPI(title="Traffic Tinting Service")
storage = Storage()


def match_rule(rule: Rule, headers: Dict[str, str], query_params: Dict[str, str], user_id: Optional[str]) -> bool:
    if rule.match_type == MatchType.HEADER:
        if rule.key and rule.value:
            return headers.get(rule.key.lower()) == rule.value
        return False
    elif rule.match_type == MatchType.QUERY:
        if rule.key and rule.value:
            return query_params.get(rule.key) == rule.value
        return False
    elif rule.match_type == MatchType.USER_ID:
        if rule.user_ids and user_id:
            return user_id in rule.user_ids
        return False
    return False


async def determine_tag(headers: Dict[str, str], query_params: Dict[str, str]) -> str:
    user_id = query_params.get("user_id")
    rules = await storage.get_rules()
    
    header_rules = [r for r in rules if r.match_type == MatchType.HEADER]
    for rule in header_rules:
        if match_rule(rule, headers, query_params, user_id):
            return rule.tag
    
    query_rules = [r for r in rules if r.match_type == MatchType.QUERY]
    for rule in query_rules:
        if match_rule(rule, headers, query_params, user_id):
            return rule.tag
    
    user_rules = [r for r in rules if r.match_type == MatchType.USER_ID]
    for rule in user_rules:
        if match_rule(rule, headers, query_params, user_id):
            return rule.tag
    
    return ""


@app.post("/rules")
async def create_rule(rule_create: RuleCreate) -> Rule:
    if rule_create.match_type == MatchType.USER_ID:
        if not rule_create.user_ids:
            raise HTTPException(status_code=400, detail="user_ids required for user_id rule")
    else:
        if not rule_create.key or not rule_create.value:
            raise HTTPException(status_code=400, detail="key and value required for header/query rule")
    
    return await storage.add_rule(rule_create)


@app.put("/rules/{rule_id}")
async def update_rule(rule_id: str, rule_create: RuleCreate) -> Rule:
    if rule_create.match_type == MatchType.USER_ID:
        if not rule_create.user_ids:
            raise HTTPException(status_code=400, detail="user_ids required for user_id rule")
    else:
        if not rule_create.key or not rule_create.value:
            raise HTTPException(status_code=400, detail="key and value required for header/query rule")
    
    return await storage.update_rule(rule_id, rule_create)


@app.delete("/rules/{rule_id}")
async def delete_rule(rule_id: str) -> Dict[str, str]:
    await storage.delete_rule(rule_id)
    return {"message": "Rule deleted"}


@app.post("/backends")
async def register_backend(backend_create: BackendCreate) -> Backend:
    return await storage.add_backend(backend_create)


@app.get("/backends")
async def list_backends() -> Dict[str, Any]:
    backends = await storage.get_backends()
    groups: Dict[str, List[Backend]] = {}
    for backend in backends:
        tags = backend.tags if backend.tags else [""]
        for tag in tags:
            if tag not in groups:
                groups[tag] = []
            groups[tag].append(backend)
    return {"groups": groups}


@app.get("/status")
async def get_status() -> Dict[str, Any]:
    rules = await storage.get_rules()
    backends = await storage.get_backends()
    counters = await storage.get_all_counters()
    
    groups: Dict[str, List[str]] = {}
    for backend in backends:
        tags = backend.tags if backend.tags else [""]
        for tag in tags:
            if tag not in groups:
                groups[tag] = []
            groups[tag].append(backend.address)
    
    return {
        "rules": rules,
        "backend_groups": groups,
        "traffic_counters": counters
    }


@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"])
async def proxy(request: Request, path: str):
    headers = {k.lower(): v for k, v in request.headers.items()}
    query_params = dict(request.query_params)
    
    tag = await determine_tag(headers, query_params)
    backend = await storage.get_backend_by_tag(tag)
    
    if not backend:
        tag = ""
        backend = await storage.get_backend_by_tag(tag)
    
    if not backend:
        raise HTTPException(status_code=503, detail="No backend available")
    
    await storage.increment_counter(tag)
    
    method = request.method
    url = f"{backend.address.rstrip('/')}/{path.lstrip('/')}"
    
    client_headers = {k: v for k, v in request.headers.items() if k.lower() not in ("host", "content-length")}
    body = await request.body()
    
    async with httpx.AsyncClient(timeout=httpx.Timeout(30.0, connect=10.0)) as client:
        req = client.build_request(
            method,
            url,
            params=request.query_params,
            headers=client_headers,
            content=body,
        )
        
        resp = await client.send(req, stream=True)
        
        response_headers = {k: v for k, v in resp.headers.items() if k.lower() not in ("content-length", "transfer-encoding")}
        response_headers["X-Traffic-Tag"] = tag if tag else "default"
        
        return StreamingResponse(
            resp.aiter_bytes(),
            status_code=resp.status_code,
            headers=response_headers
        )


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run(app, host="0.0.0.0", port=port)
