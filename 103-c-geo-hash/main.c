#include <stdio.h>
#include "base32.h"
#include "geohash.h"
#include "search.h"

static void print_bounds(const char *hash, GeoHashBounds bounds)
{
    printf("  GeoHash: %s\n", hash);
    printf("  范围: 纬度 [%.6f, %.6f], 经度 [%.6f, %.6f]\n",
           bounds.lat_min, bounds.lat_max, bounds.lon_min, bounds.lon_max);
    printf("  中心点: (%.6f, %.6f)\n",
           (bounds.lat_min + bounds.lat_max) / 2,
           (bounds.lon_min + bounds.lon_max) / 2);
}

static void print_neighbors(const char *hash)
{
    char neighbors[9][GEOHASH_PRECISION_MAX + 1];
    int i;
    
    printf("\n=== 相邻区域编码 (中心+周边8个区域):\n");
    geohash_get_neighbors(hash, neighbors);
    
    printf("  左上: %s    上: %s    右上: %s\n",
           neighbors[0], neighbors[1], neighbors[2]);
    printf("  左:   %s    中: %s    右:   %s\n",
           neighbors[3], neighbors[4], neighbors[5]);
    printf("  左下: %s    下: %s    右下: %s\n",
           neighbors[6], neighbors[7], neighbors[8]);
}

int main(void)
{
    MerchantDatabase db;
    Merchant results[MAX_SEARCH_RESULTS];
    int result_count;
    int i;
    char geohash_str[GEOHASH_PRECISION_MAX + 1];
    GeoHashBounds bounds;
    
    printf("========================================\n");
    printf("  外卖平台配送范围判断模块演示\n");
    printf("========================================\n\n");
    
    merchant_db_init(&db);
    
    printf("=== 第一部分: 地理编码转换演示\n");
    printf("----------------------------------------\n");
    
    double test_lat = 39.9042;
    double test_lon = 116.4074;
    
    printf("测试坐标: 北京 (%.6f, %.6f)\n\n", test_lat, test_lon);
    
    for (i = 1; i <= 7; i++) {
        geohash_encode(test_lat, test_lon, i, geohash_str);
        printf("精度 %d 位: %s\n", i, geohash_str);
        
        bounds = geohash_decode_bounds(geohash_str);
        printf("  覆盖范围: 纬度 %.6fkm, 经度 %.6fkm\n",
               haversine_distance(bounds.lat_min, bounds.lon_min, bounds.lat_max, bounds.lon_min) / 1000.0,
               haversine_distance(bounds.lat_min, bounds.lon_min, bounds.lat_min, bounds.lon_max) / 1000.0);
        printf("\n");
    }
    
    printf("\n=== 第二部分: 商家数据录入\n");
    printf("----------------------------------------\n");
    
    merchant_add(&db, "老北京炸酱面馆", 39.9050, 116.4080, 6);
    merchant_add(&db, "全聚德烤鸭店", 39.9030, 116.4060, 6);
    merchant_add(&db, "东来顺火锅", 39.9045, 116.4090, 6);
    merchant_add(&db, "海底捞火锅", 39.9100, 116.4150, 6);
    merchant_add(&db, "肯德基餐厅", 39.8950, 116.3980, 6);
    merchant_add(&db, "麦当劳餐厅", 39.9200, 116.4200, 6);
    
    printf("已录入 %d 个商家:\n", db.count);
    for (i = 0; i < db.count; i++) {
        printf("  %s: (%.6f, %.6f) -> %s\n",
               db.merchants[i].name,
               db.merchants[i].latitude,
               db.merchants[i].longitude,
               db.merchants[i].geohash);
    }
    
    printf("\n=== 第三部分: 相邻区域搜索演示\n");
    printf("----------------------------------------\n");
    
    geohash_encode(test_lat, test_lon, 5, geohash_str);
    printf("中心点 5 位编码: %s\n", geohash_str);
    print_neighbors(geohash_str);
    
    printf("\n=== 第四部分: 半径搜索演示\n");
    printf("----------------------------------------\n");
    
    double search_lat = 39.9042;
    double search_lon = 116.4074;
    double radius = 2000.0;
    
    printf("搜索中心: (%.6f, %.6f)\n", search_lat, search_lon);
    printf("搜索半径: %.0f 米\n", radius);
    printf("对应精度: %d 位\n\n", geohash_precision_for_radius(radius));
    
    result_count = search_merchants_by_radius(&db, search_lat, search_lon, radius, results, MAX_SEARCH_RESULTS);
    
    printf("找到 %d 个商家在 %.0f 米范围内:\n", result_count, radius);
    
    for (i = 0; i < result_count; i++) {
        double dist = haversine_distance(search_lat, search_lon,
                                          results[i].latitude, results[i].longitude);
        printf("  %d. %s\n", i + 1, results[i].name);
        printf("     位置: (%.6f, %.6f)\n", results[i].latitude, results[i].longitude);
        printf("     编码: %s\n", results[i].geohash);
        printf("     距离: %.1f 米\n", dist);
    }
    
    printf("\n=== 第五部分: 编码反向解码演示\n");
    printf("----------------------------------------\n");
    
    const char *test_hash = "wx4g0s";
    printf("测试编码: %s\n", test_hash);
    bounds = geohash_decode_bounds(test_hash);
    print_bounds(test_hash, bounds);
    
    printf("\n========================================\n");
    printf("演示结束\n");
    printf("========================================\n");
    
    return 0;
}
