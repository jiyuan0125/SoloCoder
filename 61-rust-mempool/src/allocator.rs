use crate::debug_check::{CanaryError, DebugAllocator};
use crate::slab::{SlabAllocator, SlabSize};
use std::collections::HashMap;

const MIN_POOL_SIZE: usize = 1024 * 1024;

#[derive(Debug, Clone)]
pub struct MemoryPool {
    pool: Vec<u8>,
    slab_allocators: [Option<SlabAllocator>; 5],
    large_allocations: HashMap<usize, Vec<u8>>,
    debug_allocator: DebugAllocator,
    allocation_sizes: HashMap<usize, (usize, SlabSize)>,
    next_large_id: usize,
}

impl Default for MemoryPool {
    fn default() -> Self {
        Self::new()
    }
}

impl MemoryPool {
    pub fn new() -> Self {
        Self::with_size(MIN_POOL_SIZE)
    }

    pub fn with_size(total_size: usize) -> Self {
        let actual_size = std::cmp::max(total_size, MIN_POOL_SIZE);
        let per_slab_size = actual_size / 5;

        let mut pool = Vec::with_capacity(actual_size);
        pool.resize(actual_size, 0);

        let mut slab_allocators: [Option<SlabAllocator>; 5] = [None, None, None, None, None];
        let mut current_offset = 0;

        for i in 0..5 {
            let slab_size_enum = SlabSize::from_index(i).unwrap();
            let slab_size = slab_size_enum.to_size();
            let num_slabs = per_slab_size / slab_size;

            if num_slabs > 0 {
                let allocator = SlabAllocator::new(slab_size_enum, current_offset, num_slabs);
                slab_allocators[i] = Some(allocator);
                current_offset += num_slabs * slab_size;
            }
        }

        MemoryPool {
            pool,
            slab_allocators,
            large_allocations: HashMap::new(),
            debug_allocator: DebugAllocator::new(),
            allocation_sizes: HashMap::new(),
            next_large_id: 1,
        }
    }

    pub fn pool_start(&self) -> usize {
        self.pool.as_ptr() as usize
    }

    pub fn pool_end(&self) -> usize {
        self.pool.as_ptr() as usize + self.pool.len()
    }

    pub fn is_pool_address(&self, addr: usize) -> bool {
        addr >= self.pool_start() && addr < self.pool_end()
    }

    pub fn offset_from_ptr(&self, ptr: usize) -> Option<usize> {
        if !self.is_pool_address(ptr) {
            None
        } else {
            Some(ptr - self.pool_start())
        }
    }

    pub fn ptr_from_offset(&self, offset: usize) -> Option<usize> {
        if offset >= self.pool.len() {
            None
        } else {
            Some(self.pool_start() + offset)
        }
    }

    pub fn alloc(&mut self, size: usize) -> Option<usize> {
        if size == 0 {
            return None;
        }

        if size > 1024 {
            return self.alloc_large(size);
        }

        let slab_size_enum = SlabSize::from_size(size)?;

        #[cfg(feature = "debug")]
        let actual_needed = self.debug_allocator.actual_size(size);
        #[cfg(not(feature = "debug"))]
        let actual_needed = size;

        if actual_needed > slab_size_enum.to_size() {
            let next_index = slab_size_enum.to_index() + 1;
            if next_index < 5 {
                if let Some(next_slab) = SlabSize::from_index(next_index) {
                    return self.alloc_with_slab(size, next_slab);
                }
            }
            return self.alloc_large(size);
        }

        self.alloc_with_slab(size, slab_size_enum)
    }

    fn alloc_with_slab(&mut self, size: usize, slab_size: SlabSize) -> Option<usize> {
        let index = slab_size.to_index();

        if let Some(ref mut allocator) = &mut self.slab_allocators[index] {
            if let Some(offset) = allocator.alloc() {
                #[cfg(feature = "debug")]
                {
                    let user_offset = self.debug_allocator.user_offset();
                    let slab_size_val = slab_size.to_size();

                    if offset + slab_size_val <= self.pool.len() {
                        let slab_slice = &mut self.pool[offset..offset + slab_size_val];
                        self.debug_allocator.write_canaries(slab_slice, size);
                    }
                }

                #[cfg(feature = "debug")]
                let user_offset = offset + self.debug_allocator.user_offset();
                #[cfg(not(feature = "debug"))]
                let user_offset = offset;

                self.allocation_sizes
                    .insert(user_offset, (size, slab_size));

                let ptr = self.ptr_from_offset(user_offset)?;
                return Some(ptr);
            }
        }

        self.alloc_large(size)
    }

    fn alloc_large(&mut self, size: usize) -> Option<usize> {
        let id = self.next_large_id;
        self.next_large_id += 1;

        #[cfg(feature = "debug")]
        let actual_size = self.debug_allocator.actual_size(size);
        #[cfg(not(feature = "debug"))]
        let actual_size = size;

        let mut vec = Vec::with_capacity(actual_size);
        vec.resize(actual_size, 0);

        #[cfg(feature = "debug")]
        {
            self.debug_allocator.write_canaries(&mut vec, size);
        }

        let ptr = vec.as_ptr() as usize;

        #[cfg(feature = "debug")]
        let user_ptr = ptr + self.debug_allocator.user_offset();
        #[cfg(not(feature = "debug"))]
        let user_ptr = ptr;

        self.large_allocations.insert(id, vec);
        self.allocation_sizes
            .insert(user_ptr, (size, SlabSize::Size1024));

        Some(user_ptr)
    }

    pub fn free(&mut self, ptr: usize, size: usize) {
        if ptr == 0 {
            panic!("Attempted to free null pointer");
        }

        if self.is_pool_address(ptr) {
            self.free_pool(ptr, size);
        } else {
            self.free_large(ptr, size);
        }
    }

    fn free_pool(&mut self, ptr: usize, size: usize) {
        let offset = match self.offset_from_ptr(ptr) {
            Some(o) => o,
            None => panic!(
                "Invalid pointer 0x{:x} - not within pool range [0x{:x}, 0x{:x})",
                ptr,
                self.pool_start(),
                self.pool_end()
            ),
        };

        let (allocated_size, slab_size) = match self.allocation_sizes.remove(&offset) {
            Some(s) => s,
            None => panic!(
                "Invalid pointer 0x{:x} - not a valid allocation (offset: {})",
                ptr, offset
            ),
        };

        if size > allocated_size && size > slab_size.to_size() {
            panic!(
                "Invalid free size: {} for allocation of size {} (slab size: {})",
                size,
                allocated_size,
                slab_size.to_size()
            );
        }

        #[cfg(feature = "debug")]
        {
            let slab_offset = offset - self.debug_allocator.user_offset();
            let slab_size_val = slab_size.to_size();

            if slab_offset + slab_size_val <= self.pool.len() {
                let slab_slice = &self.pool[slab_offset..slab_offset + slab_size_val];
                if let Err(e) =
                    self.debug_allocator
                        .verify_canaries(slab_slice, allocated_size, slab_offset)
                {
                    match e {
                        CanaryError::BufferTooSmall => {
                            eprintln!("Canary check failed: buffer too small");
                        }
                        CanaryError::PrefixCorrupted {
                            offset_from_start,
                            byte_offset,
                        } => {
                            let abs_addr = self.pool_start() + offset_from_start;
                            eprintln!(
                                "Prefix canary corrupted at byte offset {} (absolute address 0x{:x}) - underflow detected",
                                byte_offset, abs_addr
                            );
                        }
                        CanaryError::SuffixCorrupted {
                            offset_from_start,
                            byte_offset,
                            user_size,
                        } => {
                            let abs_addr = self.pool_start() + offset_from_start;
                            eprintln!(
                                "Suffix canary corrupted at byte offset {} (absolute address 0x{:x}) - overflow detected at +{} bytes beyond user data (user size: {})",
                                byte_offset, abs_addr, byte_offset, user_size
                            );
                        }
                    }
                }
            }

            let slab_offset = offset - self.debug_allocator.user_offset();
            self.free_slab(slab_offset, slab_size);
        }

        #[cfg(not(feature = "debug"))]
        {
            self.free_slab(offset, slab_size);
        }
    }

    fn free_slab(&mut self, offset: usize, slab_size: SlabSize) {
        let index = slab_size.to_index();

        if let Some(ref mut allocator) = &mut self.slab_allocators[index] {
            if !allocator.free(offset) {
                panic!(
                    "Failed to free slab at offset {} with size {}",
                    offset,
                    slab_size.to_size()
                );
            }
        }
    }

    fn free_large(&mut self, ptr: usize, _size: usize) {
        let mut found = false;
        let mut to_remove = None;

        #[cfg(feature = "debug")]
        let base_ptr = ptr - self.debug_allocator.user_offset();
        #[cfg(not(feature = "debug"))]
        let base_ptr = ptr;

        for (id, vec) in &self.large_allocations {
            let vec_ptr = vec.as_ptr() as usize;
            if vec_ptr == base_ptr {
                found = true;

                #[cfg(feature = "debug")]
                {
                    if let Some((allocated_size, _)) = self.allocation_sizes.get(&ptr) {
                        if let Err(e) =
                            self.debug_allocator
                                .verify_canaries(vec, *allocated_size, 0)
                        {
                            match e {
                                CanaryError::BufferTooSmall => {
                                    eprintln!("Canary check failed: buffer too small");
                                }
                                CanaryError::PrefixCorrupted {
                                    offset_from_start,
                                    byte_offset,
                                } => {
                                    let abs_addr = base_ptr + offset_from_start;
                                    eprintln!(
                                        "Prefix canary corrupted at byte offset {} (absolute address 0x{:x}) - underflow detected",
                                        byte_offset, abs_addr
                                    );
                                }
                                CanaryError::SuffixCorrupted {
                                    offset_from_start,
                                    byte_offset,
                                    user_size,
                                } => {
                                    let abs_addr = base_ptr + offset_from_start;
                                    eprintln!(
                                        "Suffix canary corrupted at byte offset {} (absolute address 0x{:x}) - overflow detected at +{} bytes beyond user data (user size: {})",
                                        byte_offset, abs_addr, byte_offset, user_size
                                    );
                                }
                            }
                        }
                    }
                }

                to_remove = Some(*id);
                break;
            }
        }

        if !found {
            panic!(
                "Invalid pointer 0x{:x} - not a valid allocation (not in pool and not a large allocation)",
                ptr
            );
        }

        if let Some(id) = to_remove {
            self.large_allocations.remove(&id);
            self.allocation_sizes.remove(&ptr);
        }
    }

    pub fn total_memory(&self) -> usize {
        self.pool.len()
    }

    pub fn used_memory(&self) -> usize {
        let mut used = 0;

        for allocator in &self.slab_allocators {
            if let Some(ref a) = allocator {
                used += a.used_slabs() * a.slab_size().to_size();
            }
        }

        for vec in self.large_allocations.values() {
            used += vec.len();
        }

        used
    }

    pub fn free_memory(&self) -> usize {
        let mut free = 0;

        for allocator in &self.slab_allocators {
            if let Some(ref a) = allocator {
                free += a.free_memory();
            }
        }

        free
    }

    pub fn fragmentation_ratio(&self) -> f64 {
        let total_free = self.free_memory();

        if total_free == 0 {
            return 0.0;
        }

        let mut max_contiguous = 0;

        for allocator in &self.slab_allocators {
            if let Some(ref a) = allocator {
                let contiguous = a.max_contiguous_free() * a.slab_size().to_size();
                if contiguous > max_contiguous {
                    max_contiguous = contiguous;
                }
            }
        }

        1.0 - (max_contiguous as f64) / (total_free as f64)
    }

    pub fn stats(&self) {
        println!("=== Memory Pool Statistics ===");
        println!("Total Memory: {} bytes", self.total_memory());
        println!("Used Memory: {} bytes", self.used_memory());
        println!("Free Memory: {} bytes", self.free_memory());
        println!(
            "Fragmentation Ratio: {:.2}%",
            self.fragmentation_ratio() * 100.0
        );
        println!();
        println!("Slab Allocators:");

        for (i, allocator) in self.slab_allocators.iter().enumerate() {
            if let Some(ref a) = allocator {
                let slab_size = a.slab_size().to_size();
                println!(
                    "  {} bytes: {}/{} slabs used, {} bytes free, max contiguous: {} bytes",
                    slab_size,
                    a.used_slabs(),
                    a.total_slabs(),
                    a.free_memory(),
                    a.max_contiguous_free() * slab_size
                );
            }
        }

        println!();
        println!(
            "Large allocations: {} allocations",
            self.large_allocations.len()
        );
    }

    pub fn get_slice(&self, ptr: usize, size: usize) -> Option<&[u8]> {
        if self.is_pool_address(ptr) {
            let offset = self.offset_from_ptr(ptr)?;
            if offset + size <= self.pool.len() {
                Some(&self.pool[offset..offset + size])
            } else {
                None
            }
        } else {
            #[cfg(feature = "debug")]
            let base_ptr = ptr - self.debug_allocator.user_offset();
            #[cfg(not(feature = "debug"))]
            let base_ptr = ptr;

            for vec in self.large_allocations.values() {
                let vec_ptr = vec.as_ptr() as usize;
                if vec_ptr == base_ptr {
                    #[cfg(feature = "debug")]
                    let offset = self.debug_allocator.user_offset();
                    #[cfg(not(feature = "debug"))]
                    let offset = 0;

                    if offset + size <= vec.len() {
                        return Some(&vec[offset..offset + size]);
                    }
                }
            }
            None
        }
    }

    pub fn get_slice_mut(&mut self, ptr: usize, size: usize) -> Option<&mut [u8]> {
        if self.is_pool_address(ptr) {
            let offset = self.offset_from_ptr(ptr)?;
            if offset + size <= self.pool.len() {
                Some(&mut self.pool[offset..offset + size])
            } else {
                None
            }
        } else {
            #[cfg(feature = "debug")]
            let base_ptr = ptr - self.debug_allocator.user_offset();
            #[cfg(not(feature = "debug"))]
            let base_ptr = ptr;

            for vec in self.large_allocations.values_mut() {
                let vec_ptr = vec.as_ptr() as usize;
                if vec_ptr == base_ptr {
                    #[cfg(feature = "debug")]
                    let offset = self.debug_allocator.user_offset();
                    #[cfg(not(feature = "debug"))]
                    let offset = 0;

                    if offset + size <= vec.len() {
                        return Some(&mut vec[offset..offset + size]);
                    }
                }
            }
            None
        }
    }
}
