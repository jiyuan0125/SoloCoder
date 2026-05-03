#include "validate.h"
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include <stdio.h>
#include <errno.h>
#include <limits.h>

const char* ap_valid_status_to_string(ap_valid_status_t status) {
    switch (status) {
        case AP_VALID_OK: return "OK";
        case AP_VALID_MISSING_REQUIRED: return "Missing required argument";
        case AP_VALID_INVALID_INT: return "Invalid integer value";
        case AP_VALID_INVALID_FLOAT: return "Invalid float value";
        case AP_VALID_OUT_OF_RANGE: return "Value out of range";
        case AP_VALID_DUPLICATE_ARG: return "Duplicate argument";
        case AP_VALID_UNKNOWN_ARG: return "Unknown argument";
        default: return "Unknown error";
    }
}

bool ap_is_valid_int(const char *value) {
    if (!value || *value == '\0') {
        return false;
    }
    
    const char *p = value;
    if (*p == '+' || *p == '-') {
        p++;
    }
    
    if (*p == '\0') {
        return false;
    }
    
    while (*p != '\0') {
        if (!isdigit((unsigned char)*p)) {
            return false;
        }
        p++;
    }
    
    return true;
}

bool ap_is_valid_float(const char *value) {
    if (!value || *value == '\0') {
        return false;
    }
    
    const char *p = value;
    bool has_dot = false;
    bool has_digit = false;
    bool has_exp = false;
    
    if (*p == '+' || *p == '-') {
        p++;
    }
    
    while (*p != '\0') {
        if (*p == '.') {
            if (has_dot || has_exp) {
                return false;
            }
            has_dot = true;
        } else if (*p == 'e' || *p == 'E') {
            if (has_exp || !has_digit) {
                return false;
            }
            has_exp = true;
            has_digit = false;
            p++;
            if (*p == '+' || *p == '-') {
                p++;
            }
            continue;
        } else if (!isdigit((unsigned char)*p)) {
            return false;
        } else {
            has_digit = true;
        }
        p++;
    }
    
    return has_digit;
}

ap_valid_status_t ap_validate_int(const char *value, int *out_result, ap_valid_error_t *error) {
    if (!value) {
        if (error) {
            error->status = AP_VALID_INVALID_INT;
            error->arg_name = NULL;
            snprintf(error->message, sizeof(error->message), "NULL value provided");
        }
        return AP_VALID_INVALID_INT;
    }
    
    if (!ap_is_valid_int(value)) {
        if (error) {
            error->status = AP_VALID_INVALID_INT;
            error->arg_name = NULL;
            snprintf(error->message, sizeof(error->message), 
                     "Invalid integer value: '%s'", value);
        }
        return AP_VALID_INVALID_INT;
    }
    
    errno = 0;
    char *endptr;
    long result = strtol(value, &endptr, 10);
    
    if (errno == ERANGE || result < (long)INT_MIN || result > (long)INT_MAX) {
        if (error) {
            error->status = AP_VALID_OUT_OF_RANGE;
            error->arg_name = NULL;
            snprintf(error->message, sizeof(error->message),
                     "Integer value out of range: '%s'", value);
        }
        return AP_VALID_OUT_OF_RANGE;
    }
    
    if (out_result) {
        *out_result = (int)result;
    }
    
    return AP_VALID_OK;
}

ap_valid_status_t ap_validate_float(const char *value, double *out_result, ap_valid_error_t *error) {
    if (!value) {
        if (error) {
            error->status = AP_VALID_INVALID_FLOAT;
            error->arg_name = NULL;
            snprintf(error->message, sizeof(error->message), "NULL value provided");
        }
        return AP_VALID_INVALID_FLOAT;
    }
    
    if (!ap_is_valid_float(value)) {
        if (error) {
            error->status = AP_VALID_INVALID_FLOAT;
            error->arg_name = NULL;
            snprintf(error->message, sizeof(error->message),
                     "Invalid float value: '%s'", value);
        }
        return AP_VALID_INVALID_FLOAT;
    }
    
    errno = 0;
    char *endptr;
    double result = strtod(value, &endptr);
    
    if (errno == ERANGE) {
        if (error) {
            error->status = AP_VALID_OUT_OF_RANGE;
            error->arg_name = NULL;
            snprintf(error->message, sizeof(error->message),
                     "Float value out of range: '%s'", value);
        }
        return AP_VALID_OUT_OF_RANGE;
    }
    
    if (out_result) {
        *out_result = result;
    }
    
    return AP_VALID_OK;
}

ap_valid_status_t ap_validate_range_int(int value, const ap_range_t *range, ap_valid_error_t *error) {
    if (!range) {
        return AP_VALID_OK;
    }
    
    if (range->has_min && value < (int)range->min) {
        if (error) {
            error->status = AP_VALID_OUT_OF_RANGE;
            error->arg_name = NULL;
            snprintf(error->message, sizeof(error->message),
                     "Value %d is less than minimum %d", value, (int)range->min);
        }
        return AP_VALID_OUT_OF_RANGE;
    }
    
    if (range->has_max && value > (int)range->max) {
        if (error) {
            error->status = AP_VALID_OUT_OF_RANGE;
            error->arg_name = NULL;
            snprintf(error->message, sizeof(error->message),
                     "Value %d is greater than maximum %d", value, (int)range->max);
        }
        return AP_VALID_OUT_OF_RANGE;
    }
    
    return AP_VALID_OK;
}

ap_valid_status_t ap_validate_range_float(double value, const ap_range_t *range, ap_valid_error_t *error) {
    if (!range) {
        return AP_VALID_OK;
    }
    
    if (range->has_min && value < range->min) {
        if (error) {
            error->status = AP_VALID_OUT_OF_RANGE;
            error->arg_name = NULL;
            snprintf(error->message, sizeof(error->message),
                     "Value %g is less than minimum %g", value, range->min);
        }
        return AP_VALID_OUT_OF_RANGE;
    }
    
    if (range->has_max && value > range->max) {
        if (error) {
            error->status = AP_VALID_OUT_OF_RANGE;
            error->arg_name = NULL;
            snprintf(error->message, sizeof(error->message),
                     "Value %g is greater than maximum %g", value, range->max);
        }
        return AP_VALID_OUT_OF_RANGE;
    }
    
    return AP_VALID_OK;
}

int ap_parse_int(const char *value, int default_val) {
    if (!value) return default_val;
    
    int result;
    if (ap_validate_int(value, &result, NULL) == AP_VALID_OK) {
        return result;
    }
    return default_val;
}

double ap_parse_float(const char *value, double default_val) {
    if (!value) return default_val;
    
    double result;
    if (ap_validate_float(value, &result, NULL) == AP_VALID_OK) {
        return result;
    }
    return default_val;
}

ap_valid_status_t ap_validate_required(const ap_arg_def_t *args, size_t count,
                                        const ap_parsed_arg_t *results, size_t result_count,
                                        ap_valid_error_t *error) {
    if (!args || count == 0) {
        return AP_VALID_OK;
    }
    
    for (size_t i = 0; i < count; i++) {
        if (args[i].required) {
            bool found = false;
            for (size_t j = 0; j < result_count; j++) {
                if (results[j].def && results[j].def->name && args[i].name &&
                    strcmp(results[j].def->name, args[i].name) == 0 && results[j].is_set) {
                    found = true;
                    break;
                }
            }
            
            if (!found) {
                if (error) {
                    error->status = AP_VALID_MISSING_REQUIRED;
                    error->arg_name = args[i].name;
                    snprintf(error->message, sizeof(error->message),
                             "Required argument '%s' is missing", args[i].name);
                }
                return AP_VALID_MISSING_REQUIRED;
            }
        }
    }
    
    return AP_VALID_OK;
}

static ap_valid_status_t ap_validate_arg_values(const ap_arg_def_t *args, size_t count,
                                                 const ap_parsed_arg_t *results, size_t result_count,
                                                 ap_valid_error_t *error) {
    for (size_t i = 0; i < result_count; i++) {
        if (!results[i].is_set) continue;
        
        const ap_arg_def_t *def = results[i].def;
        if (!def) continue;
        
        if (def->dup_policy == AP_DUP_ERROR && results[i].occurrence_count > 1) {
            if (error) {
                error->status = AP_VALID_DUPLICATE_ARG;
                error->arg_name = def->name;
                snprintf(error->message, sizeof(error->message),
                         "Argument '%s' specified multiple times (duplicates not allowed)",
                         def->name ? def->name : "");
            }
            return AP_VALID_DUPLICATE_ARG;
        }
        
        if (def->type == AP_TYPE_INT && results[i].string_value) {
            int int_val;
            ap_valid_status_t status = ap_validate_int(results[i].string_value, &int_val, error);
            if (status != AP_VALID_OK) {
                if (error) {
                    error->arg_name = def->name;
                }
                return status;
            }
            
            if (def->range.has_min || def->range.has_max) {
                status = ap_validate_range_int(int_val, &def->range, error);
                if (status != AP_VALID_OK) {
                    if (error) {
                        error->arg_name = def->name;
                    }
                    return status;
                }
            }
        }
        
        if (def->type == AP_TYPE_FLOAT && results[i].string_value) {
            double float_val;
            ap_valid_status_t status = ap_validate_float(results[i].string_value, &float_val, error);
            if (status != AP_VALID_OK) {
                if (error) {
                    error->arg_name = def->name;
                }
                return status;
            }
            
            if (def->range.has_min || def->range.has_max) {
                status = ap_validate_range_float(float_val, &def->range, error);
                if (status != AP_VALID_OK) {
                    if (error) {
                        error->arg_name = def->name;
                    }
                    return status;
                }
            }
        }
    }
    
    return AP_VALID_OK;
}

ap_valid_status_t ap_validate_parse_result(const ap_parser_t *parser, 
                                            const ap_parse_result_t *result,
                                            ap_valid_error_t *error) {
    if (!parser || !result) {
        return AP_VALID_OK;
    }
    
    ap_valid_status_t status;
    
    status = ap_validate_required(parser->global_args, parser->global_arg_count,
                                   result->global_results, result->global_result_count,
                                   error);
    if (status != AP_VALID_OK) {
        return status;
    }
    
    status = ap_validate_arg_values(parser->global_args, parser->global_arg_count,
                                     result->global_results, result->global_result_count,
                                     error);
    if (status != AP_VALID_OK) {
        return status;
    }
    
    if (result->command_path_length > 0) {
        const ap_subcommand_t *current = NULL;
        for (size_t i = 0; i < result->command_path_length; i++) {
            current = result->command_path[i];
            
            if (current->arguments && current->arg_count > 0) {
                status = ap_validate_required(current->arguments, current->arg_count,
                                               result->current_results, result->current_result_count,
                                               error);
                if (status != AP_VALID_OK) {
                    return status;
                }
                
                status = ap_validate_arg_values(current->arguments, current->arg_count,
                                                 result->current_results, result->current_result_count,
                                                 error);
                if (status != AP_VALID_OK) {
                    return status;
                }
            }
        }
    }
    
    return AP_VALID_OK;
}
