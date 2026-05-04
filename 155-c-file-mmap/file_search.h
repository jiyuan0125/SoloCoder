#ifndef FILE_SEARCH_H
#define FILE_SEARCH_H

#include "common.h"
#include "index.h"

typedef struct {
    int fd;
    off_t file_size;
    char *mmap_addr;
    int use_mmap;
} FileHandle;

int open_file_for_search(FileHandle *fh, const char *path, int use_mmap);
void close_file_handle(FileHandle *fh);

int search_file_chunk(FileHandle *fh, off_t start_offset, off_t end_offset,
                      SearchIndex *index, SearchConfig *config,
                      int start_line_number, off_t *actual_end);

int get_line_by_offset(FileHandle *fh, off_t offset, char *buffer, size_t buf_size);
off_t find_line_start(FileHandle *fh, off_t offset);
off_t find_next_line(FileHandle *fh, off_t offset);

int get_context_lines(FileHandle *fh, const MatchEntry *match, int context_lines,
                      char ***before_lines, int *before_count,
                      char ***after_lines, int *after_count);

char **alloc_context_lines(int count);
void free_context_lines(char **lines, int count);

int simple_strstr_ignore_case(const char *haystack, const char *needle);
int simple_regex_match(const char *text, const char *pattern, int case_insensitive);

#endif
