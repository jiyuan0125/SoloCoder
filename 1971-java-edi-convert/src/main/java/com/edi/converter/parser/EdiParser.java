package com.edi.converter.parser;

import com.edi.converter.exception.EdiParseException;
import com.edi.converter.model.Delimiters;
import com.edi.converter.model.EdiDataElement;
import com.edi.converter.model.EdiDocument;
import com.edi.converter.model.EdiSegment;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Component
public class EdiParser {

    public EdiDocument parse(String ediText) {
        if (ediText == null || ediText.trim().isEmpty()) {
            throw new EdiParseException("EDI 文本为空");
        }

        String normalizedText = normalizeEdiText(ediText);
        Delimiters delimiters = extractDelimiters(normalizedText);

        EdiDocument document = new EdiDocument();
        List<String> segments = splitSegments(normalizedText, delimiters);

        int segmentPosition = 0;
        for (String segmentText : segments) {
            if (segmentText.trim().isEmpty()) {
                continue;
            }
            
            EdiSegment segment = parseSegment(segmentText, delimiters, segmentPosition);
            if (segment != null) {
                document.addSegment(segment);
                segmentPosition++;
            }
        }

        return document;
    }

    private String normalizeEdiText(String ediText) {
        return ediText
                .replace("\r\n", "")
                .replace("\r", "")
                .replace("\n", "")
                .trim();
    }

    private Delimiters extractDelimiters(String ediText) {
        if (ediText.startsWith("UNA")) {
            return Delimiters.fromUna(ediText.substring(0, 9));
        }
        return Delimiters.defaultDelimiters();
    }

    private List<String> splitSegments(String ediText, Delimiters delimiters) {
        List<String> segments = new ArrayList<>();
        StringBuilder currentSegment = new StringBuilder();
        boolean inRelease = false;

        for (int i = 0; i < ediText.length(); i++) {
            char c = ediText.charAt(i);

            if (inRelease) {
                currentSegment.append(c);
                inRelease = false;
                continue;
            }

            if (c == delimiters.getReleaseCharacter()) {
                currentSegment.append(c);
                inRelease = true;
                continue;
            }

            if (c == delimiters.getSegmentTerminator()) {
                if (currentSegment.length() > 0) {
                    segments.add(currentSegment.toString());
                }
                currentSegment = new StringBuilder();
            } else {
                currentSegment.append(c);
            }
        }

        if (currentSegment.length() > 0) {
            segments.add(currentSegment.toString());
        }

        return segments;
    }

    private EdiSegment parseSegment(String segmentText, Delimiters delimiters, int position) {
        String tag = extractTag(segmentText, delimiters);
        if (tag == null || tag.isEmpty()) {
            return null;
        }

        EdiSegment segment = new EdiSegment(tag, position);
        String content = extractSegmentContent(segmentText, tag, delimiters);

        if (content == null || content.isEmpty()) {
            return segment;
        }

        List<String> dataElementTexts = splitDataElements(content, delimiters);
        int elementPosition = 1;

        for (String elementText : dataElementTexts) {
            EdiDataElement element = parseDataElement(elementText, delimiters, elementPosition);
            segment.addDataElement(element);
            elementPosition++;
        }

        return segment;
    }

    private String extractTag(String segmentText, Delimiters delimiters) {
        StringBuilder tag = new StringBuilder();
        
        for (int i = 0; i < segmentText.length() && i < 3; i++) {
            char c = segmentText.charAt(i);
            if (isDelimiter(c, delimiters)) {
                break;
            }
            tag.append(c);
        }
        
        return tag.toString();
    }

    private boolean isDelimiter(char c, Delimiters delimiters) {
        return c == delimiters.getDataElementSeparator()
                || c == delimiters.getComponentSeparator()
                || c == delimiters.getSegmentTerminator();
    }

    private String extractSegmentContent(String segmentText, String tag, Delimiters delimiters) {
        int tagLength = tag.length();
        if (segmentText.length() <= tagLength) {
            return null;
        }

        char nextChar = segmentText.charAt(tagLength);
        if (nextChar != delimiters.getDataElementSeparator()) {
            return null;
        }

        return segmentText.substring(tagLength + 1);
    }

    private List<String> splitDataElements(String content, Delimiters delimiters) {
        List<String> elements = new ArrayList<>();
        StringBuilder currentElement = new StringBuilder();
        boolean inRelease = false;

        for (int i = 0; i < content.length(); i++) {
            char c = content.charAt(i);

            if (inRelease) {
                currentElement.append(c);
                inRelease = false;
                continue;
            }

            if (c == delimiters.getReleaseCharacter()) {
                currentElement.append(c);
                inRelease = true;
                continue;
            }

            if (c == delimiters.getDataElementSeparator()) {
                elements.add(currentElement.toString());
                currentElement = new StringBuilder();
            } else {
                currentElement.append(c);
            }
        }

        elements.add(currentElement.toString());
        return elements;
    }

    private EdiDataElement parseDataElement(String elementText, Delimiters delimiters, int position) {
        EdiDataElement element = new EdiDataElement(position, removeReleaseCharacters(elementText, delimiters));

        List<String> components = splitComponents(elementText, delimiters);
        if (components.size() > 1) {
            for (String component : components) {
                element.getComponents().add(removeReleaseCharacters(component, delimiters));
            }
        }

        return element;
    }

    private List<String> splitComponents(String elementText, Delimiters delimiters) {
        List<String> components = new ArrayList<>();
        StringBuilder currentComponent = new StringBuilder();
        boolean inRelease = false;

        for (int i = 0; i < elementText.length(); i++) {
            char c = elementText.charAt(i);

            if (inRelease) {
                currentComponent.append(c);
                inRelease = false;
                continue;
            }

            if (c == delimiters.getReleaseCharacter()) {
                currentComponent.append(c);
                inRelease = true;
                continue;
            }

            if (c == delimiters.getComponentSeparator()) {
                components.add(currentComponent.toString());
                currentComponent = new StringBuilder();
            } else {
                currentComponent.append(c);
            }
        }

        components.add(currentComponent.toString());
        return components;
    }

    private String removeReleaseCharacters(String text, Delimiters delimiters) {
        if (text == null || text.isEmpty()) {
            return text;
        }

        StringBuilder result = new StringBuilder();
        boolean inRelease = false;
        char releaseChar = delimiters.getReleaseCharacter();

        for (int i = 0; i < text.length(); i++) {
            char c = text.charAt(i);

            if (inRelease) {
                result.append(c);
                inRelease = false;
                continue;
            }

            if (c == releaseChar) {
                inRelease = true;
            } else {
                result.append(c);
            }
        }

        return result.toString();
    }
}
