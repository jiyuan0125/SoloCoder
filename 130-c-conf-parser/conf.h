#ifndef CONF_H
#define CONF_H

#include <stddef.h>
#include <stdbool.h>
#include <time.h>

#ifdef __cplusplus
extern "C" {
#endif

#define CONF_MAX_KEY_LEN    256
#define CONF_MAX_VALUE_LEN  4096
#define CONF_MAX_SECTIONS   256
#define CONF_MAX_ITEMS      1024
#define CONF_MAX_RECURSION  16

typedef enum {
    CONF_TYPE_UNKNOWN = 0,
    CONF_TYPE_STRING,
    CONF_TYPE_INT,
    CONF_TYPE_FLOAT,
    CONF_TYPE_BOOL
} ConfValueType;

typedef struct {
    char            key[CONF_MAX_KEY_LEN];
    ConfValueType   type;
    union {
        char        str_val[CONF_MAX_VALUE_LEN];
        long        int_val;
        double      float_val;
        bool        bool_val;
    } value;
} ConfItem;

typedef struct {
    char        name[CONF_MAX_KEY_LEN];
    ConfItem    items[CONF_MAX_ITEMS];
    size_t      item_count;
} ConfSection;

typedef struct {
    ConfSection sections[CONF_MAX_SECTIONS];
    size_t      section_count;
    char        filepath[512];
    time_t      mtime;
    char        last_error[256];
} ConfContext;

typedef enum {
    CONF_OK = 0,
    CONF_ERR_FILE_NOT_FOUND,
    CONF_ERR_FILE_READ,
    CONF_ERR_SYNTAX,
    CONF_ERR_SECTION_NOT_FOUND,
    CONF_ERR_KEY_NOT_FOUND,
    CONF_ERR_TYPE_MISMATCH,
    CONF_ERR_CIRCULAR_REF,
    CONF_ERR_MEMORY,
    CONF_ERR_INVALID_FORMAT
} ConfError;

ConfContext* conf_create(void);
void conf_destroy(ConfContext* ctx);

ConfError conf_load_file(ConfContext* ctx, const char* filepath);
ConfError conf_reload(ConfContext* ctx);
bool conf_needs_reload(ConfContext* ctx);

const char* conf_get_string(ConfContext* ctx, const char* key, const char* default_val);
long conf_get_int(ConfContext* ctx, const char* key, long default_val);
double conf_get_float(ConfContext* ctx, const char* key, double default_val);
bool conf_get_bool(ConfContext* ctx, const char* key, bool default_val);

ConfError conf_get_string_ex(ConfContext* ctx, const char* section, const char* key, 
                              const char* default_val, char* out_buf, size_t buf_size);
ConfError conf_get_int_ex(ConfContext* ctx, const char* section, const char* key, 
                           long default_val, long* out_val);
ConfError conf_get_float_ex(ConfContext* ctx, const char* section, const char* key, 
                             double default_val, double* out_val);
ConfError conf_get_bool_ex(ConfContext* ctx, const char* section, const char* key, 
                            bool default_val, bool* out_val);

void conf_set_string(ConfContext* ctx, const char* section, const char* key, const char* value);
void conf_set_int(ConfContext* ctx, const char* section, const char* key, long value);
void conf_set_float(ConfContext* ctx, const char* section, const char* key, double value);
void conf_set_bool(ConfContext* ctx, const char* section, const char* key, bool value);

ConfSection* conf_find_section(ConfContext* ctx, const char* name);
ConfItem* conf_find_item(ConfSection* section, const char* key);
ConfSection* conf_get_or_create_section(ConfContext* ctx, const char* name);

const char* conf_error_string(ConfError err);
const char* conf_last_error(ConfContext* ctx);

ConfError conf_resolve_references(ConfContext* ctx, const char* section, const char* input, 
                                   char* output, size_t output_size, int recursion_depth);

ConfValueType conf_infer_type(const char* value);
void conf_parse_value(ConfItem* item, const char* raw_value);

#ifdef __cplusplus
}
#endif

#endif
