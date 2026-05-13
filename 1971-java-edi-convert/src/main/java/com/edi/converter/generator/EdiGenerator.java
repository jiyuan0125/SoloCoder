package com.edi.converter.generator;

import com.edi.converter.model.Delimiters;
import com.edi.converter.model.EdiDataElement;
import com.edi.converter.model.EdiDocument;
import com.edi.converter.model.EdiSegment;
import org.springframework.stereotype.Component;

@Component
public class EdiGenerator {

    public String generate(EdiDocument document, Delimiters delimiters) {
        if (document == null) {
            throw new IllegalArgumentException("EDI 文档不能为空");
        }
        if (delimiters == null) {
            delimiters = Delimiters.defaultDelimiters();
        }

        StringBuilder result = new StringBuilder();

        for (EdiSegment segment : document.getSegments()) {
            String segmentText = generateSegment(segment, delimiters);
            result.append(segmentText).append(delimiters.getSegmentTerminator());
        }

        return result.toString();
    }

    private String generateSegment(EdiSegment segment, Delimiters delimiters) {
        StringBuilder segmentBuilder = new StringBuilder();
        segmentBuilder.append(segment.getTag());

        if (!segment.getDataElements().isEmpty()) {
            segmentBuilder.append(delimiters.getDataElementSeparator());

            boolean firstElement = true;
            for (EdiDataElement element : segment.getDataElements()) {
                if (!firstElement) {
                    segmentBuilder.append(delimiters.getDataElementSeparator());
                }
                segmentBuilder.append(generateDataElement(element, delimiters));
                firstElement = false;
            }
        }

        return segmentBuilder.toString();
    }

    private String generateDataElement(EdiDataElement element, Delimiters delimiters) {
        if (element.getComponents() != null && !element.getComponents().isEmpty()) {
            StringBuilder componentBuilder = new StringBuilder();
            boolean firstComponent = true;

            for (String component : element.getComponents()) {
                if (!firstComponent) {
                    componentBuilder.append(delimiters.getComponentSeparator());
                }
                componentBuilder.append(escapeValue(component, delimiters));
                firstComponent = false;
            }

            return componentBuilder.toString();
        } else {
            return escapeValue(element.getValue(), delimiters);
        }
    }

    private String escapeValue(String value, Delimiters delimiters) {
        if (value == null) {
            return "";
        }

        StringBuilder result = new StringBuilder();
        char releaseChar = delimiters.getReleaseCharacter();

        for (int i = 0; i < value.length(); i++) {
            char c = value.charAt(i);

            if (needsEscaping(c, delimiters)) {
                result.append(releaseChar);
            }

            result.append(c);
        }

        return result.toString();
    }

    private boolean needsEscaping(char c, Delimiters delimiters) {
        return c == delimiters.getSegmentTerminator()
                || c == delimiters.getDataElementSeparator()
                || c == delimiters.getComponentSeparator()
                || c == delimiters.getReleaseCharacter();
    }
}
