#include "geohash.h"
#include "base32.h"
#include <string.h>
#include <stdlib.h>

static uint64_t encode_interval(double value, double min, double max, int bits)
{
    uint64_t result = 0;
    double mid;
    int i;
    
    for (i = 0; i < bits; i++) {
        mid = (min + max) / 2;
        result <<= 1;
        if (value >= mid) {
            result |= 1;
            min = mid;
        } else {
            max = mid;
        }
    }
    
    return result;
}

static void decode_interval(uint64_t bits, int num_bits, double *min, double *max)
{
    double mid;
    int i;
    
    for (i = 0; i < num_bits; i++) {
        mid = (*min + *max) / 2;
        if ((bits >> (num_bits - i - 1)) & 1) {
            *min = mid;
        } else {
            *max = mid;
        }
    }
}

void geohash_encode(double lat, double lon, int precision, char *result)
{
    uint64_t combined = 0;
    int total_bits;
    int lat_bits, lon_bits;
    int i;
    uint64_t lat_code, lon_code;
    double lat_min = GEOHASH_LAT_MIN;
    double lat_max = GEOHASH_LAT_MAX;
    double lon_min = GEOHASH_LON_MIN;
    double lon_max = GEOHASH_LON_MAX;
    
    if (precision <= 0 || precision > GEOHASH_PRECISION_MAX) {
        precision = 5;
    }
    
    total_bits = precision * 5;
    lon_bits = (total_bits + 1) / 2;
    lat_bits = total_bits / 2;
    
    lon_code = encode_interval(lon, lon_min, lon_max, lon_bits);
    lat_code = encode_interval(lat, lat_min, lat_max, lat_bits);
    
    for (i = 0; i < total_bits; i++) {
        combined <<= 1;
        if (i % 2 == 0) {
            int bit_pos = lon_bits - 1 - (i / 2);
            if (bit_pos >= 0) {
                combined |= (lon_code >> bit_pos) & 1;
            }
        } else {
            int bit_pos = lat_bits - 1 - ((i - 1) / 2);
            if (bit_pos >= 0) {
                combined |= (lat_code >> bit_pos) & 1;
            }
        }
    }
    
    base32_encode(combined, precision, result);
}

GeoHashBounds geohash_decode_bounds(const char *hash)
{
    GeoHashBounds bounds;
    uint64_t combined;
    int precision = (int)strlen(hash);
    int total_bits;
    int lat_bits, lon_bits;
    uint64_t lat_code = 0, lon_code = 0;
    int i;
    
    if (precision <= 0 || precision > GEOHASH_PRECISION_MAX) {
        bounds.lat_min = GEOHASH_LAT_MIN;
        bounds.lat_max = GEOHASH_LAT_MAX;
        bounds.lon_min = GEOHASH_LON_MIN;
        bounds.lon_max = GEOHASH_LON_MAX;
        return bounds;
    }
    
    total_bits = precision * 5;
    lon_bits = (total_bits + 1) / 2;
    lat_bits = total_bits / 2;
    
    combined = base32_decode(hash, precision);
    
    for (i = 0; i < total_bits; i++) {
        uint64_t bit = (combined >> (total_bits - i - 1)) & 1;
        if (i % 2 == 0) {
            lon_code <<= 1;
            lon_code |= bit;
        } else {
            lat_code <<= 1;
            lat_code |= bit;
        }
    }
    
    bounds.lat_min = GEOHASH_LAT_MIN;
    bounds.lat_max = GEOHASH_LAT_MAX;
    bounds.lon_min = GEOHASH_LON_MIN;
    bounds.lon_max = GEOHASH_LON_MAX;
    
    decode_interval(lon_code, lon_bits, &bounds.lon_min, &bounds.lon_max);
    decode_interval(lat_code, lat_bits, &bounds.lat_min, &bounds.lat_max);
    
    return bounds;
}

GeoPoint geohash_decode(const char *hash)
{
    GeoPoint point;
    GeoHashBounds bounds = geohash_decode_bounds(hash);
    
    point.latitude = (bounds.lat_min + bounds.lat_max) / 2;
    point.longitude = (bounds.lon_min + bounds.lon_max) / 2;
    
    return point;
}

static void get_adjacent(const char *hash, int dir_lat, int dir_lon, char *result)
{
    GeoHashBounds bounds = geohash_decode_bounds(hash);
    int precision = (int)strlen(hash);
    double lat_center = (bounds.lat_min + bounds.lat_max) / 2;
    double lon_center = (bounds.lon_min + bounds.lon_max) / 2;
    double lat_step = bounds.lat_max - bounds.lat_min;
    double lon_step = bounds.lon_max - bounds.lon_min;
    
    double new_lat = lat_center + dir_lat * lat_step;
    double new_lon = lon_center + dir_lon * lon_step;
    
    if (new_lat > GEOHASH_LAT_MAX) new_lat = GEOHASH_LAT_MAX;
    if (new_lat < GEOHASH_LAT_MIN) new_lat = GEOHASH_LAT_MIN;
    if (new_lon > GEOHASH_LON_MAX) new_lon = GEOHASH_LON_MAX;
    if (new_lon < GEOHASH_LON_MIN) new_lon = GEOHASH_LON_MIN;
    
    geohash_encode(new_lat, new_lon, precision, result);
}

void geohash_get_neighbors(const char *hash, char neighbors[9][GEOHASH_PRECISION_MAX + 1])
{
    int idx = 0;
    int dlat, dlon;
    int hash_len = (int)strlen(hash);
    
    for (dlat = 1; dlat >= -1; dlat--) {
        for (dlon = -1; dlon <= 1; dlon++) {
            if (dlat == 0 && dlon == 0) {
                strncpy(neighbors[idx], hash, hash_len);
                neighbors[idx][hash_len] = '\0';
            } else {
                get_adjacent(hash, dlat, dlon, neighbors[idx]);
            }
            idx++;
        }
    }
}
