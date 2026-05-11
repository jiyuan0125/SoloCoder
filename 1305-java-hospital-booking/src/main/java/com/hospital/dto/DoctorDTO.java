package com.hospital.dto;

import com.hospital.enums.DoctorTitle;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class DoctorDTO {
    private Long id;
    private String name;
    private DoctorTitle title;
    private Long departmentId;
    private String departmentName;
}
