#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <strings.h>
#include "lineedit.h"
#include "history.h"
#include "terminal.h"

static const char *keywords[] = {
    "SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER",
    "FROM", "WHERE", "ORDER BY", "GROUP BY", "JOIN", "INNER JOIN",
    "LEFT JOIN", "RIGHT JOIN", "FULL JOIN", "ON", "AS", "DISTINCT",
    "LIMIT", "OFFSET", "HAVING", "UNION", "ALL", "AND", "OR", "NOT",
    "IN", "LIKE", "BETWEEN", "IS NULL", "IS NOT NULL", "EXISTS",
    "COUNT", "SUM", "AVG", "MIN", "MAX", "UPPER", "LOWER", "SUBSTR",
    "TRIM", "LENGTH", "ROUND", "NOW", "CURRENT_DATE", "CURRENT_TIME",
    "TABLE", "DATABASE", "INDEX", "VIEW", "TRIGGER", "PROCEDURE",
    "FUNCTION", "PRIMARY KEY", "FOREIGN KEY", "UNIQUE", "NOT NULL",
    "DEFAULT", "AUTO_INCREMENT", "COMMIT", "ROLLBACK", "BEGIN",
    "SHOW", "DESCRIBE", "EXPLAIN", "USE", "HELP", "QUIT", "EXIT",
    "HISTORY", NULL
};

static const char *tables[] = {
    "users", "products", "orders", "categories", "order_items",
    "addresses", "reviews", "inventory", "employees", "departments",
    "salaries", "projects", "assignments", NULL
};

static char **sql_completion(const char *line, int *count) {
    *count = 0;
    
    if (line == NULL || line[0] == '\0') {
        return NULL;
    }
    
    size_t line_len = strlen(line);
    const char *last_word = line;
    for (int i = (int)line_len - 1; i >= 0; i--) {
        if (line[i] == ' ' || line[i] == '\t' || line[i] == '(' || line[i] == ',') {
            last_word = &line[i + 1];
            break;
        }
    }
    
    size_t prefix_len = strlen(last_word);
    if (prefix_len == 0) {
        return NULL;
    }
    
    int total = 0;
    int i;
    for (i = 0; keywords[i]; i++) total++;
    for (i = 0; tables[i]; i++) total++;
    
    char **matches = (char **)malloc(sizeof(char *) * total);
    int match_count = 0;
    
    for (i = 0; keywords[i]; i++) {
        if (strncasecmp(keywords[i], last_word, prefix_len) == 0) {
            matches[match_count] = strdup(line);
            size_t new_len = (last_word - line) + strlen(keywords[i]);
            if (new_len >= strlen(matches[match_count])) {
                free(matches[match_count]);
                matches[match_count] = (char *)malloc(new_len + 2);
                strncpy(matches[match_count], line, last_word - line);
                matches[match_count][last_word - line] = '\0';
            }
            strcat(matches[match_count], keywords[i]);
            match_count++;
        }
    }
    
    for (i = 0; tables[i]; i++) {
        if (strncasecmp(tables[i], last_word, prefix_len) == 0) {
            matches[match_count] = strdup(line);
            size_t new_len = (last_word - line) + strlen(tables[i]);
            if (new_len >= strlen(matches[match_count])) {
                free(matches[match_count]);
                matches[match_count] = (char *)malloc(new_len + 2);
                strncpy(matches[match_count], line, last_word - line);
                matches[match_count][last_word - line] = '\0';
            }
            strcat(matches[match_count], tables[i]);
            match_count++;
        }
    }
    
    if (match_count == 0) {
        free(matches);
        *count = 0;
        return NULL;
    }
    
    *count = match_count;
    return matches;
}

static void print_history(const History *h) {
    int count = history_get_count(h);
    if (count == 0) {
        printf("History is empty.\n");
        return;
    }
    printf("\n--- Command History (%d entries) ---\n", count);
    for (int i = 0; i < count; i++) {
        printf("%5d  %s\n", i + 1, history_get(h, i));
    }
    printf("--------------------------------------\n\n");
}

static void print_help(void) {
    printf("\n");
    printf("SQL Shell - Interactive Command Line\n");
    printf("======================================\n\n");
    printf("Editing Keys:\n");
    printf("  Ctrl+A / Home   - Go to beginning of line\n");
    printf("  Ctrl+E / End    - Go to end of line\n");
    printf("  Ctrl+F / Right  - Move forward one character\n");
    printf("  Ctrl+B / Left   - Move backward one character\n");
    printf("  Ctrl+U          - Kill from cursor to beginning\n");
    printf("  Ctrl+K          - Kill from cursor to end\n");
    printf("  Ctrl+C          - Cancel current line\n");
    printf("  Ctrl+D (empty)  - Exit shell\n");
    printf("  Up/Down arrows  - Navigate history\n");
    printf("  Tab             - Auto-complete\n");
    printf("\n");
    printf("Commands:\n");
    printf("  help            - Show this help\n");
    printf("  history         - Show command history\n");
    printf("  quit / exit     - Exit the shell\n");
    printf("\n");
}

int main(void) {
    History history;
    LineEdit le;
    
    history_init(&history);
    lineedit_init(&le, &history);
    lineedit_set_prompt(&le, "sql> ");
    lineedit_set_complete_callback(sql_completion);
    
    printf("Welcome to SQL Shell Demo\n");
    printf("Type 'help' for help, 'quit' to exit.\n\n");
    
    while (1) {
        char *line = lineedit_read(&le);
        
        if (line == NULL) {
            continue;
        }
        
        if (strcasecmp(line, "quit") == 0 || strcasecmp(line, "exit") == 0) {
            free(line);
            printf("Bye!\n");
            break;
        }
        
        if (strcasecmp(line, "help") == 0) {
            print_help();
            free(line);
            continue;
        }
        
        if (strcasecmp(line, "history") == 0) {
            print_history(&history);
            free(line);
            continue;
        }
        
        if (line[0] != '\0') {
            printf("Executing: %s\n", line);
        }
        
        free(line);
    }
    
    lineedit_free(&le);
    history_free(&history);
    
    return 0;
}
