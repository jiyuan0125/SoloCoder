#ifndef INDEX_H
#define INDEX_H

#include "common.h"

#define INDEX_VERSION 1
#define INDEX_MAGIC "LOGIDX"

int init_index(SearchIndex *index, const char *log_path, const char *index_path);
void free_index(SearchIndex *index);

int save_index(SearchIndex *index);
int load_index(SearchIndex *index, const char *log_path, const char *index_path);

int get_file_size(const char *path, off_t *size);
int needs_reindex(SearchIndex *index, const char *log_path);

KeywordIndex *find_or_create_keyword(SearchIndex *index, const char *keyword);
KeywordIndex *find_keyword(SearchIndex *index, const char *keyword);

int add_match_to_index(SearchIndex *index, const char *keyword, const MatchEntry *entry);
int update_indexed_size(SearchIndex *index, off_t new_size);

#endif
