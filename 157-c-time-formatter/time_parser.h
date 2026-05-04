#ifndef TIME_PARSER_H
#define TIME_PARSER_H

#include <time.h>
#include "timezone.h"

#define PARSE_ERROR_MAX_LEN 256

typedef enum {
    PARSE_OK = 0,
    PARSE_ERR_EMPTY_INPUT,
    PARSE_ERR_INVALID_FORMAT,
    PARSE_ERR_INVALID_YEAR,
    PARSE_ERR_INVALID_MONTH,
    PARSE_ERR_INVALID_DAY,
    PARSE_ERR_INVALID_HOUR,
    PARSE_ERR_INVALID_MINUTE,
    PARSE_ERR_INVALID_SECOND,
    PARSE_ERR_DATE_OUT_OF_RANGE,
    PARSE_ERR_UNKNOWN
} ParseErrorCode;

typedef struct {
    ParseErrorCode code;
    char message[PARSE_ERROR_MAX_LEN];
    int position;
} ParseError;

int parse_time_string(const char *input, const TimeZoneInfo *tz,
                      time_t *result, ParseError *error);

int parse_iso_format(const char *input, struct tm *tm, ParseError *error);

int parse_slash_format(const char *input, struct tm *tm, ParseError *error);

int parse_month_name_format(const char *input, struct tm *tm, ParseError *error);

int parse_time_component(const char *input, int *hour, int *minute, int *second,
                         ParseError *error);

int validate_datetime(struct tm *tm, ParseError *error);

const char *parse_error_to_string(ParseErrorCode code);

#endif
