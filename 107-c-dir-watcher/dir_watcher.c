#include "dir_watcher.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <dirent.h>
#include <errno.h>
#include <linux/limits.h>

typedef struct {
    int wd;
    char path[MAX_PATH_LEN];
} WatchDescriptor;

struct DirWatcher {
    int inotify_fd;
    WatchDescriptor *watch_descriptors;
    size_t watch_count;
    size_t watch_capacity;
    WatcherConfig config;
    bool running;
};

static bool should_ignore(const WatcherConfig *config, const char *name) {
    for (size_t i = 0; i < config->ignore_count; i++) {
        if (strcmp(name, config->ignore_patterns[i]) == 0) {
            return true;
        }
    }
    return false;
}

static bool should_ignore_path(const WatcherConfig *config, const char *path) {
    char tmp[MAX_PATH_LEN];
    strncpy(tmp, path, MAX_PATH_LEN - 1);
    tmp[MAX_PATH_LEN - 1] = '\0';
    
    char *token = strtok(tmp, "/");
    while (token != NULL) {
        if (should_ignore(config, token)) {
            return true;
        }
        token = strtok(NULL, "/");
    }
    return false;
}

static int add_watch_internal(DirWatcher *watcher, const char *path);

static int scan_and_watch_directory(DirWatcher *watcher, const char *dir_path) {
    DIR *dir = opendir(dir_path);
    if (!dir) {
        perror("opendir");
        return -1;
    }
    
    struct dirent *entry;
    while ((entry = readdir(dir)) != NULL) {
        if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..") == 0) {
            continue;
        }
        
        if (should_ignore(&watcher->config, entry->d_name)) {
            continue;
        }
        
        char full_path[MAX_PATH_LEN];
        int ret = snprintf(full_path, MAX_PATH_LEN, "%s/%s", dir_path, entry->d_name);
        if (ret < 0 || (size_t)ret >= MAX_PATH_LEN) {
            closedir(dir);
            return -1;
        }
        
        struct stat st;
        if (stat(full_path, &st) == -1) {
            continue;
        }
        
        if (S_ISDIR(st.st_mode)) {
            if (add_watch_internal(watcher, full_path) == -1) {
                closedir(dir);
                return -1;
            }
        }
    }
    
    closedir(dir);
    return 0;
}

static int add_watch_internal(DirWatcher *watcher, const char *path) {
    uint32_t mask = IN_CREATE | IN_DELETE | IN_MODIFY | IN_MOVED_FROM | IN_MOVED_TO;
    int wd = inotify_add_watch(watcher->inotify_fd, path, mask);
    
    if (wd == -1) {
        perror("inotify_add_watch");
        return -1;
    }
    
    if (watcher->watch_count >= watcher->watch_capacity) {
        size_t new_capacity = watcher->watch_capacity * 2;
        WatchDescriptor *new_wd = realloc(watcher->watch_descriptors, 
                                           new_capacity * sizeof(WatchDescriptor));
        if (!new_wd) {
            inotify_rm_watch(watcher->inotify_fd, wd);
            return -1;
        }
        watcher->watch_descriptors = new_wd;
        watcher->watch_capacity = new_capacity;
    }
    
    watcher->watch_descriptors[watcher->watch_count].wd = wd;
    strncpy(watcher->watch_descriptors[watcher->watch_count].path, path, MAX_PATH_LEN - 1);
    watcher->watch_descriptors[watcher->watch_count].path[MAX_PATH_LEN - 1] = '\0';
    watcher->watch_count++;
    
    return scan_and_watch_directory(watcher, path);
}

static const char *get_path_by_wd(const DirWatcher *watcher, int wd) {
    for (size_t i = 0; i < watcher->watch_count; i++) {
        if (watcher->watch_descriptors[i].wd == wd) {
            return watcher->watch_descriptors[i].path;
        }
    }
    return NULL;
}

static int remove_watch_by_wd(DirWatcher *watcher, int wd) {
    for (size_t i = 0; i < watcher->watch_count; i++) {
        if (watcher->watch_descriptors[i].wd == wd) {
            inotify_rm_watch(watcher->inotify_fd, wd);
            if (i < watcher->watch_count - 1) {
                memcpy(&watcher->watch_descriptors[i], 
                       &watcher->watch_descriptors[i + 1],
                       (watcher->watch_count - i - 1) * sizeof(WatchDescriptor));
            }
            watcher->watch_count--;
            return 0;
        }
    }
    return -1;
}

DirWatcher *dir_watcher_create(const WatcherConfig *config) {
    if (!config || !config->root_dir[0]) {
        return NULL;
    }
    
    DirWatcher *watcher = malloc(sizeof(DirWatcher));
    if (!watcher) return NULL;
    
    memset(watcher, 0, sizeof(DirWatcher));
    
    watcher->inotify_fd = inotify_init();
    if (watcher->inotify_fd == -1) {
        perror("inotify_init");
        free(watcher);
        return NULL;
    }
    
    int flags = fcntl(watcher->inotify_fd, F_GETFL);
    if (flags == -1 || fcntl(watcher->inotify_fd, F_SETFL, flags | O_NONBLOCK) == -1) {
        perror("fcntl");
        close(watcher->inotify_fd);
        free(watcher);
        return NULL;
    }
    
    watcher->watch_capacity = 64;
    watcher->watch_descriptors = malloc(watcher->watch_capacity * sizeof(WatchDescriptor));
    if (!watcher->watch_descriptors) {
        close(watcher->inotify_fd);
        free(watcher);
        return NULL;
    }
    watcher->watch_count = 0;
    
    memcpy(&watcher->config, config, sizeof(WatcherConfig));
    watcher->running = false;
    
    return watcher;
}

void dir_watcher_destroy(DirWatcher *watcher) {
    if (!watcher) return;
    
    dir_watcher_stop(watcher);
    
    for (size_t i = 0; i < watcher->watch_count; i++) {
        inotify_rm_watch(watcher->inotify_fd, watcher->watch_descriptors[i].wd);
    }
    
    free(watcher->watch_descriptors);
    close(watcher->inotify_fd);
    free(watcher);
}

int dir_watcher_start(DirWatcher *watcher) {
    if (!watcher || watcher->running) {
        return -1;
    }
    
    struct stat st;
    if (stat(watcher->config.root_dir, &st) == -1) {
        perror("stat root_dir");
        return -1;
    }
    
    if (!S_ISDIR(st.st_mode)) {
        fprintf(stderr, "%s is not a directory\n", watcher->config.root_dir);
        return -1;
    }
    
    if (add_watch_internal(watcher, watcher->config.root_dir) == -1) {
        return -1;
    }
    
    watcher->running = true;
    return 0;
}

void dir_watcher_stop(DirWatcher *watcher) {
    if (!watcher) return;
    watcher->running = false;
}

int dir_watcher_get_fd(const DirWatcher *watcher) {
    if (!watcher) return -1;
    return watcher->inotify_fd;
}

ssize_t dir_watcher_read_events(DirWatcher *watcher, char *buffer, size_t buf_size) {
    if (!watcher || !buffer || buf_size == 0) {
        return -1;
    }
    
    ssize_t len = read(watcher->inotify_fd, buffer, buf_size);
    if (len == -1 && errno != EAGAIN && errno != EWOULDBLOCK) {
        perror("read inotify");
        return -1;
    }
    
    return len;
}

void dir_watcher_process_event(DirWatcher *watcher, const struct inotify_event *event,
                                 EventCallback callback, void *user_data) {
    if (!watcher || !event || !callback) {
        return;
    }
    
    const char *dir_path = get_path_by_wd(watcher, event->wd);
    if (!dir_path) {
        return;
    }
    
    if (!(event->mask & IN_ISDIR)) {
        char full_path[MAX_PATH_LEN];
        int ret = snprintf(full_path, MAX_PATH_LEN, "%s/%s", dir_path, event->name);
        if (ret < 0 || (size_t)ret >= MAX_PATH_LEN) {
            return;
        }
        
        if (should_ignore_path(&watcher->config, full_path)) {
            return;
        }
        
        FileEvent file_event;
        memset(&file_event, 0, sizeof(FileEvent));
        strncpy(file_event.path, full_path, MAX_PATH_LEN - 1);
        file_event.path[MAX_PATH_LEN - 1] = '\0';
        file_event.timestamp = time(NULL);
        
        if (event->mask & (IN_CREATE | IN_MOVED_TO)) {
            file_event.type = EVENT_CREATE;
            callback(&file_event, user_data);
        }
        else if (event->mask & IN_MODIFY) {
            file_event.type = EVENT_MODIFY;
            callback(&file_event, user_data);
        }
        else if (event->mask & (IN_DELETE | IN_MOVED_FROM)) {
            file_event.type = EVENT_DELETE;
            callback(&file_event, user_data);
        }
    }
    else {
        if (event->mask & IN_CREATE) {
            char full_path[MAX_PATH_LEN];
            int ret = snprintf(full_path, MAX_PATH_LEN, "%s/%s", dir_path, event->name);
            if (ret >= 0 && (size_t)ret < MAX_PATH_LEN) {
                if (!should_ignore_path(&watcher->config, full_path)) {
                    add_watch_internal(watcher, full_path);
                }
            }
        }
        else if (event->mask & IN_DELETE) {
            char full_path[MAX_PATH_LEN];
            int ret = snprintf(full_path, MAX_PATH_LEN, "%s/%s", dir_path, event->name);
            if (ret >= 0 && (size_t)ret < MAX_PATH_LEN) {
                for (size_t i = 0; i < watcher->watch_count; ) {
                    if (strncmp(watcher->watch_descriptors[i].path, full_path, 
                                strlen(full_path)) == 0) {
                        remove_watch_by_wd(watcher, watcher->watch_descriptors[i].wd);
                    } else {
                        i++;
                    }
                }
            }
        }
    }
}

bool dir_watcher_is_running(const DirWatcher *watcher) {
    return watcher && watcher->running;
}
