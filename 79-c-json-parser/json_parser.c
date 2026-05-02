#include "json.h"
#include <string.h>
#include <ctype.h>
#include <math.h>
#include <errno.h>
#include <limits.h>

static void json_set_error(json_parser_t *parser, json_error_t code, const char *msg, int line, int column)
{
    parser->error = code;
    parser->error_pos.line = line;
    parser->error_pos.column = column;
    snprintf(parser->error_message, sizeof(parser->error_message), "%s", msg);
}

static int json_fill_buffer(json_parser_t *parser)
{
    if (parser->buffer_eof)
        return 0;
    
    parser->buffer_len = fread(parser->buffer, 1, sizeof(parser->buffer), parser->input);
    parser->buffer_pos = 0;
    
    if (parser->buffer_len == 0) {
        parser->buffer_eof = true;
        return 0;
    }
    return 1;
}

static int json_next_char(json_parser_t *parser)
{
    if (parser->lookahead_count > 0) {
        parser->lookahead_count--;
        int ch = parser->lookahead[parser->lookahead_count];
        if (ch == '\n') {
            parser->line++;
            parser->column = 1;
        } else {
            parser->column++;
        }
        return ch;
    }
    
    if (parser->buffer_pos >= parser->buffer_len) {
        if (!json_fill_buffer(parser))
            return EOF;
    }
    
    int ch = parser->buffer[parser->buffer_pos++];
    if (ch == '\n') {
        parser->line++;
        parser->column = 1;
    } else {
        parser->column++;
    }
    return ch;
}

static void json_unget_char(json_parser_t *parser, int ch)
{
    if (ch == EOF)
        return;
    
    if (ch == '\n') {
        parser->line--;
    } else {
        parser->column--;
    }
    
    if (parser->lookahead_count < 4) {
        parser->lookahead[parser->lookahead_count++] = (unsigned char)ch;
    }
}

static int json_peek_char(json_parser_t *parser)
{
    int ch = json_next_char(parser);
    if (ch != EOF)
        json_unget_char(parser, ch);
    return ch;
}

static void json_skip_whitespace(json_parser_t *parser)
{
    int ch;
    while ((ch = json_next_char(parser)) != EOF) {
        if (ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r') {
            json_unget_char(parser, ch);
            break;
        }
    }
}

static json_pos_t json_current_pos(json_parser_t *parser)
{
    json_pos_t pos;
    pos.line = parser->line;
    pos.column = parser->column;
    return pos;
}

static int json_hex_value(int ch)
{
    if (ch >= '0' && ch <= '9') return ch - '0';
    if (ch >= 'a' && ch <= 'f') return 10 + (ch - 'a');
    if (ch >= 'A' && ch <= 'F') return 10 + (ch - 'A');
    return -1;
}

static void json_encode_utf8(unsigned int codepoint, char *out, int *len)
{
    if (codepoint <= 0x7F) {
        out[0] = (char)codepoint;
        *len = 1;
    } else if (codepoint <= 0x7FF) {
        out[0] = 0xC0 | (codepoint >> 6);
        out[1] = 0x80 | (codepoint & 0x3F);
        *len = 2;
    } else if (codepoint <= 0xFFFF) {
        out[0] = 0xE0 | (codepoint >> 12);
        out[1] = 0x80 | ((codepoint >> 6) & 0x3F);
        out[2] = 0x80 | (codepoint & 0x3F);
        *len = 3;
    } else if (codepoint <= 0x10FFFF) {
        out[0] = 0xF0 | (codepoint >> 18);
        out[1] = 0x80 | ((codepoint >> 12) & 0x3F);
        out[2] = 0x80 | ((codepoint >> 6) & 0x3F);
        out[3] = 0x80 | (codepoint & 0x3F);
        *len = 4;
    } else {
        *len = 0;
    }
}

static char *json_grow_string(char *buf, size_t *cap, size_t needed)
{
    if (*cap == 0) *cap = 64;
    while (*cap < needed + 1)
        *cap *= 2;
    return (char *)realloc(buf, *cap);
}

static bool json_check_utf8_continuation(unsigned char ch)
{
    return (ch >= 0x80 && ch <= 0xBF);
}

static bool json_is_valid_utf8(unsigned char *bytes, int len)
{
    int i = 0;
    while (i < len) {
        unsigned char ch = bytes[i];
        if (ch <= 0x7F) {
            i++;
        } else if (ch >= 0xC2 && ch <= 0xDF) {
            if (i + 1 >= len || !json_check_utf8_continuation(bytes[i+1]))
                return false;
            i += 2;
        } else if (ch >= 0xE0 && ch <= 0xEF) {
            if (i + 2 >= len || !json_check_utf8_continuation(bytes[i+1]) ||
                !json_check_utf8_continuation(bytes[i+2]))
                return false;
            if (ch == 0xE0 && bytes[i+1] < 0xA0) return false;
            if (ch == 0xED && bytes[i+1] > 0x9F) return false;
            i += 3;
        } else if (ch >= 0xF0 && ch <= 0xF4) {
            if (i + 3 >= len || !json_check_utf8_continuation(bytes[i+1]) ||
                !json_check_utf8_continuation(bytes[i+2]) || !json_check_utf8_continuation(bytes[i+3]))
                return false;
            if (ch == 0xF0 && bytes[i+1] < 0x90) return false;
            if (ch == 0xF4 && bytes[i+1] > 0x8F) return false;
            i += 4;
        } else {
            return false;
        }
    }
    return true;
}

static json_error_t json_parse_string_content(json_parser_t *parser, char **out_str, size_t *out_len, json_pos_t *start_pos)
{
    char *buf = NULL;
    size_t buf_cap = 0;
    size_t buf_len = 0;
    int ch;
    
    while ((ch = json_next_char(parser)) != EOF) {
        if (ch == '"') {
            if (buf_len > 0) {
                if (!json_is_valid_utf8((unsigned char *)buf, (int)buf_len)) {
                    free(buf);
                    json_set_error(parser, JSON_ERROR_INVALID_UTF8, "invalid UTF-8 byte sequence", start_pos->line, start_pos->column);
                    return JSON_ERROR_INVALID_UTF8;
                }
            }
            buf = json_grow_string(buf, &buf_cap, buf_len);
            buf[buf_len] = '\0';
            *out_str = buf;
            *out_len = buf_len;
            return JSON_OK;
        }
        
        if (ch == '\\') {
            int esc_ch = json_next_char(parser);
            if (esc_ch == EOF) {
                free(buf);
                json_set_error(parser, JSON_ERROR_EOF, "unexpected end of input in escape sequence", parser->line, parser->column);
                return JSON_ERROR_EOF;
            }
            
            char utf8_buf[4];
            int utf8_len = 0;
            unsigned int codepoint = 0;
            
            switch (esc_ch) {
                case 'n': utf8_buf[0] = '\n'; utf8_len = 1; break;
                case 't': utf8_buf[0] = '\t'; utf8_len = 1; break;
                case 'r': utf8_buf[0] = '\r'; utf8_len = 1; break;
                case '\\': utf8_buf[0] = '\\'; utf8_len = 1; break;
                case '"': utf8_buf[0] = '"'; utf8_len = 1; break;
                case '/': utf8_buf[0] = '/'; utf8_len = 1; break;
                case 'b': utf8_buf[0] = '\b'; utf8_len = 1; break;
                case 'f': utf8_buf[0] = '\f'; utf8_len = 1; break;
                case 'u': {
                    int hex_val = 0;
                    for (int i = 0; i < 4; i++) {
                        int h = json_next_char(parser);
                        int hv = json_hex_value(h);
                        if (hv < 0) {
                            free(buf);
                            json_set_error(parser, JSON_ERROR_INVALID_ESCAPE, "invalid hex digit in unicode escape", parser->line, parser->column);
                            return JSON_ERROR_INVALID_ESCAPE;
                        }
                        hex_val = (hex_val << 4) | hv;
                    }
                    codepoint = hex_val;
                    
                    if (codepoint >= 0xD800 && codepoint <= 0xDBFF) {
                        int p1 = json_next_char(parser);
                        int p2 = json_next_char(parser);
                        if (p1 == '\\' && p2 == 'u') {
                            int hex2 = 0;
                            bool valid = true;
                            for (int i = 0; i < 4; i++) {
                                int h = json_next_char(parser);
                                int hv = json_hex_value(h);
                                if (hv < 0) { valid = false; break; }
                                hex2 = (hex2 << 4) | hv;
                            }
                            if (valid && hex2 >= 0xDC00 && hex2 <= 0xDFFF) {
                                codepoint = 0x10000 + ((codepoint - 0xD800) << 10) + (hex2 - 0xDC00);
                                json_encode_utf8(codepoint, utf8_buf, &utf8_len);
                            } else {
                                free(buf);
                                json_set_error(parser, JSON_ERROR_INVALID_UNICODE, "invalid surrogate pair", parser->line, parser->column);
                                return JSON_ERROR_INVALID_UNICODE;
                            }
                        } else {
                            json_unget_char(parser, p2);
                            json_unget_char(parser, p1);
                            json_encode_utf8(codepoint, utf8_buf, &utf8_len);
                        }
                    } else {
                        json_encode_utf8(codepoint, utf8_buf, &utf8_len);
                    }
                    break;
                }
                default:
                    free(buf);
                    json_set_error(parser, JSON_ERROR_INVALID_ESCAPE, "invalid escape sequence", parser->line, parser->column);
                    return JSON_ERROR_INVALID_ESCAPE;
            }
            
            if (utf8_len > 0) {
                buf = json_grow_string(buf, &buf_cap, buf_len + utf8_len);
                memcpy(buf + buf_len, utf8_buf, utf8_len);
                buf_len += utf8_len;
            }
        } else if (ch < 0x20) {
            free(buf);
            json_set_error(parser, JSON_ERROR_INVALID_ESCAPE, "control character in string", parser->line, parser->column);
            return JSON_ERROR_INVALID_ESCAPE;
        } else {
            buf = json_grow_string(buf, &buf_cap, buf_len + 1);
            buf[buf_len++] = (char)ch;
        }
    }
    
    free(buf);
    json_set_error(parser, JSON_ERROR_EOF, "unexpected end of input in string", parser->line, parser->column);
    return JSON_ERROR_EOF;
}

static json_error_t json_parse_string(json_parser_t *parser)
{
    json_pos_t start = json_current_pos(parser);
    int ch = json_next_char(parser);
    if (ch != '"') {
        json_set_error(parser, JSON_ERROR_SYNTAX, "expected string", parser->line, parser->column);
        return JSON_ERROR_SYNTAX;
    }
    
    char *str;
    size_t len;
    json_error_t err = json_parse_string_content(parser, &str, &len, &start);
    if (err != JSON_OK)
        return err;
    
    if (parser->callbacks && parser->callbacks->string)
        parser->callbacks->string(parser, parser->userdata, str, start);
    
    free(str);
    return JSON_OK;
}

static json_error_t json_parse_number(json_parser_t *parser)
{
    json_pos_t start = json_current_pos(parser);
    char num_buf[64];
    int num_len = 0;
    int ch;
    bool is_float = false;
    
    ch = json_next_char(parser);
    if (ch == '-') {
        num_buf[num_len++] = '-';
        ch = json_next_char(parser);
    }
    
    if (ch == '0') {
        num_buf[num_len++] = '0';
        ch = json_next_char(parser);
        if (ch >= '0' && ch <= '9') {
            json_set_error(parser, JSON_ERROR_INVALID_NUMBER, "leading zeros not allowed", parser->line, parser->column);
            return JSON_ERROR_INVALID_NUMBER;
        }
    } else if (ch >= '1' && ch <= '9') {
        num_buf[num_len++] = ch;
        while ((ch = json_next_char(parser)) != EOF && ch >= '0' && ch <= '9') {
            if (num_len < 63)
                num_buf[num_len++] = ch;
        }
    } else {
        json_set_error(parser, JSON_ERROR_INVALID_NUMBER, "expected digit", parser->line, parser->column);
        return JSON_ERROR_INVALID_NUMBER;
    }
    
    if (ch == '.') {
        is_float = true;
        if (num_len < 63)
            num_buf[num_len++] = '.';
        ch = json_next_char(parser);
        if (ch < '0' || ch > '9') {
            json_set_error(parser, JSON_ERROR_INVALID_NUMBER, "expected digit after decimal point", parser->line, parser->column);
            return JSON_ERROR_INVALID_NUMBER;
        }
        while (ch >= '0' && ch <= '9') {
            if (num_len < 63)
                num_buf[num_len++] = ch;
            ch = json_next_char(parser);
        }
    }
    
    if (ch == 'e' || ch == 'E') {
        is_float = true;
        if (num_len < 63)
            num_buf[num_len++] = 'e';
        ch = json_next_char(parser);
        if (ch == '+' || ch == '-') {
            if (num_len < 63)
                num_buf[num_len++] = ch;
            ch = json_next_char(parser);
        }
        if (ch < '0' || ch > '9') {
            json_set_error(parser, JSON_ERROR_INVALID_NUMBER, "expected digit in exponent", parser->line, parser->column);
            return JSON_ERROR_INVALID_NUMBER;
        }
        while (ch >= '0' && ch <= '9') {
            if (num_len < 63)
                num_buf[num_len++] = ch;
            ch = json_next_char(parser);
        }
    }
    
    if (ch != EOF)
        json_unget_char(parser, ch);
    
    num_buf[num_len] = '\0';
    
    if (is_float) {
        double val = strtod(num_buf, NULL);
        if (parser->callbacks && parser->callbacks->number_float)
            parser->callbacks->number_float(parser, parser->userdata, val, start);
    } else {
        errno = 0;
        char *endptr;
        long long val = strtoll(num_buf, &endptr, 10);
        if (errno == ERANGE) {
            double dval = strtod(num_buf, NULL);
            if (parser->callbacks && parser->callbacks->number_float)
                parser->callbacks->number_float(parser, parser->userdata, dval, start);
        } else {
            if (parser->callbacks && parser->callbacks->number_int)
                parser->callbacks->number_int(parser, parser->userdata, val, start);
        }
    }
    
    return JSON_OK;
}

static json_error_t json_parse_literal(json_parser_t *parser, const char *expected, bool *out_bool)
{
    int len = (int)strlen(expected);
    json_pos_t start = json_current_pos(parser);
    
    for (int i = 0; i < len; i++) {
        int ch = json_next_char(parser);
        if (ch == EOF) {
            json_set_error(parser, JSON_ERROR_EOF, "unexpected end of input", parser->line, parser->column);
            return JSON_ERROR_EOF;
        }
        if (ch != expected[i]) {
            json_set_error(parser, JSON_ERROR_SYNTAX, "unexpected character", parser->line, parser->column);
            return JSON_ERROR_SYNTAX;
        }
    }
    
    if (strcmp(expected, "true") == 0) {
        *out_bool = true;
        if (parser->callbacks && parser->callbacks->boolean)
            parser->callbacks->boolean(parser, parser->userdata, true, start);
    } else if (strcmp(expected, "false") == 0) {
        *out_bool = false;
        if (parser->callbacks && parser->callbacks->boolean)
            parser->callbacks->boolean(parser, parser->userdata, false, start);
    } else if (strcmp(expected, "null") == 0) {
        if (parser->callbacks && parser->callbacks->null)
            parser->callbacks->null(parser, parser->userdata, start);
    }
    return JSON_OK;
}

static json_error_t json_parse_value(json_parser_t *parser);

static json_error_t json_parse_array(json_parser_t *parser)
{
    json_pos_t start = json_current_pos(parser);
    int ch = json_next_char(parser);
    if (ch != '[') {
        json_set_error(parser, JSON_ERROR_SYNTAX, "expected '['", parser->line, parser->column);
        return JSON_ERROR_SYNTAX;
    }
    
    if (parser->callbacks && parser->callbacks->array_start)
        parser->callbacks->array_start(parser, parser->userdata, start);
    
    json_skip_whitespace(parser);
    ch = json_peek_char(parser);
    
    if (ch == ']') {
        json_next_char(parser);
    } else {
        while (1) {
            json_error_t err = json_parse_value(parser);
            if (err != JSON_OK)
                return err;
            
            json_skip_whitespace(parser);
            ch = json_next_char(parser);
            
            if (ch == ']')
                break;
            if (ch != ',') {
                json_set_error(parser, JSON_ERROR_SYNTAX, "expected ']' or ','", parser->line, parser->column);
                return JSON_ERROR_SYNTAX;
            }
            json_skip_whitespace(parser);
        }
    }
    
    json_pos_t end = json_current_pos(parser);
    if (parser->callbacks && parser->callbacks->array_end)
        parser->callbacks->array_end(parser, parser->userdata, end);
    
    return JSON_OK;
}

static json_error_t json_parse_object(json_parser_t *parser)
{
    json_pos_t start = json_current_pos(parser);
    int ch = json_next_char(parser);
    if (ch != '{') {
        json_set_error(parser, JSON_ERROR_SYNTAX, "expected '{'", parser->line, parser->column);
        return JSON_ERROR_SYNTAX;
    }
    
    if (parser->callbacks && parser->callbacks->object_start)
        parser->callbacks->object_start(parser, parser->userdata, start);
    
    json_skip_whitespace(parser);
    ch = json_peek_char(parser);
    
    if (ch == '}') {
        json_next_char(parser);
    } else {
        while (1) {
            json_pos_t key_start = json_current_pos(parser);
            ch = json_next_char(parser);
            if (ch != '"') {
                json_set_error(parser, JSON_ERROR_SYNTAX, "expected string key", parser->line, parser->column);
                return JSON_ERROR_SYNTAX;
            }
            
            char *key;
            size_t key_len;
            json_error_t err = json_parse_string_content(parser, &key, &key_len, &key_start);
            if (err != JSON_OK)
                return err;
            
            if (parser->callbacks && parser->callbacks->key)
                parser->callbacks->key(parser, parser->userdata, key, key_start);
            
            free(key);
            
            json_skip_whitespace(parser);
            ch = json_next_char(parser);
            if (ch != ':') {
                json_set_error(parser, JSON_ERROR_SYNTAX, "expected ':'", parser->line, parser->column);
                return JSON_ERROR_SYNTAX;
            }
            json_skip_whitespace(parser);
            
            err = json_parse_value(parser);
            if (err != JSON_OK)
                return err;
            
            json_skip_whitespace(parser);
            ch = json_next_char(parser);
            
            if (ch == '}')
                break;
            if (ch != ',') {
                json_set_error(parser, JSON_ERROR_SYNTAX, "expected '}' or ','", parser->line, parser->column);
                return JSON_ERROR_SYNTAX;
            }
            json_skip_whitespace(parser);
        }
    }
    
    json_pos_t end = json_current_pos(parser);
    if (parser->callbacks && parser->callbacks->object_end)
        parser->callbacks->object_end(parser, parser->userdata, end);
    
    return JSON_OK;
}

static json_error_t json_parse_value(json_parser_t *parser)
{
    json_skip_whitespace(parser);
    int ch = json_peek_char(parser);
    
    if (ch == '{')
        return json_parse_object(parser);
    if (ch == '[')
        return json_parse_array(parser);
    if (ch == '"')
        return json_parse_string(parser);
    if (ch == 't') {
        bool tmp;
        return json_parse_literal(parser, "true", &tmp);
    }
    if (ch == 'f') {
        bool tmp;
        return json_parse_literal(parser, "false", &tmp);
    }
    if (ch == 'n') {
        bool tmp;
        return json_parse_literal(parser, "null", &tmp);
    }
    if (ch == '-' || (ch >= '0' && ch <= '9'))
        return json_parse_number(parser);
    
    if (ch == EOF) {
        json_set_error(parser, JSON_ERROR_EOF, "unexpected end of input", parser->line, parser->column);
        return JSON_ERROR_EOF;
    }
    
    json_set_error(parser, JSON_ERROR_SYNTAX, "unexpected character", parser->line, parser->column);
    return JSON_ERROR_SYNTAX;
}

void json_parser_init(json_parser_t *parser, FILE *input, const json_sax_callbacks_t *callbacks, void *userdata)
{
    memset(parser, 0, sizeof(*parser));
    parser->input = input;
    parser->line = 1;
    parser->column = 1;
    parser->callbacks = callbacks;
    parser->userdata = userdata;
    parser->error = JSON_OK;
    parser->error_pos.line = 1;
    parser->error_pos.column = 1;
}

json_error_t json_parse_sax(json_parser_t *parser)
{
    json_error_t err = json_parse_value(parser);
    if (err != JSON_OK)
        return err;
    
    json_skip_whitespace(parser);
    int ch = json_next_char(parser);
    if (ch != EOF) {
        json_set_error(parser, JSON_ERROR_TRAILING_CONTENT, "unexpected trailing content", parser->line, parser->column);
        return JSON_ERROR_TRAILING_CONTENT;
    }
    
    return JSON_OK;
}

const char *json_error_message(json_parser_t *parser)
{
    return parser->error_message;
}

json_pos_t json_error_position(json_parser_t *parser)
{
    return parser->error_pos;
}
