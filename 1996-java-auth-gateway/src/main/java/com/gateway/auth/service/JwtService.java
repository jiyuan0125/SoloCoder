package com.gateway.auth.service;

import com.gateway.auth.config.JwtConfig;
import com.gateway.auth.model.JwtValidationResult;
import io.jsonwebtoken.*;
import io.jsonwebtoken.security.Keys;
import io.jsonwebtoken.security.SignatureException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.nio.charset.StandardCharsets;
import java.security.Key;
import java.time.Instant;
import java.util.Arrays;
import java.util.Base64;
import java.util.List;

@Service
public class JwtService {

    private static final Logger logger = LoggerFactory.getLogger(JwtService.class);
    private static final List<String> ALLOWED_ROLES = Arrays.asList("admin", "service");

    private final JwtConfig jwtConfig;

    private Key primaryKey;
    private Key oldKey;

    public JwtService(JwtConfig jwtConfig) {
        this.jwtConfig = jwtConfig;
    }

    @PostConstruct
    public void init() {
        primaryKey = Keys.hmacShaKeyFor(jwtConfig.getPrimarySecret().getBytes(StandardCharsets.UTF_8));
        if (jwtConfig.getOldSecret() != null && !jwtConfig.getOldSecret().trim().isEmpty()) {
            oldKey = Keys.hmacShaKeyFor(jwtConfig.getOldSecret().getBytes(StandardCharsets.UTF_8));
        }
    }

    public JwtValidationResult validateToken(String token) {
        if (token == null || token.trim().isEmpty()) {
            return JwtValidationResult.invalid("Token is empty");
        }

        String[] parts = token.split("\\.");
        if (parts.length != 3) {
            return JwtValidationResult.invalid("Invalid token format");
        }

        Claims claims = null;
        boolean usedOldKey = false;
        Exception validationException = null;

        try {
            claims = parseClaims(token, primaryKey);
        } catch (ExpiredJwtException e) {
            return JwtValidationResult.expired("Token expired");
        } catch (JwtException e) {
            validationException = e;
        }

        if (claims == null && oldKey != null && isOldKeyValid()) {
            try {
                claims = parseClaims(token, oldKey);
                usedOldKey = true;
            } catch (ExpiredJwtException e) {
                return JwtValidationResult.expired("Token expired");
            } catch (JwtException e) {
                validationException = e;
            }
        }

        if (claims == null) {
            logger.debug("Token validation failed: {}", validationException != null ? validationException.getMessage() : "Unknown error");
            return JwtValidationResult.invalid("Invalid token");
        }

        String clientId = getStringClaim(claims, "client_id", "sub");
        String role = getStringClaim(claims, "role");

        if (clientId == null || clientId.trim().isEmpty()) {
            return JwtValidationResult.invalid("Missing client_id claim");
        }

        if (role == null || !ALLOWED_ROLES.contains(role)) {
            return JwtValidationResult.invalid("Invalid or missing role claim");
        }

        return JwtValidationResult.valid(clientId, role);
    }

    public boolean isOldKeyValid() {
        if (oldKey == null) {
            return false;
        }

        Instant now = Instant.now();
        if (jwtConfig.getComputedOldKeyExpireAt() != null) {
            return now.isBefore(jwtConfig.getComputedOldKeyExpireAt());
        }

        Instant oldKeyStartTime = now.minusSeconds(jwtConfig.getOldKeyExpireMinutes() * 60L);
        return true;
    }

    public Key getPrimaryKey() {
        return primaryKey;
    }

    public Key getOldKey() {
        return oldKey;
    }

    public Instant getOldKeyExpireTime() {
        if (oldKey == null) {
            return null;
        }
        if (jwtConfig.getComputedOldKeyExpireAt() != null) {
            return jwtConfig.getComputedOldKeyExpireAt();
        }
        return Instant.now().plusSeconds(jwtConfig.getOldKeyExpireMinutes() * 60L);
    }

    private Claims parseClaims(String token, Key key) {
        return Jwts.parserBuilder()
                .setSigningKey(key)
                .build()
                .parseClaimsJws(token)
                .getBody();
    }

    private String getStringClaim(Claims claims, String... keys) {
        for (String key : keys) {
            Object value = claims.get(key);
            if (value != null) {
                return value.toString();
            }
        }
        return null;
    }

    public String extractTokenFromHeader(String authorizationHeader) {
        if (authorizationHeader == null || authorizationHeader.trim().isEmpty()) {
            return null;
        }

        String header = authorizationHeader.trim();
        if (header.startsWith("Bearer ")) {
            return header.substring(7).trim();
        }
        if (header.startsWith("JWT ")) {
            return header.substring(4).trim();
        }
        return null;
    }
}
