#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "proto.h"

static void test_varint(void) {
    printf("=== Testing varint ===\n");
    uint8_t buf[VARINT_MAX_BYTES];
    int32_t values[] = {0, 1, -1, 127, 128, -128, 16383, 16384, 2147483647, -2147483648};
    int num_values = sizeof(values) / sizeof(values[0]);

    for (int i = 0; i < num_values; i++) {
        int32_t original = values[i];
        int encoded = varint_encode(original, buf, sizeof(buf));
        if (encoded < 0) {
            printf("FAIL: varint_encode(%d) returned %d\n", original, encoded);
            continue;
        }

        int32_t decoded;
        int read = varint_decode(buf, (size_t)encoded, &decoded);
        if (read < 0) {
            printf("FAIL: varint_decode returned %d for value %d\n", read, original);
            continue;
        }

        if (original == decoded && encoded == read) {
            printf("PASS: %d -> %d bytes -> %d\n", original, encoded, decoded);
        } else {
            printf("FAIL: %d -> %d bytes -> %d (expected %d, encoded %d, read %d)\n",
                   original, encoded, decoded, original, encoded, read);
        }
    }
}

static void test_crc32(void) {
    printf("\n=== Testing CRC32 ===\n");

    const char *test1 = "123456789";
    uint32_t crc1 = crc32_ieee((const uint8_t *)test1, 9);
    printf("CRC32(\"123456789\") = 0x%08X (expected 0xCBF43926)\n", crc1);

    const char *test2 = "";
    uint32_t crc2 = crc32_ieee((const uint8_t *)test2, 0);
    printf("CRC32(\"\") = 0x%08X (expected 0x00000000)\n", crc2);

    const char *test3 = "The quick brown fox jumps over the lazy dog";
    uint32_t crc3 = crc32_ieee((const uint8_t *)test3, 43);
    printf("CRC32(\"The quick brown fox...\") = 0x%08X\n", crc3);
}

static void test_proto_simple(void) {
    printf("\n=== Testing proto encode/decode (simple) ===\n");

    proto_msg_t *msg = proto_msg_create(4);
    proto_msg_add_varint(msg, 1, 42);
    proto_msg_add_fixed32(msg, 2, 0xDEADBEEF);
    proto_msg_add_fixed64(msg, 3, 0x0123456789ABCDEFULL);

    uint8_t buf[256];
    int encoded = proto_encode(msg, buf, sizeof(buf));
    printf("Encoded size: %d bytes\n", encoded);

    proto_msg_t *decoded = proto_msg_create(4);
    int decoded_size = proto_decode(buf, (size_t)encoded, decoded);

    if (decoded_size >= 0) {
        printf("Decoded successfully, %d bytes\n", decoded_size);
        for (size_t i = 0; i < decoded->count; i++) {
            proto_field_t *f = &decoded->fields[i];
            printf("  Field tag=%u, type=%d: ", f->tag, f->type);
            switch (f->type) {
                case WIRE_TYPE_VARINT:
                    printf("varint = %d\n", f->value.varint);
                    break;
                case WIRE_TYPE_FIXED32:
                    printf("fixed32 = 0x%08X\n", f->value.fixed32);
                    break;
                case WIRE_TYPE_FIXED64:
                    printf("fixed64 = 0x%016llX\n", (unsigned long long)f->value.fixed64);
                    break;
                default:
                    printf("unknown\n");
                    break;
            }
        }
    } else {
        printf("Decode failed: %d\n", decoded_size);
    }

    proto_msg_destroy(msg);
    proto_msg_destroy(decoded);
}

static void test_proto_bytes(void) {
    printf("\n=== Testing proto encode/decode (bytes) ===\n");

    const char *str = "Hello, Binary Protocol!";

    proto_msg_t *msg = proto_msg_create(2);
    proto_msg_add_bytes(msg, 1, (const uint8_t *)str, strlen(str));

    uint8_t buf[256];
    int encoded = proto_encode(msg, buf, sizeof(buf));
    printf("Encoded size: %d bytes\n", encoded);

    proto_msg_t *decoded = proto_msg_create(2);
    int decoded_size = proto_decode(buf, (size_t)encoded, decoded);

    if (decoded_size >= 0 && decoded->count > 0) {
        proto_field_t *f = &decoded->fields[0];
        if (f->type == WIRE_TYPE_LENGTH_DELIMITED && !f->is_nested) {
            printf("Decoded bytes: len=%zu, data=\"%.*s\"\n",
                   f->value.bytes.len, (int)f->value.bytes.len,
                   (const char *)f->value.bytes.data);
        }
    } else {
        printf("Decode failed: %d\n", decoded_size);
    }

    proto_msg_destroy(msg);
    proto_msg_destroy(decoded);
}

static void test_proto_nested(void) {
    printf("\n=== Testing proto encode/decode (nested) ===\n");

    proto_msg_t *inner = proto_msg_create(2);
    proto_msg_add_varint(inner, 1, -100);
    proto_msg_add_fixed32(inner, 2, 0xCAFEBABE);

    proto_msg_t *outer = proto_msg_create(2);
    proto_msg_add_varint(outer, 1, 999);
    proto_msg_add_nested(outer, 2, inner);

    uint8_t buf[512];
    int encoded = proto_encode(outer, buf, sizeof(buf));
    printf("Encoded size (with nested): %d bytes\n", encoded);

    proto_msg_t *decoded = proto_msg_create(2);
    int decoded_size = proto_decode(buf, (size_t)encoded, decoded);

    if (decoded_size >= 0) {
        printf("Decoded successfully:\n");
        for (size_t i = 0; i < decoded->count; i++) {
            proto_field_t *f = &decoded->fields[i];
            printf("  Field tag=%u, type=%d: ", f->tag, f->type);
            if (f->type == WIRE_TYPE_VARINT) {
                printf("varint = %d\n", f->value.varint);
            } else if (f->type == WIRE_TYPE_LENGTH_DELIMITED) {
                printf("length-delimited (len=%zu)\n", f->value.bytes.len);
                printf("    Now manually decoding nested message...\n");
                proto_msg_t *nested_msg = proto_msg_create(4);
                size_t offset = 0;
                while (offset < f->value.bytes.len) {
                    int32_t key;
                    int key_size = varint_decode(f->value.bytes.data + offset,
                                                  f->value.bytes.len - offset, &key);
                    if (key_size < 0) break;
                    offset += (size_t)key_size;

                    uint32_t tag = (uint32_t)key >> 3;
                    wire_type_t wire_type = (wire_type_t)(key & 0x07);
                    printf("      Inner tag=%u, type=%d: ", tag, wire_type);

                    if (wire_type == WIRE_TYPE_VARINT) {
                        int32_t val;
                        int val_size = varint_decode(f->value.bytes.data + offset,
                                                      f->value.bytes.len - offset, &val);
                        if (val_size > 0) {
                            printf("varint = %d\n", val);
                            offset += (size_t)val_size;
                        }
                    } else if (wire_type == WIRE_TYPE_FIXED32) {
                        uint8_t *p = f->value.bytes.data + offset;
                        uint32_t val = (uint32_t)p[0] << 24 | (uint32_t)p[1] << 16 |
                                       (uint32_t)p[2] << 8 | (uint32_t)p[3];
                        printf("fixed32 = 0x%08X\n", val);
                        offset += 4;
                    }
                }
                proto_msg_destroy(nested_msg);
            }
        }
    } else {
        printf("Decode failed: %d\n", decoded_size);
    }

    proto_msg_destroy(inner);
    proto_msg_destroy(outer);
    proto_msg_destroy(decoded);
}

static void test_crc_validation(void) {
    printf("\n=== Testing CRC validation ===\n");

    proto_msg_t *msg = proto_msg_create(1);
    proto_msg_add_varint(msg, 1, 12345);

    uint8_t buf[256];
    int encoded = proto_encode(msg, buf, sizeof(buf));

    proto_msg_t *decoded = proto_msg_create(1);
    int result = proto_decode(buf, (size_t)encoded, decoded);
    printf("Good CRC: decode returned %d (expected >= 0)\n", result);
    proto_msg_destroy(decoded);

    buf[0] ^= 0xFF;
    decoded = proto_msg_create(1);
    result = proto_decode(buf, (size_t)encoded, decoded);
    printf("Bad CRC (corrupted data): decode returned %d (expected -2)\n", result);
    proto_msg_destroy(decoded);

    proto_msg_destroy(msg);
}

static void test_buffer_too_small(void) {
    printf("\n=== Testing buffer too small ===\n");

    proto_msg_t *msg = proto_msg_create(1);
    proto_msg_add_varint(msg, 1, 123456789);

    uint8_t small_buf[4];
    int result = proto_encode(msg, small_buf, sizeof(small_buf));
    printf("Encode with small buffer: returned %d (expected -1)\n", result);

    proto_msg_destroy(msg);
}

static void test_varint_too_long(void) {
    printf("\n=== Testing varint too long ===\n");

    uint8_t long_varint[10];
    for (int i = 0; i < 6; i++) {
        long_varint[i] = 0x80 | (uint8_t)i;
    }

    int32_t value;
    int result = varint_decode(long_varint, sizeof(long_varint), &value);
    printf("Decode 6-byte varint: returned %d (expected -4)\n", result);
}

int main(void) {
    test_varint();
    test_crc32();
    test_proto_simple();
    test_proto_bytes();
    test_proto_nested();
    test_crc_validation();
    test_buffer_too_small();
    test_varint_too_long();

    printf("\n=== All tests completed ===\n");
    return 0;
}
