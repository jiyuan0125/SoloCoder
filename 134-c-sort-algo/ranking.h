#ifndef RANKING_H
#define RANKING_H

#include "student.h"

typedef struct {
    int rank;
    int percentile;
    char student_id[STUDENT_ID_LEN + 1];
} RankEntry;

typedef struct {
    RankEntry* entries;
    int count;
    RankScope scope;
    int filter_value;
} RankResult;

typedef int (*CompareFunc)(const void*, const void*);

RankResult* rank_create(int capacity);
void rank_destroy(RankResult* result);

void rank_students(StudentDatabase* db, RankScope scope, int filter, RankResult* result);
void rank_students_by_indices(StudentDatabase* db, int* indices, int count, RankResult* result);

int calculate_percentile(int rank, int total);
int is_tie(const Student* a, const Student* b);

int compare_for_class(const void* a, const void* b, int class_num);
int compare_for_track(const void* a, const void* b, TrackType track);

#endif
