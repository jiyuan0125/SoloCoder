#!/usr/bin/env python3
import asyncio
import httpx
import time
import json

PROXY_URL = "http://127.0.0.1:8000"
BACKEND_URL = "http://127.0.0.1:8081"


async def test_basic_cache():
    print("\n=== 测试 1: 基础缓存功能 ===")
    async with httpx.AsyncClient() as client:
        r1 = await client.get(f"{PROXY_URL}/api/users")
        print(f"  第1次请求 - X-Cache: {r1.headers.get('X-Cache')}")
        assert r1.headers.get('X-Cache') == 'MISS', "第一次请求应该未命中"
        
        r2 = await client.get(f"{PROXY_URL}/api/users")
        print(f"  第2次请求 - X-Cache: {r2.headers.get('X-Cache')}")
        assert r2.headers.get('X-Cache') == 'HIT', "第二次请求应该命中"
        
        assert r1.json()['request_count'] == r2.json()['request_count'], "缓存命中时request_count应该相同"
        print("  ✓ 基础缓存功能测试通过")


async def test_query_params_cache():
    print("\n=== 测试 2: 查询参数缓存 ===")
    async with httpx.AsyncClient() as client:
        r1 = await client.get(f"{PROXY_URL}/api/products", params={"category": "electronics"})
        print(f"  category=electronics 第1次 - X-Cache: {r1.headers.get('X-Cache')}")
        assert r1.headers.get('X-Cache') == 'MISS'
        
        r2 = await client.get(f"{PROXY_URL}/api/products", params={"category": "electronics"})
        print(f"  category=electronics 第2次 - X-Cache: {r2.headers.get('X-Cache')}")
        assert r2.headers.get('X-Cache') == 'HIT'
        
        r3 = await client.get(f"{PROXY_URL}/api/products", params={"category": "books"})
        print(f"  category=books 第1次 - X-Cache: {r3.headers.get('X-Cache')}")
        assert r3.headers.get('X-Cache') == 'MISS', "不同参数应该是不同的缓存key"
        
        print("  ✓ 查询参数缓存测试通过")


async def test_write_strategy_invalidation():
    print("\n=== 测试 3: 失效模式写策略 ===")
    async with httpx.AsyncClient() as client:
        r1 = await client.get(f"{PROXY_URL}/api/users/1")
        print(f"  GET /api/users/1 第1次 - X-Cache: {r1.headers.get('X-Cache')}")
        
        r2 = await client.post(
            f"{PROXY_URL}/api/users",
            json={"name": "New User"},
            headers={"Content-Type": "application/json"}
        )
        print(f"  POST /api/users - X-Write-Strategy: {r2.headers.get('X-Write-Strategy')}")
        assert r2.headers.get('X-Write-Strategy') == 'invalidation'
        
        r3 = await client.get(f"{PROXY_URL}/api/users")
        print(f"  POST后 GET /api/users - X-Cache: {r3.headers.get('X-Cache')}")
        
        print("  ✓ 失效模式写策略测试通过")


async def test_admin_endpoints():
    print("\n=== 测试 4: 管理接口 ===")
    async with httpx.AsyncClient() as client:
        health = await client.get(f"{PROXY_URL}/_admin/health")
        print(f"  /_admin/health - status: {health.json()['status']}")
        assert health.status_code == 200
        
        stats = await client.get(f"{PROXY_URL}/_admin/stats")
        stats_data = stats.json()
        print(f"  /_admin/stats - 总命中率: {stats_data['total']['hit_rate']}%")
        assert stats.status_code == 200
        
        cache_stats = await client.get(f"{PROXY_URL}/_admin/cache-stats")
        print(f"  /_admin/cache-stats - 命名空间: {list(cache_stats.json().keys())}")
        assert cache_stats.status_code == 200
        
        print("  ✓ 管理接口测试通过")


async def test_concurrent_breakdown_protection():
    print("\n=== 测试 5: 缓存击穿保护（并发请求） ===")
    async with httpx.AsyncClient() as client:
        await client.post(f"{PROXY_URL}/_admin/cache/clear")
        await client.post(f"{PROXY_URL}/_admin/stats/reset")
        
        async def make_request():
            r = await client.get(f"{PROXY_URL}/api/users/999")
            return r.headers.get('X-Cache'), r.status_code
        
        tasks = [make_request() for _ in range(5)]
        results = await asyncio.gather(*tasks)
        
        print(f"  并发请求结果: {results}")
        
        stats = await client.get(f"{PROXY_URL}/_admin/stats")
        stats_data = stats.json()
        print(f"  击穿保护拦截数: {stats_data['total']['breakdown_protected']}")
        
        print("  ✓ 缓存击穿保护测试通过")


async def main():
    print("开始测试 FastAPI 旁路缓存代理...")
    print("=" * 60)
    
    await test_basic_cache()
    await test_query_params_cache()
    await test_write_strategy_invalidation()
    await test_admin_endpoints()
    await test_concurrent_breakdown_protection()
    
    print("\n" + "=" * 60)
    print("所有测试完成！")
    
    print("\n=== 最终统计信息 ===")
    async with httpx.AsyncClient() as client:
        stats = await client.get(f"{PROXY_URL}/_admin/stats")
        print(json.dumps(stats.json(), indent=2, ensure_ascii=False))


if __name__ == "__main__":
    asyncio.run(main())
