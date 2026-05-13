package com.edi.converter.validation;

import com.edi.converter.exception.EdiValidationException;
import com.edi.converter.mapping.MappingRule;
import com.edi.converter.model.EdiDataElement;
import com.edi.converter.model.EdiDocument;
import com.edi.converter.model.EdiSegment;
import org.springframework.stereotype.Service;

import java.util.*;

@Service
public class ValidationService {

    public void validateParseResult(EdiDocument document, MappingRule mappingRule) {
        List<EdiValidationException.ValidationError> errors = new ArrayList<>();

        if (mappingRule == null || mappingRule.getRequiredFields() == null) {
            return;
        }

        Map<String, List<EdiSegment>> segmentsByTag = groupSegmentsByTag(document);

        for (MappingRule.RequiredField requiredField : mappingRule.getRequiredFields()) {
            if (!requiredField.isRequired()) {
                continue;
            }

            List<EdiSegment> segments = segmentsByTag.get(requiredField.getSegmentTag());
            if (segments == null || segments.isEmpty()) {
                errors.add(createError(
                        requiredField,
                        null,
                        "段 " + requiredField.getSegmentTag() + " 不存在"
                ));
                continue;
            }

            EdiSegment segment = segments.get(0);
            EdiDataElement element = getDataElement(segment, requiredField.getElementPosition());

            if (element == null) {
                errors.add(createError(
                        requiredField,
                        segment.getPosition(),
                        "段 " + requiredField.getSegmentTag() + " 的第 " + requiredField.getElementPosition() + " 个数据元不存在"
                ));
                continue;
            }

            String actualValue;
            if (requiredField.getComponentPosition() > 0) {
                actualValue = getComponentValue(element, requiredField.getComponentPosition());
                if (actualValue == null || actualValue.trim().isEmpty()) {
                    errors.add(createError(
                            requiredField,
                            segment.getPosition(),
                            "段 " + requiredField.getSegmentTag() + " 的第 " + requiredField.getElementPosition() +
                                    " 个数据元第 " + requiredField.getComponentPosition() + " 个组件不存在或为空"
                    ));
                    continue;
                }
            } else {
                actualValue = element.getValue();
                if (actualValue == null || actualValue.trim().isEmpty()) {
                    errors.add(createError(
                            requiredField,
                            segment.getPosition(),
                            "段 " + requiredField.getSegmentTag() + " 的第 " + requiredField.getElementPosition() + " 个数据元为空"
                    ));
                    continue;
                }
            }

            if (requiredField.getExpectedValue() != null && !requiredField.getExpectedValue().isEmpty()) {
                if (!requiredField.getExpectedValue().equals(actualValue)) {
                    EdiValidationException.ValidationError error = createError(
                            requiredField,
                            segment.getPosition(),
                            "段 " + requiredField.getSegmentTag() + " 的 " +
                                    (requiredField.getFieldName() != null ? requiredField.getFieldName() : "字段") +
                                    " 值不正确"
                    );
                    error.setExpectedValue(requiredField.getExpectedValue());
                    error.setActualValue(actualValue);
                    errors.add(error);
                }
            }
        }

        if (!errors.isEmpty()) {
            throw new EdiValidationException("EDI 文档校验失败，共发现 " + errors.size() + " 个错误", errors);
        }
    }

    public void validateGenerateInput(EdiDocument document, MappingRule mappingRule) {
        List<EdiValidationException.ValidationError> errors = new ArrayList<>();

        if (document == null) {
            errors.add(new EdiValidationException.ValidationError("EDI 文档不能为空"));
            throw new EdiValidationException("EDI 文档校验失败", errors);
        }

        if (document.getSegments() == null || document.getSegments().isEmpty()) {
            errors.add(new EdiValidationException.ValidationError("EDI 文档不包含任何段"));
            throw new EdiValidationException("EDI 文档校验失败", errors);
        }

        if (mappingRule == null || mappingRule.getRequiredFields() == null) {
            return;
        }

        Map<String, List<EdiSegment>> segmentsByTag = groupSegmentsByTag(document);

        for (MappingRule.RequiredField requiredField : mappingRule.getRequiredFields()) {
            if (!requiredField.isRequired()) {
                continue;
            }

            List<EdiSegment> segments = segmentsByTag.get(requiredField.getSegmentTag());
            
            if (segments == null || segments.isEmpty()) {
                errors.add(createError(
                        requiredField,
                        null,
                        "内部格式缺少段: " + requiredField.getSegmentTag() +
                                (requiredField.getFieldName() != null ? " (" + requiredField.getFieldName() + ")" : "")
                ));
                continue;
            }

            EdiSegment segment = segments.get(0);
            EdiDataElement element = getDataElement(segment, requiredField.getElementPosition());

            if (element == null) {
                errors.add(createError(
                        requiredField,
                        segment.getPosition(),
                        "内部格式缺少数据元: 段 " + requiredField.getSegmentTag() + " 第 " +
                                requiredField.getElementPosition() + " 个"
                ));
                continue;
            }

            String actualValue;
            if (requiredField.getComponentPosition() > 0) {
                actualValue = getComponentValue(element, requiredField.getComponentPosition());
                if (actualValue == null || actualValue.trim().isEmpty()) {
                    errors.add(createError(
                            requiredField,
                            segment.getPosition(),
                            "内部格式缺少组件值: 段 " + requiredField.getSegmentTag() +
                                    " 数据元 " + requiredField.getElementPosition() +
                                    " 组件 " + requiredField.getComponentPosition()
                    ));
                }
            } else {
                actualValue = element.getValue();
                if (actualValue == null || actualValue.trim().isEmpty()) {
                    errors.add(createError(
                            requiredField,
                            segment.getPosition(),
                            "内部格式数据元值为空: 段 " + requiredField.getSegmentTag() +
                                    " 数据元 " + requiredField.getElementPosition()
                    ));
                }
            }
        }

        if (!errors.isEmpty()) {
            throw new EdiValidationException("内部格式校验失败，共发现 " + errors.size() + " 个错误", errors);
        }
    }

    private Map<String, List<EdiSegment>> groupSegmentsByTag(EdiDocument document) {
        Map<String, List<EdiSegment>> grouped = new HashMap<>();
        if (document == null || document.getSegments() == null) {
            return grouped;
        }
        for (EdiSegment segment : document.getSegments()) {
            grouped.computeIfAbsent(segment.getTag(), k -> new ArrayList<>()).add(segment);
        }
        return grouped;
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

    private EdiValidationException.ValidationError createError(
            MappingRule.RequiredField requiredField,
            Integer segmentPosition,
            String message) {
        EdiValidationException.ValidationError error = new EdiValidationException.ValidationError(message);
        error.setSegmentTag(requiredField.getSegmentTag());
        error.setSegmentPosition(segmentPosition);
        error.setElementPosition(requiredField.getElementPosition());
        error.setComponentPosition(requiredField.getComponentPosition() > 0 ? requiredField.getComponentPosition() : null);
        error.setField(requiredField.getFieldName());
        return error;
    }
}
