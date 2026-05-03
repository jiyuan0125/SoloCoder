#include "player_index.h"
#include <string.h>

static uint64_t hash_func(uint64_t player_id, int bucket_count) {
    return player_id % (uint64_t)bucket_count;
}

PlayerIndex* pi_create(int bucket_count) {
    if (bucket_count <= 0) bucket_count = 10007;
    
    PlayerIndex *index = (PlayerIndex*)malloc(sizeof(PlayerIndex));
    if (index == NULL) return NULL;
    
    index->buckets = (PlayerEntry**)calloc((size_t)bucket_count, sizeof(PlayerEntry*));
    if (index->buckets == NULL) {
        free(index);
        return NULL;
    }
    
    index->bucket_count = bucket_count;
    index->size = 0;
    
    return index;
}

void pi_destroy(PlayerIndex *index) {
    if (index == NULL) return;
    
    pi_clear(index);
    free(index->buckets);
    free(index);
}

int pi_insert(PlayerIndex *index, uint64_t player_id, int32_t score, uint64_t timestamp) {
    if (index == NULL) return 0;
    
    uint64_t bucket = hash_func(player_id, index->bucket_count);
    PlayerEntry *entry;
    
    for (entry = index->buckets[bucket]; entry != NULL; entry = entry->next) {
        if (entry->player_id == player_id) {
            entry->score = score;
            entry->timestamp = timestamp;
            return 1;
        }
    }
    
    entry = (PlayerEntry*)malloc(sizeof(PlayerEntry));
    if (entry == NULL) return 0;
    
    entry->player_id = player_id;
    entry->score = score;
    entry->timestamp = timestamp;
    entry->next = index->buckets[bucket];
    index->buckets[bucket] = entry;
    index->size++;
    
    return 1;
}

int pi_remove(PlayerIndex *index, uint64_t player_id) {
    if (index == NULL) return 0;
    
    uint64_t bucket = hash_func(player_id, index->bucket_count);
    PlayerEntry *entry = index->buckets[bucket];
    PlayerEntry *prev = NULL;
    
    for (; entry != NULL; prev = entry, entry = entry->next) {
        if (entry->player_id == player_id) {
            if (prev == NULL) {
                index->buckets[bucket] = entry->next;
            } else {
                prev->next = entry->next;
            }
            free(entry);
            index->size--;
            return 1;
        }
    }
    
    return 0;
}

int pi_get(PlayerIndex *index, uint64_t player_id, int32_t *score, uint64_t *timestamp) {
    if (index == NULL) return 0;
    
    uint64_t bucket = hash_func(player_id, index->bucket_count);
    PlayerEntry *entry;
    
    for (entry = index->buckets[bucket]; entry != NULL; entry = entry->next) {
        if (entry->player_id == player_id) {
            if (score != NULL) *score = entry->score;
            if (timestamp != NULL) *timestamp = entry->timestamp;
            return 1;
        }
    }
    
    return 0;
}

int pi_exists(PlayerIndex *index, uint64_t player_id) {
    if (index == NULL) return 0;
    return pi_get(index, player_id, NULL, NULL);
}

int pi_size(PlayerIndex *index) {
    if (index == NULL) return 0;
    return index->size;
}

void pi_clear(PlayerIndex *index) {
    if (index == NULL) return;
    
    for (int i = 0; i < index->bucket_count; i++) {
        PlayerEntry *entry = index->buckets[i];
        while (entry != NULL) {
            PlayerEntry *next = entry->next;
            free(entry);
            entry = next;
        }
        index->buckets[i] = NULL;
    }
    index->size = 0;
}
