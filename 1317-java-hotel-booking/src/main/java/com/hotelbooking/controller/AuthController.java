package com.hotelbooking.controller;

import com.hotelbooking.dto.JwtResponse;
import com.hotelbooking.dto.LoginRequest;
import com.hotelbooking.model.entity.User;
import com.hotelbooking.repository.UserRepository;
import com.hotelbooking.security.JwtUtil;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.web.bind.annotation.*;

import java.util.Optional;

@RestController
@RequestMapping("/api/auth")
@RequiredArgsConstructor
public class AuthController {

    private final AuthenticationManager authenticationManager;
    private final JwtUtil jwtUtil;
    private final UserRepository userRepository;

    @PostMapping("/login")
    public ResponseEntity<JwtResponse> login(@Valid @RequestBody LoginRequest request) {
        Authentication authentication = authenticationManager.authenticate(
                new UsernamePasswordAuthenticationToken(request.getUsername(), request.getPassword())
        );

        UserDetails userDetails = (UserDetails) authentication.getPrincipal();
        String jwt = jwtUtil.generateToken(userDetails);

        Optional<User> user = userRepository.findByUsername(request.getUsername());
        String role = user.map(u -> u.getRole().name()).orElse("CUSTOMER");
        String name = user.map(User::getName).orElse(request.getUsername());

        return ResponseEntity.ok(JwtResponse.builder()
                .token(jwt)
                .type("Bearer")
                .username(userDetails.getUsername())
                .role(role)
                .name(name)
                .build());
    }
}
