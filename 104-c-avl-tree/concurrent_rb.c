#include "concurrent_rb.h"
#include <time.h>
#include <string.h>

static uint64_t get_timestamp(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

ConcurrentLeaderboard* clb_create(int max_capacity) {
    ConcurrentLeaderboard *clb = (ConcurrentLeaderboard*)malloc(sizeof(ConcurrentLeaderboard));
    if (clb == NULL) return NULL;
    
    clb->tree = avl_create(max_capacity);
    if (clb->tree == NULL) {
        free(clb);
        return NULL;
    }
    
    if (pthread_rwlock_init(&clb->rwlock, NULL) != 0) {
        avl_destroy(clb->tree);
        free(clb);
        return NULL;
    }
    
    return clb;
}

void clb_destroy(ConcurrentLeaderboard *clb) {
    if (clb == NULL) return;
    pthread_rwlock_destroy(&clb->rwlock);
    avl_destroy(clb->tree);
    free(clb);
}

int clb_submit_score(ConcurrentLeaderboard *clb, uint64_t player_id, int32_t score) {
    if (clb == NULL) return 0;
    
    uint64_t timestamp = get_timestamp();
    
    pthread_rwlock_wrlock(&clb->rwlock);
    int result = avl_insert(clb->tree, player_id, score, timestamp);
    pthread_rwlock_unlock(&clb->rwlock);
    
    return result;
}

int clb_get_rank(ConcurrentLeaderboard *clb, uint64_t player_id, int32_t score, uint64_t timestamp) {
    if (clb == NULL) return -1;
    
    pthread_rwlock_rdlock(&clb->rwlock);
    int rank = avl_get_rank(clb->tree, player_id, score, timestamp);
    pthread_rwlock_unlock(&clb->rwlock);
    
    return rank;
}

int clb_get_by_rank(ConcurrentLeaderboard *clb, int rank, uint64_t *player_id, int32_t *score, uint64_t *timestamp) {
    if (clb == NULL) return 0;
    
    pthread_rwlock_rdlock(&clb->rwlock);
    int result = avl_get_by_rank(clb->tree, rank, player_id, score, timestamp);
    pthread_rwlock_unlock(&clb->rwlock);
    
    return result;
}

int clb_get_range(ConcurrentLeaderboard *clb, int start_rank, int end_rank,
                  uint64_t *player_ids, int32_t *scores, uint64_t *timestamps, int max_count) {
    if (clb == NULL) return 0;
    
    pthread_rwlock_rdlock(&clb->rwlock);
    int count = avl_get_range(clb->tree, start_rank, end_rank, player_ids, scores, timestamps, max_count);
    pthread_rwlock_unlock(&clb->rwlock);
    
    return count;
}

int clb_count_in_score_range(ConcurrentLeaderboard *clb, int32_t min_score, int32_t max_score) {
    if (clb == NULL) return 0;
    
    pthread_rwlock_rdlock(&clb->rwlock);
    int count = avl_count_in_score_range(clb->tree, min_score, max_score);
    pthread_rwlock_unlock(&clb->rwlock);
    
    return count;
}

int clb_size(ConcurrentLeaderboard *clb) {
    if (clb == NULL) return 0;
    
    pthread_rwlock_rdlock(&clb->rwlock);
    int size = avl_size(clb->tree);
    pthread_rwlock_unlock(&clb->rwlock);
    
    return size;
}

void clb_clear(ConcurrentLeaderboard *clb) {
    if (clb == NULL) return;
    
    pthread_rwlock_wrlock(&clb->rwlock);
    avl_clear(clb->tree);
    pthread_rwlock_unlock(&clb->rwlock);
}
