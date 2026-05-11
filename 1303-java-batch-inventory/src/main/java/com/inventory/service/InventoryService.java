package com.inventory.service;

import com.inventory.dto.*;
import com.inventory.entity.*;
import com.inventory.exception.BusinessException;
import com.inventory.exception.ResourceNotFoundException;
import com.inventory.repository.*;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Service
public class InventoryService {

    @Autowired
    private ProductRepository productRepository;

    @Autowired
    private BatchRepository batchRepository;

    @Autowired
    private InventoryRepository inventoryRepository;

    @Autowired
    private OutboundRecordRepository outboundRecordRepository;

    @Transactional
    public BatchResponse inbound(InboundRequest request) {
        Product product = productRepository.findById(request.getProductId())
                .orElseThrow(() -> new ResourceNotFoundException("商品不存在，ID: " + request.getProductId()));

        if (batchRepository.existsByProductIdAndBatchNumber(request.getProductId(), request.getBatchNumber())) {
            throw new BusinessException("该商品已存在相同批次号: " + request.getBatchNumber());
        }

        Batch batch = new Batch();
        batch.setProduct(product);
        batch.setBatchNumber(request.getBatchNumber());
        batch.setInboundDate(request.getInboundDate());
        batch.setProductionDate(request.getProductionDate());
        batch.setShelfLifeDays(request.getShelfLifeDays());
        batch.setInitialQuantity(request.getQuantity());
        Batch savedBatch = batchRepository.save(batch);

        Inventory inventory = new Inventory();
        inventory.setBatch(savedBatch);
        inventory.setProduct(product);
        inventory.setCurrentQuantity(request.getQuantity());
        inventoryRepository.save(inventory);

        return convertToBatchResponse(savedBatch, request.getQuantity());
    }

    @Transactional
    public OutboundResponse outbound(OutboundRequest request) {
        Product product = productRepository.findById(request.getProductId())
                .orElseThrow(() -> new ResourceNotFoundException("商品不存在，ID: " + request.getProductId()));

        List<Inventory> availableInventories = inventoryRepository.findAvailableByProductIdOrderByInboundDateAsc(request.getProductId());

        int totalAvailable = availableInventories.stream()
                .mapToInt(Inventory::getCurrentQuantity)
                .sum();

        if (totalAvailable < request.getQuantity()) {
            throw new BusinessException("库存不足。当前可用库存: " + totalAvailable + ", 请求出库数量: " + request.getQuantity());
        }

        int remainingQuantity = request.getQuantity();
        List<OutboundRecord> outboundRecords = new ArrayList<>();
        List<OutboundResponse.OutboundDetail> details = new ArrayList<>();

        for (Inventory inventory : availableInventories) {
            if (remainingQuantity <= 0) {
                break;
            }

            int availableQty = inventory.getCurrentQuantity();
            int qtyToTake = Math.min(availableQty, remainingQuantity);

            inventory.setCurrentQuantity(availableQty - qtyToTake);
            inventoryRepository.save(inventory);

            OutboundRecord record = new OutboundRecord();
            record.setBatch(inventory.getBatch());
            record.setProduct(product);
            record.setQuantity(qtyToTake);
            outboundRecords.add(record);

            OutboundResponse.OutboundDetail detail = new OutboundResponse.OutboundDetail();
            detail.setBatchId(inventory.getBatch().getId());
            detail.setBatchNumber(inventory.getBatch().getBatchNumber());
            detail.setQuantity(qtyToTake);
            details.add(detail);

            remainingQuantity -= qtyToTake;
        }

        outboundRecordRepository.saveAll(outboundRecords);

        OutboundResponse response = new OutboundResponse();
        response.setProductId(product.getId());
        response.setProductName(product.getName());
        response.setTotalQuantity(request.getQuantity());
        response.setOutboundTime(LocalDateTime.now());
        response.setDetails(details);

        return response;
    }

    @Transactional(readOnly = true)
    public List<InventoryResponse> getInventory() {
        return inventoryRepository.findAllAvailable().stream()
                .map(this::convertToInventoryResponse)
                .collect(Collectors.toList());
    }

    @Transactional(readOnly = true)
    public List<InventoryResponse> getInventoryByProductId(Long productId) {
        if (!productRepository.existsById(productId)) {
            throw new ResourceNotFoundException("商品不存在，ID: " + productId);
        }
        return inventoryRepository.findAvailableByProductIdOrderByInboundDateAsc(productId).stream()
                .map(this::convertToInventoryResponse)
                .collect(Collectors.toList());
    }

    private BatchResponse convertToBatchResponse(Batch batch, Integer currentQuantity) {
        BatchResponse response = new BatchResponse();
        response.setId(batch.getId());
        response.setBatchNumber(batch.getBatchNumber());
        response.setProductId(batch.getProduct().getId());
        response.setProductName(batch.getProduct().getName());
        response.setInboundDate(batch.getInboundDate());
        response.setProductionDate(batch.getProductionDate());
        response.setShelfLifeDays(batch.getShelfLifeDays());
        response.setExpiryDate(batch.getExpiryDate());
        response.setInitialQuantity(batch.getInitialQuantity());
        response.setCurrentQuantity(currentQuantity);
        response.setCreatedAt(batch.getCreatedAt());
        return response;
    }

    private InventoryResponse convertToInventoryResponse(Inventory inventory) {
        Batch batch = inventory.getBatch();
        Product product = inventory.getProduct();

        InventoryResponse response = new InventoryResponse();
        response.setId(inventory.getId());
        response.setBatchId(batch.getId());
        response.setBatchNumber(batch.getBatchNumber());
        response.setProductId(product.getId());
        response.setProductName(product.getName());
        response.setProductSku(product.getSku());
        response.setCategory(product.getCategory());
        response.setCurrentQuantity(inventory.getCurrentQuantity());
        response.setInboundDate(batch.getInboundDate());
        response.setProductionDate(batch.getProductionDate());
        response.setExpiryDate(batch.getExpiryDate());
        response.setUpdatedAt(inventory.getUpdatedAt());
        return response;
    }
}
