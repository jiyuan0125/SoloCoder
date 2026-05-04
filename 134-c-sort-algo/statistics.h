#ifndef STATISTICS_H
#define STATISTICS_H

#include "student.h"

#define SCORE_RANGE_COUNT 11

typedef struct {
    double average;
    int max_score;
    int min_score;
    double median;
    double stddev;
    int taken_count;
    int missing_count;
    int distribution[SCORE_RANGE_COUNT];
} SubjectStats;

typedef struct {
    SubjectStats subjects[SUBJECT_COUNT];
    int total_students;
    int total_missing[SUBJECT_COUNT];
} StatsResult;

StatsResult* stats_create(void);
void stats_destroy(StatsResult* result);

void calculate_statistics(StudentDatabase* db, StatsResult* result);
void calculate_statistics_for_filter(StudentDatabase* db, StatsResult* result,
                                      RankScope scope, int filter);

int score_range_index(int score);
const char* score_range_name(int index);

double calculate_median(int* values, int count);
double calculate_stddev(int* values, int count, double average);

#endif
