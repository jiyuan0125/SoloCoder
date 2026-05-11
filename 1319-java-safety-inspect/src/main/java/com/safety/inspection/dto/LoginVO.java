package com.safety.inspection.dto;

import lombok.Data;

@Data
public class LoginVO {
    private Long userId;
    private String username;
    private String realName;
    private String token;
    private String role;
}
