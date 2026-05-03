#include "search.h"
#include <string.h>
#include <math.h>

#ifndef M_PI
#define M_PI 3.14159265358979323846
#endif

#define EARTH_RADIUS 6371000.0

static int is_coverage_sufficient(int precision, double radius_meters, double center_lat)
{
    int total_bits = precision * 5;
    int lon_bits = (total_bits + 1) / 2;
    int lat_bits = total_bits / 2;
    
    double lon_deg_per_cell = 360.0 / (1LL << lon_bits);
    double lat_deg_per_cell = 180.0 / (1LL << lat_bits);
    
    double lat_rad = fabs(center_lat) * M_PI / 180.0;
    double cos_lat = cos(lat_rad);
    if (cos_lat < 0.0001) cos_lat = 0.0001;
    
    double km_per_deg_lat = 111.0;
    double km_per_deg_lon = 111.0 * cos_lat;
    
    double lon_km_per_cell = lon_deg_per_cell * km_per_deg_lon;
    double lat_km_per_cell = lat_deg_per_cell * km_per_deg_lat;
    
    double radius_km = radius_meters / 1000.0;
    
    double lon_3x3_coverage = 3 * lon_km_per_cell;
    double lat_3x3_coverage = 3 * lat_km_per_cell;
    
    double required = 2.5 * radius_km;
    
    return (lon_3x3_coverage >= required) && (lat_3x3_coverage >= required);
}

int geohash_precision_for_radius_at_lat(double radius_meters, double center_lat)
{
    int precision;
    
    for (precision = GEOHASH_PRECISION_MAX; precision >= 1; precision--) {
        if (is_coverage_sufficient(precision, radius_meters, center_lat)) {
            return precision;
        }
    }
    
    return 1;
}

int geohash_precision_for_radius(double radius_meters)
{
    return geohash_precision_for_radius_at_lat(radius_meters, 0.0);
}

void merchant_db_init(MerchantDatabase *db)
{
    db->count = 0;
    memset(db->merchants, 0, sizeof(db->merchants));
}

int merchant_add(MerchantDatabase *db, const char *name, double lat, double lon, int precision)
{
    if (db->count >= MAX_MERCHANTS) {
        return -1;
    }
    
    if (precision <= 0 || precision > GEOHASH_PRECISION_MAX) {
        precision = 6;
    }
    
    Merchant *m = &db->merchants[db->count];
    strncpy(m->name, name, sizeof(m->name) - 1);
    m->name[sizeof(m->name) - 1] = '\0';
    m->latitude = lat;
    m->longitude = lon;
    geohash_encode(lat, lon, precision, m->geohash);
    
    db->count++;
    return 0;
}

int search_merchants_by_geohash(MerchantDatabase *db, const char *center_geohash, 
                                  Merchant results[], int max_results)
{
    char neighbors[9][GEOHASH_PRECISION_MAX + 1];
    int result_count = 0;
    int i, j;
    int hash_len = (int)strlen(center_geohash);
    
    geohash_get_neighbors(center_geohash, neighbors);
    
    for (i = 0; i < db->count && result_count < max_results; i++) {
        int match = 0;
        const Merchant *m = &db->merchants[i];
        
        for (j = 0; j < 9; j++) {
            if (strncmp(m->geohash, neighbors[j], hash_len) == 0) {
                match = 1;
                break;
            }
        }
        
        if (match) {
            results[result_count++] = *m;
        }
    }
    
    return result_count;
}

double haversine_distance(double lat1, double lon1, double lat2, double lon2)
{
    double lat1_rad = lat1 * M_PI / 180.0;
    double lat2_rad = lat2 * M_PI / 180.0;
    double delta_lat = (lat2 - lat1) * M_PI / 180.0;
    double delta_lon = (lon2 - lon1) * M_PI / 180.0;
    
    double a = sin(delta_lat / 2) * sin(delta_lat / 2) +
               cos(lat1_rad) * cos(lat2_rad) *
               sin(delta_lon / 2) * sin(delta_lon / 2);
    
    double c = 2 * atan2(sqrt(a), sqrt(1 - a));
    
    return EARTH_RADIUS * c;
}

int search_merchants_by_radius(MerchantDatabase *db, double center_lat, double center_lon,
                                double radius_meters, Merchant results[], int max_results)
{
    char center_hash[GEOHASH_PRECISION_MAX + 1];
    int precision;
    Merchant geo_results[MAX_SEARCH_RESULTS];
    int geo_count;
    int result_count = 0;
    int i;
    
    precision = geohash_precision_for_radius_at_lat(radius_meters, center_lat);
    geohash_encode(center_lat, center_lon, precision, center_hash);
    
    geo_count = search_merchants_by_geohash(db, center_hash, geo_results, MAX_SEARCH_RESULTS);
    
    for (i = 0; i < geo_count && result_count < max_results; i++) {
        double distance = haversine_distance(center_lat, center_lon,
                                               geo_results[i].latitude,
                                               geo_results[i].longitude);
        
        if (distance <= radius_meters) {
            results[result_count++] = geo_results[i];
        }
    }
    
    return result_count;
}
