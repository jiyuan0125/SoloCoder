#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <getopt.h>
#include <time.h>
#include <sys/stat.h>
#include <unistd.h>
#include <limits.h>

#include "snapshot.h"
#include "scanner.h"
#include "diff.h"

#define DEFAULT_SNAPSHOT_FILE ".filewatch.snap"
#define PENDING_DELETE_FILE ".filewatch.pending"
#define MERGE_WINDOW_SECONDS 10

typedef struct PendingDelete {
    char *path;
    EntryType entry_type;
    off_t size;
    time_t mtime;
    time_t delete_time;
} PendingDelete;

typedef struct PendingList {
    PendingDelete *items;
    size_t count;
    size_t capacity;
} PendingList;

static PendingList *pending_list_create(void) {
    PendingList *list = (PendingList *)malloc(sizeof(PendingList));
    if (!list) return NULL;
    list->items = NULL;
    list->count = 0;
    list->capacity = 0;
    return list;
}

static void pending_list_destroy(PendingList *list) {
    if (!list) return;
    for (size_t i = 0; i < list->count; i++) {
        free(list->items[i].path);
    }
    free(list->items);
    free(list);
}

static int pending_list_add(PendingList *list, const char *path, EntryType type, 
                            off_t size, time_t mtime, time_t delete_time) {
    if (!list || !path) return -1;
    
    if (list->count >= list->capacity) {
        size_t new_cap = list->capacity == 0 ? 16 : list->capacity * 2;
        PendingDelete *new_items = (PendingDelete *)realloc(list->items, new_cap * sizeof(PendingDelete));
        if (!new_items) return -1;
        list->items = new_items;
        list->capacity = new_cap;
    }
    
    PendingDelete *item = &list->items[list->count];
    item->path = strdup(path);
    if (!item->path) return -1;
    item->entry_type = type;
    item->size = size;
    item->mtime = mtime;
    item->delete_time = delete_time;
    list->count++;
    return 0;
}

static int pending_list_save(PendingList *list, const char *filename) {
    if (!list || !filename) return -1;
    
    FILE *fp = fopen(filename, "wb");
    if (!fp) return -1;
    
    uint32_t magic = 0x50454E44;
    uint32_t version = 1;
    uint64_t count = list->count;
    
    fwrite(&magic, sizeof(magic), 1, fp);
    fwrite(&version, sizeof(version), 1, fp);
    fwrite(&count, sizeof(count), 1, fp);
    
    for (size_t i = 0; i < list->count; i++) {
        uint32_t path_len = (uint32_t)strlen(list->items[i].path) + 1;
        uint8_t type = (uint8_t)list->items[i].entry_type;
        uint64_t size = (uint64_t)list->items[i].size;
        uint64_t mtime = (uint64_t)list->items[i].mtime;
        uint64_t delete_time = (uint64_t)list->items[i].delete_time;
        
        fwrite(&path_len, sizeof(path_len), 1, fp);
        fwrite(list->items[i].path, 1, path_len, fp);
        fwrite(&type, sizeof(type), 1, fp);
        fwrite(&size, sizeof(size), 1, fp);
        fwrite(&mtime, sizeof(mtime), 1, fp);
        fwrite(&delete_time, sizeof(delete_time), 1, fp);
    }
    
    uint32_t end_magic = 0x50454E44;
    fwrite(&end_magic, sizeof(end_magic), 1, fp);
    
    fclose(fp);
    return 0;
}

static PendingList *pending_list_load(const char *filename) {
    if (!filename) return NULL;
    
    FILE *fp = fopen(filename, "rb");
    if (!fp) return pending_list_create();
    
    uint32_t magic, version, end_magic;
    uint64_t count;
    
    if (fread(&magic, sizeof(magic), 1, fp) != 1 || magic != 0x50454E44) {
        fclose(fp);
        return pending_list_create();
    }
    
    if (fread(&version, sizeof(version), 1, fp) != 1 || version != 1) {
        fclose(fp);
        return pending_list_create();
    }
    
    if (fread(&count, sizeof(count), 1, fp) != 1) {
        fclose(fp);
        return pending_list_create();
    }
    
    PendingList *list = pending_list_create();
    if (!list) {
        fclose(fp);
        return NULL;
    }
    
    for (uint64_t i = 0; i < count; i++) {
        uint32_t path_len;
        uint8_t type;
        uint64_t size, mtime, delete_time;
        char *path;
        
        if (fread(&path_len, sizeof(path_len), 1, fp) != 1) {
            pending_list_destroy(list);
            fclose(fp);
            return pending_list_create();
        }
        
        path = (char *)malloc(path_len);
        if (!path || fread(path, 1, (size_t)path_len, fp) != (size_t)path_len) {
            free(path);
            pending_list_destroy(list);
            fclose(fp);
            return pending_list_create();
        }
        
        if (fread(&type, sizeof(type), 1, fp) != 1 ||
            fread(&size, sizeof(size), 1, fp) != 1 ||
            fread(&mtime, sizeof(mtime), 1, fp) != 1 ||
            fread(&delete_time, sizeof(delete_time), 1, fp) != 1) {
            free(path);
            pending_list_destroy(list);
            fclose(fp);
            return pending_list_create();
        }
        
        if (pending_list_add(list, path, (EntryType)type, (off_t)size, 
                             (time_t)mtime, (time_t)delete_time) != 0) {
            free(path);
        }
        free(path);
    }
    
    if (fread(&end_magic, sizeof(end_magic), 1, fp) != 1 || end_magic != 0x50454E44) {
        pending_list_destroy(list);
        fclose(fp);
        return pending_list_create();
    }
    
    fclose(fp);
    return list;
}

static Snapshot *load_snapshot_with_backup(const char *snapshot_file) {
    Snapshot *snap = snapshot_load(snapshot_file);
    if (snap) return snap;
    
    char backup_file[PATH_MAX];
    if (snprintf(backup_file, sizeof(backup_file), "%s.bak", snapshot_file) >= (int)sizeof(backup_file)) {
        return NULL;
    }
    
    return snapshot_load(backup_file);
}

static int save_snapshot_with_backup(Snapshot *snap, const char *snapshot_file) {
    char backup_file[PATH_MAX];
    if (snprintf(backup_file, sizeof(backup_file), "%s.bak", snapshot_file) >= (int)sizeof(backup_file)) {
        return -1;
    }
    
    struct stat st;
    if (stat(snapshot_file, &st) == 0) {
        if (rename(snapshot_file, backup_file) != 0) {
            return -1;
        }
    }
    
    if (snapshot_save(snap, snapshot_file) != 0) {
        if (stat(backup_file, &st) == 0) {
            rename(backup_file, snapshot_file);
        }
        return -1;
    }
    
    return 0;
}

static void process_pending_deletes(ChangeList *changes, PendingList *pending, 
                                     Snapshot *old_snap, Snapshot *new_snap,
                                     time_t merge_window) {
    time_t now = time(NULL);
    ChangeList *final_changes = changelist_create();
    if (!final_changes) return;
    
    for (size_t i = 0; i < changes->count; i++) {
        Change *c = &changes->changes[i];
        int merged = 0;
        
        if (c->type == CHANGE_ADDED) {
            for (size_t j = 0; j < pending->count; j++) {
                PendingDelete *p = &pending->items[j];
                if (strcmp(c->path, p->path) == 0) {
                    time_t elapsed = now - p->delete_time;
                    if (elapsed <= merge_window) {
                        FileEntry *new_entry = snapshot_find_entry(new_snap, c->path);
                        if (new_entry) {
                            if (new_entry->size != p->size || new_entry->mtime != p->mtime) {
                                changelist_add(final_changes, c->path, CHANGE_MODIFIED, c->entry_type);
                            }
                        }
                        memmove(&pending->items[j], &pending->items[j + 1], 
                               (pending->count - j - 1) * sizeof(PendingDelete));
                        pending->count--;
                        merged = 1;
                        break;
                    }
                }
            }
        }
        
        if (!merged) {
            changelist_add(final_changes, c->path, c->type, c->entry_type);
        }
    }
    
    for (size_t i = 0; i < pending->count; ) {
        PendingDelete *p = &pending->items[i];
        time_t elapsed = now - p->delete_time;
        
        if (elapsed > merge_window) {
            int found = 0;
            for (size_t j = 0; j < final_changes->count; j++) {
                if (strcmp(final_changes->changes[j].path, p->path) == 0) {
                    found = 1;
                    break;
                }
            }
            
            if (!found) {
                changelist_add(final_changes, p->path, CHANGE_DELETED, p->entry_type);
            }
            
            free(p->path);
            memmove(&pending->items[i], &pending->items[i + 1], 
                   (pending->count - i - 1) * sizeof(PendingDelete));
            pending->count--;
        } else {
            i++;
        }
    }
    
    for (size_t i = 0; i < final_changes->count; ) {
        if (final_changes->changes[i].type == CHANGE_DELETED) {
            FileEntry *old_entry = NULL;
            if (old_snap) {
                old_entry = snapshot_find_entry(old_snap, final_changes->changes[i].path);
            }
            
            if (old_entry) {
                pending_list_add(pending, final_changes->changes[i].path,
                               final_changes->changes[i].entry_type,
                               old_entry->size, old_entry->mtime, now);
                
                free(final_changes->changes[i].path);
                memmove(&final_changes->changes[i], &final_changes->changes[i + 1],
                       (final_changes->count - i - 1) * sizeof(Change));
                final_changes->count--;
            } else {
                i++;
            }
        } else {
            i++;
        }
    }
    
    changelist_destroy(changes);
    *changes = *final_changes;
    free(final_changes);
}

static void print_usage(const char *prog_name) {
    printf("用法: %s [选项] <目录>\n", prog_name);
    printf("选项:\n");
    printf("  -s, --snapshot <文件>  指定快照文件路径 (默认: .filewatch.snap)\n");
    printf("  -w, --window <秒>      指定删除-新增合并时间窗口 (默认: 10秒)\n");
    printf("  -n, --no-symlinks      不跟随符号链接\n");
    printf("  -H, --no-hidden        忽略隐藏文件\n");
    printf("  -h, --help             显示此帮助信息\n");
    printf("\n");
    printf("输出格式:\n");
    printf("  [变更类型] 文件路径\n");
    printf("  变更类型: 新增, 已修改, 已删除\n");
}

int main(int argc, char *argv[]) {
    const char *snapshot_file = DEFAULT_SNAPSHOT_FILE;
    const char *target_dir = NULL;
    time_t merge_window = MERGE_WINDOW_SECONDS;
    
    ScannerOptions scan_opts = {
        .follow_symlinks = 0,
        .include_hidden = 1
    };
    
    static struct option long_options[] = {
        {"snapshot", required_argument, 0, 's'},
        {"window", required_argument, 0, 'w'},
        {"no-symlinks", no_argument, 0, 'n'},
        {"no-hidden", no_argument, 0, 'H'},
        {"help", no_argument, 0, 'h'},
        {0, 0, 0, 0}
    };
    
    int opt;
    while ((opt = getopt_long(argc, argv, "s:w:nHh", long_options, NULL)) != -1) {
        switch (opt) {
            case 's':
                snapshot_file = optarg;
                break;
            case 'w':
                merge_window = (time_t)atoi(optarg);
                if (merge_window < 1) merge_window = 1;
                break;
            case 'n':
                scan_opts.follow_symlinks = 0;
                break;
            case 'H':
                scan_opts.include_hidden = 0;
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
        fprintf(stderr, "错误: 未指定目标目录\n");
        print_usage(argv[0]);
        return 1;
    }
    
    target_dir = argv[optind];
    
    char pending_file[PATH_MAX];
    if (snprintf(pending_file, sizeof(pending_file), "%s.pending", snapshot_file) >= (int)sizeof(pending_file)) {
        fprintf(stderr, "错误: 快照文件路径过长\n");
        return 1;
    }
    
    Snapshot *old_snap = load_snapshot_with_backup(snapshot_file);
    PendingList *pending = pending_list_load(pending_file);
    if (!pending) {
        fprintf(stderr, "错误: 无法加载待处理列表\n");
        snapshot_destroy(old_snap);
        return 1;
    }
    
    Snapshot *new_snap = scanner_scan_directory(target_dir, &scan_opts);
    if (!new_snap) {
        fprintf(stderr, "错误: 无法扫描目录: %s\n", target_dir);
        snapshot_destroy(old_snap);
        pending_list_destroy(pending);
        return 1;
    }
    
    ChangeList *changes = diff_compare(old_snap, new_snap, NULL);
    if (!changes) {
        fprintf(stderr, "错误: 无法计算变更\n");
        snapshot_destroy(old_snap);
        snapshot_destroy(new_snap);
        pending_list_destroy(pending);
        return 1;
    }
    
    process_pending_deletes(changes, pending, old_snap, new_snap, merge_window);
    
    for (size_t i = 0; i < changes->count; i++) {
        Change *c = &changes->changes[i];
        const char *type_str = c->entry_type == ENTRY_DIR ? " [目录]" : "";
        printf("[%s]%s %s\n", change_type_to_string(c->type), type_str, c->path);
    }
    
    if (save_snapshot_with_backup(new_snap, snapshot_file) != 0) {
        fprintf(stderr, "警告: 无法保存快照文件\n");
    }
    
    if (pending_list_save(pending, pending_file) != 0) {
        fprintf(stderr, "警告: 无法保存待处理列表\n");
    }
    
    changelist_destroy(changes);
    snapshot_destroy(old_snap);
    snapshot_destroy(new_snap);
    pending_list_destroy(pending);
    
    return 0;
}
