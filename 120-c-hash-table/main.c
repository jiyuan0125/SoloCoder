#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "product_store.h"
#include "inventory_trans.h"
#include "filter_sort.h"

static void print_product(const Product* p) {
    if (!p) {
        printf("  (null)\n");
        return;
    }
    printf("  ID: %s, Name: %s, Qty: %d, Price: %.2f, Location: %s\n",
           p->id, p->name, p->quantity, p->price, p->location);
}

static void print_product_list(const ProductList* list, const char* title) {
    printf("\n========== %s (%zu items) ==========\n", title, list ? list->count : 0);
    if (!list || list->count == 0) {
        printf("  No items.\n");
        return;
    }
    for (size_t i = 0; i < list->count; i++) {
        printf("[%zu] ", i + 1);
        print_product(list->items[i]);
    }
}

int main(void) {
    printf("========================================\n");
    printf("   仓库库存管理系统演示\n");
    printf("========================================\n");

    InventoryStore* store = inventory_store_create(16);
    if (!store) {
        printf("错误：无法创建库存存储\n");
        return 1;
    }

    InventoryTransaction* trans = inventory_transaction_create(store);
    if (!trans) {
        printf("错误：无法创建事务管理器\n");
        inventory_store_destroy(store);
        return 1;
    }

    printf("\n----- 1. 添加一批商品 -----\n");
    
    Product products[] = {
        {"APPLE001", "红富士苹果", 100, 5.99, "A区-01-01", 0},
        {"BANANA01", "菲律宾香蕉", 50, 3.50, "A区-01-02", 0},
        {"ORANGE01", "赣南脐橙", 80, 6.80, "A区-01-03", 0},
        {"Milk001", "纯牛奶", 200, 4.50, "B区-02-01", 0},
        {"BREAD001", "全麦面包", 0, 8.90, "B区-02-02", 0},
        {"EGG0001", "土鸡蛋", 10, 15.00, "C区-03-01", 0},
        {"apple001", "这个ID会重复", 999, 1.00, "XX-XX-XX", 0},
    };
    
    int num_products = sizeof(products) / sizeof(products[0]);
    int add_count = 0;
    
    for (int i = 0; i < num_products; i++) {
        int ret = inventory_store_add(store, &products[i]);
        if (ret == INVENTORY_STORE_OK) {
            printf("添加成功: %s\n", products[i].id);
            add_count++;
        } else if (ret == INVENTORY_STORE_ERR_DUPLICATE) {
            printf("添加失败(重复ID): %s (与APPLE001是同一个，不区分大小写)\n", products[i].id);
        } else {
            printf("添加失败: %s (错误码: %d)\n", products[i].id, ret);
        }
    }
    
    printf("\n当前活跃商品数: %zu\n", inventory_store_active_count(store));

    printf("\n----- 2. 按编号查询商品 -----\n");
    
    const char* query_ids[] = {"APPLE001", "apple001", "Banana01", "NOTEXIST", "EGG0001"};
    int num_queries = sizeof(query_ids) / sizeof(query_ids[0]);
    
    for (int i = 0; i < num_queries; i++) {
        Product* p = inventory_store_find(store, query_ids[i]);
        printf("查询 \"%s\": ", query_ids[i]);
        if (p) {
            printf("找到 -> 原始ID: %s, 库存: %d\n", p->id, p->quantity);
        } else {
            printf("未找到 (返回空，不崩溃)\n");
        }
    }

    printf("\n----- 3. 入库和出库操作 -----\n");
    
    printf("当前苹果库存: %d\n", inventory_store_find(store, "APPLE001")->quantity);
    
    int ret = inventory_store_update_quantity(store, "APPLE001", 50);
    printf("入库 50 个苹果: %s\n", ret == 0 ? "成功" : "失败");
    printf("入库后苹果库存: %d\n", inventory_store_find(store, "APPLE001")->quantity);
    
    ret = inventory_store_update_quantity(store, "APPLE001", -30);
    printf("出库 30 个苹果: %s\n", ret == 0 ? "成功" : "失败");
    printf("出库后苹果库存: %d\n", inventory_store_find(store, "APPLE001")->quantity);
    
    printf("\n当前鸡蛋库存: %d\n", inventory_store_find(store, "EGG0001")->quantity);
    int available = 0;
    ret = inventory_store_update_quantity(store, "EGG0001", -20);
    if (ret == INVENTORY_STORE_ERR_NEGATIVE) {
        available = inventory_store_find(store, "EGG0001")->quantity;
        printf("尝试出库 20 个鸡蛋: 失败，库存不足。当前可用: %d\n", available);
    }

    printf("\n----- 4. 批量入库（事务语义） -----\n");
    
    StockChangeItem inbound_items[] = {
        {"APPLE001", 100},
        {"BANANA01", 200},
        {"ORANGE01", 150},
    };
    int inbound_count = sizeof(inbound_items) / sizeof(inbound_items[0]);
    
    printf("批量入库前:\n");
    for (int i = 0; i < inbound_count; i++) {
        Product* p = inventory_store_find(store, inbound_items[i].product_id);
        printf("  %s: %d\n", inbound_items[i].product_id, p ? p->quantity : -1);
    }
    
    ret = inventory_batch_inbound(trans, inbound_items, inbound_count);
    printf("批量入库 %s\n", ret == TRANS_OK ? "成功" : "失败");
    
    printf("批量入库后:\n");
    for (int i = 0; i < inbound_count; i++) {
        Product* p = inventory_store_find(store, inbound_items[i].product_id);
        printf("  %s: %d\n", inbound_items[i].product_id, p ? p->quantity : -1);
    }

    printf("\n----- 5. 批量出库（事务语义，测试失败情况） -----\n");
    
    printf("当前鸡蛋库存: %d\n", inventory_store_find(store, "EGG0001")->quantity);
    
    StockChangeItem outbound_items[] = {
        {"APPLE001", 10},
        {"EGG0001", 100},
        {"BANANA01", 5},
    };
    int outbound_count = sizeof(outbound_items) / sizeof(outbound_items[0]);
    const char* failed_id = NULL;
    int avail = 0;
    
    printf("尝试批量出库（苹果10个、鸡蛋100个、香蕉5个）...\n");
    ret = inventory_batch_outbound(trans, outbound_items, outbound_count, &failed_id, &avail);
    
    if (ret != TRANS_OK) {
        printf("批量出库失败！\n");
        printf("  失败商品: %s\n", failed_id ? failed_id : "unknown");
        printf("  该商品可用库存: %d\n", avail);
        printf("  检查苹果库存是否未变: %d (期望值220)\n", 
               inventory_store_find(store, "APPLE001")->quantity);
    } else {
        printf("批量出库成功\n");
    }

    printf("\n----- 6. 正常批量出库 -----\n");
    
    StockChangeItem outbound_items2[] = {
        {"APPLE001", 20},
        {"BANANA01", 30},
        {"ORANGE01", 15},
    };
    int outbound_count2 = sizeof(outbound_items2) / sizeof(outbound_items2[0]);
    
    printf("批量出库前:\n");
    for (int i = 0; i < outbound_count2; i++) {
        Product* p = inventory_store_find(store, outbound_items2[i].product_id);
        printf("  %s: %d\n", outbound_items2[i].product_id, p ? p->quantity : -1);
    }
    
    ret = inventory_batch_outbound(trans, outbound_items2, outbound_count2, NULL, NULL);
    printf("批量出库 %s\n", ret == TRANS_OK ? "成功" : "失败");
    
    printf("批量出库后:\n");
    for (int i = 0; i < outbound_count2; i++) {
        Product* p = inventory_store_find(store, outbound_items2[i].product_id);
        printf("  %s: %d\n", outbound_items2[i].product_id, p ? p->quantity : -1);
    }

    printf("\n----- 7. 筛选查询演示 -----\n");
    
    ProductList* all = filter_and_sort_products(store, NULL, NULL);
    print_product_list(all, "所有商品（默认按ID升序）");
    product_list_destroy(all);

    FilterConditions filter_low_stock;
    filter_conditions_init(&filter_low_stock);
    filter_low_stock.quantity_threshold_enabled = 1;
    filter_low_stock.quantity_below = 50;
    
    ProductList* low_stock = filter_and_sort_products(store, &filter_low_stock, NULL);
    print_product_list(low_stock, "库存低于50的商品（库存预警）");
    product_list_destroy(low_stock);

    FilterConditions filter_name;
    filter_conditions_init(&filter_name);
    filter_name.name_substring = "果";
    
    SortOptions sort_by_price_desc;
    sort_options_init(&sort_by_price_desc);
    sort_by_price_desc.field = SORT_FIELD_PRICE;
    sort_by_price_desc.order = SORT_DESC;
    
    ProductList* fruit_by_price = filter_and_sort_products(store, &filter_name, &sort_by_price_desc);
    print_product_list(fruit_by_price, "名称含\"果\"的商品（按价格降序）");
    product_list_destroy(fruit_by_price);

    FilterConditions filter_price;
    filter_conditions_init(&filter_price);
    filter_price.price_min_enabled = 1;
    filter_price.price_min = 5.0;
    filter_price.price_max_enabled = 1;
    filter_price.price_max = 10.0;
    
    SortOptions sort_by_qty_asc;
    sort_options_init(&sort_by_qty_asc);
    sort_by_qty_asc.field = SORT_FIELD_QUANTITY;
    sort_by_qty_asc.order = SORT_ASC;
    
    ProductList* price_range = filter_and_sort_products(store, &filter_price, &sort_by_qty_asc);
    print_product_list(price_range, "价格5-10元的商品（按库存升序）");
    product_list_destroy(price_range);

    printf("\n----- 8. 标记删除演示 -----\n");
    
    printf("删除前活跃商品数: %zu\n", inventory_store_active_count(store));
    
    ret = inventory_store_mark_deleted(store, "BREAD001");
    printf("标记删除全麦面包 (BREAD001): %s\n", ret == 0 ? "成功" : "失败");
    
    printf("删除后活跃商品数: %zu\n", inventory_store_active_count(store));
    
    Product* bread = inventory_store_find(store, "BREAD001");
    printf("查询已删除的面包: %s\n", bread ? "找到（不应该）" : "未找到（正确，返回空）");
    
    printf("\n总记录数（含已删除）: %zu\n", inventory_store_size(store));

    printf("\n========================================\n");
    printf("   演示结束\n");
    printf("========================================\n");

    inventory_transaction_destroy(trans);
    inventory_store_destroy(store);

    return 0;
}
