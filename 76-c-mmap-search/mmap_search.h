#ifndef MMAP_SEARCH_H
#define MMAP_SEARCH_H

#include <pthread.h>
#include <stddef.h>

#define SMALL_FILE_THRESHOLD (1024 * 1024)
#define BLOCK_SIZE (4 * 1024 * 1024)

typedef struct {
    char *filename;
    int matched;
    int match_count;
    pthread_mutex_t mutex;
} file_result_t;

typedef struct {
    const char *search_pattern;
    int ignore_case;
    int show_line_numbers;
    int show_filename_only;
    int show_count_only;
    file_result_t *result;
    const char *data;
    size_t data_size;
    size_t block_offset;
    size_t block_actual_size;
    int block_index;
    int total_blocks;
    int is_first_block;
    int is_last_block;
    size_t file_size;
} search_task_t;

int is_binary_file(const char *data, size_t size);
int string_match(const char *str, const char *pattern, int ignore_case);
void search_in_block(search_task_t *task);
size_t find_next_newline(const char *data, size_t start, size_t end);
size_t find_prev_newline(const char *data, size_t start, size_t end);

#endif
