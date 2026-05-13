package com.edi.converter.service;

import com.edi.converter.exception.EdiParseException;
import com.edi.converter.exception.EdiValidationException;
import com.edi.converter.generator.EdiGenerator;
import com.edi.converter.mapping.MappingRule;
import com.edi.converter.mapping.MappingService;
import com.edi.converter.model.Delimiters;
import com.edi.converter.model.EdiDocument;
import com.edi.converter.parser.EdiParser;
import com.edi.converter.validation.ValidationService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.Optional;

@Service
public class EdiConverterService {

    private final EdiParser parser;
    private final EdiGenerator generator;
    private final MappingService mappingService;
    private final ValidationService validationService;

    @Autowired
    public EdiConverterService(EdiParser parser, EdiGenerator generator,
                                MappingService mappingService, ValidationService validationService) {
        this.parser = parser;
        this.generator = generator;
        this.mappingService = mappingService;
        this.validationService = validationService;
    }

    public EdiDocument parse(String ediText, String partnerId) {
        if (ediText == null || ediText.trim().isEmpty()) {
            throw new EdiParseException("EDI 文本不能为空");
        }

        EdiDocument parsed = parser.parse(ediText);

        if (partnerId != null && !partnerId.isEmpty()) {
            Optional<MappingRule> mappingRuleOpt = mappingService.getMapping(partnerId);
            if (mappingRuleOpt.isPresent()) {
                MappingRule rule = mappingRuleOpt.get();
                validationService.validateParseResult(parsed, rule);
                return mappingService.applyMapping(parsed, rule);
            }
        }

        return parsed;
    }

    public String generate(EdiDocument document, String partnerId) {
        if (document == null) {
            throw new IllegalArgumentException("EDI 文档不能为空");
        }

        Delimiters delimiters = Delimiters.defaultDelimiters();

        if (partnerId != null && !partnerId.isEmpty()) {
            Optional<MappingRule> mappingRuleOpt = mappingService.getMapping(partnerId);
            if (mappingRuleOpt.isPresent()) {
                MappingRule rule = mappingRuleOpt.get();
                validationService.validateGenerateInput(document, rule);
                delimiters = mappingService.toDelimiters(rule.getDelimiters());
                EdiDocument sourceDocument = mappingService.reverseMapping(document, rule);
                return generator.generate(sourceDocument, delimiters);
            }
        }

        return generator.generate(document, delimiters);
    }
}
