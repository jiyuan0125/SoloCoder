#include "timezone.h"
#include <string.h>
#include <stdlib.h>

const TimeZoneInfo TIMEZONE_UTC = { "UTC", 0, 0, 0 };
const TimeZoneInfo TIMEZONE_BEIJING = { "Asia/Shanghai", 8 * 3600, 0, 0 };
const TimeZoneInfo TIMEZONE_TOKYO = { "Asia/Tokyo", 9 * 3600, 0, 0 };
const TimeZoneInfo TIMEZONE_NEWYORK = { "America/New_York", -5 * 3600, 1, -4 * 3600 };
const TimeZoneInfo TIMEZONE_LONDON = { "Europe/London", 0, 1, 3600 };

static const TimeZoneInfo *TIMEZONES[] = {
    &TIMEZONE_UTC,
    &TIMEZONE_BEIJING,
    &TIMEZONE_TOKYO,
    &TIMEZONE_NEWYORK,
    &TIMEZONE_LONDON,
    NULL
};

const TimeZoneInfo *timezone_get_by_name(const char *name) {
    if (!name) return &TIMEZONE_UTC;
    
    for (int i = 0; TIMEZONES[i] != NULL; i++) {
        if (strcmp(TIMEZONES[i]->name, name) == 0) {
            return TIMEZONES[i];
        }
    }
    return &TIMEZONE_UTC;
}

static int is_european_dst(struct tm *tm) {
    if (tm->tm_mon < 2 || tm->tm_mon > 9) {
        return 0;
    }
    if (tm->tm_mon > 2 && tm->tm_mon < 9) {
        return 1;
    }
    
    int last_sunday;
    if (tm->tm_mon == 2) {
        last_sunday = 31 - ((tm->tm_wday + 31 - tm->tm_mday) % 7);
        if (tm->tm_mday < last_sunday) return 0;
        if (tm->tm_mday > last_sunday) return 1;
        return tm->tm_hour >= 1;
    } else {
        last_sunday = 31 - ((tm->tm_wday + 31 - tm->tm_mday) % 7);
        if (tm->tm_mday < last_sunday) return 1;
        if (tm->tm_mday > last_sunday) return 0;
        return tm->tm_hour < 1;
    }
}

static int is_us_dst(struct tm *tm) {
    if (tm->tm_mon < 1 || tm->tm_mon > 10) {
        return 0;
    }
    if (tm->tm_mon > 1 && tm->tm_mon < 10) {
        return 1;
    }
    
    if (tm->tm_mon == 1) {
        int second_sunday = 8 + (6 - tm->tm_wday + tm->tm_mday - 8) % 7;
        if (second_sunday < 8) second_sunday += 7;
        
        if (tm->tm_mday < second_sunday) return 0;
        if (tm->tm_mday > second_sunday) return 1;
        return tm->tm_hour >= 2;
    } else {
        int first_sunday = 1 + (6 - tm->tm_wday + tm->tm_mday - 1) % 7;
        
        if (tm->tm_mday < first_sunday) return 1;
        if (tm->tm_mday > first_sunday) return 0;
        return tm->tm_hour < 2;
    }
}

int timezone_calculate_dst(const TimeZoneInfo *tz, time_t utc_time) {
    if (!tz || !tz->has_dst) {
        return 0;
    }
    
    struct tm tm;
    gmtime_r(&utc_time, &tm);
    
    if (strcmp(tz->name, "Europe/London") == 0 || 
        strncmp(tz->name, "Europe/", 7) == 0) {
        return is_european_dst(&tm);
    }
    
    if (strcmp(tz->name, "America/New_York") == 0 ||
        strncmp(tz->name, "America/", 8) == 0) {
        return is_us_dst(&tm);
    }
    
    return 0;
}

time_t timezone_convert_utc_to_local(time_t utc_time, const TimeZoneInfo *tz) {
    if (!tz) tz = &TIMEZONE_UTC;
    
    int offset = tz->offset_seconds;
    
    if (tz->has_dst) {
        int is_dst = timezone_calculate_dst(tz, utc_time);
        if (is_dst) {
            offset = tz->dst_offset_seconds;
        }
    }
    
    return utc_time + offset;
}

time_t timezone_convert_local_to_utc(time_t local_time, const TimeZoneInfo *tz) {
    if (!tz) tz = &TIMEZONE_UTC;
    
    time_t candidate_utc = local_time - tz->offset_seconds;
    int is_dst = timezone_calculate_dst(tz, candidate_utc);
    
    int offset = tz->offset_seconds;
    if (is_dst && tz->has_dst) {
        offset = tz->dst_offset_seconds;
    }
    
    return local_time - offset;
}

void timezone_get_local_time(time_t utc_time, const TimeZoneInfo *tz, struct tm *result) {
    time_t local_time = timezone_convert_utc_to_local(utc_time, tz);
    gmtime_r(&local_time, result);
}

int timezone_offset_to_seconds(int hours, int minutes) {
    return hours * 3600 + minutes * 60;
}

void timezone_seconds_to_offset(int seconds, int *hours, int *minutes) {
    if (hours) *hours = seconds / 3600;
    if (minutes) *minutes = (seconds % 3600) / 60;
}
