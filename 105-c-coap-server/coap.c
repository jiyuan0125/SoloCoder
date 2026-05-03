#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "coap.h"

void coap_message_init(coap_message_t *msg) {
    memset(msg, 0, sizeof(coap_message_t));
    msg->version = COAP_VERSION;
    msg->type = COAP_TYPE_CON;
}

void coap_message_free(coap_message_t *msg) {
    if (msg->options) {
        for (uint8_t i = 0; i < msg->option_count; i++) {
            if (msg->options[i].value) {
                free(msg->options[i].value);
            }
        }
        free(msg->options);
    }
    if (msg->payload) {
        free(msg->payload);
    }
    memset(msg, 0, sizeof(coap_message_t));
}

static int coap_encode_option_length(uint16_t len, uint8_t *buffer, size_t max_len) {
    if (len < 13) {
        if (max_len < 1) return -1;
        buffer[0] = len;
        return 1;
    } else if (len <= 0xff + 13) {
        if (max_len < 2) return -1;
        buffer[0] = 13;
        buffer[1] = len - 13;
        return 2;
    } else {
        if (max_len < 3) return -1;
        buffer[0] = 14;
        buffer[1] = (len - 269) >> 8;
        buffer[2] = (len - 269) & 0xff;
        return 3;
    }
}

static int coap_decode_option_length(const uint8_t *buffer, size_t buffer_len, uint8_t header_len, uint16_t *result) {
    if (header_len < 13) {
        *result = header_len;
        return 1;
    } else if (header_len == 13) {
        if (buffer_len < 2) return -1;
        *result = buffer[1] + 13;
        return 2;
    } else if (header_len == 14) {
        if (buffer_len < 3) return -1;
        *result = ((buffer[1] << 8) | buffer[2]) + 269;
        return 3;
    } else {
        return -1;
    }
}

static int coap_option_encode(const coap_option_t *opt, uint8_t *buffer, size_t buffer_size) {
    uint8_t delta_bytes[4], len_bytes[4];
    int delta_len, len_len;
    
    delta_len = coap_encode_option_length(opt->delta, delta_bytes, sizeof(delta_bytes));
    if (delta_len < 0) return -1;
    
    len_len = coap_encode_option_length(opt->length, len_bytes, sizeof(len_bytes));
    if (len_len < 0) return -1;
    
    if (buffer_size < (size_t)(delta_len + len_len + opt->length)) {
        return -1;
    }
    
    buffer[0] = (delta_bytes[0] << 4) | len_bytes[0];
    int pos = 1;
    
    for (int i = 1; i < delta_len; i++) {
        buffer[pos++] = delta_bytes[i];
    }
    for (int i = 1; i < len_len; i++) {
        buffer[pos++] = len_bytes[i];
    }
    
    if (opt->value && opt->length > 0) {
        memcpy(&buffer[pos], opt->value, opt->length);
        pos += opt->length;
    }
    
    return pos;
}

int coap_message_encode(const coap_message_t *msg, uint8_t *buffer, size_t buffer_size) {
    if (buffer_size < 4) return -1;
    
    int pos = 0;
    buffer[pos++] = (msg->version << 6) | (msg->type << 4) | msg->token_len;
    buffer[pos++] = (uint8_t)msg->code;
    buffer[pos++] = (msg->msg_id >> 8) & 0xff;
    buffer[pos++] = msg->msg_id & 0xff;
    
    if (msg->token_len > 0 && msg->token_len <= COAP_MAX_TOKEN_LEN) {
        if ((size_t)pos + msg->token_len > buffer_size) return -1;
        memcpy(&buffer[pos], msg->token, msg->token_len);
        pos += msg->token_len;
    }
    
    uint16_t prev_number = 0;
    for (uint8_t i = 0; i < msg->option_count; i++) {
        uint16_t number = prev_number + msg->options[i].delta;
        int encoded = coap_option_encode(&msg->options[i], &buffer[pos], buffer_size - pos);
        if (encoded < 0) return -1;
        pos += encoded;
        prev_number = number;
    }
    
    if (msg->payload_len > 0) {
        if ((size_t)pos + 1 + msg->payload_len > buffer_size) return -1;
        buffer[pos++] = 0xff;
        memcpy(&buffer[pos], msg->payload, msg->payload_len);
        pos += msg->payload_len;
    }
    
    return pos;
}

static int coap_option_decode(coap_option_t *opt, const uint8_t *buffer, size_t buffer_len, uint16_t *prev_number) {
    if (buffer_len < 1) return -1;
    
    uint8_t delta_header = (buffer[0] >> 4) & 0x0f;
    uint8_t len_header = buffer[0] & 0x0f;
    
    int delta_len = coap_decode_option_length(buffer, buffer_len, delta_header, &opt->delta);
    if (delta_len < 0) return -1;
    
    int len_len = coap_decode_option_length(&buffer[delta_len], buffer_len - delta_len, len_header, &opt->length);
    if (len_len < 0) return -1;
    
    int header_len = delta_len + len_len - 1;
    if (header_len < 0) return -1;
    
    if (buffer_len < (size_t)(header_len + opt->length)) return -1;
    
    if (opt->length > 0) {
        opt->value = malloc(opt->length);
        if (!opt->value) return -1;
        memcpy(opt->value, &buffer[header_len], opt->length);
    } else {
        opt->value = NULL;
    }
    
    *prev_number += opt->delta;
    return header_len + opt->length;
}

int coap_message_decode(coap_message_t *msg, const uint8_t *buffer, size_t buffer_size) {
    if (buffer_size < 4) return -1;
    
    coap_message_init(msg);
    
    int pos = 0;
    msg->version = (buffer[pos] >> 6) & 0x03;
    msg->type = (buffer[pos] >> 4) & 0x03;
    msg->token_len = buffer[pos] & 0x0f;
    pos++;
    
    msg->code = (coap_code_t)buffer[pos++];
    msg->msg_id = (buffer[pos] << 8) | buffer[pos + 1];
    pos += 2;
    
    if (msg->token_len > 0 && msg->token_len <= COAP_MAX_TOKEN_LEN) {
        if ((size_t)pos + msg->token_len > buffer_size) {
            coap_message_free(msg);
            return -1;
        }
        memcpy(msg->token, &buffer[pos], msg->token_len);
        pos += msg->token_len;
    }
    
    msg->option_count = 0;
    msg->options = NULL;
    uint16_t prev_number = 0;
    
    while ((size_t)pos < buffer_size) {
        if (buffer[pos] == 0xff) {
            pos++;
            if ((size_t)pos < buffer_size) {
                msg->payload_len = buffer_size - pos;
                msg->payload = malloc(msg->payload_len);
                if (!msg->payload) {
                    coap_message_free(msg);
                    return -1;
                }
                memcpy(msg->payload, &buffer[pos], msg->payload_len);
            }
            return buffer_size;
        }
        
        coap_option_t *new_options = realloc(msg->options, (msg->option_count + 1) * sizeof(coap_option_t));
        if (!new_options) {
            coap_message_free(msg);
            return -1;
        }
        msg->options = new_options;
        
        coap_option_t *opt = &msg->options[msg->option_count];
        memset(opt, 0, sizeof(coap_option_t));
        
        int consumed = coap_option_decode(opt, &buffer[pos], buffer_size - pos, &prev_number);
        if (consumed < 0) {
            coap_message_free(msg);
            return -1;
        }
        
        msg->option_count++;
        pos += consumed;
    }
    
    return pos;
}

coap_option_t *coap_message_add_option(coap_message_t *msg, uint16_t number, const uint8_t *value, size_t length) {
    uint16_t prev_number = 0;
    uint16_t delta;
    
    if (msg->option_count > 0) {
        for (uint8_t i = 0; i < msg->option_count; i++) {
            prev_number += msg->options[i].delta;
        }
    }
    
    if (number < prev_number) {
        return NULL;
    }
    
    delta = number - prev_number;
    
    coap_option_t *new_options = realloc(msg->options, (msg->option_count + 1) * sizeof(coap_option_t));
    if (!new_options) {
        return NULL;
    }
    msg->options = new_options;
    
    coap_option_t *opt = &msg->options[msg->option_count];
    memset(opt, 0, sizeof(coap_option_t));
    opt->delta = delta;
    opt->length = length;
    
    if (length > 0 && value) {
        opt->value = malloc(length);
        if (!opt->value) {
            return NULL;
        }
        memcpy(opt->value, value, length);
    } else {
        opt->value = NULL;
    }
    
    msg->option_count++;
    return opt;
}

coap_option_t *coap_message_get_option(const coap_message_t *msg, uint16_t number) {
    uint16_t current = 0;
    for (uint8_t i = 0; i < msg->option_count; i++) {
        current += msg->options[i].delta;
        if (current == number) {
            return &msg->options[i];
        }
    }
    return NULL;
}

coap_option_t *coap_message_get_option_by_index(const coap_message_t *msg, uint8_t index) {
    if (index >= msg->option_count) {
        return NULL;
    }
    return &msg->options[index];
}

int coap_message_set_payload(coap_message_t *msg, const uint8_t *payload, size_t length) {
    if (msg->payload) {
        free(msg->payload);
    }
    
    if (length > 0 && payload) {
        msg->payload = malloc(length);
        if (!msg->payload) {
            msg->payload_len = 0;
            return -1;
        }
        memcpy(msg->payload, payload, length);
        msg->payload_len = length;
    } else {
        msg->payload = NULL;
        msg->payload_len = 0;
    }
    return 0;
}

const char *coap_code_to_string(coap_code_t code) {
    switch (code) {
        case COAP_CODE_EMPTY: return "Empty";
        case COAP_CODE_GET: return "GET";
        case COAP_CODE_POST: return "POST";
        case COAP_CODE_PUT: return "PUT";
        case COAP_CODE_DELETE: return "DELETE";
        case COAP_CODE_201_CREATED: return "2.01 Created";
        case COAP_CODE_202_DELETED: return "2.02 Deleted";
        case COAP_CODE_203_VALID: return "2.03 Valid";
        case COAP_CODE_204_CHANGED: return "2.04 Changed";
        case COAP_CODE_205_CONTENT: return "2.05 Content";
        case COAP_CODE_400_BAD_REQUEST: return "4.00 Bad Request";
        case COAP_CODE_404_NOT_FOUND: return "4.04 Not Found";
        case COAP_CODE_405_METHOD_NOT_ALLOWED: return "4.05 Method Not Allowed";
        case COAP_CODE_500_INTERNAL_SERVER_ERROR: return "5.00 Internal Server Error";
        default: return "Unknown";
    }
}

const char *coap_type_to_string(coap_type_t type) {
    switch (type) {
        case COAP_TYPE_CON: return "CON";
        case COAP_TYPE_NON: return "NON";
        case COAP_TYPE_ACK: return "ACK";
        case COAP_TYPE_RST: return "RST";
        default: return "Unknown";
    }
}
