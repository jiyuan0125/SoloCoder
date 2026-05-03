#ifndef TEXT_REPLACE_H
#define TEXT_REPLACE_H

#include "sensitive_info.h"
#include "desensitize_rule.h"
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

#define INITIAL_MATCH_CAPACITY 32

typedef struct {
    match_result_t *matches;
    size_t count;
    size_t capacity;
} match_collection_t;

int match_collection_init(match_collection_t *collection, size_t initial_capacity);
void match_collection_free(match_collection_t *collection);
int match_collection_add(match_collection_t *collection, const match_result_t *match);

size_t collect_all_matches(const char *input, size_t len, match_collection_t *collection);

size_t desensitize_text(const char *input, size_t input_len,
                        char *output, size_t output_size,
                        size_t *required_size);

#ifdef __cplusplus
}
#endif

#endif
