#include <stdio.h>
#include <stdlib.h>
#include "ini_parser.h"

int main() {
    printf("Testing direct parse of included_config.ini...\n");
    
    ini_parser_t* parser = ini_parse("included_config.ini");
    if (!parser) {
        fprintf(stderr, "Parse failed: line %d: %s\n", 
                ini_get_error_line(parser), ini_get_error_message(parser));
        return 1;
    }
    
    const char* val = ini_get(parser, "included_section", "included_key", "not_found");
    printf("  included_section.included_key = %s\n", val);
    
    ini_free(parser);
    return 0;
}
