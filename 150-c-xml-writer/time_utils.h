#ifndef TIME_UTILS_H
#define TIME_UTILS_H

#include <time.h>

#define RFC822_DATE_LEN 32

char* Time_ToRFC822(char *buffer, size_t buffer_size, time_t timestamp);
time_t Time_FromRFC822(const char *rfc822_string);

time_t Time_ToGMT(time_t local_time);
time_t Time_FromGMT(time_t gmt_time);

#endif
