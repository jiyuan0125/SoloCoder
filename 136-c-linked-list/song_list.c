#include "song_list.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static SongNode* song_node_create(Song* song) {
    SongNode* node = (SongNode*)malloc(sizeof(SongNode));
    if (!node) return NULL;
    node->song = song;
    node->prev = NULL;
    node->next = NULL;
    return node;
}

static void song_node_destroy(SongNode* node) {
    if (node) {
        if (node->song) {
            song_destroy(node->song);
        }
        free(node);
    }
}

Song* song_create(const char* title, const char* artist, int duration, const char* filepath) {
    Song* song = (Song*)malloc(sizeof(Song));
    if (!song) return NULL;
    song->id = -1;
    song->duration = duration;
    song->is_favorite = false;
    song->source = SOURCE_MAIN;
    strncpy(song->title, title, MAX_TITLE_LEN - 1);
    song->title[MAX_TITLE_LEN - 1] = '\0';
    strncpy(song->artist, artist, MAX_ARTIST_LEN - 1);
    song->artist[MAX_ARTIST_LEN - 1] = '\0';
    strncpy(song->filepath, filepath, MAX_PATH_LEN - 1);
    song->filepath[MAX_PATH_LEN - 1] = '\0';
    return song;
}

void song_destroy(Song* song) {
    free(song);
}

SongList* song_list_create(void) {
    SongList* list = (SongList*)malloc(sizeof(SongList));
    if (!list) return NULL;
    list->head = NULL;
    list->tail = NULL;
    list->count = 0;
    list->next_id = 1;
    return list;
}

void song_list_destroy(SongList* list) {
    if (!list) return;
    SongNode* current = list->head;
    while (current) {
        SongNode* next = current->next;
        song_node_destroy(current);
        current = next;
    }
    free(list);
}

void song_list_append(SongList* list, Song* song) {
    if (!list || !song) return;
    song->id = list->next_id++;
    SongNode* node = song_node_create(song);
    if (!list->head) {
        list->head = node;
        list->tail = node;
    } else {
        node->prev = list->tail;
        list->tail->next = node;
        list->tail = node;
    }
    list->count++;
}

void song_list_insert_at(SongList* list, int position, Song* song) {
    if (!list || !song) return;
    song->id = list->next_id++;
    SongNode* node = song_node_create(song);
    
    if (list->count == 0) {
        list->head = node;
        list->tail = node;
        list->count = 1;
        return;
    }
    
    if (position <= 0) {
        node->next = list->head;
        list->head->prev = node;
        list->head = node;
        list->count++;
        return;
    }
    
    if (position >= list->count) {
        node->prev = list->tail;
        list->tail->next = node;
        list->tail = node;
        list->count++;
        return;
    }
    
    SongNode* current = song_list_get_node_at(list, position);
    if (!current) {
        song_destroy(node->song);
        free(node);
        list->next_id--;
        return;
    }
    node->prev = current->prev;
    node->next = current;
    if (current->prev) {
        current->prev->next = node;
    } else {
        list->head = node;
    }
    current->prev = node;
    list->count++;
}

Song* song_list_remove_by_id(SongList* list, int id) {
    if (!list || list->count == 0) return NULL;
    SongNode* node = song_list_find_by_id(list, id);
    return song_list_remove_by_node(list, node);
}

Song* song_list_remove_by_node(SongList* list, SongNode* node) {
    if (!list || !node) return NULL;
    Song* song = node->song;
    if (node->prev) {
        node->prev->next = node->next;
    } else {
        list->head = node->next;
    }
    if (node->next) {
        node->next->prev = node->prev;
    } else {
        list->tail = node->prev;
    }
    free(node);
    list->count--;
    return song;
}

void song_list_move(SongList* list, int from_pos, int to_pos) {
    if (!list || list->count < 2) return;
    if (from_pos < 0 || from_pos >= list->count) return;
    if (to_pos < 0 || to_pos >= list->count) return;
    if (from_pos == to_pos) return;
    
    SongNode* from_node = song_list_get_node_at(list, from_pos);
    if (!from_node) return;
    
    SongNode* to_node = song_list_get_node_at(list, to_pos);
    if (!to_node) return;
    
    if (from_node->prev) {
        from_node->prev->next = from_node->next;
    } else {
        list->head = from_node->next;
    }
    if (from_node->next) {
        from_node->next->prev = from_node->prev;
    } else {
        list->tail = from_node->prev;
    }
    
    if (to_pos > from_pos) {
        from_node->prev = to_node;
        from_node->next = to_node->next;
        if (to_node->next) {
            to_node->next->prev = from_node;
        } else {
            list->tail = from_node;
        }
        to_node->next = from_node;
    } else {
        from_node->prev = to_node->prev;
        from_node->next = to_node;
        if (to_node->prev) {
            to_node->prev->next = from_node;
        } else {
            list->head = from_node;
        }
        to_node->prev = from_node;
    }
}

SongNode* song_list_get_node_at(SongList* list, int position) {
    if (!list || position < 0 || position >= list->count) return NULL;
    SongNode* current;
    int i;
    if (position <= list->count / 2) {
        current = list->head;
        for (i = 0; i < position && current; i++) {
            current = current->next;
        }
    } else {
        current = list->tail;
        for (i = list->count - 1; i > position && current; i--) {
            current = current->prev;
        }
    }
    return current;
}

SongNode* song_list_find_by_id(SongList* list, int id) {
    if (!list) return NULL;
    SongNode* current = list->head;
    while (current) {
        if (current->song && current->song->id == id) {
            return current;
        }
        current = current->next;
    }
    return NULL;
}

int song_list_get_position(SongList* list, SongNode* node) {
    if (!list || !node) return -1;
    int pos = 0;
    SongNode* current = list->head;
    while (current) {
        if (current == node) return pos;
        current = current->next;
        pos++;
    }
    return -1;
}

void song_list_mark_favorite(SongList* list, int id, bool favorite) {
    if (!list) return;
    SongNode* node = song_list_find_by_id(list, id);
    if (node && node->song) {
        node->song->is_favorite = favorite;
    }
}

void song_list_print(SongList* list, bool favorites_only) {
    if (!list) {
        printf("(空列表)\n");
        return;
    }
    if (list->count == 0) {
        printf("(空列表)\n");
        return;
    }
    SongNode* current = list->head;
    int idx = 0;
    while (current) {
        Song* s = current->song;
        if (!favorites_only || s->is_favorite) {
            int min = s->duration / 60;
            int sec = s->duration % 60;
            printf("[%d] ID:%d - %s - %s (%02d:%02d) %s%s\n",
                   idx, s->id, s->title, s->artist, min, sec,
                   s->is_favorite ? "[❤]" : "",
                   s->source == SOURCE_TEMP ? "[临时]" : "");
        }
        current = current->next;
        idx++;
    }
}
