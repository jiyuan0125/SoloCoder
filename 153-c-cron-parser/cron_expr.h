#ifndef CRON_EXPR_H
#define CRON_EXPR_H

#include <time.h>

#define CRON_MIN_COUNT     60
#define CRON_HOUR_COUNT    24
#define CRON_DOM_COUNT     32
#define CRON_MONTH_COUNT   13
#define CRON_DOW_COUNT     8

typedef struct {
    unsigned char minute[CRON_MIN_COUNT];
    unsigned char hour[CRON_HOUR_COUNT];
    unsigned char day_of_month[CRON_DOM_COUNT];
    unsigned char month[CRON_MONTH_COUNT];
    unsigned char day_of_week[CRON_DOW_COUNT];
} cron_expr_t;

typedef enum {
    CRON_PARSE_OK = 0,
    CRON_PARSE_INVALID_FORMAT,
    CRON_PARSE_INVALID_FIELD,
    CRON_PARSE_INVALID_VALUE,
    CRON_PARSE_MEMORY_ERROR
} cron_parse_error_t;

const char* cron_parse_strerror(cron_parse_error_t err);

cron_parse_error_t cron_expr_parse(const char* expr, cron_expr_t* result);

int cron_expr_matches(const cron_expr_t* expr, const struct tm* tm);

int cron_calculate_next_run(const cron_expr_t* expr, const time_t* from_time, time_t* next_time);

int cron_is_leap_year(int year);

int cron_days_in_month(int year, int month);

void cron_expr_clear(cron_expr_t* expr);

void cron_expr_set_all(cron_expr_t* expr);

#endif
