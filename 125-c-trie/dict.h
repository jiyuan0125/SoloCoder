#ifndef DICT_H
#define DICT_H

#include "trie.h"
#include "query.h"

#define PINYIN_MAP_SIZE 256

typedef struct {
    TrieNode *root;
    TrieNode *pinyin_root;
} Dictionary;

typedef struct {
    char hanzi[MAX_WORD_LENGTH];
    char pinyin[MAX_WORD_LENGTH];
} PinyinMapEntry;

Dictionary *dict_create(void);
void dict_destroy(Dictionary *dict);

int dict_add_word(Dictionary *dict, const char *word, int weight);
int dict_remove_word(Dictionary *dict, const char *word);
int dict_update_weight(Dictionary *dict, const char *word, int weight);
int dict_has_word(Dictionary *dict, const char *word);

int dict_add_pinyin_mapping(Dictionary *dict, const char *hanzi, const char *pinyin);
int dict_query_pinyin(Dictionary *dict, const char *pinyin_prefix, Suggestion *results, int max_count);

int dict_query(Dictionary *dict, const char *input, Suggestion *results, int max_count);

#endif
