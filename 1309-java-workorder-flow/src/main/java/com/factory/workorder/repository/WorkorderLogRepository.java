package com.factory.workorder.repository;

import com.factory.workorder.entity.WorkorderLog;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface WorkorderLogRepository extends JpaRepository<WorkorderLog, Long> {

    List<WorkorderLog> findByWorkorderIdOrderByOperateTimeAsc(Long workorderId);

    List<WorkorderLog> findByWorkorderIdOrderByOperateTimeDesc(Long workorderId);
}
