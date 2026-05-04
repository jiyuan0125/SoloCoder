#ifndef RESET_H
#define RESET_H

#include "monster.h"

void monster_reset(Monster *monster);
void monster_init(Monster *monster, unsigned int pool_index);
int monster_is_leaked(Monster *monster, unsigned long long current_time, unsigned long long timeout_seconds);
unsigned long long get_current_timestamp_seconds(void);

#endif
