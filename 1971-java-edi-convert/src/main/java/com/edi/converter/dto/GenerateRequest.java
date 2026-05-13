package com.edi.converter.dto;

import com.edi.converter.model.EdiDocument;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class GenerateRequest {
    private EdiDocument document;
    private String partnerId;
}
