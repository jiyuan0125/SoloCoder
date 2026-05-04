#include "query.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

static WordEntry *g_all_words = NULL;
static size_t g_all_words_count = 0;
static size_t g_all_words_capacity = 0;

static void collect_all_words_callback(const char *word, int weight) {
    if (g_all_words_count >= g_all_words_capacity) {
        size_t new_capacity = g_all_words_capacity == 0 ? 1024 : g_all_words_capacity * 2;
        WordEntry *new_entries = (WordEntry *)realloc(
            g_all_words, new_capacity * sizeof(WordEntry));
        if (!new_entries) return;
        g_all_words = new_entries;
        g_all_words_capacity = new_capacity;
    }
    strncpy(g_all_words[g_all_words_count].word, word, MAX_WORD_LENGTH - 1);
    g_all_words[g_all_words_count].word[MAX_WORD_LENGTH - 1] = '\0';
    g_all_words[g_all_words_count].weight = weight;
    g_all_words_count++;
}

static void trie_collect_with_callback(TrieNode *node, void (*callback)(const char *, int)) {
    if (!node || !callback) return;
    if (node->word) {
        callback(node->word, node->weight);
    }
    for (size_t i = 0; i < node->child_count; i++) {
        trie_collect_with_callback(node->children[i], callback);
    }
}

static void collect_all_words(TrieNode *root) {
    g_all_words_count = 0;
    trie_collect_with_callback(root, collect_all_words_callback);
}

static void free_all_words(void) {
    free(g_all_words);
    g_all_words = NULL;
    g_all_words_count = 0;
    g_all_words_capacity = 0;
}

int suggestions_sort_by_weight(const void *a, const void *b) {
    const Suggestion *sa = (const Suggestion *)a;
    const Suggestion *sb = (const Suggestion *)b;
    return sb->weight - sa->weight;
}

int suggestions_sort_by_distance_then_weight(const void *a, const void *b) {
    const FuzzyMatch *fa = (const FuzzyMatch *)a;
    const FuzzyMatch *fb = (const FuzzyMatch *)b;
    if (fa->distance != fb->distance) {
        return fa->distance - fb->distance;
    }
    return fb->weight - fa->weight;
}

int query_exact_prefix(TrieNode *root, const char *prefix, Suggestion *results, int max_count) {
    if (!root || !results || max_count <= 0) return 0;
    
    if (!prefix || prefix[0] == '\0') {
        return query_hot_words(root, results, max_count);
    }
    
    Suggestion temp_results[100];
    int count = trie_collect_exact(root, prefix, temp_results, 100);
    
    if (count == 0) return 0;
    
    qsort(temp_results, count, sizeof(Suggestion), suggestions_sort_by_weight);
    
    int actual_count = (count < max_count) ? count : max_count;
    for (int i = 0; i < actual_count; i++) {
        results[i] = temp_results[i];
    }
    
    return actual_count;
}

int query_hot_words(TrieNode *root, Suggestion *results, int max_count) {
    if (!root || !results || max_count <= 0) return 0;
    
    collect_all_words(root);
    
    if (g_all_words_count == 0) {
        free_all_words();
        return 0;
    }
    
    qsort(g_all_words, g_all_words_count, sizeof(WordEntry), 
          (int (*)(const void *, const void *))suggestions_sort_by_weight);
    
    int actual_count = ((int)g_all_words_count < max_count) ? (int)g_all_words_count : max_count;
    for (int i = 0; i < actual_count; i++) {
        strncpy(results[i].word, g_all_words[i].word, MAX_WORD_LENGTH - 1);
        results[i].word[MAX_WORD_LENGTH - 1] = '\0';
        results[i].weight = g_all_words[i].weight;
        results[i].is_fuzzy = 0;
    }
    
    free_all_words();
    return actual_count;
}

static int min3(int a, int b, int c) {
    int m = a < b ? a : b;
    return m < c ? m : c;
}

int levenshtein_distance(const char *s1, const char *s2) {
    if (!s1 || !s2) return -1;
    
    uint32_t *codepoints1 = NULL;
    uint32_t *codepoints2 = NULL;
    int len1 = 0, len2 = 0;
    int i, j;
    
    const char *ptr = s1;
    while (*ptr) {
        uint32_t cp;
        size_t l = utf8_to_codepoint(ptr, &cp);
        if (l == 0) break;
        codepoints1 = (uint32_t *)realloc(codepoints1, (len1 + 1) * sizeof(uint32_t));
        codepoints1[len1++] = cp;
        ptr += l;
    }
    
    ptr = s2;
    while (*ptr) {
        uint32_t cp;
        size_t l = utf8_to_codepoint(ptr, &cp);
        if (l == 0) break;
        codepoints2 = (uint32_t *)realloc(codepoints2, (len2 + 1) * sizeof(uint32_t));
        codepoints2[len2++] = cp;
        ptr += l;
    }
    
    int **matrix = (int **)malloc((len1 + 1) * sizeof(int *));
    for (i = 0; i <= len1; i++) {
        matrix[i] = (int *)malloc((len2 + 1) * sizeof(int));
        matrix[i][0] = i;
    }
    for (j = 0; j <= len2; j++) {
        matrix[0][j] = j;
    }
    
    for (i = 1; i <= len1; i++) {
        for (j = 1; j <= len2; j++) {
            int cost = (codepoints1[i-1] == codepoints2[j-1]) ? 0 : 1;
            matrix[i][j] = min3(
                matrix[i-1][j] + 1,
                matrix[i][j-1] + 1,
                matrix[i-1][j-1] + cost
            );
        }
    }
    
    int result = matrix[len1][len2];
    
    for (i = 0; i <= len1; i++) {
        free(matrix[i]);
    }
    free(matrix);
    free(codepoints1);
    free(codepoints2);
    
    return result;
}

int query_fuzzy(TrieNode *root, const char *input, Suggestion *results, int max_count, int max_distance) {
    if (!root || !input || !results || max_count <= 0) return 0;
    
    collect_all_words(root);
    
    if (g_all_words_count == 0) {
        free_all_words();
        return 0;
    }
    
    FuzzyMatch *matches = (FuzzyMatch *)malloc(g_all_words_count * sizeof(FuzzyMatch));
    int match_count = 0;
    
    for (size_t i = 0; i < g_all_words_count; i++) {
        int dist = levenshtein_distance(input, g_all_words[i].word);
        if (dist >= 0 && dist <= max_distance && dist > 0) {
            strncpy(matches[match_count].word, g_all_words[i].word, MAX_WORD_LENGTH - 1);
            matches[match_count].word[MAX_WORD_LENGTH - 1] = '\0';
            matches[match_count].weight = g_all_words[i].weight;
            matches[match_count].distance = dist;
            match_count++;
        }
    }
    
    free_all_words();
    
    if (match_count == 0) {
        free(matches);
        return 0;
    }
    
    qsort(matches, match_count, sizeof(FuzzyMatch), suggestions_sort_by_distance_then_weight);
    
    int actual_count = (match_count < max_count) ? match_count : max_count;
    for (int i = 0; i < actual_count; i++) {
        strncpy(results[i].word, matches[i].word, MAX_WORD_LENGTH - 1);
        results[i].word[MAX_WORD_LENGTH - 1] = '\0';
        results[i].weight = matches[i].weight;
        results[i].is_fuzzy = 1;
    }
    
    free(matches);
    return actual_count;
}

static int utf8_strstr(const char *haystack, const char *needle) {
    if (!haystack || !needle || needle[0] == '\0') return 0;
    
    size_t needle_len = strlen(needle);
    size_t haystack_len = strlen(haystack);
    
    if (needle_len > haystack_len) return 0;
    
    for (size_t i = 0; i <= haystack_len - needle_len; i++) {
        int found = 1;
        for (size_t j = 0; j < needle_len; j++) {
            if (haystack[i + j] != needle[j]) {
                found = 0;
                break;
            }
        }
        if (found) return 1;
    }
    return 0;
}

int query_substring(TrieNode *root, const char *substring, Suggestion *results, int max_count) {
    if (!root || !substring || substring[0] == '\0' || !results || max_count <= 0) return 0;
    
    collect_all_words(root);
    
    if (g_all_words_count == 0) {
        free_all_words();
        return 0;
    }
    
    Suggestion *matches = (Suggestion *)malloc(g_all_words_count * sizeof(Suggestion));
    int match_count = 0;
    
    for (size_t i = 0; i < g_all_words_count; i++) {
        if (utf8_strstr(g_all_words[i].word, substring)) {
            strncpy(matches[match_count].word, g_all_words[i].word, MAX_WORD_LENGTH - 1);
            matches[match_count].word[MAX_WORD_LENGTH - 1] = '\0';
            matches[match_count].weight = g_all_words[i].weight;
            matches[match_count].is_fuzzy = 0;
            match_count++;
        }
    }
    
    free_all_words();
    
    if (match_count == 0) {
        free(matches);
        return 0;
    }
    
    qsort(matches, match_count, sizeof(Suggestion), suggestions_sort_by_weight);
    
    int actual_count = (match_count < max_count) ? match_count : max_count;
    for (int i = 0; i < actual_count; i++) {
        results[i] = matches[i];
    }
    
    free(matches);
    return actual_count;
}

static int is_duplicate(Suggestion *results, int count, const char *word) {
    for (int i = 0; i < count; i++) {
        if (strcmp(results[i].word, word) == 0) {
            return 1;
        }
    }
    return 0;
}

int query_combined(TrieNode *root, const char *input, Suggestion *results, int max_count) {
    if (!root || !results || max_count <= 0) return 0;
    
    Suggestion exact_matches[100];
    Suggestion substring_matches[100];
    Suggestion fuzzy_matches[100];
    
    int exact_count = 0;
    int substring_count = 0;
    int fuzzy_count = 0;
    
    if (!input || input[0] == '\0') {
        return query_hot_words(root, results, max_count);
    }
    
    exact_count = query_exact_prefix(root, input, exact_matches, 100);
    substring_count = query_substring(root, input, substring_matches, 100);
    fuzzy_count = query_fuzzy(root, input, fuzzy_matches, 100, MAX_FUZZY_DISTANCE);
    
    int total_count = 0;
    
    for (int i = 0; i < exact_count && total_count < max_count; i++) {
        if (!is_duplicate(results, total_count, exact_matches[i].word)) {
            results[total_count] = exact_matches[i];
            results[total_count].is_fuzzy = 0;
            total_count++;
        }
    }
    
    for (int i = 0; i < substring_count && total_count < max_count; i++) {
        if (!is_duplicate(results, total_count, substring_matches[i].word)) {
            results[total_count] = substring_matches[i];
            results[total_count].is_fuzzy = 0;
            total_count++;
        }
    }
    
    for (int i = 0; i < fuzzy_count && total_count < max_count; i++) {
        if (!is_duplicate(results, total_count, fuzzy_matches[i].word)) {
            results[total_count] = fuzzy_matches[i];
            results[total_count].is_fuzzy = 1;
            total_count++;
        }
    }
    
    return total_count;
}
