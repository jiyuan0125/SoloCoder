#include "checkin_streak.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

static void day_of_year_to_date(int year, int day_of_year, Date *date) {
    date->year = year;
    date->month = 1;
    date->day = day_of_year;
    
    while (date->day > days_in_month(year, date->month)) {
        date->day -= days_in_month(year, date->month);
        date->month++;
    }
}

int calculate_current_streak(UserBitmap *ub, const Date *today) {
    int today_doy;
    if (!date_to_day_of_year(today, &today_doy)) {
        return 0;
    }

    YearBitmap *yb = user_bitmap_get_year(ub, today->year);
    if (!yb) {
        return 0;
    }

    if (!year_bitmap_check(yb, today_doy)) {
        return 0;
    }

    int streak = year_bitmap_trailing_ones_from(yb, today_doy);

    if (streak < today_doy) {
        return streak;
    }

    int current_year = today->year - 1;
    while (true) {
        YearBitmap *current_yb = user_bitmap_get_year(ub, current_year);
        if (!current_yb) {
            break;
        }

        int year_trailing = year_bitmap_trailing_ones(current_yb);
        int year_total = days_in_year(current_year);

        streak += year_trailing;

        if (year_trailing < year_total) {
            break;
        }

        current_year--;
    }

    return streak;
}

int calculate_current_streak_from(UserBitmap *ub, const Date *from_date) {
    int from_doy;
    if (!date_to_day_of_year(from_date, &from_doy)) {
        return 0;
    }

    YearBitmap *yb = user_bitmap_get_year(ub, from_date->year);
    if (!yb) {
        return 0;
    }

    if (!year_bitmap_check(yb, from_doy)) {
        return 0;
    }

    int streak = year_bitmap_trailing_ones_from(yb, from_doy);

    if (streak < from_doy) {
        return streak;
    }

    int current_year = from_date->year - 1;
    while (true) {
        YearBitmap *current_yb = user_bitmap_get_year(ub, current_year);
        if (!current_yb) {
            break;
        }

        int year_trailing = year_bitmap_trailing_ones(current_yb);
        int year_total = days_in_year(current_year);

        streak += year_trailing;

        if (year_trailing < year_total) {
            break;
        }

        current_year--;
    }

    return streak;
}

int count_checkins_in_range(UserBitmap *ub, const DateRange *range) {
    int total = 0;

    if (range->start.year == range->end.year) {
        YearBitmap *yb = user_bitmap_get_year(ub, range->start.year);
        if (!yb) {
            return 0;
        }
        int start_doy, end_doy;
        if (!date_to_day_of_year(&range->start, &start_doy) ||
            !date_to_day_of_year(&range->end, &end_doy)) {
            return 0;
        }
        return year_bitmap_count_range(yb, start_doy, end_doy);
    }

    int start_doy, end_doy;

    YearBitmap *start_yb = user_bitmap_get_year(ub, range->start.year);
    if (start_yb && date_to_day_of_year(&range->start, &start_doy)) {
        int year_days = days_in_year(range->start.year);
        total += year_bitmap_count_range(start_yb, start_doy, year_days);
    }

    for (int year = range->start.year + 1; year < range->end.year; year++) {
        YearBitmap *yb = user_bitmap_get_year(ub, year);
        if (yb) {
            int year_days = days_in_year(year);
            total += year_bitmap_count_range(yb, 1, year_days);
        }
    }

    YearBitmap *end_yb = user_bitmap_get_year(ub, range->end.year);
    if (end_yb && date_to_day_of_year(&range->end, &end_doy)) {
        total += year_bitmap_count_range(end_yb, 1, end_doy);
    }

    return total;
}

int longest_streak_in_range(UserBitmap *ub, const DateRange *range) {
    int longest = 0;
    int current = 0;

    Date current_date = range->start;
    while (compare_dates(&current_date, &range->end) <= 0) {
        int doy;
        if (!date_to_day_of_year(&current_date, &doy)) {
            break;
        }

        YearBitmap *yb = user_bitmap_get_year(ub, current_date.year);
        bool checked_in = yb ? year_bitmap_check(yb, doy) : false;

        if (checked_in) {
            current++;
            if (current > longest) {
                longest = current;
            }
        } else {
            current = 0;
        }

        current_date.day++;
        int max_day = days_in_month(current_date.year, current_date.month);
        if (current_date.day > max_day) {
            current_date.day = 1;
            current_date.month++;
            if (current_date.month > 12) {
                current_date.month = 1;
                current_date.year++;
            }
        }
    }

    return longest;
}

int get_checkin_dates_in_range(UserBitmap *ub, const DateRange *range, 
                                Date *dates, int max_dates) {
    if (max_dates <= 0) return 0;

    int count = 0;
    Date current_date = range->start;

    while (compare_dates(&current_date, &range->end) <= 0 && count < max_dates) {
        int doy;
        if (!date_to_day_of_year(&current_date, &doy)) {
            break;
        }

        YearBitmap *yb = user_bitmap_get_year(ub, current_date.year);
        if (yb && year_bitmap_check(yb, doy)) {
            dates[count++] = current_date;
        }

        current_date.day++;
        int max_day = days_in_month(current_date.year, current_date.month);
        if (current_date.day > max_day) {
            current_date.day = 1;
            current_date.month++;
            if (current_date.month > 12) {
                current_date.month = 1;
                current_date.year++;
            }
        }
    }

    return count;
}

void get_checkin_stats(UserBitmap *ub, const Date *today, CheckinStats *stats) {
    memset(stats, 0, sizeof(CheckinStats));
    
    stats->current_streak = calculate_current_streak(ub, today);

    int total = 0;
    int longest = 0;

    for (int i = 0; i < ub->year_count; i++) {
        YearBitmap *yb = &ub->years[i];
        int year_days = days_in_year(yb->year);
        total += year_bitmap_count_range(yb, 1, year_days);
        
        DateRange year_range = {
            .start = {.year = yb->year, .month = 1, .day = 1},
            .end = {.year = yb->year, .month = 12, .day = days_in_month(yb->year, 12)}
        };
        
        int year_longest = year_bitmap_longest_streak_range(yb, 1, year_days);
        if (year_longest > longest) {
            longest = year_longest;
        }
    }

    stats->total_days = total;
    stats->longest_streak = longest;
}
