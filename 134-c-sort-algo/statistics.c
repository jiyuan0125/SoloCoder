#include "statistics.h"
#include "ranking.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <math.h>

static const char* score_ranges[] = {
    "0-9",   "10-19", "20-29", "30-39", "40-49",
    "50-59", "60-69", "70-79", "80-89", "90-100",
    "缺考"
};

StatsResult* stats_create(void)
{
    StatsResult* result = (StatsResult*)calloc(1, sizeof(StatsResult));
    if (!result) return NULL;
    
    for (int s = 0; s < SUBJECT_COUNT; s++) {
        result->subjects[s].max_score = 0;
        result->subjects[s].min_score = 100;
        result->subjects[s].average = 0.0;
        result->subjects[s].median = 0.0;
        result->subjects[s].stddev = 0.0;
        result->subjects[s].taken_count = 0;
        result->subjects[s].missing_count = 0;
        memset(result->subjects[s].distribution, 0, sizeof(int) * SCORE_RANGE_COUNT);
    }
    
    return result;
}

void stats_destroy(StatsResult* result)
{
    if (result) {
        free(result);
    }
}

int score_range_index(int score)
{
    if (score < 0) return 0;
    if (score >= 100) return 9;
    return score / 10;
}

const char* score_range_name(int index)
{
    if (index >= 0 && index < SCORE_RANGE_COUNT) {
        return score_ranges[index];
    }
    return "未知";
}

static int int_compare(const void* a, const void* b)
{
    return *(const int*)a - *(const int*)b;
}

double calculate_median(int* values, int count)
{
    if (!values || count <= 0) return 0.0;
    
    qsort(values, count, sizeof(int), int_compare);
    
    if (count % 2 == 1) {
        return (double)values[count / 2];
    } else {
        int mid1 = values[count / 2 - 1];
        int mid2 = values[count / 2];
        return (double)(mid1 + mid2) / 2.0;
    }
}

double calculate_stddev(int* values, int count, double average)
{
    if (!values || count <= 1) return 0.0;
    
    double sum_sq_diff = 0.0;
    for (int i = 0; i < count; i++) {
        double diff = (double)values[i] - average;
        sum_sq_diff += diff * diff;
    }
    
    return sqrt(sum_sq_diff / (double)(count - 1));
}

static void calc_stats_for_subject(StudentDatabase* db, SubjectType subj,
                                    int* indices, int filtered_count,
                                    SubjectStats* stats)
{
    if (!db || !stats) return;
    
    stats->max_score = 0;
    stats->min_score = 100;
    stats->average = 0.0;
    stats->median = 0.0;
    stats->stddev = 0.0;
    stats->taken_count = 0;
    stats->missing_count = 0;
    memset(stats->distribution, 0, sizeof(int) * SCORE_RANGE_COUNT);
    
    int count = filtered_count > 0 ? filtered_count : db->count;
    long long sum = 0;
    
    int* scores_buffer = (int*)malloc(sizeof(int) * count);
    if (!scores_buffer) return;
    
    int buffer_idx = 0;
    
    for (int i = 0; i < count; i++) {
        int student_idx = indices ? indices[i] : i;
        Student* s = &db->students[student_idx];
        
        if (s->missing[subj]) {
            stats->missing_count++;
            stats->distribution[10]++;
        } else {
            int score = s->scores[subj];
            stats->taken_count++;
            sum += score;
            scores_buffer[buffer_idx++] = score;
            
            if (score > stats->max_score) stats->max_score = score;
            if (score < stats->min_score) stats->min_score = score;
            
            int range_idx = score_range_index(score);
            stats->distribution[range_idx]++;
        }
    }
    
    if (stats->taken_count > 0) {
        stats->average = (double)sum / (double)stats->taken_count;
        stats->median = calculate_median(scores_buffer, buffer_idx);
        stats->stddev = calculate_stddev(scores_buffer, buffer_idx, stats->average);
    }
    
    free(scores_buffer);
}

void calculate_statistics(StudentDatabase* db, StatsResult* result)
{
    if (!db || !result) return;
    
    result->total_students = db->count;
    
    for (int s = 0; s < SUBJECT_COUNT; s++) {
        calc_stats_for_subject(db, (SubjectType)s, NULL, 0, &result->subjects[s]);
        result->total_missing[s] = result->subjects[s].missing_count;
    }
}

static int get_filtered_indices(StudentDatabase* db, RankScope scope, int filter,
                                 int** out_indices, int* out_count)
{
    if (!db || !out_indices || !out_count) return -1;
    
    int* indices = (int*)malloc(sizeof(int) * db->count);
    if (!indices) return -1;
    
    int count = 0;
    
    for (int i = 0; i < db->count; i++) {
        int match = 0;
        
        if (scope == RANK_CLASS) {
            match = (db->students[i].class_num == filter);
        } else if (scope == RANK_TRACK) {
            match = (db->students[i].track == (TrackType)filter);
        }
        
        if (match) {
            indices[count++] = i;
        }
    }
    
    *out_indices = indices;
    *out_count = count;
    return 0;
}

void calculate_statistics_for_filter(StudentDatabase* db, StatsResult* result,
                                      RankScope scope, int filter)
{
    if (!db || !result) return;
    
    int* indices = NULL;
    int count = 0;
    
    if (scope != RANK_GRADE) {
        if (get_filtered_indices(db, scope, filter, &indices, &count) != 0) {
            return;
        }
        result->total_students = count;
    } else {
        result->total_students = db->count;
    }
    
    for (int s = 0; s < SUBJECT_COUNT; s++) {
        calc_stats_for_subject(db, (SubjectType)s, indices, count, &result->subjects[s]);
        result->total_missing[s] = result->subjects[s].missing_count;
    }
    
    if (indices) {
        free(indices);
    }
}
