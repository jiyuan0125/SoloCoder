package com.inventory.repository;

import com.inventory.entity.OutboundRecord;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface OutboundRecordRepository extends JpaRepository<OutboundRecord, Long> {
}
