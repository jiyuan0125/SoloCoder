#ifndef STUDENT_H
#define STUDENT_H

#define MAX_STUDENTS 100000
#define SUBJECT_COUNT 5
#define CLASS_COUNT 10
#define STUDENT_ID_LEN 10

typedef enum {
    RANK_GRADE = 0,
    RANK_CLASS,
    RANK_TRACK
} RankScope;

typedef enum {
    SUBJECT_CHINESE = 0,
    SUBJECT_MATH,
    SUBJECT_ENGLISH,
    SUBJECT_PHYSICS,
    SUBJECT_CHEMISTRY
} SubjectType;

typedef enum {
    TRACK_ARTS = 0,
    TRACK_SCIENCE
} TrackType;

typedef struct {
    char id[STUDENT_ID_LEN + 1];
    int class_num;
    TrackType track;
    int scores[SUBJECT_COUNT];
    int missing[SUBJECT_COUNT];
    int total_score;
} Student;

typedef struct {
    Student* students;
    int count;
    int capacity;
} StudentDatabase;

StudentDatabase* student_db_create(int initial_capacity);
void student_db_destroy(StudentDatabase* db);
int student_db_add(StudentDatabase* db, const char* id, int class_num,
                   TrackType track, const int scores[], const int missing[]);
void student_calculate_total(Student* s);
int student_compare_rank(const void* a, const void* b);
int student_compare_id(const void* a, const void* b);
const char* subject_name(SubjectType subj);
const char* track_name(TrackType track);

#endif
