#include "smtp_client.h"
#include <stdlib.h>
#include <string.h>

static const char base64_table[] = 
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    "abcdefghijklmnopqrstuvwxyz"
    "0123456789+/";

static const int base64_decode_table[] = {
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 62, -1, -1, -1, 63,
    52, 53, 54, 55, 56, 57, 58, 59, 60, 61, -1, -1, -1, -2, -1, -1,
    -1,  0,  1,  2,  3,  4,  5,  6,  7,  8,  9, 10, 11, 12, 13, 14,
    15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, -1, -1, -1, -1, -1,
    -1, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
    41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
    -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1
};

static size_t base64_encoded_size_no_wrap(size_t input_length) {
    return ((input_length + 2) / 3) * 4;
}

static size_t base64_encoded_size_with_wrap(size_t input_length) {
    size_t chars = base64_encoded_size_no_wrap(input_length);
    size_t lines = (chars + 75) / 76;
    size_t line_breaks = lines > 0 ? (lines - 1) * 2 : 0;
    return chars + line_breaks + 1;
}

static size_t base64_decoded_size(size_t input_length) {
    return (input_length * 3) / 4 + 1;
}

static char *base64_encode_no_wrap(const unsigned char *data, size_t input_length, size_t *output_length) {
    if (data == NULL || input_length == 0) {
        if (output_length != NULL) *output_length = 0;
        return NULL;
    }

    size_t out_len = base64_encoded_size_no_wrap(input_length);
    char *output = (char *)malloc(out_len + 1);
    if (output == NULL) {
        if (output_length != NULL) *output_length = 0;
        return NULL;
    }

    size_t out_idx = 0;
    for (size_t i = 0; i < input_length; i += 3) {
        unsigned char o1 = data[i];
        unsigned char o2 = (i + 1 < input_length) ? data[i + 1] : 0;
        unsigned char o3 = (i + 2 < input_length) ? data[i + 2] : 0;

        unsigned int triple = ((unsigned int)o1 << 16) |
                               ((unsigned int)o2 << 8) |
                               (unsigned int)o3;

        output[out_idx++] = base64_table[(triple >> 18) & 0x3F];
        output[out_idx++] = base64_table[(triple >> 12) & 0x3F];
        
        if (i + 1 < input_length) {
            output[out_idx++] = base64_table[(triple >> 6) & 0x3F];
        } else {
            output[out_idx++] = '=';
        }
        
        if (i + 2 < input_length) {
            output[out_idx++] = base64_table[triple & 0x3F];
        } else {
            output[out_idx++] = '=';
        }
    }
    output[out_idx] = '\0';

    if (output_length != NULL) {
        *output_length = out_idx;
    }

    return output;
}

char *base64_encode(const unsigned char *data, size_t input_length, size_t *output_length) {
    char *no_wrap = base64_encode_no_wrap(data, input_length, NULL);
    if (no_wrap == NULL) {
        if (output_length != NULL) *output_length = 0;
        return NULL;
    }

    size_t no_wrap_len = strlen(no_wrap);
    size_t required = base64_encoded_size_with_wrap(input_length);
    char *output = (char *)malloc(required);
    if (output == NULL) {
        free(no_wrap);
        if (output_length != NULL) *output_length = 0;
        return NULL;
    }

    size_t out_idx = 0;
    for (size_t i = 0; i < no_wrap_len; i++) {
        if (i > 0 && i % 76 == 0) {
            output[out_idx++] = '\r';
            output[out_idx++] = '\n';
        }
        output[out_idx++] = no_wrap[i];
    }
    output[out_idx] = '\0';

    free(no_wrap);

    if (output_length != NULL) {
        *output_length = out_idx;
    }

    return output;
}

unsigned char *base64_decode(const char *data, size_t input_length, size_t *output_length) {
    if (data == NULL || input_length == 0) {
        if (output_length != NULL) *output_length = 0;
        return NULL;
    }

    char *filtered = (char *)malloc(input_length + 1);
    if (filtered == NULL) {
        if (output_length != NULL) *output_length = 0;
        return NULL;
    }

    size_t filtered_len = 0;
    for (size_t i = 0; i < input_length; i++) {
        char c = data[i];
        if (c != ' ' && c != '\t' && c != '\r' && c != '\n') {
            filtered[filtered_len++] = c;
        }
    }
    filtered[filtered_len] = '\0';

    if (filtered_len == 0 || filtered_len % 4 != 0) {
        free(filtered);
        if (output_length != NULL) *output_length = 0;
        return NULL;
    }

    size_t max_output = base64_decoded_size(filtered_len);
    unsigned char *output = (unsigned char *)malloc(max_output);
    if (output == NULL) {
        free(filtered);
        if (output_length != NULL) *output_length = 0;
        return NULL;
    }

    size_t out_idx = 0;
    for (size_t i = 0; i < filtered_len; i += 4) {
        int val[4];
        int pad = 0;

        for (int j = 0; j < 4; j++) {
            unsigned char c = (unsigned char)filtered[i + j];
            if (c == '=') {
                val[j] = 0;
                pad++;
            } else if (c > 127 || base64_decode_table[c] < 0) {
                free(filtered);
                free(output);
                if (output_length != NULL) *output_length = 0;
                return NULL;
            } else {
                val[j] = base64_decode_table[c];
            }
        }

        if (pad > 0 && i + 4 < filtered_len) {
            free(filtered);
            free(output);
            if (output_length != NULL) *output_length = 0;
            return NULL;
        }

        unsigned int triple = ((unsigned int)val[0] << 18) |
                               ((unsigned int)val[1] << 12) |
                               ((unsigned int)val[2] << 6) |
                               (unsigned int)val[3];

        output[out_idx++] = (triple >> 16) & 0xFF;
        if (pad < 2) {
            output[out_idx++] = (triple >> 8) & 0xFF;
        }
        if (pad < 1) {
            output[out_idx++] = triple & 0xFF;
        }
    }

    free(filtered);

    if (output_length != NULL) {
        *output_length = out_idx;
    }

    return output;
}
