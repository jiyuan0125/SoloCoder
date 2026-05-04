#include "index.h"

int get_file_size(const char *path, off_t *size) {
    struct stat st;
    if (stat(path, &st) != 0) {
        return -1;
    }
    *size = st.st_size;
    return 0;
}

int init_index(SearchIndex *index, const char *log_path, const char *index_path) {
    memset(index, 0, sizeof(SearchIndex));
    
    strncpy(index->header.log_path, log_path, MAX_PATH_LEN - 1);
    index->header.log_path[MAX_PATH_LEN - 1] = '\0';
    
    if (index_path) {
        strncpy(index->header.index_path, index_path, MAX_PATH_LEN - 1);
    } else {
        snprintf(index->header.index_path, MAX_PATH_LEN, "%s.idx", log_path);
    }
    index->header.index_path[MAX_PATH_LEN - 1] = '\0';
    
    index->header.version = INDEX_VERSION;
    index->header.indexed_size = 0;
    index->header.last_index_time = time(NULL);
    
    if (get_file_size(log_path, &index->header.log_size) != 0) {
        index->header.log_size = 0;
    }
    
    index->keyword_count = 0;
    index->allocated_keywords = 0;
    index->keywords = NULL;
    
    return 0;
}

void free_index(SearchIndex *index) {
    if (index == NULL) return;
    
    for (int i = 0; i < index->keyword_count; i++) {
        if (index->keywords[i].entries) {
            free(index->keywords[i].entries);
            index->keywords[i].entries = NULL;
        }
    }
    
    if (index->keywords) {
        free(index->keywords);
        index->keywords = NULL;
    }
    
    index->keyword_count = 0;
    index->allocated_keywords = 0;
}

int needs_reindex(SearchIndex *index, const char *log_path) {
    off_t current_size;
    if (get_file_size(log_path, &current_size) != 0) {
        return -1;
    }
    
    if (current_size > index->header.indexed_size) {
        return 1;
    }
    return 0;
}

KeywordIndex *find_keyword(SearchIndex *index, const char *keyword) {
    for (int i = 0; i < index->keyword_count; i++) {
        if (strcmp(index->keywords[i].keyword, keyword) == 0) {
            return &index->keywords[i];
        }
    }
    return NULL;
}

KeywordIndex *find_or_create_keyword(SearchIndex *index, const char *keyword) {
    KeywordIndex *existing = find_keyword(index, keyword);
    if (existing) {
        return existing;
    }
    
    if (index->keyword_count >= index->allocated_keywords) {
        int new_size = index->allocated_keywords == 0 ? 8 : index->allocated_keywords * 2;
        KeywordIndex *new_keywords = realloc(index->keywords, new_size * sizeof(KeywordIndex));
        if (new_keywords == NULL) {
            return NULL;
        }
        index->keywords = new_keywords;
        index->allocated_keywords = new_size;
    }
    
    KeywordIndex *new_idx = &index->keywords[index->keyword_count];
    strncpy(new_idx->keyword, keyword, MAX_KEYWORD_LEN - 1);
    new_idx->keyword[MAX_KEYWORD_LEN - 1] = '\0';
    new_idx->entry_count = 0;
    new_idx->allocated_count = 0;
    new_idx->entries = NULL;
    
    index->keyword_count++;
    return new_idx;
}

int add_match_to_index(SearchIndex *index, const char *keyword, const MatchEntry *entry) {
    KeywordIndex *kw_idx = find_or_create_keyword(index, keyword);
    if (kw_idx == NULL) {
        return -1;
    }
    
    if (kw_idx->entry_count >= kw_idx->allocated_count) {
        int new_size = kw_idx->allocated_count == 0 ? 16 : kw_idx->allocated_count * 2;
        MatchEntry *new_entries = realloc(kw_idx->entries, new_size * sizeof(MatchEntry));
        if (new_entries == NULL) {
            return -1;
        }
        kw_idx->entries = new_entries;
        kw_idx->allocated_count = new_size;
    }
    
    memcpy(&kw_idx->entries[kw_idx->entry_count], entry, sizeof(MatchEntry));
    kw_idx->entry_count++;
    return 0;
}

int update_indexed_size(SearchIndex *index, off_t new_size) {
    index->header.indexed_size = new_size;
    index->header.last_index_time = time(NULL);
    if (get_file_size(index->header.log_path, &index->header.log_size) != 0) {
        index->header.log_size = new_size;
    }
    return 0;
}

int save_index(SearchIndex *index) {
    FILE *fp = fopen(index->header.index_path, "wb");
    if (fp == NULL) {
        return -1;
    }
    
    char magic[8] = {0};
    memcpy(magic, INDEX_MAGIC, strlen(INDEX_MAGIC));
    if (fwrite(magic, sizeof(char), 8, fp) != 8) {
        fclose(fp);
        return -1;
    }
    
    if (fwrite(&index->header, sizeof(IndexHeader), 1, fp) != 1) {
        fclose(fp);
        return -1;
    }
    
    if (fwrite(&index->keyword_count, sizeof(int), 1, fp) != 1) {
        fclose(fp);
        return -1;
    }
    
    for (int i = 0; i < index->keyword_count; i++) {
        KeywordIndex *kw = &index->keywords[i];
        
        if (fwrite(kw->keyword, sizeof(char), MAX_KEYWORD_LEN, fp) != MAX_KEYWORD_LEN) {
            fclose(fp);
            return -1;
        }
        
        if (fwrite(&kw->entry_count, sizeof(int), 1, fp) != 1) {
            fclose(fp);
            return -1;
        }
        
        if (kw->entry_count > 0) {
            if (fwrite(kw->entries, sizeof(MatchEntry), kw->entry_count, fp) != (size_t)kw->entry_count) {
                fclose(fp);
                return -1;
            }
        }
    }
    
    fclose(fp);
    return 0;
}

int load_index(SearchIndex *index, const char *log_path, const char *index_path) {
    char actual_index_path[MAX_PATH_LEN];
    if (index_path) {
        strncpy(actual_index_path, index_path, MAX_PATH_LEN - 1);
        actual_index_path[MAX_PATH_LEN - 1] = '\0';
    } else {
        snprintf(actual_index_path, MAX_PATH_LEN, "%s.idx", log_path);
    }
    
    FILE *fp = fopen(actual_index_path, "rb");
    if (fp == NULL) {
        return -1;
    }
    
    char magic[8] = {0};
    if (fread(magic, sizeof(char), 8, fp) != 8) {
        fclose(fp);
        return -1;
    }
    
    if (memcmp(magic, INDEX_MAGIC, strlen(INDEX_MAGIC)) != 0) {
        fclose(fp);
        return -1;
    }
    
    memset(index, 0, sizeof(SearchIndex));
    
    if (fread(&index->header, sizeof(IndexHeader), 1, fp) != 1) {
        fclose(fp);
        return -1;
    }
    
    if (index->header.version != INDEX_VERSION) {
        fclose(fp);
        return -1;
    }
    
    int keyword_count;
    if (fread(&keyword_count, sizeof(int), 1, fp) != 1) {
        fclose(fp);
        return -1;
    }
    
    index->keyword_count = 0;
    index->allocated_keywords = 0;
    index->keywords = NULL;
    
    for (int i = 0; i < keyword_count; i++) {
        char keyword[MAX_KEYWORD_LEN];
        int entry_count;
        
        if (fread(keyword, sizeof(char), MAX_KEYWORD_LEN, fp) != MAX_KEYWORD_LEN) {
            free_index(index);
            fclose(fp);
            return -1;
        }
        
        if (fread(&entry_count, sizeof(int), 1, fp) != 1) {
            free_index(index);
            fclose(fp);
            return -1;
        }
        
        KeywordIndex *kw_idx = find_or_create_keyword(index, keyword);
        if (kw_idx == NULL) {
            free_index(index);
            fclose(fp);
            return -1;
        }
        
        if (entry_count > 0) {
            MatchEntry *entries = malloc(entry_count * sizeof(MatchEntry));
            if (entries == NULL) {
                free_index(index);
                fclose(fp);
                return -1;
            }
            
            if (fread(entries, sizeof(MatchEntry), entry_count, fp) != (size_t)entry_count) {
                free(entries);
                free_index(index);
                fclose(fp);
                return -1;
            }
            
            kw_idx->entries = entries;
            kw_idx->entry_count = entry_count;
            kw_idx->allocated_count = entry_count;
        }
    }
    
    fclose(fp);
    return 0;
}
