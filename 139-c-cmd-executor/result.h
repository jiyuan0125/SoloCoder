#ifndef RESULT_H
#define RESULT_H

#include "common.h"

const char *status_to_string(cmd_status_t status);
const char *status_to_color(cmd_status_t status);
void print_color(const char *text, const char *color);

int calculate_summary(cmd_result_t *results, size_t count, summary_stats_t *stats);
void print_summary(summary_stats_t *stats);

void print_result_summary(cmd_result_t *result, int summary_lines);
void print_all_results(cmd_result_t *results, size_t count, int summary_lines);

void print_complete_report(cmd_result_t *results, size_t count, int summary_lines);

char **split_lines(const char *buf, size_t len, size_t *line_count);
void free_lines(char **lines, size_t count);

#endif
