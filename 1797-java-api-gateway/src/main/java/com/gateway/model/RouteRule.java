package com.gateway.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.Valid;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotEmpty;
import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class RouteRule {

    @NotBlank
    private String id;

    @NotBlank
    private String pathPrefix;

    @NotEmpty
    @Valid
    private List<BackendServer> backends;

    @NotBlank
    private String filterChainId;
}
