use std::cell::UnsafeCell;
use std::mem::MaybeUninit;
use std::sync::atomic::{AtomicBool, AtomicUsize, Ordering};

#[repr(align(64))]
pub struct CachePadded<T> {
    value: UnsafeCell<MaybeUninit<T>>,
}

impl<T> CachePadded<T> {
    pub fn new() -> Self {
        CachePadded {
            value: UnsafeCell::new(MaybeUninit::uninit()),
        }
    }

    pub unsafe fn write(&self, value: T) {
        unsafe {
            (*self.value.get()).write(value);
        }
    }

    pub unsafe fn read(&self) -> T {
        unsafe { (*self.value.get()).assume_init_read() }
    }
}

#[repr(align(64))]
pub struct CachePaddedAtomicUsize {
    inner: AtomicUsize,
}

impl CachePaddedAtomicUsize {
    pub fn new(value: usize) -> Self {
        CachePaddedAtomicUsize {
            inner: AtomicUsize::new(value),
        }
    }

    pub fn load(&self, order: Ordering) -> usize {
        self.inner.load(order)
    }

    pub fn store(&self, value: usize, order: Ordering) {
        self.inner.store(value, order);
    }

    pub fn swap(&self, value: usize, order: Ordering) -> usize {
        self.inner.swap(value, order)
    }

    pub fn fetch_add(&self, value: usize, order: Ordering) -> usize {
        self.inner.fetch_add(value, order)
    }
}

#[repr(align(64))]
pub struct CachePaddedAtomicBool {
    inner: AtomicBool,
}

impl CachePaddedAtomicBool {
    pub fn new(value: bool) -> Self {
        CachePaddedAtomicBool {
            inner: AtomicBool::new(value),
        }
    }

    pub fn load(&self, order: Ordering) -> bool {
        self.inner.load(order)
    }

    pub fn store(&self, value: bool, order: Ordering) {
        self.inner.store(value, order);
    }

    pub fn swap(&self, value: bool, order: Ordering) -> bool {
        self.inner.swap(value, order)
    }
}
