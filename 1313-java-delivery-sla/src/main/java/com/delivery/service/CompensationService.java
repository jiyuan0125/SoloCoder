package com.delivery.service;

import com.delivery.config.DeliveryProperties;
import com.delivery.entity.Compensation;
import com.delivery.entity.DeliveryOrder;
import com.delivery.entity.ExceptionReport;
import com.delivery.enums.CompensationType;
import com.delivery.repository.CompensationRepository;
import com.delivery.repository.DeliveryOrderRepository;
import com.delivery.repository.ExceptionReportRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.Duration;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

@Service
@RequiredArgsConstructor
@Slf4j
public class CompensationService {

    private final CompensationRepository compensationRepository;
    private final DeliveryOrderRepository orderRepository;
    private final ExceptionReportRepository exceptionReportRepository;
    private final DeliveryProperties deliveryProperties;

    public static class CompensationResult {
        private boolean eligible;
        private CompensationType compensationType;
        private BigDecimal compensationAmount;
        private Long delayMinutes;
        private String message;

        public CompensationResult(boolean eligible, String message) {
            this.eligible = eligible;
            this.message = message;
        }

        public CompensationResult(boolean eligible, CompensationType type, BigDecimal amount, 
                                   long delayMinutes, String message) {
            this.eligible = eligible;
            this.compensationType = type;
            this.compensationAmount = amount;
            this.delayMinutes = delayMinutes;
            this.message = message;
        }

        public boolean isEligible() { return eligible; }
        public CompensationType getCompensationType() { return compensationType; }
        public BigDecimal getCompensationAmount() { return compensationAmount; }
        public Long getDelayMinutes() { return delayMinutes; }
        public String getMessage() { return message; }
    }

    @Transactional
    public Compensation checkAndCreateCompensation(DeliveryOrder order) {
        if (order.getPromisedDeliveryTime() == null || order.getActualDeliveryTime() == null) {
            log.warn("订单 {} 缺少时间信息，无法计算超时", order.getOrderNo());
            return null;
        }

        if (hasApprovedExemption(order)) {
            log.info("订单 {} 有已批准的豁免，不计算超时赔偿", order.getOrderNo());
            return null;
        }

        CompensationResult result = calculateCompensation(order);
        
        if (!result.isEligible()) {
            log.info("订单 {} 未超时或不符合赔偿条件: {}", order.getOrderNo(), result.getMessage());
            return null;
        }

        Optional<Compensation> existing = compensationRepository.findByOrderId(order.getId());
        if (existing.isPresent()) {
            log.warn("订单 {} 已有赔偿记录", order.getOrderNo());
            return existing.get();
        }

        Compensation compensation = Compensation.builder()
            .order(order)
            .compensationType(result.getCompensationType())
            .compensationAmount(result.getCompensationAmount())
            .originalFee(order.getTotalFee())
            .delayMinutes(result.getDelayMinutes())
            .description("自动计算超时赔偿 - " + result.getCompensationType().getDescription())
            .approved(true)
            .approvedBy("SYSTEM")
            .approvedAt(LocalDateTime.now())
            .build();

        compensation = compensationRepository.save(compensation);
        
        log.info("赔偿记录已创建 - 订单号: {}, 赔偿金额: {}, 延迟: {}分钟", 
                order.getOrderNo(), result.getCompensationAmount(), result.getDelayMinutes());
        
        return compensation;
    }

    public CompensationResult calculateCompensation(DeliveryOrder order) {
        if (order.getPromisedDeliveryTime() == null || order.getActualDeliveryTime() == null) {
            return new CompensationResult(false, "缺少时间信息");
        }

        if (hasApprovedExemption(order)) {
            return new CompensationResult(false, "存在已批准的豁免");
        }

        LocalDateTime promised = order.getPromisedDeliveryTime();
        LocalDateTime actual = order.getActualDeliveryTime();

        if (actual.isBefore(promised) || actual.isEqual(promised)) {
            return new CompensationResult(false, "未超时");
        }

        Duration delay = Duration.between(promised, actual);
        long delayMinutes = delay.toMinutes();
        long delayHours = delay.toHours();

        DeliveryProperties.CompensationConfig config = deliveryProperties.getCompensation();
        CompensationType compensationType;
        BigDecimal refundRate;

        if (delayHours < 2) {
            compensationType = CompensationType.WITHIN_2_HOURS;
            refundRate = BigDecimal.valueOf(config.getWithin2Hours());
        } else if (delayHours < 12) {
            compensationType = CompensationType.WITHIN_12_HOURS;
            refundRate = BigDecimal.valueOf(config.getWithin12Hours());
        } else {
            compensationType = CompensationType.OVER_12_HOURS;
            refundRate = BigDecimal.valueOf(config.getOver12Hours());
        }

        BigDecimal compensationAmount = order.getTotalFee().multiply(refundRate)
            .setScale(2, RoundingMode.HALF_UP);

        return new CompensationResult(true, compensationType, compensationAmount, delayMinutes,
            String.format("延迟 %d 小时，可获赔 %s 元运费", delayHours, compensationAmount));
    }

    private boolean hasApprovedExemption(DeliveryOrder order) {
        return exceptionReportRepository.findApprovedExemptionForOrder(order.getId()).isPresent();
    }

    @Transactional
    public ExceptionReport reportException(Long orderId, Long deliveryPersonId,
                                            com.delivery.enums.ExceptionReason reason,
                                            String description,
                                            boolean isTimeoutExempt) {
        DeliveryOrder order = orderRepository.findById(orderId)
            .orElseThrow(() -> new RuntimeException("订单不存在: " + orderId));

        com.delivery.entity.DeliveryPerson reporter = 
            new com.delivery.entity.DeliveryPerson();
        reporter.setId(deliveryPersonId);

        ExceptionReport report = ExceptionReport.builder()
            .order(order)
            .reporter(reporter)
            .reason(reason)
            .description(description)
            .isTimeoutExempt(isTimeoutExempt)
            .approved(false)
            .build();

        return exceptionReportRepository.save(report);
    }

    @Transactional
    public ExceptionReport approveExceptionReport(Long reportId, String approver, String comment) {
        ExceptionReport report = exceptionReportRepository.findById(reportId)
            .orElseThrow(() -> new RuntimeException("异常报告不存在: " + reportId));

        if (report.isApproved()) {
            return report;
        }

        report.setApproved(true);
        report.setApprovedBy(approver);
        report.setApprovedAt(LocalDateTime.now());
        report.setApprovalComment(comment);

        return exceptionReportRepository.save(report);
    }

    @Scheduled(cron = "0 */30 * * * *")
    @Transactional
    public void checkDelayedOrders() {
        log.info("定时任务：检查延迟订单");
        
        List<DeliveryOrder> delayedOrders = orderRepository.findDelayedOrders(LocalDateTime.now());
        
        for (DeliveryOrder order : delayedOrders) {
            if (!hasApprovedExemption(order)) {
                log.warn("发现延迟订单: {}, 承诺送达: {}", order.getOrderNo(), order.getPromisedDeliveryTime());
            }
        }
    }

    public Optional<Compensation> getCompensationByOrderId(Long orderId) {
        return compensationRepository.findByOrderId(orderId);
    }
}
