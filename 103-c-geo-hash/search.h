#ifndef SEARCH_H
#define SEARCH_H

#include "geohash.h"
#include "base32.h"

#define MAX_MERCHANTS 1000
#define MAX_SEARCH_RESULTS 100

typedef struct {
    char name[64];
    double latitude;
    double longitude;
    char geohash[GEOHASH_PRECISION_MAX + 1];
} Merchant;

typedef struct {
    Merchant merchants[MAX_MERCHANTS];
    int count;
} MerchantDatabase;

int geohash_precision_for_radius(double radius_meters);
void merchant_db_init(MerchantDatabase *db);
int merchant_add(MerchantDatabase *db, const char *name, double lat, double lon, int precision);
int search_merchants_by_geohash(MerchantDatabase *db, const char *center_geohash, 
                                  Merchant results[], int max_results);
int search_merchants_by_radius(MerchantDatabase *db, double center_lat, double center_lon,
                                double radius_meters, Merchant results[], int max_results);
double haversine_distance(double lat1, double lon1, double lat2, double lon2);

#endif
