#ifndef SONG_LIST_H
#define SONG_LIST_H

#include <stdbool.h>

#define MAX_TITLE_LEN  128
#define MAX_ARTIST_LEN 64
#define MAX_PATH_LEN   256

typedef enum {
    SOURCE_MAIN,
    SOURCE_TEMP
} SongSource;

typedef struct Song {
    int id;
    char title[MAX_TITLE_LEN];
    char artist[MAX_ARTIST_LEN];
    int duration;
    char filepath[MAX_PATH_LEN];
    bool is_favorite;
    SongSource source;
} Song;

typedef struct SongNode {
    Song* song;
    struct SongNode* prev;
    struct SongNode* next;
} SongNode;

typedef struct SongList {
    SongNode* head;
    SongNode* tail;
    int count;
    int next_id;
} SongList;

Song* song_create(const char* title, const char* artist, int duration, const char* filepath);
void song_destroy(Song* song);
SongList* song_list_create(void);
void song_list_destroy(SongList* list);
void song_list_append(SongList* list, Song* song);
void song_list_insert_at(SongList* list, int position, Song* song);
Song* song_list_remove_by_id(SongList* list, int id);
Song* song_list_remove_by_node(SongList* list, SongNode* node);
void song_list_move(SongList* list, int from_pos, int to_pos);
SongNode* song_list_get_node_at(SongList* list, int position);
SongNode* song_list_find_by_id(SongList* list, int id);
int song_list_get_position(SongList* list, SongNode* node);
void song_list_mark_favorite(SongList* list, int id, bool favorite);
void song_list_print(SongList* list, bool favorites_only);

#endif
