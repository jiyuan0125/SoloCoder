package com.loadbalancer.model;

public enum NodeStatus {
    NEW_REGISTERED,
    NORMAL_SERVICE,
    SUSPECTED_FAILURE,
    OFFLINE
}