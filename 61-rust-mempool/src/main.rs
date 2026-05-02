use mempool_allocator::MemoryPool;

fn main() {
    let mut pool = MemoryPool::with_size(1024 * 1024);

    println!("Initial state:");
    pool.stats();

    println!("\n=== Testing basic allocation ===");
    let ptr1 = pool.alloc(50).expect("Allocation failed");
    println!("Allocated 50 bytes at ptr: 0x{:x}", ptr1);

    let ptr2 = pool.alloc(100).expect("Allocation failed");
    println!("Allocated 100 bytes at ptr: 0x{:x}", ptr2);

    let ptr3 = pool.alloc(200).expect("Allocation failed");
    println!("Allocated 200 bytes at ptr: 0x{:x}", ptr3);

    let ptr4 = pool.alloc(2000).expect("Allocation failed");
    println!("Allocated 2000 bytes (large allocation) at ptr: 0x{:x}", ptr4);

    println!("\nAfter allocations:");
    pool.stats();

    if let Some(slice) = pool.get_slice_mut(ptr1, 50) {
        for i in 0..50 {
            slice[i] = i as u8;
        }
        println!("\nWrote data to first allocation");
    }

    if let Some(slice) = pool.get_slice(ptr1, 50) {
        println!("First 10 bytes: {:?}", &slice[0..10]);
    }

    println!("\n=== Testing free ===");
    pool.free(ptr2, 100);
    println!("Freed 100 byte allocation");

    pool.free(ptr3, 200);
    println!("Freed 200 byte allocation");

    pool.free(ptr4, 2000);
    println!("Freed large allocation");

    println!("\nAfter freeing some allocations:");
    pool.stats();

    println!("\n=== Testing fragmentation ===");
    let mut pointers = Vec::new();
    for _ in 0..10 {
        let ptr = pool.alloc(64).expect("Allocation failed");
        pointers.push(ptr);
        println!("Allocated 64 bytes at ptr: 0x{:x}", ptr);
    }

    for (i, &ptr) in pointers.iter().enumerate() {
        if i % 2 == 0 {
            pool.free(ptr, 64);
            println!("Freed allocation {}", i);
        }
    }

    println!("\nAfter creating fragmentation:");
    pool.stats();

    for (i, &ptr) in pointers.iter().enumerate() {
        if i % 2 != 0 {
            pool.free(ptr, 64);
        }
    }

    pool.free(ptr1, 50);

    println!("\nFinal state:");
    pool.stats();
}
