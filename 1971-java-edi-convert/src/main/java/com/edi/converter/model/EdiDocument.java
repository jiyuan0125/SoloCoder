package com.edi.converter.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.ArrayList;
import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class EdiDocument {
    private List<EdiSegment> segments = new ArrayList<>();
    
    public void addSegment(EdiSegment segment) {
        this.segments.add(segment);
    }
}
