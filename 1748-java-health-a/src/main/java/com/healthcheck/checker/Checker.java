package com.healthcheck.checker;

import com.healthcheck.model.CheckItemConfig;
import com.healthcheck.model.CheckResult;

public interface Checker {
    CheckResult execute(CheckItemConfig config, String serviceId);
    boolean supports(CheckItemConfig config);
}
