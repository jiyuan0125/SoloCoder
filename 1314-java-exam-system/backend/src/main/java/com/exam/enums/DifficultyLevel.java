package com.exam.enums;

public enum DifficultyLevel {
    LEVEL_1(1, "1星"),
    LEVEL_2(2, "2星"),
    LEVEL_3(3, "3星"),
    LEVEL_4(4, "4星"),
    LEVEL_5(5, "5星");

    private final int level;
    private final String description;

    DifficultyLevel(int level, String description) {
        this.level = level;
        this.description = description;
    }

    public int getLevel() {
        return level;
    }

    public String getDescription() {
        return description;
    }

    public static DifficultyLevel fromLevel(int level) {
        for (DifficultyLevel dl : values()) {
            if (dl.level == level) {
                return dl;
            }
        }
        throw new IllegalArgumentException("无效的难度等级: " + level);
    }
}
