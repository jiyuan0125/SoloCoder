#include "checkin_bitmap.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

static inline int count_trailing_ones(uint8_t byte) {
    if (byte == 0xFF) return 8;
    if (byte == 0x00) return 0;
    int count = 0;
    while ((byte & 0x01) && count < 8) {
        count++;
        byte >>= 1;
    }
    return count;
}

static inline int count_trailing_ones_from(uint8_t byte, int start_bit) {
    int count = 0;
    uint8_t mask = 1 << start_bit;
    while ((byte & mask) && count < (8 - start_bit)) {
        count++;
        mask <<= 1;
    }
    return count;
}

static inline int count_leading_ones(uint8_t byte) {
    if (byte == 0xFF) return 8;
    if (byte == 0x00) return 0;
    int count = 0;
    while ((byte & 0x80) && count < 8) {
        count++;
        byte <<= 1;
    }
    return count;
}

static inline int count_leading_ones_from(uint8_t byte, int start_bit) {
    int count = 0;
    uint8_t mask = 1 << start_bit;
    while (start_bit >= 0 && (byte & mask)) {
        count++;
        start_bit--;
        mask >>= 1;
    }
    return count;
}

static inline int popcount_byte(uint8_t byte) {
    static const uint8_t lookup[] = {
        0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4,
        1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
        1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
        2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
        1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
        2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
        2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
        3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
        1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
        2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
        2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
        3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
        2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
        3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
        3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
        4, 5, 5, 6, 5, 6, 6, 7, 5, 6, 6, 7, 6, 7, 7, 8
    };
    return lookup[byte];
}

void year_bitmap_init(YearBitmap *yb, int year) {
    yb->year = year;
    yb->total_days = days_in_year(year);
    yb->size_bytes = (yb->total_days + BITS_PER_BYTE - 1) / BITS_PER_BYTE;
    yb->bits = (uint8_t *)calloc(yb->size_bytes, sizeof(uint8_t));
}

void year_bitmap_free(YearBitmap *yb) {
    if (yb->bits) {
        free(yb->bits);
        yb->bits = NULL;
    }
    yb->size_bytes = 0;
    yb->total_days = 0;
}

bool year_bitmap_set(YearBitmap *yb, int day_of_year) {
    if (day_of_year < 1 || day_of_year > yb->total_days) {
        return false;
    }
    int byte_idx = (day_of_year - 1) / BITS_PER_BYTE;
    int bit_idx = (day_of_year - 1) % BITS_PER_BYTE;
    yb->bits[byte_idx] |= (1 << bit_idx);
    return true;
}

bool year_bitmap_clear(YearBitmap *yb, int day_of_year) {
    if (day_of_year < 1 || day_of_year > yb->total_days) {
        return false;
    }
    int byte_idx = (day_of_year - 1) / BITS_PER_BYTE;
    int bit_idx = (day_of_year - 1) % BITS_PER_BYTE;
    yb->bits[byte_idx] &= ~(1 << bit_idx);
    return true;
}

bool year_bitmap_check(const YearBitmap *yb, int day_of_year) {
    if (day_of_year < 1 || day_of_year > yb->total_days) {
        return false;
    }
    int byte_idx = (day_of_year - 1) / BITS_PER_BYTE;
    int bit_idx = (day_of_year - 1) % BITS_PER_BYTE;
    return (yb->bits[byte_idx] & (1 << bit_idx)) != 0;
}

int year_bitmap_count_range(const YearBitmap *yb, int start_day, int end_day) {
    if (start_day < 1) start_day = 1;
    if (end_day > yb->total_days) end_day = yb->total_days;
    if (start_day > end_day) return 0;

    int count = 0;
    int start_byte = (start_day - 1) / BITS_PER_BYTE;
    int end_byte = (end_day - 1) / BITS_PER_BYTE;
    int start_bit = (start_day - 1) % BITS_PER_BYTE;
    int end_bit = (end_day - 1) % BITS_PER_BYTE;

    if (start_byte == end_byte) {
        uint8_t mask = ((1 << (end_bit - start_bit + 1)) - 1) << start_bit;
        return popcount_byte(yb->bits[start_byte] & mask);
    }

    uint8_t start_mask = 0xFF << start_bit;
    count += popcount_byte(yb->bits[start_byte] & start_mask);

    for (int i = start_byte + 1; i < end_byte; i++) {
        count += popcount_byte(yb->bits[i]);
    }

    uint8_t end_mask = (1 << (end_bit + 1)) - 1;
    count += popcount_byte(yb->bits[end_byte] & end_mask);

    return count;
}

int year_bitmap_longest_streak_range(const YearBitmap *yb, int start_day, int end_day) {
    if (start_day < 1) start_day = 1;
    if (end_day > yb->total_days) end_day = yb->total_days;
    if (start_day > end_day) return 0;

    int longest = 0;
    int current = 0;

    for (int day = start_day; day <= end_day; day++) {
        if (year_bitmap_check(yb, day)) {
            current++;
            if (current > longest) {
                longest = current;
            }
        } else {
            current = 0;
        }
    }

    return longest;
}

int year_bitmap_get_dates_in_range(const YearBitmap *yb, int start_day, int end_day, 
                                    int *dates, int max_dates) {
    if (start_day < 1) start_day = 1;
    if (end_day > yb->total_days) end_day = yb->total_days;
    if (start_day > end_day || max_dates <= 0) return 0;

    int count = 0;
    for (int day = start_day; day <= end_day && count < max_dates; day++) {
        if (year_bitmap_check(yb, day)) {
            dates[count++] = day;
        }
    }
    return count;
}

int year_bitmap_trailing_ones(const YearBitmap *yb) {
    int last_byte_idx = yb->size_bytes - 1;
    int bits_in_last_byte = yb->total_days % BITS_PER_BYTE;
    if (bits_in_last_byte == 0) bits_in_last_byte = 8;

    int trailing = 0;
    
    uint8_t last_byte = yb->bits[last_byte_idx];
    int highest_bit = bits_in_last_byte - 1;
    
    int ones = count_leading_ones_from(last_byte, highest_bit);
    trailing += ones;

    if (ones < bits_in_last_byte) {
        return trailing;
    }

    for (int i = last_byte_idx - 1; i >= 0; i--) {
        ones = count_leading_ones(yb->bits[i]);
        trailing += ones;
        if (ones < 8) {
            break;
        }
    }

    return trailing;
}

int year_bitmap_trailing_ones_from(const YearBitmap *yb, int from_day) {
    if (from_day < 1 || from_day > yb->total_days) {
        return 0;
    }

    int byte_idx = (from_day - 1) / BITS_PER_BYTE;
    int bit_idx = (from_day - 1) % BITS_PER_BYTE;

    int trailing = 0;

    uint8_t byte = yb->bits[byte_idx];
    int ones = count_leading_ones_from(byte, bit_idx);
    trailing += ones;

    if (ones <= bit_idx) {
        return trailing;
    }

    for (int i = byte_idx - 1; i >= 0; i--) {
        ones = count_leading_ones(yb->bits[i]);
        trailing += ones;

        if (ones < 8) {
            break;
        }
    }

    return trailing;
}

uint64_t year_bitmap_get_uint64(const YearBitmap *yb, int index) {
    if (index < 0 || index * 8 >= yb->size_bytes) {
        return 0;
    }
    uint64_t result = 0;
    int bytes_to_read = 8;
    if (index * 8 + 8 > yb->size_bytes) {
        bytes_to_read = yb->size_bytes - index * 8;
    }
    for (int i = 0; i < bytes_to_read; i++) {
        result |= (uint64_t)yb->bits[index * 8 + i] << (i * 8);
    }
    return result;
}

void user_bitmap_init(UserBitmap *ub) {
    ub->years = NULL;
    ub->year_count = 0;
    ub->year_capacity = 0;
}

void user_bitmap_free(UserBitmap *ub) {
    for (int i = 0; i < ub->year_count; i++) {
        year_bitmap_free(&ub->years[i]);
    }
    free(ub->years);
    ub->years = NULL;
    ub->year_count = 0;
    ub->year_capacity = 0;
}

YearBitmap* user_bitmap_get_year(UserBitmap *ub, int year) {
    for (int i = 0; i < ub->year_count; i++) {
        if (ub->years[i].year == year) {
            return &ub->years[i];
        }
    }
    return NULL;
}

YearBitmap* user_bitmap_get_or_create_year(UserBitmap *ub, int year) {
    YearBitmap *existing = user_bitmap_get_year(ub, year);
    if (existing) {
        return existing;
    }

    if (ub->year_count >= ub->year_capacity) {
        int new_capacity = ub->year_capacity == 0 ? 4 : ub->year_capacity * 2;
        YearBitmap *new_years = (YearBitmap *)realloc(ub->years, 
                                                        new_capacity * sizeof(YearBitmap));
        if (!new_years) {
            return NULL;
        }
        ub->years = new_years;
        ub->year_capacity = new_capacity;
    }

    YearBitmap *new_year = &ub->years[ub->year_count];
    year_bitmap_init(new_year, year);
    ub->year_count++;
    return new_year;
}
