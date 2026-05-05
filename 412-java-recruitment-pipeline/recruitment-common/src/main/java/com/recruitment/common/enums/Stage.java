package com.recruitment.common.enums;

public enum Stage {
    SCREENING("简历筛选"),
    PHONE_INTERVIEW("电话面试"),
    ONSITE_INTERVIEW("现场面试"),
    OFFER("发offer"),
    ONBOARD("入职确认");

    private final String description;

    Stage(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }

    public Stage next() {
        Stage[] stages = values();
        int nextIndex = ordinal() + 1;
        if (nextIndex >= stages.length) {
            return this;
        }
        return stages[nextIndex];
    }

    public Stage previous() {
        int prevIndex = ordinal() - 1;
        if (prevIndex < 0) {
            return this;
        }
        return values()[prevIndex];
    }

    public boolean canTransitionTo(Stage target) {
        return target == next() || target == previous();
    }

    public static Stage fromName(String name) {
        for (Stage stage : values()) {
            if (stage.name().equalsIgnoreCase(name) || stage.getDescription().equals(name)) {
                return stage;
            }
        }
        return null;
    }
}
