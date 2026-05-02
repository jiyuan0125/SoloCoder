#ifndef PROTO_H
#define PROTO_H

#include <stdint.h>
#include <stddef.h>

#define VARINT_MAX_BYTES 5
#define CRC32_SIZE 4

typedef enum {
    WIRE_TYPE_VARINT = 0,
    WIRE_TYPE_LENGTH_DELIMITED = 1,
    WIRE_TYPE_FIXED32 = 2,
    WIRE_TYPE_FIXED64 = 5
} wire_type_t;

typedef enum {
    PROTO_OK = 0,
    PROTO_ERROR_BUFFER_TOO_SMALL = -1,
    PROTO_ERROR_CRC_MISMATCH = -2,
    PROTO_ERROR_INVALID_FORMAT = -3,
    PROTO_ERROR_VARINT_TOO_LONG = -4,
    PROTO_ERROR_UNKNOWN_WIRE_TYPE = -5
} proto_error_t;

typedef struct proto_field proto_field_t;
typedef struct proto_msg proto_msg_t;

struct proto_msg {
    proto_field_t *fields;
    size_t count;
    size_t capacity;
};

struct proto_field {
    uint32_t tag;
    wire_type_t type;
    union {
        int32_t varint;
        uint32_t fixed32;
        uint64_t fixed64;
        struct {
            uint8_t *data;
            size_t len;
        } bytes;
        proto_msg_t *nested;
    } value;
    uint8_t is_nested;
};

proto_msg_t *proto_msg_create(size_t capacity);
void proto_msg_destroy(proto_msg_t *msg);
int proto_msg_add_varint(proto_msg_t *msg, uint32_t tag, int32_t value);
int proto_msg_add_fixed32(proto_msg_t *msg, uint32_t tag, uint32_t value);
int proto_msg_add_fixed64(proto_msg_t *msg, uint32_t tag, uint64_t value);
int proto_msg_add_bytes(proto_msg_t *msg, uint32_t tag, const uint8_t *data, size_t len);
int proto_msg_add_nested(proto_msg_t *msg, uint32_t tag, const proto_msg_t *nested);

int varint_encode(int32_t value, uint8_t *buf, size_t buf_size);
int varint_decode(const uint8_t *buf, size_t buf_size, int32_t *value);

uint32_t crc32_ieee(const uint8_t *data, size_t len);

int proto_encode(const proto_msg_t *msg, uint8_t *buf, size_t buf_size);
int proto_decode(const uint8_t *buf, size_t len, proto_msg_t *msg);

#endif
