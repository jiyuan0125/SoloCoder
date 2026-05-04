#include "ring_buffer.h"
#include <stdlib.h>
#include <string.h>
#include <sched.h>

#define RB_MSG_PREFIX_SIZE 4
#define RB_ALIGNMENT 8
#define RB_ALIGN_UP(x) (((x) + RB_ALIGNMENT - 1) & ~(RB_ALIGNMENT - 1))
#define RB_MIN(a, b) ((a) < (b) ? (a) : (b))

static size_t round_up_power_of_two(size_t size) {
    if (size == 0) return 0;
    size--;
    size |= size >> 1;
    size |= size >> 2;
    size |= size >> 4;
    size |= size >> 8;
    size |= size >> 16;
#if SIZE_MAX > 0xFFFFFFFF
    size |= size >> 32;
#endif
    size++;
    return size;
}

static void spin_lock(_Atomic(uint32_t)* lock) {
    int backoff = 1;
    while (true) {
        if (atomic_load_explicit(lock, memory_order_relaxed) == 0) {
            if (atomic_exchange_explicit(lock, 1, memory_order_acquire) == 0) {
                return;
            }
        }
        
        for (int i = 0; i < backoff; i++) {
            __builtin_ia32_pause();
        }
        
        if (backoff < 256) {
            backoff <<= 1;
        } else {
            sched_yield();
        }
    }
}

static void spin_unlock(_Atomic(uint32_t)* lock) {
    atomic_store_explicit(lock, 0, memory_order_release);
}

RingBuffer* ring_buffer_create(const RingBufferConfig* config) {
    if (!config || config->buffer_size == 0) {
        return NULL;
    }

    size_t buffer_size = round_up_power_of_two(config->buffer_size);
    if (buffer_size < 65536) {
        buffer_size = 65536;
    }

    RingBuffer* rb = (RingBuffer*)calloc(1, sizeof(RingBuffer));
    if (!rb) {
        return NULL;
    }

    rb->buffer = (uint8_t*)malloc(buffer_size);
    if (!rb->buffer) {
        free(rb);
        return NULL;
    }

    memset(rb->buffer, 0, buffer_size);

    rb->buffer_size = buffer_size;
    rb->buffer_mask = buffer_size - 1;
    rb->policy = config->policy;

    atomic_init(&rb->write_seq, 0);
    atomic_init(&rb->read_seq, 0);
    atomic_init(&rb->write_lock, 0);
    atomic_init(&rb->is_running, true);
    atomic_init(&rb->dropped_count, 0);

    return rb;
}

void ring_buffer_destroy(RingBuffer* rb) {
    if (rb) {
        atomic_store(&rb->is_running, false);
        free(rb->buffer);
        free(rb);
    }
}

static size_t pointer_diff(size_t a, size_t b) {
    return a - b;
}

static void rb_write_at(RingBuffer* rb, size_t pos, const void* data, size_t len) {
    size_t offset = pos & rb->buffer_mask;
    size_t first_chunk = RB_MIN(len, rb->buffer_size - offset);

    memcpy(rb->buffer + offset, data, first_chunk);
    if (first_chunk < len) {
        memcpy(rb->buffer, (const uint8_t*)data + first_chunk, len - first_chunk);
    }
}

static void rb_read_at(RingBuffer* rb, size_t pos, void* data, size_t len) {
    size_t offset = pos & rb->buffer_mask;
    size_t first_chunk = RB_MIN(len, rb->buffer_size - offset);

    memcpy(data, rb->buffer + offset, first_chunk);
    if (first_chunk < len) {
        memcpy((uint8_t*)data + first_chunk, rb->buffer, len - first_chunk);
    }
}

bool ring_buffer_write(RingBuffer* rb, const void* data, size_t len) {
    if (!rb || !data || len == 0) {
        return false;
    }

    if (!atomic_load_explicit(&rb->is_running, memory_order_relaxed)) {
        return false;
    }

    if (len > rb->buffer_size / 2) {
        return false;
    }

    size_t total_len = RB_ALIGN_UP(RB_MSG_PREFIX_SIZE + len);

    spin_lock(&rb->write_lock);

    size_t write_seq = atomic_load_explicit(&rb->write_seq, memory_order_relaxed);
    size_t read_seq = atomic_load_explicit(&rb->read_seq, memory_order_acquire);
    size_t used = pointer_diff(write_seq, read_seq);
    size_t available = rb->buffer_size - used;

    if (total_len > available) {
        if (rb->policy == RB_POLICY_BLOCK) {
            spin_unlock(&rb->write_lock);
            
            int spin_count = 0;
            while (atomic_load_explicit(&rb->is_running, memory_order_relaxed)) {
                spin_lock(&rb->write_lock);
                write_seq = atomic_load_explicit(&rb->write_seq, memory_order_relaxed);
                read_seq = atomic_load_explicit(&rb->read_seq, memory_order_acquire);
                used = pointer_diff(write_seq, read_seq);
                available = rb->buffer_size - used;
                
                if (total_len <= available) {
                    break;
                }
                
                spin_unlock(&rb->write_lock);
                
                if (spin_count < 1000) {
                    spin_count++;
                    for (int i = 0; i < spin_count; i++) {
                        __builtin_ia32_pause();
                    }
                } else {
                    sched_yield();
                }
            }
            
            if (!atomic_load_explicit(&rb->is_running, memory_order_relaxed)) {
                return false;
            }
        } else {
            atomic_fetch_add_explicit(&rb->dropped_count, 1, memory_order_relaxed);
            spin_unlock(&rb->write_lock);
            return false;
        }
    }

    uint32_t len_prefix = (uint32_t)len;
    rb_write_at(rb, write_seq, &len_prefix, sizeof(len_prefix));
    rb_write_at(rb, write_seq + RB_MSG_PREFIX_SIZE, data, len);

    atomic_store_explicit(&rb->write_seq, write_seq + total_len, memory_order_release);

    spin_unlock(&rb->write_lock);
    return true;
}

bool ring_buffer_read(RingBuffer* rb, void* data, size_t max_len, size_t* bytes_read) {
    if (!rb || !data || !bytes_read || max_len == 0) {
        return false;
    }

    *bytes_read = 0;

    size_t read_seq = atomic_load_explicit(&rb->read_seq, memory_order_relaxed);
    size_t write_seq = atomic_load_explicit(&rb->write_seq, memory_order_acquire);

    if (read_seq >= write_seq) {
        return false;
    }

    size_t used = pointer_diff(write_seq, read_seq);

    if (used < RB_MSG_PREFIX_SIZE) {
        return false;
    }

    uint32_t msg_len;
    rb_read_at(rb, read_seq, &msg_len, sizeof(msg_len));

    if (msg_len == 0 || msg_len > 65536) {
        size_t skip = RB_ALIGN_UP(RB_MSG_PREFIX_SIZE + 1);
        atomic_store_explicit(&rb->read_seq, read_seq + skip, memory_order_relaxed);
        return false;
    }

    size_t total_msg_len = RB_ALIGN_UP(RB_MSG_PREFIX_SIZE + msg_len);

    if (used < total_msg_len) {
        return false;
    }

    size_t to_copy = RB_MIN(msg_len, max_len);
    rb_read_at(rb, read_seq + RB_MSG_PREFIX_SIZE, data, to_copy);

    atomic_store_explicit(&rb->read_seq, read_seq + total_msg_len, memory_order_release);

    *bytes_read = to_copy;
    return true;
}

size_t ring_buffer_available(const RingBuffer* rb) {
    if (!rb) return 0;
    size_t write_seq = atomic_load_explicit(&rb->write_seq, memory_order_acquire);
    size_t read_seq = atomic_load_explicit(&rb->read_seq, memory_order_acquire);
    size_t used = pointer_diff(write_seq, read_seq);
    return (used >= rb->buffer_size) ? 0 : (rb->buffer_size - used);
}

size_t ring_buffer_used(const RingBuffer* rb) {
    if (!rb) return 0;
    size_t write_seq = atomic_load_explicit(&rb->write_seq, memory_order_acquire);
    size_t read_seq = atomic_load_explicit(&rb->read_seq, memory_order_acquire);
    size_t used = pointer_diff(write_seq, read_seq);
    return (used > rb->buffer_size) ? rb->buffer_size : used;
}

size_t ring_buffer_capacity(const RingBuffer* rb) {
    if (!rb) return 0;
    return rb->buffer_size;
}

bool ring_buffer_is_empty(const RingBuffer* rb) {
    if (!rb) return true;
    size_t write_seq = atomic_load_explicit(&rb->write_seq, memory_order_acquire);
    size_t read_seq = atomic_load_explicit(&rb->read_seq, memory_order_acquire);
    return write_seq == read_seq;
}

bool ring_buffer_is_full(const RingBuffer* rb) {
    if (!rb) return true;
    return ring_buffer_available(rb) == 0;
}

size_t ring_buffer_dropped(const RingBuffer* rb) {
    if (!rb) return 0;
    return atomic_load_explicit(&rb->dropped_count, memory_order_relaxed);
}

void ring_buffer_set_policy(RingBuffer* rb, RingBufferPolicy policy) {
    if (rb) {
        rb->policy = policy;
    }
}
