#ifndef QUERY_H
#define QUERY_H

#include "common.h"
#include "index.h"
#include "file_search.h"

typedef struct {
    SearchIndex index;
    FileHandle file_handle;
    char log_path[MAX_PATH_LEN];
    char index_path[MAX_PATH_LEN];
    int is_initialized;
    int use_mmap;
} SearchSession;

typedef struct {
    MatchEntry entry;
    char *line_content;
    char **before_context;
    int before_count;
    char **after_context;
    int after_count;
} MatchResultWithContext;

typedef struct {
    MatchResultWithContext *results;
    int count;
    int allocated;
} FullMatchResult;

int init_search_session(SearchSession *session, const char *log_path, 
                        const char *index_path, int use_mmap);
void close_search_session(SearchSession *session);

int update_index_if_needed(SearchSession *session, SearchConfig *config);

int search_from_index(SearchSession *session, SearchConfig *config, 
                      MatchResult *result);

int get_match_with_context(SearchSession *session, const MatchEntry *entry,
                            int context_lines, MatchResultWithContext *result);

void init_full_match_result(FullMatchResult *result);
void free_full_match_result(FullMatchResult *result);
int add_to_full_match_result(FullMatchResult *result, 
                              const MatchResultWithContext *item);

int search_with_context(SearchSession *session, SearchConfig *config,
                        FullMatchResult *result);

#endif
