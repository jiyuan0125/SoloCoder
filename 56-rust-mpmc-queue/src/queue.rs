use std::cell::UnsafeCell;
use std::mem::MaybeUninit;
use std::sync::atomic::Ordering;
use std::sync::{Condvar, Mutex};

use crate::atomic::{AtomicBool, AtomicIndex, CachePadded};
use crate::error::{QueueError, QueueResult};

#[repr(align(64))]
struct Slot<T> {
    value: UnsafeCell<MaybeUninit<T>>,
    ready: AtomicBool,
}

impl<T> Slot<T> {
    fn new() -> Self {
        Slot {
            value: UnsafeCell::new(MaybeUninit::uninit()),
            ready: AtomicBool::new(false),
        }
    }
}

unsafe impl<T: Send> Send for Slot<T> {}
unsafe impl<T: Send> Sync for Slot<T> {}

struct Inner<T> {
    buffer: Vec<Slot<T>>,
    user_capacity: usize,
    buffer_size: usize,
    head: CachePadded<AtomicIndex>,
    tail: CachePadded<AtomicIndex>,
    closed: AtomicBool,
    mutex: Mutex<()>,
    not_empty: Condvar,
    not_full: Condvar,
}

impl<T> Inner<T> {
    fn new(user_capacity: usize) -> Self {
        let buffer_size = user_capacity + 1;
        let mut buffer = Vec::with_capacity(buffer_size);
        for _ in 0..buffer_size {
            buffer.push(Slot::new());
        }

        Inner {
            buffer,
            user_capacity,
            buffer_size,
            head: CachePadded::new(AtomicIndex::new(0)),
            tail: CachePadded::new(AtomicIndex::new(0)),
            closed: AtomicBool::new(false),
            mutex: Mutex::new(()),
            not_empty: Condvar::new(),
            not_full: Condvar::new(),
        }
    }

    fn is_closed(&self) -> bool {
        self.closed.load(Ordering::SeqCst)
    }

    fn len(&self) -> usize {
        let head = self.head.load(Ordering::SeqCst);
        let tail = self.tail.load(Ordering::SeqCst);
        if tail >= head {
            tail - head
        } else {
            (tail + self.buffer_size) - head
        }
    }

    fn is_full(&self) -> bool {
        let head = self.head.load(Ordering::SeqCst);
        let tail = self.tail.load(Ordering::SeqCst);
        (tail + 1) % self.buffer_size == head
    }

    fn is_empty(&self) -> bool {
        let head = self.head.load(Ordering::SeqCst);
        let tail = self.tail.load(Ordering::SeqCst);
        head == tail
    }
}

pub struct MpmcQueue<T> {
    inner: std::sync::Arc<Inner<T>>,
}

impl<T> MpmcQueue<T> {
    pub fn new(capacity: usize) -> Self {
        assert!(capacity > 0, "Capacity must be greater than 0");
        MpmcQueue {
            inner: std::sync::Arc::new(Inner::new(capacity)),
        }
    }

    pub fn try_push(&self, value: T) -> QueueResult<()> {
        if self.inner.is_closed() {
            return Err(QueueError::Closed);
        }

        let _guard = self.inner.mutex.lock().unwrap();

        if self.inner.is_closed() {
            return Err(QueueError::Closed);
        }

        if self.inner.is_full() {
            return Err(QueueError::Full);
        }

        let tail = self.inner.tail.load(Ordering::SeqCst);
        let slot = &self.inner.buffer[tail];

        unsafe {
            (*slot.value.get()).as_mut_ptr().write(value);
        }
        slot.ready.store(true, Ordering::Release);

        let next_tail = (tail + 1) % self.inner.buffer_size;
        self.inner.tail.store(next_tail, Ordering::SeqCst);

        self.inner.not_empty.notify_one();

        Ok(())
    }

    pub fn push(&self, value: T) -> QueueResult<()> {
        let mut guard = self.inner.mutex.lock().unwrap();

        loop {
            if self.inner.is_closed() {
                return Err(QueueError::Closed);
            }

            if !self.inner.is_full() {
                break;
            }

            guard = self.inner.not_full.wait(guard).unwrap();
        }

        let tail = self.inner.tail.load(Ordering::SeqCst);
        let slot = &self.inner.buffer[tail];

        unsafe {
            (*slot.value.get()).as_mut_ptr().write(value);
        }
        slot.ready.store(true, Ordering::Release);

        let next_tail = (tail + 1) % self.inner.buffer_size;
        self.inner.tail.store(next_tail, Ordering::SeqCst);

        self.inner.not_empty.notify_one();

        Ok(())
    }

    pub fn try_pop(&self) -> QueueResult<T> {
        let _guard = self.inner.mutex.lock().unwrap();

        if self.inner.is_empty() {
            if self.inner.is_closed() {
                return Err(QueueError::Closed);
            }
            return Err(QueueError::Empty);
        }

        let head = self.inner.head.load(Ordering::SeqCst);
        let slot = &self.inner.buffer[head];

        if !slot.ready.load(Ordering::Acquire) {
            if self.inner.is_closed() {
                return Err(QueueError::Closed);
            }
            return Err(QueueError::Empty);
        }

        let value = unsafe { (*slot.value.get()).as_ptr().read() };
        slot.ready.store(false, Ordering::Release);

        let next_head = (head + 1) % self.inner.buffer_size;
        self.inner.head.store(next_head, Ordering::SeqCst);

        self.inner.not_full.notify_one();

        Ok(value)
    }

    pub fn pop(&self) -> QueueResult<T> {
        let mut guard = self.inner.mutex.lock().unwrap();

        loop {
            if self.inner.is_empty() {
                if self.inner.is_closed() {
                    return Err(QueueError::Closed);
                }
                guard = self.inner.not_empty.wait(guard).unwrap();
            } else {
                break;
            }
        }

        let head = self.inner.head.load(Ordering::SeqCst);
        let slot = &self.inner.buffer[head];

        if !slot.ready.load(Ordering::Acquire) {
            if self.inner.is_closed() {
                return Err(QueueError::Closed);
            }
            return Err(QueueError::Empty);
        }

        let value = unsafe { (*slot.value.get()).as_ptr().read() };
        slot.ready.store(false, Ordering::Release);

        let next_head = (head + 1) % self.inner.buffer_size;
        self.inner.head.store(next_head, Ordering::SeqCst);

        self.inner.not_full.notify_one();

        Ok(value)
    }

    pub fn close(&self) {
        self.inner.closed.store(true, Ordering::SeqCst);
        self.inner.not_empty.notify_all();
        self.inner.not_full.notify_all();
    }

    pub fn is_closed(&self) -> bool {
        self.inner.is_closed()
    }

    pub fn len(&self) -> usize {
        self.inner.len()
    }

    pub fn capacity(&self) -> usize {
        self.inner.user_capacity
    }

    pub fn is_empty(&self) -> bool {
        self.inner.is_empty()
    }

    pub fn is_full(&self) -> bool {
        self.inner.is_full()
    }
}

impl<T> Clone for MpmcQueue<T> {
    fn clone(&self) -> Self {
        MpmcQueue {
            inner: self.inner.clone(),
        }
    }
}

pub struct IntoIter<T> {
    queue: MpmcQueue<T>,
}

impl<T> IntoIterator for MpmcQueue<T> {
    type Item = T;
    type IntoIter = IntoIter<T>;

    fn into_iter(self) -> Self::IntoIter {
        self.close();
        IntoIter { queue: self }
    }
}

impl<T> Iterator for IntoIter<T> {
    type Item = T;

    fn next(&mut self) -> Option<Self::Item> {
        self.queue.try_pop().ok()
    }
}

unsafe impl<T: Send> Send for MpmcQueue<T> {}
unsafe impl<T: Send> Sync for MpmcQueue<T> {}
