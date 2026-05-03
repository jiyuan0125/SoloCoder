#include "text_replace.h"
#include <stdlib.h>
#include <string.h>

int match_collection_init(match_collection_t *collection, size_t initial_capacity) {
    if (collection == NULL) {
        return -1;
    }
    
    if (initial_capacity == 0) {
        initial_capacity = INITIAL_MATCH_CAPACITY;
    }
    
    collection->matches = (match_result_t *)malloc(initial_capacity * sizeof(match_result_t));
    if (collection->matches == NULL) {
        collection->count = 0;
        collection->capacity = 0;
        return -1;
    }
    
    collection->count = 0;
    collection->capacity = initial_capacity;
    return 0;
}

void match_collection_free(match_collection_t *collection) {
    if (collection != NULL && collection->matches != NULL) {
        free(collection->matches);
        collection->matches = NULL;
        collection->count = 0;
        collection->capacity = 0;
    }
}

int match_collection_add(match_collection_t *collection, const match_result_t *match) {
    if (collection == NULL || match == NULL) {
        return -1;
    }
    
    if (collection->count >= collection->capacity) {
        size_t new_capacity = collection->capacity * 2;
        if (new_capacity < INITIAL_MATCH_CAPACITY) {
            new_capacity = INITIAL_MATCH_CAPACITY;
        }
        
        match_result_t *new_matches = (match_result_t *)realloc(
            collection->matches, new_capacity * sizeof(match_result_t));
        if (new_matches == NULL) {
            return -1;
        }
        
        collection->matches = new_matches;
        collection->capacity = new_capacity;
    }
    
    collection->matches[collection->count] = *match;
    collection->count++;
    return 0;
}

static void sort_matches_by_start(match_collection_t *collection) {
    for (size_t i = 0; i < collection->count; i++) {
        for (size_t j = i + 1; j < collection->count; j++) {
            if (collection->matches[i].start > collection->matches[j].start) {
                match_result_t temp = collection->matches[i];
                collection->matches[i] = collection->matches[j];
                collection->matches[j] = temp;
            }
        }
    }
}

static void remove_overlapping_matches(match_collection_t *collection) {
    if (collection->count <= 1) {
        return;
    }
    
    sort_matches_by_start(collection);
    
    match_result_t *filtered = (match_result_t *)malloc(
        collection->capacity * sizeof(match_result_t));
    if (filtered == NULL) {
        return;
    }
    
    size_t filtered_count = 0;
    size_t last_end = 0;
    
    for (size_t i = 0; i < collection->count; i++) {
        const match_result_t *current = &collection->matches[i];
        
        if (current->start >= last_end) {
            filtered[filtered_count] = *current;
            filtered_count++;
            last_end = current->start + current->length;
        }
    }
    
    free(collection->matches);
    collection->matches = filtered;
    collection->count = filtered_count;
}

size_t collect_all_matches(const char *input, size_t len, match_collection_t *collection) {
    if (input == NULL || collection == NULL) {
        return 0;
    }
    
    match_collection_t temp_collection;
    if (match_collection_init(&temp_collection, INITIAL_MATCH_CAPACITY) != 0) {
        return 0;
    }
    
    size_t pos = 0;
    
    while (pos < len && input[pos] != '\0') {
        match_result_t result;
        int found = 0;
        
        if (match_phone(input, pos, &result)) {
            match_collection_add(&temp_collection, &result);
            pos = result.start + result.length;
            found = 1;
        }
        else if (match_id_card(input, pos, &result)) {
            match_collection_add(&temp_collection, &result);
            pos = result.start + result.length;
            found = 1;
        }
        else if (match_bank_card(input, pos, &result)) {
            match_collection_add(&temp_collection, &result);
            pos = result.start + result.length;
            found = 1;
        }
        else if (match_email(input, pos, &result)) {
            match_collection_add(&temp_collection, &result);
            pos = result.start + result.length;
            found = 1;
        }
        
        if (!found) {
            pos++;
        }
    }
    
    remove_overlapping_matches(&temp_collection);
    sort_matches_by_start(&temp_collection);
    
    *collection = temp_collection;
    return collection->count;
}

static size_t calculate_output_size(const char *input, size_t input_len,
                                     const match_collection_t *collection) {
    if (collection == NULL || collection->count == 0) {
        return input_len;
    }
    
    size_t total_size = 0;
    size_t last_end = 0;
    
    for (size_t i = 0; i < collection->count; i++) {
        const match_result_t *match = &collection->matches[i];
        
        total_size += (match->start - last_end);
        
        desensitize_func_t func = get_desensitize_func(match->type);
        if (func != NULL) {
            size_t masked_len = func(input + match->start, match->length, NULL, 0);
            total_size += masked_len;
        } else {
            total_size += match->length;
        }
        
        last_end = match->start + match->length;
    }
    
    total_size += (input_len - last_end);
    
    return total_size;
}

size_t desensitize_text(const char *input, size_t input_len,
                        char *output, size_t output_size,
                        size_t *required_size) {
    if (input == NULL || required_size == NULL) {
        if (required_size != NULL) {
            *required_size = 0;
        }
        return 0;
    }
    
    match_collection_t collection;
    if (match_collection_init(&collection, INITIAL_MATCH_CAPACITY) != 0) {
        if (output != NULL && output_size > input_len) {
            memcpy(output, input, input_len);
            output[input_len] = '\0';
        }
        *required_size = input_len;
        return input_len;
    }
    
    collect_all_matches(input, input_len, &collection);
    
    size_t needed = calculate_output_size(input, input_len, &collection);
    *required_size = needed;
    
    if (output == NULL || output_size == 0) {
        match_collection_free(&collection);
        return 0;
    }
    
    if (output_size <= needed) {
        if (output_size > 0) {
            output[0] = '\0';
        }
        match_collection_free(&collection);
        return 0;
    }
    
    size_t out_idx = 0;
    size_t last_end = 0;
    
    for (size_t i = 0; i < collection.count; i++) {
        const match_result_t *match = &collection.matches[i];
        
        size_t copy_len = match->start - last_end;
        if (copy_len > 0) {
            memcpy(output + out_idx, input + last_end, copy_len);
            out_idx += copy_len;
        }
        
        desensitize_func_t func = get_desensitize_func(match->type);
        if (func != NULL) {
            size_t masked_len = func(input + match->start, match->length, NULL, 0);
            char *temp_mask = (char *)malloc(masked_len + 1);
            
            if (temp_mask != NULL) {
                func(input + match->start, match->length, temp_mask, masked_len + 1);
                
                if (out_idx + masked_len < output_size) {
                    memcpy(output + out_idx, temp_mask, masked_len);
                    out_idx += masked_len;
                }
                free(temp_mask);
            }
        } else {
            if (out_idx + match->length < output_size) {
                memcpy(output + out_idx, input + match->start, match->length);
                out_idx += match->length;
            }
        }
        
        last_end = match->start + match->length;
    }
    
    size_t remaining = input_len - last_end;
    if (remaining > 0 && out_idx + remaining < output_size) {
        memcpy(output + out_idx, input + last_end, remaining);
        out_idx += remaining;
    }
    
    if (out_idx < output_size) {
        output[out_idx] = '\0';
    } else if (output_size > 0) {
        output[output_size - 1] = '\0';
    }
    
    match_collection_free(&collection);
    return out_idx;
}
