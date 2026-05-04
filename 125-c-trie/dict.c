#include "dict.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <ctype.h>

typedef struct PinyinNode {
    char pinyin[MAX_WORD_LENGTH];
    char hanzi[MAX_WORD_LENGTH];
    int weight;
    struct PinyinNode *next;
} PinyinNode;

static PinyinNode *g_pinyin_map = NULL;

Dictionary *dict_create(void) {
    Dictionary *dict = (Dictionary *)calloc(1, sizeof(Dictionary));
    if (!dict) return NULL;
    
    dict->root = trie_create_node(0);
    if (!dict->root) {
        free(dict);
        return NULL;
    }
    
    dict->pinyin_root = trie_create_node(0);
    if (!dict->pinyin_root) {
        trie_destroy(dict->root);
        free(dict);
        return NULL;
    }
    
    return dict;
}

void dict_destroy(Dictionary *dict) {
    if (!dict) return;
    
    if (dict->root) {
        trie_destroy(dict->root);
    }
    if (dict->pinyin_root) {
        trie_destroy(dict->pinyin_root);
    }
    
    while (g_pinyin_map) {
        PinyinNode *temp = g_pinyin_map;
        g_pinyin_map = g_pinyin_map->next;
        free(temp);
    }
    
    free(dict);
}

int dict_add_word(Dictionary *dict, const char *word, int weight) {
    if (!dict || !word || word[0] == '\0') return 0;
    return trie_insert(dict->root, word, weight);
}

int dict_remove_word(Dictionary *dict, const char *word) {
    if (!dict || !word || word[0] == '\0') return 0;
    return trie_delete(dict->root, word);
}

int dict_update_weight(Dictionary *dict, const char *word, int weight) {
    if (!dict || !word || word[0] == '\0') return 0;
    return trie_update_weight(dict->root, word, weight);
}

int dict_has_word(Dictionary *dict, const char *word) {
    if (!dict || !word) return 0;
    return trie_search(dict->root, word);
}

int dict_add_pinyin_mapping(Dictionary *dict, const char *hanzi, const char *pinyin) {
    if (!dict || !hanzi || !pinyin || hanzi[0] == '\0' || pinyin[0] == '\0') return 0;
    
    int weight = 0;
    TrieNode *node = trie_find_prefix(dict->root, hanzi);
    if (node && node->word && strcmp(node->word, hanzi) == 0) {
        weight = node->weight;
    }
    
    PinyinNode *new_node = (PinyinNode *)calloc(1, sizeof(PinyinNode));
    if (!new_node) return 0;
    
    strncpy(new_node->hanzi, hanzi, MAX_WORD_LENGTH - 1);
    new_node->hanzi[MAX_WORD_LENGTH - 1] = '\0';
    strncpy(new_node->pinyin, pinyin, MAX_WORD_LENGTH - 1);
    new_node->pinyin[MAX_WORD_LENGTH - 1] = '\0';
    new_node->weight = weight;
    new_node->next = g_pinyin_map;
    g_pinyin_map = new_node;
    
    return trie_insert(dict->pinyin_root, pinyin, weight);
}

static int is_all_english(const char *str) {
    if (!str) return 0;
    for (const char *p = str; *p; p++) {
        if (!isalpha((unsigned char)*p)) {
            return 0;
        }
    }
    return 1;
}

int dict_query_pinyin(Dictionary *dict, const char *pinyin_prefix, Suggestion *results, int max_count) {
    if (!dict || !pinyin_prefix || pinyin_prefix[0] == '\0' || !results || max_count <= 0) return 0;
    
    Suggestion pinyin_matches[100];
    int pinyin_count = query_exact_prefix(dict->pinyin_root, pinyin_prefix, pinyin_matches, 100);
    
    if (pinyin_count == 0) return 0;
    
    int result_count = 0;
    
    for (int i = 0; i < pinyin_count && result_count < max_count; i++) {
        const char *matched_pinyin = pinyin_matches[i].word;
        
        PinyinNode *current = g_pinyin_map;
        while (current && result_count < max_count) {
            if (strncmp(current->pinyin, matched_pinyin, strlen(matched_pinyin)) == 0) {
                int found = 0;
                for (int j = 0; j < result_count; j++) {
                    if (strcmp(results[j].word, current->hanzi) == 0) {
                        found = 1;
                        break;
                    }
                }
                if (!found) {
                    strncpy(results[result_count].word, current->hanzi, MAX_WORD_LENGTH - 1);
                    results[result_count].word[MAX_WORD_LENGTH - 1] = '\0';
                    results[result_count].weight = current->weight;
                    results[result_count].is_fuzzy = 0;
                    result_count++;
                }
            }
            current = current->next;
        }
    }
    
    qsort(results, result_count, sizeof(Suggestion), suggestions_sort_by_weight);
    
    return result_count;
}

int dict_query(Dictionary *dict, const char *input, Suggestion *results, int max_count) {
    if (!dict || !results || max_count <= 0) return 0;
    
    Suggestion combined_results[100];
    int combined_count = 0;
    
    if (is_all_english(input) && input && input[0] != '\0') {
        combined_count = dict_query_pinyin(dict, input, combined_results, 50);
        
        if (combined_count < 50) {
            Suggestion temp_results[100];
            int temp_count = query_combined(dict->root, input, temp_results, 50);
            
            for (int i = 0; i < temp_count && combined_count < 100; i++) {
                int duplicate = 0;
                for (int j = 0; j < combined_count; j++) {
                    if (strcmp(combined_results[j].word, temp_results[i].word) == 0) {
                        duplicate = 1;
                        break;
                    }
                }
                if (!duplicate) {
                    combined_results[combined_count++] = temp_results[i];
                }
            }
        }
    } else {
        combined_count = query_combined(dict->root, input, combined_results, 100);
    }
    
    int actual_count = (combined_count < max_count) ? combined_count : max_count;
    for (int i = 0; i < actual_count; i++) {
        results[i] = combined_results[i];
    }
    
    return actual_count;
}
