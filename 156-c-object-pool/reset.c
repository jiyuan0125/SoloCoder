#include "reset.h"
#include <string.h>
#include <time.h>

void monster_reset(Monster *monster) {
    memset(monster, 0, sizeof(Monster));
    monster->state = MONSTER_STATE_IDLE;
}

void monster_init(Monster *monster, unsigned int pool_index) {
    monster_reset(monster);
    monster->pool_index = pool_index;
}

int monster_is_leaked(Monster *monster, unsigned long long current_time, unsigned long long timeout_seconds) {
    if (monster->state == MONSTER_STATE_IDLE) {
        return 0;
    }
    if (monster->borrow_timestamp == 0) {
        return 0;
    }
    if (current_time > monster->borrow_timestamp + timeout_seconds) {
        return 1;
    }
    return 0;
}

unsigned long long get_current_timestamp_seconds(void) {
    return (unsigned long long)time(NULL);
}
