package com.configcenter.repository;

import com.configcenter.model.Release;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface ReleaseRepository extends JpaRepository<Release, Long> {
    List<Release> findByProjectIdAndEnvironmentOrderByCreatedAtDesc(Long projectId, String environment);
    List<Release> findByProjectIdAndEnvironmentAndStatusIn(Long projectId, String environment, List<Release.ReleaseStatus> statuses);
}
