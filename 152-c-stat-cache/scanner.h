#ifndef SCANNER_H
#define SCANNER_H

#include "snapshot.h"

typedef struct ScannerOptions {
    int follow_symlinks;
    int include_hidden;
} ScannerOptions;

Snapshot *scanner_scan_directory(const char *root_path, ScannerOptions *options);

#endif
