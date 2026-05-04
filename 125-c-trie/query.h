#ifndef QUERY_H
#define QUERY_H

#include "trie.h"

#define MAX_FUZZY_DISTANCE 2
#define MAX_SUBSTRING_MATCHES 50

typedef struct {
    char word[MAX_WORD_LENGTH];
    int weight;
    int distance;
} FuzzyMatch;

typedef struct {
    char word[MAX_WORD_LENGTH];
    int weight;
} WordEntry;

int suggestions_sort_by_weight(const void *a, const void *b);
int suggestions_sort_by_distance_then_weight(const void *a, const void *b);

int query_exact_prefix(TrieNode *root, const char *prefix, Suggestion *results, int max_count);
int query_hot_words(TrieNode *root, Suggestion *results, int max_count);

int levenshtein_distance(const char *s1, const char *s2);
int query_fuzzy(TrieNode *root, const char *input, Suggestion *results, int max_count, int max_distance);

int query_substring(TrieNode *root, const char *substring, Suggestion *results, int max_count);

int query_combined(TrieNode *root, const char *input, Suggestion *results, int max_count);

#endif
