package com.purchase.approval.dto;

import lombok.Data;
import javax.validation.constraints.NotBlank;

@Data
public class SupplierDTO {
    @NotBlank(message = "供应商编码不能为空")
    private String supplierCode;
    
    @NotBlank(message = "供应商名称不能为空")
    private String supplierName;
    
    private String contactPerson;
    
    private String phone;
    
    private String email;
    
    private String address;
}
