#ifndef CSV_TYPES_H
#define CSV_TYPES_H

#include <stddef.h>
#include <stdint.h>

typedef enum {
    CSV_TYPE_OK,
    CSV_TYPE_NULL,
    CSV_TYPE_EMPTY,
    CSV_TYPE_OVERFLOW,
    CSV_TYPE_UNDERFLOW,
    CSV_TYPE_INVALID_FORMAT
} CSV_TypeStatus;

typedef struct CSV_TypeResult {
    CSV_TypeStatus status;
    int error_field;
    const char *error_message;
} CSV_TypeResult;

CSV_TypeResult CSV_ConvertToInt(const char *str, int *out);
CSV_TypeResult CSV_ConvertToLongLong(const char *str, long long *out);
CSV_TypeResult CSV_ConvertToDouble(const char *str, double *out);

int CSV_TypeResultIsSuccess(CSV_TypeResult result);
const char* CSV_TypeStatusToString(CSV_TypeStatus status);

#endif
