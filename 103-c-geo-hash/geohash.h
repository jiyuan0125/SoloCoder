#ifndef GEOHASH_H
#define GEOHASH_H

#include <stdint.h>

#define GEOHASH_PRECISION_MAX 12
#define GEOHASH_LAT_MIN -90.0
#define GEOHASH_LAT_MAX 90.0
#define GEOHASH_LON_MIN -180.0
#define GEOHASH_LON_MAX 180.0

typedef struct {
    double lat_min;
    double lat_max;
    double lon_min;
    double lon_max;
} GeoHashBounds;

typedef struct {
    double latitude;
    double longitude;
} GeoPoint;

void geohash_encode(double lat, double lon, int precision, char *result);
GeoHashBounds geohash_decode_bounds(const char *hash);
GeoPoint geohash_decode(const char *hash);
void geohash_get_neighbors(const char *hash, char neighbors[9][GEOHASH_PRECISION_MAX + 1]);

#endif
