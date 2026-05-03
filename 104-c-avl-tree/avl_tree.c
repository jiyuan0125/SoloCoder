#include "avl_tree.h"
#include <string.h>

static int max(int a, int b) {
    return (a > b) ? a : b;
}

static int height(AVLNode *node) {
    if (node == NULL) return 0;
    return node->height;
}

static int size(AVLNode *node) {
    if (node == NULL) return 0;
    return node->size;
}

static void update_height_and_size(AVLNode *node) {
    if (node == NULL) return;
    node->height = 1 + max(height(node->left), height(node->right));
    node->size = 1 + size(node->left) + size(node->right);
}

static int get_balance(AVLNode *node) {
    if (node == NULL) return 0;
    return height(node->left) - height(node->right);
}

static int compare(int32_t score1, uint64_t ts1, int32_t score2, uint64_t ts2) {
    if (score1 != score2) {
        return (score1 > score2) ? -1 : 1;
    }
    if (ts1 != ts2) {
        return (ts1 < ts2) ? -1 : 1;
    }
    return 0;
}

static AVLNode* right_rotate(AVLNode *y) {
    AVLNode *x = y->left;
    AVLNode *T2 = x->right;

    x->right = y;
    y->left = T2;

    update_height_and_size(y);
    update_height_and_size(x);

    return x;
}

static AVLNode* left_rotate(AVLNode *x) {
    AVLNode *y = x->right;
    AVLNode *T2 = y->left;

    y->left = x;
    x->right = T2;

    update_height_and_size(x);
    update_height_and_size(y);

    return y;
}

static AVLNode* create_node(uint64_t player_id, int32_t score, uint64_t timestamp) {
    AVLNode *node = (AVLNode*)malloc(sizeof(AVLNode));
    if (node == NULL) return NULL;
    node->player_id = player_id;
    node->score = score;
    node->timestamp = timestamp;
    node->left = NULL;
    node->right = NULL;
    node->height = 1;
    node->size = 1;
    return node;
}

static AVLNode* avl_insert_node(AVLNode *node, uint64_t player_id, int32_t score, uint64_t timestamp, int *inserted) {
    if (node == NULL) {
        *inserted = 1;
        return create_node(player_id, score, timestamp);
    }

    int cmp = compare(score, timestamp, node->score, node->timestamp);

    if (cmp < 0) {
        node->left = avl_insert_node(node->left, player_id, score, timestamp, inserted);
    } else if (cmp > 0) {
        node->right = avl_insert_node(node->right, player_id, score, timestamp, inserted);
    } else {
        *inserted = 0;
        return node;
    }

    update_height_and_size(node);

    int balance = get_balance(node);

    if (balance > 1) {
        int left_cmp = compare(score, timestamp, node->left->score, node->left->timestamp);
        if (left_cmp < 0) {
            return right_rotate(node);
        }
        if (left_cmp > 0) {
            node->left = left_rotate(node->left);
            return right_rotate(node);
        }
    }

    if (balance < -1) {
        int right_cmp = compare(score, timestamp, node->right->score, node->right->timestamp);
        if (right_cmp > 0) {
            return left_rotate(node);
        }
        if (right_cmp < 0) {
            node->right = right_rotate(node->right);
            return left_rotate(node);
        }
    }

    return node;
}

static AVLNode* get_min_node(AVLNode *node) {
    AVLNode *current = node;
    while (current->left != NULL)
        current = current->left;
    return current;
}

static AVLNode* get_max_node(AVLNode *node) {
    AVLNode *current = node;
    while (current->right != NULL)
        current = current->right;
    return current;
}

static AVLNode* avl_remove_node(AVLNode *root, uint64_t player_id, int32_t score, uint64_t timestamp, int *removed) {
    if (root == NULL) {
        *removed = 0;
        return NULL;
    }

    int cmp = compare(score, timestamp, root->score, root->timestamp);

    if (cmp < 0) {
        root->left = avl_remove_node(root->left, player_id, score, timestamp, removed);
    } else if (cmp > 0) {
        root->right = avl_remove_node(root->right, player_id, score, timestamp, removed);
    } else {
        if (root->player_id != player_id) {
            root->left = avl_remove_node(root->left, player_id, score, timestamp, removed);
            if (*removed == 0) {
                root->right = avl_remove_node(root->right, player_id, score, timestamp, removed);
            }
            if (*removed == 0) {
                return root;
            }
        } else {
            *removed = 1;
            
            if ((root->left == NULL) || (root->right == NULL)) {
                AVLNode *temp = root->left ? root->left : root->right;

                if (temp == NULL) {
                    temp = root;
                    root = NULL;
                } else {
                    *root = *temp;
                }
                free(temp);
            } else {
                AVLNode *temp = get_min_node(root->right);

                root->player_id = temp->player_id;
                root->score = temp->score;
                root->timestamp = temp->timestamp;

                int dummy = 0;
                root->right = avl_remove_node(root->right, temp->player_id, temp->score, temp->timestamp, &dummy);
            }
        }
    }

    if (root == NULL) return root;

    update_height_and_size(root);

    int balance = get_balance(root);

    if (balance > 1) {
        int left_balance = get_balance(root->left);
        if (left_balance >= 0) {
            return right_rotate(root);
        }
        if (left_balance < 0) {
            root->left = left_rotate(root->left);
            return right_rotate(root);
        }
    }

    if (balance < -1) {
        int right_balance = get_balance(root->right);
        if (right_balance <= 0) {
            return left_rotate(root);
        }
        if (right_balance > 0) {
            root->right = right_rotate(root->right);
            return left_rotate(root);
        }
    }

    return root;
}

static int avl_get_rank_node(AVLNode *node, uint64_t player_id, int32_t score, uint64_t timestamp) {
    if (node == NULL) return -1;

    int cmp = compare(score, timestamp, node->score, node->timestamp);

    if (cmp < 0) {
        return avl_get_rank_node(node->left, player_id, score, timestamp);
    } else if (cmp > 0) {
        int left_size = size(node->left);
        int right_rank = avl_get_rank_node(node->right, player_id, score, timestamp);
        if (right_rank == -1) return -1;
        return left_size + 1 + right_rank;
    } else {
        if (node->player_id == player_id) {
            return size(node->left) + 1;
        }
        int left_rank = avl_get_rank_node(node->left, player_id, score, timestamp);
        if (left_rank != -1) return left_rank;
        int right_rank = avl_get_rank_node(node->right, player_id, score, timestamp);
        if (right_rank == -1) return -1;
        return size(node->left) + 1 + right_rank;
    }
}

static int avl_get_by_rank_node(AVLNode *node, int rank, uint64_t *player_id, int32_t *score, uint64_t *timestamp) {
    if (node == NULL) return 0;

    int left_size = size(node->left);

    if (rank <= left_size) {
        return avl_get_by_rank_node(node->left, rank, player_id, score, timestamp);
    } else if (rank == left_size + 1) {
        *player_id = node->player_id;
        *score = node->score;
        *timestamp = node->timestamp;
        return 1;
    } else {
        return avl_get_by_rank_node(node->right, rank - left_size - 1, player_id, score, timestamp);
    }
}

static int avl_get_range_optimized(AVLNode *node, int start_rank, int end_rank,
                                     int prefix_count,
                                     uint64_t *player_ids, int32_t *scores, uint64_t *timestamps,
                                     int max_count, int *filled_count) {
    if (node == NULL || *filled_count >= max_count) return 0;

    int left_size = size(node->left);
    int node_rank = prefix_count + left_size + 1;

    if (node_rank > end_rank) {
        if (left_size > 0 && prefix_count < end_rank) {
            avl_get_range_optimized(node->left, start_rank, end_rank, prefix_count,
                                     player_ids, scores, timestamps, max_count, filled_count);
        }
        return 1;
    }

    if (node_rank < start_rank) {
        int right_prefix = prefix_count + left_size + 1;
        avl_get_range_optimized(node->right, start_rank, end_rank, right_prefix,
                                 player_ids, scores, timestamps, max_count, filled_count);
        return 1;
    }

    if (left_size > 0 && prefix_count < end_rank) {
        avl_get_range_optimized(node->left, start_rank, end_rank, prefix_count,
                                 player_ids, scores, timestamps, max_count, filled_count);
    }

    if (*filled_count < max_count && node_rank >= start_rank && node_rank <= end_rank) {
        player_ids[*filled_count] = node->player_id;
        scores[*filled_count] = node->score;
        timestamps[*filled_count] = node->timestamp;
        (*filled_count)++;
    }

    if (*filled_count < max_count) {
        int right_prefix = prefix_count + left_size + 1;
        avl_get_range_optimized(node->right, start_rank, end_rank, right_prefix,
                                 player_ids, scores, timestamps, max_count, filled_count);
    }

    return 1;
}

static int avl_count_ge_score(AVLNode *node, int32_t score) {
    if (node == NULL) return 0;
    
    if (node->score >= score) {
        return 1 + size(node->left) + avl_count_ge_score(node->right, score);
    } else {
        return avl_count_ge_score(node->right, score);
    }
}

static int avl_count_gt_score(AVLNode *node, int32_t score) {
    if (node == NULL) return 0;
    
    if (node->score > score) {
        return 1 + size(node->left) + avl_count_gt_score(node->right, score);
    } else {
        return avl_count_gt_score(node->right, score);
    }
}

static void avl_destroy_node(AVLNode *node) {
    if (node == NULL) return;
    avl_destroy_node(node->left);
    avl_destroy_node(node->right);
    free(node);
}

AVLTree* avl_create(int max_capacity) {
    AVLTree *tree = (AVLTree*)malloc(sizeof(AVLTree));
    if (tree == NULL) return NULL;
    tree->root = NULL;
    tree->max_capacity = max_capacity;
    tree->current_size = 0;
    return tree;
}

void avl_destroy(AVLTree *tree) {
    if (tree == NULL) return;
    avl_destroy_node(tree->root);
    free(tree);
}

int avl_insert(AVLTree *tree, uint64_t player_id, int32_t score, uint64_t timestamp) {
    if (tree == NULL) return 0;

    int inserted = 0;
    tree->root = avl_insert_node(tree->root, player_id, score, timestamp, &inserted);
    
    if (inserted) {
        tree->current_size++;
        
        if (tree->max_capacity > 0 && tree->current_size > tree->max_capacity) {
            AVLNode *max_node = get_max_node(tree->root);
            if (max_node != NULL) {
                int removed = 0;
                tree->root = avl_remove_node(tree->root, max_node->player_id, 
                                              max_node->score, max_node->timestamp, &removed);
                if (removed) {
                    tree->current_size--;
                }
            }
        }
    }
    
    return inserted;
}

int avl_remove(AVLTree *tree, uint64_t player_id, int32_t score, uint64_t timestamp) {
    if (tree == NULL) return 0;

    int removed = 0;
    tree->root = avl_remove_node(tree->root, player_id, score, timestamp, &removed);
    
    if (removed) {
        tree->current_size--;
    }
    
    return removed;
}

int avl_get_rank(AVLTree *tree, uint64_t player_id, int32_t score, uint64_t timestamp) {
    if (tree == NULL || tree->root == NULL) return -1;
    return avl_get_rank_node(tree->root, player_id, score, timestamp);
}

int avl_get_by_rank(AVLTree *tree, int rank, uint64_t *player_id, int32_t *score, uint64_t *timestamp) {
    if (tree == NULL || tree->root == NULL || rank < 1) return 0;
    return avl_get_by_rank_node(tree->root, rank, player_id, score, timestamp);
}

int avl_get_range(AVLTree *tree, int start_rank, int end_rank, 
                  uint64_t *player_ids, int32_t *scores, uint64_t *timestamps, int max_count) {
    if (tree == NULL || tree->root == NULL || start_rank < 1 || end_rank < start_rank) return 0;
    
    int filled_count = 0;
    
    avl_get_range_optimized(tree->root, start_rank, end_rank, 0,
                            player_ids, scores, timestamps, max_count, &filled_count);
    
    return filled_count;
}

int avl_count_in_score_range(AVLTree *tree, int32_t min_score, int32_t max_score) {
    if (tree == NULL || tree->root == NULL) return 0;
    
    int count_ge_min = avl_count_ge_score(tree->root, min_score);
    int count_gt_max = avl_count_gt_score(tree->root, max_score);
    
    return count_ge_min - count_gt_max;
}

int avl_size(AVLTree *tree) {
    if (tree == NULL) return 0;
    return tree->current_size;
}

void avl_clear(AVLTree *tree) {
    if (tree == NULL) return;
    avl_destroy_node(tree->root);
    tree->root = NULL;
    tree->current_size = 0;
}
