#include "checkin_common.h"
#include <stdlib.h>
#include <string.h>

static const int days_per_month[] = {0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31};

bool is_leap_year(int year) {
    if (year % 4 != 0) return false;
    if (year % 100 != 0) return true;
    if (year % 400 != 0) return false;
    return true;
}

int days_in_year(int year) {
    return is_leap_year(year) ? 366 : 365;
}

int days_in_month(int year, int month) {
    if (month < 1 || month > 12) return 0;
    int days = days_per_month[month];
    if (month == 2 && is_leap_year(year)) {
        days++;
    }
    return days;
}

int day_of_year(int year, int month, int day) {
    int doy = 0;
    for (int i = 1; i < month; i++) {
        doy += days_in_month(year, i);
    }
    doy += day;
    return doy;
}

bool date_to_day_of_year(const Date *date, int *day_of_year_out) {
    if (!is_valid_date(date->year, date->month, date->day)) {
        return false;
    }
    *day_of_year_out = day_of_year(date->year, date->month, date->day);
    return true;
}

bool is_valid_date(int year, int month, int day) {
    if (year < 1970 || year > 2100) return false;
    if (month < 1 || month > 12) return false;
    int max_day = days_in_month(year, month);
    if (day < 1 || day > max_day) return false;
    return true;
}

int compare_dates(const Date *d1, const Date *d2) {
    if (d1->year != d2->year) return d1->year - d2->year;
    if (d1->month != d2->month) return d1->month - d2->month;
    return d1->day - d2->day;
}

void get_current_date(Date *date) {
    time_t t = time(NULL);
    struct tm *tm_info = localtime(&t);
    date->year = tm_info->tm_year + 1900;
    date->month = tm_info->tm_mon + 1;
    date->day = tm_info->tm_mday;
}
