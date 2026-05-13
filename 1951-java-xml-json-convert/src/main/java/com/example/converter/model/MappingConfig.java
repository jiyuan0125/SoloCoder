package com.example.converter.model;

import lombok.Data;
import java.util.List;

@Data
public class MappingConfig {
    private String name;
    private MappingDirection direction;
    private String attributePrefix = "";
    private String rootElementName = "root";
    private String xsdSchema;
    private String jsonSchema;
    private List<FieldMapping> fieldMappings;
}
