#include "csv_types.h"
#include <stdlib.h>
#include <string.h>
#include <limits.h>
#include <errno.h>
#include <ctype.h>

static const char* g_status_messages[] = {
    "OK",
    "NULL pointer",
    "Empty string",
    "Overflow",
    "Underflow",
    "Invalid format"
};

static int CSV_IsWhitespaceOnly(const char *str) {
    if (!str) return 1;
    while (*str) {
        if (!isspace((unsigned char)*str)) return 0;
        str++;
    }
    return 1;
}

CSV_TypeResult CSV_ConvertToInt(const char *str, int *out) {
    CSV_TypeResult result;
    result.status = CSV_TYPE_OK;
    result.error_field = 0;
    result.error_message = NULL;

    if (out) *out = 0;

    if (!str) {
        result.status = CSV_TYPE_NULL;
        result.error_message = g_status_messages[CSV_TYPE_NULL];
        return result;
    }

    if (CSV_IsWhitespaceOnly(str)) {
        result.status = CSV_TYPE_EMPTY;
        result.error_message = g_status_messages[CSV_TYPE_EMPTY];
        return result;
    }

    errno = 0;
    char *endptr = NULL;
    long val = strtol(str, &endptr, 10);

    if (endptr == str || *endptr != '\0') {
        result.status = CSV_TYPE_INVALID_FORMAT;
        result.error_message = g_status_messages[CSV_TYPE_INVALID_FORMAT];
        return result;
    }

    if (errno == ERANGE) {
        if (val == LONG_MAX) {
            result.status = CSV_TYPE_OVERFLOW;
            result.error_message = g_status_messages[CSV_TYPE_OVERFLOW];
        } else {
            result.status = CSV_TYPE_UNDERFLOW;
            result.error_message = g_status_messages[CSV_TYPE_UNDERFLOW];
        }
        return result;
    }

    if (val > INT_MAX) {
        result.status = CSV_TYPE_OVERFLOW;
        result.error_message = g_status_messages[CSV_TYPE_OVERFLOW];
        return result;
    }

    if (val < INT_MIN) {
        result.status = CSV_TYPE_UNDERFLOW;
        result.error_message = g_status_messages[CSV_TYPE_UNDERFLOW];
        return result;
    }

    if (out) *out = (int)val;
    result.status = CSV_TYPE_OK;
    result.error_message = g_status_messages[CSV_TYPE_OK];
    return result;
}

CSV_TypeResult CSV_ConvertToLongLong(const char *str, long long *out) {
    CSV_TypeResult result;
    result.status = CSV_TYPE_OK;
    result.error_field = 0;
    result.error_message = NULL;

    if (out) *out = 0;

    if (!str) {
        result.status = CSV_TYPE_NULL;
        result.error_message = g_status_messages[CSV_TYPE_NULL];
        return result;
    }

    if (CSV_IsWhitespaceOnly(str)) {
        result.status = CSV_TYPE_EMPTY;
        result.error_message = g_status_messages[CSV_TYPE_EMPTY];
        return result;
    }

    errno = 0;
    char *endptr = NULL;
    long long val = strtoll(str, &endptr, 10);

    if (endptr == str || *endptr != '\0') {
        result.status = CSV_TYPE_INVALID_FORMAT;
        result.error_message = g_status_messages[CSV_TYPE_INVALID_FORMAT];
        return result;
    }

    if (errno == ERANGE) {
        if (val == LLONG_MAX) {
            result.status = CSV_TYPE_OVERFLOW;
            result.error_message = g_status_messages[CSV_TYPE_OVERFLOW];
        } else {
            result.status = CSV_TYPE_UNDERFLOW;
            result.error_message = g_status_messages[CSV_TYPE_UNDERFLOW];
        }
        return result;
    }

    if (out) *out = val;
    result.status = CSV_TYPE_OK;
    result.error_message = g_status_messages[CSV_TYPE_OK];
    return result;
}

CSV_TypeResult CSV_ConvertToDouble(const char *str, double *out) {
    CSV_TypeResult result;
    result.status = CSV_TYPE_OK;
    result.error_field = 0;
    result.error_message = NULL;

    if (out) *out = 0.0;

    if (!str) {
        result.status = CSV_TYPE_NULL;
        result.error_message = g_status_messages[CSV_TYPE_NULL];
        return result;
    }

    if (CSV_IsWhitespaceOnly(str)) {
        result.status = CSV_TYPE_EMPTY;
        result.error_message = g_status_messages[CSV_TYPE_EMPTY];
        return result;
    }

    errno = 0;
    char *endptr = NULL;
    double val = strtod(str, &endptr);

    if (endptr == str || *endptr != '\0') {
        result.status = CSV_TYPE_INVALID_FORMAT;
        result.error_message = g_status_messages[CSV_TYPE_INVALID_FORMAT];
        return result;
    }

    if (errno == ERANGE) {
        result.status = CSV_TYPE_OVERFLOW;
        result.error_message = g_status_messages[CSV_TYPE_OVERFLOW];
        return result;
    }

    if (out) *out = val;
    result.status = CSV_TYPE_OK;
    result.error_message = g_status_messages[CSV_TYPE_OK];
    return result;
}

int CSV_TypeResultIsSuccess(CSV_TypeResult result) {
    return result.status == CSV_TYPE_OK;
}

const char* CSV_TypeStatusToString(CSV_TypeStatus status) {
    if (status < 0 || status >= (int)(sizeof(g_status_messages) / sizeof(g_status_messages[0]))) {
        return "Unknown";
    }
    return g_status_messages[status];
}
