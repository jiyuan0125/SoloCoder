#include "shuffle.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

static void shuffle_array(int* arr, int n) {
    if (n <= 1) return;
    for (int i = n - 1; i > 0; i--) {
        int j = rand() % (i + 1);
        int temp = arr[i];
        arr[i] = arr[j];
        arr[j] = temp;
    }
}

static void ensure_shuffled_excluding_current(ShuffleManager* sm) {
    if (!sm || sm->total_count <= 0) return;
    
    if (sm->total_count == 1) {
        sm->indices[0] = 0;
        sm->current_pos = -1;
        return;
    }
    
    int current_idx = sm->current_playing_idx;
    int attempts = 0;
    do {
        shuffle_array(sm->indices, sm->total_count);
        attempts++;
    } while (sm->indices[0] == current_idx && attempts < 100);
    
    if (sm->indices[0] == current_idx) {
        for (int i = 1; i < sm->total_count; i++) {
            if (sm->indices[i] != current_idx) {
                int temp = sm->indices[0];
                sm->indices[0] = sm->indices[i];
                sm->indices[i] = temp;
                break;
            }
        }
    }
    
    sm->current_pos = -1;
}

ShuffleManager* shuffle_manager_create(void) {
    ShuffleManager* sm = (ShuffleManager*)malloc(sizeof(ShuffleManager));
    if (!sm) return NULL;
    sm->indices = NULL;
    sm->total_count = 0;
    sm->current_pos = 0;
    sm->current_playing_idx = -1;
    srand((unsigned int)time(NULL));
    return sm;
}

void shuffle_manager_destroy(ShuffleManager* sm) {
    if (!sm) return;
    if (sm->indices) {
        free(sm->indices);
    }
    free(sm);
}

void shuffle_manager_reset(ShuffleManager* sm, int total_songs, int current_playing_index) {
    if (!sm) return;
    
    if (sm->indices) {
        free(sm->indices);
        sm->indices = NULL;
    }
    
    sm->total_count = total_songs;
    sm->current_playing_idx = current_playing_index;
    sm->current_pos = 0;
    
    if (total_songs <= 0) return;
    
    sm->indices = (int*)malloc(sizeof(int) * total_songs);
    if (!sm->indices) {
        sm->total_count = 0;
        return;
    }
    
    for (int i = 0; i < total_songs; i++) {
        sm->indices[i] = i;
    }
    
    ensure_shuffled_excluding_current(sm);
}

int shuffle_manager_get_next(ShuffleManager* sm) {
    if (!sm || sm->total_count <= 0) return -1;
    
    if (sm->total_count == 1) {
        return 0;
    }
    
    if (sm->current_pos >= sm->total_count - 1) {
        ensure_shuffled_excluding_current(sm);
    }
    
    sm->current_pos++;
    if (sm->current_pos >= sm->total_count) {
        sm->current_pos = 0;
    }
    
    int result = sm->indices[sm->current_pos];
    sm->current_playing_idx = result;
    return result;
}

int shuffle_manager_get_prev(ShuffleManager* sm) {
    if (!sm || sm->total_count <= 0) return -1;
    
    if (sm->total_count == 1) {
        return 0;
    }
    
    if (sm->current_pos > 0) {
        sm->current_pos--;
    } else {
        sm->current_pos = sm->total_count - 1;
    }
    
    int result = sm->indices[sm->current_pos];
    sm->current_playing_idx = result;
    return result;
}
