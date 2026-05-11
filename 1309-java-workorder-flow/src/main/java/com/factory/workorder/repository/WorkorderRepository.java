package com.factory.workorder.repository;

import com.factory.workorder.entity.Workorder;
import com.factory.workorder.enums.WorkorderPriority;
import com.factory.workorder.enums.WorkorderStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface WorkorderRepository extends JpaRepository<Workorder, Long> {

    Optional<Workorder> findByOrderNo(String orderNo);

    List<Workorder> findByStatus(WorkorderStatus status);

    List<Workorder> findByPriority(WorkorderPriority priority);

    List<Workorder> findByCurrentHandler(String currentHandler);

    List<Workorder> findByStatusNot(WorkorderStatus status);

    @Query("SELECT w FROM Workorder w WHERE w.status <> 'COMPLETED'")
    List<Workorder> findAllNotCompleted();
}
