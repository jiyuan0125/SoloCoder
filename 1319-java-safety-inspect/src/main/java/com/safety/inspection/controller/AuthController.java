package com.safety.inspection.controller;

import com.safety.inspection.common.Result;
import com.safety.inspection.dto.LoginDTO;
import com.safety.inspection.dto.LoginVO;
import com.safety.inspection.service.UserService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/auth")
@RequiredArgsConstructor
public class AuthController {

    private final UserService userService;

    @PostMapping("/login")
    public Result<LoginVO> login(@RequestBody @Valid LoginDTO loginDTO) {
        return Result.success(userService.login(loginDTO));
    }
}
