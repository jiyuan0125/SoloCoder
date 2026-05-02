use std::sync::atomic::{AtomicUsize, Ordering};

#[repr(align(64))]
pub struct CachePadded<T> {
    value: T,
}

impl<T> CachePadded<T> {
    pub fn new(value: T) -> Self {
        CachePadded { value }
    }

    pub fn into_inner(self) -> T {
        self.value
    }
}

impl<T> std::ops::Deref for CachePadded<T> {
    type Target = T;

    fn deref(&self) -> &Self::Target {
        &self.value
    }
}

impl<T> std::ops::DerefMut for CachePadded<T> {
    fn deref_mut(&mut self) -> &mut Self::Target {
        &mut self.value
    }
}

#[repr(align(64))]
pub struct AtomicIndex {
    inner: AtomicUsize,
}

impl AtomicIndex {
    pub fn new(initial: usize) -> Self {
        AtomicIndex {
            inner: AtomicUsize::new(initial),
        }
    }

    pub fn load(&self, order: Ordering) -> usize {
        self.inner.load(order)
    }

    pub fn store(&self, val: usize, order: Ordering) {
        self.inner.store(val, order)
    }

    pub fn compare_exchange(
        &self,
        current: usize,
        new: usize,
        success: Ordering,
        failure: Ordering,
    ) -> Result<usize, usize> {
        self.inner.compare_exchange(current, new, success, failure)
    }

    pub fn fetch_add(&self, val: usize, order: Ordering) -> usize {
        self.inner.fetch_add(val, order)
    }

    pub fn fetch_sub(&self, val: usize, order: Ordering) -> usize {
        self.inner.fetch_sub(val, order)
    }
}

#[repr(align(64))]
pub struct AtomicBool {
    inner: std::sync::atomic::AtomicBool,
}

impl AtomicBool {
    pub fn new(initial: bool) -> Self {
        AtomicBool {
            inner: std::sync::atomic::AtomicBool::new(initial),
        }
    }

    pub fn load(&self, order: Ordering) -> bool {
        self.inner.load(order)
    }

    pub fn store(&self, val: bool, order: Ordering) {
        self.inner.store(val, order)
    }

    pub fn swap(&self, val: bool, order: Ordering) -> bool {
        self.inner.swap(val, order)
    }
}
