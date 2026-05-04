#include "player.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

const char* player_mode_to_string(PlayMode mode) {
    switch (mode) {
        case PLAY_MODE_SEQUENTIAL: return "顺序播放";
        case PLAY_MODE_REPEAT_ALL:  return "列表循环";
        case PLAY_MODE_REPEAT_ONE:  return "单曲循环";
        case PLAY_MODE_SHUFFLE:     return "随机播放";
        default: return "未知模式";
    }
}

Player* player_create(void) {
    Player* player = (Player*)malloc(sizeof(Player));
    if (!player) return NULL;
    player->main_list = song_list_create();
    player->temp_list = song_list_create();
    player->current_node = NULL;
    player->current_source = LIST_MAIN;
    player->main_continue_point = NULL;
    player->mode = PLAY_MODE_SEQUENTIAL;
    player->shuffle_mgr = shuffle_manager_create();
    player->is_playing = false;
    player->is_paused = false;
    if (!player->main_list || !player->temp_list || !player->shuffle_mgr) {
        player_destroy(player);
        return NULL;
    }
    return player;
}

void player_destroy(Player* player) {
    if (!player) return;
    if (player->main_list) song_list_destroy(player->main_list);
    if (player->temp_list) song_list_destroy(player->temp_list);
    if (player->shuffle_mgr) shuffle_manager_destroy(player->shuffle_mgr);
    free(player);
}

void player_add_song(Player* player, Song* song) {
    if (!player || !song) return;
    song->source = SOURCE_MAIN;
    song_list_append(player->main_list, song);
}

void player_insert_song_at(Player* player, int position, Song* song) {
    if (!player || !song) return;
    song->source = SOURCE_MAIN;
    song_list_insert_at(player->main_list, position, song);
}

void player_insert_temp_song(Player* player, Song* song) {
    if (!player || !song) return;
    song->source = SOURCE_TEMP;
    song_list_append(player->temp_list, song);
    
    if (player->current_source == LIST_MAIN && player->current_node && 
        player->main_continue_point == NULL) {
        player->main_continue_point = player->current_node->next;
        if (!player->main_continue_point) {
            player->main_continue_point = player->main_list->head;
        }
    }
}

static SongNode* find_in_either_list(Player* player, int song_id, WhichList* out_source) {
    if (!player) return NULL;
    SongNode* node = song_list_find_by_id(player->main_list, song_id);
    if (node) {
        if (out_source) *out_source = LIST_MAIN;
        return node;
    }
    node = song_list_find_by_id(player->temp_list, song_id);
    if (node) {
        if (out_source) *out_source = LIST_TEMP;
        return node;
    }
    return NULL;
}

void player_remove_song(Player* player, int song_id) {
    if (!player) return;
    
    WhichList source;
    SongNode* node = find_in_either_list(player, song_id, &source);
    if (!node) return;
    
    SongList* list = (source == LIST_MAIN) ? player->main_list : player->temp_list;
    bool is_current = (node == player->current_node);
    SongNode* next_node = NULL;
    SongNode* continue_point_next = NULL;
    bool is_continue_point = (node == player->main_continue_point && source == LIST_MAIN);
    
    if (is_current) {
        if (node->next) {
            next_node = node->next;
        } else {
            next_node = list->head;
            if (next_node == node) {
                next_node = NULL;
            }
        }
    }
    
    if (is_continue_point) {
        if (node->next) {
            continue_point_next = node->next;
        } else {
            continue_point_next = player->main_list->head;
        }
        if (continue_point_next == node) {
            continue_point_next = NULL;
        }
    }
    
    Song* removed_song = song_list_remove_by_node(list, node);
    
    if (is_current) {
        player->current_node = next_node;
        if (!player->current_node && source == LIST_TEMP) {
            player->current_node = player->main_continue_point;
            player->current_source = LIST_MAIN;
            player->main_continue_point = NULL;
            if (!player->current_node) {
                player->current_node = player->main_list->head;
            }
        }
    }
    
    if (is_continue_point) {
        player->main_continue_point = continue_point_next;
    }
    
    song_destroy(removed_song);
}

void player_move_song(Player* player, int from_pos, int to_pos) {
    if (!player) return;
    song_list_move(player->main_list, from_pos, to_pos);
}

void player_set_favorite(Player* player, int song_id, bool favorite) {
    if (!player) return;
    song_list_mark_favorite(player->main_list, song_id, favorite);
    song_list_mark_favorite(player->temp_list, song_id, favorite);
}

static void reset_shuffle_if_needed(Player* player) {
    if (!player || player->mode != PLAY_MODE_SHUFFLE) return;
    int total = player->main_list->count + player->temp_list->count;
    if (total == 0) {
        shuffle_manager_reset(player->shuffle_mgr, 0, -1);
        return;
    }
    int current_idx = -1;
    if (player->current_node) {
        if (player->current_source == LIST_MAIN) {
            current_idx = song_list_get_position(player->main_list, player->current_node);
        } else {
            current_idx = player->main_list->count + 
                          song_list_get_position(player->temp_list, player->current_node);
        }
    }
    shuffle_manager_reset(player->shuffle_mgr, total, current_idx);
}

void player_set_mode(Player* player, PlayMode mode) {
    if (!player) return;
    player->mode = mode;
    if (mode == PLAY_MODE_SHUFFLE) {
        reset_shuffle_if_needed(player);
    }
}

PlayMode player_get_mode(Player* player) {
    if (!player) return PLAY_MODE_SEQUENTIAL;
    return player->mode;
}

Song* player_play(Player* player) {
    if (!player) return NULL;
    if (player->current_node) {
        player->is_playing = true;
        player->is_paused = false;
        return player->current_node->song;
    }
    if (player->main_list->count > 0) {
        player->current_node = player->main_list->head;
        player->current_source = LIST_MAIN;
        player->is_playing = true;
        player->is_paused = false;
        if (player->mode == PLAY_MODE_SHUFFLE) {
            reset_shuffle_if_needed(player);
        }
        return player->current_node->song;
    }
    if (player->temp_list->count > 0) {
        player->current_node = player->temp_list->head;
        player->current_source = LIST_TEMP;
        player->is_playing = true;
        player->is_paused = false;
        return player->current_node->song;
    }
    return NULL;
}

Song* player_play_at(Player* player, int position) {
    if (!player) return NULL;
    SongNode* node = song_list_get_node_at(player->main_list, position);
    if (node) {
        player->current_node = node;
        player->current_source = LIST_MAIN;
        player->is_playing = true;
        player->is_paused = false;
        if (player->mode == PLAY_MODE_SHUFFLE) {
            reset_shuffle_if_needed(player);
        }
        return node->song;
    }
    return NULL;
}

void player_pause(Player* player) {
    if (!player) return;
    if (player->is_playing) {
        player->is_paused = true;
    }
}

void player_resume(Player* player) {
    if (!player) return;
    if (player->is_paused) {
        player->is_paused = false;
    }
}

void player_stop(Player* player) {
    if (!player) return;
    player->is_playing = false;
    player->is_paused = false;
}

static SongNode* get_node_by_combined_index(Player* player, int idx) {
    if (!player) return NULL;
    if (idx < player->main_list->count) {
        return song_list_get_node_at(player->main_list, idx);
    }
    int temp_idx = idx - player->main_list->count;
    if (temp_idx < player->temp_list->count) {
        return song_list_get_node_at(player->temp_list, temp_idx);
    }
    return NULL;
}

static WhichList get_source_by_combined_index(Player* player, int idx) {
    if (!player) return LIST_MAIN;
    if (idx < player->main_list->count) {
        return LIST_MAIN;
    }
    return LIST_TEMP;
}

static Song* handle_temp_song_finished(Player* player) {
    if (!player || !player->current_node) return NULL;
    if (player->current_source != LIST_TEMP) return NULL;
    
    SongNode* finished_node = player->current_node;
    Song* finished_song = finished_node->song;
    
    player->current_node = finished_node->next;
    
    Song* removed = song_list_remove_by_node(player->temp_list, finished_node);
    (void)removed;
    
    if (player->current_node) {
        return player->current_node->song;
    }
    
    if (player->main_continue_point) {
        player->current_node = player->main_continue_point;
        player->current_source = LIST_MAIN;
        player->main_continue_point = NULL;
    } else {
        player->current_node = player->main_list->head;
        player->current_source = LIST_MAIN;
    }
    
    if (!player->current_node) {
        player->is_playing = false;
        return NULL;
    }
    
    return player->current_node->song;
}

Song* player_next(Player* player) {
    if (!player) return NULL;
    
    int total = player->main_list->count + player->temp_list->count;
    if (total == 0) {
        player->is_playing = false;
        return NULL;
    }
    
    if (player->current_source == LIST_TEMP && player->current_node) {
        Song* next_song = handle_temp_song_finished(player);
        if (next_song) {
            player->is_playing = true;
            return next_song;
        }
    }
    
    if (player->temp_list->count > 0 && player->current_source == LIST_MAIN) {
        player->main_continue_point = player->current_node ? 
                                       player->current_node->next : player->main_list->head;
        if (!player->main_continue_point) {
            player->main_continue_point = player->main_list->head;
        }
        player->current_node = player->temp_list->head;
        player->current_source = LIST_TEMP;
        player->is_playing = true;
        return player->current_node->song;
    }
    
    switch (player->mode) {
        case PLAY_MODE_REPEAT_ONE:
            player->is_playing = true;
            return player->current_node ? player->current_node->song : NULL;
        
        case PLAY_MODE_SHUFFLE: {
            if (total <= 1) {
                player->is_playing = true;
                return player->current_node ? player->current_node->song : NULL;
            }
            int next_idx = shuffle_manager_get_next(player->shuffle_mgr);
            if (next_idx >= 0) {
                player->current_node = get_node_by_combined_index(player, next_idx);
                player->current_source = get_source_by_combined_index(player, next_idx);
                player->is_playing = true;
                return player->current_node ? player->current_node->song : NULL;
            }
            break;
        }
        
        case PLAY_MODE_SEQUENTIAL:
        case PLAY_MODE_REPEAT_ALL:
        default: {
            if (!player->current_node) {
                player->current_node = player->main_list->head;
                player->current_source = LIST_MAIN;
                player->is_playing = (player->current_node != NULL);
                return player->current_node ? player->current_node->song : NULL;
            }
            
            if (player->current_node->next) {
                player->current_node = player->current_node->next;
            } else {
                if (player->mode == PLAY_MODE_REPEAT_ALL) {
                    player->current_node = (player->current_source == LIST_MAIN) ? 
                                            player->main_list->head : player->temp_list->head;
                } else {
                    player->is_playing = false;
                    return NULL;
                }
            }
            player->is_playing = true;
            return player->current_node ? player->current_node->song : NULL;
        }
    }
    
    return NULL;
}

Song* player_prev(Player* player) {
    if (!player) return NULL;
    
    int total = player->main_list->count + player->temp_list->count;
    if (total == 0) {
        return NULL;
    }
    
    if (!player->current_node) {
        return player_play(player);
    }
    
    if (player->mode == PLAY_MODE_REPEAT_ONE) {
        return player->current_node->song;
    }
    
    if (player->mode == PLAY_MODE_SHUFFLE && total > 1) {
        int prev_idx = shuffle_manager_get_prev(player->shuffle_mgr);
        if (prev_idx >= 0) {
            player->current_node = get_node_by_combined_index(player, prev_idx);
            player->current_source = get_source_by_combined_index(player, prev_idx);
            return player->current_node ? player->current_node->song : NULL;
        }
    }
    
    if (player->current_node->prev) {
        player->current_node = player->current_node->prev;
    } else {
        SongList* list = (player->current_source == LIST_MAIN) ? 
                         player->main_list : player->temp_list;
        player->current_node = list->tail;
    }
    
    return player->current_node ? player->current_node->song : NULL;
}

Song* player_get_current_song(Player* player) {
    if (!player || !player->current_node) return NULL;
    return player->current_node->song;
}

int player_get_current_position(Player* player) {
    if (!player || !player->current_node) return -1;
    if (player->current_source == LIST_MAIN) {
        return song_list_get_position(player->main_list, player->current_node);
    }
    return -1;
}

void player_print_list(Player* player, bool favorites_only) {
    if (!player) return;
    printf("=== 播放列表 ===\n");
    printf("模式: %s\n", player_mode_to_string(player->mode));
    printf("状态: %s\n", player->is_playing ? (player->is_paused ? "暂停" : "播放中") : "停止");
    
    printf("\n--- 主列表 (%d 首) ---\n", player->main_list->count);
    song_list_print(player->main_list, favorites_only);
    
    if (player->temp_list->count > 0) {
        printf("\n--- 临时插队 (%d 首) ---\n", player->temp_list->count);
        song_list_print(player->temp_list, favorites_only);
    }
    
    Song* current = player_get_current_song(player);
    if (current) {
        int min = current->duration / 60;
        int sec = current->duration % 60;
        printf("\n>>> 当前播放: [%d] %s - %s (%02d:%02d) %s\n",
               current->id, current->title, current->artist, min, sec,
               current->source == SOURCE_TEMP ? "[临时]" : "");
    } else {
        printf("\n>>> 当前播放: (无)\n");
    }
}

int player_get_total_songs(Player* player) {
    if (!player) return 0;
    return player->main_list->count + player->temp_list->count;
}
