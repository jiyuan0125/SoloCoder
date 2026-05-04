#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include "checkin_user.h"
#include "checkin_common.h"

#define NUM_USERS 500
#define MAX_CHECKIN_DATES 100

static void generate_random_checkins(UserData *user, int start_year, int end_year) {
    for (int year = start_year; year <= end_year; year++) {
        int days = days_in_year(year);
        for (int doy = 1; doy <= days; doy++) {
            if (rand() % 100 < 70) {
                Date date;
                date.year = year;
                date.month = 1;
                date.day = doy;
                while (date.day > days_in_month(date.year, date.month)) {
                    date.day -= days_in_month(date.year, date.month);
                    date.month++;
                }
                user_checkin(user, &date);
            }
        }
    }
}

static void generate_continuous_streak(UserData *user, const Date *start, int days) {
    Date current = *start;
    for (int i = 0; i < days; i++) {
        user_checkin(user, &current);
        current.day++;
        int max_day = days_in_month(current.year, current.month);
        if (current.day > max_day) {
            current.day = 1;
            current.month++;
            if (current.month > 12) {
                current.month = 1;
                current.year++;
            }
        }
    }
}

static void print_date(const Date *date) {
    printf("%04d-%02d-%02d", date->year, date->month, date->day);
}

int main() {
    srand(time(NULL));

    printf("========================================\n");
    printf("    用户签到记录模块演示\n");
    printf("========================================\n\n");

    UserManager um;
    user_manager_init(&um);

    printf("步骤 1: 创建 %d 个用户并生成随机签到数据...\n", NUM_USERS);
    printf("        (2024年、2025年、2026年三年数据)\n\n");

    for (UserId id = 1; id <= NUM_USERS; id++) {
        UserData *user = user_manager_get_or_create_user(&um, id);
        generate_random_checkins(user, 2024, 2026);
    }

    printf("已创建 %d 个用户，每个用户约 70%% 的签到概率\n\n", um.user_count);

    printf("步骤 2: 演示跨年连续签到计算...\n\n");

    UserData *special_user = user_manager_get_or_create_user(&um, 99999);
    
    Date dec30_2025 = {2025, 12, 30};
    Date dec31_2025 = {2025, 12, 31};
    Date jan1_2026 = {2026, 1, 1};
    Date jan2_2026 = {2026, 1, 2};
    
    user_checkin(special_user, &dec30_2025);
    user_checkin(special_user, &dec31_2025);
    user_checkin(special_user, &jan1_2026);
    user_checkin(special_user, &jan2_2026);

    printf("用户 99999 签到记录:\n");
    printf("  2025-12-30: %s\n", user_has_checked_in(special_user, &dec30_2025) ? "已签到" : "未签到");
    printf("  2025-12-31: %s\n", user_has_checked_in(special_user, &dec31_2025) ? "已签到" : "未签到");
    printf("  2026-01-01: %s\n", user_has_checked_in(special_user, &jan1_2026) ? "已签到" : "未签到");
    printf("  2026-01-02: %s\n", user_has_checked_in(special_user, &jan2_2026) ? "已签到" : "未签到");

    int streak = user_get_current_streak(special_user, &jan2_2026);
    printf("\n从 2026-01-02 往前计算连续签到天数: %d 天 (期望: 4 天)\n\n", streak);

    printf("步骤 3: 演示区间统计功能...\n\n");

    UserData *test_user = user_manager_get_or_create_user(&um, 88888);
    
    Date jan1_2025 = {2025, 1, 1};
    Date jan10_2025 = {2025, 1, 10};
    Date jan20_2025 = {2025, 1, 20};
    Date jan31_2025 = {2025, 1, 31};
    
    for (int i = 1; i <= 10; i++) {
        Date d = {2025, 1, i};
        user_checkin(test_user, &d);
    }
    for (int i = 15; i <= 25; i++) {
        Date d = {2025, 1, i};
        user_checkin(test_user, &d);
    }

    DateRange jan_range = {
        .start = {2025, 1, 1},
        .end = {2025, 1, 31}
    };

    int total = user_count_checkins_in_range(test_user, &jan_range);
    int longest = user_longest_streak_in_range(test_user, &jan_range);

    printf("用户 88888 2025年1月签到情况:\n");
    printf("  1-10日: 连续签到\n");
    printf("  11-14日: 未签到\n");
    printf("  15-25日: 连续签到\n");
    printf("  26-31日: 未签到\n\n");
    printf("  1月总签到天数: %d 天 (期望: 21 天)\n", total);
    printf("  1月最长连续签到: %d 天 (期望: 11 天)\n\n", longest);

    printf("步骤 4: 查询签到日期列表...\n\n");

    Date checkin_dates[31];
    int date_count = user_get_checkin_dates_in_range(test_user, &jan_range, checkin_dates, 31);

    printf("用户 88888 2025年1月签到日期列表 (%d 天):\n", date_count);
    for (int i = 0; i < date_count; i++) {
        printf("  ");
        print_date(&checkin_dates[i]);
        if ((i + 1) % 7 == 0) printf("\n");
    }
    printf("\n\n");

    printf("步骤 5: 演示清除签到记录...\n\n");

    Date jan5_2025 = {2025, 1, 5};
    printf("用户 88888 在 2025-01-05 的签到状态: %s\n", 
           user_has_checked_in(test_user, &jan5_2025) ? "已签到" : "未签到");
    
    user_uncheckin(test_user, &jan5_2025);
    printf("清除 2025-01-05 签到记录后: %s\n", 
           user_has_checked_in(test_user, &jan5_2025) ? "已签到" : "未签到");

    total = user_count_checkins_in_range(test_user, &jan_range);
    printf("1月总签到天数变为: %d 天 (期望: 20 天)\n\n", total);

    printf("步骤 6: 随机用户统计信息...\n\n");

    for (int i = 0; i < 5; i++) {
        UserId random_id = (rand() % NUM_USERS) + 1;
        UserData *user = user_manager_get_user(&um, random_id);
        if (user) {
            CheckinStats stats;
            Date today = {2026, 5, 4};
            user_get_stats(user, &today, &stats);
            
            printf("用户 %lu:\n", (unsigned long)user->id);
            printf("  总签到天数: %d\n", stats.total_days);
            printf("  历史最长连续: %d 天\n", stats.longest_streak);
            printf("  当前连续签到(截至2026-05-04): %d 天\n", stats.current_streak);
            printf("\n");
        }
    }

    printf("========================================\n");
    printf("    演示完成\n");
    printf("========================================\n\n");

    printf("内存使用统计:\n");
    printf("  用户数: %d\n", um.user_count);
    printf("  每个用户每年存储: 约 %d 字节 (闰年 %d 字节)\n", 
           (365 + 7) / 8, (366 + 7) / 8);
    printf("  500 用户 x 3 年 ≈ %d KB\n", 
           (500 * 3 * 46) / 1024);

    user_manager_free(&um);

    return 0;
}
