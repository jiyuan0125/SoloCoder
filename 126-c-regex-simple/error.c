#include <stdio.h>
#include <string.h>
#include "regex.h"

const char *regex_strerror(RegexError err) {
    switch (err) {
        case REGEX_OK:
            return "Success";
            
        case REGEX_ERR_SYNTAX:
            return "Syntax error";
            
        case REGEX_ERR_MEMORY:
            return "Memory allocation failed";
            
        case REGEX_ERR_UNMATCHED_BRACKET:
            return "Unmatched '[' - missing closing ']'";
            
        case REGEX_ERR_UNMATCHED_PAREN:
            return "Unmatched parenthesis";
            
        case REGEX_ERR_MISSING_PREV:
            return "Quantifier (*, +, ?) without preceding character";
            
        case REGEX_ERR_INVALID_RANGE:
            return "Invalid character range (e.g., [z-a])";
            
        case REGEX_ERR_INVALID_ESCAPE:
            return "Incomplete escape sequence";
            
        case REGEX_NO_MATCH:
            return "No match found";
            
        default:
            return "Unknown error";
    }
}
