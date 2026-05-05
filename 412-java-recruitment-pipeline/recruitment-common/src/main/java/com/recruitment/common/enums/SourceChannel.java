package com.recruitment.common.enums;

public enum SourceChannel {
    BOSS("Boss直聘"),
    LAGOU("拉勾网"),
    ZHILIAN("智联招聘"),
    INTERNAL("内部推荐"),
    HEADHUNTER("猎头"),
    UNIVERSITY("校园招聘"),
    OTHER("其他");

    private final String description;

    SourceChannel(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }

    public static SourceChannel fromName(String name) {
        for (SourceChannel channel : values()) {
            if (channel.name().equalsIgnoreCase(name) || channel.getDescription().equals(name)) {
                return channel;
            }
        }
        return OTHER;
    }
}
