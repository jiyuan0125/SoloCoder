#include "ring_buffer.h"
#include <stdlib.h>
#include <string.h>
#include <sched.h>

#define RB_ALIGNMENT 8
#define RB_ALIGN_UP(x) (((x) + RB_ALIGNMENT - 1) & ~(RB_ALIGNMENT - 1))
#define RB_MSG_HEADER_SIZE 8
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

RingBuffer* ring_buffer_create(const RingBufferConfig* config) {
    if (!config || config->buffer_size == 0) {
        return NULL;
    }

    size_t buffer_size = round_up_power_of_two(config->buffer_size);
    if (buffer_size < 4096) {
        buffer_size = 4096;
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
    atomic_init(&rb->commit_seq, 0);
    atomic_init(&rb->read_seq, 0);
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

static bool rb_try_reserve(RingBuffer* rb, size_t len, size_t* out_pos) {
    size_t total_needed = RB_ALIGN_UP(RB_MSG_HEADER_SIZE + len);
    
    if (total_needed > rb->buffer_size / 2) {
        return false;
    }

    while (true) {
        size_t write_seq = atomic_load_explicit(&rb->write_seq, memory_order_relaxed);
        size_t read_seq = atomic_load_explicit(&rb->read_seq, memory_order_acquire);
        
        size_t used = pointer_diff(write_seq, read_seq);
        size_t available = rb->buffer_size - used;

        if (total_needed > available) {
            if (rb->policy == RB_POLICY_BLOCK) {
                return false;
            }
            
            size_t commit_seq = atomic_load_explicit(&rb->commit_seq, memory_order_acquire);
            size_t committed_used = pointer_diff(commit_seq, read_seq);
            
            if (committed_used + total_needed <= rb->buffer_size) {
            } else {
                size_t new_read_seq = read_seq + (committed_used + total_needed - rb->buffer_size);
                new_read_seq = RB_ALIGN_UP(new_read_seq);
                atomic_store_explicit(&rb->read_seq, new_read_seq, memory_order_release);
                atomic_fetch_add_explicit(&rb->dropped_count, 1, memory_order_relaxed);
            }
        }

        size_t new_write_seq = write_seq + total_needed;
        
        if (atomic_compare_exchange_weak_explicit(
                &rb->write_seq,
                &write_seq,
                new_write_seq,
                memory_order_acquire,
                memory_order_relaxed)) {
            *out_pos = write_seq;
            return true;
        }
    }
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

    size_t pos;
    int spin_count = 0;

    while (true) {
        if (!atomic_load_explicit(&rb->is_running, memory_order_relaxed)) {
            return false;
        }

        if (rb_try_reserve(rb, len, &pos)) {
            uint32_t header[2];
            header[0] = 0;
            header[1] = (uint32_t)len;
            
            rb_write_at(rb, pos, header, sizeof(header));
            rb_write_at(rb, pos + RB_MSG_HEADER_SIZE, data, len);
            
            atomic_thread_fence(memory_order_release);
            
            uint32_t magic = RB_MSG_COMMITTED;
            rb_write_at(rb, pos, &magic, sizeof(magic));
            
            return true;
        }

        if (rb->policy == RB_POLICY_DROP) {
            spin_count++;
            if (spin_count > 100) {
                atomic_fetch_add_explicit(&rb->dropped_count, 1, memory_order_relaxed);
                return false;
            }
        } else {
            spin_count++;
            if (spin_count < 1000) {
                for (volatile int i = 0; i < 20; i++);
            } else {
                sched_yield();
            }
        }
    }
}

static bool rb_is_msg_committed(RingBuffer* rb, size_t pos) {
    uint32_t magic;
    rb_read_at(rb, pos, &magic, sizeof(magic));
    return magic == RB_MSG_COMMITTED;
}

static bool rb_try_read_msg(RingBuffer* rb, void* data, size_t max_len, size_t* bytes_read) {
    size_t read_seq = atomic_load_explicit(&rb->read_seq, memory_order_relaxed);
    size_t commit_seq = atomic_load_explicit(&rb->commit_seq, memory_order_acquire);
    size_t write_seq = atomic_load_explicit(&rb->write_seq, memory_order_acquire);

    while (read_seq < commit_seq) {
        if (!rb_is_msg_committed(rb, read_seq)) {
            size_t skip = RB_ALIGN_UP(RB_MSG_HEADER_SIZE + 1);
            read_seq += skip;
            continue;
        }

        uint32_t header[2];
        rb_read_at(rb, read_seq, header, sizeof(header));
        
        uint32_t magic = header[0];
        uint32_t len = header[1];

        if (magic != RB_MSG_COMMITTED) {
            size_t skip = RB_ALIGN_UP(RB_MSG_HEADER_SIZE + 1);
            read_seq += skip;
            continue;
        }

        if (len == 0 || len > 65536) {
            size_t skip = RB_ALIGN_UP(RB_MSG_HEADER_SIZE + 1);
            read_seq += skip;
            continue;
        }

        size_t total_len = RB_ALIGN_UP(RB_MSG_HEADER_SIZE + len);
        
        if (read_seq + total_len > commit_seq) {
            break;
        }

        size_t to_copy = RB_MIN(len, max_len);
        rb_read_at(rb, read_seq + RB_MSG_HEADER_SIZE, data, to_copy);
        
        atomic_store_explicit(&rb->read_seq, read_seq + total_len, memory_order_release);
        
        *bytes_read = to_copy;
        return true;
    }

    while (read_seq < write_seq) {
        int spin = 0;
        while (!rb_is_msg_committed(rb, read_seq) && spin < 100) {
            spin++;
            for (volatile int i = 0; i < 10; i++);
        }

        if (!rb_is_msg_committed(rb, read_seq)) {
            break;
        }

        uint32_t header[2];
        rb_read_at(rb, read_seq, header, sizeof(header));
        
        uint32_t magic = header[0];
        uint32_t len = header[1];

        if (magic != RB_MSG_COMMITTED) {
            break;
        }

        if (len == 0 || len > 65536) {
            break;
        }

        size_t to_copy = RB_MIN(len, max_len);
        rb_read_at(rb, read_seq + RB_MSG_HEADER_SIZE, data, to_copy);
        
        size_t total_len = RB_ALIGN_UP(RB_MSG_HEADER_SIZE + len);
        atomic_store_explicit(&rb->read_seq, read_seq + total_len, memory_order_release);
        atomic_store_explicit(&rb->commit_seq, read_seq + total_len, memory_order_release);
        
        *bytes_read = to_copy;
        return true;
    }

    *bytes_read = 0;
    return false;
}

static void rb_advance_commit(RingBuffer* rb) {
    size_t read_seq = atomic_load_explicit(&rb->read_seq, memory_order_relaxed);
    size_t commit_seq = atomic_load_explicit(&rb->commit_seq, memory_order_relaxed);
    size_t write_seq = atomic_load_explicit(&rb->write_seq, memory_order_acquire);

    while (commit_seq < write_seq) {
        if (!rb_is_msg_committed(rb, commit_seq)) {
            break;
        }

        uint32_t header[2];
        rb_read_at(rb, commit_seq, header, sizeof(header));
        
        if (header[0] != RB_MSG_COMMITTED) {
            break;
        }

        uint32_t len = header[1];
        if (len == 0 || len > 65536) {
            break;
        }

        size_t total_len = RB_ALIGN_UP(RB_MSG_HEADER_SIZE + len);
        commit_seq += total_len;
    }

    atomic_store_explicit(&rb->commit_seq, commit_seq, memory_order_release);
}

bool ring_buffer_read(RingBuffer* rb, void* data, size_t max_len, size_t* bytes_read) {
    if (!rb || !data || !bytes_read || max_len == 0) {
        return false;
    }

    *bytes_read = 0;

    rb_advance_commit(rb);

    size_t read_seq = atomic_load_explicit(&rb->read_seq, memory_order_relaxed);
    size_t commit_seq = atomic_load_explicit(&rb->commit_seq, memory_order_acquire);

    if (read_seq >= commit_seq) {
        return false;
    }

    return rb_try_read_msg(rb, data, max_len, bytes_read);
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
