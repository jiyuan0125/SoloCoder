package com.example.lightweightqueue.model;

public enum DeliveryFailureType {
    TIMEOUT,
    FORMAT_ERROR,
    CONSUMER_NOT_EXIST,
    MAX_RETRY_EXCEEDED
}
