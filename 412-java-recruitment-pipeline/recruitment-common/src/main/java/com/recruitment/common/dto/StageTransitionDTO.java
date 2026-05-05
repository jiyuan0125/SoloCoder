package com.recruitment.common.dto;

import com.recruitment.common.enums.Stage;

import java.time.LocalDateTime;

public class StageTransitionDTO {
    private Stage fromStage;
    private Stage toStage;
    private LocalDateTime transitionTime;
    private String operator;
    private String remark;

    public StageTransitionDTO() {
    }

    public StageTransitionDTO(Stage fromStage, Stage toStage, LocalDateTime transitionTime) {
        this.fromStage = fromStage;
        this.toStage = toStage;
        this.transitionTime = transitionTime;
    }

    public Stage getFromStage() {
        return fromStage;
    }

    public void setFromStage(Stage fromStage) {
        this.fromStage = fromStage;
    }

    public Stage getToStage() {
        return toStage;
    }

    public void setToStage(Stage toStage) {
        this.toStage = toStage;
    }

    public LocalDateTime getTransitionTime() {
        return transitionTime;
    }

    public void setTransitionTime(LocalDateTime transitionTime) {
        this.transitionTime = transitionTime;
    }

    public String getOperator() {
        return operator;
    }

    public void setOperator(String operator) {
        this.operator = operator;
    }

    public String getRemark() {
        return remark;
    }

    public void setRemark(String remark) {
        this.remark = remark;
    }
}
