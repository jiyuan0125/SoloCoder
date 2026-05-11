package com.hotelbooking.dto;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class JwtResponse {
    private String token;
    private String type;
    private String username;
    private String role;
    private String name;
}
