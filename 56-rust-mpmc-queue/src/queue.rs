use crate::atomic::{CachePadded, CachePaddedAtomicBool, CachePaddedAtomicUsize};
use std::error::Error;
use std::fmt;
use std::mem::MaybeUninit;
use std::sync::atomic::Ordering::{Acquire, Relaxed, Release, SeqCst};
use std::sync::{Arc, Condvar, Mutex};
use std::thread;

#[derive(Debug, PartialEq, Eq)]
pub enum PushError<T> {
    Full(T),
    Closed(T),
}

impl<T> fmt::Display for PushError<T> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            PushError::Full(_) => write!(f, "queue is full"),
            PushError::Closed(_) => write!(f, "queue is closed"),
        }
    }
}

impl<T: fmt::Debug> Error for PushError<T> {}

#[derive(Debug, PartialEq, Eq)]
pub enum PopError {
    Empty,
    Closed,
}

impl fmt::Display for PopError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            PopError::Empty => write!(f, "queue is empty"),
            PopError::Closed => write!(f, "queue is closed"),
        }
    }
}

impl Error for PopError {}

struct Slot<T> {
    value: CachePadded<MaybeUninit<T>>,
    ready: CachePaddedAtomicBool,
}

impl<T> Slot<T> {
    fn new() -> Self {
        Slot {
            value: CachePadded::new(),
            ready: CachePaddedAtomicBool::new(false),
        }
    }
}

struct Inner<T> {
    buffer: Box<[Slot<T>]>,
    capacity: usize,
    buffer_len: usize,
    head: CachePaddedAtomicUsize,
    tail: CachePaddedAtomicUsize,
    closed: CachePaddedAtomicBool,
    mutex: Mutex<()>,
    not_empty: Condvar,
    not_full: Condvar,
}

impl<T> Inner<T> {
    fn new(capacity: usize) -> Self {
        let buffer_len = capacity + 1;
        let mut buffer = Vec::with_capacity(buffer_len);
        for _ in 0..buffer_len {
            buffer.push(Slot::new());
        }
        Inner {
            buffer: buffer.into_boxed_slice(),
            capacity,
            buffer_len,
            head: CachePaddedAtomicUsize::new(0),
            tail: CachePaddedAtomicUsize::new(0),
            closed: CachePaddedAtomicBool::new(false),
            mutex: Mutex::new(()),
            not_empty: Condvar::new(),
            not_full: Condvar::new(),
        }
    }

    fn is_closed(&self) -> bool {
        self.closed.load(Acquire)
    }

    fn len(&self) -> usize {
        let head = self.head.load(Acquire);
        let tail = self.tail.load(Acquire);
        if tail >= head {
            tail - head
        } else {
            self.buffer_len - head + tail
        }
    }

    fn is_empty(&self) -> bool {
        self.head.load(Acquire) == self.tail.load(Acquire)
    }

    fn is_full(&self) -> bool {
        let head = self.head.load(Acquire);
        let tail = self.tail.load(Acquire);
        (tail + 1) % self.buffer_len == head
    }

    unsafe fn do_push(&self, value: T) -> Result<(), T> {
        loop {
            if self.is_closed() {
                return Err(value);
            }

            let tail = self.tail.load(Acquire);
            let next_tail = (tail + 1) % self.buffer_len;
            let head = self.head.load(Acquire);

            if next_tail == head {
                return Err(value);
            }

            if self
                .tail
                .compare_exchange_weak(tail, next_tail, Release, Relaxed)
                .is_ok()
            {
                let slot = &self.buffer[tail];
                unsafe {
                    slot.value.write(MaybeUninit::new(value));
                }
                slot.ready.store(true, Release);
                return Ok(());
            }

            thread::yield_now();
        }
    }

    unsafe fn do_pop(&self) -> Option<T> {
        loop {
            let head = self.head.load(Acquire);
            let tail = self.tail.load(Acquire);

            if head == tail {
                return None;
            }

            let next_head = (head + 1) % self.buffer_len;

            if self
                .head
                .compare_exchange_weak(head, next_head, Release, Relaxed)
                .is_ok()
            {
                let slot = &self.buffer[head];

                while !slot.ready.load(Acquire) {
                    thread::yield_now();
                }

                let value = unsafe { slot.value.read().assume_init() };
                slot.ready.store(false, Release);

                return Some(value);
            }

            thread::yield_now();
        }
    }
}

impl<T> Drop for Inner<T> {
    fn drop(&mut self) {
        while let Some(_) = unsafe { self.do_pop() } {}
    }
}

unsafe impl<T: Send> Send for Inner<T> {}
unsafe impl<T: Send> Sync for Inner<T> {}

pub struct Queue<T> {
    inner: Arc<Inner<T>>,
}

impl<T> Queue<T> {
    pub fn new(capacity: usize) -> Self {
        assert!(capacity > 0, "capacity must be greater than 0");
        Queue {
            inner: Arc::new(Inner::new(capacity)),
        }
    }

    pub fn push(&self, value: T) -> Result<(), PushError<T>> {
        if self.inner.is_closed() {
            return Err(PushError::Closed(value));
        }

        let mut value = value;
        loop {
            match unsafe { self.inner.do_push(value) } {
                Ok(()) => {
                    self.inner.not_empty.notify_one();
                    return Ok(());
                }
                Err(v) => {
                    value = v;
                    if self.inner.is_closed() {
                        return Err(PushError::Closed(value));
                    }
                    let guard = self.inner.mutex.lock().unwrap();
                    let _guard = self
                        .inner
                        .not_full
                        .wait_while(guard, |_| {
                            !self.inner.is_closed() && self.inner.is_full()
                        })
                        .unwrap();
                    if self.inner.is_closed() {
                        return Err(PushError::Closed(value));
                    }
                }
            }
        }
    }

    pub fn try_push(&self, value: T) -> Result<(), PushError<T>> {
        if self.inner.is_closed() {
            return Err(PushError::Closed(value));
        }

        match unsafe { self.inner.do_push(value) } {
            Ok(()) => {
                self.inner.not_empty.notify_one();
                Ok(())
            }
            Err(v) => {
                if self.inner.is_closed() {
                    Err(PushError::Closed(v))
                } else {
                    Err(PushError::Full(v))
                }
            }
        }
    }

    pub fn pop(&self) -> Result<T, PopError> {
        loop {
            match unsafe { self.inner.do_pop() } {
                Some(value) => {
                    self.inner.not_full.notify_one();
                    return Ok(value);
                }
                None => {
                    if self.inner.is_closed() {
                        return Err(PopError::Closed);
                    }
                    let guard = self.inner.mutex.lock().unwrap();
                    let _guard = self
                        .inner
                        .not_empty
                        .wait_while(guard, |_| {
                            !self.inner.is_closed() && self.inner.is_empty()
                        })
                        .unwrap();
                    if self.inner.is_closed() && self.inner.is_empty() {
                        return Err(PopError::Closed);
                    }
                }
            }
        }
    }

    pub fn try_pop(&self) -> Result<T, PopError> {
        match unsafe { self.inner.do_pop() } {
            Some(value) => {
                self.inner.not_full.notify_one();
                Ok(value)
            }
            None => {
                if self.inner.is_closed() {
                    Err(PopError::Closed)
                } else {
                    Err(PopError::Empty)
                }
            }
        }
    }

    pub fn close(&self) {
        if !self.inner.closed.swap(true, SeqCst) {
            self.inner.not_empty.notify_all();
            self.inner.not_full.notify_all();
        }
    }

    pub fn is_closed(&self) -> bool {
        self.inner.is_closed()
    }

    pub fn len(&self) -> usize {
        self.inner.len()
    }

    pub fn is_empty(&self) -> bool {
        self.inner.is_empty()
    }

    pub fn capacity(&self) -> usize {
        self.inner.capacity
    }

    pub fn into_iter(self) -> IntoIter<T> {
        self.close();
        IntoIter { queue: self }
    }
}

impl<T> Clone for Queue<T> {
    fn clone(&self) -> Self {
        Queue {
            inner: self.inner.clone(),
        }
    }
}

pub struct IntoIter<T> {
    queue: Queue<T>,
}

impl<T> Iterator for IntoIter<T> {
    type Item = T;

    fn next(&mut self) -> Option<Self::Item> {
        self.queue.try_pop().ok()
    }
}

pub fn channel<T>(capacity: usize) -> (Sender<T>, Receiver<T>) {
    let queue = Queue::new(capacity);
    (
        Sender { inner: queue.inner.clone() },
        Receiver { inner: queue.inner },
    )
}

pub struct Sender<T> {
    inner: Arc<Inner<T>>,
}

impl<T> Sender<T> {
    pub fn push(&self, value: T) -> Result<(), PushError<T>> {
        if self.inner.is_closed() {
            return Err(PushError::Closed(value));
        }

        let mut value = value;
        loop {
            match unsafe { self.inner.do_push(value) } {
                Ok(()) => {
                    self.inner.not_empty.notify_one();
                    return Ok(());
                }
                Err(v) => {
                    value = v;
                    if self.inner.is_closed() {
                        return Err(PushError::Closed(value));
                    }
                    let guard = self.inner.mutex.lock().unwrap();
                    let _guard = self
                        .inner
                        .not_full
                        .wait_while(guard, |_| {
                            !self.inner.is_closed() && self.inner.is_full()
                        })
                        .unwrap();
                    if self.inner.is_closed() {
                        return Err(PushError::Closed(value));
                    }
                }
            }
        }
    }

    pub fn try_push(&self, value: T) -> Result<(), PushError<T>> {
        if self.inner.is_closed() {
            return Err(PushError::Closed(value));
        }

        match unsafe { self.inner.do_push(value) } {
            Ok(()) => {
                self.inner.not_empty.notify_one();
                Ok(())
            }
            Err(v) => {
                if self.inner.is_closed() {
                    Err(PushError::Closed(v))
                } else {
                    Err(PushError::Full(v))
                }
            }
        }
    }

    pub fn close(&self) {
        if !self.inner.closed.swap(true, SeqCst) {
            self.inner.not_empty.notify_all();
            self.inner.not_full.notify_all();
        }
    }

    pub fn is_closed(&self) -> bool {
        self.inner.is_closed()
    }

    pub fn len(&self) -> usize {
        self.inner.len()
    }

    pub fn capacity(&self) -> usize {
        self.inner.capacity
    }
}

impl<T> Clone for Sender<T> {
    fn clone(&self) -> Self {
        Sender {
            inner: self.inner.clone(),
        }
    }
}

pub struct Receiver<T> {
    inner: Arc<Inner<T>>,
}

impl<T> Receiver<T> {
    pub fn pop(&self) -> Result<T, PopError> {
        loop {
            match unsafe { self.inner.do_pop() } {
                Some(value) => {
                    self.inner.not_full.notify_one();
                    return Ok(value);
                }
                None => {
                    if self.inner.is_closed() {
                        return Err(PopError::Closed);
                    }
                    let guard = self.inner.mutex.lock().unwrap();
                    let _guard = self
                        .inner
                        .not_empty
                        .wait_while(guard, |_| {
                            !self.inner.is_closed() && self.inner.is_empty()
                        })
                        .unwrap();
                    if self.inner.is_closed() && self.inner.is_empty() {
                        return Err(PopError::Closed);
                    }
                }
            }
        }
    }

    pub fn try_pop(&self) -> Result<T, PopError> {
        match unsafe { self.inner.do_pop() } {
            Some(value) => {
                self.inner.not_full.notify_one();
                Ok(value)
            }
            None => {
                if self.inner.is_closed() {
                    Err(PopError::Closed)
                } else {
                    Err(PopError::Empty)
                }
            }
        }
    }

    pub fn close(&self) {
        if !self.inner.closed.swap(true, SeqCst) {
            self.inner.not_empty.notify_all();
            self.inner.not_full.notify_all();
        }
    }

    pub fn is_closed(&self) -> bool {
        self.inner.is_closed()
    }

    pub fn len(&self) -> usize {
        self.inner.len()
    }

    pub fn is_empty(&self) -> bool {
        self.inner.is_empty()
    }

    pub fn capacity(&self) -> usize {
        self.inner.capacity
    }
}

impl<T> Clone for Receiver<T> {
    fn clone(&self) -> Self {
        Receiver {
            inner: self.inner.clone(),
        }
    }
}
