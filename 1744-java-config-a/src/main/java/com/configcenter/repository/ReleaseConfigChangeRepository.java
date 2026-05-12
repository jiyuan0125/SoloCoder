package com.configcenter.repository;

import com.configcenter.model.ReleaseConfigChange;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface ReleaseConfigChangeRepository extends JpaRepository<ReleaseConfigChange, Long> {
    List<ReleaseConfigChange> findByReleaseId(Long releaseId);
}
