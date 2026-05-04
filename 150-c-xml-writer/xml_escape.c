#include "xml_escape.h"
#include <string.h>
#include <stdlib.h>

static const struct {
    char original;
    const char *replacement;
    size_t replacement_len;
} xml_entities[] = {
    {'<', "&lt;", 4},
    {'>', "&gt;", 4},
    {'&', "&amp;", 5},
    {'"', "&quot;", 6},
    {'\'', "&apos;", 6}
};

#define XML_ENTITY_COUNT (sizeof(xml_entities) / sizeof(xml_entities[0]))

size_t XML_Escape_Length(const char *input) {
    if (input == NULL) return 0;
    
    size_t len = 0;
    for (const char *p = input; *p != '\0'; p++) {
        int found = 0;
        for (size_t i = 0; i < XML_ENTITY_COUNT; i++) {
            if (*p == xml_entities[i].original) {
                len += xml_entities[i].replacement_len;
                found = 1;
                break;
            }
        }
        if (!found) {
            len++;
        }
    }
    return len;
}

size_t XML_Escape(char *output, size_t output_size, const char *input) {
    if (input == NULL || output == NULL || output_size == 0) {
        if (output != NULL && output_size > 0) {
            output[0] = '\0';
        }
        return 0;
    }
    
    char fixed_input[4096];
    XML_ValidateAndFixUTF8(fixed_input, sizeof(fixed_input), input);
    
    size_t pos = 0;
    for (const char *p = fixed_input; *p != '\0' && pos < output_size - 1; p++) {
        int found = 0;
        for (size_t i = 0; i < XML_ENTITY_COUNT; i++) {
            if (*p == xml_entities[i].original) {
                size_t rlen = xml_entities[i].replacement_len;
                if (pos + rlen > output_size - 1) {
                    break;
                }
                memcpy(output + pos, xml_entities[i].replacement, rlen);
                pos += rlen;
                found = 1;
                break;
            }
        }
        if (!found) {
            output[pos++] = *p;
        }
    }
    
    output[pos] = '\0';
    return pos;
}

int XML_IsValidUTF8(const char *input) {
    if (input == NULL) return 0;
    
    const unsigned char *p = (const unsigned char *)input;
    while (*p != '\0') {
        if (*p <= 0x7F) {
            p++;
        } else if ((*p & 0xE0) == 0xC0) {
            if ((*(p + 1) & 0xC0) != 0x80) return 0;
            p += 2;
        } else if ((*p & 0xF0) == 0xE0) {
            if ((*(p + 1) & 0xC0) != 0x80) return 0;
            if ((*(p + 2) & 0xC0) != 0x80) return 0;
            p += 3;
        } else if ((*p & 0xF8) == 0xF0) {
            if ((*(p + 1) & 0xC0) != 0x80) return 0;
            if ((*(p + 2) & 0xC0) != 0x80) return 0;
            if ((*(p + 3) & 0xC0) != 0x80) return 0;
            p += 4;
        } else {
            return 0;
        }
    }
    return 1;
}

size_t XML_ValidateAndFixUTF8(char *output, size_t output_size, const char *input) {
    if (input == NULL || output == NULL || output_size == 0) {
        if (output != NULL && output_size > 0) {
            output[0] = '\0';
        }
        return 0;
    }
    
    const unsigned char *p = (const unsigned char *)input;
    size_t pos = 0;
    
    while (*p != '\0' && pos < output_size - 1) {
        if (*p <= 0x7F) {
            output[pos++] = (char)*p++;
        } else if ((*p & 0xE0) == 0xC0) {
            if ((*(p + 1) & 0xC0) != 0x80) {
                if (pos + 3 < output_size - 1) {
                    memcpy(output + pos, "\xEF\xBF\xBD", 3);
                    pos += 3;
                }
                p++;
            } else {
                if (pos + 2 < output_size - 1) {
                    output[pos++] = (char)*p++;
                    output[pos++] = (char)*p++;
                } else {
                    break;
                }
            }
        } else if ((*p & 0xF0) == 0xE0) {
            if ((*(p + 1) & 0xC0) != 0x80 || (*(p + 2) & 0xC0) != 0x80) {
                if (pos + 3 < output_size - 1) {
                    memcpy(output + pos, "\xEF\xBF\xBD", 3);
                    pos += 3;
                }
                p++;
            } else {
                if (pos + 3 < output_size - 1) {
                    output[pos++] = (char)*p++;
                    output[pos++] = (char)*p++;
                    output[pos++] = (char)*p++;
                } else {
                    break;
                }
            }
        } else if ((*p & 0xF8) == 0xF0) {
            if ((*(p + 1) & 0xC0) != 0x80 || 
                (*(p + 2) & 0xC0) != 0x80 ||
                (*(p + 3) & 0xC0) != 0x80) {
                if (pos + 3 < output_size - 1) {
                    memcpy(output + pos, "\xEF\xBF\xBD", 3);
                    pos += 3;
                }
                p++;
            } else {
                if (pos + 4 < output_size - 1) {
                    output[pos++] = (char)*p++;
                    output[pos++] = (char)*p++;
                    output[pos++] = (char)*p++;
                    output[pos++] = (char)*p++;
                } else {
                    break;
                }
            }
        } else {
            if (pos + 3 < output_size - 1) {
                memcpy(output + pos, "\xEF\xBF\xBD", 3);
                pos += 3;
            }
            p++;
        }
    }
    
    output[pos] = '\0';
    return pos;
}

size_t XML_CDATA_SafeLength(const char *input) {
    if (input == NULL) return 0;
    
    size_t len = 0;
    size_t consecutive_brackets = 0;
    
    for (const char *p = input; *p != '\0'; p++) {
        if (*p == ']') {
            consecutive_brackets++;
            len++;
        } else if (*p == '>' && consecutive_brackets >= 2) {
            len += 3;
            consecutive_brackets = 0;
        } else {
            consecutive_brackets = 0;
            len++;
        }
    }
    return len;
}

size_t XML_CDATA_Safe(char *output, size_t output_size, const char *input) {
    if (input == NULL || output == NULL || output_size == 0) {
        if (output != NULL && output_size > 0) {
            output[0] = '\0';
        }
        return 0;
    }
    
    char fixed_input[8192];
    XML_ValidateAndFixUTF8(fixed_input, sizeof(fixed_input), input);
    
    size_t pos = 0;
    const char *p = fixed_input;
    
    while (*p != '\0' && pos < output_size - 1) {
        if (*p == ']' && *(p + 1) == ']' && *(p + 2) == '>') {
            if (pos + 15 >= output_size - 1) break;
            output[pos++] = ']';
            output[pos++] = ']';
            memcpy(output + pos, "]]><![CDATA[", 12);
            pos += 12;
            output[pos++] = '>';
            p += 3;
        } else {
            output[pos++] = *p++;
        }
    }
    
    output[pos] = '\0';
    return pos;
}
