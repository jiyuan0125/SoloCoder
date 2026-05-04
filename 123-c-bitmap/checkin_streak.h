#ifndef CHECKIN_STREAK_H
#define CHECKIN_STREAK_H

#include "checkin_common.h"
#include "checkin_bitmap.h"

typedef struct {
    int total_days;
    int longest_streak;
    int current_streak;
} CheckinStats;

int calculate_current_streak(UserBitmap *ub, const Date *today);
int calculate_current_streak_from(UserBitmap *ub, const Date *from_date);
int count_checkins_in_range(UserBitmap *ub, const DateRange *range);
int longest_streak_in_range(UserBitmap *ub, const DateRange *range);
int get_checkin_dates_in_range(UserBitmap *ub, const DateRange *range, 
                                Date *dates, int max_dates);
void get_checkin_stats(UserBitmap *ub, const Date *today, CheckinStats *stats);

#endif
