#include "search.h"
#include <string.h>
#include <math.h>

#ifndef M_PI
#define M_PI 3.14159265358979323846
#endif

#define EARTH_RADIUS 6371000.0

int geohash_precision_for_radius(double radius_meters)
{
    if (radius_meters <= 0) {
        return 6;
    }
    
    if (radius_meters <= 5) return 9;
    if (radius_meters <= 20) return 8;
    if (radius_meters <= 150) return 7;
    if (radius_meters <= 1200) return 6;
    if (radius_meters <= 5000) return 5;
    if (radius_meters <= 39000) return 4;
    if (radius_meters <= 156000) return 3;
    if (radius_meters <= 1250000) return 2;
    return 1;
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
    
    precision = geohash_precision_for_radius(radius_meters);
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
