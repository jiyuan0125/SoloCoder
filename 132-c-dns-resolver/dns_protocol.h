#ifndef DNS_PROTOCOL_H
#define DNS_PROTOCOL_H

#include <stdint.h>
#include <stddef.h>

#define DNS_PORT 53
#define DNS_HEADER_SIZE 12
#define DNS_MAX_PACKET_SIZE 512
#define DNS_MAX_NAME_LENGTH 255
#define DNS_MAX_LABEL_LENGTH 63

#define DNS_QR_QUERY 0
#define DNS_QR_RESPONSE 1

#define DNS_OPCODE_QUERY 0
#define DNS_OPCODE_IQUERY 1
#define DNS_OPCODE_STATUS 2

#define DNS_RCODE_NO_ERROR 0
#define DNS_RCODE_FORMAT_ERROR 1
#define DNS_RCODE_SERVER_FAILURE 2
#define DNS_RCODE_NAME_ERROR 3
#define DNS_RCODE_NOT_IMPLEMENTED 4
#define DNS_RCODE_REFUSED 5

#define DNS_TYPE_A 1
#define DNS_TYPE_NS 2
#define DNS_TYPE_CNAME 5
#define DNS_TYPE_SOA 6
#define DNS_TYPE_PTR 12
#define DNS_TYPE_MX 15
#define DNS_TYPE_TXT 16
#define DNS_TYPE_AAAA 28

#define DNS_CLASS_IN 1

#define DNS_FLAG_QR (1 << 15)
#define DNS_FLAG_OPCODE_SHIFT 11
#define DNS_FLAG_OPCODE_MASK (0xF << 11)
#define DNS_FLAG_AA (1 << 10)
#define DNS_FLAG_TC (1 << 9)
#define DNS_FLAG_RD (1 << 8)
#define DNS_FLAG_RA (1 << 7)
#define DNS_FLAG_Z (7 << 4)
#define DNS_FLAG_RCODE_MASK 0xF

typedef struct {
    uint16_t id;
    uint16_t flags;
    uint16_t qdcount;
    uint16_t ancount;
    uint16_t nscount;
    uint16_t arcount;
} dns_header_t;

typedef struct {
    char name[DNS_MAX_NAME_LENGTH];
    uint16_t qtype;
    uint16_t qclass;
} dns_question_t;

typedef struct {
    char name[DNS_MAX_NAME_LENGTH];
    uint16_t type;
    uint16_t class;
    uint32_t ttl;
    uint16_t rdlength;
    uint8_t rdata[512];
    char rdata_name[DNS_MAX_NAME_LENGTH];
    uint16_t mx_preference;
} dns_rr_t;

typedef struct {
    dns_header_t header;
    dns_question_t question;
    dns_rr_t answers[128];
    int answer_count;
} dns_message_t;

int dns_parse_header(const uint8_t *packet, size_t packet_len, dns_header_t *header);
int dns_construct_header(uint8_t *packet, size_t packet_size, const dns_header_t *header);
int dns_parse_name(const uint8_t *packet, size_t packet_len, size_t *offset,
                    char *name_buf, size_t name_buf_size);
int dns_construct_name(uint8_t *packet, size_t packet_size, size_t *offset,
                        const char *name);
int dns_parse_question(const uint8_t *packet, size_t packet_len, size_t *offset,
                        dns_question_t *question);
int dns_construct_question(uint8_t *packet, size_t packet_size, size_t *offset,
                            const dns_question_t *question);
int dns_parse_rr(const uint8_t *packet, size_t packet_len, size_t *offset,
                  dns_rr_t *rr);
int dns_construct_query(uint8_t *packet, size_t packet_size, uint16_t id,
                         const char *name, uint16_t qtype, uint16_t qclass);
int dns_parse_response(const uint8_t *packet, size_t packet_len, dns_message_t *msg);

int dns_type_to_string(uint16_t type, char *buf, size_t buf_size);
int dns_string_to_type(const char *str, uint16_t *type);
void dns_format_a_record(const dns_rr_t *rr, char *buf, size_t buf_size);
void dns_format_aaaa_record(const dns_rr_t *rr, char *buf, size_t buf_size);
void dns_format_mx_record(const dns_rr_t *rr, const uint8_t *packet, size_t packet_len,
                           char *buf, size_t buf_size);
void dns_format_cname_record(const dns_rr_t *rr, const uint8_t *packet, size_t packet_len,
                              char *buf, size_t buf_size);
void dns_format_ns_record(const dns_rr_t *rr, const uint8_t *packet, size_t packet_len,
                           char *buf, size_t buf_size);

#endif
