#include "file_search.h"

int open_file_for_search(FileHandle *fh, const char *path, int use_mmap) {
    memset(fh, 0, sizeof(FileHandle));
    
    fh->fd = open(path, O_RDONLY);
    if (fh->fd < 0) {
        return -1;
    }
    
    struct stat st;
    if (fstat(fh->fd, &st) != 0) {
        close(fh->fd);
        fh->fd = -1;
        return -1;
    }
    fh->file_size = st.st_size;
    fh->use_mmap = use_mmap;
    
    if (use_mmap && fh->file_size > 0) {
        fh->mmap_addr = mmap(NULL, fh->file_size, PROT_READ, MAP_PRIVATE, fh->fd, 0);
        if (fh->mmap_addr == MAP_FAILED) {
            fh->mmap_addr = NULL;
            fh->use_mmap = 0;
        }
    }
    
    return 0;
}

void close_file_handle(FileHandle *fh) {
    if (fh->mmap_addr && fh->mmap_addr != MAP_FAILED) {
        munmap(fh->mmap_addr, fh->file_size);
        fh->mmap_addr = NULL;
    }
    if (fh->fd >= 0) {
        close(fh->fd);
        fh->fd = -1;
    }
}

static int read_char_at(FileHandle *fh, off_t offset, char *c) {
    if (offset < 0 || offset >= fh->file_size) {
        return -1;
    }
    
    if (fh->use_mmap && fh->mmap_addr) {
        *c = fh->mmap_addr[offset];
        return 0;
    }
    
    if (lseek(fh->fd, offset, SEEK_SET) != offset) {
        return -1;
    }
    
    ssize_t read_bytes = read(fh->fd, c, 1);
    return (read_bytes == 1) ? 0 : -1;
}

static ssize_t read_range(FileHandle *fh, off_t offset, char *buffer, size_t size) {
    if (offset < 0 || offset >= fh->file_size) {
        return -1;
    }
    
    size_t available = (size_t)(fh->file_size - offset);
    if (size > available) {
        size = available;
    }
    
    if (fh->use_mmap && fh->mmap_addr) {
        memcpy(buffer, fh->mmap_addr + offset, size);
        return (ssize_t)size;
    }
    
    if (lseek(fh->fd, offset, SEEK_SET) != offset) {
        return -1;
    }
    
    return read(fh->fd, buffer, size);
}

off_t find_line_start(FileHandle *fh, off_t offset) {
    if (offset <= 0) {
        return 0;
    }
    
    off_t pos = offset - 1;
    char c;
    
    while (pos >= 0) {
        if (read_char_at(fh, pos, &c) != 0) {
            break;
        }
        if (c == '\n') {
            return pos + 1;
        }
        pos--;
    }
    
    return 0;
}

off_t find_next_line(FileHandle *fh, off_t offset) {
    char c;
    off_t pos = offset;
    
    while (pos < fh->file_size) {
        if (read_char_at(fh, pos, &c) != 0) {
            break;
        }
        if (c == '\n') {
            return pos + 1;
        }
        pos++;
    }
    
    return fh->file_size;
}

int get_line_by_offset(FileHandle *fh, off_t offset, char *buffer, size_t buf_size) {
    if (buf_size == 0) {
        return -1;
    }
    
    off_t line_start = find_line_start(fh, offset);
    off_t line_end = find_next_line(fh, line_start);
    
    size_t line_len = (size_t)(line_end - line_start);
    if (line_len >= buf_size) {
        line_len = buf_size - 1;
    }
    
    ssize_t read_bytes = read_range(fh, line_start, buffer, line_len);
    if (read_bytes < 0) {
        buffer[0] = '\0';
        return -1;
    }
    
    buffer[read_bytes] = '\0';
    trim_newline(buffer);
    
    return (int)strlen(buffer);
}

int simple_strstr_ignore_case(const char *haystack, const char *needle) {
    if (!haystack || !needle) return 0;
    
    size_t needle_len = strlen(needle);
    size_t haystack_len = strlen(haystack);
    
    if (needle_len == 0) return 1;
    if (haystack_len < needle_len) return 0;
    
    for (size_t i = 0; i <= haystack_len - needle_len; i++) {
        int match = 1;
        for (size_t j = 0; j < needle_len; j++) {
            if (tolower((unsigned char)haystack[i + j]) != tolower((unsigned char)needle[j])) {
                match = 0;
                break;
            }
        }
        if (match) {
            return 1;
        }
    }
    return 0;
}

static int regex_match_internal(const char *text, const char *pattern, int case_insensitive);

static int char_match(char t, char p, int case_insensitive) {
    if (case_insensitive) {
        return tolower((unsigned char)t) == tolower((unsigned char)p);
    }
    return t == p;
}

static int class_match(char t, const char *start, const char *end, int case_insensitive) {
    int negated = 0;
    const char *p = start;
    
    if (*p == '^') {
        negated = 1;
        p++;
    }
    
    int found = 0;
    while (p < end) {
        if (p + 2 < end && p[1] == '-') {
            char low = p[0];
            char high = p[2];
            if (case_insensitive) {
                low = tolower((unsigned char)low);
                high = tolower((unsigned char)high);
            }
            char tc = case_insensitive ? tolower((unsigned char)t) : t;
            if (tc >= low && tc <= high) {
                found = 1;
                break;
            }
            p += 3;
        } else {
            if (char_match(t, *p, case_insensitive)) {
                found = 1;
                break;
            }
            p++;
        }
    }
    
    return negated ? !found : found;
}

static int regex_match_internal(const char *text, const char *pattern, int case_insensitive) {
    if (*pattern == '\0') {
        return 1;
    }
    
    if (*pattern == '$' && pattern[1] == '\0') {
        return *text == '\0';
    }
    
    if (*text == '\0' && *pattern != '*' && *(pattern + 1) != '*') {
        return 0;
    }
    
    if (*pattern == '^') {
        return regex_match_internal(text, pattern + 1, case_insensitive);
    }
    
    if (*(pattern + 1) == '*') {
        int match_zero = regex_match_internal(text, pattern + 2, case_insensitive);
        if (match_zero) {
            return 1;
        }
        
        if (*pattern == '.') {
            while (*text != '\0') {
                if (regex_match_internal(text + 1, pattern + 2, case_insensitive)) {
                    return 1;
                }
                text++;
            }
        } else if (*pattern == '[') {
            const char *class_end = pattern + 1;
            int depth = 1;
            while (*class_end != '\0' && depth > 0) {
                if (*class_end == '\\') class_end++;
                else if (*class_end == ']') depth--;
                class_end++;
            }
            while (*text != '\0' && class_match(*text, pattern + 1, class_end - 1, case_insensitive)) {
                if (regex_match_internal(text + 1, class_end, case_insensitive)) {
                    return 1;
                }
                text++;
            }
        } else {
            while (*text != '\0' && char_match(*text, *pattern, case_insensitive)) {
                if (regex_match_internal(text + 1, pattern + 2, case_insensitive)) {
                    return 1;
                }
                text++;
            }
        }
        return 0;
    }
    
    if (*(pattern + 1) == '+') {
        if (*text == '\0') return 0;
        
        int matched = 0;
        if (*pattern == '.') {
            matched = 1;
        } else if (*pattern == '[') {
            const char *class_end = pattern + 1;
            int depth = 1;
            while (*class_end != '\0' && depth > 0) {
                if (*class_end == '\\') class_end++;
                else if (*class_end == ']') depth--;
                class_end++;
            }
            matched = class_match(*text, pattern + 1, class_end - 1, case_insensitive);
        } else {
            matched = char_match(*text, *pattern, case_insensitive);
        }
        
        if (!matched) return 0;
        return regex_match_internal(text + 1, pattern, case_insensitive) ||
               regex_match_internal(text + 1, pattern + 2, case_insensitive);
    }
    
    if (*pattern == '\\' && pattern[1] != '\0') {
        if (char_match(*text, pattern[1], case_insensitive)) {
            return regex_match_internal(text + 1, pattern + 2, case_insensitive);
        }
        return 0;
    }
    
    if (*pattern == '.') {
        return regex_match_internal(text + 1, pattern + 1, case_insensitive);
    }
    
    if (*pattern == '[') {
        const char *class_end = pattern + 1;
        int depth = 1;
        while (*class_end != '\0' && depth > 0) {
            if (*class_end == '\\') class_end++;
            else if (*class_end == ']') depth--;
            class_end++;
        }
        if (class_match(*text, pattern + 1, class_end - 1, case_insensitive)) {
            return regex_match_internal(text + 1, class_end, case_insensitive);
        }
        return 0;
    }
    
    if (char_match(*text, *pattern, case_insensitive)) {
        return regex_match_internal(text + 1, pattern + 1, case_insensitive);
    }
    
    return 0;
}

int simple_regex_match(const char *text, const char *pattern, int case_insensitive) {
    if (!text || !pattern) return 0;
    
    if (pattern[0] == '^') {
        return regex_match_internal(text, pattern, case_insensitive);
    }
    
    while (*text != '\0') {
        if (regex_match_internal(text, pattern, case_insensitive)) {
            return 1;
        }
        text++;
    }
    return 0;
}

int search_file_chunk(FileHandle *fh, off_t start_offset, off_t end_offset,
                      SearchIndex *index, SearchConfig *config,
                      int start_line_number, off_t *actual_end) {
    if (config->keyword_count == 0) {
        return 0;
    }
    
    char buffer[BUFFER_SIZE + 1];
    off_t current_offset = find_line_start(fh, start_offset);
    int line_number = start_line_number;
    
    if (actual_end) {
        *actual_end = current_offset;
    }
    
    while (current_offset < end_offset && current_offset < fh->file_size) {
        off_t line_end = find_next_line(fh, current_offset);
        size_t line_len = (size_t)(line_end - current_offset);
        
        if (line_len > 0) {
            if (line_len >= BUFFER_SIZE) {
                line_len = BUFFER_SIZE - 1;
            }
            
            ssize_t read_bytes = read_range(fh, current_offset, buffer, line_len);
            if (read_bytes < 0) {
                current_offset = line_end;
                line_number++;
                continue;
            }
            buffer[read_bytes] = '\0';
            trim_newline(buffer);
            
            for (int i = 0; i < config->keyword_count; i++) {
                SearchKeyword *kw = &config->keywords[i];
                int matched = 0;
                
                if (kw->is_regex) {
                    matched = simple_regex_match(buffer, kw->keyword, 0);
                } else {
                    matched = simple_strstr_ignore_case(buffer, kw->keyword);
                }
                
                if (matched) {
                    MatchEntry entry;
                    memset(&entry, 0, sizeof(entry));
                    strncpy(entry.keyword, kw->keyword, MAX_KEYWORD_LEN - 1);
                    entry.keyword[MAX_KEYWORD_LEN - 1] = '\0';
                    entry.line_number = line_number + 1;
                    entry.file_offset = current_offset;
                    entry.line_length = strlen(buffer);
                    
                    add_match_to_index(index, kw->keyword, &entry);
                    kw->match_count++;
                }
            }
        }
        
        if (actual_end) {
            *actual_end = line_end;
        }
        current_offset = line_end;
        line_number++;
    }
    
    return 0;
}

char **alloc_context_lines(int count) {
    if (count <= 0) return NULL;
    char **lines = malloc(count * sizeof(char *));
    if (lines == NULL) return NULL;
    for (int i = 0; i < count; i++) {
        lines[i] = NULL;
    }
    return lines;
}

void free_context_lines(char **lines, int count) {
    if (lines == NULL) return;
    for (int i = 0; i < count; i++) {
        if (lines[i]) {
            free(lines[i]);
            lines[i] = NULL;
        }
    }
    free(lines);
}

int get_context_lines(FileHandle *fh, const MatchEntry *match, int context_lines,
                      char ***before_lines, int *before_count,
                      char ***after_lines, int *after_count) {
    *before_lines = NULL;
    *before_count = 0;
    *after_lines = NULL;
    *after_count = 0;
    
    if (context_lines <= 0) {
        return 0;
    }
    
    char line_buffer[BUFFER_SIZE];
    off_t current_offset = match->file_offset;
    
    char **before = alloc_context_lines(context_lines);
    if (before == NULL) {
        return -1;
    }
    
    off_t offsets_before[CONTEXT_LINES * 2];
    int count_before = 0;
    
    off_t offset = find_line_start(fh, current_offset);
    for (int i = 0; i < context_lines && offset > 0; i++) {
        off_t scan_pos = offset - 1;
        off_t prev_line_start = 0;
        int found_newline = 0;
        
        while (scan_pos >= 0) {
            char c;
            if (read_char_at(fh, scan_pos, &c) != 0) {
                break;
            }
            if (c == '\n') {
                prev_line_start = scan_pos + 1;
                found_newline = 1;
                break;
            }
            scan_pos--;
        }
        
        if (found_newline && prev_line_start < offset) {
            offsets_before[count_before++] = prev_line_start;
            offset = prev_line_start;
        } else if (!found_newline && offset > 0) {
            prev_line_start = 0;
            if (prev_line_start < offset) {
                offsets_before[count_before++] = prev_line_start;
            }
            break;
        } else {
            break;
        }
    }
    
    for (int i = count_before - 1, j = 0; i >= 0 && j < context_lines; i--, j++) {
        if (get_line_by_offset(fh, offsets_before[i], line_buffer, BUFFER_SIZE) >= 0) {
            before[j] = safe_strdup(line_buffer);
        }
    }
    *before_lines = before;
    *before_count = count_before;
    
    char **after = alloc_context_lines(context_lines);
    if (after == NULL) {
        free_context_lines(before, count_before);
        *before_lines = NULL;
        return -1;
    }
    
    current_offset = match->file_offset;
    int count_after = 0;
    
    for (int i = 0; i < context_lines; i++) {
        off_t next_line = find_next_line(fh, current_offset);
        if (next_line >= fh->file_size) {
            break;
        }
        
        if (get_line_by_offset(fh, next_line, line_buffer, BUFFER_SIZE) >= 0) {
            after[count_after] = safe_strdup(line_buffer);
            count_after++;
        }
        current_offset = next_line;
    }
    
    *after_lines = after;
    *after_count = count_after;
    
    return 0;
}
