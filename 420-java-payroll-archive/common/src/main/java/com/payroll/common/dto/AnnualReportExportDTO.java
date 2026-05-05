package com.payroll.common.dto;

import java.util.List;

public class AnnualReportExportDTO {

    private AnnualReportDTO summary;
    private List<PayrollArchiveDTO> archives;

    public AnnualReportExportDTO() {
    }

    public AnnualReportExportDTO(AnnualReportDTO summary, List<PayrollArchiveDTO> archives) {
        this.summary = summary;
        this.archives = archives;
    }

    public AnnualReportDTO getSummary() {
        return summary;
    }

    public void setSummary(AnnualReportDTO summary) {
        this.summary = summary;
    }

    public List<PayrollArchiveDTO> getArchives() {
        return archives;
    }

    public void setArchives(List<PayrollArchiveDTO> archives) {
        this.archives = archives;
    }
}
