#ifndef CODE_GENERATOR_H
#define CODE_GENERATOR_H

#include "common.h"

typedef struct CodeGenerator CodeGenerator;

CodeGenerator* code_generator_create(void);
void code_generator_destroy(CodeGenerator* gen);

ReturnCode code_generator_generate_single(CodeGenerator* gen, char* out_code, size_t out_size);
ReturnCode code_generator_generate_batch(CodeGenerator* gen, char** out_codes, int count);
void code_generator_free_batch(char** codes, int count);

bool code_generator_is_valid_char(char c);
int code_generator_char_to_index(char c);

#endif
