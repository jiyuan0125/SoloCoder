#include "base64_stream.h"
#include "base64_encode.h"
#include "base64_decode.h"
#include <string.h>

#define BASE64_INVALID 0xFF

static const char base64_table[] = {
    'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H',
    'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P',
    'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X',
    'Y', 'Z', 'a', 'b', 'c', 'd', 'e', 'f',
    'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
    'o', 'p', 'q', 'r', 's', 't', 'u', 'v',
    'w', 'x', 'y', 'z', '0', '1', '2', '3',
    '4', '5', '6', '7', '8', '9', '+', '/'
};

static const uint8_t decode_table[256] = {
    0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x3E, 0xFF, 0xFF, 0xFF, 0x3F,
    0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E,
    0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28,
    0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x2F, 0x30, 0x31, 0x32, 0x33, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
    0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF
};

void base64_encode_stream_init(base64_encode_stream_t *stream)
{
    if (stream) {
        memset(stream, 0, sizeof(base64_encode_stream_t));
    }
}

int base64_encode_stream_update(base64_encode_stream_t *stream,
                                 const uint8_t *input, size_t input_len,
                                 char *output, size_t output_size,
                                 size_t *output_len)
{
    if (!stream || !input || !output || !output_len)
        return BASE64_ERROR;
    
    *output_len = 0;
    
    if (input_len == 0)
        return BASE64_OK;
    
    size_t out_idx = 0;
    size_t col = stream->column;
    size_t total_available = output_size;
    
    const uint8_t *data = input;
    size_t data_len = input_len;
    
    if (stream->leftover_len > 0) {
        size_t needed = 3 - stream->leftover_len;
        if (data_len < needed) {
            memcpy(&stream->leftover[stream->leftover_len], data, data_len);
            stream->leftover_len += data_len;
            *output_len = 0;
            return BASE64_OK;
        }
        
        uint8_t combined[3];
        memcpy(combined, stream->leftover, stream->leftover_len);
        memcpy(&combined[stream->leftover_len], data, needed);
        
        if (total_available - out_idx < 5)
            return BASE64_BUFFER_SMALL;
        
        uint32_t val = ((uint32_t)combined[0] << 16) |
                       ((uint32_t)combined[1] << 8) |
                       ((uint32_t)combined[2]);
        
        output[out_idx++] = base64_table[(val >> 18) & 0x3F];
        output[out_idx++] = base64_table[(val >> 12) & 0x3F];
        output[out_idx++] = base64_table[(val >> 6) & 0x3F];
        output[out_idx++] = base64_table[val & 0x3F];
        
        col += 4;
        if (col >= BASE64_LINE_LENGTH) {
            output[out_idx++] = '\n';
            col = 0;
        }
        
        data += needed;
        data_len -= needed;
        stream->leftover_len = 0;
    }
    
    size_t i;
    for (i = 0; i + 2 < data_len; i += 3) {
        if (total_available - out_idx < 5)
            return BASE64_BUFFER_SMALL;
        
        uint32_t val = ((uint32_t)data[i] << 16) |
                       ((uint32_t)data[i + 1] << 8) |
                       ((uint32_t)data[i + 2]);
        
        output[out_idx++] = base64_table[(val >> 18) & 0x3F];
        output[out_idx++] = base64_table[(val >> 12) & 0x3F];
        output[out_idx++] = base64_table[(val >> 6) & 0x3F];
        output[out_idx++] = base64_table[val & 0x3F];
        
        col += 4;
        if (col >= BASE64_LINE_LENGTH) {
            output[out_idx++] = '\n';
            col = 0;
        }
    }
    
    size_t rem = data_len - i;
    if (rem > 0) {
        memcpy(stream->leftover, &data[i], rem);
        stream->leftover_len = rem;
    }
    
    stream->column = col;
    *output_len = out_idx;
    
    return BASE64_OK;
}

int base64_encode_stream_final(base64_encode_stream_t *stream,
                                char *output, size_t output_size,
                                size_t *output_len)
{
    if (!stream || !output || !output_len)
        return BASE64_ERROR;
    
    *output_len = 0;
    
    if (stream->leftover_len == 0) {
        if (output_size < 1)
            return BASE64_BUFFER_SMALL;
        output[0] = '\0';
        *output_len = 0;
        return BASE64_OK;
    }
    
    if (output_size < 6)
        return BASE64_BUFFER_SMALL;
    
    size_t out_idx = 0;
    uint32_t val = ((uint32_t)stream->leftover[0] << 16);
    if (stream->leftover_len > 1)
        val |= ((uint32_t)stream->leftover[1] << 8);
    
    output[out_idx++] = base64_table[(val >> 18) & 0x3F];
    output[out_idx++] = base64_table[(val >> 12) & 0x3F];
    
    if (stream->leftover_len == 2) {
        output[out_idx++] = base64_table[(val >> 6) & 0x3F];
        output[out_idx++] = '=';
    } else {
        output[out_idx++] = '=';
        output[out_idx++] = '=';
    }
    
    stream->column += 4;
    if (stream->column >= BASE64_LINE_LENGTH) {
        output[out_idx++] = '\n';
    }
    
    if (out_idx > 0 && output[out_idx - 1] == '\n') {
        out_idx--;
    }
    
    output[out_idx] = '\0';
    *output_len = out_idx;
    
    memset(stream, 0, sizeof(base64_encode_stream_t));
    
    return BASE64_OK;
}

void base64_decode_stream_init(base64_decode_stream_t *stream)
{
    if (stream) {
        memset(stream, 0, sizeof(base64_decode_stream_t));
    }
}

int base64_decode_stream_update(base64_decode_stream_t *stream,
                                 const char *input, size_t input_len,
                                 uint8_t *output, size_t output_size,
                                 size_t *output_len)
{
    if (!stream || !input || !output || !output_len)
        return BASE64_ERROR;
    
    *output_len = 0;
    
    if (input_len == 0)
        return BASE64_OK;
    
    size_t out_idx = 0;
    size_t total_available = output_size;
    size_t group_idx = stream->group_idx;
    
    size_t i;
    for (i = 0; i < input_len; i++) {
        unsigned char c = (unsigned char)input[i];
        
        if (c == '=')
            continue;
        
        uint8_t val = decode_table[c];
        if (val == BASE64_INVALID)
            continue;
        
        stream->group[group_idx++] = val;
        
        if (group_idx == 4) {
            if (total_available - out_idx < 3)
                return BASE64_BUFFER_SMALL;
            
            output[out_idx++] = (uint8_t)((stream->group[0] << 2) | (stream->group[1] >> 4));
            output[out_idx++] = (uint8_t)((stream->group[1] << 4) | (stream->group[2] >> 2));
            output[out_idx++] = (uint8_t)((stream->group[2] << 6) | stream->group[3]);
            group_idx = 0;
        }
    }
    
    stream->group_idx = group_idx;
    *output_len = out_idx;
    
    return BASE64_OK;
}

int base64_decode_stream_final(base64_decode_stream_t *stream,
                                uint8_t *output, size_t output_size,
                                size_t *output_len)
{
    if (!stream || !output || !output_len)
        return BASE64_ERROR;
    
    *output_len = 0;
    
    if (stream->group_idx <= 1) {
        memset(stream, 0, sizeof(base64_decode_stream_t));
        return BASE64_OK;
    }
    
    size_t out_idx = 0;
    
    if (stream->group_idx >= 2) {
        if (output_size - out_idx < 1)
            return BASE64_BUFFER_SMALL;
        output[out_idx++] = (uint8_t)((stream->group[0] << 2) | (stream->group[1] >> 4));
    }
    
    if (stream->group_idx == 3) {
        if (output_size - out_idx < 1)
            return BASE64_BUFFER_SMALL;
        output[out_idx++] = (uint8_t)((stream->group[1] << 4) | (stream->group[2] >> 2));
    }
    
    *output_len = out_idx;
    memset(stream, 0, sizeof(base64_decode_stream_t));
    
    return BASE64_OK;
}
