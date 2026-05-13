package com.example.converter.exception;

public class MappingNotFoundException extends RuntimeException {
    public MappingNotFoundException(String message) {
        super(message);
    }
}
