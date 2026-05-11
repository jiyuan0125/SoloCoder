package com.delivery.controller;

import com.delivery.entity.Customer;
import com.delivery.entity.DeliveryPerson;
import com.delivery.entity.DeliveryZone;
import com.delivery.entity.Holiday;
import com.delivery.repository.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.time.LocalDate;
import java.util.List;

@RestController
@RequestMapping("/api/admin")
@RequiredArgsConstructor
@Slf4j
public class AdminController {

    private final CustomerRepository customerRepository;
    private final DeliveryPersonRepository deliveryPersonRepository;
    private final DeliveryZoneRepository deliveryZoneRepository;
    private final HolidayRepository holidayRepository;

    @PostMapping("/customers")
    public ResponseEntity<OrderController.ApiResponse<Customer>> createCustomer(@RequestBody Customer customer) {
        Customer saved = customerRepository.save(customer);
        return ResponseEntity.ok(OrderController.ApiResponse.success("客户创建成功", saved));
    }

    @GetMapping("/customers")
    public ResponseEntity<OrderController.ApiResponse<List<Customer>>> getAllCustomers() {
        List<Customer> customers = customerRepository.findAll();
        return ResponseEntity.ok(OrderController.ApiResponse.success(customers));
    }

    @PostMapping("/delivery-persons")
    public ResponseEntity<OrderController.ApiResponse<DeliveryPerson>> createDeliveryPerson(
            @RequestBody DeliveryPerson person) {
        DeliveryPerson saved = deliveryPersonRepository.save(person);
        return ResponseEntity.ok(OrderController.ApiResponse.success("配送员创建成功", saved));
    }

    @GetMapping("/delivery-persons")
    public ResponseEntity<OrderController.ApiResponse<List<DeliveryPerson>>> getAllDeliveryPersons() {
        List<DeliveryPerson> persons = deliveryPersonRepository.findAll();
        return ResponseEntity.ok(OrderController.ApiResponse.success(persons));
    }

    @PostMapping("/zones")
    public ResponseEntity<OrderController.ApiResponse<DeliveryZone>> createZone(@RequestBody DeliveryZone zone) {
        DeliveryZone saved = deliveryZoneRepository.save(zone);
        return ResponseEntity.ok(OrderController.ApiResponse.success("区域创建成功", saved));
    }

    @GetMapping("/zones")
    public ResponseEntity<OrderController.ApiResponse<List<DeliveryZone>>> getAllZones() {
        List<DeliveryZone> zones = deliveryZoneRepository.findAll();
        return ResponseEntity.ok(OrderController.ApiResponse.success(zones));
    }

    @PostMapping("/holidays")
    public ResponseEntity<OrderController.ApiResponse<Holiday>> createHoliday(@RequestBody Holiday holiday) {
        Holiday saved = holidayRepository.save(holiday);
        return ResponseEntity.ok(OrderController.ApiResponse.success("节假日创建成功", saved));
    }

    @GetMapping("/holidays")
    public ResponseEntity<OrderController.ApiResponse<List<Holiday>>> getAllHolidays() {
        List<Holiday> holidays = holidayRepository.findAll();
        return ResponseEntity.ok(OrderController.ApiResponse.success(holidays));
    }

    @GetMapping("/holidays/check")
    public ResponseEntity<OrderController.ApiResponse<Boolean>> checkHoliday(@RequestParam LocalDate date) {
        boolean isHoliday = holidayRepository.findByHolidayDate(date).isPresent();
        return ResponseEntity.ok(OrderController.ApiResponse.success(isHoliday));
    }
}
