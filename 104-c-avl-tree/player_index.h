#ifndef PLAYER_INDEX_H
#define PLAYER_INDEX_H

#include <stdint.h>
#include <stdlib.h>

typedef struct PlayerEntry {
    uint64_t player_id;
    int32_t score;
    uint64_t timestamp;
    struct PlayerEntry *next;
} PlayerEntry;

typedef struct PlayerIndex {
    PlayerEntry **buckets;
    int bucket_count;
    int size;
} PlayerIndex;

PlayerIndex* pi_create(int bucket_count);
void pi_destroy(PlayerIndex *index);

int pi_insert(PlayerIndex *index, uint64_t player_id, int32_t score, uint64_t timestamp);
int pi_remove(PlayerIndex *index, uint64_t player_id);
int pi_get(PlayerIndex *index, uint64_t player_id, int32_t *score, uint64_t *timestamp);
int pi_exists(PlayerIndex *index, uint64_t player_id);
int pi_size(PlayerIndex *index);
void pi_clear(PlayerIndex *index);

#endif
