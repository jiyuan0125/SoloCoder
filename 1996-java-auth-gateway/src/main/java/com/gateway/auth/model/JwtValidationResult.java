package com.gateway.auth.model;

public class JwtValidationResult {

    public enum Status {
        VALID,
        EXPIRED,
        INVALID
    }

    private final Status status;
    private final String clientId;
    private final String role;
    private final String message;

    private JwtValidationResult(Status status, String clientId, String role, String message) {
        this.status = status;
        this.clientId = clientId;
        this.role = role;
        this.message = message;
    }

    public static JwtValidationResult valid(String clientId, String role) {
        return new JwtValidationResult(Status.VALID, clientId, role, null);
    }

    public static JwtValidationResult expired(String message) {
        return new JwtValidationResult(Status.EXPIRED, null, null, message);
    }

    public static JwtValidationResult invalid(String message) {
        return new JwtValidationResult(Status.INVALID, null, null, message);
    }

    public Status getStatus() {
        return status;
    }

    public String getClientId() {
        return clientId;
    }

    public String getRole() {
        return role;
    }

    public String getMessage() {
        return message;
    }
}
