#include "base64_encode.h"
#include <string.h>

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

size_t base64_encode_size(size_t input_len)
{
    if (input_len == 0)
        return 1;
    
    size_t base64_len = ((input_len + 2) / 3) * 4;
    size_t line_breaks = (base64_len + BASE64_LINE_LENGTH - 1) / BASE64_LINE_LENGTH - 1;
    
    return base64_len + line_breaks + 1;
}

int base64_encode(const uint8_t *input, size_t input_len,
                  char *output, size_t output_size,
                  size_t *output_len)
{
    size_t required = base64_encode_size(input_len);
    if (output_size < required) {
        if (output_len)
            *output_len = required;
        return BASE64_BUFFER_SMALL;
    }
    
    if (input_len == 0) {
        if (output)
            output[0] = '\0';
        if (output_len)
            *output_len = 0;
        return BASE64_OK;
    }
    
    size_t idx = 0;
    size_t col = 0;
    
    size_t i;
    for (i = 0; i + 2 < input_len; i += 3) {
        uint32_t val = ((uint32_t)input[i] << 16) |
                       ((uint32_t)input[i + 1] << 8) |
                       ((uint32_t)input[i + 2]);
        
        output[idx++] = base64_table[(val >> 18) & 0x3F];
        output[idx++] = base64_table[(val >> 12) & 0x3F];
        output[idx++] = base64_table[(val >> 6) & 0x3F];
        output[idx++] = base64_table[val & 0x3F];
        
        col += 4;
        if (col >= BASE64_LINE_LENGTH) {
            output[idx++] = '\n';
            col = 0;
        }
    }
    
    size_t rem = input_len - i;
    if (rem > 0) {
        uint32_t val = ((uint32_t)input[i] << 16);
        if (rem > 1)
            val |= ((uint32_t)input[i + 1] << 8);
        
        output[idx++] = base64_table[(val >> 18) & 0x3F];
        output[idx++] = base64_table[(val >> 12) & 0x3F];
        
        if (rem == 3) {
            output[idx++] = base64_table[(val >> 6) & 0x3F];
            output[idx++] = base64_table[val & 0x3F];
        } else if (rem == 2) {
            output[idx++] = base64_table[(val >> 6) & 0x3F];
            output[idx++] = '=';
        } else {
            output[idx++] = '=';
            output[idx++] = '=';
        }
        
        col += 4;
        if (col >= BASE64_LINE_LENGTH) {
            output[idx++] = '\n';
            col = 0;
        }
    }
    
    if (idx > 0 && output[idx - 1] == '\n') {
        idx--;
    }
    
    output[idx] = '\0';
    
    if (output_len)
        *output_len = idx;
    
    return BASE64_OK;
}
