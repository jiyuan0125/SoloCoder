package com.example.proxy.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Route {
    
    private String id;
    private String path;
    private RouteType type;
    private List<Backend> backends;
    
    public enum RouteType {
        EXACT,
        PREFIX
    }
}
