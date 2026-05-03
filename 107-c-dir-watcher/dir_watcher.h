#ifndef DIR_WATCHER_H
#define DIR_WATCHER_H

#include "common.h"
#include <sys/inotify.h>
#include <linux/limits.h>

#ifndef NAME_MAX
#define NAME_MAX 255
#endif

#define INOTIFY_BUF_SIZE (1024 * (sizeof(struct inotify_event) + NAME_MAX + 1))

typedef struct DirWatcher DirWatcher;
typedef void (*EventCallback)(const FileEvent *event, void *user_data);

typedef struct {
    char root_dir[MAX_PATH_LEN];
    char ignore_patterns[MAX_IGNORE_PATTERNS][MAX_PATH_LEN];
    size_t ignore_count;
} WatcherConfig;

DirWatcher *dir_watcher_create(const WatcherConfig *config);
void dir_watcher_destroy(DirWatcher *watcher);

int dir_watcher_start(DirWatcher *watcher);
void dir_watcher_stop(DirWatcher *watcher);

int dir_watcher_get_fd(const DirWatcher *watcher);
ssize_t dir_watcher_read_events(DirWatcher *watcher, char *buffer, size_t buf_size);
void dir_watcher_process_event(DirWatcher *watcher, const struct inotify_event *event, 
                                 EventCallback callback, void *user_data);

bool dir_watcher_is_running(const DirWatcher *watcher);

#endif
