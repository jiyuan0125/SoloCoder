#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include "dict.h"

#define MAX_INPUT 256

typedef struct {
    const char *word;
    int weight;
    const char *pinyin;
} SampleWord;

static SampleWord sample_words[] = {
    {"iphone", 10000, "iphone"},
    {"ipad", 8000, "ipad"},
    {"ipod", 5000, "ipod"},
    {"apple", 9000, "apple"},
    {"app store", 7000, "appstore"},
    {"android", 8500, "android"},
    {"手机", 9500, "shouji"},
    {"手机壳", 6000, "shoujike"},
    {"手机充电器", 5500, "shoujichongdianqi"},
    {"北京", 9000, "beijing"},
    {"北京天气", 7500, "beijingtianqi"},
    {"北京天安门", 7000, "beijingtiananmen"},
    {"北京大学", 8000, "beijingdaxue"},
    {"北京医院", 6500, "beijingyiyuan"},
    {"上海", 8800, "shanghai"},
    {"上海天气", 7200, "shanghaitianqi"},
    {"上海外滩", 6800, "shanghaiwaitan"},
    {"淘宝", 9200, "taobao"},
    {"天猫", 8500, "tianmao"},
    {"京东", 8800, "jingdong"},
    {"拼多多", 8200, "pingduoduo"},
    {"微信", 9500, "weixin"},
    {"微信支付", 7500, "weixinzhifu"},
    {"王者荣耀", 9000, "wangzherongyao"},
    {"和平精英", 8500, "hepingjingying"},
    {"原神", 8800, "yuanshen"},
    {"英雄联盟", 9200, "yingxionglianmeng"},
    {"电视剧", 7000, "dianshiju"},
    {"电影", 7500, "dianying"},
    {"音乐", 6800, "yinyue"},
    {"小说", 6500, "xiaoshuo"},
    {"新闻", 7200, "xinwen"},
    {"天气预报", 8000, "tianqiyubao"},
    {"地图", 6000, "ditu"},
    {"翻译", 5500, "fanyi"},
    {"计算器", 5000, "jisuanqi"},
    {"日历", 4800, "rili"},
    {"闹钟", 4500, "naozhong"},
    {"相机", 6200, "xiangji"},
    {"相册", 5800, "xiangce"},
    {"浏览器", 6500, "liulanqi"},
    {"邮件", 5200, "youjian"},
    {"设置", 4500, "shezhi"},
    {"文件管理", 4800, "wenjianguanli"},
    {"应用商店", 5500, "yingyongshangdian"},
    {"游戏中心", 6000, "youxizhongxin"},
    {"视频", 7000, "shipin"},
    {"短视频", 7500, "duanshipin"},
    {"直播", 6800, "zhibo"},
    {NULL, 0, NULL}
};

static void init_dictionary(Dictionary *dict) {
    printf("正在加载词典...\n");
    
    for (int i = 0; sample_words[i].word != NULL; i++) {
        dict_add_word(dict, sample_words[i].word, sample_words[i].weight);
        if (sample_words[i].pinyin != NULL) {
            dict_add_pinyin_mapping(dict, sample_words[i].word, sample_words[i].pinyin);
        }
    }
    
    printf("词典加载完成，共加载 %d 个词条\n", (int)(sizeof(sample_words)/sizeof(SampleWord) - 1));
}

static void print_suggestions(Suggestion *results, int count) {
    if (count == 0) {
        printf("  没有找到匹配的建议\n");
        return;
    }
    
    for (int i = 0; i < count; i++) {
        if (results[i].is_fuzzy) {
            printf("  %d. %s (权重: %d) [模糊匹配]\n", 
                   i + 1, results[i].word, results[i].weight);
        } else {
            printf("  %d. %s (权重: %d)\n", 
                   i + 1, results[i].word, results[i].weight);
        }
    }
}

static void test_performance(Dictionary *dict) {
    printf("\n--- 性能测试 ---\n");
    
    const char *test_queries[] = {
        "i", "ip", "iph", "iphone",
        "北", "北京", "北京天",
        "shouji", "bj",
        NULL
    };
    
    clock_t start, end;
    double total_time = 0;
    int query_count = 0;
    
    for (int i = 0; test_queries[i] != NULL; i++) {
        Suggestion results[MAX_SUGGESTIONS];
        
        start = clock();
        int count = dict_query(dict, test_queries[i], results, MAX_SUGGESTIONS);
        end = clock();
        
        double time_ms = ((double)(end - start)) / CLOCKS_PER_SEC * 1000;
        total_time += time_ms;
        query_count++;
        
        printf("  查询 \"%s\": 找到 %d 个结果, 耗时 %.3f 毫秒\n", 
               test_queries[i], count, time_ms);
    }
    
    printf("\n  平均查询时间: %.3f 毫秒\n", total_time / query_count);
    printf("  (性能要求: < 10 毫秒)\n");
}

static void demo_update_operations(Dictionary *dict) {
    printf("\n--- 词典更新操作演示 ---\n");
    
    printf("\n1. 添加新词 \"测试功能\" (权重: 8000)\n");
    if (dict_add_word(dict, "测试功能", 8000)) {
        printf("   添加成功\n");
    } else {
        printf("   添加失败\n");
    }
    
    printf("\n2. 查询 \"测试\":\n");
    Suggestion results[MAX_SUGGESTIONS];
    int count = dict_query(dict, "测试", results, MAX_SUGGESTIONS);
    print_suggestions(results, count);
    
    printf("\n3. 更新 \"测试功能\" 的权重为 9000\n");
    if (dict_update_weight(dict, "测试功能", 9000)) {
        printf("   更新成功\n");
    } else {
        printf("   更新失败（词不存在）\n");
    }
    
    printf("\n4. 再次查询 \"测试\":\n");
    count = dict_query(dict, "测试", results, MAX_SUGGESTIONS);
    print_suggestions(results, count);
    
    printf("\n5. 删除 \"测试功能\"\n");
    if (dict_remove_word(dict, "测试功能")) {
        printf("   删除成功\n");
    } else {
        printf("   删除失败（词不存在）\n");
    }
    
    printf("\n6. 最后查询 \"测试\":\n");
    count = dict_query(dict, "测试", results, MAX_SUGGESTIONS);
    print_suggestions(results, count);
}

static void demo_special_cases(Dictionary *dict) {
    printf("\n--- 特殊情况演示 ---\n");
    
    Suggestion results[MAX_SUGGESTIONS];
    int count;
    
    printf("\n1. 空前缀（热门推荐）:\n");
    count = dict_query(dict, "", results, MAX_SUGGESTIONS);
    print_suggestions(results, count);
    
    printf("\n2. 模糊匹配示例 - 输入 \"iphne\" (应为 \"iphone\"):\n");
    count = dict_query(dict, "iphne", results, MAX_SUGGESTIONS);
    print_suggestions(results, count);
    
    printf("\n3. 中文子串匹配 - 输入 \"北京天\" (匹配 \"北京天气\", \"北京天安门\"):\n");
    count = dict_query(dict, "北京天", results, MAX_SUGGESTIONS);
    print_suggestions(results, count);
    
    printf("\n4. 拼音搜索 - 输入 \"shouji\":\n");
    count = dict_query(dict, "shouji", results, MAX_SUGGESTIONS);
    print_suggestions(results, count);
    
    printf("\n5. 英文前缀匹配 - 输入 \"ip\":\n");
    count = dict_query(dict, "ip", results, MAX_SUGGESTIONS);
    print_suggestions(results, count);
}

static void interactive_mode(Dictionary *dict) {
    printf("\n--- 交互模式 ---\n");
    printf("输入前缀查询建议词，输入 :quit 退出，:help 查看帮助\n\n");
    
    char input[MAX_INPUT];
    Suggestion results[MAX_SUGGESTIONS];
    
    while (1) {
        printf("> ");
        if (fgets(input, sizeof(input), stdin) == NULL) {
            break;
        }
        
        input[strcspn(input, "\n")] = '\0';
        
        if (strcmp(input, ":quit") == 0 || strcmp(input, ":q") == 0) {
            break;
        }
        
        if (strcmp(input, ":help") == 0 || strcmp(input, ":h") == 0) {
            printf("  帮助:\n");
            printf("    输入任意文字进行前缀搜索\n");
            printf("    空输入显示热门推荐\n");
            printf("    :quit/:q - 退出程序\n");
            printf("    :help/:h - 显示此帮助\n");
            printf("    :test - 运行性能测试\n");
            printf("    :demo - 运行更新操作演示\n");
            printf("    :special - 运行特殊情况演示\n");
            continue;
        }
        
        if (strcmp(input, ":test") == 0) {
            test_performance(dict);
            continue;
        }
        
        if (strcmp(input, ":demo") == 0) {
            demo_update_operations(dict);
            continue;
        }
        
        if (strcmp(input, ":special") == 0) {
            demo_special_cases(dict);
            continue;
        }
        
        clock_t start = clock();
        int count = dict_query(dict, input, results, MAX_SUGGESTIONS);
        clock_t end = clock();
        
        double time_ms = ((double)(end - start)) / CLOCKS_PER_SEC * 1000;
        
        if (input[0] == '\0') {
            printf("  [热门推荐] (耗时 %.3f 毫秒):\n", time_ms);
        } else {
            printf("  搜索 \"%s\" (耗时 %.3f 毫秒):\n", input, time_ms);
        }
        print_suggestions(results, count);
        printf("\n");
    }
}

int main(void) {
    printf("========================================\n");
    printf("    搜索引擎输入联想模块演示\n");
    printf("========================================\n\n");
    
    Dictionary *dict = dict_create();
    if (!dict) {
        fprintf(stderr, "创建词典失败！\n");
        return 1;
    }
    
    init_dictionary(dict);
    
    printf("\n功能说明:\n");
    printf("  1. 前缀匹配: 输入前缀快速查找匹配词\n");
    printf("  2. 权重排序: 结果按搜索热度从高到低排列\n");
    printf("  3. 动态更新: 支持添加、删除、修改权重\n");
    printf("  4. 模糊匹配: 允许1-2个字符差异（如 \"iphne\" -> \"iphone\"）\n");
    printf("  5. 子串匹配: 中文支持中间词匹配（如 \"北京天\" -> \"北京天气\"）\n");
    printf("  6. 拼音搜索: 输入拼音查找中文词（如 \"shouji\" -> \"手机\"）\n");
    printf("  7. 热门推荐: 空输入返回权重最高的词\n");
    
    interactive_mode(dict);
    
    printf("\n正在清理资源...\n");
    dict_destroy(dict);
    printf("程序已退出。\n");
    
    return 0;
}
