#ifndef TIME_FORMATTER_H
#define TIME_FORMATTER_H

#include <time.h>
#include "timezone.h"

#define TIME_FORMAT_MAX_LEN 128

typedef enum {
    FORMAT_ABSOLUTE,
    FORMAT_RELATIVE,
    FORMAT_DATE_ONLY,
    FORMAT_TIME_ONLY,
    FORMAT_FRIENDLY
} TimeFormatType;

int format_absolute_time(time_t utc_time, const TimeZoneInfo *tz, 
                         char *buffer, size_t buffer_size);

int format_relative_time(time_t utc_time, time_t now_utc, 
                         const TimeZoneInfo *tz,
                         char *buffer, size_t buffer_size);

int format_date_only(time_t utc_time, const TimeZoneInfo *tz,
                     char *buffer, size_t buffer_size);

int format_time_only(time_t utc_time, const TimeZoneInfo *tz,
                     char *buffer, size_t buffer_size);

int format_friendly_time(time_t utc_time, time_t now_utc,
                         const TimeZoneInfo *tz,
                         char *buffer, size_t buffer_size);

int format_time(time_t utc_time, time_t now_utc,
                const TimeZoneInfo *tz, TimeFormatType format_type,
                char *buffer, size_t buffer_size);

#endif
