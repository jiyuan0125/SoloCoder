package com.messagequeue.filter;

import com.messagequeue.model.Message;
import lombok.AllArgsConstructor;

@AllArgsConstructor
public class HeaderEqualsExpression implements FilterExpression {
    private final String headerName;
    private final String expectedValue;

    @Override
    public boolean evaluate(Message message) {
        String actualValue = message.getHeaders().get(headerName);
        return expectedValue.equals(actualValue);
    }
}
