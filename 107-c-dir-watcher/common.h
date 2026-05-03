#ifndef COMMON_H
#define COMMON_H

#include <stddef.h>
#include <stdbool.h>
#include <sys/types.h>
#include <time.h>

#define DEFAULT_DEBOUNCE_MS 500
#define MAX_PATH_LEN 4096
#define MAX_IGNORE_PATTERNS 128

typedef enum {
    EVENT_CREATE,
    EVENT_MODIFY,
    EVENT_DELETE
} EventType;

typedef struct {
    char path[MAX_PATH_LEN];
    EventType type;
    time_t timestamp;
} FileEvent;

typedef struct {
    char **paths;
    size_t count;
    size_t capacity;
} PathList;

typedef void (*BuildCallback)(const PathList *changed_files, void *user_data);

PathList *path_list_create(size_t initial_capacity);
void path_list_destroy(PathList *list);
bool path_list_append(PathList *list, const char *path);
bool path_list_contains(const PathList *list, const char *path);
void path_list_clear(PathList *list);
char *path_list_join(const PathList *list, const char *separator);

#endif
