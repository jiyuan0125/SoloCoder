package com.loadbalancer.dto;

import lombok.Data;

import java.util.ArrayList;
import java.util.List;

@Data
public class NodeRegistrationRequest {
    private String host;
    private int port;
    private int weight = 1;
    private List<String> tags = new ArrayList<>();
}
