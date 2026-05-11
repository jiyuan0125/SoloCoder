package com.hotelbooking.exception;

public class BookingException extends RuntimeException {
    private final String errorCode;

    public BookingException(String message) {
        super(message);
        this.errorCode = "BOOKING_ERROR";
    }

    public BookingException(String errorCode, String message) {
        super(message);
        this.errorCode = errorCode;
    }

    public String getErrorCode() {
        return errorCode;
    }
}
