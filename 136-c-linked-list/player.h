#ifndef PLAYER_H
#define PLAYER_H

#include "song_list.h"
#include "shuffle.h"

typedef enum {
    PLAY_MODE_SEQUENTIAL,
    PLAY_MODE_REPEAT_ALL,
    PLAY_MODE_REPEAT_ONE,
    PLAY_MODE_SHUFFLE
} PlayMode;

typedef enum {
    LIST_MAIN,
    LIST_TEMP
} WhichList;

typedef struct Player {
    SongList* main_list;
    SongList* temp_list;
    
    SongNode* current_node;
    WhichList current_source;
    
    SongNode* main_continue_point;
    
    PlayMode mode;
    ShuffleManager* shuffle_mgr;
    
    bool is_playing;
    bool is_paused;
} Player;

Player* player_create(void);
void player_destroy(Player* player);

void player_add_song(Player* player, Song* song);
void player_insert_song_at(Player* player, int position, Song* song);
void player_insert_temp_song(Player* player, Song* song);

void player_remove_song(Player* player, int song_id);
void player_move_song(Player* player, int from_pos, int to_pos);

void player_set_favorite(Player* player, int song_id, bool favorite);

void player_set_mode(Player* player, PlayMode mode);
PlayMode player_get_mode(Player* player);

Song* player_play(Player* player);
Song* player_play_at(Player* player, int position);
void player_pause(Player* player);
void player_resume(Player* player);
void player_stop(Player* player);

Song* player_next(Player* player);
Song* player_prev(Player* player);

Song* player_get_current_song(Player* player);
int player_get_current_position(Player* player);

void player_print_list(Player* player, bool favorites_only);
int player_get_total_songs(Player* player);

const char* player_mode_to_string(PlayMode mode);

#endif
