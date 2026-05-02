#include "proto.h"
#include <stdlib.h>
#include <string.h>

static void write_u32_be(uint8_t *buf, uint32_t value) {
    buf[0] = (uint8_t)((value >> 24) & 0xFF);
    buf[1] = (uint8_t)((value >> 16) & 0xFF);
    buf[2] = (uint8_t)((value >> 8) & 0xFF);
    buf[3] = (uint8_t)(value & 0xFF);
}

static uint32_t read_u32_be(const uint8_t *buf) {
    return (uint32_t)buf[0] << 24 |
           (uint32_t)buf[1] << 16 |
           (uint32_t)buf[2] << 8 |
           (uint32_t)buf[3];
}

static void write_u64_be(uint8_t *buf, uint64_t value) {
    buf[0] = (uint8_t)((value >> 56) & 0xFF);
    buf[1] = (uint8_t)((value >> 48) & 0xFF);
    buf[2] = (uint8_t)((value >> 40) & 0xFF);
    buf[3] = (uint8_t)((value >> 32) & 0xFF);
    buf[4] = (uint8_t)((value >> 24) & 0xFF);
    buf[5] = (uint8_t)((value >> 16) & 0xFF);
    buf[6] = (uint8_t)((value >> 8) & 0xFF);
    buf[7] = (uint8_t)(value & 0xFF);
}

static uint64_t read_u64_be(const uint8_t *buf) {
    return (uint64_t)buf[0] << 56 |
           (uint64_t)buf[1] << 48 |
           (uint64_t)buf[2] << 40 |
           (uint64_t)buf[3] << 32 |
           (uint64_t)buf[4] << 24 |
           (uint64_t)buf[5] << 16 |
           (uint64_t)buf[6] << 8 |
           (uint64_t)buf[7];
}

proto_msg_t *proto_msg_create(size_t capacity) {
    if (capacity == 0) {
        capacity = 8;
    }

    proto_msg_t *msg = (proto_msg_t *)malloc(sizeof(proto_msg_t));
    if (!msg) {
        return NULL;
    }

    msg->fields = (proto_field_t *)malloc(capacity * sizeof(proto_field_t));
    if (!msg->fields) {
        free(msg);
        return NULL;
    }

    msg->count = 0;
    msg->capacity = capacity;
    return msg;
}

void proto_msg_destroy(proto_msg_t *msg) {
    if (!msg) {
        return;
    }

    for (size_t i = 0; i < msg->count; i++) {
        proto_field_t *field = &msg->fields[i];
        if (field->is_nested && field->value.nested) {
            proto_msg_destroy(field->value.nested);
        } else if (field->type == WIRE_TYPE_LENGTH_DELIMITED && field->value.bytes.data) {
            free(field->value.bytes.data);
        }
    }

    free(msg->fields);
    free(msg);
}

static int proto_msg_ensure_capacity(proto_msg_t *msg) {
    if (msg->count >= msg->capacity) {
        size_t new_capacity = msg->capacity * 2;
        proto_field_t *new_fields = (proto_field_t *)realloc(
            msg->fields, new_capacity * sizeof(proto_field_t));
        if (!new_fields) {
            return PROTO_ERROR_BUFFER_TOO_SMALL;
        }
        msg->fields = new_fields;
        msg->capacity = new_capacity;
    }
    return PROTO_OK;
}

int proto_msg_add_varint(proto_msg_t *msg, uint32_t tag, int32_t value) {
    if (!msg) {
        return PROTO_ERROR_INVALID_FORMAT;
    }

    int ret = proto_msg_ensure_capacity(msg);
    if (ret != PROTO_OK) {
        return ret;
    }

    proto_field_t *field = &msg->fields[msg->count++];
    field->tag = tag;
    field->type = WIRE_TYPE_VARINT;
    field->value.varint = value;
    field->is_nested = 0;

    return PROTO_OK;
}

int proto_msg_add_fixed32(proto_msg_t *msg, uint32_t tag, uint32_t value) {
    if (!msg) {
        return PROTO_ERROR_INVALID_FORMAT;
    }

    int ret = proto_msg_ensure_capacity(msg);
    if (ret != PROTO_OK) {
        return ret;
    }

    proto_field_t *field = &msg->fields[msg->count++];
    field->tag = tag;
    field->type = WIRE_TYPE_FIXED32;
    field->value.fixed32 = value;
    field->is_nested = 0;

    return PROTO_OK;
}

int proto_msg_add_fixed64(proto_msg_t *msg, uint32_t tag, uint64_t value) {
    if (!msg) {
        return PROTO_ERROR_INVALID_FORMAT;
    }

    int ret = proto_msg_ensure_capacity(msg);
    if (ret != PROTO_OK) {
        return ret;
    }

    proto_field_t *field = &msg->fields[msg->count++];
    field->tag = tag;
    field->type = WIRE_TYPE_FIXED64;
    field->value.fixed64 = value;
    field->is_nested = 0;

    return PROTO_OK;
}

int proto_msg_add_bytes(proto_msg_t *msg, uint32_t tag, const uint8_t *data, size_t len) {
    if (!msg || !data) {
        return PROTO_ERROR_INVALID_FORMAT;
    }

    int ret = proto_msg_ensure_capacity(msg);
    if (ret != PROTO_OK) {
        return ret;
    }

    uint8_t *copy = (uint8_t *)malloc(len);
    if (!copy) {
        return PROTO_ERROR_BUFFER_TOO_SMALL;
    }
    memcpy(copy, data, len);

    proto_field_t *field = &msg->fields[msg->count++];
    field->tag = tag;
    field->type = WIRE_TYPE_LENGTH_DELIMITED;
    field->value.bytes.data = copy;
    field->value.bytes.len = len;
    field->is_nested = 0;

    return PROTO_OK;
}

static proto_msg_t *proto_msg_clone(const proto_msg_t *src);

static proto_msg_t *proto_msg_clone(const proto_msg_t *src) {
    if (!src) {
        return NULL;
    }

    proto_msg_t *dest = proto_msg_create(src->count > 0 ? src->count : 8);
    if (!dest) {
        return NULL;
    }

    for (size_t i = 0; i < src->count; i++) {
        const proto_field_t *src_field = &src->fields[i];
        int ret = proto_msg_ensure_capacity(dest);
        if (ret != PROTO_OK) {
            proto_msg_destroy(dest);
            return NULL;
        }

        proto_field_t *dest_field = &dest->fields[dest->count];
        dest_field->tag = src_field->tag;
        dest_field->type = src_field->type;
        dest_field->is_nested = src_field->is_nested;

        if (src_field->is_nested) {
            dest_field->value.nested = proto_msg_clone(src_field->value.nested);
            if (!dest_field->value.nested) {
                proto_msg_destroy(dest);
                return NULL;
            }
        } else if (src_field->type == WIRE_TYPE_LENGTH_DELIMITED) {
            dest_field->value.bytes.data = (uint8_t *)malloc(src_field->value.bytes.len);
            if (!dest_field->value.bytes.data) {
                proto_msg_destroy(dest);
                return NULL;
            }
            memcpy(dest_field->value.bytes.data, src_field->value.bytes.data,
                   src_field->value.bytes.len);
            dest_field->value.bytes.len = src_field->value.bytes.len;
        } else if (src_field->type == WIRE_TYPE_VARINT) {
            dest_field->value.varint = src_field->value.varint;
        } else if (src_field->type == WIRE_TYPE_FIXED32) {
            dest_field->value.fixed32 = src_field->value.fixed32;
        } else if (src_field->type == WIRE_TYPE_FIXED64) {
            dest_field->value.fixed64 = src_field->value.fixed64;
        }

        dest->count++;
    }

    return dest;
}

int proto_msg_add_nested(proto_msg_t *msg, uint32_t tag, const proto_msg_t *nested) {
    if (!msg || !nested) {
        return PROTO_ERROR_INVALID_FORMAT;
    }

    int ret = proto_msg_ensure_capacity(msg);
    if (ret != PROTO_OK) {
        return ret;
    }

    proto_msg_t *clone = proto_msg_clone(nested);
    if (!clone) {
        return PROTO_ERROR_BUFFER_TOO_SMALL;
    }

    proto_field_t *field = &msg->fields[msg->count];
    field->tag = tag;
    field->type = WIRE_TYPE_LENGTH_DELIMITED;
    field->is_nested = 1;
    field->value.nested = clone;

    msg->count++;
    return PROTO_OK;
}

static int calculate_field_size(const proto_field_t *field) {
    int size = 0;

    int32_t key = (int32_t)((field->tag << 3) | field->type);
    uint8_t temp_buf[VARINT_MAX_BYTES];
    int key_size = varint_encode(key, temp_buf, sizeof(temp_buf));
    if (key_size < 0) {
        return key_size;
    }
    size += key_size;

    switch (field->type) {
        case WIRE_TYPE_VARINT: {
            int val_size = varint_encode(field->value.varint, temp_buf, sizeof(temp_buf));
            if (val_size < 0) {
                return val_size;
            }
            size += val_size;
            break;
        }
        case WIRE_TYPE_FIXED32:
            size += 4;
            break;
        case WIRE_TYPE_FIXED64:
            size += 8;
            break;
        case WIRE_TYPE_LENGTH_DELIMITED:
            size += 4;
            if (field->is_nested && field->value.nested) {
                for (size_t i = 0; i < field->value.nested->count; i++) {
                    int nested_size = calculate_field_size(&field->value.nested->fields[i]);
                    if (nested_size < 0) {
                        return nested_size;
                    }
                    size += nested_size;
                }
            } else {
                size += (int)field->value.bytes.len;
            }
            break;
        default:
            return PROTO_ERROR_UNKNOWN_WIRE_TYPE;
    }

    return size;
}

static int encode_field(const proto_field_t *field, uint8_t *buf, size_t buf_size, size_t *offset) {
    int32_t key = (int32_t)((field->tag << 3) | field->type);
    int key_size = varint_encode(key, buf + *offset, buf_size - *offset);
    if (key_size < 0) {
        return key_size;
    }
    *offset += (size_t)key_size;

    switch (field->type) {
        case WIRE_TYPE_VARINT: {
            int val_size = varint_encode(field->value.varint, buf + *offset, buf_size - *offset);
            if (val_size < 0) {
                return val_size;
            }
            *offset += (size_t)val_size;
            break;
        }
        case WIRE_TYPE_FIXED32:
            if (*offset + 4 > buf_size) {
                return PROTO_ERROR_BUFFER_TOO_SMALL;
            }
            write_u32_be(buf + *offset, field->value.fixed32);
            *offset += 4;
            break;
        case WIRE_TYPE_FIXED64:
            if (*offset + 8 > buf_size) {
                return PROTO_ERROR_BUFFER_TOO_SMALL;
            }
            write_u64_be(buf + *offset, field->value.fixed64);
            *offset += 8;
            break;
        case WIRE_TYPE_LENGTH_DELIMITED: {
            size_t len_offset = *offset;
            *offset += 4;

            if (field->is_nested && field->value.nested) {
                for (size_t i = 0; i < field->value.nested->count; i++) {
                    int ret = encode_field(&field->value.nested->fields[i], buf, buf_size, offset);
                    if (ret < 0) {
                        return ret;
                    }
                }
            } else {
                if (*offset + field->value.bytes.len > buf_size) {
                    return PROTO_ERROR_BUFFER_TOO_SMALL;
                }
                memcpy(buf + *offset, field->value.bytes.data, field->value.bytes.len);
                *offset += field->value.bytes.len;
            }

            uint32_t data_len = (uint32_t)(*offset - len_offset - 4);
            write_u32_be(buf + len_offset, data_len);
            break;
        }
        default:
            return PROTO_ERROR_UNKNOWN_WIRE_TYPE;
    }

    return PROTO_OK;
}

int proto_encode(const proto_msg_t *msg, uint8_t *buf, size_t buf_size) {
    if (!msg || !buf) {
        return PROTO_ERROR_INVALID_FORMAT;
    }

    int total_size = 0;
    for (size_t i = 0; i < msg->count; i++) {
        int field_size = calculate_field_size(&msg->fields[i]);
        if (field_size < 0) {
            return field_size;
        }
        total_size += field_size;
    }

    if ((size_t)total_size + CRC32_SIZE > buf_size) {
        return PROTO_ERROR_BUFFER_TOO_SMALL;
    }

    size_t offset = 0;
    for (size_t i = 0; i < msg->count; i++) {
        int ret = encode_field(&msg->fields[i], buf, buf_size, &offset);
        if (ret < 0) {
            return ret;
        }
    }

    uint32_t crc = crc32_ieee(buf, offset);
    write_u32_be(buf + offset, crc);
    offset += CRC32_SIZE;

    return (int)offset;
}

static int decode_field(const uint8_t *buf, size_t buf_len, size_t *offset, proto_msg_t *msg);

static int decode_length_delimited(const uint8_t *buf, size_t buf_len, size_t *offset,
                                    proto_field_t *field) {
    if (*offset + 4 > buf_len) {
        return PROTO_ERROR_INVALID_FORMAT;
    }

    uint32_t data_len = read_u32_be(buf + *offset);
    *offset += 4;

    if (*offset + data_len > buf_len) {
        return PROTO_ERROR_INVALID_FORMAT;
    }

    field->value.bytes.data = (uint8_t *)malloc(data_len);
    if (!field->value.bytes.data) {
        return PROTO_ERROR_BUFFER_TOO_SMALL;
    }
    memcpy(field->value.bytes.data, buf + *offset, data_len);
    field->value.bytes.len = data_len;
    *offset += data_len;

    return PROTO_OK;
}

static int decode_field(const uint8_t *buf, size_t buf_len, size_t *offset, proto_msg_t *msg) {
    int ret = proto_msg_ensure_capacity(msg);
    if (ret != PROTO_OK) {
        return ret;
    }

    proto_field_t *field = &msg->fields[msg->count];

    int32_t key;
    int key_size = varint_decode(buf + *offset, buf_len - *offset, &key);
    if (key_size < 0) {
        return key_size;
    }
    *offset += (size_t)key_size;

    field->tag = (uint32_t)key >> 3;
    field->type = (wire_type_t)(key & 0x07);
    field->is_nested = 0;
    field->value.nested = NULL;

    switch (field->type) {
        case WIRE_TYPE_VARINT: {
            int32_t value;
            int val_size = varint_decode(buf + *offset, buf_len - *offset, &value);
            if (val_size < 0) {
                return val_size;
            }
            field->value.varint = value;
            *offset += (size_t)val_size;
            break;
        }
        case WIRE_TYPE_FIXED32:
            if (*offset + 4 > buf_len) {
                return PROTO_ERROR_INVALID_FORMAT;
            }
            field->value.fixed32 = read_u32_be(buf + *offset);
            *offset += 4;
            break;
        case WIRE_TYPE_FIXED64:
            if (*offset + 8 > buf_len) {
                return PROTO_ERROR_INVALID_FORMAT;
            }
            field->value.fixed64 = read_u64_be(buf + *offset);
            *offset += 8;
            break;
        case WIRE_TYPE_LENGTH_DELIMITED:
            ret = decode_length_delimited(buf, buf_len, offset, field);
            if (ret < 0) {
                return ret;
            }
            break;
        default:
            return PROTO_ERROR_UNKNOWN_WIRE_TYPE;
    }

    msg->count++;
    return PROTO_OK;
}

int proto_decode(const uint8_t *buf, size_t len, proto_msg_t *msg) {
    if (!buf || len < CRC32_SIZE || !msg) {
        return PROTO_ERROR_INVALID_FORMAT;
    }

    size_t data_len = len - CRC32_SIZE;
    uint32_t expected_crc = read_u32_be(buf + data_len);
    uint32_t actual_crc = crc32_ieee(buf, data_len);

    if (expected_crc != actual_crc) {
        return PROTO_ERROR_CRC_MISMATCH;
    }

    msg->count = 0;
    size_t offset = 0;

    while (offset < data_len) {
        int ret = decode_field(buf, data_len, &offset, msg);
        if (ret < 0) {
            return ret;
        }
    }

    return (int)len;
}
