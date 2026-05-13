package com.messagequeue.filter;

import com.messagequeue.model.Message;

public interface FilterExpression {
    boolean evaluate(Message message);
}
