package com.gateway.filter;

import com.gateway.model.FilterResult;
import com.gateway.model.RequestContext;

public interface GatewayFilter {

    String getName();

    FilterResult filter(RequestContext context);
}
