#ifndef DIR_SCAN_H
#define DIR_SCAN_H

#include "common.h"

FileList *scan_directory(const char *path, const ScanConfig *config);

void scan_config_init(ScanConfig *config);
void scan_config_free(ScanConfig *config);

#endif
