package com.messagequeue.filter;

import com.messagequeue.model.Message;
import lombok.AllArgsConstructor;

@AllArgsConstructor
public class AndExpression implements FilterExpression {
    private final FilterExpression left;
    private final FilterExpression right;

    @Override
    public boolean evaluate(Message message) {
        return left.evaluate(message) && right.evaluate(message);
    }
}
