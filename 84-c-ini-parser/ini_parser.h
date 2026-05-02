#ifndef INI_PARSER_H
#define INI_PARSER_H

#ifdef __cplusplus
extern "C" {
#endif

#define INI_MAX_INCLUDE_DEPTH 10

typedef struct ini_error_t {
    int line_number;
    char message[256];
} ini_error_t;

typedef struct ini_kv_t {
    char* key;
    char* value;
    char* raw_line;
    int modified;
} ini_kv_t;

typedef struct ini_section_t {
    char* name;
    char* raw_header;
    ini_kv_t* entries;
    int entry_count;
    int entry_capacity;
    char** trailing_lines;
    int trailing_count;
    int trailing_capacity;
} ini_section_t;

typedef struct ini_parser_t {
    ini_section_t* sections;
    int section_count;
    int section_capacity;
    char** leading_lines;
    int leading_count;
    int leading_capacity;
    ini_error_t last_error;
} ini_parser_t;

ini_parser_t* ini_parse(const char* path);
void ini_free(ini_parser_t* parser);

const char* ini_get(const ini_parser_t* parser, const char* section, const char* key, const char* default_value);
const char* ini_get_string(const ini_parser_t* parser, const char* section, const char* key, const char* default_value);
int ini_get_int(const ini_parser_t* parser, const char* section, const char* key, int default_value);
int ini_get_bool(const ini_parser_t* parser, const char* section, const char* key, int default_value);

int ini_set(ini_parser_t* parser, const char* section, const char* key, const char* value);
int ini_save(const ini_parser_t* parser, const char* path);

const char* ini_get_error_message(const ini_parser_t* parser);
int ini_get_error_line(const ini_parser_t* parser);

#ifdef __cplusplus
}
#endif

#endif /* INI_PARSER_H */
