package com.delivery.service;

import com.delivery.entity.DeliveryOrder;
import com.delivery.entity.DeliveryPerson;
import com.delivery.entity.DeliveryZone;
import com.delivery.enums.OrderStatus;
import com.delivery.enums.TimeSlot;
import com.delivery.repository.DeliveryOrderRepository;
import com.delivery.repository.DeliveryPersonRepository;
import com.delivery.repository.DeliveryZoneRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import java.util.*;

@Service
@RequiredArgsConstructor
@Slf4j
public class DispatchService {

    private final DeliveryPersonRepository deliveryPersonRepository;
    private final DeliveryZoneRepository deliveryZoneRepository;
    private final DeliveryOrderRepository orderRepository;
    private final OrderService orderService;

    public static class AssignmentResult {
        private boolean success;
        private Long deliveryPersonId;
        private String deliveryPersonName;
        private String zoneCode;
        private String message;

        public AssignmentResult(boolean success, String message) {
            this.success = success;
            this.message = message;
        }

        public AssignmentResult(boolean success, Long deliveryPersonId, String deliveryPersonName, 
                                String zoneCode, String message) {
            this.success = success;
            this.deliveryPersonId = deliveryPersonId;
            this.deliveryPersonName = deliveryPersonName;
            this.zoneCode = zoneCode;
            this.message = message;
        }

        public boolean isSuccess() { return success; }
        public Long getDeliveryPersonId() { return deliveryPersonId; }
        public String getDeliveryPersonName() { return deliveryPersonName; }
        public String getZoneCode() { return zoneCode; }
        public String getMessage() { return message; }
    }

    @Transactional
    public AssignmentResult assignDeliveryPerson(Long orderId, TimeSlot timeSlot) {
        DeliveryOrder order = orderRepository.findById(orderId)
            .orElseThrow(() -> new RuntimeException("订单不存在: " + orderId));
        
        if (order.getStatus() != OrderStatus.CONFIRMED) {
            return new AssignmentResult(false, "订单状态不正确，需要已确认状态");
        }

        List<DeliveryZone> activeZones = deliveryZoneRepository.findByActiveTrue();
        
        DeliveryZone targetZone = findNearestZone(order, activeZones);
        if (targetZone == null) {
            return new AssignmentResult(false, "未找到合适的配送区域");
        }

        List<DeliveryPerson> availablePersons = findAvailableDeliveryPersons(targetZone, timeSlot);
        
        if (availablePersons.isEmpty()) {
            adjustZoneAssignment(targetZone);
            availablePersons = findAvailableDeliveryPersons(targetZone, timeSlot);
        }
        
        if (availablePersons.isEmpty()) {
            return new AssignmentResult(false, "当前时段该区域无可用配送员");
        }

        DeliveryPerson selectedPerson = selectBestDeliveryPerson(availablePersons, order);
        
        return assignToPerson(order, selectedPerson, targetZone);
    }

    private AssignmentResult assignToPerson(DeliveryOrder order, DeliveryPerson person, DeliveryZone zone) {
        order.setDeliveryPerson(person);
        orderRepository.save(order);
        
        orderService.dispatchOrder(order.getId(), person.getId(), zone.getZoneCode());
        
        log.info("订单分配成功 - 订单号: {}, 配送员: {}, 区域: {}", 
                order.getOrderNo(), person.getName(), zone.getZoneName());
        
        return new AssignmentResult(true, person.getId(), person.getName(), zone.getZoneCode(),
            String.format("已分配给配送员 %s", person.getName()));
    }

    private DeliveryZone findNearestZone(DeliveryOrder order, List<DeliveryZone> zones) {
        if (order.getReceiverLatitude() == null || order.getReceiverLongitude() == null) {
            return zones.isEmpty() ? null : zones.get(0);
        }

        double minDistance = Double.MAX_VALUE;
        DeliveryZone nearestZone = null;

        for (DeliveryZone zone : zones) {
            double distance = calculateHaversineDistance(
                order.getReceiverLatitude(), order.getReceiverLongitude(),
                zone.getCenterLatitude(), zone.getCenterLongitude()
            );
            
            if (distance < zone.getRadius() && distance < minDistance) {
                minDistance = distance;
                nearestZone = zone;
            }
        }

        if (nearestZone == null && !zones.isEmpty()) {
            nearestZone = zones.get(0);
        }

        return nearestZone;
    }

    private List<DeliveryPerson> findAvailableDeliveryPersons(DeliveryZone zone, TimeSlot timeSlot) {
        List<DeliveryPerson> zonePersons = deliveryPersonRepository
            .findByActiveTrueAndCurrentZoneCode(zone.getZoneCode());
        
        List<DeliveryPerson> available = new ArrayList<>();
        
        for (DeliveryPerson person : zonePersons) {
            if (isPersonAvailable(person, timeSlot)) {
                available.add(person);
            }
        }
        
        return available;
    }

    private boolean isPersonAvailable(DeliveryPerson person, TimeSlot timeSlot) {
        List<DeliveryOrder> activeOrders = orderRepository
            .findActiveOrdersByDeliveryPerson(person.getId());
        
        for (DeliveryOrder order : activeOrders) {
            if (order.getPreferredTimeSlot() == timeSlot) {
                return false;
            }
        }
        
        return true;
    }

    private DeliveryPerson selectBestDeliveryPerson(List<DeliveryPerson> available, DeliveryOrder order) {
        Map<DeliveryPerson, Integer> loadMap = new HashMap<>();
        
        for (DeliveryPerson person : available) {
            int load = orderRepository.findActiveOrdersByDeliveryPerson(person.getId()).size();
            loadMap.put(person, load);
        }
        
        return loadMap.entrySet().stream()
            .min(Map.Entry.comparingByValue())
            .map(Map.Entry::getKey)
            .orElse(available.get(0));
    }

    @Transactional
    public void adjustZoneAssignment(DeliveryZone overloadedZone) {
        log.info("区域 {} 订单量过大，开始动态调整配送员", overloadedZone.getZoneName());
        
        List<DeliveryZone> activeZones = deliveryZoneRepository.findByActiveTrue();
        List<DeliveryPerson> allActivePersons = deliveryPersonRepository.findByActiveTrue();
        
        for (DeliveryPerson person : allActivePersons) {
            if (person.getCurrentZoneCode() == null) {
                person.setCurrentZoneCode(overloadedZone.getZoneCode());
                deliveryPersonRepository.save(person);
                log.info("配送员 {} 临时分配到区域 {}", person.getName(), overloadedZone.getZoneName());
            }
        }
    }

    @Transactional
    public void reassignDeliveryPerson(Long orderId, Long newPersonId) {
        DeliveryOrder order = orderRepository.findById(orderId)
            .orElseThrow(() -> new RuntimeException("订单不存在: " + orderId));
        
        DeliveryPerson newPerson = deliveryPersonRepository.findById(newPersonId)
            .orElseThrow(() -> new RuntimeException("配送员不存在: " + newPersonId));
        
        if (!newPerson.isActive()) {
            throw new RuntimeException("配送员不可用");
        }
        
        order.setDeliveryPerson(newPerson);
        orderRepository.save(order);
        
        log.info("订单重新分配 - 订单号: {}, 新配送员: {}", order.getOrderNo(), newPerson.getName());
    }

    public List<DeliveryPerson> getAvailableDeliveryPersons(String zoneCode, TimeSlot timeSlot) {
        List<DeliveryPerson> zonePersons;
        
        if (zoneCode != null) {
            zonePersons = deliveryPersonRepository.findByActiveTrueAndCurrentZoneCode(zoneCode);
        } else {
            zonePersons = deliveryPersonRepository.findByActiveTrue();
        }
        
        List<DeliveryPerson> available = new ArrayList<>();
        for (DeliveryPerson person : zonePersons) {
            if (isPersonAvailable(person, timeSlot)) {
                available.add(person);
            }
        }
        
        return available;
    }

    public List<DeliveryZone> getAllZones() {
        return deliveryZoneRepository.findByActiveTrue();
    }

    private double calculateHaversineDistance(double lat1, double lon1, double lat2, double lon2) {
        final int R = 6371;
        
        double latDistance = Math.toRadians(lat2 - lat1);
        double lonDistance = Math.toRadians(lon2 - lon1);
        
        double a = Math.sin(latDistance / 2) * Math.sin(latDistance / 2)
                 + Math.cos(Math.toRadians(lat1)) * Math.cos(Math.toRadians(lat2))
                 * Math.sin(lonDistance / 2) * Math.sin(lonDistance / 2);
        
        double c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
        
        return R * c;
    }
}
