import asyncio
import aiohttp
import time
import json

BASE_URL = "http://localhost:8000"


async def test_semaphore_service():
    print("=" * 60)
    print("Test 1: Create semaphore with capacity 3")
    print("=" * 60)
    
    async with aiohttp.ClientSession() as session:
        async with session.post(
            f"{BASE_URL}/semaphores",
            json={"name": "test_semaphore", "capacity": 3}
        ) as resp:
            print(f"Status: {resp.status}")
            data = await resp.json()
            print(f"Response: {json.dumps(data, indent=2)}")
            assert resp.status == 201, "Failed to create semaphore"
            print("✓ Semaphore created successfully\n")

    print("=" * 60)
    print("Test 2: Acquire 3 licenses (should all succeed immediately)")
    print("=" * 60)
    
    acquire_ids = []
    async with aiohttp.ClientSession() as session:
        for i in range(3):
            start_time = time.time()
            async with session.post(
                f"{BASE_URL}/semaphores/test_semaphore/acquire",
                json={}
            ) as resp:
                elapsed = time.time() - start_time
                data = await resp.json()
                print(f"Acquire {i+1}: status={resp.status}, elapsed={elapsed:.3f}s, id={data.get('acquire_id')[:8]}...")
                assert resp.status == 200, f"Acquire {i+1} failed"
                acquire_ids.append(data['acquire_id'])
    
    print("✓ All 3 acquires succeeded\n")

    print("=" * 60)
    print("Test 3: Check status (should have 3 holders)")
    print("=" * 60)
    
    async with aiohttp.ClientSession() as session:
        async with session.get(
            f"{BASE_URL}/semaphores/test_semaphore/status"
        ) as resp:
            data = await resp.json()
            print(json.dumps(data, indent=2))
            assert data['current_holders_count'] == 3
            assert data['capacity'] == 3
    print("✓ Status check passed\n")

    print("=" * 60)
    print("Test 4: Test FIFO queue - acquire 2 more (should wait)")
    print("=" * 60)
    
    queued_results = []
    
    async def queued_acquire(num):
        async with aiohttp.ClientSession() as session:
            start_time = time.time()
            async with session.post(
                f"{BASE_URL}/semaphores/test_semaphore/acquire",
                json={"timeout": 10.0}
            ) as resp:
                elapsed = time.time() - start_time
                data = await resp.json()
                result = {
                    "num": num,
                    "status": resp.status,
                    "elapsed": elapsed,
                    "acquire_id": data.get('acquire_id')
                }
                print(f"Queued acquire {num}: status={resp.status}, elapsed={elapsed:.3f}s")
                queued_results.append(result)
                return result

    start_time = time.time()
    await asyncio.gather(
        queued_acquire(1),
        queued_acquire(2),
    )
    
    print(f"Note: These should be waiting. Let's check queue status...")
    
    async with aiohttp.ClientSession() as session:
        async with session.get(
            f"{BASE_URL}/semaphores/test_semaphore/status"
        ) as resp:
            data = await resp.json()
            print(f"Wait queue length: {data['wait_queue_length']}")
    
    print("=" * 60)
    print("Test 5: Release one holder to wake up queue (FIFO order)")
    print("=" * 60)
    
    async with aiohttp.ClientSession() as session:
        async with session.post(
            f"{BASE_URL}/semaphores/test_semaphore/release",
            json={"acquire_id": acquire_ids[0]}
        ) as resp:
            print(f"Release status: {resp.status}")
            assert resp.status == 204
    
    await asyncio.sleep(1)
    
    print(f"Queued results after release: {len(queued_results)} tasks completed")
    assert len(queued_results) >= 1, "Should have at least one queued acquire completed"
    
    print("✓ FIFO queue wakeup test passed\n")

    print("=" * 60)
    print("Test 6: Test timeout (acquire with 2s timeout when full)")
    print("=" * 60)
    
    async with aiohttp.ClientSession() as session:
        start_time = time.time()
        async with session.post(
            f"{BASE_URL}/semaphores/test_semaphore/acquire",
            json={"timeout": 2.0}
        ) as resp:
            elapsed = time.time() - start_time
            data = await resp.json()
            print(f"Status: {resp.status}")
            print(f"Elapsed: {elapsed:.3f}s")
            print(f"Response: {json.dumps(data, indent=2)}")
            assert resp.status == 408, "Should timeout with 408"
            assert 1.8 < elapsed < 2.5, f"Timeout should be around 2s, got {elapsed}"
    
    print("✓ Timeout test passed\n")

    print("=" * 60)
    print("Test 7: Dynamic capacity increase (from 3 to 5)")
    print("=" * 60)
    
    async with aiohttp.ClientSession() as session:
        async with session.put(
            f"{BASE_URL}/semaphores/test_semaphore/config",
            json={"capacity": 5}
        ) as resp:
            print(f"Update config status: {resp.status}")
            assert resp.status == 204
        
        async with session.get(
            f"{BASE_URL}/semaphores/test_semaphore/status"
        ) as resp:
            data = await resp.json()
            print(f"New capacity: {data['capacity']}")
            assert data['capacity'] == 5
            print(f"Current holders: {data['current_holders_count']}")
            print(f"Wait queue: {data['wait_queue_length']}")
    
    print("✓ Dynamic capacity increase test passed\n")

    print("=" * 60)
    print("Test 8: List all semaphores")
    print("=" * 60)
    
    async with aiohttp.ClientSession() as session:
        async with session.get(f"{BASE_URL}/semaphores") as resp:
            data = await resp.json()
            print(f"Semaphores: {data['semaphores']}")
            assert "test_semaphore" in data['semaphores']
    
    print("✓ List semaphores test passed\n")

    print("=" * 60)
    print("All tests completed successfully!")
    print("=" * 60)


if __name__ == "__main__":
    asyncio.run(test_semaphore_service())
