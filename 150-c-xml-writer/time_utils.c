#include "time_utils.h"
#include <string.h>
#include <stdio.h>

static const char *weekdays[] = {"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"};
static const char *months[] = {"Jan", "Feb", "Mar", "Apr", "May", "Jun", 
                                "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"};

char* Time_ToRFC822(char *buffer, size_t buffer_size, time_t timestamp) {
    if (buffer == NULL || buffer_size < RFC822_DATE_LEN) {
        return NULL;
    }
    
    struct tm *gmt_time = gmtime(&timestamp);
    if (gmt_time == NULL) {
        buffer[0] = '\0';
        return NULL;
    }
    
    snprintf(buffer, buffer_size, "%s, %02d %s %04d %02d:%02d:%02d GMT",
             weekdays[gmt_time->tm_wday],
             gmt_time->tm_mday,
             months[gmt_time->tm_mon],
             gmt_time->tm_year + 1900,
             gmt_time->tm_hour,
             gmt_time->tm_min,
             gmt_time->tm_sec);
    
    return buffer;
}

static int month_to_int(const char *month_name) {
    for (int i = 0; i < 12; i++) {
        if (strcasecmp(month_name, months[i]) == 0) {
            return i;
        }
    }
    return -1;
}

time_t Time_FromRFC822(const char *rfc822_string) {
    if (rfc822_string == NULL) {
        return (time_t)-1;
    }
    
    struct tm tm;
    memset(&tm, 0, sizeof(tm));
    
    char weekday[4], month[4], timezone[8];
    int day, year, hour, minute, second;
    
    int parsed = sscanf(rfc822_string, "%3[^,], %d %3s %d %d:%d:%d %7s",
                        weekday, &day, month, &year, &hour, &minute, &second, timezone);
    
    if (parsed < 7) {
        parsed = sscanf(rfc822_string, "%d %3s %d %d:%d:%d %7s",
                        &day, month, &year, &hour, &minute, &second, timezone);
        if (parsed < 6) {
            return (time_t)-1;
        }
    }
    
    int mon = month_to_int(month);
    if (mon < 0) {
        return (time_t)-1;
    }
    
    tm.tm_mday = day;
    tm.tm_mon = mon;
    tm.tm_year = (year >= 1900) ? (year - 1900) : year;
    tm.tm_hour = hour;
    tm.tm_min = minute;
    tm.tm_sec = second;
    tm.tm_isdst = 0;
    
    time_t result = timegm(&tm);
    if (result == (time_t)-1) {
        return (time_t)-1;
    }
    
    if (parsed >= 8) {
        if (strcasecmp(timezone, "GMT") == 0 || 
            strcasecmp(timezone, "UTC") == 0 ||
            strcmp(timezone, "Z") == 0) {
            return result;
        } else if (timezone[0] == '+' || timezone[0] == '-') {
            int offset_hour, offset_min = 0;
            if (strlen(timezone) == 5) {
                offset_hour = (timezone[1] - '0') * 10 + (timezone[2] - '0');
                offset_min = (timezone[3] - '0') * 10 + (timezone[4] - '0');
            } else if (strlen(timezone) == 3) {
                offset_hour = (timezone[1] - '0') * 10 + (timezone[2] - '0');
            } else {
                return result;
            }
            
            int total_offset = offset_hour * 3600 + offset_min * 60;
            if (timezone[0] == '+') {
                result -= total_offset;
            } else {
                result += total_offset;
            }
        }
    }
    
    return result;
}

time_t Time_ToGMT(time_t local_time) {
    struct tm *local_tm = localtime(&local_time);
    if (local_tm == NULL) {
        return local_time;
    }
    
    struct tm gmt_tm = *local_tm;
    gmt_tm.tm_isdst = 0;
    return timegm(&gmt_tm);
}

time_t Time_FromGMT(time_t gmt_time) {
    struct tm *gmt_tm = gmtime(&gmt_time);
    if (gmt_tm == NULL) {
        return gmt_time;
    }
    
    struct tm local_tm = *gmt_tm;
    local_tm.tm_isdst = -1;
    return mktime(&local_tm);
}
