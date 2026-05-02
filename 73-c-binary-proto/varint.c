#include "proto.h"
#include <stdint.h>

static uint32_t zigzag_encode(int32_t n) {
    return (uint32_t)((n << 1) ^ (n >> 31));
}

static int32_t zigzag_decode(uint32_t n) {
    return (int32_t)((n >> 1) ^ (-(n & 1)));
}

int varint_encode(int32_t value, uint8_t *buf, size_t buf_size) {
    uint32_t uvalue = zigzag_encode(value);
    size_t size = 0;

    do {
        if (size >= buf_size) {
            return PROTO_ERROR_BUFFER_TOO_SMALL;
        }
        if (size >= VARINT_MAX_BYTES) {
            return PROTO_ERROR_VARINT_TOO_LONG;
        }

        uint8_t byte = (uint8_t)(uvalue & 0x7F);
        uvalue >>= 7;

        if (uvalue != 0) {
            byte |= 0x80;
        }

        buf[size++] = byte;
    } while (uvalue != 0);

    return (int)size;
}

int varint_decode(const uint8_t *buf, size_t buf_size, int32_t *value) {
    uint32_t result = 0;
    size_t shift = 0;
    size_t i = 0;

    while (i < buf_size && i < VARINT_MAX_BYTES) {
        uint8_t byte = buf[i];
        result |= (uint32_t)(byte & 0x7F) << shift;
        shift += 7;
        i++;

        if ((byte & 0x80) == 0) {
            *value = zigzag_decode(result);
            return (int)i;
        }
    }

    if (i >= VARINT_MAX_BYTES) {
        return PROTO_ERROR_VARINT_TOO_LONG;
    }

    return PROTO_ERROR_INVALID_FORMAT;
}
