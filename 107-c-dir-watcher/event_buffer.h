#ifndef EVENT_BUFFER_H
#define EVENT_BUFFER_H

#include "common.h"

typedef struct EventBuffer EventBuffer;
typedef void (*FlushCallback)(const PathList *created, const PathList *modified, 
                               const PathList *deleted, void *user_data);

EventBuffer *event_buffer_create(int debounce_ms);
void event_buffer_destroy(EventBuffer *buffer);

void event_buffer_add_event(EventBuffer *buffer, const FileEvent *event);
bool event_buffer_should_flush(const EventBuffer *buffer);
void event_buffer_flush(EventBuffer *buffer, FlushCallback callback, void *user_data);
void event_buffer_clear(EventBuffer *buffer);

time_t event_buffer_get_last_event_time(const EventBuffer *buffer);
int event_buffer_get_debounce_ms(const EventBuffer *buffer);
void event_buffer_set_debounce_ms(EventBuffer *buffer, int debounce_ms);

bool event_buffer_has_pending_events(const EventBuffer *buffer);

#endif
