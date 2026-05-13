package com.example.converter.service;

import com.example.converter.dto.FieldError;
import com.example.converter.exception.ValidationException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.networknt.schema.JsonSchema;
import com.networknt.schema.JsonSchemaFactory;
import com.networknt.schema.ValidationMessage;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.xml.sax.SAXException;

import javax.xml.XMLConstants;
import javax.xml.transform.Source;
import javax.xml.transform.stream.StreamSource;
import javax.xml.validation.Schema;
import javax.xml.validation.SchemaFactory;
import javax.xml.validation.Validator;
import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.StringReader;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;

@Service
@RequiredArgsConstructor
@Slf4j
public class ValidationService {

    private final ObjectMapper objectMapper;

    public void validateXml(String xmlData, String xsdSchema) {
        try {
            SchemaFactory factory = SchemaFactory.newInstance(XMLConstants.W3C_XML_SCHEMA_NS_URI);
            factory.setProperty(XMLConstants.ACCESS_EXTERNAL_DTD, "");
            factory.setProperty(XMLConstants.ACCESS_EXTERNAL_SCHEMA, "");

            Schema schema = factory.newSchema(new StreamSource(new StringReader(xsdSchema)));
            Validator validator = schema.newValidator();

            validator.setErrorHandler(new org.xml.sax.ErrorHandler() {
                private final List<FieldError> errors = new ArrayList<>();

                @Override
                public void warning(org.xml.sax.SAXParseException e) {
                    addError(e, "warning");
                }

                @Override
                public void error(org.xml.sax.SAXParseException e) {
                    addError(e, "error");
                }

                @Override
                public void fatalError(org.xml.sax.SAXParseException e) {
                    addError(e, "fatal");
                }

                private void addError(org.xml.sax.SAXParseException e, String level) {
                    errors.add(new FieldError(
                            "line " + e.getLineNumber() + ", column " + e.getColumnNumber(),
                            "valid XML structure",
                            e.getMessage(),
                            level + ": " + e.getMessage()
                    ));
                }
            });

            Source source = new StreamSource(new ByteArrayInputStream(xmlData.getBytes()));
            validator.validate(source);

        } catch (SAXException | IOException e) {
            log.error("XML validation failed", e);
            throw new ValidationException("XML validation failed: " + e.getMessage());
        }
    }

    public void validateJson(String jsonData, String jsonSchema) {
        try {
            JsonSchemaFactory factory = JsonSchemaFactory.getInstance();
            JsonSchema schema = factory.getSchema(jsonSchema);

            JsonNode jsonNode = objectMapper.readTree(jsonData);
            Set<ValidationMessage> errors = schema.validate(jsonNode);

            if (!errors.isEmpty()) {
                List<FieldError> fieldErrors = new ArrayList<>();
                for (ValidationMessage vm : errors) {
                    String field = vm.getPath().isEmpty() ? "root" : vm.getPath();
                    fieldErrors.add(new FieldError(
                            field,
                            vm.getMessage(),
                            jsonNode.at(vm.getPath()).toString(),
                            vm.getMessage()
                    ));
                }
                throw new ValidationException("JSON validation failed", fieldErrors);
            }

        } catch (IOException e) {
            log.error("JSON validation failed", e);
            throw new ValidationException("JSON validation failed: " + e.getMessage());
        }
    }
}
