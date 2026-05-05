package com.example.common.request;

import com.example.common.dto.TemplateItemDTO;
import com.example.common.enums.PositionType;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import java.util.List;

public class CreateTemplateRequest {
    @NotBlank(message = "模板名称不能为空")
    private String name;
    @NotNull(message = "岗位类型不能为空")
    private PositionType positionType;
    private String description;
    private List<TemplateItemDTO> items;
    private boolean isDefault;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public PositionType getPositionType() {
        return positionType;
    }

    public void setPositionType(PositionType positionType) {
        this.positionType = positionType;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public List<TemplateItemDTO> getItems() {
        return items;
    }

    public void setItems(List<TemplateItemDTO> items) {
        this.items = items;
    }

    public boolean isDefault() {
        return isDefault;
    }

    public void setDefault(boolean isDefault) {
        this.isDefault = isDefault;
    }
}
