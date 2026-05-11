package com.example.groupbuy.exception;

import lombok.Getter;

@Getter
public class GroupBuyException extends RuntimeException {
    private final int code;
    
    public GroupBuyException(String message) {
        super(message);
        this.code = 500;
    }
    
    public GroupBuyException(int code, String message) {
        super(message);
        this.code = code;
    }
}
