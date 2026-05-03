#include "argparse.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

static const ap_arg_def_t *ap_find_arg_by_name(const ap_arg_def_t *args, size_t count, const char *name) {
    for (size_t i = 0; i < count; i++) {
        if (args[i].name && strcmp(args[i].name, name) == 0) {
            return &args[i];
        }
    }
    return NULL;
}

static const ap_arg_def_t *ap_find_arg_by_short(const ap_arg_def_t *args, size_t count, char short_name) {
    for (size_t i = 0; i < count; i++) {
        if (args[i].short_name == short_name) {
            return &args[i];
        }
    }
    return NULL;
}

static const ap_subcommand_t *ap_find_subcommand(ap_subcommand_t **subcmds, size_t count, const char *name) {
    for (size_t i = 0; i < count; i++) {
        if (subcmds[i] && subcmds[i]->name && strcmp(subcmds[i]->name, name) == 0) {
            return subcmds[i];
        }
    }
    return NULL;
}

static ap_parsed_arg_t *ap_create_result_array(const ap_arg_def_t *args, size_t count) {
    if (count == 0) return NULL;
    ap_parsed_arg_t *results = (ap_parsed_arg_t*)malloc(count * sizeof(ap_parsed_arg_t));
    if (!results) return NULL;
    
    for (size_t i = 0; i < count; i++) {
        results[i].def = &args[i];
        results[i].string_value = args[i].default_value;
        results[i].numeric.bool_value = false;
        results[i].numeric.int_value = 0;
        results[i].numeric.float_value = 0.0;
        results[i].occurrence_count = 0;
        results[i].is_set = false;
    }
    return results;
}

ap_parser_t* ap_parser_create(const char *program_name, const char *description) {
    ap_parser_t *parser = (ap_parser_t*)malloc(sizeof(ap_parser_t));
    if (!parser) return NULL;
    
    parser->program_name = program_name ? strdup(program_name) : NULL;
    parser->description = description ? strdup(description) : NULL;
    parser->global_args = NULL;
    parser->global_arg_count = 0;
    parser->subcommands = NULL;
    parser->subcommand_count = 0;
    
    return parser;
}

void ap_parser_destroy(ap_parser_t *parser) {
    if (!parser) return;
    
    free((void*)parser->program_name);
    free((void*)parser->description);
    free(parser->global_args);
    
    for (size_t i = 0; i < parser->subcommand_count; i++) {
        if (parser->subcommands[i]) {
            ap_subcommand_destroy(parser->subcommands[i]);
        }
    }
    free(parser->subcommands);
    free(parser);
}

void ap_parser_add_global_arg(ap_parser_t *parser, const ap_arg_def_t *arg) {
    if (!parser || !arg) return;
    
    size_t new_count = parser->global_arg_count + 1;
    ap_arg_def_t *new_args = (ap_arg_def_t*)realloc(parser->global_args, new_count * sizeof(ap_arg_def_t));
    if (!new_args) return;
    
    parser->global_args = new_args;
    memcpy(&parser->global_args[parser->global_arg_count], arg, sizeof(ap_arg_def_t));
    parser->global_arg_count = new_count;
}

void ap_parser_add_subcommand(ap_parser_t *parser, ap_subcommand_t *subcommand) {
    if (!parser || !subcommand) return;
    
    size_t new_count = parser->subcommand_count + 1;
    ap_subcommand_t **new_subcmds = (ap_subcommand_t**)realloc(parser->subcommands, new_count * sizeof(ap_subcommand_t*));
    if (!new_subcmds) return;
    
    parser->subcommands = new_subcmds;
    parser->subcommands[parser->subcommand_count] = subcommand;
    parser->subcommand_count = new_count;
}

ap_subcommand_t* ap_subcommand_create(const char *name, const char *description) {
    ap_subcommand_t *subcmd = (ap_subcommand_t*)malloc(sizeof(ap_subcommand_t));
    if (!subcmd) return NULL;
    
    subcmd->name = name ? strdup(name) : NULL;
    subcmd->description = description ? strdup(description) : NULL;
    subcmd->arguments = NULL;
    subcmd->arg_count = 0;
    subcmd->subcommands = NULL;
    subcmd->subcommand_count = 0;
    
    return subcmd;
}

void ap_subcommand_destroy(ap_subcommand_t *subcmd) {
    if (!subcmd) return;
    
    free((void*)subcmd->name);
    free((void*)subcmd->description);
    free(subcmd->arguments);
    
    for (size_t i = 0; i < subcmd->subcommand_count; i++) {
        if (subcmd->subcommands[i]) {
            ap_subcommand_destroy(subcmd->subcommands[i]);
        }
    }
    free(subcmd->subcommands);
    free(subcmd);
}

void ap_subcommand_add_arg(ap_subcommand_t *subcmd, const ap_arg_def_t *arg) {
    if (!subcmd || !arg) return;
    
    size_t new_count = subcmd->arg_count + 1;
    ap_arg_def_t *new_args = (ap_arg_def_t*)realloc(subcmd->arguments, new_count * sizeof(ap_arg_def_t));
    if (!new_args) return;
    
    subcmd->arguments = new_args;
    memcpy(&subcmd->arguments[subcmd->arg_count], arg, sizeof(ap_arg_def_t));
    subcmd->arg_count = new_count;
}

void ap_subcommand_add_subcommand(ap_subcommand_t *parent, ap_subcommand_t *child) {
    if (!parent || !child) return;
    
    size_t new_count = parent->subcommand_count + 1;
    ap_subcommand_t **new_subcmds = (ap_subcommand_t**)realloc(parent->subcommands, new_count * sizeof(ap_subcommand_t*));
    if (!new_subcmds) return;
    
    parent->subcommands = new_subcmds;
    parent->subcommands[parent->subcommand_count] = child;
    parent->subcommand_count = new_count;
}

static const ap_arg_def_t *ap_find_help_arg(const ap_arg_def_t *args, size_t count, const char *arg) {
    if (strcmp(arg, "--help") == 0) {
        return ap_find_arg_by_name(args, count, "help");
    }
    if (strcmp(arg, "-h") == 0) {
        const ap_arg_def_t *help_arg = ap_find_arg_by_name(args, count, "help");
        if (help_arg && help_arg->short_name == 'h') {
            return help_arg;
        }
    }
    return NULL;
}

ap_parse_result_t* ap_parse(const ap_parser_t *parser, int argc, char *argv[]) {
    if (!parser || argc < 1) return NULL;
    
    ap_parse_result_t *result = (ap_parse_result_t*)malloc(sizeof(ap_parse_result_t));
    if (!result) return NULL;
    
    memset(result, 0, sizeof(ap_parse_result_t));
    
    result->global_results = ap_create_result_array(parser->global_args, parser->global_arg_count);
    result->global_result_count = parser->global_arg_count;
    result->command_path = NULL;
    result->command_path_length = 0;
    result->current_results = NULL;
    result->current_result_count = 0;
    result->help_requested = false;
    result->help_subcommand = NULL;
    
    const ap_subcommand_t *current_subcmd = NULL;
    const ap_subcommand_t **command_path = NULL;
    size_t path_len = 0;
    
    int arg_idx = 1;
    size_t positional_idx = 0;
    
    while (arg_idx < argc) {
        const char *current_arg = argv[arg_idx];
        
        if (strcmp(current_arg, "--help") == 0 || strcmp(current_arg, "-h") == 0) {
            result->help_requested = true;
            result->help_subcommand = current_subcmd;
            arg_idx++;
            continue;
        }
        
        const ap_arg_def_t *all_args[256];
        size_t all_arg_count = 0;
        
        for (size_t i = 0; i < parser->global_arg_count && all_arg_count < 256; i++) {
            all_args[all_arg_count++] = &parser->global_args[i];
        }
        
        if (current_subcmd) {
            for (size_t i = 0; i < current_subcmd->arg_count && all_arg_count < 256; i++) {
                all_args[all_arg_count++] = &current_subcmd->arguments[i];
            }
        }
        
        if (strncmp(current_arg, "--", 2) == 0) {
            const char *arg_name = current_arg + 2;
            const ap_arg_def_t *found = NULL;
            for (size_t i = 0; i < all_arg_count; i++) {
                if (all_args[i]->name && strcmp(all_args[i]->name, arg_name) == 0) {
                    found = all_args[i];
                    break;
                }
            }
            
            if (!found) {
                arg_idx++;
                continue;
            }
            
            ap_parsed_arg_t *res = NULL;
            if (found->type == AP_TYPE_BOOL) {
                for (size_t i = 0; i < result->global_result_count; i++) {
                    if (result->global_results[i].def == found) {
                        res = &result->global_results[i];
                        break;
                    }
                }
                if (!res && current_subcmd) {
                    for (size_t i = 0; i < result->current_result_count; i++) {
                        if (result->current_results[i].def == found) {
                            res = &result->current_results[i];
                            break;
                        }
                    }
                }
                if (res) {
                    res->is_set = true;
                    res->occurrence_count++;
                    res->numeric.bool_value = true;
                }
            } else {
                const char *value = NULL;
                char *eq_pos = strchr(current_arg, '=');
                if (eq_pos) {
                    value = eq_pos + 1;
                } else if (arg_idx + 1 < argc) {
                    arg_idx++;
                    value = argv[arg_idx];
                }
                
                if (res == NULL) {
                    for (size_t i = 0; i < result->global_result_count; i++) {
                        if (result->global_results[i].def == found) {
                            res = &result->global_results[i];
                            break;
                        }
                    }
                }
                if (!res && current_subcmd) {
                    for (size_t i = 0; i < result->current_result_count; i++) {
                        if (result->current_results[i].def == found) {
                            res = &result->current_results[i];
                            break;
                        }
                    }
                }
                if (res && value) {
                    res->is_set = true;
                    res->string_value = value;
                    if (found->dup_policy == AP_DUP_ACCUMULATE) {
                        res->occurrence_count++;
                    } else {
                        res->occurrence_count = 1;
                    }
                    
                    if (found->type == AP_TYPE_INT) {
                        res->numeric.int_value = atoi(value);
                    } else if (found->type == AP_TYPE_FLOAT) {
                        res->numeric.float_value = atof(value);
                    }
                }
            }
            arg_idx++;
        } else if (current_arg[0] == '-' && current_arg[1] != '\0') {
            int char_idx = 1;
            while (current_arg[char_idx] != '\0') {
                char short_opt = current_arg[char_idx];
                const ap_arg_def_t *found = NULL;
                for (size_t i = 0; i < all_arg_count; i++) {
                    if (all_args[i]->short_name == short_opt) {
                        found = all_args[i];
                        break;
                    }
                }
                
                if (found) {
                    ap_parsed_arg_t *res = NULL;
                    for (size_t i = 0; i < result->global_result_count; i++) {
                        if (result->global_results[i].def == found) {
                            res = &result->global_results[i];
                            break;
                        }
                    }
                    if (!res && current_subcmd) {
                        for (size_t i = 0; i < result->current_result_count; i++) {
                            if (result->current_results[i].def == found) {
                                res = &result->current_results[i];
                                break;
                            }
                        }
                    }
                    
                    if (res) {
                        if (found->type == AP_TYPE_BOOL) {
                            res->is_set = true;
                            res->occurrence_count++;
                            res->numeric.bool_value = true;
                        } else {
                            const char *value = NULL;
                            if (current_arg[char_idx + 1] != '\0') {
                                value = &current_arg[char_idx + 1];
                                char_idx = strlen(current_arg);
                            } else if (arg_idx + 1 < argc) {
                                arg_idx++;
                                value = argv[arg_idx];
                            }
                            
                            if (value) {
                                res->is_set = true;
                                res->string_value = value;
                                if (found->dup_policy == AP_DUP_ACCUMULATE) {
                                    res->occurrence_count++;
                                } else {
                                    res->occurrence_count = 1;
                                }
                                
                                if (found->type == AP_TYPE_INT) {
                                    res->numeric.int_value = atoi(value);
                                } else if (found->type == AP_TYPE_FLOAT) {
                                    res->numeric.float_value = atof(value);
                                }
                            }
                        }
                    }
                }
                char_idx++;
            }
            arg_idx++;
        } else {
            const ap_subcommand_t *found_subcmd = NULL;
            
            if (!current_subcmd && parser->subcommands) {
                found_subcmd = ap_find_subcommand(parser->subcommands, parser->subcommand_count, current_arg);
            } else if (current_subcmd && current_subcmd->subcommands) {
                found_subcmd = ap_find_subcommand(current_subcmd->subcommands, current_subcmd->subcommand_count, current_arg);
            }
            
            if (found_subcmd) {
                size_t new_path_len = path_len + 1;
                const ap_subcommand_t **new_path = (const ap_subcommand_t**)realloc(
                    command_path, new_path_len * sizeof(const ap_subcommand_t*)
                );
                if (new_path) {
                    command_path = new_path;
                    command_path[path_len] = found_subcmd;
                    path_len = new_path_len;
                    current_subcmd = found_subcmd;
                    
                    free(result->current_results);
                    result->current_results = ap_create_result_array(current_subcmd->arguments, current_subcmd->arg_count);
                    result->current_result_count = current_subcmd->arg_count;
                }
                arg_idx++;
                positional_idx = 0;
            } else {
                ap_arg_def_t positional_defs[64];
                size_t positional_count = 0;
                
                for (size_t i = 0; i < parser->global_arg_count; i++) {
                    if (parser->global_args[i].type == AP_TYPE_POSITIONAL) {
                        positional_defs[positional_count++] = parser->global_args[i];
                    }
                }
                if (current_subcmd) {
                    for (size_t i = 0; i < current_subcmd->arg_count; i++) {
                        if (current_subcmd->arguments[i].type == AP_TYPE_POSITIONAL) {
                            positional_defs[positional_count++] = current_subcmd->arguments[i];
                        }
                    }
                }
                
                if (positional_idx < positional_count) {
                    const ap_arg_def_t *def = &positional_defs[positional_idx];
                    ap_parsed_arg_t *res = NULL;
                    
                    for (size_t i = 0; i < result->global_result_count; i++) {
                        if (result->global_results[i].def->name && def->name &&
                            strcmp(result->global_results[i].def->name, def->name) == 0) {
                            res = &result->global_results[i];
                            break;
                        }
                    }
                    if (!res && current_subcmd) {
                        for (size_t i = 0; i < result->current_result_count; i++) {
                            if (result->current_results[i].def->name && def->name &&
                                strcmp(result->current_results[i].def->name, def->name) == 0) {
                                res = &result->current_results[i];
                                break;
                            }
                        }
                    }
                    
                    if (res) {
                        res->is_set = true;
                        res->string_value = current_arg;
                        res->occurrence_count = 1;
                        
                        if (def->type == AP_TYPE_POSITIONAL) {
                        } else if (def->type == AP_TYPE_INT) {
                            res->numeric.int_value = atoi(current_arg);
                        } else if (def->type == AP_TYPE_FLOAT) {
                            res->numeric.float_value = atof(current_arg);
                        }
                    }
                    positional_idx++;
                }
                arg_idx++;
            }
        }
    }
    
    result->command_path = command_path;
    result->command_path_length = path_len;
    
    return result;
}

void ap_parse_result_destroy(ap_parse_result_t *result) {
    if (!result) return;
    free(result->global_results);
    free(result->current_results);
    free(result->command_path);
    free(result);
}

const ap_parsed_arg_t* ap_result_get_arg(const ap_parse_result_t *result, const char *name) {
    if (!result || !name) return NULL;
    
    for (size_t i = 0; i < result->current_result_count; i++) {
        if (result->current_results[i].def->name && 
            strcmp(result->current_results[i].def->name, name) == 0) {
            return &result->current_results[i];
        }
    }
    
    for (size_t i = 0; i < result->global_result_count; i++) {
        if (result->global_results[i].def->name && 
            strcmp(result->global_results[i].def->name, name) == 0) {
            return &result->global_results[i];
        }
    }
    
    return NULL;
}

bool ap_result_is_set(const ap_parse_result_t *result, const char *name) {
    const ap_parsed_arg_t *arg = ap_result_get_arg(result, name);
    return arg ? arg->is_set : false;
}

const char* ap_result_get_string(const ap_parse_result_t *result, const char *name) {
    const ap_parsed_arg_t *arg = ap_result_get_arg(result, name);
    return arg ? arg->string_value : NULL;
}

int ap_result_get_int(const ap_parse_result_t *result, const char *name) {
    const ap_parsed_arg_t *arg = ap_result_get_arg(result, name);
    return arg ? arg->numeric.int_value : 0;
}

double ap_result_get_float(const ap_parse_result_t *result, const char *name) {
    const ap_parsed_arg_t *arg = ap_result_get_arg(result, name);
    return arg ? arg->numeric.float_value : 0.0;
}

bool ap_result_get_bool(const ap_parse_result_t *result, const char *name) {
    const ap_parsed_arg_t *arg = ap_result_get_arg(result, name);
    return arg ? arg->numeric.bool_value : false;
}

int ap_result_get_count(const ap_parse_result_t *result, const char *name) {
    const ap_parsed_arg_t *arg = ap_result_get_arg(result, name);
    return arg ? arg->occurrence_count : 0;
}

void ap_arg_set_range_int(ap_arg_def_t *arg, int min, int max) {
    if (!arg) return;
    arg->range.has_min = true;
    arg->range.has_max = true;
    arg->range.min = (double)min;
    arg->range.max = (double)max;
}

void ap_arg_set_range_float(ap_arg_def_t *arg, double min, double max) {
    if (!arg) return;
    arg->range.has_min = true;
    arg->range.has_max = true;
    arg->range.min = min;
    arg->range.max = max;
}

void ap_arg_set_required(ap_arg_def_t *arg, bool required) {
    if (arg) arg->required = required;
}

void ap_arg_set_default(ap_arg_def_t *arg, const char *default_value) {
    if (arg) arg->default_value = default_value;
}

void ap_arg_set_dup_policy(ap_arg_def_t *arg, ap_dup_policy_t policy) {
    if (arg) arg->dup_policy = policy;
}
