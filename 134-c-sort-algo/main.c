#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include "student.h"
#include "ranking.h"
#include "statistics.h"

#define DEFAULT_STUDENT_COUNT 1000

static void generate_student_id(int index, char* out_id, int max_len)
{
    snprintf(out_id, max_len, "S%06d", index + 1);
}

static int random_score(void)
{
    return rand() % 101;
}

static int random_missing(int student_idx)
{
    if ((student_idx % 97 == 0) || (student_idx % 113 == 0)) {
        return 1;
    }
    return 0;
}

static void generate_test_data(StudentDatabase* db, int count)
{
    for (int i = 0; i < count; i++) {
        char id[STUDENT_ID_LEN + 1];
        generate_student_id(i, id, sizeof(id));
        
        int class_num = (i % CLASS_COUNT) + 1;
        TrackType track = (i % 3 == 0) ? TRACK_ARTS : TRACK_SCIENCE;
        
        int scores[SUBJECT_COUNT];
        int missing[SUBJECT_COUNT] = {0};
        
        for (int s = 0; s < SUBJECT_COUNT; s++) {
            scores[s] = random_score();
            missing[s] = random_missing(i + s);
        }
        
        if (i == 0 || i == 1) {
            scores[SUBJECT_CHINESE] = 95;
            scores[SUBJECT_MATH] = 100;
            scores[SUBJECT_ENGLISH] = 90;
            scores[SUBJECT_PHYSICS] = 85;
            scores[SUBJECT_CHEMISTRY] = 80;
            memset(missing, 0, sizeof(missing));
        }
        
        if (i == 2) {
            scores[SUBJECT_CHINESE] = 90;
            scores[SUBJECT_MATH] = 95;
            scores[SUBJECT_ENGLISH] = 90;
            scores[SUBJECT_PHYSICS] = 85;
            scores[SUBJECT_CHEMISTRY] = 80;
            memset(missing, 0, sizeof(missing));
        }
        
        student_db_add(db, id, class_num, track, scores, missing);
    }
}

static void print_student(const Student* s, int total_count)
{
    printf("学号: %s | 班级: %d班 | %s\n", 
           s->id, s->class_num, track_name(s->track));
    printf("  成绩: ");
    for (int i = 0; i < SUBJECT_COUNT; i++) {
        if (s->missing[i]) {
            printf("%s:缺考 ", subject_name((SubjectType)i));
        } else {
            printf("%s:%d ", subject_name((SubjectType)i), s->scores[i]);
        }
    }
    printf("| 总分: %d\n", s->total_score);
}

static void print_rank_summary(RankResult* result, const char* title)
{
    if (!result || result->count == 0) return;
    
    printf("\n========== %s (共%d人) ==========\n", title, result->count);
    printf("前10名:\n");
    printf("%-6s | %-10s | %-6s | %-6s\n", 
           "名次", "学号", "百分位", "说明");
    printf("----------------------------------------------------\n");
    
    for (int i = 0; i < 10 && i < result->count; i++) {
        const char* tie_note = "";
        if (i > 0 && result->entries[i].rank == result->entries[i-1].rank) {
            tie_note = "并列";
        }
        
        printf("%-6d | %-10s | %-6d%% | %s\n",
               result->entries[i].rank,
               result->entries[i].student_id,
               result->entries[i].percentile,
               tie_note);
    }
    
    printf("...\n");
}

static void print_stats_summary(StatsResult* result, const char* title)
{
    if (!result) return;
    
    printf("\n========== %s (共%d人) ==========\n", title, result->total_students);
    printf("%-6s | %-8s | %-8s | %-8s | %-8s | %-8s | %-6s | %-6s\n",
           "科目", "平均分", "最高分", "最低分", "中位数", "标准差", "参考", "缺考");
    printf("--------------------------------------------------------------------------------\n");
    
    for (int s = 0; s < SUBJECT_COUNT; s++) {
        SubjectStats* stats = &result->subjects[s];
        printf("%-6s | %-8.1f | %-8d | %-8d | %-8.1f | %-8.1f | %-6d | %-6d\n",
               subject_name((SubjectType)s),
               stats->average,
               stats->max_score,
               stats->min_score,
               stats->median,
               stats->stddev,
               stats->taken_count,
               stats->missing_count);
    }
}

static void print_score_distribution(StatsResult* result, SubjectType subj)
{
    if (!result) return;
    
    SubjectStats* stats = &result->subjects[subj];
    printf("\n%s 分数段分布 (参考:%d人, 缺考:%d人)\n",
           subject_name(subj), stats->taken_count, stats->missing_count);
    printf("----------------------------------------\n");
    
    for (int i = 0; i < SCORE_RANGE_COUNT; i++) {
        if (stats->distribution[i] > 0) {
            int count = stats->distribution[i];
            double pct = stats->taken_count > 0 ? 
                (double)count / stats->taken_count * 100 : 0;
            
            int bar_len = (int)(pct / 2);
            char bar[51];
            memset(bar, '=', bar_len);
            bar[bar_len] = '\0';
            
            printf("%-6s: %4d人 (%5.1f%%) %s\n",
                   score_range_name(i), count, pct, bar);
        }
    }
}

static int find_student_by_id(StudentDatabase* db, const char* id)
{
    for (int i = 0; i < db->count; i++) {
        if (strcmp(db->students[i].id, id) == 0) {
            return i;
        }
    }
    return -1;
}

static int find_rank_by_id(RankResult* result, const char* id)
{
    for (int i = 0; i < result->count; i++) {
        if (strcmp(result->entries[i].student_id, id) == 0) {
            return i;
        }
    }
    return -1;
}

int main(int argc, char* argv[])
{
    int student_count = DEFAULT_STUDENT_COUNT;
    
    if (argc > 1) {
        student_count = atoi(argv[1]);
        if (student_count < 10 || student_count > MAX_STUDENTS) {
            student_count = DEFAULT_STUDENT_COUNT;
        }
    }
    
    srand((unsigned int)time(NULL));
    
    printf("========================================\n");
    printf("    考试成绩排名系统 v1.0\n");
    printf("========================================\n\n");
    
    printf("正在生成 %d 名学生的成绩数据...\n", student_count);
    printf("(其中 S000001 和 S000002 为并列第1名测试数据)\n\n");
    
    StudentDatabase* db = student_db_create(student_count);
    if (!db) {
        fprintf(stderr, "创建数据库失败\n");
        return 1;
    }
    
    generate_test_data(db, student_count);
    
    printf("学生数据样例 (前5名):\n");
    for (int i = 0; i < 5 && i < db->count; i++) {
        print_student(&db->students[i], db->count);
        printf("\n");
    }
    
    clock_t start, end;
    
    printf("\n---------- 执行排名计算 ----------\n");
    
    start = clock();
    RankResult* grade_rank = rank_create(db->count);
    rank_students(db, RANK_GRADE, 0, grade_rank);
    end = clock();
    printf("年级排名完成，耗时: %.2f ms\n", 
           (double)(end - start) * 1000.0 / CLOCKS_PER_SEC);
    
    RankResult* class1_rank = rank_create(db->count);
    rank_students(db, RANK_CLASS, 1, class1_rank);
    
    RankResult* arts_rank = rank_create(db->count);
    rank_students(db, RANK_TRACK, TRACK_ARTS, arts_rank);
    
    RankResult* science_rank = rank_create(db->count);
    rank_students(db, RANK_TRACK, TRACK_SCIENCE, science_rank);
    
    print_rank_summary(grade_rank, "年级总排名");
    print_rank_summary(class1_rank, "1班排名");
    print_rank_summary(arts_rank, "文科排名");
    print_rank_summary(science_rank, "理科排名");
    
    printf("\n---------- 并列排名验证 ----------\n");
    printf("测试并列排名逻辑: S000001 和 S000002 应为并列第1名\n");
    printf("预期: 第1名有2人，下一个为第3名\n\n");
    
    for (int i = 0; i < 4 && i < grade_rank->count; i++) {
        printf("第%d名: %s", 
               grade_rank->entries[i].rank,
               grade_rank->entries[i].student_id);
        if (i > 0 && grade_rank->entries[i].rank == grade_rank->entries[i-1].rank) {
            printf(" (与上一名并列)");
        }
        printf("\n");
    }
    
    printf("\n---------- 百分位示例 ----------\n");
    const char* sample_ids[] = {"S000001", "S000003", "S000010", "S000100", "S000500"};
    int sample_count = sizeof(sample_ids) / sizeof(sample_ids[0]);
    
    printf("%-10s | %-6s | %-6s | %-6s | 说明\n", 
           "学号", "年级名", "班级名", "百分位");
    printf("-----------------------------------------------------\n");
    
    for (int i = 0; i < sample_count; i++) {
        int grade_idx = find_rank_by_id(grade_rank, sample_ids[i]);
        int class_idx = find_rank_by_id(class1_rank, sample_ids[i]);
        
        if (grade_idx >= 0) {
            int student_idx = find_student_by_id(db, sample_ids[i]);
            int class_num = student_idx >= 0 ? db->students[student_idx].class_num : 0;
            
            printf("%-10s | 第%-4d名 | %d班第%-2d名 | %-6d%% | 超过了 %d%% 的学生\n",
                   sample_ids[i],
                   grade_rank->entries[grade_idx].rank,
                   class_num,
                   class_idx >= 0 ? class1_rank->entries[class_idx].rank : -1,
                   grade_rank->entries[grade_idx].percentile,
                   grade_rank->entries[grade_idx].percentile);
        }
    }
    
    printf("\n---------- 执行统计分析 ----------\n");
    
    StatsResult* grade_stats = stats_create();
    calculate_statistics(db, grade_stats);
    
    print_stats_summary(grade_stats, "全年级成绩统计");
    print_score_distribution(grade_stats, SUBJECT_MATH);
    print_score_distribution(grade_stats, SUBJECT_CHINESE);
    
    printf("\n---------- 缺考处理验证 ----------\n");
    printf("缺考学生成绩记为0参与排名，但不计入平均分统计\n");
    printf("各科目缺考人数: ");
    for (int s = 0; s < SUBJECT_COUNT; s++) {
        printf("%s:%d人 ", subject_name((SubjectType)s), 
               grade_stats->subjects[s].missing_count);
    }
    printf("\n");
    
    rank_destroy(grade_rank);
    rank_destroy(class1_rank);
    rank_destroy(arts_rank);
    rank_destroy(science_rank);
    stats_destroy(grade_stats);
    student_db_destroy(db);
    
    printf("\n========================================\n");
    printf("   演示完成！系统可处理 %d 学生规模\n", MAX_STUDENTS);
    printf("========================================\n");
    
    return 0;
}
