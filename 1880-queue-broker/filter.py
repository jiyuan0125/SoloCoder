import re
import json
from typing import Any, Optional


class FilterExpression:
    def __init__(self, expression: str):
        self.expression = expression.strip()
        self.tokens = self._tokenize()
        self.position = 0

    def _tokenize(self) -> list:
        tokens = []
        i = 0
        while i < len(self.expression):
            char = self.expression[i]
            
            if char.isspace():
                i += 1
                continue
            
            if char == '(':
                tokens.append(('LPAREN', '('))
                i += 1
            elif char == ')':
                tokens.append(('RPAREN', ')'))
                i += 1
            elif char in ('>', '<', '=', '!'):
                if i + 1 < len(self.expression) and self.expression[i+1] == '=':
                    tokens.append(('OPERATOR', char + '='))
                    i += 2
                else:
                    tokens.append(('OPERATOR', char))
                    i += 1
            elif char == '"':
                j = i + 1
                value = ''
                while j < len(self.expression) and self.expression[j] != '"':
                    value += self.expression[j]
                    j += 1
                tokens.append(('STRING', value))
                i = j + 1
            elif char.isdigit() or char == '-' and i + 1 < len(self.expression) and self.expression[i+1].isdigit():
                j = i
                if char == '-':
                    j += 1
                while j < len(self.expression) and (self.expression[j].isdigit() or self.expression[j] == '.'):
                    j += 1
                num_str = self.expression[i:j]
                if '.' in num_str:
                    tokens.append(('NUMBER', float(num_str)))
                else:
                    tokens.append(('NUMBER', int(num_str)))
                i = j
            elif char.isalpha() or char == '_':
                j = i
                while j < len(self.expression) and (self.expression[j].isalnum() or self.expression[j] == '_'):
                    j += 1
                word = self.expression[i:j]
                if word.lower() in ('and', 'or', 'contains', 'not', 'true', 'false'):
                    tokens.append((word.upper(), word.lower()))
                else:
                    tokens.append(('IDENTIFIER', word))
                i = j
            else:
                i += 1
        
        return tokens

    def _current(self) -> Optional[tuple]:
        if self.position < len(self.tokens):
            return self.tokens[self.position]
        return None

    def _consume(self) -> Optional[tuple]:
        if self.position < len(self.tokens):
            token = self.tokens[self.position]
            self.position += 1
            return token
        return None

    def _match(self, token_type: str) -> bool:
        current = self._current()
        return current and current[0] == token_type

    def parse(self, context: dict) -> bool:
        self.position = 0
        return self._parse_or(context)

    def _parse_or(self, context: dict) -> bool:
        left = self._parse_and(context)
        while self._match('OR'):
            self._consume()
            right = self._parse_and(context)
            left = left or right
        return left

    def _parse_and(self, context: dict) -> bool:
        left = self._parse_not(context)
        while self._match('AND'):
            self._consume()
            right = self._parse_not(context)
            left = left and right
        return left

    def _parse_not(self, context: dict) -> bool:
        if self._match('NOT'):
            self._consume()
            return not self._parse_comparison(context)
        return self._parse_comparison(context)

    def _parse_comparison(self, context: dict) -> bool:
        if self._match('LPAREN'):
            self._consume()
            result = self._parse_or(context)
            if self._match('RPAREN'):
                self._consume()
            return result
        
        left = self._parse_value(context)
        
        if self._match('OPERATOR'):
            operator = self._consume()[1]
            right = self._parse_value(context)
            return self._apply_comparison(left, operator, right)
        elif self._match('CONTAINS'):
            self._consume()
            right = self._parse_value(context)
            if isinstance(left, str) and isinstance(right, str):
                return right in left
            return False
        
        return bool(left) if left is not None else False

    def _parse_value(self, context: dict) -> Any:
        if self._match('NUMBER'):
            return self._consume()[1]
        elif self._match('STRING'):
            return self._consume()[1]
        elif self._match('TRUE'):
            self._consume()
            return True
        elif self._match('FALSE'):
            self._consume()
            return False
        elif self._match('IDENTIFIER'):
            name = self._consume()[1]
            return self._get_value(context, name)
        return None

    def _get_value(self, context: dict, name: str) -> Any:
        if name in context:
            return context[name]
        return None

    def _apply_comparison(self, left: Any, operator: str, right: Any) -> bool:
        try:
            if operator == '>':
                return left > right
            elif operator == '<':
                return left < right
            elif operator == '>=':
                return left >= right
            elif operator == '<=':
                return left <= right
            elif operator == '=':
                return left == right
            elif operator == '!=':
                return left != right
        except (TypeError, ValueError):
            return False
        return False


def evaluate_filter(expression: str, content: str) -> bool:
    try:
        context = json.loads(content)
    except json.JSONDecodeError:
        context = {'content': content}
    
    filter_expr = FilterExpression(expression)
    return filter_expr.parse(context)
