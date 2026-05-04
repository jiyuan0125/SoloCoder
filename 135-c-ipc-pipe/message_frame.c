#include "message_frame.h"
#include <stdlib.h>
#include <string.h>
#include <arpa/inet.h>

int message_buffer_init(message_buffer_t *mb, size_t initial_size) {
    if (initial_size < MESSAGE_HEADER_SIZE) {
        initial_size = MESSAGE_HEADER_SIZE;
    }
    
    mb->buffer = (uint8_t *)malloc(initial_size);
    if (mb->buffer == NULL) {
        return -1;
    }
    
    mb->buffer_size = initial_size;
    mb->data_len = 0;
    mb->expected_len = 0;
    return 0;
}

void message_buffer_destroy(message_buffer_t *mb) {
    if (mb->buffer != NULL) {
        free(mb->buffer);
        mb->buffer = NULL;
    }
    mb->buffer_size = 0;
    mb->data_len = 0;
    mb->expected_len = 0;
}

static int message_buffer_resize(message_buffer_t *mb, size_t new_size) {
    if (new_size <= mb->buffer_size) {
        return 0;
    }
    
    uint8_t *new_buf = (uint8_t *)realloc(mb->buffer, new_size);
    if (new_buf == NULL) {
        return -1;
    }
    
    mb->buffer = new_buf;
    mb->buffer_size = new_size;
    return 0;
}

int message_buffer_append(message_buffer_t *mb, const uint8_t *data, size_t len) {
    size_t needed = mb->data_len + len;
    
    if (message_buffer_resize(mb, needed) != 0) {
        return -1;
    }
    
    memcpy(mb->buffer + mb->data_len, data, len);
    mb->data_len += len;
    
    if (mb->expected_len == 0 && mb->data_len >= MESSAGE_HEADER_SIZE) {
        mb->expected_len = MESSAGE_HEADER_SIZE + message_frame_decode_length(mb->buffer);
    }
    
    return 0;
}

int message_buffer_has_complete_message(const message_buffer_t *mb) {
    if (mb->data_len < MESSAGE_HEADER_SIZE) {
        return 0;
    }
    
    size_t total_msg_len = MESSAGE_HEADER_SIZE + message_frame_decode_length(mb->buffer);
    return (mb->data_len >= total_msg_len) ? 1 : 0;
}

uint32_t message_buffer_get_message_length(const message_buffer_t *mb) {
    if (mb->data_len < MESSAGE_HEADER_SIZE) {
        return 0;
    }
    return message_frame_decode_length(mb->buffer);
}

int message_buffer_extract_message(message_buffer_t *mb, uint8_t **msg_data, size_t *msg_len) {
    if (!message_buffer_has_complete_message(mb)) {
        return -1;
    }
    
    uint32_t body_len = message_frame_decode_length(mb->buffer);
    size_t total_len = MESSAGE_HEADER_SIZE + body_len;
    
    *msg_data = (uint8_t *)malloc(body_len);
    if (*msg_data == NULL) {
        return -1;
    }
    
    memcpy(*msg_data, mb->buffer + MESSAGE_HEADER_SIZE, body_len);
    *msg_len = body_len;
    
    message_buffer_consume(mb, total_len);
    
    return 0;
}

void message_buffer_consume(message_buffer_t *mb, size_t len) {
    if (len == 0 || len > mb->data_len) {
        return;
    }
    
    size_t remaining = mb->data_len - len;
    if (remaining > 0) {
        memmove(mb->buffer, mb->buffer + len, remaining);
    }
    mb->data_len = remaining;
    
    if (mb->data_len >= MESSAGE_HEADER_SIZE) {
        mb->expected_len = MESSAGE_HEADER_SIZE + message_frame_decode_length(mb->buffer);
    } else {
        mb->expected_len = 0;
    }
}

uint32_t message_frame_encode_length(uint32_t len) {
    return htonl(len);
}

uint32_t message_frame_decode_length(const uint8_t *header) {
    uint32_t len;
    memcpy(&len, header, sizeof(len));
    return ntohl(len);
}

size_t message_frame_calculate_total_size(size_t msg_len) {
    return MESSAGE_HEADER_SIZE + msg_len;
}

int message_frame_pack(uint8_t *dest, size_t dest_len, const uint8_t *msg_data, size_t msg_len) {
    size_t total_needed = message_frame_calculate_total_size(msg_len);
    
    if (dest_len < total_needed) {
        return -1;
    }
    
    uint32_t encoded_len = message_frame_encode_length((uint32_t)msg_len);
    memcpy(dest, &encoded_len, MESSAGE_HEADER_SIZE);
    memcpy(dest + MESSAGE_HEADER_SIZE, msg_data, msg_len);
    
    return (int)total_needed;
}
