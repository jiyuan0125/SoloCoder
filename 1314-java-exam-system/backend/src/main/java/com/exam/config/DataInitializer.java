package com.exam.config;

import com.exam.entity.*;
import com.exam.enums.DifficultyLevel;
import com.exam.enums.QuestionType;
import com.exam.enums.UserRole;
import com.exam.repository.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

@Slf4j
@Component
@RequiredArgsConstructor
public class DataInitializer implements CommandLineRunner {

    private final UserRepository userRepository;
    private final KnowledgeCategoryRepository categoryRepository;
    private final QuestionRepository questionRepository;

    @Override
    @Transactional
    public void run(String... args) {
        if (userRepository.count() == 0) {
            initializeUsers();
        }
        
        if (categoryRepository.count() == 0) {
            initializeCategoriesAndQuestions();
        }
    }

    private void initializeUsers() {
        log.info("初始化用户数据...");

        User admin = new User();
        admin.setUsername("admin");
        admin.setPassword("admin123");
        admin.setRealName("系统管理员");
        admin.setRole(UserRole.ADMIN);
        admin.setEmail("admin@exam.com");
        userRepository.save(admin);

        User teacher = new User();
        teacher.setUsername("teacher");
        teacher.setPassword("teacher123");
        teacher.setRealName("张老师");
        teacher.setRole(UserRole.TEACHER);
        teacher.setEmail("teacher@exam.com");
        userRepository.save(teacher);

        User student = new User();
        student.setUsername("student");
        student.setPassword("student123");
        student.setRealName("李同学");
        student.setRole(UserRole.STUDENT);
        student.setEmail("student@exam.com");
        userRepository.save(student);

        log.info("用户数据初始化完成");
    }

    private void initializeCategoriesAndQuestions() {
        log.info("初始化分类和题目数据...");

        KnowledgeCategory javaCategory = new KnowledgeCategory();
        javaCategory.setName("Java基础");
        javaCategory.setDescription("Java编程语言基础知识");
        javaCategory.setSortOrder(1);
        categoryRepository.save(javaCategory);

        KnowledgeCategory dbCategory = new KnowledgeCategory();
        dbCategory.setName("数据库");
        dbCategory.setDescription("数据库相关知识");
        dbCategory.setSortOrder(2);
        categoryRepository.save(dbCategory);

        KnowledgeCategory networkCategory = new KnowledgeCategory();
        networkCategory.setName("网络协议");
        networkCategory.setDescription("计算机网络和协议知识");
        networkCategory.setSortOrder(3);
        categoryRepository.save(networkCategory);

        createJavaQuestions(javaCategory);
        createDatabaseQuestions(dbCategory);
        createNetworkQuestions(networkCategory);

        log.info("分类和题目数据初始化完成");
    }

    private void createJavaQuestions(KnowledgeCategory category) {
        Question q1 = createQuestion(
                category,
                QuestionType.SINGLE_CHOICE,
                DifficultyLevel.LEVEL_2,
                "Java中，以下哪个关键字用于定义常量？",
                10,
                null,
                Arrays.asList(
                        createOption("A", "final", true),
                        createOption("B", "static", false),
                        createOption("C", "const", false),
                        createOption("D", "constant", false)
                ),
                null,
                "final关键字可以用来修饰变量、方法和类，表示不可改变"
        );
        questionRepository.save(q1);

        Question q2 = createQuestion(
                category,
                QuestionType.SINGLE_CHOICE,
                DifficultyLevel.LEVEL_3,
                "Java中，String类的哪个方法可以将字符串转换为小写？",
                10,
                null,
                Arrays.asList(
                        createOption("A", "toLowerCase()", true),
                        createOption("B", "toLower()", false),
                        createOption("C", "lower()", false),
                        createOption("D", "changeCase()", false)
                ),
                null,
                "String类的toLowerCase()方法可以将字符串中的所有字符转换为小写"
        );
        questionRepository.save(q2);

        Question q3 = createQuestion(
                category,
                QuestionType.MULTIPLE_CHOICE,
                DifficultyLevel.LEVEL_3,
                "Java中，以下哪些是合法的访问修饰符？（多选）",
                10,
                null,
                Arrays.asList(
                        createOption("A", "public", true),
                        createOption("B", "private", true),
                        createOption("C", "internal", false),
                        createOption("D", "protected", true)
                ),
                null,
                "Java的四种访问修饰符是：public、private、protected和默认（包访问）"
        );
        questionRepository.save(q3);

        Question q4 = createQuestion(
                category,
                QuestionType.TRUE_FALSE,
                DifficultyLevel.LEVEL_1,
                "Java中，int类型的默认值是0。",
                5,
                null,
                Arrays.asList(
                        createOption("A", "正确", true),
                        createOption("B", "错误", false)
                ),
                null,
                "Java中，基本类型int的默认值是0"
        );
        questionRepository.save(q4);

        Question q5 = createQuestion(
                category,
                QuestionType.FILL_BLANK,
                DifficultyLevel.LEVEL_2,
                "Java中，用于创建对象的关键字是______。",
                10,
                2,
                null,
                Arrays.asList("new", "NEW"),
                "new关键字用于实例化对象"
        );
        questionRepository.save(q5);

        Question q6 = createQuestion(
                category,
                QuestionType.SINGLE_CHOICE,
                DifficultyLevel.LEVEL_4,
                "Java中，以下哪个集合类是线程安全的？",
                10,
                null,
                Arrays.asList(
                        createOption("A", "ArrayList", false),
                        createOption("B", "HashMap", false),
                        createOption("C", "Vector", true),
                        createOption("D", "LinkedList", false)
                ),
                null,
                "Vector是Java最早的集合类之一，它的所有方法都是同步的，是线程安全的"
        );
        questionRepository.save(q6);

        Question q7 = createQuestion(
                category,
                QuestionType.MULTIPLE_CHOICE,
                DifficultyLevel.LEVEL_4,
                "以下哪些是Java 8引入的新特性？（多选）",
                15,
                null,
                Arrays.asList(
                        createOption("A", "Lambda表达式", true),
                        createOption("B", "Stream API", true),
                        createOption("C", "泛型", false),
                        createOption("D", "接口默认方法", true)
                ),
                null,
                "泛型是Java 5引入的特性，Lambda、Stream API和接口默认方法是Java 8引入的"
        );
        questionRepository.save(q7);

        Question q8 = createQuestion(
                category,
                QuestionType.FILL_BLANK,
                DifficultyLevel.LEVEL_3,
                "Java中，实现多线程的两种主要方式是继承Thread类和实现______接口。",
                10,
                2,
                null,
                Arrays.asList("Runnable", "runnable", "RUNNABLE"),
                "可以通过继承Thread类或实现Runnable接口来创建线程"
        );
        questionRepository.save(q8);
    }

    private void createDatabaseQuestions(KnowledgeCategory category) {
        Question q1 = createQuestion(
                category,
                QuestionType.SINGLE_CHOICE,
                DifficultyLevel.LEVEL_2,
                "SQL中，用于查询数据的关键字是？",
                10,
                null,
                Arrays.asList(
                        createOption("A", "SELECT", true),
                        createOption("B", "UPDATE", false),
                        createOption("C", "INSERT", false),
                        createOption("D", "DELETE", false)
                ),
                null,
                "SELECT语句用于从数据库中查询数据"
        );
        questionRepository.save(q1);

        Question q2 = createQuestion(
                category,
                QuestionType.MULTIPLE_CHOICE,
                DifficultyLevel.LEVEL_3,
                "以下哪些是常见的关系型数据库？（多选）",
                10,
                null,
                Arrays.asList(
                        createOption("A", "MySQL", true),
                        createOption("B", "MongoDB", false),
                        createOption("C", "PostgreSQL", true),
                        createOption("D", "Oracle", true)
                ),
                null,
                "MongoDB是文档型NoSQL数据库，MySQL、PostgreSQL和Oracle都是关系型数据库"
        );
        questionRepository.save(q2);

        Question q3 = createQuestion(
                category,
                QuestionType.TRUE_FALSE,
                DifficultyLevel.LEVEL_2,
                "在SQL中，INNER JOIN只返回两个表中匹配的记录。",
                5,
                null,
                Arrays.asList(
                        createOption("A", "正确", true),
                        createOption("B", "错误", false)
                ),
                null,
                "INNER JOIN（内连接）只返回两个表中在连接条件上匹配的记录"
        );
        questionRepository.save(q3);

        Question q4 = createQuestion(
                category,
                QuestionType.FILL_BLANK,
                DifficultyLevel.LEVEL_3,
                "SQL中，用于对查询结果进行分组的关键字是______。",
                10,
                2,
                null,
                Arrays.asList("GROUP BY", "group by", "GROUP BY", "group by"),
                "GROUP BY子句用于将结果集按照一列或多列进行分组"
        );
        questionRepository.save(q4);

        Question q5 = createQuestion(
                category,
                QuestionType.SINGLE_CHOICE,
                DifficultyLevel.LEVEL_4,
                "在MySQL中，以下哪个存储引擎支持事务？",
                10,
                null,
                Arrays.asList(
                        createOption("A", "MyISAM", false),
                        createOption("B", "InnoDB", true),
                        createOption("C", "MEMORY", false),
                        createOption("D", "CSV", false)
                ),
                null,
                "InnoDB是MySQL的默认存储引擎，支持ACID事务"
        );
        questionRepository.save(q5);
    }

    private void createNetworkQuestions(KnowledgeCategory category) {
        Question q1 = createQuestion(
                category,
                QuestionType.SINGLE_CHOICE,
                DifficultyLevel.LEVEL_2,
                "HTTP协议默认使用的端口号是？",
                10,
                null,
                Arrays.asList(
                        createOption("A", "21", false),
                        createOption("B", "22", false),
                        createOption("C", "80", true),
                        createOption("D", "443", false)
                ),
                null,
                "HTTP默认端口是80，HTTPS默认端口是443"
        );
        questionRepository.save(q1);

        Question q2 = createQuestion(
                category,
                QuestionType.MULTIPLE_CHOICE,
                DifficultyLevel.LEVEL_3,
                "以下哪些是TCP协议的特点？（多选）",
                10,
                null,
                Arrays.asList(
                        createOption("A", "面向连接", true),
                        createOption("B", "可靠传输", true),
                        createOption("C", "无连接", false),
                        createOption("D", "流量控制", true)
                ),
                null,
                "TCP是面向连接的、可靠的传输协议，支持流量控制和拥塞控制"
        );
        questionRepository.save(q2);

        Question q3 = createQuestion(
                category,
                QuestionType.TRUE_FALSE,
                DifficultyLevel.LEVEL_2,
                "IP地址192.168.1.1属于私网地址。",
                5,
                null,
                Arrays.asList(
                        createOption("A", "正确", true),
                        createOption("B", "错误", false)
                ),
                null,
                "192.168.0.0/16是私网地址段，用于局域网内部"
        );
        questionRepository.save(q3);

        Question q4 = createQuestion(
                category,
                QuestionType.FILL_BLANK,
                DifficultyLevel.LEVEL_2,
                "DNS的主要功能是将域名解析为______地址。",
                10,
                2,
                null,
                Arrays.asList("IP", "ip", "Ip", "互联网协议"),
                "DNS（域名系统）的主要功能是将人类可读的域名解析为机器可读的IP地址"
        );
        questionRepository.save(q4);

        Question q5 = createQuestion(
                category,
                QuestionType.SINGLE_CHOICE,
                DifficultyLevel.LEVEL_4,
                "OSI模型中，路由器工作在哪一层？",
                10,
                null,
                Arrays.asList(
                        createOption("A", "物理层", false),
                        createOption("B", "数据链路层", false),
                        createOption("C", "网络层", true),
                        createOption("D", "传输层", false)
                ),
                null,
                "路由器工作在网络层，负责在不同网络之间转发数据包"
        );
        questionRepository.save(q5);
    }

    private Question createQuestion(
            KnowledgeCategory category,
            QuestionType type,
            DifficultyLevel difficulty,
            String content,
            int defaultScore,
            Integer answerTimeLimit,
            List<QuestionOption> options,
            List<String> answers,
            String explanation
    ) {
        Question question = new Question();
        question.setType(type);
        question.setCategory(category);
        question.setDifficulty(difficulty);
        question.setContent(content);
        question.setDefaultScore(defaultScore);
        question.setAnswerTimeLimit(answerTimeLimit);
        question.setIgnoreCase(true);
        question.setExplanation(explanation);

        if (options != null) {
            for (int i = 0; i < options.size(); i++) {
                QuestionOption option = options.get(i);
                option.setQuestion(question);
                option.setSortOrder(i);
            }
            question.setOptions(options);
        }

        if (answers != null && !answers.isEmpty()) {
            List<QuestionAnswer> questionAnswers = new ArrayList<>();
            for (String answer : answers) {
                QuestionAnswer qa = new QuestionAnswer();
                qa.setQuestion(question);
                qa.setAnswer(answer);
                questionAnswers.add(qa);
            }
            question.setAnswers(questionAnswers);
        }

        return question;
    }

    private QuestionOption createOption(String key, String content, boolean correct) {
        QuestionOption option = new QuestionOption();
        option.setOptionKey(key);
        option.setContent(content);
        option.setCorrect(correct);
        return option;
    }
}
