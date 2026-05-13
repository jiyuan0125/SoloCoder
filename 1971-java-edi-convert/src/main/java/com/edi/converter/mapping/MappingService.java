package com.edi.converter.mapping;

import com.edi.converter.model.Delimiters;
import com.edi.converter.model.EdiDataElement;
import com.edi.converter.model.EdiDocument;
import com.edi.converter.model.EdiSegment;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class MappingService {

    private final Map<String, MappingRule> mappings = new ConcurrentHashMap<>();

    public MappingRule saveMapping(MappingRule mapping) {
        if (mapping.getPartnerId() == null || mapping.getPartnerId().isEmpty()) {
            throw new IllegalArgumentException("交易伙伴ID不能为空");
        }
        mappings.put(mapping.getPartnerId(), mapping);
        return mapping;
    }

    public Optional<MappingRule> getMapping(String partnerId) {
        return Optional.ofNullable(mappings.get(partnerId));
    }

    public List<MappingRule> getAllMappings() {
        return new ArrayList<>(mappings.values());
    }

    public void deleteMapping(String partnerId) {
        mappings.remove(partnerId);
    }

    public EdiDocument applyMapping(EdiDocument source, MappingRule rule) {
        if (rule == null) {
            return source;
        }

        EdiDocument target = new EdiDocument();
        Map<String, List<EdiSegment>> sourceSegmentsByTag = groupSegmentsByTag(source);

        for (MappingRule.SegmentMapping segmentMapping : rule.getSegmentMappings()) {
            List<EdiSegment> sourceSegments = sourceSegmentsByTag.getOrDefault(
                    segmentMapping.getSourceSegmentTag(), new ArrayList<>());

            int occurrence = segmentMapping.getOccurrence();
            for (int i = 0; i < Math.min(sourceSegments.size(), occurrence); i++) {
                EdiSegment sourceSegment = sourceSegments.get(i);
                EdiSegment targetSegment = mapSegment(sourceSegment, segmentMapping);
                if (targetSegment != null) {
                    target.addSegment(targetSegment);
                }
            }
        }

        return target;
    }

    public EdiDocument reverseMapping(EdiDocument target, MappingRule rule) {
        if (rule == null) {
            return target;
        }

        EdiDocument source = new EdiDocument();
        Map<String, List<EdiSegment>> targetSegmentsByTag = groupSegmentsByTag(target);

        for (MappingRule.SegmentMapping segmentMapping : rule.getSegmentMappings()) {
            List<EdiSegment> targetSegments = targetSegmentsByTag.getOrDefault(
                    segmentMapping.getTargetSegmentTag(), new ArrayList<>());

            int occurrence = segmentMapping.getOccurrence();
            for (int i = 0; i < Math.min(targetSegments.size(), occurrence); i++) {
                EdiSegment targetSegment = targetSegments.get(i);
                EdiSegment sourceSegment = reverseMapSegment(targetSegment, segmentMapping);
                if (sourceSegment != null) {
                    source.addSegment(sourceSegment);
                }
            }
        }

        return source;
    }

    public Delimiters toDelimiters(MappingRule.DelimiterConfig config) {
        if (config == null) {
            return Delimiters.defaultDelimiters();
        }
        return new Delimiters(
                config.getSegmentTerminator(),
                config.getDataElementSeparator(),
                config.getComponentSeparator(),
                config.getReleaseCharacter()
        );
    }

    private Map<String, List<EdiSegment>> groupSegmentsByTag(EdiDocument document) {
        Map<String, List<EdiSegment>> grouped = new HashMap<>();
        for (EdiSegment segment : document.getSegments()) {
            grouped.computeIfAbsent(segment.getTag(), k -> new ArrayList<>()).add(segment);
        }
        return grouped;
    }

    private EdiSegment mapSegment(EdiSegment sourceSegment, MappingRule.SegmentMapping segmentMapping) {
        EdiSegment targetSegment = new EdiSegment(
                segmentMapping.getTargetSegmentTag() != null ? segmentMapping.getTargetSegmentTag() : sourceSegment.getTag(),
                sourceSegment.getPosition()
        );

        List<EdiDataElement> targetElements = new ArrayList<>();

        for (MappingRule.DataElementMapping elementMapping : segmentMapping.getDataElementMappings()) {
            EdiDataElement sourceElement = getDataElement(sourceSegment, elementMapping.getSourceElementPosition());
            if (sourceElement == null && elementMapping.getDefaultValue() == null) {
                continue;
            }

            EdiDataElement targetElement = new EdiDataElement();
            targetElement.setPosition(elementMapping.getTargetElementPosition());

            if (sourceElement != null) {
                if (elementMapping.getSourceComponentPosition() > 0) {
                    String componentValue = getComponentValue(sourceElement, elementMapping.getSourceComponentPosition());
                    if (componentValue != null) {
                        targetElement.setValue(componentValue);
                    } else if (elementMapping.getDefaultValue() != null) {
                        targetElement.setValue(elementMapping.getDefaultValue());
                    }
                } else {
                    targetElement.setValue(sourceElement.getValue() != null ? sourceElement.getValue() : elementMapping.getDefaultValue());
                    if (sourceElement.getComponents() != null && !sourceElement.getComponents().isEmpty()) {
                        targetElement.setComponents(new ArrayList<>(sourceElement.getComponents()));
                    }
                }
            } else if (elementMapping.getDefaultValue() != null) {
                targetElement.setValue(elementMapping.getDefaultValue());
            }

            while (targetElements.size() < elementMapping.getTargetElementPosition()) {
                targetElements.add(new EdiDataElement(targetElements.size() + 1, ""));
            }
            targetElements.set(elementMapping.getTargetElementPosition() - 1, targetElement);
        }

        targetSegment.setDataElements(targetElements);
        return targetSegment;
    }

    private EdiSegment reverseMapSegment(EdiSegment targetSegment, MappingRule.SegmentMapping segmentMapping) {
        EdiSegment sourceSegment = new EdiSegment(
                segmentMapping.getSourceSegmentTag() != null ? segmentMapping.getSourceSegmentTag() : targetSegment.getTag(),
                targetSegment.getPosition()
        );

        List<EdiDataElement> sourceElements = new ArrayList<>();

        for (MappingRule.DataElementMapping elementMapping : segmentMapping.getDataElementMappings()) {
            EdiDataElement targetElement = getDataElement(targetSegment, elementMapping.getTargetElementPosition());
            if (targetElement == null && elementMapping.getDefaultValue() == null) {
                continue;
            }

            EdiDataElement sourceElement = new EdiDataElement();
            sourceElement.setPosition(elementMapping.getSourceElementPosition());

            if (targetElement != null) {
                if (elementMapping.getTargetComponentPosition() > 0) {
                    String componentValue = getComponentValue(targetElement, elementMapping.getTargetComponentPosition());
                    if (componentValue != null) {
                        sourceElement.setValue(componentValue);
                    } else if (elementMapping.getDefaultValue() != null) {
                        sourceElement.setValue(elementMapping.getDefaultValue());
                    }
                } else {
                    sourceElement.setValue(targetElement.getValue() != null ? targetElement.getValue() : elementMapping.getDefaultValue());
                    if (targetElement.getComponents() != null && !targetElement.getComponents().isEmpty()) {
                        sourceElement.setComponents(new ArrayList<>(targetElement.getComponents()));
                    }
                }
            } else if (elementMapping.getDefaultValue() != null) {
                sourceElement.setValue(elementMapping.getDefaultValue());
            }

            while (sourceElements.size() < elementMapping.getSourceElementPosition()) {
                sourceElements.add(new EdiDataElement(sourceElements.size() + 1, ""));
            }
            sourceElements.set(elementMapping.getSourceElementPosition() - 1, sourceElement);
        }

        sourceSegment.setDataElements(sourceElements);
        return sourceSegment;
    }

    private EdiDataElement getDataElement(EdiSegment segment, int position) {
        if (segment == null || segment.getDataElements() == null) {
            return null;
        }
        for (EdiDataElement element : segment.getDataElements()) {
            if (element.getPosition() == position) {
                return element;
            }
        }
        return null;
    }

    private String getComponentValue(EdiDataElement element, int position) {
        if (element == null || element.getComponents() == null) {
            return null;
        }
        int index = position - 1;
        if (index >= 0 && index < element.getComponents().size()) {
            return element.getComponents().get(index);
        }
        return null;
    }
}
