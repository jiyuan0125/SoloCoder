#ifndef FILTER_SORT_H
#define FILTER_SORT_H

#include "product_store.h"

#define FILTER_OK              0
#define FILTER_ERR_NULL       -1
#define FILTER_ERR_MEMORY     -2

typedef enum {
    SORT_FIELD_ID,
    SORT_FIELD_NAME,
    SORT_FIELD_PRICE,
    SORT_FIELD_QUANTITY
} SortField;

typedef enum {
    SORT_ASC,
    SORT_DESC
} SortOrder;

typedef struct {
    const char* name_substring;
    int price_min_enabled;
    double price_min;
    int price_max_enabled;
    double price_max;
    int quantity_threshold_enabled;
    int quantity_below;
} FilterConditions;

typedef struct {
    SortField field;
    SortOrder order;
} SortOptions;

typedef struct {
    Product** items;
    size_t count;
    size_t capacity;
} ProductList;

ProductList* product_list_create(size_t initial_capacity);
void product_list_destroy(ProductList* list);
int product_list_append(ProductList* list, Product* product);

ProductList* filter_products(const InventoryStore* store, const FilterConditions* filter);
void sort_product_list(ProductList* list, const SortOptions* options);

ProductList* filter_and_sort_products(const InventoryStore* store, 
                                        const FilterConditions* filter,
                                        const SortOptions* sort);

void filter_conditions_init(FilterConditions* filter);
void sort_options_init(SortOptions* options);

int strstr_case_insensitive(const char* haystack, const char* needle);

#endif
