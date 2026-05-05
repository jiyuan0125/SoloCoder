package com.reimbursement.server.repository;

import com.reimbursement.server.entity.CostCenter;
import com.reimbursement.server.entity.Reimbursement;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class InMemoryDataStore {
    private final Map<String, Reimbursement> reimbursementMap = new ConcurrentHashMap<>();
    private final Map<String, CostCenter> costCenterMap = new ConcurrentHashMap<>();

    public InMemoryDataStore() {
        initializeCostCenters();
    }

    private void initializeCostCenters() {
        costCenterMap.put("CC001", new CostCenter("CC001", "研发中心", "M001", "张经理", new BigDecimal("100000.00")));
        costCenterMap.put("CC002", new CostCenter("CC002", "市场中心", "M002", "李经理", new BigDecimal("50000.00")));
        costCenterMap.put("CC003", new CostCenter("CC003", "销售中心", "M003", "王经理", new BigDecimal("80000.00")));
        costCenterMap.put("CC004", new CostCenter("CC004", "行政中心", "M004", "赵经理", new BigDecimal("30000.00")));
        costCenterMap.put("CC005", new CostCenter("CC005", "财务中心", "M005", "陈经理", new BigDecimal("20000.00")));
        costCenterMap.put("CC006", new CostCenter("CC006", "人力中心", "M006", "刘经理", new BigDecimal("25000.00")));
    }

    public String generateId() {
        return UUID.randomUUID().toString().replace("-", "").substring(0, 12).toUpperCase();
    }

    public Map<String, Reimbursement> getReimbursementMap() {
        return reimbursementMap;
    }

    public Map<String, CostCenter> getCostCenterMap() {
        return costCenterMap;
    }

    public CostCenter getCostCenterById(String id) {
        return costCenterMap.get(id);
    }

    public Reimbursement getReimbursementById(String id) {
        return reimbursementMap.get(id);
    }

    public void saveReimbursement(Reimbursement reimbursement) {
        reimbursementMap.put(reimbursement.getId(), reimbursement);
    }
}
