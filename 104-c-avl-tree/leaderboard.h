#ifndef LEADERBOARD_H
#define LEADERBOARD_H

#include "concurrent_rb.h"
#include <stdint.h>

typedef struct PlayerScore {
    uint64_t player_id;
    int32_t score;
    uint64_t timestamp;
    int rank;
} PlayerScore;

typedef struct Leaderboard {
    ConcurrentLeaderboard *clb;
} Leaderboard;

Leaderboard* lb_create(int max_capacity);
void lb_destroy(Leaderboard *lb);

int lb_submit_score(Leaderboard *lb, uint64_t player_id, int32_t score);
int lb_get_player_rank(Leaderboard *lb, uint64_t player_id, int32_t score, uint64_t timestamp);
int lb_get_player_by_rank(Leaderboard *lb, int rank, PlayerScore *result);

int lb_get_top_n(Leaderboard *lb, int n, PlayerScore *results, int max_results);
int lb_get_rank_range(Leaderboard *lb, int start_rank, int end_rank, 
                      PlayerScore *results, int max_results);

int lb_count_players_in_score_range(Leaderboard *lb, int32_t min_score, int32_t max_score);

int lb_get_total_players(Leaderboard *lb);
void lb_clear(Leaderboard *lb);

#endif
