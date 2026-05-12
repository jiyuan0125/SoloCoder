package com.example.circuitbreaker.core;

import java.net.ConnectException;
import java.net.SocketTimeoutException;
import java.util.Arrays;
import java.util.List;

public class FailureClassifier {

    private static final List<Class<? extends Exception>> CONNECTION_EXCEPTIONS = Arrays.asList(
            ConnectException.class,
            SocketTimeoutException.class,
            java.util.concurrent.TimeoutException.class,
            java.io.InterruptedIOException.class
    );

    public static boolean shouldCountAsFailure(Throwable exception) {
        if (exception == null) {
            return false;
        }

        for (Class<? extends Exception> exClass : CONNECTION_EXCEPTIONS) {
            if (exClass.isInstance(exception) || isCausedBy(exception, exClass)) {
                return true;
            }
        }

        if (exception instanceof org.springframework.web.client.HttpServerErrorException) {
            return true;
        }

        return false;
    }

    public static boolean shouldCountAsFailure(int httpStatusCode) {
        return httpStatusCode >= 500;
    }

    public static boolean isClientError(int httpStatusCode) {
        return httpStatusCode >= 400 && httpStatusCode < 500;
    }

    private static boolean isCausedBy(Throwable exception, Class<? extends Exception> causeClass) {
        Throwable cause = exception.getCause();
        while (cause != null) {
            if (causeClass.isInstance(cause)) {
                return true;
            }
            cause = cause.getCause();
        }
        return false;
    }
}
