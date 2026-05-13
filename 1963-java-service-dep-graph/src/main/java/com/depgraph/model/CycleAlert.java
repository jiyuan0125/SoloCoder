package com.depgraph.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class CycleAlert {
    private boolean hasCycle;
    private String message;
    private List<String> cycle;

    public static CycleAlert none() {
        return new CycleAlert(false, "No cyclic dependencies detected", null);
    }

    public static CycleAlert found(List<String> cycle) {
        return new CycleAlert(true, "Cyclic dependency detected", cycle);
    }
}
