#ifndef COMMON_H
#define COMMON_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <sys/stat.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/mman.h>
#include <ctype.h>

#define CONTEXT_LINES 3
#define MAX_KEYWORD_LEN 256
#define MAX_PATH_LEN 1024
#define BUFFER_SIZE (64 * 1024)
#define MAX_KEYWORDS 10

typedef struct {
    char keyword[MAX_KEYWORD_LEN];
    int line_number;
    off_t file_offset;
    size_t line_length;
} MatchEntry;

typedef struct {
    char keyword[MAX_KEYWORD_LEN];
    int is_regex;
    int match_count;
} SearchKeyword;

typedef struct {
    char log_path[MAX_PATH_LEN];
    char index_path[MAX_PATH_LEN];
    off_t log_size;
    off_t indexed_size;
    time_t last_index_time;
    int version;
} IndexHeader;

typedef struct {
    char keyword[MAX_KEYWORD_LEN];
    int entry_count;
    int allocated_count;
    MatchEntry *entries;
} KeywordIndex;

typedef struct {
    IndexHeader header;
    int keyword_count;
    int allocated_keywords;
    KeywordIndex *keywords;
} SearchIndex;

typedef struct {
    SearchKeyword keywords[MAX_KEYWORDS];
    int keyword_count;
    int context_lines;
    int escape_regex;
    int verbose;
} SearchConfig;

typedef struct {
    MatchEntry *matches;
    int match_count;
    int allocated_count;
} MatchResult;

void free_match_result(MatchResult *result);
void init_match_result(MatchResult *result);
int add_match_to_result(MatchResult *result, const MatchEntry *entry);

char *safe_strdup(const char *str);
char *escape_regex_chars(const char *str);
char *trim_newline(char *str);

#endif
