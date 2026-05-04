#include "dns_protocol.h"
#include <string.h>
#include <arpa/inet.h>
#include <stdio.h>

static inline uint16_t read_uint16(const uint8_t *buf) {
    return (buf[0] << 8) | buf[1];
}

static inline uint32_t read_uint32(const uint8_t *buf) {
    return (buf[0] << 24) | (buf[1] << 16) | (buf[2] << 8) | buf[3];
}

static inline void write_uint16(uint8_t *buf, uint16_t val) {
    buf[0] = (val >> 8) & 0xFF;
    buf[1] = val & 0xFF;
}

static inline void write_uint32(uint8_t *buf, uint32_t val) {
    buf[0] = (val >> 24) & 0xFF;
    buf[1] = (val >> 16) & 0xFF;
    buf[2] = (val >> 8) & 0xFF;
    buf[3] = val & 0xFF;
}

int dns_parse_header(const uint8_t *packet, size_t packet_len, dns_header_t *header) {
    if (packet_len < DNS_HEADER_SIZE) {
        return -1;
    }

    header->id = read_uint16(packet);
    header->flags = read_uint16(packet + 2);
    header->qdcount = read_uint16(packet + 4);
    header->ancount = read_uint16(packet + 6);
    header->nscount = read_uint16(packet + 8);
    header->arcount = read_uint16(packet + 10);

    return 0;
}

int dns_construct_header(uint8_t *packet, size_t packet_size, const dns_header_t *header) {
    if (packet_size < DNS_HEADER_SIZE) {
        return -1;
    }

    write_uint16(packet, header->id);
    write_uint16(packet + 2, header->flags);
    write_uint16(packet + 4, header->qdcount);
    write_uint16(packet + 6, header->ancount);
    write_uint16(packet + 8, header->nscount);
    write_uint16(packet + 10, header->arcount);

    return 0;
}

static int check_pointer_loop(size_t ptr_history[], int ptr_count, size_t offset) {
    for (int i = 0; i < ptr_count; i++) {
        if (ptr_history[i] == offset) {
            return 1;
        }
    }
    return 0;
}

static int dns_parse_name_recursive(const uint8_t *packet, size_t packet_len,
                                      size_t *offset, char *name_buf, size_t name_buf_size,
                                      size_t ptr_history[], int ptr_count) {
    if (ptr_count >= 10) {
        return -1;
    }

    size_t pos = *offset;
    size_t name_pos = 0;
    int first = 1;

    while (pos < packet_len) {
        uint8_t len = packet[pos];

        if (len == 0) {
            pos++;
            if (name_pos < name_buf_size) {
                name_buf[name_pos] = '\0';
            }
            *offset = pos;
            return 0;
        }

        if ((len & 0xC0) == 0xC0) {
            size_t ptr_offset = ((len & 0x3F) << 8) | packet[pos + 1];

            if (check_pointer_loop(ptr_history, ptr_count, ptr_offset)) {
                return -1;
            }

            ptr_history[ptr_count] = ptr_offset;
            *offset = ptr_offset;

            int result = dns_parse_name_recursive(packet, packet_len, offset,
                                                    name_buf + name_pos,
                                                    name_buf_size - name_pos,
                                                    ptr_history, ptr_count + 1);

            if (result == 0 && name_pos > 0) {
            }

            *offset = pos + 2;
            return result;
        }

        if (len > DNS_MAX_LABEL_LENGTH) {
            return -1;
        }

        pos++;

        if (first) {
            first = 0;
        } else {
            if (name_pos + 1 >= name_buf_size) {
                return -1;
            }
            name_buf[name_pos++] = '.';
        }

        if (name_pos + len >= name_buf_size) {
            return -1;
        }

        memcpy(name_buf + name_pos, packet + pos, len);
        name_pos += len;
        pos += len;
    }

    return -1;
}

int dns_parse_name(const uint8_t *packet, size_t packet_len, size_t *offset,
                    char *name_buf, size_t name_buf_size) {
    size_t ptr_history[10];
    memset(ptr_history, 0xFF, sizeof(ptr_history));

    return dns_parse_name_recursive(packet, packet_len, offset, name_buf, name_buf_size,
                                     ptr_history, 0);
}

int dns_construct_name(uint8_t *packet, size_t packet_size, size_t *offset,
                        const char *name) {
    size_t pos = *offset;
    const char *start = name;

    while (*start) {
        const char *dot = strchr(start, '.');
        size_t label_len;

        if (dot) {
            label_len = dot - start;
        } else {
            label_len = strlen(start);
        }

        if (label_len > DNS_MAX_LABEL_LENGTH || label_len == 0) {
            return -1;
        }

        if (pos + label_len + 2 > packet_size) {
            return -1;
        }

        packet[pos++] = (uint8_t)label_len;
        memcpy(packet + pos, start, label_len);
        pos += label_len;

        if (dot) {
            start = dot + 1;
        } else {
            break;
        }
    }

    if (pos + 1 > packet_size) {
        return -1;
    }

    packet[pos++] = 0;
    *offset = pos;
    return 0;
}

int dns_parse_question(const uint8_t *packet, size_t packet_len, size_t *offset,
                        dns_question_t *question) {
    int result = dns_parse_name(packet, packet_len, offset,
                                  question->name, sizeof(question->name));
    if (result != 0) {
        return -1;
    }

    if (*offset + 4 > packet_len) {
        return -1;
    }

    question->qtype = read_uint16(packet + *offset);
    *offset += 2;

    question->qclass = read_uint16(packet + *offset);
    *offset += 2;

    return 0;
}

int dns_construct_question(uint8_t *packet, size_t packet_size, size_t *offset,
                            const dns_question_t *question) {
    int result = dns_construct_name(packet, packet_size, offset, question->name);
    if (result != 0) {
        return -1;
    }

    if (*offset + 4 > packet_size) {
        return -1;
    }

    write_uint16(packet + *offset, question->qtype);
    *offset += 2;

    write_uint16(packet + *offset, question->qclass);
    *offset += 2;

    return 0;
}

int dns_parse_rr(const uint8_t *packet, size_t packet_len, size_t *offset,
                  dns_rr_t *rr) {
    int result = dns_parse_name(packet, packet_len, offset,
                                  rr->name, sizeof(rr->name));
    if (result != 0) {
        return -1;
    }

    if (*offset + 10 > packet_len) {
        return -1;
    }

    rr->type = read_uint16(packet + *offset);
    *offset += 2;

    rr->class = read_uint16(packet + *offset);
    *offset += 2;

    rr->ttl = read_uint32(packet + *offset);
    *offset += 4;

    rr->rdlength = read_uint16(packet + *offset);
    *offset += 2;

    if (*offset + rr->rdlength > packet_len) {
        return -1;
    }

    memset(rr->rdata, 0, sizeof(rr->rdata));
    memset(rr->rdata_name, 0, sizeof(rr->rdata_name));
    rr->mx_preference = 0;

    if (rr->rdlength > 0 && rr->rdlength < sizeof(rr->rdata)) {
        memcpy(rr->rdata, packet + *offset, rr->rdlength);

        size_t rdata_offset = *offset;

        switch (rr->type) {
            case DNS_TYPE_CNAME:
            case DNS_TYPE_NS:
            case DNS_TYPE_PTR:
                if (dns_parse_name(packet, packet_len, &rdata_offset,
                                    rr->rdata_name, sizeof(rr->rdata_name)) != 0) {
                    rr->rdata_name[0] = '\0';
                }
                break;

            case DNS_TYPE_MX:
                if (rr->rdlength >= 2) {
                    rr->mx_preference = read_uint16(packet + rdata_offset);
                    rdata_offset += 2;
                    if (dns_parse_name(packet, packet_len, &rdata_offset,
                                        rr->rdata_name, sizeof(rr->rdata_name)) != 0) {
                        rr->rdata_name[0] = '\0';
                    }
                }
                break;

            default:
                rr->rdata_name[0] = '\0';
                break;
        }
    }

    *offset += rr->rdlength;

    return 0;
}

int dns_construct_query(uint8_t *packet, size_t packet_size, uint16_t id,
                         const char *name, uint16_t qtype, uint16_t qclass) {
    dns_header_t header;
    memset(&header, 0, sizeof(header));

    header.id = id;
    header.flags = DNS_FLAG_RD;
    header.qdcount = 1;

    size_t offset = 0;
    if (dns_construct_header(packet, packet_size, &header) != 0) {
        return -1;
    }

    offset = DNS_HEADER_SIZE;

    dns_question_t question;
    memset(&question, 0, sizeof(question));
    strncpy(question.name, name, sizeof(question.name) - 1);
    question.qtype = qtype;
    question.qclass = qclass;

    if (dns_construct_question(packet, packet_size, &offset, &question) != 0) {
        return -1;
    }

    return (int)offset;
}

int dns_parse_response(const uint8_t *packet, size_t packet_len, dns_message_t *msg) {
    memset(msg, 0, sizeof(*msg));

    if (dns_parse_header(packet, packet_len, &msg->header) != 0) {
        return -1;
    }

    size_t offset = DNS_HEADER_SIZE;

    uint16_t qdcount = msg->header.qdcount;
    if (qdcount > 0) {
        if (dns_parse_question(packet, packet_len, &offset, &msg->question) != 0) {
            return -1;
        }
    }

    uint16_t ancount = msg->header.ancount;
    msg->answer_count = 0;

    for (int i = 0; i < ancount && i < 128; i++) {
        if (dns_parse_rr(packet, packet_len, &offset, &msg->answers[i]) != 0) {
            return -1;
        }
        msg->answer_count++;
    }

    return 0;
}

int dns_type_to_string(uint16_t type, char *buf, size_t buf_size) {
    const char *name = NULL;

    switch (type) {
        case DNS_TYPE_A: name = "A"; break;
        case DNS_TYPE_NS: name = "NS"; break;
        case DNS_TYPE_CNAME: name = "CNAME"; break;
        case DNS_TYPE_SOA: name = "SOA"; break;
        case DNS_TYPE_PTR: name = "PTR"; break;
        case DNS_TYPE_MX: name = "MX"; break;
        case DNS_TYPE_TXT: name = "TXT"; break;
        case DNS_TYPE_AAAA: name = "AAAA"; break;
        default:
            if (buf_size > 10) {
                snprintf(buf, buf_size, "TYPE%u", type);
                return 0;
            }
            return -1;
    }

    if (name && buf_size > strlen(name)) {
        strcpy(buf, name);
        return 0;
    }

    return -1;
}

int dns_string_to_type(const char *str, uint16_t *type) {
    if (strcasecmp(str, "A") == 0) *type = DNS_TYPE_A;
    else if (strcasecmp(str, "NS") == 0) *type = DNS_TYPE_NS;
    else if (strcasecmp(str, "CNAME") == 0) *type = DNS_TYPE_CNAME;
    else if (strcasecmp(str, "SOA") == 0) *type = DNS_TYPE_SOA;
    else if (strcasecmp(str, "PTR") == 0) *type = DNS_TYPE_PTR;
    else if (strcasecmp(str, "MX") == 0) *type = DNS_TYPE_MX;
    else if (strcasecmp(str, "TXT") == 0) *type = DNS_TYPE_TXT;
    else if (strcasecmp(str, "AAAA") == 0) *type = DNS_TYPE_AAAA;
    else return -1;

    return 0;
}

void dns_format_a_record(const dns_rr_t *rr, char *buf, size_t buf_size) {
    if (rr->rdlength != 4 || buf_size < 16) {
        if (buf_size > 0) buf[0] = '\0';
        return;
    }

    snprintf(buf, buf_size, "%u.%u.%u.%u",
             rr->rdata[0], rr->rdata[1], rr->rdata[2], rr->rdata[3]);
}

void dns_format_aaaa_record(const dns_rr_t *rr, char *buf, size_t buf_size) {
    if (rr->rdlength != 16 || buf_size < 40) {
        if (buf_size > 0) buf[0] = '\0';
        return;
    }

    char tmp[INET6_ADDRSTRLEN];
    const struct in6_addr *addr = (const struct in6_addr *)rr->rdata;

    if (inet_ntop(AF_INET6, addr, tmp, sizeof(tmp)) == NULL) {
        if (buf_size > 0) buf[0] = '\0';
        return;
    }

    strncpy(buf, tmp, buf_size - 1);
    buf[buf_size - 1] = '\0';
}

void dns_format_mx_record(const dns_rr_t *rr, const uint8_t *packet, size_t packet_len,
                           char *buf, size_t buf_size) {
    (void)packet;
    (void)packet_len;

    if (buf_size < 32) {
        if (buf_size > 0) buf[0] = '\0';
        return;
    }

    if (rr->rdata_name[0] != '\0') {
        snprintf(buf, buf_size, "%d %s", rr->mx_preference, rr->rdata_name);
    } else {
        snprintf(buf, buf_size, "%d", rr->mx_preference);
    }
}

void dns_format_cname_record(const dns_rr_t *rr, const uint8_t *packet, size_t packet_len,
                              char *buf, size_t buf_size) {
    (void)packet;
    (void)packet_len;

    if (buf_size < 2) {
        if (buf_size > 0) buf[0] = '\0';
        return;
    }

    if (rr->rdata_name[0] != '\0') {
        strncpy(buf, rr->rdata_name, buf_size - 1);
        buf[buf_size - 1] = '\0';
    } else {
        buf[0] = '\0';
    }
}

void dns_format_ns_record(const dns_rr_t *rr, const uint8_t *packet, size_t packet_len,
                           char *buf, size_t buf_size) {
    dns_format_cname_record(rr, packet, packet_len, buf, buf_size);
}
