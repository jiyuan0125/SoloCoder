#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "argparse.h"
#include "help.h"
#include "validate.h"

static int cmd_add(const ap_parse_result_t *result) {
    const char *path = ap_result_get_string(result, "path");
    bool verbose = ap_result_get_bool(result, "verbose");
    bool force = ap_result_get_bool(result, "force");
    bool dry_run = ap_result_get_bool(result, "dry-run");
    int verbosity = ap_result_get_count(result, "verbose");
    
    printf("[add] Command executed\n");
    if (path) {
        printf("  Path: %s\n", path);
    }
    printf("  Verbose: %s (level: %d)\n", verbose ? "yes" : "no", verbosity);
    printf("  Force: %s\n", force ? "yes" : "no");
    printf("  Dry run: %s\n", dry_run ? "yes" : "no");
    
    return 0;
}

static int cmd_commit(const ap_parse_result_t *result) {
    const char *message = ap_result_get_string(result, "message");
    const char *author = ap_result_get_string(result, "author");
    bool all = ap_result_get_bool(result, "all");
    bool amend = ap_result_get_bool(result, "amend");
    bool signoff = ap_result_get_bool(result, "signoff");
    
    printf("[commit] Command executed\n");
    if (message) {
        printf("  Message: %s\n", message);
    }
    if (author) {
        printf("  Author: %s\n", author);
    }
    printf("  All: %s\n", all ? "yes" : "no");
    printf("  Amend: %s\n", amend ? "yes" : "no");
    printf("  Signoff: %s\n", signoff ? "yes" : "no");
    
    return 0;
}

static int cmd_status(const ap_parse_result_t *result) {
    bool short_flag = ap_result_get_bool(result, "short");
    bool porcelain = ap_result_get_bool(result, "porcelain");
    bool branch = ap_result_get_bool(result, "branch");
    const char *path = ap_result_get_string(result, "path");
    
    printf("[status] Command executed\n");
    printf("  Short format: %s\n", short_flag ? "yes" : "no");
    printf("  Porcelain: %s\n", porcelain ? "yes" : "no");
    printf("  Show branch: %s\n", branch ? "yes" : "no");
    if (path) {
        printf("  Path: %s\n", path);
    }
    
    return 0;
}

static int cmd_log(const ap_parse_result_t *result) {
    int max_count = ap_result_get_int(result, "max-count");
    bool oneline = ap_result_get_bool(result, "oneline");
    bool graph = ap_result_get_bool(result, "graph");
    bool decorate = ap_result_get_bool(result, "decorate");
    const char *author = ap_result_get_string(result, "author");
    const char *since = ap_result_get_string(result, "since");
    
    printf("[log] Command executed\n");
    printf("  Max count: %d\n", max_count);
    printf("  Oneline: %s\n", oneline ? "yes" : "no");
    printf("  Graph: %s\n", graph ? "yes" : "no");
    printf("  Decorate: %s\n", decorate ? "yes" : "no");
    if (author) {
        printf("  Author: %s\n", author);
    }
    if (since) {
        printf("  Since: %s\n", since);
    }
    
    return 0;
}

static int cmd_remote_add(const ap_parse_result_t *result) {
    const char *name = ap_result_get_string(result, "name");
    const char *url = ap_result_get_string(result, "url");
    bool fetch = ap_result_get_bool(result, "fetch");
    bool mirror = ap_result_get_bool(result, "mirror");
    const char *master = ap_result_get_string(result, "master");
    
    printf("[remote add] Command executed\n");
    printf("  Remote name: %s\n", name ? name : "(none)");
    printf("  URL: %s\n", url ? url : "(none)");
    printf("  Fetch: %s\n", fetch ? "yes" : "no");
    printf("  Mirror: %s\n", mirror ? "yes" : "no");
    if (master) {
        printf("  Master: %s\n", master);
    }
    
    return 0;
}

static int cmd_remote_remove(const ap_parse_result_t *result) {
    const char *name = ap_result_get_string(result, "name");
    
    printf("[remote remove] Command executed\n");
    printf("  Remote name: %s\n", name ? name : "(none)");
    
    return 0;
}

static int cmd_remote_list(const ap_parse_result_t *result) {
    bool verbose = ap_result_get_bool(result, "verbose");
    
    printf("[remote list] Command executed\n");
    printf("  Verbose: %s\n", verbose ? "yes" : "no");
    
    return 0;
}

int main(int argc, char *argv[]) {
    ap_parser_t *parser = ap_parser_create("git", "A fast, scalable, distributed revision control system");
    
    ap_arg_def_t verbose_arg = {
        .name = "verbose",
        .short_name = 'v',
        .type = AP_TYPE_BOOL,
        .description = "Be verbose",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_ACCUMULATE
    };
    ap_parser_add_global_arg(parser, &verbose_arg);
    
    ap_arg_def_t version_arg = {
        .name = "version",
        .short_name = 'V',
        .type = AP_TYPE_BOOL,
        .description = "Print version information and exit",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_parser_add_global_arg(parser, &version_arg);
    
    ap_subcommand_t *add_cmd = ap_subcommand_create("add", "Add file contents to the index");
    ap_arg_def_t add_path_arg = {
        .name = "path",
        .short_name = '\0',
        .type = AP_TYPE_POSITIONAL,
        .description = "Files to add content from",
        .default_value = NULL,
        .required = true,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(add_cmd, &add_path_arg);
    
    ap_arg_def_t add_force_arg = {
        .name = "force",
        .short_name = 'f',
        .type = AP_TYPE_BOOL,
        .description = "Allow adding otherwise ignored files",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(add_cmd, &add_force_arg);
    
    ap_arg_def_t add_dry_run_arg = {
        .name = "dry-run",
        .short_name = 'n',
        .type = AP_TYPE_BOOL,
        .description = "Don't actually add the file(s), just show if they exist",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(add_cmd, &add_dry_run_arg);
    ap_parser_add_subcommand(parser, add_cmd);
    
    ap_subcommand_t *commit_cmd = ap_subcommand_create("commit", "Record changes to the repository");
    ap_arg_def_t commit_msg_arg = {
        .name = "message",
        .short_name = 'm',
        .type = AP_TYPE_STRING,
        .description = "Use the given message as the commit message",
        .default_value = NULL,
        .required = true,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(commit_cmd, &commit_msg_arg);
    
    ap_arg_def_t commit_author_arg = {
        .name = "author",
        .short_name = '\0',
        .type = AP_TYPE_STRING,
        .description = "Override the commit author",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(commit_cmd, &commit_author_arg);
    
    ap_arg_def_t commit_all_arg = {
        .name = "all",
        .short_name = 'a',
        .type = AP_TYPE_BOOL,
        .description = "Tell the command to automatically stage files that have been modified and deleted",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(commit_cmd, &commit_all_arg);
    
    ap_arg_def_t commit_amend_arg = {
        .name = "amend",
        .short_name = '\0',
        .type = AP_TYPE_BOOL,
        .description = "Replace the tip of the current branch by creating a new commit",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(commit_cmd, &commit_amend_arg);
    
    ap_arg_def_t commit_signoff_arg = {
        .name = "signoff",
        .short_name = 's',
        .type = AP_TYPE_BOOL,
        .description = "Add a Signed-off-by trailer by the committer at the end of the commit log message",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(commit_cmd, &commit_signoff_arg);
    ap_parser_add_subcommand(parser, commit_cmd);
    
    ap_subcommand_t *status_cmd = ap_subcommand_create("status", "Show the working tree status");
    ap_arg_def_t status_short_arg = {
        .name = "short",
        .short_name = 's',
        .type = AP_TYPE_BOOL,
        .description = "Give the output in the short-format",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(status_cmd, &status_short_arg);
    
    ap_arg_def_t status_porcelain_arg = {
        .name = "porcelain",
        .short_name = '\0',
        .type = AP_TYPE_BOOL,
        .description = "Give the output in an easy-to-parse format for scripts",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(status_cmd, &status_porcelain_arg);
    
    ap_arg_def_t status_branch_arg = {
        .name = "branch",
        .short_name = 'b',
        .type = AP_TYPE_BOOL,
        .description = "Show the branch and tracking info even in short-format",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(status_cmd, &status_branch_arg);
    
    ap_arg_def_t status_path_arg = {
        .name = "path",
        .short_name = '\0',
        .type = AP_TYPE_POSITIONAL,
        .description = "Limit to specific path",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(status_cmd, &status_path_arg);
    ap_parser_add_subcommand(parser, status_cmd);
    
    ap_subcommand_t *log_cmd = ap_subcommand_create("log", "Show commit logs");
    ap_arg_def_t log_max_count_arg = {
        .name = "max-count",
        .short_name = 'n',
        .type = AP_TYPE_INT,
        .description = "Limit the number of commits to output",
        .default_value = "0",
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_arg_set_range_int(&log_max_count_arg, 1, 10000);
    ap_subcommand_add_arg(log_cmd, &log_max_count_arg);
    
    ap_arg_def_t log_oneline_arg = {
        .name = "oneline",
        .short_name = '\0',
        .type = AP_TYPE_BOOL,
        .description = "This is a shorthand for --pretty=oneline --abbrev-commit",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(log_cmd, &log_oneline_arg);
    
    ap_arg_def_t log_graph_arg = {
        .name = "graph",
        .short_name = 'g',
        .type = AP_TYPE_BOOL,
        .description = "Draw a text-based graphical representation of the commit history",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(log_cmd, &log_graph_arg);
    
    ap_arg_def_t log_decorate_arg = {
        .name = "decorate",
        .short_name = '\0',
        .type = AP_TYPE_BOOL,
        .description = "Print out the ref names of any commits that are shown",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(log_cmd, &log_decorate_arg);
    
    ap_arg_def_t log_author_arg = {
        .name = "author",
        .short_name = '\0',
        .type = AP_TYPE_STRING,
        .description = "Limit the commits output to ones with author header lines that match the specified pattern",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(log_cmd, &log_author_arg);
    
    ap_arg_def_t log_since_arg = {
        .name = "since",
        .short_name = '\0',
        .type = AP_TYPE_STRING,
        .description = "Show commits more recent than a specific date",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(log_cmd, &log_since_arg);
    ap_parser_add_subcommand(parser, log_cmd);
    
    ap_subcommand_t *remote_cmd = ap_subcommand_create("remote", "Manage set of tracked repositories");
    
    ap_subcommand_t *remote_add_cmd = ap_subcommand_create("add", "Add a remote named <name> for the repository at <url>");
    ap_arg_def_t remote_add_name_arg = {
        .name = "name",
        .short_name = '\0',
        .type = AP_TYPE_POSITIONAL,
        .description = "Name of the remote",
        .default_value = NULL,
        .required = true,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(remote_add_cmd, &remote_add_name_arg);
    
    ap_arg_def_t remote_add_url_arg = {
        .name = "url",
        .short_name = '\0',
        .type = AP_TYPE_POSITIONAL,
        .description = "URL of the remote repository",
        .default_value = NULL,
        .required = true,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(remote_add_cmd, &remote_add_url_arg);
    
    ap_arg_def_t remote_add_fetch_arg = {
        .name = "fetch",
        .short_name = 'f',
        .type = AP_TYPE_BOOL,
        .description = "Fetch the remote branches immediately",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(remote_add_cmd, &remote_add_fetch_arg);
    
    ap_arg_def_t remote_add_mirror_arg = {
        .name = "mirror",
        .short_name = 'm',
        .type = AP_TYPE_BOOL,
        .description = "Set up a mirror",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(remote_add_cmd, &remote_add_mirror_arg);
    
    ap_arg_def_t remote_add_master_arg = {
        .name = "master",
        .short_name = 'M',
        .type = AP_TYPE_STRING,
        .description = "Specify a master branch",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(remote_add_cmd, &remote_add_master_arg);
    ap_subcommand_add_subcommand(remote_cmd, remote_add_cmd);
    
    ap_subcommand_t *remote_remove_cmd = ap_subcommand_create("remove", "Remove the remote named <name>");
    ap_arg_def_t remote_remove_name_arg = {
        .name = "name",
        .short_name = '\0',
        .type = AP_TYPE_POSITIONAL,
        .description = "Name of the remote to remove",
        .default_value = NULL,
        .required = true,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(remote_remove_cmd, &remote_remove_name_arg);
    ap_subcommand_add_subcommand(remote_cmd, remote_remove_cmd);
    
    ap_subcommand_t *remote_list_cmd = ap_subcommand_create("list", "List all remotes");
    ap_arg_def_t remote_list_verbose_arg = {
        .name = "verbose",
        .short_name = 'v',
        .type = AP_TYPE_BOOL,
        .description = "Be verbose: show remote URLs after names",
        .default_value = NULL,
        .required = false,
        .dup_policy = AP_DUP_LAST
    };
    ap_subcommand_add_arg(remote_list_cmd, &remote_list_verbose_arg);
    ap_subcommand_add_subcommand(remote_cmd, remote_list_cmd);
    
    ap_parser_add_subcommand(parser, remote_cmd);
    
    ap_parse_result_t *result = ap_parse(parser, argc, argv);
    
    if (result->help_requested) {
        if (result->help_subcommand) {
            ap_help_print_subcommand(parser, result->help_subcommand);
        } else {
            ap_help_print_global(parser);
        }
        ap_parse_result_destroy(result);
        ap_parser_destroy(parser);
        return 0;
    }
    
    if (ap_result_get_bool(result, "version")) {
        printf("git version 2.40.0 (simplified demo)\n");
        ap_parse_result_destroy(result);
        ap_parser_destroy(parser);
        return 0;
    }
    
    ap_valid_error_t error;
    memset(&error, 0, sizeof(error));
    
    ap_valid_status_t status = ap_validate_parse_result(parser, result, &error);
    if (status != AP_VALID_OK) {
        fprintf(stderr, "Error: %s\n", error.message);
        if (result->command_path_length > 0) {
            const ap_subcommand_t *last = result->command_path[result->command_path_length - 1];
            ap_help_print_subcommand(parser, last);
        } else {
            ap_help_print_global(parser);
        }
        ap_parse_result_destroy(result);
        ap_parser_destroy(parser);
        return 1;
    }
    
    int ret = 0;
    
    if (result->command_path_length > 0) {
        const ap_subcommand_t *last = result->command_path[result->command_path_length - 1];
        const char *full_cmd = last->name;
        
        if (result->command_path_length == 2) {
            const ap_subcommand_t *first = result->command_path[0];
            if (strcmp(first->name, "remote") == 0) {
                if (strcmp(full_cmd, "add") == 0) {
                    ret = cmd_remote_add(result);
                } else if (strcmp(full_cmd, "remove") == 0) {
                    ret = cmd_remote_remove(result);
                } else if (strcmp(full_cmd, "list") == 0) {
                    ret = cmd_remote_list(result);
                }
            }
        } else if (result->command_path_length == 1) {
            if (strcmp(full_cmd, "add") == 0) {
                ret = cmd_add(result);
            } else if (strcmp(full_cmd, "commit") == 0) {
                ret = cmd_commit(result);
            } else if (strcmp(full_cmd, "status") == 0) {
                ret = cmd_status(result);
            } else if (strcmp(full_cmd, "log") == 0) {
                ret = cmd_log(result);
            }
        }
    } else {
        if (argc == 1) {
            ap_help_print_global(parser);
        } else {
            fprintf(stderr, "Error: Unknown command or missing command\n\n");
            ap_help_print_global(parser);
            ret = 1;
        }
    }
    
    ap_parse_result_destroy(result);
    ap_parser_destroy(parser);
    
    return ret;
}
