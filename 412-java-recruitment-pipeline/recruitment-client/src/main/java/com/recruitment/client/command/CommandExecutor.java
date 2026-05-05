package com.recruitment.client.command;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.recruitment.client.http.HttpClient;
import com.recruitment.common.dto.*;
import com.recruitment.common.enums.SourceChannel;
import com.recruitment.common.enums.Stage;
import com.recruitment.common.request.*;
import com.recruitment.common.response.ApiResponse;
import com.recruitment.common.response.CandidateListResponse;
import com.recruitment.common.response.StatisticsResponse;

import java.io.IOException;
import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;

public class CommandExecutor {
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;

    public CommandExecutor() {
        this.httpClient = new HttpClient();
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
    }

    public void execute(CommandParser parser) throws IOException {
        String command = parser.getCommand();
        
        if (command == null || "help".equals(command)) {
            printHelp();
            return;
        }

        switch (command) {
            case "create-candidate":
                createCandidate(parser);
                break;
            case "get-candidate":
                getCandidate(parser);
                break;
            case "query-candidates":
                queryCandidates(parser);
                break;
            case "add-application":
                addApplication(parser);
                break;
            case "update-stage":
                updateStage(parser);
                break;
            case "add-interview":
                addInterview(parser);
                break;
            case "send-offer":
                sendOffer(parser);
                break;
            case "respond-offer":
                respondOffer(parser);
                break;
            case "confirm-onboard":
                confirmOnboard(parser);
                break;
            case "get-application":
                getApplication(parser);
                break;
            case "add-contact":
                addContact(parser);
                break;
            case "get-contacts":
                getContacts(parser);
                break;
            case "statistics":
                getStatistics();
                break;
            default:
                System.out.println("未知命令: " + command);
                printHelp();
        }
    }

    private void createCandidate(CommandParser parser) throws IOException {
        String name = parser.getOption("name");
        String phone = parser.getOption("phone");
        
        if (name == null || phone == null) {
            System.out.println("用法: create-candidate --name <姓名> --phone <手机号> [--email <邮箱>] [--resume <简历>] [--source <来源>] [--position <岗位>]");
            return;
        }

        CreateCandidateRequest request = new CreateCandidateRequest();
        request.setName(name);
        request.setPhone(phone);
        request.setEmail(parser.getOption("email"));
        request.setResume(parser.getOption("resume"));
        
        String source = parser.getOption("source");
        if (source != null) {
            request.setSourceChannel(SourceChannel.fromName(source));
        }
        request.setPosition(parser.getOption("position"));

        @SuppressWarnings("unchecked")
        ApiResponse<CandidateDTO> response = httpClient.post("/api/candidates", request, ApiResponse.class);
        printResponse(response);
    }

    private void getCandidate(CommandParser parser) throws IOException {
        String id = parser.getOption("id");
        if (id == null) {
            System.out.println("用法: get-candidate --id <候选人ID>");
            return;
        }

        @SuppressWarnings("unchecked")
        ApiResponse<CandidateListResponse> response = httpClient.get("/api/candidates/" + id, ApiResponse.class);
        printResponse(response);
    }

    private void queryCandidates(CommandParser parser) throws IOException {
        QueryCandidatesRequest request = new QueryCandidatesRequest();
        request.setPosition(parser.getOption("position"));
        
        String stageStr = parser.getOption("stage");
        if (stageStr != null) {
            request.setStage(Stage.fromName(stageStr));
        }
        
        String sourceStr = parser.getOption("source");
        if (sourceStr != null) {
            request.setSourceChannel(SourceChannel.fromName(sourceStr));
        }
        
        request.setPage(parser.getIntOption("page", 0));
        request.setSize(parser.getIntOption("size", 20));

        @SuppressWarnings("unchecked")
        ApiResponse<List<CandidateListResponse>> response = httpClient.post("/api/candidates/query", request, ApiResponse.class);
        printResponse(response);
    }

    private void addApplication(CommandParser parser) throws IOException {
        String candidateId = parser.getOption("candidate-id");
        String position = parser.getOption("position");
        
        if (candidateId == null || position == null) {
            System.out.println("用法: add-application --candidate-id <候选人ID> --position <岗位>");
            return;
        }

        @SuppressWarnings("unchecked")
        ApiResponse<ApplicationDTO> response = httpClient.post(
                "/api/applications/add?candidateId=" + candidateId + "&position=" + position, null, ApiResponse.class);
        printResponse(response);
    }

    private void updateStage(CommandParser parser) throws IOException {
        String applicationId = parser.getOption("application-id");
        String direction = parser.getOption("direction");
        
        if (applicationId == null || direction == null) {
            System.out.println("用法: update-stage --application-id <申请ID> --direction <next|previous> [--remark <备注>] [--operator <操作人>]");
            return;
        }

        UpdateStageRequest request = new UpdateStageRequest();
        request.setApplicationId(applicationId);
        request.setDirection(direction);
        request.setRemark(parser.getOption("remark"));
        request.setOperator(parser.getOption("operator"));

        @SuppressWarnings("unchecked")
        ApiResponse<ApplicationDTO> response = httpClient.post("/api/applications/stage", request, ApiResponse.class);
        printResponse(response);
    }

    private void addInterview(CommandParser parser) throws IOException {
        String applicationId = parser.getOption("application-id");
        if (applicationId == null) {
            System.out.println("用法: add-interview --application-id <申请ID> [--round <轮次>] [--interviewer <面试官>] [--score <评分>] [--comment <评语>]");
            return;
        }

        AddInterviewRequest request = new AddInterviewRequest();
        request.setApplicationId(applicationId);
        request.setRoundNumber(parser.getIntOption("round", null));
        request.setInterviewer(parser.getOption("interviewer"));
        request.setScore(parser.getIntOption("score", null));
        request.setComment(parser.getOption("comment"));
        request.setInterviewTime(LocalDateTime.now());

        @SuppressWarnings("unchecked")
        ApiResponse<ApplicationDTO> response = httpClient.post("/api/applications/interview", request, ApiResponse.class);
        printResponse(response);
    }

    private void sendOffer(CommandParser parser) throws IOException {
        String applicationId = parser.getOption("application-id");
        String salaryStr = parser.getOption("salary");
        
        if (applicationId == null || salaryStr == null) {
            System.out.println("用法: send-offer --application-id <申请ID> --salary <薪资>");
            return;
        }

        SendOfferRequest request = new SendOfferRequest();
        request.setApplicationId(applicationId);
        request.setSalary(new BigDecimal(salaryStr));

        @SuppressWarnings("unchecked")
        ApiResponse<ApplicationDTO> response = httpClient.post("/api/applications/offer/send", request, ApiResponse.class);
        printResponse(response);
    }

    private void respondOffer(CommandParser parser) throws IOException {
        String applicationId = parser.getOption("application-id");
        String acceptedStr = parser.getOption("accepted");
        
        if (applicationId == null || acceptedStr == null) {
            System.out.println("用法: respond-offer --application-id <申请ID> --accepted <true|false>");
            return;
        }

        RespondOfferRequest request = new RespondOfferRequest();
        request.setApplicationId(applicationId);
        request.setAccepted(Boolean.parseBoolean(acceptedStr));

        @SuppressWarnings("unchecked")
        ApiResponse<ApplicationDTO> response = httpClient.post("/api/applications/offer/respond", request, ApiResponse.class);
        printResponse(response);
    }

    private void confirmOnboard(CommandParser parser) throws IOException {
        String applicationId = parser.getOption("application-id");
        if (applicationId == null) {
            System.out.println("用法: confirm-onboard --application-id <申请ID>");
            return;
        }

        @SuppressWarnings("unchecked")
        ApiResponse<ApplicationDTO> response = httpClient.post(
                "/api/applications/onboard?applicationId=" + applicationId, null, ApiResponse.class);
        printResponse(response);
    }

    private void getApplication(CommandParser parser) throws IOException {
        String id = parser.getOption("id");
        if (id == null) {
            System.out.println("用法: get-application --id <申请ID>");
            return;
        }

        @SuppressWarnings("unchecked")
        ApiResponse<CandidateListResponse> response = httpClient.get("/api/applications/" + id, ApiResponse.class);
        printResponse(response);
    }

    private void addContact(CommandParser parser) throws IOException {
        String candidateId = parser.getOption("candidate-id");
        String content = parser.getOption("content");
        
        if (candidateId == null || content == null) {
            System.out.println("用法: add-contact --candidate-id <候选人ID> --content <联系内容> [--type <类型>] [--operator <操作人>]");
            return;
        }

        AddContactRecordRequest request = new AddContactRecordRequest();
        request.setCandidateId(candidateId);
        request.setContent(content);
        request.setContactType(parser.getOption("type"));
        request.setOperator(parser.getOption("operator"));

        @SuppressWarnings("unchecked")
        ApiResponse<ContactRecordDTO> response = httpClient.post("/api/contacts", request, ApiResponse.class);
        printResponse(response);
    }

    private void getContacts(CommandParser parser) throws IOException {
        String candidateId = parser.getOption("candidate-id");
        if (candidateId == null) {
            System.out.println("用法: get-contacts --candidate-id <候选人ID>");
            return;
        }

        @SuppressWarnings("unchecked")
        ApiResponse<List<ContactRecordDTO>> response = httpClient.get("/api/contacts/candidate/" + candidateId, ApiResponse.class);
        printResponse(response);
    }

    private void getStatistics() throws IOException {
        @SuppressWarnings("unchecked")
        ApiResponse<StatisticsResponse> response = httpClient.get("/api/statistics", ApiResponse.class);
        printResponse(response);
    }

    private void printResponse(ApiResponse<?> response) {
        if (response.isSuccess()) {
            System.out.println("成功!");
            if (response.getData() != null) {
                try {
                    String json = objectMapper.writerWithDefaultPrettyPrinter().writeValueAsString(response.getData());
                    System.out.println(json);
                } catch (Exception e) {
                    System.out.println(response.getData());
                }
            }
        } else {
            System.out.println("失败，错误码: " + response.getCode());
            System.out.println("错误信息: " + response.getMessage());
        }
    }

    private void printHelp() {
        System.out.println("=== 招聘漏斗系统 CLI 帮助 ===");
        System.out.println();
        System.out.println("候选人管理:");
        System.out.println("  create-candidate --name <姓名> --phone <手机号> [--email <邮箱>] [--resume <简历>] [--source <来源>] [--position <岗位>]");
        System.out.println("  get-candidate --id <候选人ID>");
        System.out.println("  query-candidates [--position <岗位>] [--stage <阶段>] [--source <来源>]");
        System.out.println();
        System.out.println("申请管理:");
        System.out.println("  add-application --candidate-id <候选人ID> --position <岗位>");
        System.out.println("  update-stage --application-id <申请ID> --direction <next|previous> [--remark <备注>]");
        System.out.println("  add-interview --application-id <申请ID> [--round <轮次>] [--interviewer <面试官>] [--score <评分>]");
        System.out.println("  send-offer --application-id <申请ID> --salary <薪资>");
        System.out.println("  respond-offer --application-id <申请ID> --accepted <true|false>");
        System.out.println("  confirm-onboard --application-id <申请ID>");
        System.out.println("  get-application --id <申请ID>");
        System.out.println();
        System.out.println("联系记录:");
        System.out.println("  add-contact --candidate-id <候选人ID> --content <内容> [--type <类型>]");
        System.out.println("  get-contacts --candidate-id <候选人ID>");
        System.out.println();
        System.out.println("统计:");
        System.out.println("  statistics");
        System.out.println();
        System.out.println("阶段值: SCREENING(简历筛选), PHONE_INTERVIEW(电话面试), ONSITE_INTERVIEW(现场面试), OFFER(发offer), ONBOARD(入职确认)");
        System.out.println("来源值: BOSS, LAGOU, ZHILIAN, INTERNAL(内部推荐), HEADHUNTER(猎头), UNIVERSITY(校园招聘), OTHER");
    }

    public void close() throws IOException {
        httpClient.close();
    }
}
