#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "base64_encode.h"
#include "base64_decode.h"
#include "base64_stream.h"

static int test_basic_encode_decode(void)
{
    printf("=== Test 1: Basic Encode/Decode ===\n");
    
    const uint8_t original[] = {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
                                 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
                                 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
                                 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F,
                                 'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd', '!',
                                 0xFF, 0xFE, 0xFD, 0xFC, 0xFB, 0xFA};
    
    size_t original_len = sizeof(original);
    
    size_t encode_size = base64_encode_size(original_len);
    char *encoded = (char *)malloc(encode_size);
    if (!encoded) {
        printf("  FAIL: Memory allocation failed\n");
        return -1;
    }
    
    size_t encoded_len;
    int ret = base64_encode(original, original_len, encoded, encode_size, &encoded_len);
    if (ret != BASE64_OK) {
        printf("  FAIL: Encode failed with code %d\n", ret);
        free(encoded);
        return -1;
    }
    
    printf("  Original length: %zu bytes\n", original_len);
    printf("  Encoded length: %zu bytes\n", encoded_len);
    printf("  Encoded:\n%s\n", encoded);
    
    size_t decode_size = base64_decode_size(encoded_len);
    uint8_t *decoded = (uint8_t *)malloc(decode_size);
    if (!decoded) {
        printf("  FAIL: Memory allocation failed\n");
        free(encoded);
        return -1;
    }
    
    size_t decoded_len;
    ret = base64_decode(encoded, encoded_len, decoded, decode_size, &decoded_len);
    if (ret != BASE64_OK) {
        printf("  FAIL: Decode failed with code %d\n", ret);
        free(encoded);
        free(decoded);
        return -1;
    }
    
    printf("  Decoded length: %zu bytes\n", decoded_len);
    
    if (decoded_len != original_len) {
        printf("  FAIL: Length mismatch! Original: %zu, Decoded: %zu\n", original_len, decoded_len);
        free(encoded);
        free(decoded);
        return -1;
    }
    
    if (memcmp(original, decoded, original_len) != 0) {
        printf("  FAIL: Data mismatch!\n");
        free(encoded);
        free(decoded);
        return -1;
    }
    
    printf("  PASS: Data matches!\n\n");
    
    free(encoded);
    free(decoded);
    return 0;
}

static int test_empty_input(void)
{
    printf("=== Test 2: Empty Input ===\n");
    
    char encode_buf[10];
    size_t encode_len;
    
    int ret = base64_encode(NULL, 0, encode_buf, sizeof(encode_buf), &encode_len);
    if (ret != BASE64_OK || encode_len != 0 || encode_buf[0] != '\0') {
        printf("  FAIL: Empty encode failed\n");
        return -1;
    }
    printf("  Empty encode: PASS (output is empty string)\n");
    
    uint8_t decode_buf[10];
    size_t decode_len;
    
    ret = base64_decode("", 0, decode_buf, sizeof(decode_buf), &decode_len);
    if (ret != BASE64_OK || decode_len != 0) {
        printf("  FAIL: Empty decode failed\n");
        return -1;
    }
    printf("  Empty decode: PASS (output is empty)\n\n");
    
    return 0;
}

static int test_padding(void)
{
    printf("=== Test 3: Padding Handling ===\n");
    
    const char *test1 = "QQ==";
    const char *test2 = "QUI=";
    const char *test3 = "QUJD";
    
    uint8_t buf[10];
    size_t len;
    int ret;
    
    ret = base64_decode(test1, strlen(test1), buf, sizeof(buf), &len);
    if (ret != BASE64_OK || len != 1 || buf[0] != 'A') {
        printf("  FAIL: Test 1 (QQ==) failed\n");
        return -1;
    }
    printf("  'QQ==' -> 'A' (1 byte): PASS\n");
    
    ret = base64_decode(test2, strlen(test2), buf, sizeof(buf), &len);
    if (ret != BASE64_OK || len != 2 || buf[0] != 'A' || buf[1] != 'B') {
        printf("  FAIL: Test 2 (QUI=) failed\n");
        return -1;
    }
    printf("  'QUI=' -> 'AB' (2 bytes): PASS\n");
    
    ret = base64_decode(test3, strlen(test3), buf, sizeof(buf), &len);
    if (ret != BASE64_OK || len != 3 || buf[0] != 'A' || buf[1] != 'B' || buf[2] != 'C') {
        printf("  FAIL: Test 3 (QUJD) failed\n");
        return -1;
    }
    printf("  'QUJD' -> 'ABC' (3 bytes): PASS\n\n");
    
    return 0;
}

static int test_stream_encode(void)
{
    printf("=== Test 4: Stream Encode ===\n");
    
    const uint8_t original[] = "The quick brown fox jumps over the lazy dog. "
                                "This is a longer test string to verify that "
                                "stream encoding works correctly with multiple chunks.";
    size_t original_len = strlen((const char *)original);
    
    size_t total_encode_size = base64_encode_size(original_len);
    char *expected = (char *)malloc(total_encode_size);
    size_t expected_len;
    base64_encode(original, original_len, expected, total_encode_size, &expected_len);
    
    base64_encode_stream_t stream;
    base64_encode_stream_init(&stream);
    
    char *stream_encoded = (char *)malloc(total_encode_size + 100);
    size_t stream_idx = 0;
    
    const size_t chunk_size = 7;
    size_t i;
    for (i = 0; i < original_len; i += chunk_size) {
        size_t this_chunk = (i + chunk_size > original_len) ? (original_len - i) : chunk_size;
        
        char chunk_output[100];
        size_t chunk_output_len;
        
        int ret = base64_encode_stream_update(&stream, original + i, this_chunk,
                                                chunk_output, sizeof(chunk_output), &chunk_output_len);
        if (ret != BASE64_OK) {
            printf("  FAIL: Stream update failed at chunk %zu\n", i / chunk_size);
            free(expected);
            free(stream_encoded);
            return -1;
        }
        
        memcpy(stream_encoded + stream_idx, chunk_output, chunk_output_len);
        stream_idx += chunk_output_len;
    }
    
    char final_output[100];
    size_t final_output_len;
    int ret = base64_encode_stream_final(&stream, final_output, sizeof(final_output), &final_output_len);
    if (ret != BASE64_OK) {
        printf("  FAIL: Stream final failed\n");
        free(expected);
        free(stream_encoded);
        return -1;
    }
    
    memcpy(stream_encoded + stream_idx, final_output, final_output_len);
    stream_idx += final_output_len;
    stream_encoded[stream_idx] = '\0';
    
    printf("  Non-stream output:\n%s\n", expected);
    printf("  Stream output:\n%s\n", stream_encoded);
    
    if (strcmp(expected, stream_encoded) != 0) {
        printf("  FAIL: Stream output mismatch!\n");
        free(expected);
        free(stream_encoded);
        return -1;
    }
    printf("  PASS: Stream encode matches non-stream encode!\n\n");
    
    free(expected);
    free(stream_encoded);
    return 0;
}

static int test_stream_decode(void)
{
    printf("=== Test 5: Stream Decode ===\n");
    
    const char *encoded = "VGhlIHF1aWNrIGJyb3duIGZveCBqdW1wcyBvdmVyIHRoZSBsYXp5IGRvZy4g"
                          "VGhpcyBpcyBhIGxvbmdlciB0ZXN0IHN0cmluZyB0byB2ZXJpZnkgdGhhdCBz"
                          "dHJlYW0gZGVjb2Rpbmcgd29ya3Mu";
    
    size_t encoded_len = strlen(encoded);
    
    size_t decode_size = base64_decode_size(encoded_len);
    uint8_t *expected = (uint8_t *)malloc(decode_size);
    size_t expected_len;
    base64_decode(encoded, encoded_len, expected, decode_size, &expected_len);
    
    base64_decode_stream_t stream;
    base64_decode_stream_init(&stream);
    
    uint8_t *stream_decoded = (uint8_t *)malloc(decode_size + 100);
    size_t stream_idx = 0;
    
    const size_t chunk_size = 5;
    size_t i;
    for (i = 0; i < encoded_len; i += chunk_size) {
        size_t this_chunk = (i + chunk_size > encoded_len) ? (encoded_len - i) : chunk_size;
        
        uint8_t chunk_output[100];
        size_t chunk_output_len;
        
        int ret = base64_decode_stream_update(&stream, encoded + i, this_chunk,
                                                chunk_output, sizeof(chunk_output), &chunk_output_len);
        if (ret != BASE64_OK) {
            printf("  FAIL: Stream update failed at chunk %zu\n", i / chunk_size);
            free(expected);
            free(stream_decoded);
            return -1;
        }
        
        memcpy(stream_decoded + stream_idx, chunk_output, chunk_output_len);
        stream_idx += chunk_output_len;
    }
    
    uint8_t final_output[100];
    size_t final_output_len;
    int ret = base64_decode_stream_final(&stream, final_output, sizeof(final_output), &final_output_len);
    if (ret != BASE64_OK) {
        printf("  FAIL: Stream final failed\n");
        free(expected);
        free(stream_decoded);
        return -1;
    }
    
    memcpy(stream_decoded + stream_idx, final_output, final_output_len);
    stream_idx += final_output_len;
    
    expected[expected_len] = '\0';
    stream_decoded[stream_idx] = '\0';
    
    printf("  Non-stream output: %s\n", expected);
    printf("  Stream output: %s\n", stream_decoded);
    
    if (expected_len != stream_idx || memcmp(expected, stream_decoded, expected_len) != 0) {
        printf("  FAIL: Stream output mismatch!\n");
        free(expected);
        free(stream_decoded);
        return -1;
    }
    printf("  PASS: Stream decode matches non-stream decode!\n\n");
    
    free(expected);
    free(stream_decoded);
    return 0;
}

static int test_buffer_size_check(void)
{
    printf("=== Test 6: Buffer Size Check ===\n");
    
    const uint8_t original[] = "Hello, World!";
    size_t original_len = strlen((const char *)original);
    
    size_t required = base64_encode_size(original_len);
    char *small_buf = (char *)malloc(required - 1);
    size_t out_len;
    
    int ret = base64_encode(original, original_len, small_buf, required - 1, &out_len);
    if (ret != BASE64_BUFFER_SMALL || out_len != required) {
        printf("  FAIL: Encode buffer size check failed\n");
        free(small_buf);
        return -1;
    }
    printf("  Encode: small buffer correctly returns BASE64_BUFFER_SMALL with required size %zu\n", required);
    
    const char *encoded = "SGVsbG8sIFdvcmxkIQ==";
    size_t encoded_len = strlen(encoded);
    size_t decode_required = base64_decode_size(encoded_len);
    
    uint8_t *small_decode_buf = (uint8_t *)malloc(decode_required - 1);
    ret = base64_decode(encoded, encoded_len, small_decode_buf, decode_required - 1, &out_len);
    if (ret != BASE64_BUFFER_SMALL || out_len != decode_required) {
        printf("  FAIL: Decode buffer size check failed\n");
        free(small_buf);
        free(small_decode_buf);
        return -1;
    }
    printf("  Decode: small buffer correctly returns BASE64_BUFFER_SMALL with required size %zu\n", decode_required);
    printf("  PASS: Buffer size checks work!\n\n");
    
    free(small_buf);
    free(small_decode_buf);
    return 0;
}

int main(void)
{
    printf("========================================\n");
    printf("Base64 Encode/Decode Module Test Suite\n");
    printf("========================================\n\n");
    
    int failed = 0;
    
    if (test_basic_encode_decode() != 0) failed++;
    if (test_empty_input() != 0) failed++;
    if (test_padding() != 0) failed++;
    if (test_stream_encode() != 0) failed++;
    if (test_stream_decode() != 0) failed++;
    if (test_buffer_size_check() != 0) failed++;
    
    printf("========================================\n");
    if (failed == 0) {
        printf("All tests PASSED!\n");
        return 0;
    } else {
        printf("%d test(s) FAILED!\n", failed);
        return 1;
    }
}
