package com.exam.dto;

import lombok.Data;

@Data
public class QuestionOptionDTO {
    private String optionKey;
    
    private String content;
    
    private boolean correct;
}
