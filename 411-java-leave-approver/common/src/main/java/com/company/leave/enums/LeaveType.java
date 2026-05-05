package com.company.leave.enums;

public enum LeaveType {
    ANNUAL("年假", false, false),
    SICK("病假", true, true),
    PERSONAL("事假", true, false),
    MARRIAGE("婚假", false, true),
    MATERNITY("产假", false, true),
    PATERNITY("陪产假", false, false);

    private final String description;
    private final boolean countWeekends;
    private final boolean requiresAttachment;

    LeaveType(String description, boolean countWeekends, boolean requiresAttachment) {
        this.description = description;
        this.countWeekends = countWeekends;
        this.requiresAttachment = requiresAttachment;
    }

    public String getDescription() {
        return description;
    }

    public boolean isCountWeekends() {
        return countWeekends;
    }

    public boolean isRequiresAttachment() {
        return requiresAttachment;
    }
}
