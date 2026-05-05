package com.recruitment.common.dto;

import com.recruitment.common.enums.Stage;

import java.math.BigDecimal;

public class StageStatisticsDTO {
    private Stage stage;
    private Integer count;
    private BigDecimal conversionRate;
    private BigDecimal avgDaysInStage;

    public StageStatisticsDTO() {
        this.count = 0;
        this.conversionRate = BigDecimal.ZERO;
        this.avgDaysInStage = BigDecimal.ZERO;
    }

    public Stage getStage() {
        return stage;
    }

    public void setStage(Stage stage) {
        this.stage = stage;
    }

    public Integer getCount() {
        return count;
    }

    public void setCount(Integer count) {
        this.count = count;
    }

    public BigDecimal getConversionRate() {
        return conversionRate;
    }

    public void setConversionRate(BigDecimal conversionRate) {
        this.conversionRate = conversionRate;
    }

    public BigDecimal getAvgDaysInStage() {
        return avgDaysInStage;
    }

    public void setAvgDaysInStage(BigDecimal avgDaysInStage) {
        this.avgDaysInStage = avgDaysInStage;
    }
}
