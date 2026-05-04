#ifndef DUPLICATE_H
#define DUPLICATE_H

#include "common.h"

DuplicateReport *find_duplicates(FileList *files);

void print_duplicate_report(const DuplicateReport *report);

typedef void (*report_callback_t)(const DuplicateGroup *group, void *user_data);
void iterate_duplicates(const DuplicateReport *report, report_callback_t callback, void *user_data);

#endif
