package com.purchase.approval.service;

import com.purchase.approval.dto.BlacklistDTO;
import com.purchase.approval.dto.SupplierDTO;
import com.purchase.approval.entity.Supplier;
import com.purchase.approval.enums.BlacklistReason;
import com.purchase.approval.exception.BusinessException;
import com.purchase.approval.repository.SupplierRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor
public class SupplierService {

    private final SupplierRepository supplierRepository;

    public Supplier createSupplier(SupplierDTO dto) {
        if (supplierRepository.findBySupplierCode(dto.getSupplierCode()).isPresent()) {
            throw new BusinessException("供应商编码已存在");
        }

        Supplier supplier = new Supplier();
        supplier.setSupplierCode(dto.getSupplierCode());
        supplier.setSupplierName(dto.getSupplierName());
        supplier.setContactPerson(dto.getContactPerson());
        supplier.setPhone(dto.getPhone());
        supplier.setEmail(dto.getEmail());
        supplier.setAddress(dto.getAddress());
        supplier.setIsBlacklisted(false);

        return supplierRepository.save(supplier);
    }

    public Supplier getSupplier(Long id) {
        return supplierRepository.findById(id)
                .orElseThrow(() -> new BusinessException("供应商不存在"));
    }

    public List<Supplier> getAllSuppliers() {
        return supplierRepository.findAll();
    }

    public List<Supplier> getBlacklistedSuppliers() {
        return supplierRepository.findByIsBlacklistedTrue();
    }

    public boolean isSupplierBlacklisted(Long supplierId) {
        Supplier supplier = getSupplier(supplierId);
        return supplier.isCurrentlyBlacklisted();
    }

    public void validateSupplierNotBlacklisted(Long supplierId) {
        if (supplierId == null) {
            return;
        }
        if (isSupplierBlacklisted(supplierId)) {
            Supplier supplier = getSupplier(supplierId);
            String reason = supplier.getBlacklistReason() != null 
                ? supplier.getBlacklistReason().getDescription() 
                : "黑名单供应商";
            throw new BusinessException(String.format("供应商[%s]处于黑名单中，原因：%s", 
                supplier.getSupplierName(), reason));
        }
    }

    @Transactional
    public Supplier addToBlacklist(BlacklistDTO dto) {
        Supplier supplier = getSupplier(dto.getSupplierId());

        if (supplier.isCurrentlyBlacklisted()) {
            throw new BusinessException("该供应商已在黑名单中");
        }

        if (dto.getBlacklistEndDate().isBefore(LocalDate.now())) {
            throw new BusinessException("黑名单截止日期不能早于今天");
        }

        BlacklistReason reason;
        try {
            reason = BlacklistReason.valueOf(dto.getBlacklistReason());
        } catch (IllegalArgumentException e) {
            throw new BusinessException("无效的黑名单原因");
        }

        supplier.setIsBlacklisted(true);
        supplier.setBlacklistReason(reason);
        supplier.setBlacklistStartDate(LocalDate.now());
        supplier.setBlacklistEndDate(dto.getBlacklistEndDate());
        supplier.setBlacklistRemarks(dto.getRemarks());

        log.info("供应商[{}]已被加入黑名单，截止日期：{}", 
            supplier.getSupplierName(), dto.getBlacklistEndDate());

        return supplierRepository.save(supplier);
    }

    @Transactional
    public Supplier removeFromBlacklist(Long supplierId) {
        Supplier supplier = getSupplier(supplierId);

        if (!supplier.getIsBlacklisted()) {
            throw new BusinessException("该供应商不在黑名单中");
        }

        supplier.setIsBlacklisted(false);
        supplier.setBlacklistEndDate(LocalDate.now());
        log.info("供应商[{}]已从黑名单中移除", supplier.getSupplierName());

        return supplierRepository.save(supplier);
    }

    @Transactional
    public void expireBlacklistedSuppliers() {
        LocalDate today = LocalDate.now();
        List<Supplier> expiringSuppliers = supplierRepository.findBlacklistedSuppliersExpiringBefore(today);

        for (Supplier supplier : expiringSuppliers) {
            if (supplier.getBlacklistEndDate() != null && 
                !supplier.getBlacklistEndDate().isAfter(today)) {
                supplier.setIsBlacklisted(false);
                log.info("供应商[{}]黑名单已到期自动解除", supplier.getSupplierName());
            }
        }

        if (!expiringSuppliers.isEmpty()) {
            supplierRepository.saveAll(expiringSuppliers);
        }
    }

    @Scheduled(cron = "0 0 0 * * ?")
    @Transactional
    public void scheduleBlacklistExpiration() {
        expireBlacklistedSuppliers();
    }
}
