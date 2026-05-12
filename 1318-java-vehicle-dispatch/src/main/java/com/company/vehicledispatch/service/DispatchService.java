package com.company.vehicledispatch.service;

import com.company.vehicledispatch.config.AppConfig;
import com.company.vehicledispatch.dto.DispatchRequestDTO;
import com.company.vehicledispatch.dto.DispatchResultDTO;
import com.company.vehicledispatch.entity.Department;
import com.company.vehicledispatch.entity.DispatchRecord;
import com.company.vehicledispatch.entity.DispatchRequest;
import com.company.vehicledispatch.entity.Vehicle;
import com.company.vehicledispatch.enums.RequestStatus;
import com.company.vehicledispatch.enums.VehicleStatus;
import com.company.vehicledispatch.exception.BusinessException;
import com.company.vehicledispatch.repository.DepartmentRepository;
import com.company.vehicledispatch.repository.DispatchRecordRepository;
import com.company.vehicledispatch.repository.DispatchRequestRepository;
import com.company.vehicledispatch.repository.VehicleRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

@Service
public class DispatchService {

    @Autowired
    private DispatchRequestRepository dispatchRequestRepository;

    @Autowired
    private DispatchRecordRepository dispatchRecordRepository;

    @Autowired
    private VehicleRepository vehicleRepository;

    @Autowired
    private DepartmentRepository departmentRepository;

    @Autowired
    private VehicleService vehicleService;

    @Autowired
    private AppConfig appConfig;

    public List<DispatchRequest> getAllRequests() {
        return dispatchRequestRepository.findAll();
    }

    public Optional<DispatchRequest> getRequestById(Long id) {
        return dispatchRequestRepository.findById(id);
    }

    public List<DispatchRequest> getRequestsByStatus(RequestStatus status) {
        return dispatchRequestRepository.findByStatus(status);
    }

    public List<DispatchRequest> getRequestsByVehicle(Long vehicleId) {
        return dispatchRequestRepository.findByAssignedVehicleId(vehicleId);
    }

    @Transactional
    public DispatchResultDTO createRequest(DispatchRequestDTO dto) {
        if (dto.getEndDateTime().isBefore(dto.getStartDateTime())) {
            throw new BusinessException("结束时间不能早于开始时间");
        }

        if (dto.getEndDateTime().isEqual(dto.getStartDateTime())) {
            throw new BusinessException("结束时间不能等于开始时间");
        }

        Department department = departmentRepository.findById(dto.getDepartmentId())
                .orElseThrow(() -> new BusinessException("部门不存在: " + dto.getDepartmentId()));

        boolean sameDepartmentContinuous = dto.getSameDepartmentContinuous() != null ? 
                dto.getSameDepartmentContinuous() : false;

        List<Vehicle> availableVehicles = findAvailableVehicles(dto.getRequiredSeats(),
                dto.getStartDateTime(), dto.getEndDateTime(), sameDepartmentContinuous, dto.getDepartmentId());

        if (availableVehicles.isEmpty()) {
            return DispatchResultDTO.builder()
                    .success(false)
                    .message("没有可用车辆，请调整时间或座位数要求")
                    .availableVehicles(new ArrayList<>())
                    .build();
        }

        DispatchRequest request = new DispatchRequest();
        request.setDepartment(department);
        request.setStartDateTime(dto.getStartDateTime());
        request.setEndDateTime(dto.getEndDateTime());
        request.setPurpose(dto.getPurpose());
        request.setDestination(dto.getDestination());
        request.setRequiredSeats(dto.getRequiredSeats());
        request.setEstimatedDistance(dto.getEstimatedDistance());
        request.setStatus(RequestStatus.PENDING);
        request.setRemarks(dto.getRemarks());
        request.setSameDepartmentContinuous(dto.getSameDepartmentContinuous() != null ? dto.getSameDepartmentContinuous() : false);
        request.setDispatchConfirmed(false);
        request.setCreatedAt(LocalDateTime.now());
        request.setUpdatedAt(LocalDateTime.now());

        Vehicle selectedVehicle = availableVehicles.get(0);
        request.setAssignedVehicle(selectedVehicle);

        List<String> warnings = new ArrayList<>();
        if (selectedVehicle.getFuelWarning()) {
            warnings.add("车辆油量/电量低于20%，请安排加油/充电");
        }

        DispatchRequest savedRequest = dispatchRequestRepository.save(request);

        return DispatchResultDTO.builder()
                .success(true)
                .message("用车申请已提交，系统已自动匹配车辆")
                .request(savedRequest)
                .assignedVehicle(selectedVehicle)
                .warnings(warnings)
                .availableVehicles(availableVehicles)
                .build();
    }

    public List<Vehicle> findAvailableVehicles(Integer requiredSeats, LocalDateTime startDateTime, LocalDateTime endDateTime) {
        return findAvailableVehicles(requiredSeats, startDateTime, endDateTime, false, null);
    }

    public List<Vehicle> findAvailableVehicles(Integer requiredSeats, LocalDateTime startDateTime, 
                                                LocalDateTime endDateTime, boolean sameDepartmentContinuous, 
                                                Long departmentId) {
        List<Vehicle> availableVehicles = new ArrayList<>();
        List<Vehicle> allVehicles = vehicleRepository.findAll();

        for (Vehicle vehicle : allVehicles) {
            if (isVehicleAvailable(vehicle, requiredSeats, startDateTime, endDateTime, sameDepartmentContinuous, departmentId)) {
                availableVehicles.add(vehicle);
            }
        }

        return availableVehicles;
    }

    public boolean isVehicleAvailable(Vehicle vehicle, Integer requiredSeats, LocalDateTime startDateTime, LocalDateTime endDateTime) {
        return isVehicleAvailable(vehicle, requiredSeats, startDateTime, endDateTime, false, null);
    }

    public boolean isVehicleAvailable(Vehicle vehicle, Integer requiredSeats, LocalDateTime startDateTime, 
                                      LocalDateTime endDateTime, boolean sameDepartmentContinuous, Long departmentId) {
        if (vehicle.getStatus() != VehicleStatus.AVAILABLE) {
            return false;
        }

        if (vehicle.getSeats() < requiredSeats) {
            return false;
        }

        double fuelPercentage = vehicle.getFuelLevel();
        if (fuelPercentage <= appConfig.getVehicle().getFuelUnavailableThreshold()) {
            return false;
        }

        if (vehicle.getInsuranceExpiryDate().isBefore(LocalDateTime.now().toLocalDate())) {
            return false;
        }

        if (vehicle.getAnnualInspectionExpiryDate().isBefore(LocalDateTime.now().toLocalDate())) {
            return false;
        }

        List<DispatchRequest> conflictingRequests = dispatchRequestRepository.findConflictingRequests(
                vehicle.getId(), startDateTime, endDateTime);

        if (!conflictingRequests.isEmpty()) {
            return false;
        }

        if (!sameDepartmentContinuous) {
            LocalDateTime bufferStart = startDateTime.minusMinutes(appConfig.getDispatch().getBufferMinutes());
            List<DispatchRequest> requestsWithinBuffer = dispatchRequestRepository.findRequestsWithinBuffer(
                    vehicle.getId(), startDateTime, bufferStart);

            if (!requestsWithinBuffer.isEmpty()) {
                return false;
            }
        }

        return true;
    }

    @Transactional
    public DispatchRequest approveRequest(Long id, Long vehicleId) {
        DispatchRequest request = dispatchRequestRepository.findById(id)
                .orElseThrow(() -> new BusinessException("申请不存在: " + id));

        if (request.getStatus() != RequestStatus.PENDING) {
            throw new BusinessException("只有待审批的申请才能批准");
        }

        Vehicle vehicle;
        if (vehicleId != null) {
            vehicle = vehicleRepository.findById(vehicleId)
                    .orElseThrow(() -> new BusinessException("车辆不存在: " + vehicleId));

            if (!isVehicleAvailable(vehicle, request.getRequiredSeats(),
                    request.getStartDateTime(), request.getEndDateTime())) {
                throw new BusinessException("所选车辆不可用");
            }
        } else {
            List<Vehicle> availableVehicles = findAvailableVehicles(request.getRequiredSeats(),
                    request.getStartDateTime(), request.getEndDateTime());

            if (availableVehicles.isEmpty()) {
                throw new BusinessException("没有可用车辆");
            }

            vehicle = availableVehicles.get(0);
        }

        request.setAssignedVehicle(vehicle);
        request.setStatus(RequestStatus.APPROVED);
        request.setUpdatedAt(LocalDateTime.now());

        return dispatchRequestRepository.save(request);
    }

    @Transactional
    public DispatchRequest rejectRequest(Long id, String reason) {
        DispatchRequest request = dispatchRequestRepository.findById(id)
                .orElseThrow(() -> new BusinessException("申请不存在: " + id));

        if (request.getStatus() != RequestStatus.PENDING) {
            throw new BusinessException("只有待审批的申请才能拒绝");
        }

        request.setStatus(RequestStatus.REJECTED);
        request.setRemarks(reason);
        request.setUpdatedAt(LocalDateTime.now());

        return dispatchRequestRepository.save(request);
    }

    @Transactional
    public DispatchRecord startDispatch(Long requestId) {
        DispatchRequest request = dispatchRequestRepository.findById(requestId)
                .orElseThrow(() -> new BusinessException("申请不存在: " + requestId));

        if (request.getStatus() != RequestStatus.APPROVED) {
            throw new BusinessException("只有已批准的申请才能开始调度");
        }

        Vehicle vehicle = request.getAssignedVehicle();
        if (vehicle == null) {
            throw new BusinessException("申请未分配车辆");
        }

        DispatchRecord record = new DispatchRecord();
        record.setRequest(request);
        record.setVehicle(vehicle);
        record.setActualStartDateTime(LocalDateTime.now());
        record.setStartMileage(vehicle.getCurrentMileage());
        record.setCreatedAt(LocalDateTime.now());
        record.setUpdatedAt(LocalDateTime.now());

        vehicle.setStatus(VehicleStatus.IN_USE);
        vehicleRepository.save(vehicle);

        request.setStatus(RequestStatus.IN_USE);
        request.setUpdatedAt(LocalDateTime.now());
        dispatchRequestRepository.save(request);

        return dispatchRecordRepository.save(record);
    }

    @Transactional
    public DispatchRecord completeDispatch(Long recordId, Double endMileage, Double actualFuelConsumption) {
        DispatchRecord record = dispatchRecordRepository.findById(recordId)
                .orElseThrow(() -> new BusinessException("调度记录不存在: " + recordId));

        if (record.getActualEndDateTime() != null) {
            throw new BusinessException("该调度已完成");
        }

        Vehicle vehicle = record.getVehicle();

        if (endMileage < record.getStartMileage()) {
            throw new BusinessException("结束里程不能小于开始里程");
        }

        double actualDistance = endMileage - record.getStartMileage();
        double estimatedFuelConsumption = 0.0;

        if (record.getRequest().getEstimatedDistance() != null) {
            estimatedFuelConsumption = (record.getRequest().getEstimatedDistance() * 2 / 100.0) * vehicle.getAverageFuelConsumption();
        }

        record.setActualEndDateTime(LocalDateTime.now());
        record.setEndMileage(endMileage);
        record.setActualDistance(actualDistance);
        record.setEstimatedFuelConsumption(estimatedFuelConsumption);
        record.setActualFuelConsumption(actualFuelConsumption);
        record.setUpdatedAt(LocalDateTime.now());

        if (estimatedFuelConsumption > 0) {
            double deviation = Math.abs(actualFuelConsumption - estimatedFuelConsumption) / estimatedFuelConsumption;
            record.setFuelAbnormal(deviation > appConfig.getFuelDeviationThreshold());
        } else {
            record.setFuelAbnormal(false);
        }

        vehicle.setCurrentMileage(endMileage);
        vehicle.setStatus(VehicleStatus.AVAILABLE);
        if (actualFuelConsumption != null) {
            double newFuelLevel = vehicle.getFuelLevel() - actualFuelConsumption;
            if (newFuelLevel < 0) newFuelLevel = 0;
            vehicle.setFuelLevel(newFuelLevel);
        }
        vehicleService.updateVehicleStatuses(vehicle);
        vehicleRepository.save(vehicle);

        DispatchRequest request = record.getRequest();
        request.setStatus(RequestStatus.COMPLETED);
        request.setUpdatedAt(LocalDateTime.now());
        dispatchRequestRepository.save(request);

        return dispatchRecordRepository.save(record);
    }

    @Transactional
    public DispatchRequest cancelRequest(Long id) {
        DispatchRequest request = dispatchRequestRepository.findById(id)
                .orElseThrow(() -> new BusinessException("申请不存在: " + id));

        if (request.getStatus() == RequestStatus.COMPLETED || request.getStatus() == RequestStatus.CANCELLED) {
            throw new BusinessException("该申请无法取消");
        }

        request.setStatus(RequestStatus.CANCELLED);
        request.setUpdatedAt(LocalDateTime.now());

        return dispatchRequestRepository.save(request);
    }
}
