#ifndef MONSTER_H
#define MONSTER_H

typedef enum MonsterState {
    MONSTER_STATE_IDLE = 0,
    MONSTER_STATE_ALIVE,
    MONSTER_STATE_DEAD
} MonsterState;

typedef struct Monster {
    unsigned int monster_id;
    unsigned int template_id;
    float pos_x;
    float pos_y;
    float pos_z;
    float rotation;
    int max_hp;
    int current_hp;
    int level;
    MonsterState state;
    unsigned long long borrow_timestamp;
    unsigned int pool_index;
} Monster;

#endif
