#ifndef SHUFFLE_H
#define SHUFFLE_H

#include "song_list.h"

typedef struct ShuffleManager {
    int* indices;
    int total_count;
    int current_pos;
    int current_playing_idx;
} ShuffleManager;

ShuffleManager* shuffle_manager_create(void);
void shuffle_manager_destroy(ShuffleManager* sm);
void shuffle_manager_reset(ShuffleManager* sm, int total_songs, int current_playing_index);
int shuffle_manager_get_next(ShuffleManager* sm);
int shuffle_manager_get_prev(ShuffleManager* sm);

#endif
