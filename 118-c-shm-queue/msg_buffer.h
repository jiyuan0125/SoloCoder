#ifndef MSG_BUFFER_H
#define MSG_BUFFER_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

#define MSG_BUFFER_MAGIC     0x51554553
#define MSG_BUFFER_VERSION   1

typedef struct msg_buffer_header {
    uint32_t magic;
    uint32_t version;
    uint32_t max_msg_len;
    uint32_t buffer_size;
    volatile uint32_t read_idx;
    volatile uint32_t write_idx;
    uint32_t reserved[8];
} msg_buffer_header_t;

typedef struct msg_buffer {
    msg_buffer_header_t *header;
    uint8_t *data;
    size_t total_size;
} msg_buffer_t;

size_t msg_buffer_calculate_size(size_t buffer_size);

int msg_buffer_init(msg_buffer_t *buf, void *shm_addr, size_t shm_size,
                    size_t max_msg_len, size_t buffer_size);
int msg_buffer_attach(msg_buffer_t *buf, void *shm_addr, size_t shm_size);
void msg_buffer_detach(msg_buffer_t *buf);

bool msg_buffer_can_write(msg_buffer_t *buf, size_t msg_len);
bool msg_buffer_can_read(msg_buffer_t *buf);
size_t msg_buffer_available_space(msg_buffer_t *buf);
size_t msg_buffer_available_data(msg_buffer_t *buf);

int msg_buffer_write(msg_buffer_t *buf, const void *data, size_t len);
int msg_buffer_read(msg_buffer_t *buf, void *data, size_t *len);
int msg_buffer_peek(msg_buffer_t *buf, void *data, size_t *len);

bool msg_buffer_is_valid(msg_buffer_t *buf);

#ifdef __cplusplus
}
#endif

#endif
