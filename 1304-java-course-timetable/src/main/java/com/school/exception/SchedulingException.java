package com.school.exception;

public class SchedulingException extends RuntimeException {
    private final ConflictType conflictType;

    public SchedulingException(String message, ConflictType conflictType) {
        super(message);
        this.conflictType = conflictType;
    }

    public ConflictType getConflictType() {
        return conflictType;
    }

    public enum ConflictType {
        TEACHER_CONFLICT,
        CLASSROOM_CONFLICT,
        CAPACITY_EXCEEDED,
        TIMESLOT_INVALID
    }
}
