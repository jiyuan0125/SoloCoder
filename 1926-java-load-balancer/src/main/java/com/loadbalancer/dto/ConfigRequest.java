package com.loadbalancer.dto;

import javax.validation.constraints.Min;

public class ConfigRequest {

    @Min(value = 1, message = "vnodes_per_node must be at least 1")
    private Integer vnodes_per_node;

    @Min(value = 1, message = "session_window_seconds must be at least 1")
    private Long session_window_seconds;

    public Integer getVnodes_per_node() {
        return vnodes_per_node;
    }

    public void setVnodes_per_node(Integer vnodes_per_node) {
        this.vnodes_per_node = vnodes_per_node;
    }

    public Long getSession_window_seconds() {
        return session_window_seconds;
    }

    public void setSession_window_seconds(Long session_window_seconds) {
        this.session_window_seconds = session_window_seconds;
    }
}
