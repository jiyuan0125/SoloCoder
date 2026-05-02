#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include "mmap_search.h"

int is_binary_file(const char *data, size_t size) {
    for (size_t i = 0; i < size; i++) {
        if (data[i] == '\0') {
            return 1;
        }
    }
    return 0;
}

static int char_icmp(char a, char b) {
    return tolower((unsigned char)a) == tolower((unsigned char)b);
}

int string_match(const char *str, const char *pattern, int ignore_case) {
    if (!pattern || !*pattern) return 0;
    if (!str) return 0;

    size_t pattern_len = strlen(pattern);
    size_t str_len = strlen(str);

    if (pattern_len > str_len) return 0;

    for (size_t i = 0; i <= str_len - pattern_len; i++) {
        int matched = 1;
        for (size_t j = 0; j < pattern_len; j++) {
            if (ignore_case) {
                if (!char_icmp(str[i + j], pattern[j])) {
                    matched = 0;
                    break;
                }
            } else {
                if (str[i + j] != pattern[j]) {
                    matched = 0;
                    break;
                }
            }
        }
        if (matched) {
            return 1;
        }
    }
    return 0;
}

size_t find_next_newline(const char *data, size_t start, size_t end) {
    for (size_t i = start; i < end; i++) {
        if (data[i] == '\n') {
            return i;
        }
    }
    return end;
}

size_t find_prev_newline(const char *data, size_t start, size_t end) {
    if (start == 0) return 0;
    for (size_t i = start - 1; i > end; i--) {
        if (data[i] == '\n') {
            return i + 1;
        }
    }
    return end;
}

static size_t count_lines(const char *data, size_t start, size_t end) {
    size_t count = 0;
    for (size_t i = start; i < end; i++) {
        if (data[i] == '\n') {
            count++;
        }
    }
    return count;
}

static void print_match(file_result_t *result, const char *filename, 
                        int line_num, const char *line_content, size_t line_len,
                        int show_line_numbers, int show_filename_only, int show_count_only) {
    pthread_mutex_lock(&result->mutex);
    
    result->matched = 1;
    result->match_count++;

    if (show_count_only) {
        pthread_mutex_unlock(&result->mutex);
        return;
    }

    if (show_filename_only) {
        pthread_mutex_unlock(&result->mutex);
        return;
    }

    printf("%s:%d:", filename, line_num);
    fwrite(line_content, 1, line_len, stdout);
    if (line_content[line_len - 1] != '\n') {
        printf("\n");
    }
    fflush(stdout);

    pthread_mutex_unlock(&result->mutex);
}

void search_in_block(search_task_t *task) {
    const char *data = task->data;
    const char *pattern = task->search_pattern;
    file_result_t *result = task->result;
    const char *filename = result->filename;

    size_t block_start = task->block_offset;
    size_t block_end = block_start + task->block_actual_size;
    
    size_t search_start, search_end;
    
    if (task->is_first_block) {
        search_start = 0;
    } else {
        search_start = find_prev_newline(data, block_start, 0);
    }
    
    if (task->is_last_block) {
        search_end = task->file_size;
    } else {
        search_end = find_next_newline(data, block_end, task->file_size);
        if (search_end < task->file_size) {
            search_end++;
        }
    }

    size_t lines_before = 0;
    if (search_start > 0) {
        lines_before = count_lines(data, 0, search_start);
    }

    size_t current_line_start = search_start;
    int current_line_num = (int)lines_before + 1;

    for (size_t i = search_start; i < search_end; i++) {
        if (data[i] == '\n' || i == search_end - 1) {
            size_t line_end = (data[i] == '\n') ? i + 1 : i + 1;
            size_t line_len = line_end - current_line_start;
            
            char *line_content = (char *)malloc(line_len + 1);
            if (!line_content) {
                current_line_start = line_end;
                current_line_num++;
                continue;
            }
            memcpy(line_content, data + current_line_start, line_len);
            line_content[line_len] = '\0';

            if (string_match(line_content, pattern, task->ignore_case)) {
                print_match(result, filename, current_line_num, 
                           line_content, line_len,
                           task->show_line_numbers, 
                           task->show_filename_only, 
                           task->show_count_only);
            }

            free(line_content);
            current_line_start = line_end;
            current_line_num++;
        }
    }
}
