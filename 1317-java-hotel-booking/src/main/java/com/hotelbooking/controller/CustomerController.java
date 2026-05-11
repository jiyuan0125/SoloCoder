package com.hotelbooking.controller;

import com.hotelbooking.exception.ResourceNotFoundException;
import com.hotelbooking.model.entity.Customer;
import com.hotelbooking.model.enums.MemberLevel;
import com.hotelbooking.repository.CustomerRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/customers")
@RequiredArgsConstructor
public class CustomerController {

    private final CustomerRepository customerRepository;

    @GetMapping
    public ResponseEntity<List<Customer>> getAllCustomers() {
        return ResponseEntity.ok(customerRepository.findAll());
    }

    @GetMapping("/{id}")
    public ResponseEntity<Customer> getCustomer(@PathVariable Long id) {
        Customer customer = customerRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("客户不存在"));
        return ResponseEntity.ok(customer);
    }

    @GetMapping("/phone/{phone}")
    public ResponseEntity<Customer> getCustomerByPhone(@PathVariable String phone) {
        Customer customer = customerRepository.findByPhone(phone)
                .orElseThrow(() -> new ResourceNotFoundException("客户不存在"));
        return ResponseEntity.ok(customer);
    }

    @PostMapping
    public ResponseEntity<Customer> createCustomer(@RequestBody Customer customer) {
        return ResponseEntity.ok(customerRepository.save(customer));
    }

    @PutMapping("/{id}")
    public ResponseEntity<Customer> updateCustomer(@PathVariable Long id, @RequestBody Customer customerDetails) {
        Customer customer = customerRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("客户不存在"));
        
        if (customerDetails.getName() != null) customer.setName(customerDetails.getName());
        if (customerDetails.getPhone() != null) customer.setPhone(customerDetails.getPhone());
        if (customerDetails.getEmail() != null) customer.setEmail(customerDetails.getEmail());
        if (customerDetails.getIdCard() != null) customer.setIdCard(customerDetails.getIdCard());
        if (customerDetails.getMemberLevel() != null) customer.setMemberLevel(customerDetails.getMemberLevel());
        
        return ResponseEntity.ok(customerRepository.save(customer));
    }

    @PutMapping("/{id}/upgrade-level")
    public ResponseEntity<Customer> upgradeMemberLevel(
            @PathVariable Long id,
            @RequestParam MemberLevel newLevel) {
        Customer customer = customerRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("客户不存在"));
        
        customer.setMemberLevel(newLevel);
        return ResponseEntity.ok(customerRepository.save(customer));
    }

    @GetMapping("/levels")
    public ResponseEntity<MemberLevel[]> getMemberLevels() {
        return ResponseEntity.ok(MemberLevel.values());
    }
}
