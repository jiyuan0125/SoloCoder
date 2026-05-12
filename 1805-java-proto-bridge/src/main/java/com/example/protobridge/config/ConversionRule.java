package com.example.protobridge.config;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import java.util.List;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ConversionRule {

    private String id;

    @NotBlank(message = "路径不能为空")
    private String path;

    @NotBlank(message = "方法不能为空")
    private String method;

    @NotNull(message = "请求类型不能为空")
    private Byte requestType;

    private Byte responseType;

    private List<FieldMapping> requestMappings;

    private List<FieldMapping> responseMappings;

    private String description;

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class FieldMapping {

        @NotBlank(message = "JSON字段名不能为空")
        private String jsonField;

        @NotNull(message = "偏移量不能为空")
        private Integer offset;

        @NotNull(message = "长度不能为空")
        private Integer length;

        @NotBlank(message = "类型不能为空")
        private String type;

        private String description;
    }
}
