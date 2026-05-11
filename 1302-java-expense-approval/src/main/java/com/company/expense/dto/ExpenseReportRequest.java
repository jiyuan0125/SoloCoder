package com.company.expense.dto;

import com.company.expense.enums.ExpenseType;
import lombok.Data;

import javax.validation.constraints.*;
import java.math.BigDecimal;

@Data
public class ExpenseReportRequest {

    private Long id;

    @NotBlank(message = "申请人员工编号不能为空")
    private String applicantEmployeeId;

    @NotNull(message = "报销类型不能为空")
    private ExpenseType expenseType;

    @NotNull(message = "报销金额不能为空")
    @DecimalMin(value = "0.01", message = "报销金额必须大于0")
    @DecimalMax(value = "99999.99", message = "报销金额不能超过99999.99元")
    private BigDecimal amount;

    @NotBlank(message = "报销说明不能为空")
    @Size(min = 10, message = "报销说明至少需要10个字")
    @Size(max = 1000, message = "报销说明长度不能超过1000个字符")
    private String description;
}
