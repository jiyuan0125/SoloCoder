#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <getopt.h>
#include <time.h>
#include <inttypes.h>

#include "common.h"
#include "index.h"
#include "file_search.h"
#include "query.h"

static void print_usage(const char *prog_name) {
    printf("Usage: %s [OPTIONS] <log_file> <keyword1> [keyword2] ...\n", prog_name);
    printf("\nOptions:\n");
    printf("  -h, --help              Show this help message\n");
    printf("  -c, --context N         Show N lines of context (default: 3)\n");
    printf("  -i, --index PATH        Use specified index file (default: <log_file>.idx)\n");
    printf("  -r, --regex             Treat keywords as regular expressions\n");
    printf("  -e, --escape            Escape regex special characters in keywords\n");
    printf("  -n, --no-mmap           Don't use memory-mapped file I/O\n");
    printf("  -v, --verbose           Show verbose output\n");
    printf("  -f, --force-rebuild     Force rebuild of index even if it exists\n");
    printf("\n");
    printf("Examples:\n");
    printf("  %s access.log error                    Search for 'error'\n", prog_name);
    printf("  %s access.log \"500 Internal\" \"404\"   Search for multiple keywords\n", prog_name);
    printf("  %s -c 5 access.log error               Show 5 lines of context\n", prog_name);
    printf("  %s -r access.log \"^GET.*\\.php\"       Use regex pattern\n", prog_name);
    printf("\n");
}

int main(int argc, char *argv[]) {
    int opt;
    int context_lines = CONTEXT_LINES;
    int use_mmap = 1;
    int is_regex = 0;
    int escape_regex = 0;
    int verbose = 0;
    int force_rebuild = 0;
    char *index_path = NULL;
    
    static struct option long_options[] = {
        {"help",          no_argument,       0, 'h'},
        {"context",       required_argument, 0, 'c'},
        {"index",         required_argument, 0, 'i'},
        {"regex",         no_argument,       0, 'r'},
        {"escape",        no_argument,       0, 'e'},
        {"no-mmap",       no_argument,       0, 'n'},
        {"verbose",       no_argument,       0, 'v'},
        {"force-rebuild", no_argument,       0, 'f'},
        {0, 0, 0, 0}
    };
    
    while ((opt = getopt_long(argc, argv, "hc:i:renvf", long_options, NULL)) != -1) {
        switch (opt) {
            case 'h':
                print_usage(argv[0]);
                return 0;
            case 'c':
                context_lines = atoi(optarg);
                if (context_lines < 0) context_lines = 0;
                break;
            case 'i':
                index_path = optarg;
                break;
            case 'r':
                is_regex = 1;
                break;
            case 'e':
                escape_regex = 1;
                break;
            case 'n':
                use_mmap = 0;
                break;
            case 'v':
                verbose = 1;
                break;
            case 'f':
                force_rebuild = 1;
                break;
            default:
                print_usage(argv[0]);
                return 1;
        }
    }
    
    if (optind >= argc) {
        fprintf(stderr, "Error: No log file specified\n");
        print_usage(argv[0]);
        return 1;
    }
    
    const char *log_path = argv[optind++];
    
    if (optind >= argc) {
        fprintf(stderr, "Error: No search keywords specified\n");
        print_usage(argv[0]);
        return 1;
    }
    
    SearchConfig config;
    memset(&config, 0, sizeof(config));
    config.context_lines = context_lines;
    config.escape_regex = escape_regex;
    config.verbose = verbose;
    
    while (optind < argc && config.keyword_count < MAX_KEYWORDS) {
        const char *raw_keyword = argv[optind++];
        SearchKeyword *kw = &config.keywords[config.keyword_count];
        
        if (escape_regex && !is_regex) {
            char *escaped = escape_regex_chars(raw_keyword);
            if (escaped) {
                strncpy(kw->keyword, escaped, MAX_KEYWORD_LEN - 1);
                free(escaped);
            } else {
                strncpy(kw->keyword, raw_keyword, MAX_KEYWORD_LEN - 1);
            }
        } else {
            strncpy(kw->keyword, raw_keyword, MAX_KEYWORD_LEN - 1);
        }
        kw->keyword[MAX_KEYWORD_LEN - 1] = '\0';
        kw->is_regex = is_regex;
        kw->match_count = 0;
        config.keyword_count++;
    }
    
    if (force_rebuild && index_path) {
        unlink(index_path);
    } else if (force_rebuild) {
        char default_idx[MAX_PATH_LEN];
        snprintf(default_idx, MAX_PATH_LEN, "%s.idx", log_path);
        unlink(default_idx);
    }
    
    clock_t start_time = clock();
    double index_time = 0;
    double search_time = 0;
    
    SearchSession session;
    if (init_search_session(&session, log_path, index_path, use_mmap) != 0) {
        fprintf(stderr, "Error: Cannot open log file: %s\n", log_path);
        return 1;
    }
    
    if (verbose) {
        printf("Log file: %s\n", log_path);
        printf("Index file: %s\n", session.index.header.index_path);
        printf("File size: %" PRId64 " bytes\n", (int64_t)session.file_handle.file_size);
        printf("Indexed size: %" PRId64 " bytes\n", (int64_t)session.index.header.indexed_size);
        printf("Using mmap: %s\n", use_mmap ? "yes" : "no");
        printf("\n");
    }
    
    clock_t index_start = clock();
    int updated = update_index_if_needed(&session, &config);
    clock_t index_end = clock();
    index_time = (double)(index_end - index_start) / CLOCKS_PER_SEC;
    
    if (updated > 0 && verbose) {
        printf("Index updated (scanned new content)\n");
    }
    
    if (verbose) {
        printf("Keywords to search: %d\n", config.keyword_count);
        for (int i = 0; i < config.keyword_count; i++) {
            printf("  [%d] \"%s\" (%s)\n", i + 1, 
                   config.keywords[i].keyword,
                   config.keywords[i].is_regex ? "regex" : "text");
        }
        printf("\n");
    }
    
    for (int i = 0; i < config.keyword_count; i++) {
        config.keywords[i].match_count = 0;
    }
    
    clock_t search_start = clock();
    
    FullMatchResult result;
    init_full_match_result(&result);
    
    if (search_with_context(&session, &config, &result) != 0) {
        fprintf(stderr, "Error: Search failed\n");
        close_search_session(&session);
        return 1;
    }
    
    clock_t search_end = clock();
    search_time = (double)(search_end - search_start) / CLOCKS_PER_SEC;
    
    save_index(&session.index);
    
    clock_t end_time = clock();
    double total_time = (double)(end_time - start_time) / CLOCKS_PER_SEC;
    
    printf("=");
    for (int i = 0; i < 70; i++) printf("=");
    printf("\n");
    printf("SEARCH RESULTS\n");
    printf("=");
    for (int i = 0; i < 70; i++) printf("=");
    printf("\n\n");
    
    printf("Statistics:\n");
    printf("  Log file: %s\n", log_path);
    printf("  File size: %" PRId64 " bytes\n", (int64_t)session.file_handle.file_size);
    printf("  Total matches: %d\n", result.count);
    printf("\n");
    
    printf("Keyword summary:\n");
    for (int i = 0; i < config.keyword_count; i++) {
        KeywordIndex *kw_idx = find_keyword(&session.index, config.keywords[i].keyword);
        int total_in_index = kw_idx ? kw_idx->entry_count : 0;
        printf("  \"%s\": %d matches in this search, %d total in index\n",
               config.keywords[i].keyword, 
               config.keywords[i].match_count,
               total_in_index);
    }
    printf("\n");
    
    if (result.count > 0) {
        printf("Match details:\n");
        printf("-");
        for (int i = 0; i < 70; i++) printf("-");
        printf("\n");
        
        for (int i = 0; i < result.count; i++) {
            MatchResultWithContext *item = &result.results[i];
            
            printf("\n[%d] Line %d, Offset: %" PRId64 " bytes\n",
                   i + 1, item->entry.line_number, (int64_t)item->entry.file_offset);
            printf("  Keyword: \"%s\"\n", item->entry.keyword);
            
            if (item->before_count > 0) {
                printf("\n  -- Context before --\n");
                for (int j = 0; j < item->before_count; j++) {
                    printf("     %d: %s\n", 
                           item->entry.line_number - item->before_count + j,
                           item->before_context[j]);
                }
            }
            
            printf("\n  >> Match line <<\n");
            printf("     %d: %s\n", item->entry.line_number, 
                   item->line_content ? item->line_content : "(null)");
            
            if (item->after_count > 0) {
                printf("\n  -- Context after --\n");
                for (int j = 0; j < item->after_count; j++) {
                    printf("     %d: %s\n", 
                           item->entry.line_number + j + 1,
                           item->after_context[j]);
                }
            }
            
            printf("\n");
            printf("-");
            for (int k = 0; k < 70; k++) printf("-");
            printf("\n");
        }
    } else {
        printf("No matches found.\n");
    }
    
    printf("\n");
    printf("=");
    for (int i = 0; i < 70; i++) printf("=");
    printf("\n");
    printf("PERFORMANCE\n");
    printf("=");
    for (int i = 0; i < 70; i++) printf("=");
    printf("\n\n");
    printf("  Index update time: %.3f seconds\n", index_time);
    printf("  Search time:       %.3f seconds\n", search_time);
    printf("  Total time:        %.3f seconds\n", total_time);
    printf("\n");
    
    free_full_match_result(&result);
    close_search_session(&session);
    
    return 0;
}
