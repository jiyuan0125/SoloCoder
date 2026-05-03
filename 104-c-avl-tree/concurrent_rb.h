#ifndef CONCURRENT_RB_H
#define CONCURRENT_RB_H

#include "avl_tree.h"
#include <pthread.h>

typedef struct ConcurrentLeaderboard {
    AVLTree *tree;
    pthread_rwlock_t rwlock;
} ConcurrentLeaderboard;

ConcurrentLeaderboard* clb_create(int max_capacity);
void clb_destroy(ConcurrentLeaderboard *clb);

int clb_submit_score(ConcurrentLeaderboard *clb, uint64_t player_id, int32_t score);
int clb_get_rank(ConcurrentLeaderboard *clb, uint64_t player_id, int32_t score, uint64_t timestamp);
int clb_get_by_rank(ConcurrentLeaderboard *clb, int rank, uint64_t *player_id, int32_t *score, uint64_t *timestamp);
int clb_get_range(ConcurrentLeaderboard *clb, int start_rank, int end_rank,
                  uint64_t *player_ids, int32_t *scores, uint64_t *timestamps, int max_count);
int clb_count_in_score_range(ConcurrentLeaderboard *clb, int32_t min_score, int32_t max_score);

int clb_size(ConcurrentLeaderboard *clb);
void clb_clear(ConcurrentLeaderboard *clb);

#endif
