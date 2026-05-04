#define _GNU_SOURCE
#include "cron_expr.h"
#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <ctype.h>

#define MAX_EXPR_LEN 256

static const char* cron_error_messages[] = {
    "OK",
    "Invalid cron expression format",
    "Invalid field in expression",
    "Invalid value in field",
    "Memory error"
};

const char* cron_parse_strerror(cron_parse_error_t err) {
    if (err < 0 || err >= (int)(sizeof(cron_error_messages)/sizeof(cron_error_messages[0]))) {
        return "Unknown error";
    }
    return cron_error_messages[err];
}

void cron_expr_clear(cron_expr_t* expr) {
    if (!expr) return;
    memset(expr->minute, 0, sizeof(expr->minute));
    memset(expr->hour, 0, sizeof(expr->hour));
    memset(expr->day_of_month, 0, sizeof(expr->day_of_month));
    memset(expr->month, 0, sizeof(expr->month));
    memset(expr->day_of_week, 0, sizeof(expr->day_of_week));
}

void cron_expr_set_all(cron_expr_t* expr) {
    if (!expr) return;
    int i;
    for (i = 0; i < CRON_MIN_COUNT; i++) expr->minute[i] = 1;
    for (i = 0; i < CRON_HOUR_COUNT; i++) expr->hour[i] = 1;
    for (i = 1; i < CRON_DOM_COUNT; i++) expr->day_of_month[i] = 1;
    for (i = 1; i < CRON_MONTH_COUNT; i++) expr->month[i] = 1;
    for (i = 0; i < CRON_DOW_COUNT; i++) expr->day_of_week[i] = 1;
}

int cron_is_leap_year(int year) {
    if (year % 4 != 0) return 0;
    if (year % 100 != 0) return 1;
    if (year % 400 != 0) return 0;
    return 1;
}

int cron_days_in_month(int year, int month) {
    static const int days[] = {0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31};
    if (month < 1 || month > 12) return 0;
    if (month == 2 && cron_is_leap_year(year)) return 29;
    return days[month];
}

static int parse_number(const char** str) {
    int val = 0;
    if (!isdigit(**str)) return -1;
    while (isdigit(**str)) {
        val = val * 10 + (**str - '0');
        (*str)++;
    }
    return val;
}

static cron_parse_error_t parse_field(const char* field, unsigned char* bits, 
                                        int min_val, int max_val, int offset) {
    const char* p = field;
    if (*p == '\0') return CRON_PARSE_INVALID_FIELD;

    memset(bits, 0, max_val - min_val + 1);

    while (*p != '\0') {
        int start, end, step = 1;
        
        if (*p == '*') {
            p++;
            start = min_val;
            end = max_val;
            
            if (*p == '/') {
                p++;
                step = parse_number(&p);
                if (step <= 0) return CRON_PARSE_INVALID_VALUE;
            }
        } else {
            start = parse_number(&p);
            if (start < 0) return CRON_PARSE_INVALID_VALUE;
            
            if (*p == '-') {
                p++;
                end = parse_number(&p);
                if (end < 0) return CRON_PARSE_INVALID_VALUE;
            } else {
                end = start;
            }
            
            if (*p == '/') {
                p++;
                step = parse_number(&p);
                if (step <= 0) return CRON_PARSE_INVALID_VALUE;
            }
        }

        if (start < min_val || end > max_val || start > end) {
            return CRON_PARSE_INVALID_VALUE;
        }

        int i;
        for (i = start; i <= end; i += step) {
            int idx = i - min_val + offset;
            if (idx >= 0 && idx < (max_val - min_val + 1)) {
                bits[idx] = 1;
            }
        }

        if (*p == ',') {
            p++;
        } else if (*p != '\0') {
            return CRON_PARSE_INVALID_FORMAT;
        }
    }

    return CRON_PARSE_OK;
}

static char* strtrim(char* str) {
    while (isspace((unsigned char)*str)) str++;
    if (*str == '\0') return str;
    char* end = str + strlen(str) - 1;
    while (end > str && isspace((unsigned char)*end)) end--;
    *(end + 1) = '\0';
    return str;
}

cron_parse_error_t cron_expr_parse(const char* expr, cron_expr_t* result) {
    if (!expr || !result) return CRON_PARSE_INVALID_FORMAT;

    char expr_copy[MAX_EXPR_LEN];
    strncpy(expr_copy, expr, MAX_EXPR_LEN - 1);
    expr_copy[MAX_EXPR_LEN - 1] = '\0';

    char* fields[5];
    int field_count = 0;
    char* p = strtok(expr_copy, " \t");
    
    while (p != NULL && field_count < 5) {
        fields[field_count++] = strtrim(p);
        p = strtok(NULL, " \t");
    }

    if (field_count != 5) return CRON_PARSE_INVALID_FORMAT;

    cron_expr_clear(result);

    cron_parse_error_t err;
    
    err = parse_field(fields[0], result->minute, 0, 59, 0);
    if (err != CRON_PARSE_OK) return err;
    
    err = parse_field(fields[1], result->hour, 0, 23, 0);
    if (err != CRON_PARSE_OK) return err;
    
    err = parse_field(fields[2], result->day_of_month, 1, 31, 1);
    if (err != CRON_PARSE_OK) return err;
    
    err = parse_field(fields[3], result->month, 1, 12, 1);
    if (err != CRON_PARSE_OK) return err;
    
    err = parse_field(fields[4], result->day_of_week, 0, 7, 0);
    if (err != CRON_PARSE_OK) return err;

    if (result->day_of_week[7]) {
        result->day_of_week[0] = 1;
    }

    return CRON_PARSE_OK;
}

int cron_expr_matches(const cron_expr_t* expr, const struct tm* tm) {
    if (!expr || !tm) return 0;

    int minute = tm->tm_min;
    int hour = tm->tm_hour;
    int day_of_month = tm->tm_mday;
    int month = tm->tm_mon + 1;
    int day_of_week = tm->tm_wday;

    if (minute < 0 || minute >= CRON_MIN_COUNT) return 0;
    if (hour < 0 || hour >= CRON_HOUR_COUNT) return 0;
    if (day_of_month < 1 || day_of_month >= CRON_DOM_COUNT) return 0;
    if (month < 1 || month >= CRON_MONTH_COUNT) return 0;
    if (day_of_week < 0 || day_of_week >= 7) return 0;

    if (!expr->minute[minute]) return 0;
    if (!expr->hour[hour]) return 0;
    if (!expr->month[month]) return 0;

    int dom_matches = expr->day_of_month[day_of_month];
    int dow_matches = expr->day_of_week[day_of_week];

    int dom_is_wildcard = (expr->day_of_month[1] && expr->day_of_month[2] && 
                            expr->day_of_month[3] && expr->day_of_month[4] &&
                            expr->day_of_month[5] && expr->day_of_month[6] &&
                            expr->day_of_month[7] && expr->day_of_month[8] &&
                            expr->day_of_month[9] && expr->day_of_month[10] &&
                            expr->day_of_month[11] && expr->day_of_month[12] &&
                            expr->day_of_month[13] && expr->day_of_month[14] &&
                            expr->day_of_month[15] && expr->day_of_month[16] &&
                            expr->day_of_month[17] && expr->day_of_month[18] &&
                            expr->day_of_month[19] && expr->day_of_month[20] &&
                            expr->day_of_month[21] && expr->day_of_month[22] &&
                            expr->day_of_month[23] && expr->day_of_month[24] &&
                            expr->day_of_month[25] && expr->day_of_month[26] &&
                            expr->day_of_month[27] && expr->day_of_month[28] &&
                            expr->day_of_month[29] && expr->day_of_month[30] &&
                            expr->day_of_month[31]);

    int dow_is_wildcard = (expr->day_of_week[0] && expr->day_of_week[1] && 
                            expr->day_of_week[2] && expr->day_of_week[3] &&
                            expr->day_of_week[4] && expr->day_of_week[5] &&
                            expr->day_of_week[6]);

    if (dom_is_wildcard && dow_is_wildcard) {
        return 1;
    } else if (dom_is_wildcard) {
        return dow_matches;
    } else if (dow_is_wildcard) {
        return dom_matches;
    } else {
        return (dom_matches || dow_matches);
    }
}

static int is_dom_all_set(const unsigned char* dom) {
    int i;
    for (i = 1; i < 32; i++) {
        if (!dom[i]) return 0;
    }
    return 1;
}

static int is_dow_all_set(const unsigned char* dow) {
    int i;
    for (i = 0; i < 7; i++) {
        if (!dow[i]) return 0;
    }
    return 1;
}

int cron_calculate_next_run(const cron_expr_t* expr, const time_t* from_time, time_t* next_time) {
    if (!expr || !from_time || !next_time) return -1;

    time_t t = *from_time;
    t -= (t % 60);

    int dom_all = is_dom_all_set(expr->day_of_month);
    int dow_all = is_dow_all_set(expr->day_of_week);

    int iterations = 0;
    const int max_iterations = 535680;

    while (iterations < max_iterations) {
        struct tm* tm_ptr = localtime(&t);
        if (!tm_ptr) return -1;

        int minute = tm_ptr->tm_min;
        int hour = tm_ptr->tm_hour;
        int day_of_month = tm_ptr->tm_mday;
        int month = tm_ptr->tm_mon + 1;
        int day_of_week = tm_ptr->tm_wday;
        int year = tm_ptr->tm_year + 1900;

        int month_ok = expr->month[month];
        if (!month_ok) {
            t = t + (cron_days_in_month(year, month) - day_of_month + 1) * 86400;
            t -= (t % 86400);
            iterations++;
            continue;
        }

        int dom_ok = expr->day_of_month[day_of_month];
        int dow_ok = expr->day_of_week[day_of_week];
        
        int day_ok = 0;
        if (dom_all && dow_all) {
            day_ok = 1;
        } else if (dom_all) {
            day_ok = dow_ok;
        } else if (dow_all) {
            day_ok = dom_ok;
        } else {
            day_ok = (dom_ok || dow_ok);
        }

        if (!day_ok) {
            t += 86400;
            t -= (t % 86400);
            iterations++;
            continue;
        }

        int hour_ok = expr->hour[hour];
        if (!hour_ok) {
            t += 3600;
            t -= (t % 3600);
            iterations++;
            continue;
        }

        int min_ok = expr->minute[minute];
        if (!min_ok) {
            t += 60;
            iterations++;
            continue;
        }

        *next_time = t;
        return 0;
    }

    return -1;
}
