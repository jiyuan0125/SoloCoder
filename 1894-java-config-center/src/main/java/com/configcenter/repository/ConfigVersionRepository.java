package com.configcenter.repository;

import com.configcenter.entity.ConfigVersion;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import java.util.List;
import java.util.Optional;

@Repository
public interface ConfigVersionRepository extends JpaRepository<ConfigVersion, Long> {
    List<ConfigVersion> findByApplicationIdAndEnvironmentIdOrderByConfigKeyAsc(Long applicationId, Long environmentId);
    
    @Query("SELECT MAX(cv.versionNumber) FROM ConfigVersion cv WHERE cv.applicationId = :applicationId AND cv.environmentId = :environmentId")
    Optional<Long> findMaxVersionNumber(@Param("applicationId") Long applicationId, @Param("environmentId") Long environmentId);
    
    List<ConfigVersion> findByApplicationIdAndEnvironmentIdAndVersionNumber(Long applicationId, Long environmentId, Long versionNumber);
    
    List<ConfigVersion> findByApplicationIdAndEnvironmentIdAndVersionNumberIn(Long applicationId, Long environmentId, List<Long> versionNumbers);
}
