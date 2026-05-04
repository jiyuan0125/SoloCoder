#include "trie.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#define INITIAL_CHILD_CAPACITY 4

static char *my_strdup(const char *s) {
    if (!s) return NULL;
    size_t len = strlen(s) + 1;
    char *dup = (char *)malloc(len);
    if (!dup) return NULL;
    memcpy(dup, s, len);
    return dup;
}

TrieNode *trie_create_node(uint32_t codepoint) {
    TrieNode *node = (TrieNode *)calloc(1, sizeof(TrieNode));
    if (!node) return NULL;
    node->codepoint = codepoint;
    node->weight = 0;
    node->word = NULL;
    node->children = NULL;
    node->child_count = 0;
    node->child_capacity = 0;
    return node;
}

void trie_destroy(TrieNode *root) {
    if (!root) return;
    for (size_t i = 0; i < root->child_count; i++) {
        trie_destroy(root->children[i]);
    }
    free(root->children);
    free(root->word);
    free(root);
}

static TrieNode *trie_find_child(TrieNode *node, uint32_t codepoint) {
    for (size_t i = 0; i < node->child_count; i++) {
        if (node->children[i]->codepoint == codepoint) {
            return node->children[i];
        }
    }
    return NULL;
}

static void trie_add_child(TrieNode *parent, TrieNode *child) {
    if (parent->child_count >= parent->child_capacity) {
        size_t new_capacity = parent->child_capacity == 0 ? 
            INITIAL_CHILD_CAPACITY : parent->child_capacity * 2;
        TrieNode **new_children = (TrieNode **)realloc(
            parent->children, new_capacity * sizeof(TrieNode *));
        if (!new_children) return;
        parent->children = new_children;
        parent->child_capacity = new_capacity;
    }
    parent->children[parent->child_count++] = child;
}

static int trie_remove_child(TrieNode *parent, TrieNode *child) {
    for (size_t i = 0; i < parent->child_count; i++) {
        if (parent->children[i] == child) {
            for (size_t j = i; j < parent->child_count - 1; j++) {
                parent->children[j] = parent->children[j + 1];
            }
            parent->child_count--;
            return 1;
        }
    }
    return 0;
}

int trie_insert(TrieNode *root, const char *word, int weight) {
    if (!root || !word || word[0] == '\0') return 0;
    
    TrieNode *current = root;
    const char *ptr = word;
    
    while (*ptr != '\0') {
        uint32_t codepoint;
        size_t len = utf8_to_codepoint(ptr, &codepoint);
        if (len == 0) return 0;
        
        TrieNode *child = trie_find_child(current, codepoint);
        if (!child) {
            child = trie_create_node(codepoint);
            if (!child) return 0;
            trie_add_child(current, child);
        }
        current = child;
        ptr += len;
    }
    
    if (current->word) {
        free(current->word);
    }
    current->word = my_strdup(word);
    current->weight = weight;
    return 1;
}

static int trie_delete_recursive(TrieNode *node, const char *ptr, int *deleted) {
    if (*ptr == '\0') {
        if (node->word) {
            free(node->word);
            node->word = NULL;
            node->weight = 0;
            *deleted = 1;
            return (node->child_count == 0);
        }
        return 0;
    }
    
    uint32_t codepoint;
    size_t len = utf8_to_codepoint(ptr, &codepoint);
    if (len == 0) return 0;
    
    TrieNode *child = trie_find_child(node, codepoint);
    if (!child) return 0;
    
    int should_delete = trie_delete_recursive(child, ptr + len, deleted);
    if (should_delete) {
        trie_remove_child(node, child);
        trie_destroy(child);
        return (node->child_count == 0 && node->word == NULL);
    }
    return 0;
}

int trie_delete(TrieNode *root, const char *word) {
    if (!root || !word || word[0] == '\0') return 0;
    int deleted = 0;
    trie_delete_recursive(root, word, &deleted);
    return deleted;
}

int trie_update_weight(TrieNode *root, const char *word, int weight) {
    if (!root || !word || word[0] == '\0') return 0;
    
    TrieNode *current = root;
    const char *ptr = word;
    
    while (*ptr != '\0') {
        uint32_t codepoint;
        size_t len = utf8_to_codepoint(ptr, &codepoint);
        if (len == 0) return 0;
        
        TrieNode *child = trie_find_child(current, codepoint);
        if (!child) return 0;
        current = child;
        ptr += len;
    }
    
    if (current->word) {
        current->weight = weight;
        return 1;
    }
    return 0;
}

int trie_search(TrieNode *root, const char *word) {
    if (!root || !word) return 0;
    
    TrieNode *current = root;
    const char *ptr = word;
    
    while (*ptr != '\0') {
        uint32_t codepoint;
        size_t len = utf8_to_codepoint(ptr, &codepoint);
        if (len == 0) return 0;
        
        TrieNode *child = trie_find_child(current, codepoint);
        if (!child) return 0;
        current = child;
        ptr += len;
    }
    
    return (current->word != NULL);
}

TrieNode *trie_find_prefix(TrieNode *root, const char *prefix) {
    if (!root || !prefix || prefix[0] == '\0') return root;
    
    TrieNode *current = root;
    const char *ptr = prefix;
    
    while (*ptr != '\0') {
        uint32_t codepoint;
        size_t len = utf8_to_codepoint(ptr, &codepoint);
        if (len == 0) return NULL;
        
        TrieNode *child = trie_find_child(current, codepoint);
        if (!child) return NULL;
        current = child;
        ptr += len;
    }
    
    return current;
}

size_t utf8_to_codepoint(const char *str, uint32_t *codepoint) {
    if (!str || !codepoint) return 0;
    
    unsigned char c = (unsigned char)*str;
    if (c <= 0x7F) {
        *codepoint = c;
        return 1;
    } else if ((c & 0xE0) == 0xC0) {
        if ((str[1] & 0xC0) != 0x80) return 0;
        *codepoint = ((c & 0x1F) << 6) | (str[1] & 0x3F);
        return 2;
    } else if ((c & 0xF0) == 0xE0) {
        if ((str[1] & 0xC0) != 0x80 || (str[2] & 0xC0) != 0x80) return 0;
        *codepoint = ((c & 0x0F) << 12) | ((str[1] & 0x3F) << 6) | (str[2] & 0x3F);
        return 3;
    } else if ((c & 0xF8) == 0xF0) {
        if ((str[1] & 0xC0) != 0x80 || (str[2] & 0xC0) != 0x80 || (str[3] & 0xC0) != 0x80) return 0;
        *codepoint = ((c & 0x07) << 18) | ((str[1] & 0x3F) << 12) | ((str[2] & 0x3F) << 6) | (str[3] & 0x3F);
        return 4;
    }
    return 0;
}

size_t codepoint_to_utf8(uint32_t codepoint, char *buf) {
    if (!buf) return 0;
    
    if (codepoint <= 0x7F) {
        buf[0] = (char)codepoint;
        return 1;
    } else if (codepoint <= 0x7FF) {
        buf[0] = (char)(0xC0 | (codepoint >> 6));
        buf[1] = (char)(0x80 | (codepoint & 0x3F));
        return 2;
    } else if (codepoint <= 0xFFFF) {
        buf[0] = (char)(0xE0 | (codepoint >> 12));
        buf[1] = (char)(0x80 | ((codepoint >> 6) & 0x3F));
        buf[2] = (char)(0x80 | (codepoint & 0x3F));
        return 3;
    } else if (codepoint <= 0x10FFFF) {
        buf[0] = (char)(0xF0 | (codepoint >> 18));
        buf[1] = (char)(0x80 | ((codepoint >> 12) & 0x3F));
        buf[2] = (char)(0x80 | ((codepoint >> 6) & 0x3F));
        buf[3] = (char)(0x80 | (codepoint & 0x3F));
        return 4;
    }
    return 0;
}

size_t utf8_strlen(const char *str) {
    if (!str) return 0;
    size_t count = 0;
    while (*str) {
        uint32_t codepoint;
        size_t len = utf8_to_codepoint(str, &codepoint);
        if (len == 0) break;
        count++;
        str += len;
    }
    return count;
}

int trie_collect_all(TrieNode *node, Suggestion *results, int *count, int max_count) {
    if (!node || !results || !count || *count >= max_count) return 0;
    
    if (node->word) {
        strncpy(results[*count].word, node->word, MAX_WORD_LENGTH - 1);
        results[*count].word[MAX_WORD_LENGTH - 1] = '\0';
        results[*count].weight = node->weight;
        results[*count].is_fuzzy = 0;
        (*count)++;
        if (*count >= max_count) return 1;
    }
    
    for (size_t i = 0; i < node->child_count; i++) {
        if (trie_collect_all(node->children[i], results, count, max_count)) {
            return 1;
        }
    }
    return 0;
}

int trie_collect_exact(TrieNode *root, const char *prefix, Suggestion *results, int max_count) {
    if (!root || !results) return 0;
    
    TrieNode *prefix_node = trie_find_prefix(root, prefix);
    if (!prefix_node) return 0;
    
    int count = 0;
    trie_collect_all(prefix_node, results, &count, max_count);
    return count;
}
