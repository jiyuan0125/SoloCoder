#ifndef CHECKIN_BITMAP_H
#define CHECKIN_BITMAP_H

#include "checkin_common.h"

#define BITS_PER_BYTE 8
#define BITS_PER_UINT64 64

typedef struct {
    int year;
    uint8_t *bits;
    int size_bytes;
    int total_days;
} YearBitmap;

typedef struct {
    YearBitmap *years;
    int year_count;
    int year_capacity;
} UserBitmap;

void year_bitmap_init(YearBitmap *yb, int year);
void year_bitmap_free(YearBitmap *yb);
bool year_bitmap_set(YearBitmap *yb, int day_of_year);
bool year_bitmap_clear(YearBitmap *yb, int day_of_year);
bool year_bitmap_check(const YearBitmap *yb, int day_of_year);
int year_bitmap_count_range(const YearBitmap *yb, int start_day, int end_day);
int year_bitmap_longest_streak_range(const YearBitmap *yb, int start_day, int end_day);
int year_bitmap_get_dates_in_range(const YearBitmap *yb, int start_day, int end_day, 
                                    int *dates, int max_dates);
int year_bitmap_trailing_ones(const YearBitmap *yb);
int year_bitmap_trailing_ones_from(const YearBitmap *yb, int from_day);
uint64_t year_bitmap_get_uint64(const YearBitmap *yb, int index);

void user_bitmap_init(UserBitmap *ub);
void user_bitmap_free(UserBitmap *ub);
YearBitmap* user_bitmap_get_year(UserBitmap *ub, int year);
YearBitmap* user_bitmap_get_or_create_year(UserBitmap *ub, int year);

#endif
