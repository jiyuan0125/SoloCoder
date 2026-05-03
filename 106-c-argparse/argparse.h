#ifndef ARGPARSE_H
#define ARGPARSE_H

#include <stddef.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    AP_TYPE_BOOL,
    AP_TYPE_STRING,
    AP_TYPE_INT,
    AP_TYPE_FLOAT,
    AP_TYPE_POSITIONAL
} ap_arg_type_t;

typedef enum {
    AP_DUP_LAST,
    AP_DUP_ERROR,
    AP_DUP_ACCUMULATE
} ap_dup_policy_t;

typedef struct ap_range_t {
    double min;
    double max;
    bool has_min;
    bool has_max;
} ap_range_t;

typedef struct ap_arg_def_t {
    const char *name;
    char short_name;
    ap_arg_type_t type;
    const char *description;
    const char *default_value;
    bool required;
    ap_range_t range;
    ap_dup_policy_t dup_policy;
} ap_arg_def_t;

typedef struct ap_subcommand_t ap_subcommand_t;

struct ap_subcommand_t {
    const char *name;
    const char *description;
    ap_arg_def_t *arguments;
    size_t arg_count;
    ap_subcommand_t **subcommands;
    size_t subcommand_count;
};

typedef struct ap_parsed_arg_t {
    const ap_arg_def_t *def;
    const char *string_value;
    union {
        int int_value;
        double float_value;
        bool bool_value;
    } numeric;
    int occurrence_count;
    bool is_set;
} ap_parsed_arg_t;

typedef struct ap_parser_t {
    const char *program_name;
    const char *description;
    ap_arg_def_t *global_args;
    size_t global_arg_count;
    ap_subcommand_t **subcommands;
    size_t subcommand_count;
} ap_parser_t;

typedef struct ap_parse_result_t {
    ap_parsed_arg_t *global_results;
    size_t global_result_count;
    const ap_subcommand_t **command_path;
    size_t command_path_length;
    ap_parsed_arg_t *current_results;
    size_t current_result_count;
    bool help_requested;
    const ap_subcommand_t *help_subcommand;
} ap_parse_result_t;

ap_parser_t* ap_parser_create(const char *program_name, const char *description);
void ap_parser_destroy(ap_parser_t *parser);

void ap_parser_add_global_arg(ap_parser_t *parser, const ap_arg_def_t *arg);
void ap_parser_add_subcommand(ap_parser_t *parser, ap_subcommand_t *subcommand);

ap_subcommand_t* ap_subcommand_create(const char *name, const char *description);
void ap_subcommand_destroy(ap_subcommand_t *subcmd);
void ap_subcommand_add_arg(ap_subcommand_t *subcmd, const ap_arg_def_t *arg);
void ap_subcommand_add_subcommand(ap_subcommand_t *parent, ap_subcommand_t *child);

ap_parse_result_t* ap_parse(const ap_parser_t *parser, int argc, char *argv[]);
void ap_parse_result_destroy(ap_parse_result_t *result);

const ap_parsed_arg_t* ap_result_get_arg(const ap_parse_result_t *result, const char *name);
bool ap_result_is_set(const ap_parse_result_t *result, const char *name);
const char* ap_result_get_string(const ap_parse_result_t *result, const char *name);
int ap_result_get_int(const ap_parse_result_t *result, const char *name);
double ap_result_get_float(const ap_parse_result_t *result, const char *name);
bool ap_result_get_bool(const ap_parse_result_t *result, const char *name);
int ap_result_get_count(const ap_parse_result_t *result, const char *name);

void ap_arg_set_range_int(ap_arg_def_t *arg, int min, int max);
void ap_arg_set_range_float(ap_arg_def_t *arg, double min, double max);
void ap_arg_set_required(ap_arg_def_t *arg, bool required);
void ap_arg_set_default(ap_arg_def_t *arg, const char *default_value);
void ap_arg_set_dup_policy(ap_arg_def_t *arg, ap_dup_policy_t policy);

#ifdef __cplusplus
}
#endif

#endif
