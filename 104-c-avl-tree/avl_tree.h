#ifndef AVL_TREE_H
#define AVL_TREE_H

#include <stdint.h>
#include <stdlib.h>

typedef struct AVLNode {
    uint64_t player_id;
    int32_t score;
    uint64_t timestamp;
    int height;
    int size;
    struct AVLNode *left;
    struct AVLNode *right;
} AVLNode;

typedef struct AVLTree {
    AVLNode *root;
    int max_capacity;
    int current_size;
} AVLTree;

AVLTree* avl_create(int max_capacity);
void avl_destroy(AVLTree *tree);

int avl_insert(AVLTree *tree, uint64_t player_id, int32_t score, uint64_t timestamp);
int avl_remove(AVLTree *tree, uint64_t player_id, int32_t score, uint64_t timestamp);

int avl_get_rank(AVLTree *tree, uint64_t player_id, int32_t score, uint64_t timestamp);
int avl_get_by_rank(AVLTree *tree, int rank, uint64_t *player_id, int32_t *score, uint64_t *timestamp);
int avl_get_range(AVLTree *tree, int start_rank, int end_rank, 
                  uint64_t *player_ids, int32_t *scores, uint64_t *timestamps, int max_count);

int avl_count_in_score_range(AVLTree *tree, int32_t min_score, int32_t max_score);

int avl_size(AVLTree *tree);
void avl_clear(AVLTree *tree);

#endif
