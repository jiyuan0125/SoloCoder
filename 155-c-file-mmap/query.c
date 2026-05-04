#include "query.h"

int init_search_session(SearchSession *session, const char *log_path, 
                        const char *index_path, int use_mmap) {
    memset(session, 0, sizeof(SearchSession));
    
    strncpy(session->log_path, log_path, MAX_PATH_LEN - 1);
    session->log_path[MAX_PATH_LEN - 1] = '\0';
    
    if (index_path) {
        strncpy(session->index_path, index_path, MAX_PATH_LEN - 1);
        session->index_path[MAX_PATH_LEN - 1] = '\0';
    } else {
        snprintf(session->index_path, MAX_PATH_LEN, "%s.idx", log_path);
    }
    
    session->use_mmap = use_mmap;
    
    if (open_file_for_search(&session->file_handle, log_path, use_mmap) != 0) {
        return -1;
    }
    
    if (load_index(&session->index, log_path, index_path) != 0) {
        if (init_index(&session->index, log_path, index_path) != 0) {
            close_file_handle(&session->file_handle);
            return -1;
        }
    }
    
    session->is_initialized = 1;
    return 0;
}

void close_search_session(SearchSession *session) {
    if (!session) return;
    
    if (session->is_initialized) {
        save_index(&session->index);
        free_index(&session->index);
        close_file_handle(&session->file_handle);
    }
    
    memset(session, 0, sizeof(SearchSession));
}

static int count_indexed_lines(SearchSession *session) {
    int max_line = 0;
    for (int i = 0; i < session->index.keyword_count; i++) {
        KeywordIndex *kw = &session->index.keywords[i];
        for (int j = 0; j < kw->entry_count; j++) {
            if (kw->entries[j].line_number > max_line) {
                max_line = kw->entries[j].line_number;
            }
        }
    }
    return max_line;
}

static off_t get_last_indexed_offset(SearchSession *session) {
    if (session->index.header.indexed_size > 0) {
        return session->index.header.indexed_size;
    }
    
    off_t max_offset = 0;
    for (int i = 0; i < session->index.keyword_count; i++) {
        KeywordIndex *kw = &session->index.keywords[i];
        for (int j = 0; j < kw->entry_count; j++) {
            off_t end_offset = kw->entries[j].file_offset + (off_t)kw->entries[j].line_length;
            if (end_offset > max_offset) {
                max_offset = end_offset;
            }
        }
    }
    return max_offset;
}

static int has_new_keywords(SearchSession *session, SearchConfig *config) {
    for (int i = 0; i < config->keyword_count; i++) {
        if (find_keyword(&session->index, config->keywords[i].keyword) == NULL) {
            return 1;
        }
    }
    return 0;
}

static void split_keywords(SearchSession *session, SearchConfig *config,
                           SearchConfig *new_keywords_config,
                           SearchConfig *existing_keywords_config) {
    memset(new_keywords_config, 0, sizeof(SearchConfig));
    memset(existing_keywords_config, 0, sizeof(SearchConfig));
    
    new_keywords_config->context_lines = config->context_lines;
    new_keywords_config->escape_regex = config->escape_regex;
    new_keywords_config->verbose = config->verbose;
    
    existing_keywords_config->context_lines = config->context_lines;
    existing_keywords_config->escape_regex = config->escape_regex;
    existing_keywords_config->verbose = config->verbose;
    
    for (int i = 0; i < config->keyword_count; i++) {
        if (find_keyword(&session->index, config->keywords[i].keyword) == NULL) {
            memcpy(&new_keywords_config->keywords[new_keywords_config->keyword_count],
                   &config->keywords[i], sizeof(SearchKeyword));
            new_keywords_config->keyword_count++;
        } else {
            memcpy(&existing_keywords_config->keywords[existing_keywords_config->keyword_count],
                   &config->keywords[i], sizeof(SearchKeyword));
            existing_keywords_config->keyword_count++;
        }
    }
}

int update_index_if_needed(SearchSession *session, SearchConfig *config) {
    if (!session->is_initialized) {
        return -1;
    }
    
    off_t current_size;
    if (get_file_size(session->log_path, &current_size) != 0) {
        return -1;
    }
    
    off_t last_offset = get_last_indexed_offset(session);
    int has_new_kw = has_new_keywords(session, config);
    int has_incremental = (current_size > last_offset);
    int updated = 0;
    
    if (has_new_kw) {
        SearchConfig new_kw_config;
        SearchConfig existing_kw_config;
        split_keywords(session, config, &new_kw_config, &existing_kw_config);
        
        if (new_kw_config.keyword_count > 0) {
            off_t actual_end = 0;
            search_file_chunk(&session->file_handle, 0, current_size,
                              &session->index, &new_kw_config, 0, &actual_end);
            
            if (actual_end > session->index.header.indexed_size) {
                update_indexed_size(&session->index, actual_end);
            }
            updated = 1;
        }
    }
    
    if (has_incremental) {
        int start_line = count_indexed_lines(session);
        off_t actual_end = last_offset;
        
        search_file_chunk(&session->file_handle, last_offset, current_size,
                          &session->index, config, start_line, &actual_end);
        
        if (actual_end > session->index.header.indexed_size) {
            update_indexed_size(&session->index, actual_end);
        }
        updated = 1;
    }
    
    return updated ? 1 : 0;
}

int search_from_index(SearchSession *session, SearchConfig *config, 
                      MatchResult *result) {
    if (!session->is_initialized) {
        return -1;
    }
    
    init_match_result(result);
    
    for (int i = 0; i < config->keyword_count; i++) {
        SearchKeyword *kw = &config->keywords[i];
        KeywordIndex *kw_idx = find_keyword(&session->index, kw->keyword);
        
        if (kw_idx) {
            for (int j = 0; j < kw_idx->entry_count; j++) {
                add_match_to_result(result, &kw_idx->entries[j]);
                kw->match_count++;
            }
        }
    }
    
    return 0;
}

void init_full_match_result(FullMatchResult *result) {
    result->results = NULL;
    result->count = 0;
    result->allocated = 0;
}

static void free_match_result_item(MatchResultWithContext *item) {
    if (item->line_content) {
        free(item->line_content);
        item->line_content = NULL;
    }
    free_context_lines(item->before_context, item->before_count);
    free_context_lines(item->after_context, item->after_count);
    item->before_context = NULL;
    item->after_context = NULL;
    item->before_count = 0;
    item->after_count = 0;
}

void free_full_match_result(FullMatchResult *result) {
    if (result->results) {
        for (int i = 0; i < result->count; i++) {
            free_match_result_item(&result->results[i]);
        }
        free(result->results);
        result->results = NULL;
    }
    result->count = 0;
    result->allocated = 0;
}

int add_to_full_match_result(FullMatchResult *result, 
                              const MatchResultWithContext *item) {
    if (result->count >= result->allocated) {
        int new_size = result->allocated == 0 ? 16 : result->allocated * 2;
        MatchResultWithContext *new_results = realloc(result->results, 
                                           new_size * sizeof(MatchResultWithContext));
        if (new_results == NULL) {
            return -1;
        }
        result->results = new_results;
        result->allocated = new_size;
    }
    
    MatchResultWithContext *dst = &result->results[result->count];
    memset(dst, 0, sizeof(MatchResultWithContext));
    
    memcpy(&dst->entry, &item->entry, sizeof(MatchEntry));
    
    if (item->line_content) {
        dst->line_content = safe_strdup(item->line_content);
    }
    
    dst->before_count = item->before_count;
    if (item->before_count > 0 && item->before_context) {
        dst->before_context = alloc_context_lines(item->before_count);
        if (dst->before_context) {
            for (int i = 0; i < item->before_count; i++) {
                if (item->before_context[i]) {
                    dst->before_context[i] = safe_strdup(item->before_context[i]);
                }
            }
        }
    }
    
    dst->after_count = item->after_count;
    if (item->after_count > 0 && item->after_context) {
        dst->after_context = alloc_context_lines(item->after_count);
        if (dst->after_context) {
            for (int i = 0; i < item->after_count; i++) {
                if (item->after_context[i]) {
                    dst->after_context[i] = safe_strdup(item->after_context[i]);
                }
            }
        }
    }
    
    result->count++;
    return 0;
}

int get_match_with_context(SearchSession *session, const MatchEntry *entry,
                            int context_lines, MatchResultWithContext *result) {
    memset(result, 0, sizeof(MatchResultWithContext));
    memcpy(&result->entry, entry, sizeof(MatchEntry));
    
    char buffer[BUFFER_SIZE];
    if (get_line_by_offset(&session->file_handle, entry->file_offset, buffer, BUFFER_SIZE) < 0) {
        return -1;
    }
    result->line_content = safe_strdup(buffer);
    
    if (context_lines > 0) {
        get_context_lines(&session->file_handle, entry, context_lines,
                         &result->before_context, &result->before_count,
                         &result->after_context, &result->after_count);
    }
    
    return 0;
}

static int compare_match_entries(const void *a, const void *b) {
    const MatchEntry *ma = (const MatchEntry *)a;
    const MatchEntry *mb = (const MatchEntry *)b;
    return ma->line_number - mb->line_number;
}

int search_with_context(SearchSession *session, SearchConfig *config,
                        FullMatchResult *result) {
    if (!session->is_initialized) {
        return -1;
    }
    
    init_full_match_result(result);
    
    MatchResult simple_result;
    init_match_result(&simple_result);
    
    if (search_from_index(session, config, &simple_result) != 0) {
        free_match_result(&simple_result);
        return -1;
    }
    
    if (simple_result.match_count > 1) {
        qsort(simple_result.matches, simple_result.match_count, 
              sizeof(MatchEntry), compare_match_entries);
    }
    
    for (int i = 0; i < simple_result.match_count; i++) {
        MatchResultWithContext item;
        if (get_match_with_context(session, &simple_result.matches[i],
                                    config->context_lines, &item) == 0) {
            add_to_full_match_result(result, &item);
            free_match_result_item(&item);
        }
    }
    
    free_match_result(&simple_result);
    return 0;
}
