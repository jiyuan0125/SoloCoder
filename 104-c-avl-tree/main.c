#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>
#include <unistd.h>
#include <time.h>
#include <sys/time.h>

#include "leaderboard.h"

#define TEST_PLAYERS 10000
#define MAX_CAPACITY 100000
#define CONCURRENT_THREADS 10
#define SUBMISSIONS_PER_THREAD 1000

typedef struct {
    Leaderboard *lb;
    int thread_id;
    int start_player;
    int count;
} ThreadArg;

static int g_scores[TEST_PLAYERS];

static uint64_t get_timestamp_us(void) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (uint64_t)tv.tv_sec * 1000000ULL + (uint64_t)tv.tv_usec;
}

static void* submit_scores_thread(void *arg) {
    ThreadArg *targ = (ThreadArg*)arg;
    Leaderboard *lb = targ->lb;
    
    for (int i = 0; i < SUBMISSIONS_PER_THREAD; i++) {
        uint64_t player_id = (uint64_t)(targ->start_player + (i % targ->count));
        int32_t score = (int32_t)(rand() % 10000);
        lb_submit_score(lb, player_id, score);
    }
    
    return NULL;
}

static void* query_thread(void *arg) {
    ThreadArg *targ = (ThreadArg*)arg;
    Leaderboard *lb = targ->lb;
    
    for (int i = 0; i < 100; i++) {
        int rank = (rand() % 100) + 1;
        PlayerScore ps;
        lb_get_player_by_rank(lb, rank, &ps);
        
        int total = lb_get_total_players(lb);
        if (total > 0) {
            int r = rand() % total + 1;
            int start = (r > 10) ? r - 10 : 1;
            int end = r;
            
            PlayerScore results[10];
            lb_get_rank_range(lb, start, end, results, 10);
        }
        
        int min_score = rand() % 5000;
        int max_score = min_score + 1000;
        lb_count_players_in_score_range(lb, min_score, max_score);
    }
    
    return NULL;
}

static void demo_basic_operations(void) {
    printf("========================================\n");
    printf("Demo 1: Basic Operations\n");
    printf("========================================\n\n");
    
    Leaderboard *lb = lb_create(100);
    
    printf("Inserting 10 players with different scores...\n");
    
    struct {
        uint64_t id;
        int32_t score;
    } players[] = {
        {1, 1000}, {2, 2500}, {3, 1500}, {4, 3000}, {5, 2000},
        {6, 3500}, {7, 1800}, {8, 2800}, {9, 1200}, {10, 2200}
    };
    
    for (int i = 0; i < 10; i++) {
        lb_submit_score(lb, players[i].id, players[i].score);
        g_scores[players[i].id] = players[i].score;
    }
    
    printf("Total players: %d\n\n", lb_get_total_players(lb));
    
    printf("Top 5 players:\n");
    PlayerScore top5[5];
    int count = lb_get_top_n(lb, 5, top5, 5);
    for (int i = 0; i < count; i++) {
        printf("  Rank %d: Player %lu, Score %d\n", 
               top5[i].rank, (unsigned long)top5[i].player_id, top5[i].score);
    }
    printf("\n");
    
    printf("Players ranked 3-7:\n");
    PlayerScore range[10];
    count = lb_get_rank_range(lb, 3, 7, range, 10);
    for (int i = 0; i < count; i++) {
        printf("  Rank %d: Player %lu, Score %d\n", 
               range[i].rank, (unsigned long)range[i].player_id, range[i].score);
    }
    printf("\n");
    
    int min_s = 2000, max_s = 3000;
    int c = lb_count_players_in_score_range(lb, min_s, max_s);
    printf("Players with score between %d and %d: %d\n\n", min_s, max_s, c);
    
    printf("Testing same score (earlier submission ranks higher):\n");
    lb_submit_score(lb, 100, 2500);
    lb_submit_score(lb, 101, 2500);
    lb_submit_score(lb, 102, 2500);
    
    printf("Inserted 3 more players with score 2500\n");
    
    PlayerScore ps;
    for (int i = 1; i <= 13; i++) {
        if (lb_get_player_by_rank(lb, i, &ps)) {
            if (ps.score == 2500) {
                printf("  Rank %d: Player %lu, Score %d\n", 
                       ps.rank, (unsigned long)ps.player_id, ps.score);
            }
        }
    }
    printf("\n");
    
    lb_destroy(lb);
}

static void demo_capacity_limit(void) {
    printf("========================================\n");
    printf("Demo 2: Capacity Limit (Auto-eviction)\n");
    printf("========================================\n\n");
    
    int max_cap = 100;
    Leaderboard *lb = lb_create(max_cap);
    
    printf("Creating leaderboard with max capacity: %d\n", max_cap);
    
    for (int i = 0; i < 200; i++) {
        uint64_t player_id = (uint64_t)(i + 1);
        int32_t score = (int32_t)(i * 10);
        lb_submit_score(lb, player_id, score);
    }
    
    printf("Inserted 200 players (scores 0 to 1990)\n");
    printf("Current size: %d (should be %d)\n", lb_get_total_players(lb), max_cap);
    
    printf("\nBottom 5 players (should be highest ranks, lowest scores in top 100):\n");
    PlayerScore bottom[5];
    int count = lb_get_rank_range(lb, 96, 100, bottom, 5);
    for (int i = 0; i < count; i++) {
        printf("  Rank %d: Player %lu, Score %d\n", 
               bottom[i].rank, (unsigned long)bottom[i].player_id, bottom[i].score);
    }
    
    printf("\nTop 5 players:\n");
    PlayerScore top[5];
    count = lb_get_top_n(lb, 5, top, 5);
    for (int i = 0; i < count; i++) {
        printf("  Rank %d: Player %lu, Score %d\n", 
               top[i].rank, (unsigned long)top[i].player_id, top[i].score);
    }
    printf("\n");
    
    lb_destroy(lb);
}

static void demo_concurrency(void) {
    printf("========================================\n");
    printf("Demo 3: Concurrent Operations\n");
    printf("========================================\n\n");
    
    Leaderboard *lb = lb_create(MAX_CAPACITY);
    
    int num_writer_threads = CONCURRENT_THREADS;
    int num_reader_threads = 5;
    pthread_t writers[CONCURRENT_THREADS];
    pthread_t readers[5];
    ThreadArg writer_args[CONCURRENT_THREADS];
    ThreadArg reader_args[5];
    
    printf("Starting %d writer threads, each submitting %d scores...\n", 
           num_writer_threads, SUBMISSIONS_PER_THREAD);
    printf("Starting %d concurrent reader threads...\n", num_reader_threads);
    
    uint64_t start_time = get_timestamp_us();
    
    for (int i = 0; i < num_writer_threads; i++) {
        writer_args[i].lb = lb;
        writer_args[i].thread_id = i;
        writer_args[i].start_player = i * 1000 + 1;
        writer_args[i].count = 1000;
        pthread_create(&writers[i], NULL, submit_scores_thread, &writer_args[i]);
    }
    
    for (int i = 0; i < num_reader_threads; i++) {
        reader_args[i].lb = lb;
        reader_args[i].thread_id = i;
        reader_args[i].start_player = 0;
        reader_args[i].count = 0;
        pthread_create(&readers[i], NULL, query_thread, &reader_args[i]);
    }
    
    for (int i = 0; i < num_writer_threads; i++) {
        pthread_join(writers[i], NULL);
    }
    
    for (int i = 0; i < num_reader_threads; i++) {
        pthread_join(readers[i], NULL);
    }
    
    uint64_t end_time = get_timestamp_us();
    double elapsed = (double)(end_time - start_time) / 1000000.0;
    
    printf("\nConcurrent test completed!\n");
    printf("Total submissions: %d\n", num_writer_threads * SUBMISSIONS_PER_THREAD);
    printf("Total players in leaderboard: %d\n", lb_get_total_players(lb));
    printf("Elapsed time: %.3f seconds\n", elapsed);
    printf("Throughput: %.2f ops/sec\n", 
           (double)(num_writer_threads * SUBMISSIONS_PER_THREAD) / elapsed);
    
    printf("\nTop 10 players after concurrent submissions:\n");
    PlayerScore top10[10];
    int count = lb_get_top_n(lb, 10, top10, 10);
    for (int i = 0; i < count; i++) {
        printf("  Rank %d: Player %lu, Score %d\n", 
               top10[i].rank, (unsigned long)top10[i].player_id, top10[i].score);
    }
    printf("\n");
    
    lb_destroy(lb);
}

static void demo_performance(void) {
    printf("========================================\n");
    printf("Demo 4: Performance Benchmark\n");
    printf("========================================\n\n");
    
    Leaderboard *lb = lb_create(1000000);
    
    int num_players = 100000;
    uint64_t *player_ids = (uint64_t*)malloc(num_players * sizeof(uint64_t));
    int32_t *scores = (int32_t*)malloc(num_players * sizeof(int32_t));
    uint64_t *timestamps = (uint64_t*)malloc(num_players * sizeof(uint64_t));
    
    printf("Generating %d random player scores...\n", num_players);
    srand((unsigned int)time(NULL));
    
    for (int i = 0; i < num_players; i++) {
        player_ids[i] = (uint64_t)(i + 1);
        scores[i] = (int32_t)(rand() % 1000000);
    }
    
    printf("Inserting %d players...\n", num_players);
    uint64_t start_time = get_timestamp_us();
    
    for (int i = 0; i < num_players; i++) {
        lb_submit_score(lb, player_ids[i], scores[i]);
    }
    
    uint64_t end_time = get_timestamp_us();
    double insert_time = (double)(end_time - start_time) / 1000000.0;
    
    printf("Insertion time: %.3f seconds\n", insert_time);
    printf("Insertion rate: %.2f players/sec\n", (double)num_players / insert_time);
    printf("Total players: %d\n\n", lb_get_total_players(lb));
    
    printf("Benchmarking rank queries...\n");
    int query_count = 10000;
    start_time = get_timestamp_us();
    
    PlayerScore ps;
    for (int i = 0; i < query_count; i++) {
        int rank = (rand() % num_players) + 1;
        lb_get_player_by_rank(lb, rank, &ps);
    }
    
    end_time = get_timestamp_us();
    double query_time = (double)(end_time - start_time) / 1000000.0;
    
    printf("Rank query time (%d queries): %.3f seconds\n", query_count, query_time);
    printf("Rank query rate: %.2f queries/sec\n", (double)query_count / query_time);
    printf("Average query latency: %.2f microseconds\n\n", 
           (double)(end_time - start_time) / (double)query_count);
    
    printf("Benchmarking range queries (10 players each)...\n");
    query_count = 5000;
    start_time = get_timestamp_us();
    
    PlayerScore results[10];
    for (int i = 0; i < query_count; i++) {
        int start = (rand() % (num_players - 10)) + 1;
        lb_get_rank_range(lb, start, start + 9, results, 10);
    }
    
    end_time = get_timestamp_us();
    double range_time = (double)(end_time - start_time) / 1000000.0;
    
    printf("Range query time (%d queries): %.3f seconds\n", query_count, range_time);
    printf("Range query rate: %.2f queries/sec\n", (double)query_count / range_time);
    printf("\n");
    
    printf("Benchmarking score range count queries...\n");
    query_count = 5000;
    start_time = get_timestamp_us();
    
    for (int i = 0; i < query_count; i++) {
        int min_s = rand() % 500000;
        int max_s = min_s + 100000;
        lb_count_players_in_score_range(lb, min_s, max_s);
    }
    
    end_time = get_timestamp_us();
    double count_time = (double)(end_time - start_time) / 1000000.0;
    
    printf("Score count query time (%d queries): %.3f seconds\n", query_count, count_time);
    printf("Score count query rate: %.2f queries/sec\n", (double)query_count / count_time);
    printf("\n");
    
    free(player_ids);
    free(scores);
    free(timestamps);
    lb_destroy(lb);
}

int main(void) {
    srand((unsigned int)time(NULL));
    
    printf("\n");
    printf("========================================\n");
    printf("   Game Real-time Leaderboard System\n");
    printf("========================================\n\n");
    
    demo_basic_operations();
    demo_capacity_limit();
    demo_concurrency();
    demo_performance();
    
    printf("========================================\n");
    printf("   All demos completed successfully!\n");
    printf("========================================\n\n");
    
    return 0;
}
