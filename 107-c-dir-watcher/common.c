#include "common.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

PathList *path_list_create(size_t initial_capacity) {
    if (initial_capacity == 0) {
        initial_capacity = 16;
    }
    
    PathList *list = malloc(sizeof(PathList));
    if (!list) return NULL;
    
    list->paths = malloc(initial_capacity * sizeof(char *));
    if (!list->paths) {
        free(list);
        return NULL;
    }
    
    list->count = 0;
    list->capacity = initial_capacity;
    memset(list->paths, 0, initial_capacity * sizeof(char *));
    
    return list;
}

void path_list_destroy(PathList *list) {
    if (!list) return;
    
    for (size_t i = 0; i < list->count; i++) {
        free(list->paths[i]);
    }
    free(list->paths);
    free(list);
}

bool path_list_append(PathList *list, const char *path) {
    if (!list || !path) return false;
    
    if (list->count >= list->capacity) {
        size_t new_capacity = list->capacity * 2;
        char **new_paths = realloc(list->paths, new_capacity * sizeof(char *));
        if (!new_paths) return false;
        
        list->paths = new_paths;
        list->capacity = new_capacity;
    }
    
    list->paths[list->count] = strdup(path);
    if (!list->paths[list->count]) return false;
    
    list->count++;
    return true;
}

bool path_list_contains(const PathList *list, const char *path) {
    if (!list || !path) return false;
    
    for (size_t i = 0; i < list->count; i++) {
        if (strcmp(list->paths[i], path) == 0) {
            return true;
        }
    }
    return false;
}

void path_list_clear(PathList *list) {
    if (!list) return;
    
    for (size_t i = 0; i < list->count; i++) {
        free(list->paths[i]);
        list->paths[i] = NULL;
    }
    list->count = 0;
}

char *path_list_join(const PathList *list, const char *separator) {
    if (!list || list->count == 0 || !separator) {
        return NULL;
    }
    
    size_t total_len = 0;
    size_t sep_len = strlen(separator);
    
    for (size_t i = 0; i < list->count; i++) {
        total_len += strlen(list->paths[i]);
        if (i < list->count - 1) {
            total_len += sep_len;
        }
    }
    
    char *result = malloc(total_len + 1);
    if (!result) return NULL;
    
    char *ptr = result;
    for (size_t i = 0; i < list->count; i++) {
        size_t path_len = strlen(list->paths[i]);
        memcpy(ptr, list->paths[i], path_len);
        ptr += path_len;
        
        if (i < list->count - 1) {
            memcpy(ptr, separator, sep_len);
            ptr += sep_len;
        }
    }
    
    *ptr = '\0';
    return result;
}
