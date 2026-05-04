#ifndef TIMEZONE_H
#define TIMEZONE_H

#include <time.h>

typedef struct {
    const char *name;
    int offset_seconds;
    int has_dst;
    int dst_offset_seconds;
} TimeZoneInfo;

typedef struct {
    int is_dst;
    int total_offset;
} TimeZoneConversion;

extern const TimeZoneInfo TIMEZONE_UTC;
extern const TimeZoneInfo TIMEZONE_BEIJING;
extern const TimeZoneInfo TIMEZONE_TOKYO;
extern const TimeZoneInfo TIMEZONE_NEWYORK;
extern const TimeZoneInfo TIMEZONE_LONDON;

const TimeZoneInfo *timezone_get_by_name(const char *name);

int timezone_calculate_dst(const TimeZoneInfo *tz, time_t utc_time);

time_t timezone_convert_utc_to_local(time_t utc_time, const TimeZoneInfo *tz);

time_t timezone_convert_local_to_utc(time_t local_time, const TimeZoneInfo *tz);

void timezone_get_local_time(time_t utc_time, const TimeZoneInfo *tz, struct tm *result);

int timezone_offset_to_seconds(int hours, int minutes);

void timezone_seconds_to_offset(int seconds, int *hours, int *minutes);

#endif
