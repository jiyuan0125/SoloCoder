#include <stdio.h>
#include <string.h>
#include "geohash.h"
#include "base32.h"

static void test_geohash_decode_bounds(void)
{
    const char *hash = "wx4g0";
    GeoHashBounds bounds;
    
    printf("Test 1: geohash_decode_bounds(\"%s\")\n", hash);
    bounds = geohash_decode_bounds(hash);
    printf("  lat: [%.6f, %.6f]\n", bounds.lat_min, bounds.lat_max);
    printf("  lon: [%.6f, %.6f]\n", bounds.lon_min, bounds.lon_max);
}

static void test_geohash_encode(void)
{
    char result[GEOHASH_PRECISION_MAX + 1];
    double lat = 39.9042;
    double lon = 116.4074;
    
    printf("Test 2: geohash_encode(%.6f, %.6f, 5)\n", lat, lon);
    geohash_encode(lat, lon, 5, result);
    printf("  Result: %s\n", result);
}

static void test_adjacent_simple(void)
{
    const char *hash = "wx4g0";
    GeoHashBounds bounds;
    double lat_center, lon_center;
    double lat_step, lon_step;
    double new_lat, new_lon;
    char result[GEOHASH_PRECISION_MAX + 1];
    int precision = 5;
    
    printf("Test 3: Simulating get_adjacent\n");
    bounds = geohash_decode_bounds(hash);
    printf("  bounds decoded\n");
    
    lat_center = (bounds.lat_min + bounds.lat_max) / 2;
    lon_center = (bounds.lon_min + bounds.lon_max) / 2;
    lat_step = bounds.lat_max - bounds.lat_min;
    lon_step = bounds.lon_max - bounds.lon_min;
    
    printf("  center: (%.6f, %.6f)\n", lat_center, lon_center);
    printf("  step: (%.6f, %.6f)\n", lat_step, lon_step);
    
    new_lat = lat_center + 1 * lat_step;
    new_lon = lon_center + 1 * lon_step;
    
    printf("  new: (%.6f, %.6f)\n", new_lat, new_lon);
    
    printf("  calling geohash_encode...\n");
    geohash_encode(new_lat, new_lon, precision, result);
    printf("  result: %s\n", result);
}

int main(void)
{
    printf("=== Start of test ===\n\n");
    
    test_geohash_decode_bounds();
    printf("\n");
    
    test_geohash_encode();
    printf("\n");
    
    test_adjacent_simple();
    printf("\n");
    
    printf("=== End of test ===\n");
    return 0;
}
