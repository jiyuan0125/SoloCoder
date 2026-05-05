package com.recruitment.server.service;

import com.recruitment.common.dto.*;
import com.recruitment.common.enums.*;
import com.recruitment.common.request.*;
import com.recruitment.common.response.ApiResponse;
import com.recruitment.common.response.ApplicationDetailResponse;
import com.recruitment.server.repository.*;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
public class ApplicationService {
    private final ApplicationRepository applicationRepository;
    private final CandidateRepository candidateRepository;
    private final ContactRecordRepository contactRecordRepository;
    private final EmployeeArchiveRepository employeeArchiveRepository;

    public ApplicationService(ApplicationRepository applicationRepository,
                              CandidateRepository candidateRepository,
                              ContactRecordRepository contactRecordRepository,
                              EmployeeArchiveRepository employeeArchiveRepository) {
        this.applicationRepository = applicationRepository;
        this.candidateRepository = candidateRepository;
        this.contactRecordRepository = contactRecordRepository;
        this.employeeArchiveRepository = employeeArchiveRepository;
    }

    public ApiResponse<ApplicationDTO> updateStage(UpdateStageRequest request) {
        Optional<ApplicationDTO> appOpt = applicationRepository.findById(request.getApplicationId());
        if (!appOpt.isPresent()) {
            return ApiResponse.error(ErrorCode.APPLICATION_NOT_FOUND);
        }

        ApplicationDTO application = appOpt.get();
        
        if (application.getStatus() != ApplicationStatus.IN_PROGRESS) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "申请已结束，状态为: " + application.getStatus().getDescription());
        }

        Stage currentStage = application.getCurrentStage();
        Stage targetStage;

        if ("next".equalsIgnoreCase(request.getDirection())) {
            targetStage = currentStage.next();
        } else if ("previous".equalsIgnoreCase(request.getDirection())) {
            targetStage = currentStage.previous();
        } else {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "方向参数错误，只能是 next 或 previous");
        }

        if (targetStage == currentStage) {
            return ApiResponse.error(ErrorCode.INVALID_STAGE_TRANSITION);
        }

        if (targetStage == Stage.OFFER && application.isNotRecommended()) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "该候选人被标记为不推荐，无法发放offer");
        }

        StageTransitionDTO transition = new StageTransitionDTO(currentStage, targetStage, LocalDateTime.now());
        transition.setOperator(request.getOperator());
        transition.setRemark(request.getRemark());
        application.getStageTransitions().add(transition);
        application.setCurrentStage(targetStage);
        application.setUpdatedAt(LocalDateTime.now());

        applicationRepository.save(application);
        return ApiResponse.success(application);
    }

    public ApiResponse<ApplicationDTO> addInterviewRound(AddInterviewRequest request) {
        Optional<ApplicationDTO> appOpt = applicationRepository.findById(request.getApplicationId());
        if (!appOpt.isPresent()) {
            return ApiResponse.error(ErrorCode.APPLICATION_NOT_FOUND);
        }

        ApplicationDTO application = appOpt.get();

        if (request.getScore() != null && (request.getScore() < 1 || request.getScore() > 5)) {
            return ApiResponse.error(ErrorCode.INVALID_SCORE);
        }

        InterviewRoundDTO round = new InterviewRoundDTO();
        round.setRoundNumber(request.getRoundNumber() != null ? request.getRoundNumber() : application.getInterviewRounds().size() + 1);
        round.setInterviewer(request.getInterviewer());
        round.setScore(request.getScore());
        round.setComment(request.getComment());
        round.setInterviewTime(request.getInterviewTime() != null ? request.getInterviewTime() : LocalDateTime.now());

        application.getInterviewRounds().add(round);
        application.setUpdatedAt(LocalDateTime.now());

        BigDecimal avgScore = application.getAverageScore();
        if (avgScore.compareTo(BigDecimal.ZERO) > 0 && avgScore.compareTo(BigDecimal.valueOf(3)) < 0) {
            application.setNotRecommended(true);
        }

        applicationRepository.save(application);
        return ApiResponse.success(application);
    }

    public ApiResponse<ApplicationDTO> sendOffer(SendOfferRequest request) {
        Optional<ApplicationDTO> appOpt = applicationRepository.findById(request.getApplicationId());
        if (!appOpt.isPresent()) {
            return ApiResponse.error(ErrorCode.APPLICATION_NOT_FOUND);
        }

        ApplicationDTO application = appOpt.get();

        if (application.getStatus() != ApplicationStatus.IN_PROGRESS) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "申请已结束");
        }

        if (application.getCurrentStage() != Stage.OFFER) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "当前阶段不在发offer阶段");
        }

        if (application.isNotRecommended()) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "该候选人被标记为不推荐，无法发放offer");
        }

        if (application.getOffer() != null && application.getOffer().isNeedReapproval()) {
            return ApiResponse.error(ErrorCode.OFFER_EXPIRED);
        }

        OfferDTO offer = new OfferDTO();
        offer.setId(UUID.randomUUID().toString());
        offer.setApplicationId(application.getId());
        offer.setSalary(request.getSalary());
        offer.setSentTime(LocalDateTime.now());
        offer.setExpireTime(LocalDateTime.now().plusDays(7));
        offer.setStatus(OfferStatus.PENDING);
        offer.setNeedReapproval(false);

        application.setOffer(offer);
        application.setUpdatedAt(LocalDateTime.now());

        applicationRepository.save(application);
        return ApiResponse.success(application);
    }

    public ApiResponse<ApplicationDTO> respondOffer(RespondOfferRequest request) {
        Optional<ApplicationDTO> appOpt = applicationRepository.findById(request.getApplicationId());
        if (!appOpt.isPresent()) {
            return ApiResponse.error(ErrorCode.APPLICATION_NOT_FOUND);
        }

        ApplicationDTO application = appOpt.get();

        if (application.getOffer() == null) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "不存在offer");
        }

        OfferDTO offer = application.getOffer();

        if (offer.getStatus() != OfferStatus.PENDING) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "offer已处理，状态为: " + offer.getStatus().getDescription());
        }

        LocalDateTime now = LocalDateTime.now();
        if (now.isAfter(offer.getExpireTime())) {
            offer.setStatus(OfferStatus.EXPIRED);
            offer.setNeedReapproval(true);
            application.setUpdatedAt(now);
            applicationRepository.save(application);
            return ApiResponse.error(ErrorCode.OFFER_EXPIRED);
        }

        if (Boolean.TRUE.equals(request.getAccepted())) {
            offer.setStatus(OfferStatus.ACCEPTED);
            offer.setResponseTime(now);
            application.setStatus(ApplicationStatus.OFFER_ACCEPTED);
            application.setUpdatedAt(now);

            abandonOtherApplications(application.getCandidateId(), application.getId());
        } else {
            offer.setStatus(OfferStatus.REJECTED);
            offer.setResponseTime(now);
            application.setStatus(ApplicationStatus.REJECTED);
            application.setUpdatedAt(now);
        }

        applicationRepository.save(application);
        return ApiResponse.success(application);
    }

    private void abandonOtherApplications(String candidateId, String acceptedApplicationId) {
        List<ApplicationDTO> otherApps = applicationRepository.findByCandidateIdExcluding(candidateId, acceptedApplicationId);
        for (ApplicationDTO app : otherApps) {
            if (app.getStatus() == ApplicationStatus.IN_PROGRESS) {
                app.setStatus(ApplicationStatus.ABANDONED);
                app.setAbandonReason("已接受其他岗位offer");
                app.setUpdatedAt(LocalDateTime.now());
                applicationRepository.save(app);
            }
        }
    }

    public ApiResponse<ApplicationDTO> confirmOnboard(String applicationId) {
        Optional<ApplicationDTO> appOpt = applicationRepository.findById(applicationId);
        if (!appOpt.isPresent()) {
            return ApiResponse.error(ErrorCode.APPLICATION_NOT_FOUND);
        }

        ApplicationDTO application = appOpt.get();

        if (application.getStatus() != ApplicationStatus.OFFER_ACCEPTED) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "只有接受offer后才能确认入职");
        }

        if (application.getCurrentStage() != Stage.ONBOARD) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "请先推进到入职确认阶段");
        }

        application.setStatus(ApplicationStatus.ONBOARDED);
        application.setUpdatedAt(LocalDateTime.now());

        createEmployeeArchive(application);

        applicationRepository.save(application);
        return ApiResponse.success(application);
    }

    private void createEmployeeArchive(ApplicationDTO application) {
        Optional<CandidateDTO> candidateOpt = candidateRepository.findById(application.getCandidateId());
        if (!candidateOpt.isPresent()) {
            return;
        }

        CandidateDTO candidate = candidateOpt.get();
        EmployeeArchiveDTO archive = new EmployeeArchiveDTO();
        archive.setId(UUID.randomUUID().toString());
        archive.setCandidateId(candidate.getId());
        archive.setName(candidate.getName());
        archive.setPhone(candidate.getPhone());
        archive.setEmail(candidate.getEmail());
        archive.setPosition(application.getPosition());
        if (application.getOffer() != null) {
            archive.setSalary(application.getOffer().getSalary());
        }
        archive.setSourceChannel(candidate.getSourceChannel());
        archive.setOnboardDate(LocalDateTime.now());
        archive.setCreatedAt(LocalDateTime.now());

        employeeArchiveRepository.save(archive);
    }

    public ApiResponse<ApplicationDetailResponse> getApplicationDetail(String applicationId) {
        Optional<ApplicationDTO> appOpt = applicationRepository.findById(applicationId);
        if (!appOpt.isPresent()) {
            return ApiResponse.error(ErrorCode.APPLICATION_NOT_FOUND);
        }

        ApplicationDTO application = appOpt.get();
        Optional<CandidateDTO> candidateOpt = candidateRepository.findById(application.getCandidateId());
        CandidateDTO candidate = candidateOpt.orElse(null);

        List<ContactRecordDTO> contactRecords = contactRecordRepository.findByCandidateId(application.getCandidateId());

        return ApiResponse.success(new ApplicationDetailResponse(candidate, application, contactRecords));
    }

    public ApiResponse<ApplicationDTO> addApplication(String candidateId, String position) {
        Optional<CandidateDTO> candidateOpt = candidateRepository.findById(candidateId);
        if (!candidateOpt.isPresent()) {
            return ApiResponse.error(ErrorCode.CANDIDATE_NOT_FOUND);
        }

        ApplicationDTO application = new ApplicationDTO();
        application.setId(UUID.randomUUID().toString());
        application.setCandidateId(candidateId);
        application.setPosition(position);
        application.setCreatedAt(LocalDateTime.now());
        application.setUpdatedAt(LocalDateTime.now());

        applicationRepository.save(application);
        return ApiResponse.success(application);
    }

    @Scheduled(fixedRate = 60000)
    public void checkExpiredOffers() {
        List<ApplicationDTO> applications = applicationRepository.findByStatus(ApplicationStatus.IN_PROGRESS);
        LocalDateTime now = LocalDateTime.now();
        
        for (ApplicationDTO app : applications) {
            if (app.getOffer() != null && app.getOffer().getStatus() == OfferStatus.PENDING) {
                if (now.isAfter(app.getOffer().getExpireTime())) {
                    app.getOffer().setStatus(OfferStatus.EXPIRED);
                    app.getOffer().setNeedReapproval(true);
                    app.setUpdatedAt(now);
                    applicationRepository.save(app);
                }
            }
        }
    }
}
