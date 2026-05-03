#include "event_buffer.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>

#define VIM_DELETE_CREATE_WINDOW_MS 200

typedef struct {
    char path[MAX_PATH_LEN];
    EventType type;
    time_t timestamp_sec;
    long timestamp_usec;
} BufferedEvent;

struct EventBuffer {
    BufferedEvent *events;
    size_t count;
    size_t capacity;
    int debounce_ms;
    time_t last_event_sec;
    long last_event_usec;
};

static long get_time_diff_ms(time_t sec1, long usec1, time_t sec2, long usec2) {
    long diff_sec = (long)(sec1 - sec2);
    long diff_usec = usec1 - usec2;
    return diff_sec * 1000 + diff_usec / 1000;
}

static void get_current_time(time_t *sec, long *usec) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    *sec = tv.tv_sec;
    *usec = tv.tv_usec;
}

EventBuffer *event_buffer_create(int debounce_ms) {
    EventBuffer *buffer = malloc(sizeof(EventBuffer));
    if (!buffer) return NULL;
    
    memset(buffer, 0, sizeof(EventBuffer));
    
    buffer->capacity = 128;
    buffer->events = malloc(buffer->capacity * sizeof(BufferedEvent));
    if (!buffer->events) {
        free(buffer);
        return NULL;
    }
    
    buffer->count = 0;
    buffer->debounce_ms = debounce_ms > 0 ? debounce_ms : DEFAULT_DEBOUNCE_MS;
    buffer->last_event_sec = 0;
    buffer->last_event_usec = 0;
    
    return buffer;
}

void event_buffer_destroy(EventBuffer *buffer) {
    if (!buffer) return;
    free(buffer->events);
    free(buffer);
}

static int find_event_by_path(const EventBuffer *buffer, const char *path) {
    for (size_t i = 0; i < buffer->count; i++) {
        if (strcmp(buffer->events[i].path, path) == 0) {
            return (int)i;
        }
    }
    return -1;
}

static int find_delete_event_for_path(const EventBuffer *buffer, const char *path,
                                        time_t create_sec, long create_usec) {
    for (size_t i = 0; i < buffer->count; i++) {
        if (buffer->events[i].type == EVENT_DELETE &&
            strcmp(buffer->events[i].path, path) == 0) {
            long diff = get_time_diff_ms(create_sec, create_usec,
                                          buffer->events[i].timestamp_sec,
                                          buffer->events[i].timestamp_usec);
            if (diff >= 0 && diff <= VIM_DELETE_CREATE_WINDOW_MS) {
                return (int)i;
            }
        }
    }
    return -1;
}

void event_buffer_add_event(EventBuffer *buffer, const FileEvent *event) {
    if (!buffer || !event) return;
    
    time_t now_sec;
    long now_usec;
    get_current_time(&now_sec, &now_usec);
    
    buffer->last_event_sec = now_sec;
    buffer->last_event_usec = now_usec;
    
    int existing_idx = find_event_by_path(buffer, event->path);
    
    if (event->type == EVENT_CREATE && existing_idx == -1) {
        int delete_idx = find_delete_event_for_path(buffer, event->path, now_sec, now_usec);
        if (delete_idx >= 0) {
            buffer->events[delete_idx].type = EVENT_MODIFY;
            buffer->events[delete_idx].timestamp_sec = now_sec;
            buffer->events[delete_idx].timestamp_usec = now_usec;
            return;
        }
    }
    
    if (existing_idx >= 0) {
        BufferedEvent *existing = &buffer->events[existing_idx];
        
        if (existing->type == EVENT_CREATE) {
            if (event->type == EVENT_DELETE) {
                if (buffer->count > 1) {
                    memmove(&buffer->events[existing_idx],
                            &buffer->events[existing_idx + 1],
                            (buffer->count - existing_idx - 1) * sizeof(BufferedEvent));
                    buffer->count--;
                } else {
                    buffer->count--;
                }
            }
            return;
        }
        
        if (existing->type == EVENT_DELETE) {
            if (event->type == EVENT_CREATE) {
                long diff = get_time_diff_ms(now_sec, now_usec,
                                              existing->timestamp_sec,
                                              existing->timestamp_usec);
                if (diff <= VIM_DELETE_CREATE_WINDOW_MS) {
                    existing->type = EVENT_MODIFY;
                    existing->timestamp_sec = now_sec;
                    existing->timestamp_usec = now_usec;
                    return;
                }
            }
            existing->timestamp_sec = now_sec;
            existing->timestamp_usec = now_usec;
            return;
        }
        
        existing->timestamp_sec = now_sec;
        existing->timestamp_usec = now_usec;
        return;
    }
    
    if (buffer->count >= buffer->capacity) {
        size_t new_capacity = buffer->capacity * 2;
        BufferedEvent *new_events = realloc(buffer->events,
                                               new_capacity * sizeof(BufferedEvent));
        if (!new_events) return;
        
        buffer->events = new_events;
        buffer->capacity = new_capacity;
    }
    
    strncpy(buffer->events[buffer->count].path, event->path, MAX_PATH_LEN - 1);
    buffer->events[buffer->count].path[MAX_PATH_LEN - 1] = '\0';
    buffer->events[buffer->count].type = event->type;
    buffer->events[buffer->count].timestamp_sec = now_sec;
    buffer->events[buffer->count].timestamp_usec = now_usec;
    buffer->count++;
}

bool event_buffer_should_flush(const EventBuffer *buffer) {
    if (!buffer || buffer->count == 0) {
        return false;
    }
    
    time_t now_sec;
    long now_usec;
    get_current_time(&now_sec, &now_usec);
    
    long diff = get_time_diff_ms(now_sec, now_usec,
                                  buffer->last_event_sec,
                                  buffer->last_event_usec);
    
    return diff >= buffer->debounce_ms;
}

void event_buffer_flush(EventBuffer *buffer, FlushCallback callback, void *user_data) {
    if (!buffer || !callback || buffer->count == 0) {
        return;
    }
    
    PathList *created = path_list_create(32);
    PathList *modified = path_list_create(32);
    PathList *deleted = path_list_create(32);
    
    if (!created || !modified || !deleted) {
        path_list_destroy(created);
        path_list_destroy(modified);
        path_list_destroy(deleted);
        return;
    }
    
    for (size_t i = 0; i < buffer->count; i++) {
        switch (buffer->events[i].type) {
            case EVENT_CREATE:
                path_list_append(created, buffer->events[i].path);
                break;
            case EVENT_MODIFY:
                path_list_append(modified, buffer->events[i].path);
                break;
            case EVENT_DELETE:
                path_list_append(deleted, buffer->events[i].path);
                break;
        }
    }
    
    callback(created, modified, deleted, user_data);
    
    path_list_destroy(created);
    path_list_destroy(modified);
    path_list_destroy(deleted);
    
    event_buffer_clear(buffer);
}

void event_buffer_clear(EventBuffer *buffer) {
    if (!buffer) return;
    buffer->count = 0;
    buffer->last_event_sec = 0;
    buffer->last_event_usec = 0;
}

time_t event_buffer_get_last_event_time(const EventBuffer *buffer) {
    if (!buffer) return 0;
    return buffer->last_event_sec;
}

int event_buffer_get_debounce_ms(const EventBuffer *buffer) {
    if (!buffer) return 0;
    return buffer->debounce_ms;
}

void event_buffer_set_debounce_ms(EventBuffer *buffer, int debounce_ms) {
    if (!buffer) return;
    buffer->debounce_ms = debounce_ms > 0 ? debounce_ms : DEFAULT_DEBOUNCE_MS;
}

bool event_buffer_has_pending_events(const EventBuffer *buffer) {
    return buffer && buffer->count > 0;
}
