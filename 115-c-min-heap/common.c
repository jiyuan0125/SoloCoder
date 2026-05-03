#include "common.h"

int get_priority_time_limit(int priority) {
    switch (priority) {
        case 1: return PRIORITY_1_TIME_LIMIT;
        case 2: return PRIORITY_2_TIME_LIMIT;
        case 3: return PRIORITY_3_TIME_LIMIT;
        case 4: return PRIORITY_4_TIME_LIMIT;
        case 5: return PRIORITY_5_TIME_LIMIT;
        default: return PRIORITY_5_TIME_LIMIT;
    }
}

const char* get_priority_description(int priority) {
    switch (priority) {
        case 1: return "濒危";
        case 2: return "危重";
        case 3: return "紧急";
        case 4: return "较急";
        case 5: return "非急";
        default: return "未知";
    }
}
