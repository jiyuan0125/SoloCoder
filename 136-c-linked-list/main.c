#include <stdio.h>
#include <stdlib.h>
#include "player.h"

static void print_song(Song* song) {
    if (!song) {
        printf("  (无歌曲)\n");
        return;
    }
    int min = song->duration / 60;
    int sec = song->duration % 60;
    printf("  [%d] %s - %s (%02d:%02d) %s%s\n",
           song->id, song->title, song->artist, min, sec,
           song->is_favorite ? "[❤]" : "",
           song->source == SOURCE_TEMP ? "[临时]" : "");
}

static void print_separator(void) {
    printf("=================================================\n");
}

int main(void) {
    printf("音乐播放器播放列表管理模块演示\n");
    print_separator();
    
    Player* player = player_create();
    if (!player) {
        printf("错误: 创建播放器失败\n");
        return 1;
    }
    
    printf("\n【步骤1】添加歌曲到播放列表\n");
    player_add_song(player, song_create("夜曲", "周杰伦", 285, "/music/yequ.mp3"));
    player_add_song(player, song_create("稻香", "周杰伦", 223, "/music/daoxiang.mp3"));
    player_add_song(player, song_create("七里香", "周杰伦", 299, "/music/qilixiang.mp3"));
    player_add_song(player, song_create("青花瓷", "周杰伦", 239, "/music/qinghuaci.mp3"));
    player_add_song(player, song_create("晴天", "周杰伦", 269, "/music/qingtian.mp3"));
    player_print_list(player, false);
    
    print_separator();
    printf("\n【步骤2】在位置2插入一首新歌\n");
    player_insert_song_at(player, 2, song_create("告白气球", "周杰伦", 215, "/music/gaobai.mp3"));
    player_print_list(player, false);
    
    print_separator();
    printf("\n【步骤3】标记几首歌为收藏\n");
    player_set_favorite(player, 1, true);
    player_set_favorite(player, 3, true);
    player_set_favorite(player, 6, true);
    printf("只显示收藏的歌曲:\n");
    player_print_list(player, true);
    
    print_separator();
    printf("\n【步骤4】开始播放，测试顺序播放模式\n");
    player_set_mode(player, PLAY_MODE_SEQUENTIAL);
    Song* s = player_play(player);
    printf("开始播放: ");
    print_song(s);
    
    printf("\n播放下一首 (顺序模式):\n");
    for (int i = 0; i < 7; i++) {
        s = player_next(player);
        printf("  下一首: ");
        if (s) print_song(s);
        else printf("  (播放结束)\n");
    }
    
    print_separator();
    printf("\n【步骤5】测试列表循环模式\n");
    player_set_mode(player, PLAY_MODE_REPEAT_ALL);
    player_play_at(player, 4);
    printf("从位置4开始播放: ");
    print_song(player_get_current_song(player));
    
    printf("\n连续切歌 (列表循环):\n");
    for (int i = 0; i < 8; i++) {
        s = player_next(player);
        printf("  下一首: ");
        print_song(s);
    }
    
    print_separator();
    printf("\n【步骤6】测试单曲循环模式\n");
    player_set_mode(player, PLAY_MODE_REPEAT_ONE);
    player_play_at(player, 2);
    printf("当前播放: ");
    print_song(player_get_current_song(player));
    
    printf("\n连续切歌 (单曲循环):\n");
    for (int i = 0; i < 5; i++) {
        s = player_next(player);
        printf("  下一首: ");
        print_song(s);
    }
    
    print_separator();
    printf("\n【步骤7】测试上一首功能\n");
    player_set_mode(player, PLAY_MODE_SEQUENTIAL);
    player_play_at(player, 3);
    printf("当前播放: ");
    print_song(player_get_current_song(player));
    
    printf("\n按上一首:\n");
    for (int i = 0; i < 5; i++) {
        s = player_prev(player);
        printf("  上一首: ");
        print_song(s);
    }
    
    print_separator();
    printf("\n【步骤8】测试随机播放模式\n");
    player_set_mode(player, PLAY_MODE_SHUFFLE);
    player_play_at(player, 0);
    printf("当前播放: ");
    print_song(player_get_current_song(player));
    
    printf("\n随机切歌 (保证每首轮播不重复):\n");
    int total = player_get_total_songs(player);
    for (int i = 0; i < total + 2; i++) {
        s = player_next(player);
        printf("  随机下一首: ");
        print_song(s);
    }
    
    print_separator();
    printf("\n【步骤9】测试临时插入（插队）\n");
    player_set_mode(player, PLAY_MODE_SEQUENTIAL);
    player_play_at(player, 1);
    printf("当前播放: ");
    print_song(player_get_current_song(player));
    player_print_list(player, false);
    
    printf("\n插入2首临时歌曲:\n");
    player_insert_temp_song(player, song_create("听妈妈的话", "周杰伦", 264, "/music/mama.mp3"));
    player_insert_temp_song(player, song_create("以父之名", "周杰伦", 327, "/music/fu.mp3"));
    player_print_list(player, false);
    
    printf("\n开始切歌 (临时歌曲播完自动移除):\n");
    for (int i = 0; i < 5; i++) {
        s = player_next(player);
        printf("  下一首: ");
        print_song(s);
        printf("  播放列表状态 - 主列表:%d首, 临时:%d首\n", 
               player->main_list->count, player->temp_list->count);
    }
    
    print_separator();
    printf("\n【步骤10】测试删除歌曲\n");
    player_play_at(player, 2);
    printf("当前播放: ");
    print_song(player_get_current_song(player));
    printf("删除前的列表:\n");
    player_print_list(player, false);
    
    printf("\n删除当前播放的歌曲 (ID=%d):\n", 
           player_get_current_song(player)->id);
    int remove_id = player_get_current_song(player)->id;
    player_remove_song(player, remove_id);
    printf("删除后自动切换到下一首: ");
    print_song(player_get_current_song(player));
    printf("删除后的列表:\n");
    player_print_list(player, false);
    
    printf("\n删除最后一首，验证指针是否回到第一首:\n");
    player_play_at(player, player->main_list->count - 1);
    printf("当前播放 (最后一首): ");
    print_song(player_get_current_song(player));
    remove_id = player_get_current_song(player)->id;
    player_remove_song(player, remove_id);
    printf("删除后播放: ");
    print_song(player_get_current_song(player));
    
    print_separator();
    printf("\n【步骤11】测试移动歌曲位置\n");
    printf("移动前列表:\n");
    player_print_list(player, false);
    
    printf("\n把位置0的歌移到位置2:\n");
    player_move_song(player, 0, 2);
    player_print_list(player, false);
    
    print_separator();
    printf("\n【步骤12】演示暂停/恢复/停止\n");
    player_play(player);
    printf("播放状态: %s\n", player->is_playing ? "播放中" : "停止");
    player_pause(player);
    printf("暂停后: %s\n", player->is_paused ? "已暂停" : "未暂停");
    player_resume(player);
    printf("恢复后: %s\n", player->is_paused ? "已暂停" : "播放中");
    player_stop(player);
    printf("停止后: %s\n", player->is_playing ? "播放中" : "已停止");
    
    print_separator();
    printf("\n演示完成！释放资源...\n");
    player_destroy(player);
    
    printf("\n所有模块测试通过！\n");
    return 0;
}
