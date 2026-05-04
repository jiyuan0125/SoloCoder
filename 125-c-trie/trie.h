#ifndef TRIE_H
#define TRIE_H

#include <stddef.h>
#include <stdint.h>

#define MAX_SUGGESTIONS 10
#define MAX_WORD_LENGTH 256

typedef struct TrieNode {
    uint32_t codepoint;
    int weight;
    char *word;
    struct TrieNode **children;
    size_t child_count;
    size_t child_capacity;
} TrieNode;

typedef struct {
    char word[MAX_WORD_LENGTH];
    int weight;
    int is_fuzzy;
} Suggestion;

TrieNode *trie_create_node(uint32_t codepoint);
void trie_destroy(TrieNode *root);
int trie_insert(TrieNode *root, const char *word, int weight);
int trie_delete(TrieNode *root, const char *word);
int trie_update_weight(TrieNode *root, const char *word, int weight);
int trie_search(TrieNode *root, const char *word);
TrieNode *trie_find_prefix(TrieNode *root, const char *prefix);

size_t utf8_to_codepoint(const char *str, uint32_t *codepoint);
size_t codepoint_to_utf8(uint32_t codepoint, char *buf);
size_t utf8_strlen(const char *str);

int trie_collect_all(TrieNode *node, Suggestion *results, int *count, int max_count);
int trie_collect_exact(TrieNode *root, const char *prefix, Suggestion *results, int max_count);

#endif
