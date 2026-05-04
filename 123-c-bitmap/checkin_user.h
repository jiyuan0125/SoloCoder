#ifndef CHECKIN_USER_H
#define CHECKIN_USER_H

#include "checkin_common.h"
#include "checkin_bitmap.h"
#include "checkin_streak.h"

typedef uint64_t UserId;

typedef struct {
    UserId id;
    UserBitmap bitmap;
} UserData;

typedef struct {
    UserData *users;
    int user_count;
    int user_capacity;
} UserManager;

void user_manager_init(UserManager *um);
void user_manager_free(UserManager *um);
UserData* user_manager_get_user(UserManager *um, UserId id);
UserData* user_manager_get_or_create_user(UserManager *um, UserId id);
bool user_manager_remove_user(UserManager *um, UserId id);

bool user_checkin(UserData *user, const Date *date);
bool user_uncheckin(UserData *user, const Date *date);
bool user_has_checked_in(UserData *user, const Date *date);
int user_get_current_streak(UserData *user, const Date *today);
int user_count_checkins_in_range(UserData *user, const DateRange *range);
int user_longest_streak_in_range(UserData *user, const DateRange *range);
int user_get_checkin_dates_in_range(UserData *user, const DateRange *range, 
                                      Date *dates, int max_dates);
void user_get_stats(UserData *user, const Date *today, CheckinStats *stats);

#endif
