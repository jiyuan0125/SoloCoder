#include "dir_scan.h"
#include <stdio.h>
#include <string.h>
#include <dirent.h>
#include <sys/stat.h>
#include <libgen.h>

typedef struct {
    ino_t *inodes;
    size_t count;
    size_t capacity;
} InodeSet;

static InodeSet *inode_set_create(size_t initial_capacity) {
    InodeSet *set = malloc(sizeof(InodeSet));
    if (!set) return NULL;
    set->count = 0;
    set->capacity = initial_capacity > 0 ? initial_capacity : 128;
    set->inodes = malloc(set->capacity * sizeof(ino_t));
    if (!set->inodes) {
        free(set);
        return NULL;
    }
    return set;
}

static bool inode_set_contains(InodeSet *set, ino_t inode) {
    if (!set) return false;
    for (size_t i = 0; i < set->count; i++) {
        if (set->inodes[i] == inode) return true;
    }
    return false;
}

static void inode_set_add(InodeSet *set, ino_t inode) {
    if (!set) return;
    if (inode_set_contains(set, inode)) return;
    
    if (set->count >= set->capacity) {
        size_t new_capacity = set->capacity * 2;
        ino_t *new_inodes = realloc(set->inodes, new_capacity * sizeof(ino_t));
        if (new_inodes) {
            set->inodes = new_inodes;
            set->capacity = new_capacity;
        } else {
            return;
        }
    }
    set->inodes[set->count++] = inode;
}

static void inode_set_free(InodeSet *set) {
    if (set) {
        free(set->inodes);
        free(set);
    }
}

static bool should_exclude_file(const char *filename, const ScanConfig *config) {
    if (!config) return false;
    return pattern_list_match(&config->exclude_patterns, filename);
}

static bool should_exclude_dir(const char *dirname, const ScanConfig *config) {
    if (!config) return false;
    return pattern_list_match(&config->exclude_dirs, dirname);
}

static bool is_size_valid(off_t size, const ScanConfig *config) {
    if (!config) return true;
    if (config->min_size > 0 && size < config->min_size) return false;
    if (config->max_size > 0 && size > config->max_size) return false;
    return true;
}

static void scan_recursive(const char *path, const ScanConfig *config, 
                            FileList *files, InodeSet *inode_set) {
    DIR *dir = opendir(path);
    if (!dir) return;

    struct dirent *entry;
    while ((entry = readdir(dir)) != NULL) {
        if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..") == 0) {
            continue;
        }

        char full_path[MAX_PATH_LEN];
        if (snprintf(full_path, sizeof(full_path), "%s/%s", path, entry->d_name) >= 
            (int)sizeof(full_path)) {
            continue;
        }

        struct stat st;
        if (lstat(full_path, &st) != 0) {
            continue;
        }

        if (S_ISLNK(st.st_mode)) {
            if (config && config->follow_symlinks) {
                if (stat(full_path, &st) != 0) {
                    continue;
                }
            } else {
                continue;
            }
        }

        if (S_ISDIR(st.st_mode)) {
            if (!should_exclude_dir(entry->d_name, config)) {
                scan_recursive(full_path, config, files, inode_set);
            }
        } else if (S_ISREG(st.st_mode)) {
            if (!should_exclude_file(entry->d_name, config) && 
                is_size_valid(st.st_size, config)) {
                
                if (inode_set_contains(inode_set, st.st_ino)) {
                    continue;
                }
                inode_set_add(inode_set, st.st_ino);

                FileInfo info;
                memset(&info, 0, sizeof(info));
                strncpy(info.path, full_path, MAX_PATH_LEN - 1);
                info.size = st.st_size;
                info.inode = st.st_ino;
                file_list_append(files, &info);
            }
        }
    }
    closedir(dir);
}

void scan_config_init(ScanConfig *config) {
    if (!config) return;
    memset(config, 0, sizeof(ScanConfig));
    config->max_open_files = 32;
}

void scan_config_free(ScanConfig *config) {
    if (!config) return;
    pattern_list_clear(&config->exclude_dirs);
    pattern_list_clear(&config->exclude_patterns);
}

FileList *scan_directory(const char *path, const ScanConfig *config) {
    if (!path) return NULL;

    struct stat st;
    if (stat(path, &st) != 0) {
        return NULL;
    }
    if (!S_ISDIR(st.st_mode)) {
        return NULL;
    }

    FileList *files = file_list_create(0);
    if (!files) return NULL;

    InodeSet *inode_set = inode_set_create(0);
    if (!inode_set) {
        file_list_free(files);
        return NULL;
    }

    scan_recursive(path, config, files, inode_set);

    inode_set_free(inode_set);
    return files;
}
