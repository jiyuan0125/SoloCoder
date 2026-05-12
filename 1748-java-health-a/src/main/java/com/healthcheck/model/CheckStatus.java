package com.healthcheck.model;

public enum CheckStatus {
    NORMAL,
    WARNING,
    ERROR,
    UNKNOWN;

    public static CheckStatus nextWorse(CheckStatus current) {
        return switch (current) {
            case NORMAL -> WARNING;
            case WARNING -> ERROR;
            case ERROR, UNKNOWN -> ERROR;
        };
    }

    public static CheckStatus fromBoolean(boolean success) {
        return success ? NORMAL : ERROR;
    }
}
