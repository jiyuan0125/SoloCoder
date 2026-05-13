package com.edi.converter.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.ArrayList;
import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class EdiSegment {
    private String tag;
    private int position;
    private List<EdiDataElement> dataElements = new ArrayList<>();
    
    public EdiSegment(String tag, int position) {
        this.tag = tag;
        this.position = position;
    }
    
    public void addDataElement(EdiDataElement element) {
        this.dataElements.add(element);
    }
}
