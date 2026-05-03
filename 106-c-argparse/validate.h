#ifndef AP_VALIDATE_H
#define AP_VALIDATE_H

#include "argparse.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    AP_VALID_OK = 0,
    AP_VALID_MISSING_REQUIRED,
    AP_VALID_INVALID_INT,
    AP_VALID_INVALID_FLOAT,
    AP_VALID_OUT_OF_RANGE,
    AP_VALID_DUPLICATE_ARG,
    AP_VALID_UNKNOWN_ARG
} ap_valid_status_t;

typedef struct ap_valid_error_t {
    ap_valid_status_t status;
    const char *arg_name;
    char message[256];
} ap_valid_error_t;

ap_valid_status_t ap_validate_parse_result(const ap_parser_t *parser, 
                                            const ap_parse_result_t *result,
                                            ap_valid_error_t *error);

ap_valid_status_t ap_validate_required(const ap_arg_def_t *args, size_t count,
                                        const ap_parsed_arg_t *results, size_t result_count,
                                        ap_valid_error_t *error);

ap_valid_status_t ap_validate_int(const char *value, int *out_result, ap_valid_error_t *error);
ap_valid_status_t ap_validate_float(const char *value, double *out_result, ap_valid_error_t *error);

ap_valid_status_t ap_validate_range_int(int value, const ap_range_t *range, ap_valid_error_t *error);
ap_valid_status_t ap_validate_range_float(double value, const ap_range_t *range, ap_valid_error_t *error);

bool ap_is_valid_int(const char *value);
bool ap_is_valid_float(const char *value);

int ap_parse_int(const char *value, int default_val);
double ap_parse_float(const char *value, double default_val);

const char* ap_valid_status_to_string(ap_valid_status_t status);

#ifdef __cplusplus
}
#endif

#endif
