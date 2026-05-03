#include "filter_sort.h"
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

#define INITIAL_LIST_CAPACITY  16

void filter_conditions_init(FilterConditions* filter) {
    if (!filter) return;
    memset(filter, 0, sizeof(FilterConditions));
}

void sort_options_init(SortOptions* options) {
    if (!options) return;
    options->field = SORT_FIELD_ID;
    options->order = SORT_ASC;
}

int strstr_case_insensitive(const char* haystack, const char* needle) {
    if (!haystack || !needle || needle[0] == '\0') {
        return 1;
    }
    
    for (const char* h = haystack; *h != '\0'; h++) {
        const char* h_ptr = h;
        const char* n_ptr = needle;
        
        while (*n_ptr != '\0' && *h_ptr != '\0') {
            int hc = tolower((unsigned char)*h_ptr);
            int nc = tolower((unsigned char)*n_ptr);
            if (hc != nc) {
                break;
            }
            h_ptr++;
            n_ptr++;
        }
        
        if (*n_ptr == '\0') {
            return 1;
        }
    }
    
    return 0;
}

ProductList* product_list_create(size_t initial_capacity) {
    if (initial_capacity == 0) {
        initial_capacity = INITIAL_LIST_CAPACITY;
    }
    
    ProductList* list = (ProductList*)malloc(sizeof(ProductList));
    if (!list) return NULL;
    
    list->items = (Product**)malloc(initial_capacity * sizeof(Product*));
    if (!list->items) {
        free(list);
        return NULL;
    }
    
    list->count = 0;
    list->capacity = initial_capacity;
    
    return list;
}

void product_list_destroy(ProductList* list) {
    if (!list) return;
    free(list->items);
    free(list);
}

static int product_list_reserve(ProductList* list, size_t new_capacity) {
    if (new_capacity <= list->capacity) {
        return FILTER_OK;
    }
    
    Product** new_items = (Product**)realloc(list->items, new_capacity * sizeof(Product*));
    if (!new_items) {
        return FILTER_ERR_MEMORY;
    }
    
    list->items = new_items;
    list->capacity = new_capacity;
    return FILTER_OK;
}

int product_list_append(ProductList* list, Product* product) {
    if (!list || !product) return FILTER_ERR_NULL;
    
    if (list->count >= list->capacity) {
        int ret = product_list_reserve(list, list->capacity * 2);
        if (ret != FILTER_OK) {
            return ret;
        }
    }
    
    list->items[list->count++] = product;
    return FILTER_OK;
}

static int product_matches_filter(const Product* product, const FilterConditions* filter) {
    if (!filter) {
        return 1;
    }
    
    if (filter->name_substring && filter->name_substring[0] != '\0') {
        if (!strstr_case_insensitive(product->name, filter->name_substring)) {
            return 0;
        }
    }
    
    if (filter->price_min_enabled) {
        if (product->price < filter->price_min) {
            return 0;
        }
    }
    
    if (filter->price_max_enabled) {
        if (product->price > filter->price_max) {
            return 0;
        }
    }
    
    if (filter->quantity_threshold_enabled) {
        if (product->quantity >= filter->quantity_below) {
            return 0;
        }
    }
    
    return 1;
}

ProductList* filter_products(const InventoryStore* store, const FilterConditions* filter) {
    if (!store) return NULL;
    
    size_t all_count = 0;
    Product** all_products = inventory_store_get_all_active(store, &all_count);
    
    if (all_count == 0 || !all_products) {
        inventory_store_free_product_array(all_products, all_count);
        return product_list_create(0);
    }
    
    ProductList* result = product_list_create(all_count);
    if (!result) {
        inventory_store_free_product_array(all_products, all_count);
        return NULL;
    }
    
    for (size_t i = 0; i < all_count; i++) {
        if (product_matches_filter(all_products[i], filter)) {
            product_list_append(result, all_products[i]);
        }
    }
    
    inventory_store_free_product_array(all_products, all_count);
    return result;
}

static int sort_compare_id_asc(const void* a, const void* b) {
    Product* const* pa = (Product* const*)a;
    Product* const* pb = (Product* const*)b;
    return strcasecmp_custom((*pa)->id, (*pb)->id);
}

static int sort_compare_id_desc(const void* a, const void* b) {
    return -sort_compare_id_asc(a, b);
}

static int sort_compare_name_asc(const void* a, const void* b) {
    Product* const* pa = (Product* const*)a;
    Product* const* pb = (Product* const*)b;
    return strcasecmp_custom((*pa)->name, (*pb)->name);
}

static int sort_compare_name_desc(const void* a, const void* b) {
    return -sort_compare_name_asc(a, b);
}

static int sort_compare_price_asc(const void* a, const void* b) {
    Product* const* pa = (Product* const*)a;
    Product* const* pb = (Product* const*)b;
    if ((*pa)->price < (*pb)->price) return -1;
    if ((*pa)->price > (*pb)->price) return 1;
    return 0;
}

static int sort_compare_price_desc(const void* a, const void* b) {
    return -sort_compare_price_asc(a, b);
}

static int sort_compare_quantity_asc(const void* a, const void* b) {
    Product* const* pa = (Product* const*)a;
    Product* const* pb = (Product* const*)b;
    return (*pa)->quantity - (*pb)->quantity;
}

static int sort_compare_quantity_desc(const void* a, const void* b) {
    return -sort_compare_quantity_asc(a, b);
}

void sort_product_list(ProductList* list, const SortOptions* options) {
    if (!list || list->count <= 1 || !options) {
        return;
    }
    
    int (*compar)(const void*, const void*) = NULL;
    
    switch (options->field) {
        case SORT_FIELD_ID:
            compar = (options->order == SORT_ASC) ? sort_compare_id_asc : sort_compare_id_desc;
            break;
        case SORT_FIELD_NAME:
            compar = (options->order == SORT_ASC) ? sort_compare_name_asc : sort_compare_name_desc;
            break;
        case SORT_FIELD_PRICE:
            compar = (options->order == SORT_ASC) ? sort_compare_price_asc : sort_compare_price_desc;
            break;
        case SORT_FIELD_QUANTITY:
            compar = (options->order == SORT_ASC) ? sort_compare_quantity_asc : sort_compare_quantity_desc;
            break;
        default:
            compar = sort_compare_id_asc;
            break;
    }
    
    qsort(list->items, list->count, sizeof(Product*), compar);
}

ProductList* filter_and_sort_products(const InventoryStore* store, 
                                        const FilterConditions* filter,
                                        const SortOptions* sort) {
    ProductList* result = filter_products(store, filter);
    if (result && sort) {
        sort_product_list(result, sort);
    }
    return result;
}
