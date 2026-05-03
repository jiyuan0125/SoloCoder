#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "csv_parser.h"
#include "csv_types.h"

static int g_row_count = 0;

typedef struct {
    int has_error;
    int error_field;
    const char *error_message;
} TypeConvertContext;

static void OnRowCallback(const CSV_Row *row, void *user_data) {
    (void)user_data;
    g_row_count++;

    printf("=== 第 %d 行 (字段数: %zu, 空行: %s) ===\n",
           g_row_count, row->field_count, row->is_empty ? "是" : "否");

    for (size_t i = 0; i < row->field_count; i++) {
        const char *value = CSV_RowGetField(row, i);
        if (value) {
            printf("  字段[%zu]: \"%s\"\n", i, value);
        } else {
            printf("  字段[%zu]: (NULL)\n", i);
        }
    }
    printf("\n");
}

static void TestBasicParsing(void) {
    printf("\n========================================\n");
    printf("测试1: 基本CSV解析\n");
    printf("========================================\n\n");

    const char *csv = "id,name,price,quantity\n"
                      "1,Apple,5.99,100\n"
                      "2,Banana,3.49,200\n"
                      "3,Orange,4.25,150";

    g_row_count = 0;
    CSV_Parser *parser = CSV_ParserCreate();
    CSV_ParserSetDelimiter(parser, CSV_DELIMITER_COMMA);
    CSV_ParserSetEmptyLinePolicy(parser, CSV_EMPTY_LINE_SKIP);
    CSV_ParserSetRowCallback(parser, OnRowCallback, NULL);

    CSV_ParserFeed(parser, csv, strlen(csv));
    CSV_ParserFinish(parser);

    CSV_ParserDestroy(parser);
}

static void TestQuotedFields(void) {
    printf("\n========================================\n");
    printf("测试2: 引号处理（分隔符、换行符、转义引号）\n");
    printf("========================================\n\n");

    const char *csv = "product_id,description,notes\n"
                      "101,\"Apple, Inc.\",\"Normal product\"\n"
                      "102,\"Say \"\"Hello\"\" World\",\"Contains quotes\"\n"
                      "103,\"Line1\nLine2\nLine3\",\"Multi-line field\"\n"
                      "104,\"A,B,C\",\"Commas inside\"";

    g_row_count = 0;
    CSV_Parser *parser = CSV_ParserCreate();
    CSV_ParserSetDelimiter(parser, CSV_DELIMITER_COMMA);
    CSV_ParserSetEmptyLinePolicy(parser, CSV_EMPTY_LINE_SKIP);
    CSV_ParserSetRowCallback(parser, OnRowCallback, NULL);

    CSV_ParserFeed(parser, csv, strlen(csv));
    CSV_ParserFinish(parser);

    CSV_ParserDestroy(parser);
}

static void TestMixedNewlines(void) {
    printf("\n========================================\n");
    printf("测试3: 混合换行符 (\\n, \\r\\n, \\r)\n");
    printf("========================================\n\n");

    char csv[256];
    int pos = 0;

    const char *line1 = "field1,field2,field3";
    memcpy(csv + pos, line1, strlen(line1));
    pos += strlen(line1);
    csv[pos++] = '\n';

    const char *line2 = "a,b,c";
    memcpy(csv + pos, line2, strlen(line2));
    pos += strlen(line2);
    csv[pos++] = '\r';
    csv[pos++] = '\n';

    const char *line3 = "x,y,z";
    memcpy(csv + pos, line3, strlen(line3));
    pos += strlen(line3);
    csv[pos++] = '\r';

    const char *line4 = "1,2,3";
    memcpy(csv + pos, line4, strlen(line4));
    pos += strlen(line4);
    csv[pos] = '\0';

    g_row_count = 0;
    CSV_Parser *parser = CSV_ParserCreate();
    CSV_ParserSetDelimiter(parser, CSV_DELIMITER_COMMA);
    CSV_ParserSetEmptyLinePolicy(parser, CSV_EMPTY_LINE_SKIP);
    CSV_ParserSetRowCallback(parser, OnRowCallback, NULL);

    CSV_ParserFeed(parser, csv, pos);
    CSV_ParserFinish(parser);

    CSV_ParserDestroy(parser);
}

static void TestEmptyLines(void) {
    printf("\n========================================\n");
    printf("测试4: 空行处理（保留模式）\n");
    printf("========================================\n\n");

    const char *csv = "header1,header2\n"
                      "\n"
                      "data1,data2\n"
                      "\n"
                      "\n"
                      "data3,data4\n";

    g_row_count = 0;
    CSV_Parser *parser = CSV_ParserCreate();
    CSV_ParserSetDelimiter(parser, CSV_DELIMITER_COMMA);
    CSV_ParserSetEmptyLinePolicy(parser, CSV_EMPTY_LINE_KEEP);
    CSV_ParserSetRowCallback(parser, OnRowCallback, NULL);

    CSV_ParserFeed(parser, csv, strlen(csv));
    CSV_ParserFinish(parser);

    CSV_ParserDestroy(parser);

    printf("\n--- 同样的数据，跳过空行模式 ---\n\n");

    g_row_count = 0;
    parser = CSV_ParserCreate();
    CSV_ParserSetDelimiter(parser, CSV_DELIMITER_COMMA);
    CSV_ParserSetEmptyLinePolicy(parser, CSV_EMPTY_LINE_SKIP);
    CSV_ParserSetRowCallback(parser, OnRowCallback, NULL);

    CSV_ParserFeed(parser, csv, strlen(csv));
    CSV_ParserFinish(parser);

    CSV_ParserDestroy(parser);
}

static void TestStreaming(void) {
    printf("\n========================================\n");
    printf("测试5: 流式解析（模拟大文件分块读取）\n");
    printf("========================================\n\n");

    const char *csv_chunks[] = {
        "id,name,price\n",
        "1,\"Product ",
        "A\",10.50\n",
        "2,\"Product ",
        "B\",20.",
        "00\n",
        "3,Normal,15."
    };
    const int chunk_count = 7;

    g_row_count = 0;
    CSV_Parser *parser = CSV_ParserCreate();
    CSV_ParserSetDelimiter(parser, CSV_DELIMITER_COMMA);
    CSV_ParserSetEmptyLinePolicy(parser, CSV_EMPTY_LINE_SKIP);
    CSV_ParserSetRowCallback(parser, OnRowCallback, NULL);

    for (int i = 0; i < chunk_count; i++) {
        printf("--- 提供数据块 %d: \"%s\" ---\n",
               i + 1, csv_chunks[i]);
        CSV_ParserFeed(parser, csv_chunks[i], strlen(csv_chunks[i]));
    }

    printf("--- 调用 Finish 处理剩余数据 ---\n");
    CSV_ParserFinish(parser);

    CSV_ParserDestroy(parser);
}

static void TestTypeConversion(void) {
    printf("\n========================================\n");
    printf("测试6: 字段类型转换\n");
    printf("========================================\n\n");

    const char *test_cases[] = {
        "123",
        "-456",
        "3.14159",
        "2.5e10",
        "",
        "  ",
        "abc",
        "9999999999999999999",
        NULL
    };

    for (int i = 0; test_cases[i] != NULL; i++) {
        const char *str = test_cases[i];
        printf("测试值: \"%s\"\n", str ? str : "(NULL)");

        if (str != NULL) {
            int int_val = 0;
            CSV_TypeResult int_result = CSV_ConvertToInt(str, &int_val);
            printf("  转int:  ");
            if (CSV_TypeResultIsSuccess(int_result)) {
                printf("成功, 值=%d\n", int_val);
            } else {
                printf("失败, 状态=%s\n", CSV_TypeStatusToString(int_result.status));
            }

            long long ll_val = 0;
            CSV_TypeResult ll_result = CSV_ConvertToLongLong(str, &ll_val);
            printf("  转long long:  ");
            if (CSV_TypeResultIsSuccess(ll_result)) {
                printf("成功, 值=%lld\n", ll_val);
            } else {
                printf("失败, 状态=%s\n", CSV_TypeStatusToString(ll_result.status));
            }

            double d_val = 0.0;
            CSV_TypeResult d_result = CSV_ConvertToDouble(str, &d_val);
            printf("  转double:  ");
            if (CSV_TypeResultIsSuccess(d_result)) {
                printf("成功, 值=%g\n", d_val);
            } else {
                printf("失败, 状态=%s\n", CSV_TypeStatusToString(d_result.status));
            }
        } else {
            CSV_TypeResult result = CSV_ConvertToInt(NULL, NULL);
            printf("  状态=%s\n", CSV_TypeStatusToString(result.status));
        }
        printf("\n");
    }
}

static void TestTabDelimiter(void) {
    printf("\n========================================\n");
    printf("测试7: 制表符分隔模式\n");
    printf("========================================\n\n");

    const char *csv = "id\tname\tprice\n"
                      "1\tApple\t5.99\n"
                      "2\t\"Banana, Yellow\"\t3.49\n"
                      "3\tOrange\t4.25";

    g_row_count = 0;
    CSV_Parser *parser = CSV_ParserCreate();
    CSV_ParserSetDelimiter(parser, CSV_DELIMITER_TAB);
    CSV_ParserSetEmptyLinePolicy(parser, CSV_EMPTY_LINE_SKIP);
    CSV_ParserSetRowCallback(parser, OnRowCallback, NULL);

    CSV_ParserFeed(parser, csv, strlen(csv));
    CSV_ParserFinish(parser);

    CSV_ParserDestroy(parser);
}

static void TestEdgeCases(void) {
    printf("\n========================================\n");
    printf("测试8: 其他边界情况\n");
    printf("========================================\n\n");

    const char *csv = ",,,\n"
                      "a,,b,c\n"
                      ",empty,before,\n"
                      "\"\",\"\",\"empty quoted\"\n"
                      "\"a,b\"\"c,d\",normal";

    g_row_count = 0;
    CSV_Parser *parser = CSV_ParserCreate();
    CSV_ParserSetDelimiter(parser, CSV_DELIMITER_COMMA);
    CSV_ParserSetEmptyLinePolicy(parser, CSV_EMPTY_LINE_SKIP);
    CSV_ParserSetRowCallback(parser, OnRowCallback, NULL);

    CSV_ParserFeed(parser, csv, strlen(csv));
    CSV_ParserFinish(parser);

    CSV_ParserDestroy(parser);
}

int main(void) {
    printf("========================================\n");
    printf("CSV 解析模块演示程序\n");
    printf("========================================\n");

    TestBasicParsing();
    TestQuotedFields();
    TestMixedNewlines();
    TestEmptyLines();
    TestStreaming();
    TestTypeConversion();
    TestTabDelimiter();
    TestEdgeCases();

    printf("\n========================================\n");
    printf("所有测试完成!\n");
    printf("========================================\n");

    return 0;
}
