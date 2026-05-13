package com.edi.converter.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.ArrayList;
import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class EdiDataElement {
    private int position;
    private String value;
    private List<String> components = new ArrayList<>();
    
    public EdiDataElement(int position, String value) {
        this.position = position;
        this.value = value;
    }
}
