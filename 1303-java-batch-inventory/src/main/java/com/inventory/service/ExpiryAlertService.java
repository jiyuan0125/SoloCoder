package com.inventory.service;

import com.inventory.dto.ExpiryAlertResponse;
import com.inventory.entity.Batch;
import com.inventory.entity.Inventory;
import com.inventory.entity.Product;
import com.inventory.repository.InventoryRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import java.time.LocalDate;
import java.time.temporal.ChronoUnit;
import java.util.List;
import java.util.stream.Collectors;

@Service
public class ExpiryAlertService {

    @Autowired
    private InventoryRepository inventoryRepository;

    private static final int DEFAULT_ALERT_DAYS = 30;

    @Transactional(readOnly = true)
    public List<ExpiryAlertResponse> getExpiryAlerts() {
        return getExpiryAlerts(DEFAULT_ALERT_DAYS);
    }

    @Transactional(readOnly = true)
    public List<ExpiryAlertResponse> getExpiryAlerts(int daysThreshold) {
        LocalDate today = LocalDate.now();
        LocalDate alertThresholdDate = today.plusDays(daysThreshold);

        List<Inventory> inventories = inventoryRepository.findAllAvailable();

        return inventories.stream()
                .filter(inv -> {
                    Batch batch = inv.getBatch();
                    LocalDate expiryDate = batch.getProductionDate().plusDays(batch.getShelfLifeDays());
                    return !expiryDate.isAfter(alertThresholdDate);
                })
                .map(inv -> convertToAlertResponse(inv, today))
                .sorted((a, b) -> {
                    if (a.getDaysUntilExpiry() == null && b.getDaysUntilExpiry() == null) return 0;
                    if (a.getDaysUntilExpiry() == null) return -1;
                    if (b.getDaysUntilExpiry() == null) return 1;
                    return a.getDaysUntilExpiry().compareTo(b.getDaysUntilExpiry());
                })
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public List<ExpiryAlertResponse> getExpiredProducts() {
        LocalDate today = LocalDate.now();

        List<Inventory> inventories = inventoryRepository.findAllAvailable();

        return inventories.stream()
                .filter(inv -> {
                    Batch batch = inv.getBatch();
                    LocalDate expiryDate = batch.getProductionDate().plusDays(batch.getShelfLifeDays());
                    return expiryDate.isBefore(today);
                })
                .map(inv -> convertToAlertResponse(inv, today))
                .sorted((a, b) -> {
                    if (a.getDaysUntilExpiry() == null && b.getDaysUntilExpiry() == null) return 0;
                    if (a.getDaysUntilExpiry() == null) return -1;
                    if (b.getDaysUntilExpiry() == null) return 1;
                    return a.getDaysUntilExpiry().compareTo(b.getDaysUntilExpiry());
                })
                .collect(Collectors.toList());
    }

    private ExpiryAlertResponse convertToAlertResponse(Inventory inventory, LocalDate today) {
        Batch batch = inventory.getBatch();
        Product product = inventory.getProduct();
        LocalDate expiryDate = batch.getProductionDate().plusDays(batch.getShelfLifeDays());

        ExpiryAlertResponse response = new ExpiryAlertResponse();
        response.setBatchId(batch.getId());
        response.setBatchNumber(batch.getBatchNumber());
        response.setProductId(product.getId());
        response.setProductName(product.getName());
        response.setProductSku(product.getSku());
        response.setCategory(product.getCategory());
        response.setCurrentQuantity(inventory.getCurrentQuantity());
        response.setProductionDate(batch.getProductionDate());
        response.setShelfLifeDays(batch.getShelfLifeDays());
        response.setExpiryDate(expiryDate);
        
        long daysUntilExpiry = ChronoUnit.DAYS.between(today, expiryDate);
        response.setDaysUntilExpiry(daysUntilExpiry);

        return response;
    }
}
