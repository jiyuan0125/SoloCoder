#include "time_parser.h"
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <ctype.h>

static const char *MONTH_NAMES[] = {
    "january", "february", "march", "april", "may", "june",
    "july", "august", "september", "october", "november", "december",
    "jan", "feb", "mar", "apr", "may", "jun",
    "jul", "aug", "sep", "oct", "nov", "dec",
    NULL
};

static const int MONTH_NUMBERS[] = {
    1, 2, 3, 4, 5, 6,
    7, 8, 9, 10, 11, 12,
    1, 2, 3, 4, 5, 6,
    7, 8, 9, 10, 11, 12
};

static void set_error(ParseError *error, ParseErrorCode code, const char *message, int position) {
    if (!error) return;
    error->code = code;
    error->position = position;
    snprintf(error->message, PARSE_ERROR_MAX_LEN, "%s", message ? message : parse_error_to_string(code));
}

const char *parse_error_to_string(ParseErrorCode code) {
    switch (code) {
        case PARSE_OK: return "解析成功";
        case PARSE_ERR_EMPTY_INPUT: return "输入为空";
        case PARSE_ERR_INVALID_FORMAT: return "无法识别的时间格式";
        case PARSE_ERR_INVALID_YEAR: return "年份格式无效";
        case PARSE_ERR_INVALID_MONTH: return "月份格式无效";
        case PARSE_ERR_INVALID_DAY: return "日期格式无效";
        case PARSE_ERR_INVALID_HOUR: return "小时格式无效";
        case PARSE_ERR_INVALID_MINUTE: return "分钟格式无效";
        case PARSE_ERR_INVALID_SECOND: return "秒格式无效";
        case PARSE_ERR_DATE_OUT_OF_RANGE: return "日期超出有效范围";
        case PARSE_ERR_UNKNOWN: return "未知错误";
        default: return "未知错误";
    }
}

static int is_leap_year(int year) {
    return (year % 4 == 0 && (year % 100 != 0 || year % 400 == 0));
}

static int days_in_month(int year, int month) {
    int days[] = {31,28,31,30,31,30,31,31,30,31,30,31};
    if (month < 1 || month > 12) return 0;
    if (month == 2 && is_leap_year(year)) return 29;
    return days[month - 1];
}

int validate_datetime(struct tm *tm, ParseError *error) {
    if (tm->tm_year < 0) {
        set_error(error, PARSE_ERR_INVALID_YEAR, "年份不能为负数", 0);
        return -1;
    }
    if (tm->tm_year + 1900 < 1970) {
        set_error(error, PARSE_ERR_DATE_OUT_OF_RANGE, "年份不能早于1970年", 0);
        return -1;
    }
    if (tm->tm_mon < 1 || tm->tm_mon > 12) {
        char msg[PARSE_ERROR_MAX_LEN];
        snprintf(msg, PARSE_ERROR_MAX_LEN, "月份 %d 无效，有效范围是 1-12", tm->tm_mon);
        set_error(error, PARSE_ERR_INVALID_MONTH, msg, 0);
        return -1;
    }
    
    int max_day = days_in_month(tm->tm_year + 1900, tm->tm_mon);
    if (tm->tm_mday < 1 || tm->tm_mday > max_day) {
        char msg[PARSE_ERROR_MAX_LEN];
        snprintf(msg, PARSE_ERROR_MAX_LEN, "日期 %d 无效，%d年%d月有效范围是 1-%d",
                 tm->tm_mday, tm->tm_year + 1900, tm->tm_mon, max_day);
        set_error(error, PARSE_ERR_INVALID_DAY, msg, 0);
        return -1;
    }
    
    if (tm->tm_hour < 0 || tm->tm_hour > 23) {
        char msg[PARSE_ERROR_MAX_LEN];
        snprintf(msg, PARSE_ERROR_MAX_LEN, "小时 %d 无效，有效范围是 0-23", tm->tm_hour);
        set_error(error, PARSE_ERR_INVALID_HOUR, msg, 0);
        return -1;
    }
    if (tm->tm_min < 0 || tm->tm_min > 59) {
        char msg[PARSE_ERROR_MAX_LEN];
        snprintf(msg, PARSE_ERROR_MAX_LEN, "分钟 %d 无效，有效范围是 0-59", tm->tm_min);
        set_error(error, PARSE_ERR_INVALID_MINUTE, msg, 0);
        return -1;
    }
    if (tm->tm_sec < 0 || tm->tm_sec > 59) {
        char msg[PARSE_ERROR_MAX_LEN];
        snprintf(msg, PARSE_ERROR_MAX_LEN, "秒 %d 无效，有效范围是 0-59", tm->tm_sec);
        set_error(error, PARSE_ERR_INVALID_SECOND, msg, 0);
        return -1;
    }
    
    return 0;
}

static const char *skip_whitespace(const char *p) {
    while (*p && isspace(*p)) p++;
    return p;
}

static int parse_int(const char **p, int *result) {
    const char *start = *p;
    if (!isdigit(*start)) return 0;
    
    int value = 0;
    while (isdigit(**p)) {
        value = value * 10 + (**p - '0');
        (*p)++;
    }
    *result = value;
    return 1;
}

static int parse_month_name(const char **p, int *month) {
    const char *start = *p;
    char lower[32];
    int i = 0;
    
    while (*start && isalpha(*start) && i < 31) {
        lower[i] = tolower(*start);
        i++;
        start++;
    }
    if (i == 0) return 0;
    lower[i] = '\0';
    
    for (int j = 0; MONTH_NAMES[j] != NULL; j++) {
        if (strcmp(lower, MONTH_NAMES[j]) == 0) {
            *month = MONTH_NUMBERS[j];
            *p = start;
            return 1;
        }
    }
    return 0;
}

int parse_time_component(const char *input, int *hour, int *minute, int *second,
                         ParseError *error) {
    const char *p = input;
    
    if (!parse_int(&p, hour)) {
        set_error(error, PARSE_ERR_INVALID_HOUR, "无法解析小时", (int)(p - input));
        return -1;
    }
    
    if (*p == ':') p++;
    else {
        set_error(error, PARSE_ERR_INVALID_FORMAT, "时间部分格式错误，缺少冒号分隔符", (int)(p - input));
        return -1;
    }
    
    if (!parse_int(&p, minute)) {
        set_error(error, PARSE_ERR_INVALID_MINUTE, "无法解析分钟", (int)(p - input));
        return -1;
    }
    
    *second = 0;
    if (*p == ':') {
        p++;
        if (!parse_int(&p, second)) {
            set_error(error, PARSE_ERR_INVALID_SECOND, "无法解析秒", (int)(p - input));
            return -1;
        }
    }
    
    return 0;
}

int parse_iso_format(const char *input, struct tm *tm, ParseError *error) {
    const char *p = input;
    memset(tm, 0, sizeof(struct tm));
    tm->tm_isdst = -1;
    
    int year, month, day;
    int hour = 0, minute = 0, second = 0;
    
    p = skip_whitespace(p);
    int start_pos = (int)(p - input);
    
    if (!parse_int(&p, &year)) {
        set_error(error, PARSE_ERR_INVALID_YEAR, "无法解析年份", start_pos);
        return -1;
    }
    
    if (*p != '-') {
        set_error(error, PARSE_ERR_INVALID_FORMAT, "日期格式错误，应为 '-' 分隔符", (int)(p - input));
        return -1;
    }
    p++;
    
    if (!parse_int(&p, &month)) {
        set_error(error, PARSE_ERR_INVALID_MONTH, "无法解析月份", (int)(p - input));
        return -1;
    }
    
    if (*p != '-') {
        set_error(error, PARSE_ERR_INVALID_FORMAT, "日期格式错误，应为 '-' 分隔符", (int)(p - input));
        return -1;
    }
    p++;
    
    if (!parse_int(&p, &day)) {
        set_error(error, PARSE_ERR_INVALID_DAY, "无法解析日期", (int)(p - input));
        return -1;
    }
    
    p = skip_whitespace(p);
    if (*p != '\0') {
        if (parse_time_component(p, &hour, &minute, &second, error) != 0) {
            return -1;
        }
    }
    
    tm->tm_year = year - 1900;
    tm->tm_mon = month;
    tm->tm_mday = day;
    tm->tm_hour = hour;
    tm->tm_min = minute;
    tm->tm_sec = second;
    
    return validate_datetime(tm, error);
}

int parse_slash_format(const char *input, struct tm *tm, ParseError *error) {
    const char *p = input;
    memset(tm, 0, sizeof(struct tm));
    tm->tm_isdst = -1;
    
    int year, month, day;
    int hour = 0, minute = 0, second = 0;
    
    p = skip_whitespace(p);
    
    if (!parse_int(&p, &year)) {
        set_error(error, PARSE_ERR_INVALID_YEAR, "无法解析年份", (int)(p - input));
        return -1;
    }
    
    if (*p != '/') {
        set_error(error, PARSE_ERR_INVALID_FORMAT, "日期格式错误，应为 '/' 分隔符", (int)(p - input));
        return -1;
    }
    p++;
    
    if (!parse_int(&p, &month)) {
        set_error(error, PARSE_ERR_INVALID_MONTH, "无法解析月份", (int)(p - input));
        return -1;
    }
    
    if (*p != '/') {
        set_error(error, PARSE_ERR_INVALID_FORMAT, "日期格式错误，应为 '/' 分隔符", (int)(p - input));
        return -1;
    }
    p++;
    
    if (!parse_int(&p, &day)) {
        set_error(error, PARSE_ERR_INVALID_DAY, "无法解析日期", (int)(p - input));
        return -1;
    }
    
    p = skip_whitespace(p);
    if (*p != '\0') {
        if (parse_time_component(p, &hour, &minute, &second, error) != 0) {
            return -1;
        }
    }
    
    tm->tm_year = year - 1900;
    tm->tm_mon = month;
    tm->tm_mday = day;
    tm->tm_hour = hour;
    tm->tm_min = minute;
    tm->tm_sec = second;
    
    return validate_datetime(tm, error);
}

int parse_month_name_format(const char *input, struct tm *tm, ParseError *error) {
    const char *p = input;
    memset(tm, 0, sizeof(struct tm));
    tm->tm_isdst = -1;
    
    int year = 0, month = 0, day = 0;
    int hour = 0, minute = 0, second = 0;
    
    p = skip_whitespace(p);
    
    if (!parse_month_name(&p, &month)) {
        set_error(error, PARSE_ERR_INVALID_MONTH, "无法识别的月份名称", (int)(p - input));
        return -1;
    }
    
    p = skip_whitespace(p);
    
    if (!parse_int(&p, &day)) {
        set_error(error, PARSE_ERR_INVALID_DAY, "无法解析日期", (int)(p - input));
        return -1;
    }
    
    if (*p == ',') p++;
    p = skip_whitespace(p);
    
    if (!parse_int(&p, &year)) {
        time_t now = time(NULL);
        struct tm *now_tm = localtime(&now);
        year = now_tm->tm_year + 1900;
    }
    
    p = skip_whitespace(p);
    if (*p != '\0') {
        if (parse_time_component(p, &hour, &minute, &second, error) != 0) {
            return -1;
        }
    }
    
    tm->tm_year = year - 1900;
    tm->tm_mon = month;
    tm->tm_mday = day;
    tm->tm_hour = hour;
    tm->tm_min = minute;
    tm->tm_sec = second;
    
    return validate_datetime(tm, error);
}

static time_t tm_to_time_t(struct tm *tm, const TimeZoneInfo *tz) {
    struct tm tmp = *tm;
    tmp.tm_mon--;
    
    time_t local_time;
    if (tz->has_dst) {
        tmp.tm_isdst = -1;
        local_time = mktime(&tmp);
    } else {
        tmp.tm_isdst = 0;
        local_time = mktime(&tmp);
        
        if (local_time != (time_t)-1) {
            local_time = timezone_convert_local_to_utc(local_time, tz);
        }
    }
    
    return local_time;
}

int parse_time_string(const char *input, const TimeZoneInfo *tz,
                      time_t *result, ParseError *error) {
    if (!input || !input[0]) {
        set_error(error, PARSE_ERR_EMPTY_INPUT, NULL, 0);
        return -1;
    }
    
    if (!tz) tz = &TIMEZONE_UTC;
    
    struct tm tm;
    int found = 0;
    
    if (strchr(input, '-')) {
        if (parse_iso_format(input, &tm, error) == 0) {
            found = 1;
        }
    }
    
    if (!found && strchr(input, '/')) {
        ParseError slash_error;
        if (parse_slash_format(input, &tm, &slash_error) == 0) {
            found = 1;
        } else if (error) {
            *error = slash_error;
        }
    }
    
    if (!found) {
        const char *p = skip_whitespace(input);
        if (isalpha(*p)) {
            ParseError name_error;
            if (parse_month_name_format(input, &tm, &name_error) == 0) {
                found = 1;
            } else if (error) {
                *error = name_error;
            }
        }
    }
    
    if (!found) {
        set_error(error, PARSE_ERR_INVALID_FORMAT, 
                  "无法识别的时间格式。支持的格式：\n"
                  "  - ISO格式: 2026-05-03 14:30:00\n"
                  "  - 斜杠格式: 2026/05/03 14:30\n"
                  "  - 英文格式: May 3 2026 或 May 3, 2026",
                  0);
        return -1;
    }
    
    time_t t = tm_to_time_t(&tm, tz);
    if (t == (time_t)-1) {
        set_error(error, PARSE_ERR_DATE_OUT_OF_RANGE, "日期超出有效范围", 0);
        return -1;
    }
    
    if (result) *result = t;
    return 0;
}
