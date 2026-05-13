import asyncio
import time
from semaphore_manager import Semaphore, SemaphoreManager


async def test_basic_acquire_release():
    print("Test 1: Basic acquire/release")
    sem = Semaphore("test", capacity=3)
    
    id1 = await sem.acquire(timeout=1.0)
    id2 = await sem.acquire(timeout=1.0)
    id3 = await sem.acquire(timeout=1.0)
    
    assert id1 and id2 and id3, "First 3 acquires should succeed"
    assert sem.current_holders_count == 3
    assert sem.wait_queue_length == 0
    print(f"  Acquired 3 licenses: {id1[:8]}..., {id2[:8]}..., {id3[:8]}...")
    
    success = await sem.release(id1)
    assert success, "Release should succeed"
    assert sem.current_holders_count == 2
    print("  Released 1, count now 2")
    print("  ✓ Basic test passed\n")


async def test_fifo_queue():
    print("Test 2: FIFO queue order")
    sem = Semaphore("test", capacity=1)
    results = []
    
    async def acquire_and_record(name):
        start = time.time()
        acquire_id = await sem.acquire(timeout=5.0)
        elapsed = time.time() - start
        results.append((name, acquire_id, elapsed))
        await asyncio.sleep(0.1)
        await sem.release(acquire_id)
    
    id0 = await sem.acquire(timeout=1.0)
    assert id0, "First acquire should succeed immediately"
    print("  First acquired, now starting 3 queued acquires...")
    
    await asyncio.gather(
        acquire_and_record("A"),
        acquire_and_record("B"),
        acquire_and_record("C"),
    )
    
    await sem.release(id0)
    
    await asyncio.sleep(0.5)
    
    names = [r[0] for r in results]
    print(f"  Completion order: {names}")
    assert names == ["A", "B", "C"], f"Should be FIFO order A,B,C but got {names}"
    print("  ✓ FIFO queue test passed\n")


async def test_timeout():
    print("Test 3: Timeout mechanism")
    sem = Semaphore("test", capacity=1)
    
    id1 = await sem.acquire(timeout=1.0)
    assert id1, "First acquire should succeed"
    
    start = time.time()
    id2 = await sem.acquire(timeout=1.0)
    elapsed = time.time() - start
    
    assert id2 is None, "Second acquire should timeout"
    assert 0.9 < elapsed < 1.5, f"Timeout should be ~1s, got {elapsed:.3f}s"
    print(f"  Timeout after {elapsed:.3f}s")
    print("  ✓ Timeout test passed\n")


async def test_dynamic_capacity_increase():
    print("Test 4: Dynamic capacity increase wakes up waiters")
    sem = Semaphore("test", capacity=1)
    results = []
    
    async def acquire_and_record(name):
        acquire_id = await sem.acquire(timeout=5.0)
        results.append((name, acquire_id))
        return acquire_id
    
    id1 = await sem.acquire(timeout=1.0)
    assert id1
    
    task2 = asyncio.create_task(acquire_and_record("A"))
    task3 = asyncio.create_task(acquire_and_record("B"))
    
    await asyncio.sleep(0.2)
    assert sem.wait_queue_length == 2, "Should have 2 waiting"
    print(f"  2 waiting, increasing capacity from 1 to 3...")
    
    await sem.update_capacity(3)
    await asyncio.sleep(0.2)
    
    assert sem.wait_queue_length == 0, "Should have 0 waiting after capacity increase"
    assert sem.current_holders_count == 3, "Should have 3 holders now"
    print(f"  Now have {sem.current_holders_count} holders")
    
    await task2
    await task3
    print("  ✓ Dynamic capacity test passed\n")


async def test_dynamic_capacity_decrease():
    print("Test 5: Dynamic capacity decrease")
    sem = Semaphore("test", capacity=3)
    
    id1 = await sem.acquire(timeout=1.0)
    id2 = await sem.acquire(timeout=1.0)
    id3 = await sem.acquire(timeout=1.0)
    assert sem.current_holders_count == 3
    
    print("  Acquired 3, decreasing capacity to 1...")
    await sem.update_capacity(1)
    
    assert sem.current_holders_count == 3, "Existing holders should not be affected"
    assert sem.capacity == 1
    
    id4 = await sem.acquire(timeout=0.5)
    assert id4 is None, "New acquire should fail (capacity 1 but 3 holders)"
    
    print("  Releasing 2 holders (count=3->1)...")
    await sem.release(id2)
    await sem.release(id3)
    assert sem.current_holders_count == 1
    
    id5 = await sem.acquire(timeout=0.5)
    assert id5 is None, "Still cannot acquire (count=1 == capacity=1)"
    
    print("  Releasing last holder (count=1->0)...")
    await sem.release(id1)
    assert sem.current_holders_count == 0
    
    id6 = await sem.acquire(timeout=1.0)
    assert id6, "Now should be able to acquire"
    print("  ✓ Capacity decrease test passed\n")


async def test_manager():
    print("Test 6: SemaphoreManager")
    manager = SemaphoreManager()
    
    created = await manager.create_semaphore("test1", 5)
    assert created, "Should create new semaphore"
    
    created2 = await manager.create_semaphore("test1", 5)
    assert not created2, "Should not create duplicate"
    
    assert manager.get_semaphore("test1") is not None
    assert manager.get_semaphore("nonexistent") is None
    
    semaphores = manager.list_all()
    assert "test1" in semaphores
    
    deleted = await manager.delete_semaphore("test1")
    assert deleted
    
    assert manager.get_semaphore("test1") is None
    print("  ✓ SemaphoreManager test passed\n")


async def main():
    print("=" * 60)
    print("Core Semaphore Logic Tests")
    print("=" * 60 + "\n")
    
    await test_basic_acquire_release()
    await test_fifo_queue()
    await test_timeout()
    await test_dynamic_capacity_increase()
    await test_dynamic_capacity_decrease()
    await test_manager()
    
    print("=" * 60)
    print("All core tests passed!")
    print("=" * 60)


if __name__ == "__main__":
    asyncio.run(main())
