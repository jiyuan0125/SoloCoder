#ifndef MESSAGE_FRAME_H
#define MESSAGE_FRAME_H

#include <stdint.h>
#include <stddef.h>

#define MESSAGE_HEADER_SIZE 4

typedef struct {
    uint8_t *buffer;
    size_t buffer_size;
    size_t data_len;
    size_t expected_len;
} message_buffer_t;

int message_buffer_init(message_buffer_t *mb, size_t initial_size);
void message_buffer_destroy(message_buffer_t *mb);
int message_buffer_append(message_buffer_t *mb, const uint8_t *data, size_t len);
int message_buffer_has_complete_message(const message_buffer_t *mb);
uint32_t message_buffer_get_message_length(const message_buffer_t *mb);
int message_buffer_extract_message(message_buffer_t *mb, uint8_t **msg_data, size_t *msg_len);
void message_buffer_consume(message_buffer_t *mb, size_t len);

uint32_t message_frame_encode_length(uint32_t len);
uint32_t message_frame_decode_length(const uint8_t *header);
size_t message_frame_calculate_total_size(size_t msg_len);
int message_frame_pack(uint8_t *dest, size_t dest_len, const uint8_t *msg_data, size_t msg_len);

#endif
