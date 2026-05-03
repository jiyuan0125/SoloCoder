#include "msg_buffer.h"
#include <string.h>
#include <stdio.h>

size_t msg_buffer_calculate_size(size_t buffer_size) {
    return sizeof(msg_buffer_header_t) + buffer_size;
}

static bool buf_is_empty(msg_buffer_header_t *hdr) {
    return hdr->read_idx == hdr->write_idx;
}

static size_t buf_used_space(msg_buffer_header_t *hdr) {
    if (hdr->write_idx >= hdr->read_idx) {
        return hdr->write_idx - hdr->read_idx;
    } else {
        return hdr->buffer_size - hdr->read_idx + hdr->write_idx;
    }
}

static size_t buf_free_space(msg_buffer_header_t *hdr) {
    return hdr->buffer_size - 1 - buf_used_space(hdr);
}

int msg_buffer_init(msg_buffer_t *buf, void *shm_addr, size_t shm_size,
                    size_t max_msg_len, size_t buffer_size) {
    memset(buf, 0, sizeof(msg_buffer_t));
    
    if (shm_size < msg_buffer_calculate_size(buffer_size)) {
        return -1;
    }
    
    buf->header = (msg_buffer_header_t *)shm_addr;
    buf->data = (uint8_t *)shm_addr + sizeof(msg_buffer_header_t);
    buf->total_size = shm_size;
    
    memset(buf->header, 0, sizeof(msg_buffer_header_t));
    buf->header->magic = MSG_BUFFER_MAGIC;
    buf->header->version = MSG_BUFFER_VERSION;
    buf->header->max_msg_len = (uint32_t)max_msg_len;
    buf->header->buffer_size = (uint32_t)buffer_size;
    buf->header->read_idx = 0;
    buf->header->write_idx = 0;
    
    return 0;
}

int msg_buffer_attach(msg_buffer_t *buf, void *shm_addr, size_t shm_size) {
    memset(buf, 0, sizeof(msg_buffer_t));
    
    if (shm_size < sizeof(msg_buffer_header_t)) {
        return -1;
    }
    
    buf->header = (msg_buffer_header_t *)shm_addr;
    
    if (buf->header->magic != MSG_BUFFER_MAGIC) {
        return -1;
    }
    
    if (shm_size < msg_buffer_calculate_size(buf->header->buffer_size)) {
        return -1;
    }
    
    buf->data = (uint8_t *)shm_addr + sizeof(msg_buffer_header_t);
    buf->total_size = shm_size;
    
    return 0;
}

void msg_buffer_detach(msg_buffer_t *buf) {
    memset(buf, 0, sizeof(msg_buffer_t));
}

bool msg_buffer_is_valid(msg_buffer_t *buf) {
    if (buf == NULL || buf->header == NULL) {
        return false;
    }
    return buf->header->magic == MSG_BUFFER_MAGIC;
}

bool msg_buffer_can_write(msg_buffer_t *buf, size_t msg_len) {
    if (!msg_buffer_is_valid(buf)) {
        return false;
    }
    
    size_t total_len = 4 + msg_len;
    
    if (msg_len > buf->header->max_msg_len) {
        return false;
    }
    
    size_t free_space = buf_free_space(buf->header);
    return free_space >= total_len;
}

bool msg_buffer_can_read(msg_buffer_t *buf) {
    if (!msg_buffer_is_valid(buf)) {
        return false;
    }
    return !buf_is_empty(buf->header);
}

size_t msg_buffer_available_space(msg_buffer_t *buf) {
    if (!msg_buffer_is_valid(buf)) {
        return 0;
    }
    size_t free_space = buf_free_space(buf->header);
    if (free_space > 4) {
        return free_space - 4;
    }
    return 0;
}

size_t msg_buffer_available_data(msg_buffer_t *buf) {
    if (!msg_buffer_is_valid(buf)) {
        return 0;
    }
    return buf_used_space(buf->header);
}

int msg_buffer_write(msg_buffer_t *buf, const void *data, size_t len) {
    if (!msg_buffer_is_valid(buf)) {
        return -1;
    }
    
    if (len > buf->header->max_msg_len) {
        return -1;
    }
    
    if (!msg_buffer_can_write(buf, len)) {
        return -1;
    }
    
    msg_buffer_header_t *hdr = buf->header;
    uint8_t *dst = buf->data;
    const uint8_t *src = (const uint8_t *)data;
    uint32_t write_idx = hdr->write_idx;
    
    uint32_t len_be = __builtin_bswap32((uint32_t)len);
    
    if (write_idx + 4 <= hdr->buffer_size) {
        memcpy(dst + write_idx, &len_be, 4);
    } else {
        size_t first_part = hdr->buffer_size - write_idx;
        memcpy(dst + write_idx, &len_be, first_part);
        memcpy(dst, ((uint8_t *)&len_be) + first_part, 4 - first_part);
    }
    
    write_idx = (write_idx + 4) % hdr->buffer_size;
    
    if (write_idx + len <= hdr->buffer_size) {
        memcpy(dst + write_idx, src, len);
    } else {
        size_t first_part = hdr->buffer_size - write_idx;
        memcpy(dst + write_idx, src, first_part);
        memcpy(dst, src + first_part, len - first_part);
    }
    
    hdr->write_idx = (write_idx + (uint32_t)len) % hdr->buffer_size;
    
    return 0;
}

static int read_message_len(msg_buffer_t *buf, uint32_t *out_len) {
    msg_buffer_header_t *hdr = buf->header;
    uint8_t *src = buf->data;
    uint32_t read_idx = hdr->read_idx;
    uint32_t len_be;
    
    if (read_idx + 4 <= hdr->buffer_size) {
        memcpy(&len_be, src + read_idx, 4);
    } else {
        size_t first_part = hdr->buffer_size - read_idx;
        uint8_t *len_bytes = (uint8_t *)&len_be;
        memcpy(len_bytes, src + read_idx, first_part);
        memcpy(len_bytes + first_part, src, 4 - first_part);
    }
    
    *out_len = __builtin_bswap32(len_be);
    
    if (*out_len > hdr->max_msg_len) {
        return -1;
    }
    
    return 0;
}

int msg_buffer_read(msg_buffer_t *buf, void *data, size_t *len) {
    if (!msg_buffer_is_valid(buf)) {
        return -1;
    }
    
    if (buf_is_empty(buf->header)) {
        return -1;
    }
    
    msg_buffer_header_t *hdr = buf->header;
    uint8_t *src = buf->data;
    uint8_t *dst = (uint8_t *)data;
    uint32_t read_idx = hdr->read_idx;
    uint32_t msg_len;
    
    if (read_message_len(buf, &msg_len) < 0) {
        return -1;
    }
    
    read_idx = (read_idx + 4) % hdr->buffer_size;
    
    if (read_idx + msg_len <= hdr->buffer_size) {
        memcpy(dst, src + read_idx, msg_len);
    } else {
        size_t first_part = hdr->buffer_size - read_idx;
        memcpy(dst, src + read_idx, first_part);
        memcpy(dst + first_part, src, msg_len - first_part);
    }
    
    hdr->read_idx = (read_idx + msg_len) % hdr->buffer_size;
    *len = msg_len;
    
    return 0;
}

int msg_buffer_peek(msg_buffer_t *buf, void *data, size_t *len) {
    if (!msg_buffer_is_valid(buf)) {
        return -1;
    }
    
    if (buf_is_empty(buf->header)) {
        return -1;
    }
    
    msg_buffer_header_t *hdr = buf->header;
    uint8_t *src = buf->data;
    uint8_t *dst = (uint8_t *)data;
    uint32_t read_idx = hdr->read_idx;
    uint32_t msg_len;
    
    if (read_message_len(buf, &msg_len) < 0) {
        return -1;
    }
    
    read_idx = (read_idx + 4) % hdr->buffer_size;
    
    if (read_idx + msg_len <= hdr->buffer_size) {
        memcpy(dst, src + read_idx, msg_len);
    } else {
        size_t first_part = hdr->buffer_size - read_idx;
        memcpy(dst, src + read_idx, first_part);
        memcpy(dst + first_part, src, msg_len - first_part);
    }
    
    *len = msg_len;
    
    return 0;
}
