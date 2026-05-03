#ifndef AP_HELP_H
#define AP_HELP_H

#include "argparse.h"

#ifdef __cplusplus
extern "C" {
#endif

void ap_help_print(const ap_parser_t *parser, const ap_subcommand_t **command_path, size_t path_length);
void ap_help_print_global(const ap_parser_t *parser);
void ap_help_print_subcommand(const ap_parser_t *parser, const ap_subcommand_t **command_path, size_t path_length);

char* ap_help_generate(const ap_parser_t *parser, const ap_subcommand_t **command_path, size_t path_length);
void ap_help_free(char *help_text);

#ifdef __cplusplus
}
#endif

#endif
