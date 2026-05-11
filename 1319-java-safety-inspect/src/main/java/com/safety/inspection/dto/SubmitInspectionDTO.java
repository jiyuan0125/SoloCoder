package com.safety.inspection.dto;

import jakarta.validation.Valid;
import jakarta.validation.constraints.NotEmpty;
import lombok.Data;

import java.util.List;

@Data
public class SubmitInspectionDTO {

    @NotEmpty(message = "检查记录不能为空")
    @Valid
    private List<InspectionRecordDTO> records;

    private String remark;
}
