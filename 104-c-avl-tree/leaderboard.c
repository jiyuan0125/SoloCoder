#include "leaderboard.h"
#include <string.h>

Leaderboard* lb_create(int max_capacity) {
    Leaderboard *lb = (Leaderboard*)malloc(sizeof(Leaderboard));
    if (lb == NULL) return NULL;
    
    lb->clb = clb_create(max_capacity);
    if (lb->clb == NULL) {
        free(lb);
        return NULL;
    }
    
    return lb;
}

void lb_destroy(Leaderboard *lb) {
    if (lb == NULL) return;
    clb_destroy(lb->clb);
    free(lb);
}

int lb_submit_score(Leaderboard *lb, uint64_t player_id, int32_t score) {
    if (lb == NULL) return 0;
    return clb_submit_score(lb->clb, player_id, score);
}

int lb_submit_score_and_get_rank(Leaderboard *lb, uint64_t player_id, int32_t score, int *out_rank) {
    if (lb == NULL || out_rank == NULL) {
        if (out_rank != NULL) *out_rank = -1;
        return 0;
    }
    return clb_submit_score_and_get_rank(lb->clb, player_id, score, out_rank);
}

int lb_get_player_rank(Leaderboard *lb, uint64_t player_id, int32_t score, uint64_t timestamp) {
    if (lb == NULL) return -1;
    return clb_get_rank(lb->clb, player_id, score, timestamp);
}

int lb_get_rank_by_player_id(Leaderboard *lb, uint64_t player_id) {
    if (lb == NULL) return -1;
    return clb_get_rank_by_player_id(lb->clb, player_id);
}

int lb_get_player_info(Leaderboard *lb, uint64_t player_id, int32_t *out_score, uint64_t *out_timestamp) {
    if (lb == NULL) return 0;
    return clb_get_player_info(lb->clb, player_id, out_score, out_timestamp);
}

int lb_get_player_by_rank(Leaderboard *lb, int rank, PlayerScore *result) {
    if (lb == NULL || result == NULL) return 0;
    
    uint64_t player_id;
    int32_t score;
    uint64_t timestamp;
    
    int ret = clb_get_by_rank(lb->clb, rank, &player_id, &score, &timestamp);
    if (ret == 0) return 0;
    
    result->player_id = player_id;
    result->score = score;
    result->timestamp = timestamp;
    result->rank = rank;
    
    return 1;
}

int lb_get_top_n(Leaderboard *lb, int n, PlayerScore *results, int max_results) {
    if (lb == NULL || results == NULL || n < 1 || max_results < 1) return 0;
    
    int actual_n = (n < max_results) ? n : max_results;
    return lb_get_rank_range(lb, 1, actual_n, results, max_results);
}

int lb_get_rank_range(Leaderboard *lb, int start_rank, int end_rank, 
                      PlayerScore *results, int max_results) {
    if (lb == NULL || results == NULL || start_rank < 1 || end_rank < start_rank || max_results < 1) {
        return 0;
    }
    
    int count = end_rank - start_rank + 1;
    if (count > max_results) count = max_results;
    
    uint64_t *player_ids = (uint64_t*)malloc(count * sizeof(uint64_t));
    int32_t *scores = (int32_t*)malloc(count * sizeof(int32_t));
    uint64_t *timestamps = (uint64_t*)malloc(count * sizeof(uint64_t));
    
    if (player_ids == NULL || scores == NULL || timestamps == NULL) {
        free(player_ids);
        free(scores);
        free(timestamps);
        return 0;
    }
    
    int filled = clb_get_range(lb->clb, start_rank, end_rank, player_ids, scores, timestamps, count);
    
    for (int i = 0; i < filled; i++) {
        results[i].player_id = player_ids[i];
        results[i].score = scores[i];
        results[i].timestamp = timestamps[i];
        results[i].rank = start_rank + i;
    }
    
    free(player_ids);
    free(scores);
    free(timestamps);
    
    return filled;
}

int lb_count_players_in_score_range(Leaderboard *lb, int32_t min_score, int32_t max_score) {
    if (lb == NULL) return 0;
    return clb_count_in_score_range(lb->clb, min_score, max_score);
}

int lb_get_total_players(Leaderboard *lb) {
    if (lb == NULL) return 0;
    return clb_size(lb->clb);
}

void lb_clear(Leaderboard *lb) {
    if (lb == NULL) return;
    clb_clear(lb->clb);
}
