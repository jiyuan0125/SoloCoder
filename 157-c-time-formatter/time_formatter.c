#include "time_formatter.h"
#include <stdio.h>
#include <string.h>

static void normalize_tm_day(struct tm *tm) {
    int days_in_month[] = {31,28,31,30,31,30,31,31,30,31,30,31};
    int is_leap = (tm->tm_year % 4 == 0 && (tm->tm_year % 100 != 0 || tm->tm_year % 400 == 0));
    
    while (tm->tm_mday > days_in_month[tm->tm_mon] + (tm->tm_mon == 1 ? is_leap : 0)) {
        tm->tm_mday -= days_in_month[tm->tm_mon] + (tm->tm_mon == 1 ? is_leap : 0);
        tm->tm_mon++;
        if (tm->tm_mon > 11) {
            tm->tm_mon = 0;
            tm->tm_year++;
            is_leap = (tm->tm_year % 4 == 0 && (tm->tm_year % 100 != 0 || tm->tm_year % 400 == 0));
        }
    }
    while (tm->tm_mday < 1) {
        tm->tm_mon--;
        if (tm->tm_mon < 0) {
            tm->tm_mon = 11;
            tm->tm_year--;
            is_leap = (tm->tm_year % 4 == 0 && (tm->tm_year % 100 != 0 || tm->tm_year % 400 == 0));
        }
        tm->tm_mday += days_in_month[tm->tm_mon] + (tm->tm_mon == 1 ? is_leap : 0);
    }
}

static int days_between(struct tm *t1, struct tm *t2) {
    long long jd1 = t1->tm_year * 365LL + t1->tm_year / 4 - t1->tm_year / 100 + t1->tm_year / 400;
    int month_days[] = {0,31,59,90,120,151,181,212,243,273,304,334};
    jd1 += month_days[t1->tm_mon] + t1->tm_mday;
    if (t1->tm_mon > 1 && (t1->tm_year % 4 == 0 && (t1->tm_year % 100 != 0 || t1->tm_year % 400 == 0))) {
        jd1++;
    }
    
    long long jd2 = t2->tm_year * 365LL + t2->tm_year / 4 - t2->tm_year / 100 + t2->tm_year / 400;
    jd2 += month_days[t2->tm_mon] + t2->tm_mday;
    if (t2->tm_mon > 1 && (t2->tm_year % 4 == 0 && (t2->tm_year % 100 != 0 || t2->tm_year % 400 == 0))) {
        jd2++;
    }
    
    return (int)(jd2 - jd1);
}

int format_absolute_time(time_t utc_time, const TimeZoneInfo *tz,
                         char *buffer, size_t buffer_size) {
    if (!buffer || buffer_size == 0) return -1;
    
    struct tm local_tm;
    timezone_get_local_time(utc_time, tz, &local_tm);
    
    int result = snprintf(buffer, buffer_size, "%04d-%02d-%02d %02d:%02d:%02d",
                          local_tm.tm_year + 1900,
                          local_tm.tm_mon + 1,
                          local_tm.tm_mday,
                          local_tm.tm_hour,
                          local_tm.tm_min,
                          local_tm.tm_sec);
    
    return (result >= 0 && (size_t)result < buffer_size) ? 0 : -1;
}

int format_relative_time(time_t utc_time, time_t now_utc,
                         const TimeZoneInfo *tz,
                         char *buffer, size_t buffer_size) {
    if (!buffer || buffer_size == 0) return -1;
    
    time_t diff = now_utc - utc_time;
    
    if (diff < 0) {
        return format_absolute_time(utc_time, tz, buffer, buffer_size);
    }
    
    struct tm local_now, local_time;
    timezone_get_local_time(now_utc, tz, &local_now);
    timezone_get_local_time(utc_time, tz, &local_time);
    
    if (diff < 60) {
        snprintf(buffer, buffer_size, "刚刚");
        return 0;
    }
    
    if (diff < 3600) {
        int minutes = (int)(diff / 60);
        snprintf(buffer, buffer_size, "%d分钟前", minutes);
        return 0;
    }
    
    struct tm now_date = local_now;
    struct tm time_date = local_time;
    now_date.tm_hour = 0;
    now_date.tm_min = 0;
    now_date.tm_sec = 0;
    time_date.tm_hour = 0;
    time_date.tm_min = 0;
    time_date.tm_sec = 0;
    
    int days = days_between(&time_date, &now_date);
    
    if (days == 0 && diff < 86400) {
        int hours = (int)(diff / 3600);
        snprintf(buffer, buffer_size, "%d小时前", hours);
        return 0;
    }
    
    if (days == 1) {
        snprintf(buffer, buffer_size, "昨天 %02d:%02d", local_time.tm_hour, local_time.tm_min);
        return 0;
    }
    
    if (days == 2) {
        snprintf(buffer, buffer_size, "前天 %02d:%02d", local_time.tm_hour, local_time.tm_min);
        return 0;
    }
    
    if (days < 30) {
        snprintf(buffer, buffer_size, "%d天前", days);
        return 0;
    }
    
    return format_date_only(utc_time, tz, buffer, buffer_size);
}

int format_date_only(time_t utc_time, const TimeZoneInfo *tz,
                     char *buffer, size_t buffer_size) {
    if (!buffer || buffer_size == 0) return -1;
    
    struct tm local_tm;
    timezone_get_local_time(utc_time, tz, &local_tm);
    
    int result = snprintf(buffer, buffer_size, "%04d-%02d-%02d",
                          local_tm.tm_year + 1900,
                          local_tm.tm_mon + 1,
                          local_tm.tm_mday);
    
    return (result >= 0 && (size_t)result < buffer_size) ? 0 : -1;
}

int format_time_only(time_t utc_time, const TimeZoneInfo *tz,
                     char *buffer, size_t buffer_size) {
    if (!buffer || buffer_size == 0) return -1;
    
    struct tm local_tm;
    timezone_get_local_time(utc_time, tz, &local_tm);
    
    int result = snprintf(buffer, buffer_size, "%02d:%02d",
                          local_tm.tm_hour,
                          local_tm.tm_min);
    
    return (result >= 0 && (size_t)result < buffer_size) ? 0 : -1;
}

int format_friendly_time(time_t utc_time, time_t now_utc,
                         const TimeZoneInfo *tz,
                         char *buffer, size_t buffer_size) {
    if (!buffer || buffer_size == 0) return -1;
    
    time_t diff = now_utc - utc_time;
    
    if (diff < 0) {
        return format_absolute_time(utc_time, tz, buffer, buffer_size);
    }
    
    struct tm local_now, local_time;
    timezone_get_local_time(now_utc, tz, &local_now);
    timezone_get_local_time(utc_time, tz, &local_time);
    
    if (diff < 60) {
        snprintf(buffer, buffer_size, "刚刚");
        return 0;
    }
    
    if (diff < 3600) {
        int minutes = (int)(diff / 60);
        if (minutes == 1) {
            snprintf(buffer, buffer_size, "1分钟前");
        } else {
            snprintf(buffer, buffer_size, "%d分钟前", minutes);
        }
        return 0;
    }
    
    struct tm now_date = local_now;
    struct tm time_date = local_time;
    now_date.tm_hour = 0;
    now_date.tm_min = 0;
    now_date.tm_sec = 0;
    time_date.tm_hour = 0;
    time_date.tm_min = 0;
    time_date.tm_sec = 0;
    
    int days = days_between(&time_date, &now_date);
    const char *period;
    int display_hour = local_time.tm_hour;
    
    if (local_time.tm_hour >= 0 && local_time.tm_hour < 6) {
        period = "凌晨";
    } else if (local_time.tm_hour >= 6 && local_time.tm_hour < 9) {
        period = "早上";
    } else if (local_time.tm_hour >= 9 && local_time.tm_hour < 12) {
        period = "上午";
    } else if (local_time.tm_hour == 12) {
        period = "中午";
    } else if (local_time.tm_hour > 12 && local_time.tm_hour < 18) {
        period = "下午";
        display_hour = local_time.tm_hour - 12;
    } else {
        period = "晚上";
        display_hour = local_time.tm_hour - 12;
        if (display_hour == 0) display_hour = 12;
    }
    
    if (days == 0 && diff < 86400) {
        int hours = (int)(diff / 3600);
        if (hours <= 6) {
            snprintf(buffer, buffer_size, "%d小时前", hours);
        } else {
            if (local_time.tm_min == 0) {
                snprintf(buffer, buffer_size, "%s%d点", period, display_hour);
            } else {
                snprintf(buffer, buffer_size, "%s%d点%d分", period, display_hour, local_time.tm_min);
            }
        }
        return 0;
    }
    
    if (days == 1) {
        if (local_time.tm_min == 0) {
            snprintf(buffer, buffer_size, "昨天%s%d点", period, display_hour);
        } else {
            snprintf(buffer, buffer_size, "昨天%s%d点%d分", period, display_hour, local_time.tm_min);
        }
        return 0;
    }
    
    if (days == 2) {
        if (local_time.tm_min == 0) {
            snprintf(buffer, buffer_size, "前天%s%d点", period, display_hour);
        } else {
            snprintf(buffer, buffer_size, "前天%s%d点%d分", period, display_hour, local_time.tm_min);
        }
        return 0;
    }
    
    if (days < 7) {
        const char *weekdays[] = {"周日", "周一", "周二", "周三", "周四", "周五", "周六"};
        if (local_time.tm_min == 0) {
            snprintf(buffer, buffer_size, "%s%s%d点", weekdays[local_time.tm_wday], period, display_hour);
        } else {
            snprintf(buffer, buffer_size, "%s%s%d点%d分", weekdays[local_time.tm_wday], period, display_hour, local_time.tm_min);
        }
        return 0;
    }
    
    if (days < 30) {
        snprintf(buffer, buffer_size, "%d天前", days);
        return 0;
    }
    
    if (local_time.tm_year == local_now.tm_year) {
        snprintf(buffer, buffer_size, "%d月%d日", local_time.tm_mon + 1, local_time.tm_mday);
        return 0;
    }
    
    snprintf(buffer, buffer_size, "%d年%d月%d日", local_time.tm_year + 1900, local_time.tm_mon + 1, local_time.tm_mday);
    return 0;
}

int format_time(time_t utc_time, time_t now_utc,
                const TimeZoneInfo *tz, TimeFormatType format_type,
                char *buffer, size_t buffer_size) {
    switch (format_type) {
        case FORMAT_ABSOLUTE:
            return format_absolute_time(utc_time, tz, buffer, buffer_size);
        case FORMAT_RELATIVE:
            return format_relative_time(utc_time, now_utc, tz, buffer, buffer_size);
        case FORMAT_DATE_ONLY:
            return format_date_only(utc_time, tz, buffer, buffer_size);
        case FORMAT_TIME_ONLY:
            return format_time_only(utc_time, tz, buffer, buffer_size);
        case FORMAT_FRIENDLY:
            return format_friendly_time(utc_time, now_utc, tz, buffer, buffer_size);
        default:
            return -1;
    }
}
