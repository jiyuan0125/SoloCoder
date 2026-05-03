#ifndef COAP_H
#define COAP_H

#include <stdint.h>
#include <stddef.h>

#define COAP_VERSION 1
#define COAP_MAX_TOKEN_LEN 8
#define COAP_MAX_PAYLOAD_SIZE 512

typedef enum {
    COAP_TYPE_CON = 0,
    COAP_TYPE_NON = 1,
    COAP_TYPE_ACK = 2,
    COAP_TYPE_RST = 3
} coap_type_t;

typedef enum {
    COAP_CODE_EMPTY = 0,
    COAP_CODE_GET = 1,
    COAP_CODE_POST = 2,
    COAP_CODE_PUT = 3,
    COAP_CODE_DELETE = 4,
    COAP_CODE_201_CREATED = 65,
    COAP_CODE_202_DELETED = 66,
    COAP_CODE_203_VALID = 67,
    COAP_CODE_204_CHANGED = 68,
    COAP_CODE_205_CONTENT = 69,
    COAP_CODE_400_BAD_REQUEST = 128,
    COAP_CODE_404_NOT_FOUND = 132,
    COAP_CODE_405_METHOD_NOT_ALLOWED = 133,
    COAP_CODE_500_INTERNAL_SERVER_ERROR = 160
} coap_code_t;

typedef enum {
    COAP_OPTION_IF_MATCH = 1,
    COAP_OPTION_URI_HOST = 3,
    COAP_OPTION_ETAG = 4,
    COAP_OPTION_IF_NONE_MATCH = 5,
    COAP_OPTION_OBSERVE = 6,
    COAP_OPTION_URI_PORT = 7,
    COAP_OPTION_LOCATION_PATH = 8,
    COAP_OPTION_URI_PATH = 11,
    COAP_OPTION_CONTENT_FORMAT = 12,
    COAP_OPTION_MAX_AGE = 14,
    COAP_OPTION_URI_QUERY = 15,
    COAP_OPTION_ACCEPT = 17,
    COAP_OPTION_LOCATION_QUERY = 20,
    COAP_OPTION_PROXY_URI = 35,
    COAP_OPTION_PROXY_SCHEME = 39,
    COAP_OPTION_SIZE1 = 60
} coap_option_num_t;

typedef enum {
    COAP_CONTENT_FORMAT_TEXT_PLAIN = 0,
    COAP_CONTENT_FORMAT_APPLICATION_LINK_FORMAT = 40,
    COAP_CONTENT_FORMAT_APPLICATION_XML = 41,
    COAP_CONTENT_FORMAT_APPLICATION_OCTET_STREAM = 42,
    COAP_CONTENT_FORMAT_APPLICATION_EXI = 47,
    COAP_CONTENT_FORMAT_APPLICATION_JSON = 50
} coap_content_format_t;

typedef struct {
    uint16_t delta;
    uint16_t length;
    uint8_t *value;
} coap_option_t;

typedef struct {
    uint8_t version;
    coap_type_t type;
    uint8_t token_len;
    coap_code_t code;
    uint16_t msg_id;
    uint8_t token[COAP_MAX_TOKEN_LEN];
    coap_option_t *options;
    uint8_t option_count;
    uint8_t *payload;
    size_t payload_len;
} coap_message_t;

void coap_message_init(coap_message_t *msg);
void coap_message_free(coap_message_t *msg);

int coap_message_encode(const coap_message_t *msg, uint8_t *buffer, size_t buffer_size);
int coap_message_decode(coap_message_t *msg, const uint8_t *buffer, size_t buffer_size);

coap_option_t *coap_message_add_option(coap_message_t *msg, uint16_t number, const uint8_t *value, size_t length);
coap_option_t *coap_message_get_option(const coap_message_t *msg, uint16_t number);
coap_option_t *coap_message_get_option_by_index(const coap_message_t *msg, uint8_t index);

int coap_message_set_payload(coap_message_t *msg, const uint8_t *payload, size_t length);
const char *coap_code_to_string(coap_code_t code);
const char *coap_type_to_string(coap_type_t type);

#endif
