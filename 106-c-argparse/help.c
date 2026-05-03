#include "help.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define AP_HELP_INDENT "  "
#define AP_HELP_WIDTH 80

static const char* ap_arg_type_to_string(ap_arg_type_t type) {
    switch (type) {
        case AP_TYPE_BOOL: return "flag";
        case AP_TYPE_STRING: return "string";
        case AP_TYPE_INT: return "integer";
        case AP_TYPE_FLOAT: return "float";
        case AP_TYPE_POSITIONAL: return "positional";
        default: return "unknown";
    }
}

static void ap_help_print_arg(const ap_arg_def_t *arg, int is_positional) {
    if (is_positional) {
        printf("%s<%s>", AP_HELP_INDENT, arg->name);
        if (!arg->required) {
            printf(" (optional)");
        }
    } else {
        int has_short = (arg->short_name != '\0');
        if (has_short) {
            printf("%s-%c", AP_HELP_INDENT, arg->short_name);
            if (arg->name) {
                printf(", --%s", arg->name);
            }
        } else {
            printf("%s    --%s", AP_HELP_INDENT, arg->name);
        }
        
        if (arg->type != AP_TYPE_BOOL) {
            printf(" <%s>", ap_arg_type_to_string(arg->type));
        }
    }
    
    int padding = 25;
    if (is_positional) {
        int len = (int)strlen(arg->name) + 3;
        if (!arg->required) len += 11;
        padding -= len;
    } else {
        int len = 4;
        if (arg->short_name != '\0') len += 3;
        if (arg->name) len += 4 + (int)strlen(arg->name);
        if (arg->type != AP_TYPE_BOOL) {
            len += 1 + (int)strlen(ap_arg_type_to_string(arg->type)) + 1;
        }
        padding -= len;
    }
    
    if (padding < 2) padding = 2;
    for (int i = 0; i < padding; i++) printf(" ");
    
    if (arg->description) {
        printf("%s", arg->description);
    }
    
    if (arg->default_value) {
        printf(" [default: %s]", arg->default_value);
    }
    
    if (arg->required) {
        printf(" [required]");
    }
    
    if (arg->range.has_min || arg->range.has_max) {
        printf(" [");
        if (arg->range.has_min) {
            printf("min: %.0f", arg->range.min);
        }
        if (arg->range.has_min && arg->range.has_max) {
            printf(", ");
        }
        if (arg->range.has_max) {
            printf("max: %.0f", arg->range.max);
        }
        printf("]");
    }
    
    printf("\n");
}

void ap_help_print_subcommand(const ap_parser_t *parser, const ap_subcommand_t *subcmd) {
    if (!parser) return;
    
    printf("Usage: %s", parser->program_name ? parser->program_name : "program");
    
    if (subcmd) {
        printf(" %s", subcmd->name ? subcmd->name : "");
    }
    
    bool has_options = false;
    bool has_positionals = false;
    bool has_subcmds = false;
    
    if (parser->global_args) {
        for (size_t i = 0; i < parser->global_arg_count; i++) {
            if (parser->global_args[i].type == AP_TYPE_POSITIONAL) {
                has_positionals = true;
            } else {
                has_options = true;
            }
        }
    }
    
    if (subcmd && subcmd->arguments) {
        for (size_t i = 0; i < subcmd->arg_count; i++) {
            if (subcmd->arguments[i].type == AP_TYPE_POSITIONAL) {
                has_positionals = true;
            } else {
                has_options = true;
            }
        }
    }
    
    if (subcmd && subcmd->subcommands && subcmd->subcommand_count > 0) {
        has_subcmds = true;
    } else if (!subcmd && parser->subcommands && parser->subcommand_count > 0) {
        has_subcmds = true;
    }
    
    if (has_options) printf(" [options]");
    if (has_subcmds) printf(" <command> [args]");
    if (has_positionals) {
        if (parser->global_args) {
            for (size_t i = 0; i < parser->global_arg_count; i++) {
                if (parser->global_args[i].type == AP_TYPE_POSITIONAL) {
                    if (parser->global_args[i].required) {
                        printf(" <%s>", parser->global_args[i].name);
                    } else {
                        printf(" [<%s>]", parser->global_args[i].name);
                    }
                }
            }
        }
        if (subcmd && subcmd->arguments) {
            for (size_t i = 0; i < subcmd->arg_count; i++) {
                if (subcmd->arguments[i].type == AP_TYPE_POSITIONAL) {
                    if (subcmd->arguments[i].required) {
                        printf(" <%s>", subcmd->arguments[i].name);
                    } else {
                        printf(" [<%s>]", subcmd->arguments[i].name);
                    }
                }
            }
        }
    }
    
    printf("\n\n");
    
    if (subcmd && subcmd->description) {
        printf("%s\n\n", subcmd->description);
    } else if (parser->description) {
        printf("%s\n\n", parser->description);
    }
    
    if (has_options) {
        printf("Options:\n");
        printf("%s-h, --help%sShow this help message and exit\n", 
               AP_HELP_INDENT, "                              ");
        
        if (parser->global_args) {
            for (size_t i = 0; i < parser->global_arg_count; i++) {
                if (parser->global_args[i].type != AP_TYPE_POSITIONAL) {
                    ap_help_print_arg(&parser->global_args[i], 0);
                }
            }
        }
        
        if (subcmd && subcmd->arguments) {
            for (size_t i = 0; i < subcmd->arg_count; i++) {
                if (subcmd->arguments[i].type != AP_TYPE_POSITIONAL) {
                    ap_help_print_arg(&subcmd->arguments[i], 0);
                }
            }
        }
        printf("\n");
    }
    
    if (has_positionals) {
        printf("Positional arguments:\n");
        if (parser->global_args) {
            for (size_t i = 0; i < parser->global_arg_count; i++) {
                if (parser->global_args[i].type == AP_TYPE_POSITIONAL) {
                    ap_help_print_arg(&parser->global_args[i], 1);
                }
            }
        }
        if (subcmd && subcmd->arguments) {
            for (size_t i = 0; i < subcmd->arg_count; i++) {
                if (subcmd->arguments[i].type == AP_TYPE_POSITIONAL) {
                    ap_help_print_arg(&subcmd->arguments[i], 1);
                }
            }
        }
        printf("\n");
    }
    
    if (has_subcmds) {
        printf("Commands:\n");
        if (subcmd && subcmd->subcommands) {
            for (size_t i = 0; i < subcmd->subcommand_count; i++) {
                if (subcmd->subcommands[i]) {
                    printf("%s%s", AP_HELP_INDENT, subcmd->subcommands[i]->name);
                    int padding = 20 - (int)strlen(subcmd->subcommands[i]->name);
                    if (padding < 2) padding = 2;
                    for (int j = 0; j < padding; j++) printf(" ");
                    if (subcmd->subcommands[i]->description) {
                        printf("%s", subcmd->subcommands[i]->description);
                    }
                    printf("\n");
                }
            }
        } else if (!subcmd && parser->subcommands) {
            for (size_t i = 0; i < parser->subcommand_count; i++) {
                if (parser->subcommands[i]) {
                    printf("%s%s", AP_HELP_INDENT, parser->subcommands[i]->name);
                    int padding = 20 - (int)strlen(parser->subcommands[i]->name);
                    if (padding < 2) padding = 2;
                    for (int j = 0; j < padding; j++) printf(" ");
                    if (parser->subcommands[i]->description) {
                        printf("%s", parser->subcommands[i]->description);
                    }
                    printf("\n");
                }
            }
        }
        printf("\n");
        printf("Use '%s <command> --help' for more information about a command.\n",
               parser->program_name ? parser->program_name : "program");
    }
}

void ap_help_print_global(const ap_parser_t *parser) {
    ap_help_print_subcommand(parser, NULL);
}

void ap_help_print(const ap_parser_t *parser, const ap_subcommand_t *subcmd) {
    ap_help_print_subcommand(parser, subcmd);
}

char* ap_help_generate(const ap_parser_t *parser, const ap_subcommand_t *subcmd) {
    (void)parser;
    (void)subcmd;
    return NULL;
}

void ap_help_free(char *help_text) {
    if (help_text) {
        free(help_text);
    }
}
