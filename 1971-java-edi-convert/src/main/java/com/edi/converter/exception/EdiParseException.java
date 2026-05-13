package com.edi.converter.exception;

public class EdiParseException extends RuntimeException {
    
    private String segmentTag;
    private Integer segmentPosition;
    private Integer elementPosition;
    private Integer componentPosition;
    
    public EdiParseException(String message) {
        super(message);
    }
    
    public EdiParseException(String message, String segmentTag, Integer segmentPosition) {
        super(message);
        this.segmentTag = segmentTag;
        this.segmentPosition = segmentPosition;
    }
    
    public EdiParseException(String message, String segmentTag, Integer segmentPosition, Integer elementPosition) {
        super(message);
        this.segmentTag = segmentTag;
        this.segmentPosition = segmentPosition;
        this.elementPosition = elementPosition;
    }
    
    public EdiParseException(String message, String segmentTag, Integer segmentPosition, Integer elementPosition, Integer componentPosition) {
        super(message);
        this.segmentTag = segmentTag;
        this.segmentPosition = segmentPosition;
        this.elementPosition = elementPosition;
        this.componentPosition = componentPosition;
    }

    public String getSegmentTag() {
        return segmentTag;
    }

    public Integer getSegmentPosition() {
        return segmentPosition;
    }

    public Integer getElementPosition() {
        return elementPosition;
    }

    public Integer getComponentPosition() {
        return componentPosition;
    }
}
