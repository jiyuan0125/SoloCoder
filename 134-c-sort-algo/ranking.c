#include "ranking.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

RankResult* rank_create(int capacity)
{
    RankResult* result = (RankResult*)malloc(sizeof(RankResult));
    if (!result) return NULL;
    
    result->count = 0;
    result->scope = RANK_GRADE;
    result->filter_value = 0;
    result->entries = (RankEntry*)malloc(sizeof(RankEntry) * capacity);
    
    if (!result->entries) {
        free(result);
        return NULL;
    }
    
    return result;
}

void rank_destroy(RankResult* result)
{
    if (result) {
        free(result->entries);
        free(result);
    }
}

int calculate_percentile(int rank, int total)
{
    if (total <= 0) return 0;
    if (rank <= 0) return 100;
    if (rank >= total) return 0;
    
    int better = total - rank;
    double percentile = (double)better / total * 100.0;
    
    if (percentile >= 100.0) return 99;
    if (percentile < 0.0) return 0;
    
    return (int)(percentile + 0.5);
}

int is_tie(const Student* a, const Student* b)
{
    if (a->total_score != b->total_score) return 0;
    
    for (int i = 0; i < SUBJECT_COUNT; i++) {
        int score_a = a->missing[i] ? 0 : a->scores[i];
        int score_b = b->missing[i] ? 0 : b->scores[i];
        if (score_a != score_b) return 0;
    }
    
    return 1;
}

static void assign_ranks_with_ties(RankResult* result, Student* sorted_students, int count)
{
    if (count <= 0 || !result) return;
    
    int current_rank = 1;
    int same_rank_count = 1;
    
    strncpy(result->entries[0].student_id, sorted_students[0].id, STUDENT_ID_LEN);
    result->entries[0].student_id[STUDENT_ID_LEN] = '\0';
    result->entries[0].rank = 1;
    result->entries[0].percentile = calculate_percentile(1, count);
    
    for (int i = 1; i < count; i++) {
        strncpy(result->entries[i].student_id, sorted_students[i].id, STUDENT_ID_LEN);
        result->entries[i].student_id[STUDENT_ID_LEN] = '\0';
        
        if (is_tie(&sorted_students[i], &sorted_students[i-1])) {
            result->entries[i].rank = current_rank;
            same_rank_count++;
        } else {
            current_rank = current_rank + same_rank_count;
            result->entries[i].rank = current_rank;
            same_rank_count = 1;
        }
        
        result->entries[i].percentile = calculate_percentile(result->entries[i].rank, count);
    }
    
    result->count = count;
}

void rank_students_by_indices(StudentDatabase* db, int* indices, int count, RankResult* result)
{
    if (!db || !indices || count <= 0 || !result) return;
    
    Student* temp = (Student*)malloc(sizeof(Student) * count);
    if (!temp) return;
    
    for (int i = 0; i < count; i++) {
        temp[i] = db->students[indices[i]];
    }
    
    qsort(temp, count, sizeof(Student), student_compare_rank);
    assign_ranks_with_ties(result, temp, count);
    
    free(temp);
}

static int filter_by_class(const Student* s, int class_num)
{
    return (s->class_num == class_num);
}

static int filter_by_track(const Student* s, TrackType track)
{
    return (s->track == track);
}

void rank_students(StudentDatabase* db, RankScope scope, int filter, RankResult* result)
{
    if (!db || db->count <= 0 || !result) return;
    
    result->scope = scope;
    result->filter_value = filter;
    
    if (scope == RANK_GRADE) {
        Student* temp = (Student*)malloc(sizeof(Student) * db->count);
        if (!temp) return;
        
        for (int i = 0; i < db->count; i++) {
            temp[i] = db->students[i];
        }
        
        qsort(temp, db->count, sizeof(Student), student_compare_rank);
        assign_ranks_with_ties(result, temp, db->count);
        
        free(temp);
        return;
    }
    
    int* indices = (int*)malloc(sizeof(int) * db->count);
    if (!indices) return;
    
    int filtered_count = 0;
    for (int i = 0; i < db->count; i++) {
        int match = 0;
        if (scope == RANK_CLASS) {
            match = filter_by_class(&db->students[i], filter);
        } else if (scope == RANK_TRACK) {
            match = filter_by_track(&db->students[i], (TrackType)filter);
        }
        
        if (match) {
            indices[filtered_count++] = i;
        }
    }
    
    Student* temp = (Student*)malloc(sizeof(Student) * filtered_count);
    if (!temp) {
        free(indices);
        return;
    }
    
    for (int i = 0; i < filtered_count; i++) {
        temp[i] = db->students[indices[i]];
    }
    
    qsort(temp, filtered_count, sizeof(Student), student_compare_rank);
    assign_ranks_with_ties(result, temp, filtered_count);
    
    free(indices);
    free(temp);
}

int compare_for_class(const void* a, const void* b, int class_num)
{
    (void)class_num;
    return student_compare_rank(a, b);
}

int compare_for_track(const void* a, const void* b, TrackType track)
{
    (void)track;
    return student_compare_rank(a, b);
}
