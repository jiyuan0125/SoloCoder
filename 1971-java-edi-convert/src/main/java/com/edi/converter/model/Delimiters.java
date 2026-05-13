package com.edi.converter.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class Delimiters {
    private char segmentTerminator = '\'';
    private char dataElementSeparator = '+';
    private char componentSeparator = ':';
    private char releaseCharacter = '?';
    
    public static Delimiters defaultDelimiters() {
        return new Delimiters();
    }
    
    public static Delimiters fromUna(String unaSegment) {
        if (unaSegment == null || !unaSegment.startsWith("UNA") || unaSegment.length() < 6) {
            return defaultDelimiters();
        }
        Delimiters delimiters = new Delimiters();
        delimiters.setComponentSeparator(unaSegment.charAt(3));
        delimiters.setDataElementSeparator(unaSegment.charAt(4));
        delimiters.setReleaseCharacter(unaSegment.charAt(6));
        return delimiters;
    }
}
