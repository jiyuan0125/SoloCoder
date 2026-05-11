package com.exam.util;

import com.exam.entity.Question;
import com.exam.entity.QuestionOption;
import com.exam.enums.QuestionType;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.util.*;
import java.util.stream.Collectors;

public class ShuffleUtil {

    private static final ObjectMapper objectMapper = new ObjectMapper();
    private static final Random random = new Random();

    public static <T> List<T> shuffleList(List<T> list) {
        if (list == null || list.isEmpty()) {
            return new ArrayList<>();
        }
        List<T> shuffled = new ArrayList<>(list);
        Collections.shuffle(shuffled, random);
        return shuffled;
    }

    public static ShuffledOptions shuffleOptions(Question question) {
        if (question.getType() == QuestionType.FILL_BLANK) {
            return ShuffledOptions.builder()
                    .originalOptions(new ArrayList<>())
                    .shuffledOptions(new ArrayList<>())
                    .shuffledToOriginalKeyMap(new HashMap<>())
                    .originalToShuffledKeyMap(new HashMap<>())
                    .correctAnswerKeysOriginal(new ArrayList<>())
                    .correctAnswerKeysShuffled(new ArrayList<>())
                    .build();
        }

        List<QuestionOption> options = question.getOptions();
        if (options == null || options.isEmpty()) {
            return ShuffledOptions.builder()
                    .originalOptions(new ArrayList<>())
                    .shuffledOptions(new ArrayList<>())
                    .shuffledToOriginalKeyMap(new HashMap<>())
                    .originalToShuffledKeyMap(new HashMap<>())
                    .correctAnswerKeysOriginal(new ArrayList<>())
                    .correctAnswerKeysShuffled(new ArrayList<>())
                    .build();
        }

        List<QuestionOption> originalOptions = new ArrayList<>(options);
        List<QuestionOption> shuffledOptions = new ArrayList<>(originalOptions);
        Collections.shuffle(shuffledOptions, random);

        Map<String, String> originalToShuffledKeyMap = new LinkedHashMap<>();
        Map<String, String> shuffledToOriginalKeyMap = new LinkedHashMap<>();
        List<String> correctAnswerKeysOriginal = new ArrayList<>();
        List<String> correctAnswerKeysShuffled = new ArrayList<>();

        char newKeyChar = 'A';
        for (int i = 0; i < shuffledOptions.size(); i++) {
            QuestionOption shuffledOption = shuffledOptions.get(i);
            String newKey = String.valueOf(newKeyChar++);
            
            originalToShuffledKeyMap.put(shuffledOption.getOptionKey(), newKey);
            shuffledToOriginalKeyMap.put(newKey, shuffledOption.getOptionKey());

            if (shuffledOption.isCorrect()) {
                correctAnswerKeysOriginal.add(shuffledOption.getOptionKey());
                correctAnswerKeysShuffled.add(newKey);
            }
        }

        Collections.sort(correctAnswerKeysShuffled);

        return ShuffledOptions.builder()
                .originalOptions(originalOptions)
                .shuffledOptions(shuffledOptions)
                .originalToShuffledKeyMap(originalToShuffledKeyMap)
                .shuffledToOriginalKeyMap(shuffledToOriginalKeyMap)
                .correctAnswerKeysOriginal(correctAnswerKeysOriginal)
                .correctAnswerKeysShuffled(correctAnswerKeysShuffled)
                .build();
    }

    public static String serializeShuffledOptions(ShuffledOptions shuffledOptions) {
        try {
            ShuffleStorageData data = ShuffleStorageData.builder()
                    .shuffledToOriginalKeyMap(shuffledOptions.getShuffledToOriginalKeyMap())
                    .originalToShuffledKeyMap(shuffledOptions.getOriginalToShuffledKeyMap())
                    .correctAnswerKeysOriginal(shuffledOptions.getCorrectAnswerKeysOriginal())
                    .correctAnswerKeysShuffled(shuffledOptions.getCorrectAnswerKeysShuffled())
                    .build();
            return objectMapper.writeValueAsString(data);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("序列化打乱选项失败", e);
        }
    }

    public static ShuffleStorageData deserializeShuffledOptions(String shuffledOptionsJson) {
        if (shuffledOptionsJson == null || shuffledOptionsJson.isEmpty()) {
            return ShuffleStorageData.builder()
                    .shuffledToOriginalKeyMap(new HashMap<>())
                    .originalToShuffledKeyMap(new HashMap<>())
                    .correctAnswerKeysOriginal(new ArrayList<>())
                    .correctAnswerKeysShuffled(new ArrayList<>())
                    .build();
        }
        try {
            return objectMapper.readValue(shuffledOptionsJson, ShuffleStorageData.class);
        } catch (JsonProcessingException e) {
            throw new RuntimeException("反序列化打乱选项失败", e);
        }
    }

    public static String convertUserAnswerToOriginal(String userAnswer, ShuffleStorageData shuffleData) {
        if (userAnswer == null || userAnswer.isEmpty() || shuffleData == null) {
            return userAnswer;
        }

        if (shuffleData.getShuffledToOriginalKeyMap() == null || shuffleData.getShuffledToOriginalKeyMap().isEmpty()) {
            return userAnswer;
        }

        String[] userAnswers = userAnswer.split(",");
        List<String> originalAnswers = new ArrayList<>();

        for (String answer : userAnswers) {
            String trimmed = answer.trim();
            if (shuffleData.getShuffledToOriginalKeyMap().containsKey(trimmed)) {
                originalAnswers.add(shuffleData.getShuffledToOriginalKeyMap().get(trimmed));
            } else {
                originalAnswers.add(trimmed);
            }
        }

        Collections.sort(originalAnswers);
        return String.join(",", originalAnswers);
    }

    public static String convertOriginalAnswerToShuffled(String originalAnswer, ShuffleStorageData shuffleData) {
        if (originalAnswer == null || originalAnswer.isEmpty() || shuffleData == null) {
            return originalAnswer;
        }

        if (shuffleData.getOriginalToShuffledKeyMap() == null || shuffleData.getOriginalToShuffledKeyMap().isEmpty()) {
            return originalAnswer;
        }

        String[] originalAnswers = originalAnswer.split(",");
        List<String> shuffledAnswers = new ArrayList<>();

        for (String answer : originalAnswers) {
            String trimmed = answer.trim();
            if (shuffleData.getOriginalToShuffledKeyMap().containsKey(trimmed)) {
                shuffledAnswers.add(shuffleData.getOriginalToShuffledKeyMap().get(trimmed));
            } else {
                shuffledAnswers.add(trimmed);
            }
        }

        Collections.sort(shuffledAnswers);
        return String.join(",", shuffledAnswers);
    }

    @Data
    @Builder
    @AllArgsConstructor
    @NoArgsConstructor
    public static class ShuffledOptions {
        private List<QuestionOption> originalOptions;
        private List<QuestionOption> shuffledOptions;
        private Map<String, String> originalToShuffledKeyMap;
        private Map<String, String> shuffledToOriginalKeyMap;
        private List<String> correctAnswerKeysOriginal;
        private List<String> correctAnswerKeysShuffled;
    }

    @Data
    @Builder
    @AllArgsConstructor
    @NoArgsConstructor
    public static class ShuffleStorageData {
        private Map<String, String> shuffledToOriginalKeyMap;
        private Map<String, String> originalToShuffledKeyMap;
        private List<String> correctAnswerKeysOriginal;
        private List<String> correctAnswerKeysShuffled;
    }
}
