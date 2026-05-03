#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <poll.h>
#include <unistd.h>
#include <errno.h>

#include "common.h"
#include "dir_watcher.h"
#include "event_buffer.h"
#include "build_trigger.h"

static volatile sig_atomic_t g_running = 1;

typedef struct {
    DirWatcher *watcher;
    EventBuffer *buffer;
    BuildTrigger *trigger;
    const char *watch_dir;
    int debounce_ms;
    const char *build_command;
} AppContext;

static void signal_handler(int sig) {
    (void)sig;
    g_running = 0;
}

static void setup_signal_handlers(void) {
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    
    sa.sa_handler = signal_handler;
    sigemptyset(&sa.sa_mask);
    
    if (sigaction(SIGINT, &sa, NULL) == -1) {
        perror("sigaction SIGINT");
        exit(1);
    }
    
    if (sigaction(SIGTERM, &sa, NULL) == -1) {
        perror("sigaction SIGTERM");
        exit(1);
    }
}

static void on_file_event(const FileEvent *event, void *user_data) {
    EventBuffer *buffer = (EventBuffer *)user_data;
    
    const char *type_str = "";
    switch (event->type) {
        case EVENT_CREATE: type_str = "CREATE"; break;
        case EVENT_MODIFY: type_str = "MODIFY"; break;
        case EVENT_DELETE: type_str = "DELETE"; break;
    }
    
    printf("[RAW EVENT] %s: %s\n", type_str, event->path);
    event_buffer_add_event(buffer, event);
}

static void on_buffer_flush(const PathList *created, const PathList *modified,
                             const PathList *deleted, void *user_data) {
    AppContext *ctx = (AppContext *)user_data;
    PathList *all_changed = path_list_create(32);
    
    printf("\n========== CHANGE DETECTED ==========\n");
    
    if (created && created->count > 0) {
        printf("Created (%zu files):\n", created->count);
        for (size_t i = 0; i < created->count; i++) {
            printf("  + %s\n", created->paths[i]);
            if (!path_list_contains(all_changed, created->paths[i])) {
                path_list_append(all_changed, created->paths[i]);
            }
        }
    }
    
    if (modified && modified->count > 0) {
        printf("Modified (%zu files):\n", modified->count);
        for (size_t i = 0; i < modified->count; i++) {
            printf("  * %s\n", modified->paths[i]);
            if (!path_list_contains(all_changed, modified->paths[i])) {
                path_list_append(all_changed, modified->paths[i]);
            }
        }
    }
    
    if (deleted && deleted->count > 0) {
        printf("Deleted (%zu files):\n", deleted->count);
        for (size_t i = 0; i < deleted->count; i++) {
            printf("  - %s\n", deleted->paths[i]);
            if (!path_list_contains(all_changed, deleted->paths[i])) {
                path_list_append(all_changed, deleted->paths[i]);
            }
        }
    }
    
    printf("=======================================\n\n");
    
    if (ctx->trigger && all_changed->count > 0) {
        printf("Triggering build...\n");
        build_trigger_request(ctx->trigger, all_changed);
    }
    
    path_list_destroy(all_changed);
}

static void on_build_callback(const PathList *changed_files, void *user_data) {
    (void)user_data;
    
    if (changed_files && changed_files->count > 0) {
        char *joined = path_list_join(changed_files, " ");
        if (joined) {
            printf("[BUILD] Changed files: %s\n", joined);
            free(joined);
        }
    }
}

static void print_usage(const char *prog_name) {
    printf("Usage: %s [OPTIONS] <watch_directory>\n", prog_name);
    printf("\nOptions:\n");
    printf("  -d <ms>       Debounce time in milliseconds (default: 500)\n");
    printf("  -c <command>  Build command to execute on changes (optional)\n");
    printf("  -i <pattern>  Ignore pattern (can be specified multiple times)\n");
    printf("  -h            Show this help message\n");
    printf("\nDefault ignore patterns: node_modules, .git, dist, build\n");
}

int main(int argc, char *argv[]) {
    int opt;
    int debounce_ms = DEFAULT_DEBOUNCE_MS;
    const char *build_command = NULL;
    const char *watch_dir = NULL;
    
    WatcherConfig watcher_config;
    memset(&watcher_config, 0, sizeof(watcher_config));
    
    const char *default_ignores[] = {"node_modules", ".git", "dist", "build"};
    size_t default_ignore_count = sizeof(default_ignores) / sizeof(default_ignores[0]);
    
    for (size_t i = 0; i < default_ignore_count && i < MAX_IGNORE_PATTERNS; i++) {
        strncpy(watcher_config.ignore_patterns[watcher_config.ignore_count],
                default_ignores[i], MAX_PATH_LEN - 1);
        watcher_config.ignore_count++;
    }
    
    while ((opt = getopt(argc, argv, "d:c:i:h")) != -1) {
        switch (opt) {
            case 'd':
                debounce_ms = atoi(optarg);
                if (debounce_ms <= 0) debounce_ms = DEFAULT_DEBOUNCE_MS;
                break;
            case 'c':
                build_command = optarg;
                break;
            case 'i':
                if (watcher_config.ignore_count < MAX_IGNORE_PATTERNS) {
                    strncpy(watcher_config.ignore_patterns[watcher_config.ignore_count],
                            optarg, MAX_PATH_LEN - 1);
                    watcher_config.ignore_count++;
                }
                break;
            case 'h':
                print_usage(argv[0]);
                return 0;
            default:
                print_usage(argv[0]);
                return 1;
        }
    }
    
    if (optind >= argc) {
        fprintf(stderr, "Error: Watch directory not specified\n");
        print_usage(argv[0]);
        return 1;
    }
    
    watch_dir = argv[optind];
    strncpy(watcher_config.root_dir, watch_dir, MAX_PATH_LEN - 1);
    
    setup_signal_handlers();
    
    printf("========================================\n");
    printf("Directory Watcher - File Change Monitor\n");
    printf("========================================\n");
    printf("Watching directory: %s\n", watch_dir);
    printf("Debounce time: %d ms\n", debounce_ms);
    if (build_command) {
        printf("Build command: %s\n", build_command);
    }
    printf("Ignore patterns:\n");
    for (size_t i = 0; i < watcher_config.ignore_count; i++) {
        printf("  - %s\n", watcher_config.ignore_patterns[i]);
    }
    printf("========================================\n");
    printf("Press Ctrl+C to exit\n\n");
    
    AppContext ctx;
    memset(&ctx, 0, sizeof(ctx));
    ctx.watch_dir = watch_dir;
    ctx.debounce_ms = debounce_ms;
    ctx.build_command = build_command;
    
    ctx.watcher = dir_watcher_create(&watcher_config);
    if (!ctx.watcher) {
        fprintf(stderr, "Failed to create directory watcher\n");
        return 1;
    }
    
    ctx.buffer = event_buffer_create(debounce_ms);
    if (!ctx.buffer) {
        fprintf(stderr, "Failed to create event buffer\n");
        dir_watcher_destroy(ctx.watcher);
        return 1;
    }
    
    if (dir_watcher_start(ctx.watcher) != 0) {
        fprintf(stderr, "Failed to start directory watcher\n");
        event_buffer_destroy(ctx.buffer);
        dir_watcher_destroy(ctx.watcher);
        return 1;
    }
    
    if (build_command) {
        BuildConfig build_config;
        memset(&build_config, 0, sizeof(build_config));
        build_config.command = build_command;
        build_config.use_shell = true;
        
        ctx.trigger = build_trigger_create(&build_config, on_build_callback, NULL);
        if (!ctx.trigger) {
            fprintf(stderr, "Warning: Failed to create build trigger\n");
        }
    }
    
    char inotify_buffer[INOTIFY_BUF_SIZE];
    struct pollfd fds[1];
    fds[0].fd = dir_watcher_get_fd(ctx.watcher);
    fds[0].events = POLLIN;
    
    while (g_running && dir_watcher_is_running(ctx.watcher)) {
        int timeout = event_buffer_has_pending_events(ctx.buffer) ? 100 : 1000;
        int poll_ret = poll(fds, 1, timeout);
        
        if (poll_ret == -1) {
            if (errno == EINTR) continue;
            perror("poll");
            break;
        }
        
        if (poll_ret > 0 && (fds[0].revents & POLLIN)) {
            ssize_t len = dir_watcher_read_events(ctx.watcher, inotify_buffer, sizeof(inotify_buffer));
            
            if (len > 0) {
                ssize_t i = 0;
                while (i < len) {
                    struct inotify_event *event = (struct inotify_event *)&inotify_buffer[i];
                    dir_watcher_process_event(ctx.watcher, event, on_file_event, ctx.buffer);
                    i += sizeof(struct inotify_event) + event->len;
                }
            }
        }
        
        if (event_buffer_should_flush(ctx.buffer)) {
            event_buffer_flush(ctx.buffer, on_buffer_flush, &ctx);
        }
    }
    
    printf("\nShutting down...\n");
    
    if (event_buffer_has_pending_events(ctx.buffer)) {
        printf("Flushing remaining events...\n");
        event_buffer_flush(ctx.buffer, on_buffer_flush, &ctx);
    }
    
    if (ctx.trigger) {
        printf("Waiting for build to complete...\n");
        build_trigger_wait(ctx.trigger);
        build_trigger_destroy(ctx.trigger);
    }
    
    dir_watcher_stop(ctx.watcher);
    dir_watcher_destroy(ctx.watcher);
    event_buffer_destroy(ctx.buffer);
    
    printf("Done. Exiting.\n");
    
    return 0;
}
