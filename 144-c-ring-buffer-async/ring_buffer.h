#ifndef RING_BUFFER_H
#define RING_BUFFER_H

#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>
#include <stdatomic.h>

typedef enum {
    RB_POLICY_DROP,
    RB_POLICY_BLOCK
} RingBufferPolicy;

typedef struct {
    size_t buffer_size;
    RingBufferPolicy policy;
} RingBufferConfig;

typedef struct {
    uint8_t* buffer;
    size_t buffer_size;
    size_t buffer_mask;
    
    _Atomic(size_t) write_seq;
    _Atomic(size_t) read_seq;
    
    _Atomic(uint32_t) write_lock;
    
    RingBufferPolicy policy;
    _Atomic(bool) is_running;
    _Atomic(size_t) dropped_count;
} RingBuffer;

RingBuffer* ring_buffer_create(const RingBufferConfig* config);
void ring_buffer_destroy(RingBuffer* rb);

bool ring_buffer_write(RingBuffer* rb, const void* data, size_t len);
bool ring_buffer_read(RingBuffer* rb, void* data, size_t max_len, size_t* bytes_read);

size_t ring_buffer_available(const RingBuffer* rb);
size_t ring_buffer_used(const RingBuffer* rb);
size_t ring_buffer_capacity(const RingBuffer* rb);
bool ring_buffer_is_empty(const RingBuffer* rb);
bool ring_buffer_is_full(const RingBuffer* rb);
size_t ring_buffer_dropped(const RingBuffer* rb);

void ring_buffer_set_policy(RingBuffer* rb, RingBufferPolicy policy);

#endif
