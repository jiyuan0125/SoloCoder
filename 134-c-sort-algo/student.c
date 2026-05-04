#include "student.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

static const char* subject_names[] = {
    "语文", "数学", "英语", "物理", "化学"
};

static const char* track_names[] = {
    "文科", "理科"
};

StudentDatabase* student_db_create(int initial_capacity)
{
    StudentDatabase* db = (StudentDatabase*)malloc(sizeof(StudentDatabase));
    if (!db) return NULL;
    
    db->capacity = initial_capacity > 0 ? initial_capacity : 100;
    db->count = 0;
    db->students = (Student*)malloc(sizeof(Student) * db->capacity);
    
    if (!db->students) {
        free(db);
        return NULL;
    }
    
    return db;
}

void student_db_destroy(StudentDatabase* db)
{
    if (db) {
        free(db->students);
        free(db);
    }
}

static int student_db_resize(StudentDatabase* db)
{
    int new_capacity = db->capacity * 2;
    Student* new_students = (Student*)realloc(db->students, 
                                                 sizeof(Student) * new_capacity);
    if (!new_students) return -1;
    
    db->students = new_students;
    db->capacity = new_capacity;
    return 0;
}

int student_db_add(StudentDatabase* db, const char* id, int class_num,
                   TrackType track, const int scores[], const int missing[])
{
    if (!db || !id || !scores || !missing) return -1;
    
    if (db->count >= db->capacity) {
        if (student_db_resize(db) != 0) {
            return -1;
        }
    }
    
    Student* s = &db->students[db->count];
    memset(s, 0, sizeof(Student));
    
    strncpy(s->id, id, STUDENT_ID_LEN);
    s->id[STUDENT_ID_LEN] = '\0';
    s->class_num = class_num;
    s->track = track;
    
    memcpy(s->scores, scores, sizeof(int) * SUBJECT_COUNT);
    memcpy(s->missing, missing, sizeof(int) * SUBJECT_COUNT);
    
    student_calculate_total(s);
    db->count++;
    
    return db->count - 1;
}

void student_calculate_total(Student* s)
{
    if (!s) return;
    
    s->total_score = 0;
    for (int i = 0; i < SUBJECT_COUNT; i++) {
        if (s->missing[i]) {
            s->total_score += 0;
        } else {
            s->total_score += s->scores[i];
        }
    }
}

int student_compare_rank(const void* a, const void* b)
{
    const Student* sa = (const Student*)a;
    const Student* sb = (const Student*)b;
    
    if (sa->total_score != sb->total_score) {
        return sb->total_score - sa->total_score;
    }
    
    int sa_math = sa->missing[SUBJECT_MATH] ? 0 : sa->scores[SUBJECT_MATH];
    int sb_math = sb->missing[SUBJECT_MATH] ? 0 : sb->scores[SUBJECT_MATH];
    if (sa_math != sb_math) {
        return sb_math - sa_math;
    }
    
    int sa_chinese = sa->missing[SUBJECT_CHINESE] ? 0 : sa->scores[SUBJECT_CHINESE];
    int sb_chinese = sb->missing[SUBJECT_CHINESE] ? 0 : sb->scores[SUBJECT_CHINESE];
    if (sa_chinese != sb_chinese) {
        return sb_chinese - sa_chinese;
    }
    
    int sa_english = sa->missing[SUBJECT_ENGLISH] ? 0 : sa->scores[SUBJECT_ENGLISH];
    int sb_english = sb->missing[SUBJECT_ENGLISH] ? 0 : sb->scores[SUBJECT_ENGLISH];
    if (sa_english != sb_english) {
        return sb_english - sa_english;
    }
    
    return strcmp(sa->id, sb->id);
}

int student_compare_id(const void* a, const void* b)
{
    const Student* sa = (const Student*)a;
    const Student* sb = (const Student*)b;
    return strcmp(sa->id, sb->id);
}

const char* subject_name(SubjectType subj)
{
    if (subj >= 0 && subj < SUBJECT_COUNT) {
        return subject_names[subj];
    }
    return "未知";
}

const char* track_name(TrackType track)
{
    if (track == TRACK_ARTS || track == TRACK_SCIENCE) {
        return track_names[track];
    }
    return "未知";
}
