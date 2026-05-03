#include <stdio.h>
#include <string.h>
#include <math.h>

#ifndef M_PI
#define M_PI 3.14159265358979323846
#endif

#include "base32.h"
#include "geohash.h"
#include "search.h"

static void test_precision_at_latitude(void)
{
    printf("=== 测试: 不同纬度下的精度选择\n");
    printf("----------------------------------------\n");
    
    double radius = 800.0;
    double lats[] = {0.0, 30.0, 45.0, 60.0, 70.0, 80.0, 85.0, 89.0};
    int num_lats = sizeof(lats) / sizeof(lats[0]);
    int i;
    
    printf("搜索半径: %.0f 米\n\n", radius);
    printf("  纬度    精度    经度格子    纬度格子    3×3经度覆盖  3×3纬度覆盖\n");
    printf("  ------- ------- ----------- ----------- ------------- -------------\n");
    
    for (i = 0; i < num_lats; i++) {
        double lat = lats[i];
        int precision = geohash_precision_for_radius_at_lat(radius, lat);
        
        int total_bits = precision * 5;
        int lon_bits = (total_bits + 1) / 2;
        int lat_bits = total_bits / 2;
        
        double lon_deg_per_cell = 360.0 / (1LL << lon_bits);
        double lat_deg_per_cell = 180.0 / (1LL << lat_bits);
        
        double lat_rad = fabs(lat) * M_PI / 180.0;
        double cos_lat = cos(lat_rad);
        if (cos_lat < 0.0001) cos_lat = 0.0001;
        
        double km_per_deg_lon = 111.0 * cos_lat;
        
        double lon_km_per_cell = lon_deg_per_cell * km_per_deg_lon * 1000.0;
        double lat_km_per_cell = lat_deg_per_cell * 111.0 * 1000.0;
        
        printf("  %6.1f°    %d 位    %6.1f 米    %6.1f 米    %7.1f 米    %7.1f 米\n",
               lat, precision, lon_km_per_cell, lat_km_per_cell,
               3 * lon_km_per_cell, 3 * lat_km_per_cell);
    }
}

static void test_high_latitude_search(void)
{
    printf("\n=== 测试: 高纬度搜索验证 (lat=80°)\n");
    printf("----------------------------------------\n");
    
    MerchantDatabase db;
    Merchant results[MAX_SEARCH_RESULTS];
    int result_count;
    int i;
    
    double center_lat = 80.0;
    double center_lon = 0.0;
    double radius = 800.0;
    
    merchant_db_init(&db);
    
    merchant_add(&db, "中心商家", center_lat, center_lon, 6);
    
    double east_500m_lon = center_lon + 500.0 / (111000.0 * cos(80.0 * M_PI / 180.0));
    merchant_add(&db, "正东 500 米", center_lat, east_500m_lon, 6);
    
    merchant_add(&db, "正南 500 米", center_lat - 500.0 / 111000.0, center_lon, 6);
    merchant_add(&db, "正西 500 米", center_lat, center_lon - 500.0 / (111000.0 * cos(80.0 * M_PI / 180.0)), 6);
    merchant_add(&db, "正北 500 米", center_lat + 500.0 / 111000.0, center_lon, 6);
    
    merchant_add(&db, "正东 1000 米", center_lat, center_lon + 1000.0 / (111000.0 * cos(80.0 * M_PI / 180.0)), 6);
    merchant_add(&db, "正南 1000 米", center_lat - 1000.0 / 111000.0, center_lon, 6);
    
    printf("中心坐标: (%.6f, %.6f)\n", center_lat, center_lon);
    printf("搜索半径: %.0f 米\n", radius);
    printf("选择精度: %d 位\n\n", geohash_precision_for_radius_at_lat(radius, center_lat));
    
    printf("已录入商家:\n");
    for (i = 0; i < db.count; i++) {
        double dist = haversine_distance(center_lat, center_lon,
                                          db.merchants[i].latitude,
                                          db.merchants[i].longitude);
        printf("  %s: 编码=%s, 距离中心=%.1f 米\n",
               db.merchants[i].name, db.merchants[i].geohash, dist);
    }
    
    printf("\n搜索结果 (%.0f 米范围内):\n", radius);
    result_count = search_merchants_by_radius(&db, center_lat, center_lon, radius, results, MAX_SEARCH_RESULTS);
    
    printf("  找到 %d 个商家:\n", result_count);
    for (i = 0; i < result_count; i++) {
        double dist = haversine_distance(center_lat, center_lon,
                                          results[i].latitude,
                                          results[i].longitude);
        printf("    %s: 距离=%.1f 米\n", results[i].name, dist);
    }
    
    printf("\n预期应该找到: 中心、正东500米、正南500米、正西500米、正北500米 (共5个)\n");
}

int main(void)
{
    printf("========================================\n");
    printf("  高纬度问题修复验证\n");
    printf("========================================\n");
    
    test_precision_at_latitude();
    test_high_latitude_search();
    
    printf("\n========================================\n");
    printf("测试完成\n");
    printf("========================================\n");
    
    return 0;
}
