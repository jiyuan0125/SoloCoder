#ifndef CHECKIN_COMMON_H
#define CHECKIN_COMMON_H

#include <stdint.h>
#include <stdbool.h>
#include <time.h>

#define MAX_YEARS 10
#define INITIAL_USER_CAPACITY 1024

typedef struct {
    int year;
    int month;
    int day;
} Date;

typedef struct {
    Date start;
    Date end;
} DateRange;

bool is_leap_year(int year);
int days_in_year(int year);
int days_in_month(int year, int month);
int day_of_year(int year, int month, int day);
bool date_to_day_of_year(const Date *date, int *day_of_year);
bool is_valid_date(int year, int month, int day);
int compare_dates(const Date *d1, const Date *d2);
void get_current_date(Date *date);

#endif
