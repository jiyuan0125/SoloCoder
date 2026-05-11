package com.retail.memberpoints.config;

import com.retail.memberpoints.dto.EarnPointsRequest;
import com.retail.memberpoints.dto.RegisterMemberRequest;
import com.retail.memberpoints.entity.Member;
import com.retail.memberpoints.repository.MemberRepository;
import com.retail.memberpoints.service.MemberService;
import com.retail.memberpoints.service.PointsService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.CommandLineRunner;
import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;

@Component
@Profile("!test")
@Slf4j
@RequiredArgsConstructor
public class DataInitializer implements CommandLineRunner {

    private final MemberRepository memberRepository;
    private final MemberService memberService;
    private final PointsService pointsService;

    @Override
    public void run(String... args) {
        if (memberRepository.count() > 0) {
            log.info("数据库已存在数据，跳过初始化");
            return;
        }

        log.info("开始初始化示例数据...");

        Member member1 = memberService.register(RegisterMemberRequest.builder()
                .phone("13800138001")
                .name("张三")
                .build());
        
        Member member2 = memberService.register(RegisterMemberRequest.builder()
                .phone("13800138002")
                .name("李四")
                .build());
        
        Member member3 = memberService.register(RegisterMemberRequest.builder()
                .phone("13800138003")
                .name("王五")
                .build());

        pointsService.earnPoints(EarnPointsRequest.builder()
                .memberId(member1.getId())
                .orderNo("ORD001")
                .orderAmount(new BigDecimal("500"))
                .build());

        pointsService.earnPoints(EarnPointsRequest.builder()
                .memberId(member1.getId())
                .orderNo("ORD002")
                .orderAmount(new BigDecimal("600"))
                .build());

        pointsService.earnPoints(EarnPointsRequest.builder()
                .memberId(member2.getId())
                .orderNo("ORD003")
                .orderAmount(new BigDecimal("3000"))
                .build());

        pointsService.earnPoints(EarnPointsRequest.builder()
                .memberId(member2.getId())
                .orderNo("ORD004")
                .orderAmount(new BigDecimal("2500"))
                .build());

        pointsService.earnPoints(EarnPointsRequest.builder()
                .memberId(member3.getId())
                .orderNo("ORD005")
                .orderAmount(new BigDecimal("8000"))
                .build());

        pointsService.earnPoints(EarnPointsRequest.builder()
                .memberId(member3.getId())
                .orderNo("ORD006")
                .orderAmount(new BigDecimal("15000"))
                .build());

        log.info("示例数据初始化完成");
        log.info("会员1: 张三, 累计消费 1100元 -> 应升级为银卡");
        log.info("会员2: 李四, 累计消费 5500元 -> 应升级为金卡");
        log.info("会员3: 王五, 累计消费 23000元 -> 应升级为钻石卡");
    }
}
