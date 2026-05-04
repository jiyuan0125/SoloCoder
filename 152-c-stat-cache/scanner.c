#include "scanner.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <dirent.h>
#include <sys/stat.h>
#include <unistd.h>
#include <limits.h>

static ScannerOptions default_options = {
    .follow_symlinks = 0,
    .include_hidden = 1
};

static int scan_recursive(Snapshot *snap, const char *current_path, const char *base_path, ScannerOptions *options) {
    DIR *dir = opendir(current_path);
    if (!dir) return -1;
    
    struct dirent *entry;
    while ((entry = readdir(dir)) != NULL) {
        if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..") == 0) {
            continue;
        }
        
        if (!options->include_hidden && entry->d_name[0] == '.') {
            continue;
        }
        
        char full_path[PATH_MAX];
        char rel_path[PATH_MAX];
        
        if (snprintf(full_path, sizeof(full_path), "%s/%s", current_path, entry->d_name) >= (int)sizeof(full_path)) {
            continue;
        }
        
        if (strcmp(base_path, ".") == 0 || strcmp(base_path, current_path) == 0) {
            if (snprintf(rel_path, sizeof(rel_path), "%s", entry->d_name) >= (int)sizeof(rel_path)) {
                continue;
            }
        } else {
            const char *rel_start = current_path + strlen(base_path);
            if (*rel_start == '/') rel_start++;
            
            if (strlen(rel_start) > 0) {
                if (snprintf(rel_path, sizeof(rel_path), "%s/%s", rel_start, entry->d_name) >= (int)sizeof(rel_path)) {
                    continue;
                }
            } else {
                if (snprintf(rel_path, sizeof(rel_path), "%s", entry->d_name) >= (int)sizeof(rel_path)) {
                    continue;
                }
            }
        }
        
        struct stat st;
        int stat_result;
        
        if (options->follow_symlinks) {
            stat_result = stat(full_path, &st);
        } else {
            stat_result = lstat(full_path, &st);
        }
        
        if (stat_result != 0) {
            continue;
        }
        
        if (S_ISDIR(st.st_mode)) {
            if (snapshot_add_entry(snap, rel_path, 0, st.st_mtime, ENTRY_DIR) != 0) {
                closedir(dir);
                return -1;
            }
            if (scan_recursive(snap, full_path, base_path, options) != 0) {
                closedir(dir);
                return -1;
            }
        } else if (S_ISREG(st.st_mode)) {
            if (snapshot_add_entry(snap, rel_path, st.st_size, st.st_mtime, ENTRY_FILE) != 0) {
                closedir(dir);
                return -1;
            }
        }
    }
    
    closedir(dir);
    return 0;
}

Snapshot *scanner_scan_directory(const char *root_path, ScannerOptions *options) {
    if (!root_path) return NULL;
    
    ScannerOptions *opts = options ? options : &default_options;
    
    struct stat root_st;
    if (stat(root_path, &root_st) != 0 || !S_ISDIR(root_st.st_mode)) {
        return NULL;
    }
    
    Snapshot *snap = snapshot_create();
    if (!snap) return NULL;
    
    char abs_root[PATH_MAX];
    if (realpath(root_path, abs_root) == NULL) {
        strncpy(abs_root, root_path, sizeof(abs_root) - 1);
        abs_root[sizeof(abs_root) - 1] = '\0';
    }
    
    if (scan_recursive(snap, abs_root, abs_root, opts) != 0) {
        snapshot_destroy(snap);
        return NULL;
    }
    
    snapshot_sort(snap);
    return snap;
}
