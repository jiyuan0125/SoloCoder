#include "checkin_user.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

void user_manager_init(UserManager *um) {
    um->users = NULL;
    um->user_count = 0;
    um->user_capacity = 0;
}

void user_manager_free(UserManager *um) {
    for (int i = 0; i < um->user_count; i++) {
        user_bitmap_free(&um->users[i].bitmap);
    }
    free(um->users);
    um->users = NULL;
    um->user_count = 0;
    um->user_capacity = 0;
}

UserData* user_manager_get_user(UserManager *um, UserId id) {
    for (int i = 0; i < um->user_count; i++) {
        if (um->users[i].id == id) {
            return &um->users[i];
        }
    }
    return NULL;
}

UserData* user_manager_get_or_create_user(UserManager *um, UserId id) {
    UserData *existing = user_manager_get_user(um, id);
    if (existing) {
        return existing;
    }

    if (um->user_count >= um->user_capacity) {
        int new_capacity = um->user_capacity == 0 ? INITIAL_USER_CAPACITY : um->user_capacity * 2;
        UserData *new_users = (UserData *)realloc(um->users, 
                                                    new_capacity * sizeof(UserData));
        if (!new_users) {
            return NULL;
        }
        um->users = new_users;
        um->user_capacity = new_capacity;
    }

    UserData *new_user = &um->users[um->user_count];
    new_user->id = id;
    user_bitmap_init(&new_user->bitmap);
    um->user_count++;
    return new_user;
}

bool user_manager_remove_user(UserManager *um, UserId id) {
    for (int i = 0; i < um->user_count; i++) {
        if (um->users[i].id == id) {
            user_bitmap_free(&um->users[i].bitmap);
            if (i < um->user_count - 1) {
                memmove(&um->users[i], &um->users[i + 1], 
                        (um->user_count - i - 1) * sizeof(UserData));
            }
            um->user_count--;
            return true;
        }
    }
    return false;
}

bool user_checkin(UserData *user, const Date *date) {
    if (!is_valid_date(date->year, date->month, date->day)) {
        return false;
    }
    YearBitmap *yb = user_bitmap_get_or_create_year(&user->bitmap, date->year);
    if (!yb) {
        return false;
    }
    int doy;
    if (!date_to_day_of_year(date, &doy)) {
        return false;
    }
    return year_bitmap_set(yb, doy);
}

bool user_uncheckin(UserData *user, const Date *date) {
    if (!is_valid_date(date->year, date->month, date->day)) {
        return false;
    }
    YearBitmap *yb = user_bitmap_get_year(&user->bitmap, date->year);
    if (!yb) {
        return false;
    }
    int doy;
    if (!date_to_day_of_year(date, &doy)) {
        return false;
    }
    return year_bitmap_clear(yb, doy);
}

bool user_has_checked_in(UserData *user, const Date *date) {
    if (!is_valid_date(date->year, date->month, date->day)) {
        return false;
    }
    YearBitmap *yb = user_bitmap_get_year(&user->bitmap, date->year);
    if (!yb) {
        return false;
    }
    int doy;
    if (!date_to_day_of_year(date, &doy)) {
        return false;
    }
    return year_bitmap_check(yb, doy);
}

int user_get_current_streak(UserData *user, const Date *today) {
    return calculate_current_streak(&user->bitmap, today);
}

int user_count_checkins_in_range(UserData *user, const DateRange *range) {
    return count_checkins_in_range(&user->bitmap, range);
}

int user_longest_streak_in_range(UserData *user, const DateRange *range) {
    return longest_streak_in_range(&user->bitmap, range);
}

int user_get_checkin_dates_in_range(UserData *user, const DateRange *range, 
                                      Date *dates, int max_dates) {
    return get_checkin_dates_in_range(&user->bitmap, range, dates, max_dates);
}

void user_get_stats(UserData *user, const Date *today, CheckinStats *stats) {
    get_checkin_stats(&user->bitmap, today, stats);
}
