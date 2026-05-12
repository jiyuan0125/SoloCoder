package com.example.retry.service;

import org.springframework.stereotype.Service;

@Service
public class ExternalPaymentService {

    public String executePayment(String operationId, String payload) {
        return "Payment successful for operation: " + operationId;
    }
}
