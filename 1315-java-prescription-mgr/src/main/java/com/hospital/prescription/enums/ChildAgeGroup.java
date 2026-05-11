package com.hospital.prescription.enums;

public enum ChildAgeGroup {
    NEWBORN("新生儿", 0, 28),
    INFANT("婴幼儿", 29, 365),
    PRESCHOOL("学龄前", 366, 2190),
    SCHOOL_AGE("学龄期", 2191, 4380),
    ADOLESCENT("青少年", 4381, 6570);

    private final String description;
    private final int minDays;
    private final int maxDays;

    ChildAgeGroup(String description, int minDays, int maxDays) {
        this.description = description;
        this.minDays = minDays;
        this.maxDays = maxDays;
    }

    public String getDescription() {
        return description;
    }

    public static ChildAgeGroup fromAgeInDays(int days) {
        for (ChildAgeGroup group : values()) {
            if (days >= group.minDays && days <= group.maxDays) {
                return group;
            }
        }
        return null;
    }
}
