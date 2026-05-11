package com.retail.memberpoints.service;

import com.retail.memberpoints.dto.RegisterMemberRequest;
import com.retail.memberpoints.entity.Member;
import com.retail.memberpoints.enums.MemberLevel;
import com.retail.memberpoints.exception.BusinessException;
import com.retail.memberpoints.repository.MemberRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.util.List;

@Service
@Slf4j
@RequiredArgsConstructor
public class MemberService {

    private final MemberRepository memberRepository;

    @Transactional
    public Member register(RegisterMemberRequest request) {
        if (memberRepository.existsByPhone(request.getPhone())) {
            throw BusinessException.memberAlreadyExists();
        }

        Member member = Member.builder()
                .phone(request.getPhone())
                .name(request.getName())
                .totalSpending(BigDecimal.ZERO)
                .pointsBalance(0)
                .level(MemberLevel.NORMAL)
                .build();

        Member saved = memberRepository.save(member);
        log.info("会员注册成功，ID: {}, 手机号: {}", saved.getId(), saved.getPhone());
        return saved;
    }

    public Member getById(Long memberId) {
        return memberRepository.findById(memberId)
                .orElseThrow(BusinessException::memberNotFound);
    }

    public Member getByPhone(String phone) {
        return memberRepository.findByPhone(phone)
                .orElseThrow(BusinessException::memberNotFound);
    }

    public List<Member> getAllMembers() {
        return memberRepository.findAll();
    }

    @Transactional
    public void addSpending(Long memberId, BigDecimal amount) {
        Member member = getById(memberId);
        member.setTotalSpending(member.getTotalSpending().add(amount));
        
        MemberLevel newLevel = MemberLevel.calculateLevel(member.getTotalSpending());
        if (newLevel != member.getLevel()) {
            log.info("会员 {} 等级变化: {} -> {}", member.getId(), member.getLevel(), newLevel);
            member.setLevel(newLevel);
        }
        
        memberRepository.save(member);
    }

    @Transactional
    public void updatePointsBalance(Long memberId, int pointsChange) {
        Member member = getById(memberId);
        int newBalance = member.getPointsBalance() + pointsChange;
        log.debug("会员 {} 积分余额变化: {} -> {} (变化: {})", 
                member.getId(), member.getPointsBalance(), newBalance, pointsChange);
        member.setPointsBalance(newBalance);
        memberRepository.save(member);
    }

    public MemberLevel checkAndUpdateLevel(Member member) {
        MemberLevel calculatedLevel = MemberLevel.calculateLevel(member.getTotalSpending());
        if (calculatedLevel != member.getLevel()) {
            log.info("会员 {} 等级更新: {} -> {}", member.getId(), member.getLevel(), calculatedLevel);
            member.setLevel(calculatedLevel);
            memberRepository.save(member);
        }
        return member.getLevel();
    }
}
