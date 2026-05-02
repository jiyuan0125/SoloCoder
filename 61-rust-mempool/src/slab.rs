const SLAB_SIZES: [usize; 5] = [64, 128, 256, 512, 1024];

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum SlabSize {
    Size64,
    Size128,
    Size256,
    Size512,
    Size1024,
}

impl SlabSize {
    pub fn from_size(size: usize) -> Option<Self> {
        match size {
            0..=64 => Some(SlabSize::Size64),
            65..=128 => Some(SlabSize::Size128),
            129..=256 => Some(SlabSize::Size256),
            257..=512 => Some(SlabSize::Size512),
            513..=1024 => Some(SlabSize::Size1024),
            _ => None,
        }
    }

    pub fn to_size(self) -> usize {
        match self {
            SlabSize::Size64 => 64,
            SlabSize::Size128 => 128,
            SlabSize::Size256 => 256,
            SlabSize::Size512 => 512,
            SlabSize::Size1024 => 1024,
        }
    }

    pub fn from_index(index: usize) -> Option<Self> {
        match index {
            0 => Some(SlabSize::Size64),
            1 => Some(SlabSize::Size128),
            2 => Some(SlabSize::Size256),
            3 => Some(SlabSize::Size512),
            4 => Some(SlabSize::Size1024),
            _ => None,
        }
    }

    pub fn to_index(self) -> usize {
        match self {
            SlabSize::Size64 => 0,
            SlabSize::Size128 => 1,
            SlabSize::Size256 => 2,
            SlabSize::Size512 => 3,
            SlabSize::Size1024 => 4,
        }
    }

    pub fn slab_sizes() -> &'static [usize] {
        &SLAB_SIZES
    }
}

#[derive(Debug, Clone)]
pub struct SlabAllocator {
    slab_size: SlabSize,
    base_offset: usize,
    total_slabs: usize,
    used_slabs: usize,
    free_list: Vec<usize>,
}

impl SlabAllocator {
    pub fn new(slab_size: SlabSize, base_offset: usize, total_slabs: usize) -> Self {
        let mut free_list = Vec::with_capacity(total_slabs);
        for i in 0..total_slabs {
            free_list.push(i);
        }
        free_list.reverse();

        SlabAllocator {
            slab_size,
            base_offset,
            total_slabs,
            used_slabs: 0,
            free_list,
        }
    }

    pub fn slab_size(&self) -> SlabSize {
        self.slab_size
    }

    pub fn total_slabs(&self) -> usize {
        self.total_slabs
    }

    pub fn used_slabs(&self) -> usize {
        self.used_slabs
    }

    pub fn free_slabs(&self) -> usize {
        self.total_slabs - self.used_slabs
    }

    pub fn free_memory(&self) -> usize {
        self.free_slabs() * self.slab_size.to_size()
    }

    pub fn max_contiguous_free(&self) -> usize {
        if self.free_slabs() == 0 {
            return 0;
        }

        let mut bitmap = vec![true; self.total_slabs];
        for &idx in &self.free_list {
            if idx < self.total_slabs {
                bitmap[idx] = false;
            }
        }

        let mut max_count = 0;
        let mut current_count = 0;

        for &is_used in &bitmap {
            if !is_used {
                current_count += 1;
                if current_count > max_count {
                    max_count = current_count;
                }
            } else {
                current_count = 0;
            }
        }

        max_count
    }

    pub fn alloc(&mut self) -> Option<usize> {
        if let Some(slab_idx) = self.free_list.pop() {
            self.used_slabs += 1;
            Some(self.base_offset + slab_idx * self.slab_size.to_size())
        } else {
            None
        }
    }

    pub fn free(&mut self, offset: usize) -> bool {
        if offset < self.base_offset {
            return false;
        }

        let relative_offset = offset - self.base_offset;
        let slab_size = self.slab_size.to_size();

        if relative_offset % slab_size != 0 {
            return false;
        }

        let slab_idx = relative_offset / slab_size;

        if slab_idx >= self.total_slabs {
            return false;
        }

        if self.free_list.contains(&slab_idx) {
            return false;
        }

        self.free_list.push(slab_idx);
        self.used_slabs -= 1;
        true
    }

    pub fn contains_offset(&self, offset: usize) -> bool {
        let size = self.total_slabs * self.slab_size.to_size();
        offset >= self.base_offset && offset < self.base_offset + size
    }

    pub fn get_slab_index(&self, offset: usize) -> Option<usize> {
        if !self.contains_offset(offset) {
            return None;
        }

        let relative_offset = offset - self.base_offset;
        let slab_size = self.slab_size.to_size();

        if relative_offset % slab_size != 0 {
            return None;
        }

        Some(relative_offset / slab_size)
    }
}
