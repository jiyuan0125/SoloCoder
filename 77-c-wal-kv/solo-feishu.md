# Solo Coder 填表数据

## 77-c-wal-kv — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | .335769888099319:6216b3bf58bbd261d142ad8ab631fa3f_69f62ff157c8d33337f23b51.69f62ff357c8d33337f23b54.69f62ff1d962c6cda068b6ab:Trae CN.T(2026/5/3 01:10:11) |
| 第一轮Session ID | .335769888099319:6216b3bf58bbd261d142ad8ab631fa3f_69f62ff157c8d33337f23b51.69f62ff357c8d33337f23b54.69f62ff1d962c6cda068b6ab:Trae CN.T(2026/5/3 01:10:11) |
| 轮次 | 1 |
| User Prompt | 用 C 写一个持久化 KV 存储引擎。支持 CRUD 操作，数据先写 WAL 再更新内存索引。 存储架构： 1. 所有写操作（put/delete）先追加到 WAL 文件（write-ahead log），WAL 格式：4 字节 CRC + 4 字节记录长度 + 1 字节类型(0=PUT,1=DELETE) + 2 字节 key 长度 + key + 4 字节 value 长度 + value 2. 内存中维护一个 memtable（跳表实现），存最近写入的 key-value 3. memtable 大小超过 4MB 时 flush 到磁盘成为 SSTable（排序的 key-value 文件，格式：4 字节 key 长度 + key + 4 字节 value 长度 + value） 读操作： 4. get(key)：先查 memtable，查不到就查 SSTable（从最新的开始往前找，找到就返回） 5. scan(prefix)：返回所有匹配前缀的 key-value，按 key 字典序排列 后台任务： 6. 定期执行 compaction：把多个 SSTable 合并成一个，新值覆盖旧值，遇到 DELETE 记录就删除 7. compaction 过程不能阻塞正常的读写操作 崩溃恢复： 8. 启动时回放 WAL 重建 memtable 9. WAL 最后一条记录如果 CRC 校验失败或长度字段指向文件末尾之外，丢弃该条（模拟写到一半崩溃），不能因为一条坏记录导致整个恢复失败 代码分 wal.c、memtable.c、sstable.c、compaction.c、db.c 几个文件，gcc 编译能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：compaction 不处理 DELETE 记录——PROMPT 要求"遇到 DELETE 记录就删除"，但 sstable_create_from_memtable 把 tombstone 当普通条目写入 SSTable（value_len=0），SSTable 格式本身也没有 deleted 标记字段；sstable_get 硬编码 deleted=false，已删除的 key flush 到 SSTable 后会被当作空值的有效 key 返回；compaction 的 load_sstable_to_memtable 用 skiplist_put 回写，丢失 deleted 标记，合并后 tombstone 变成普通空值条目。实际复现场景：put 一个 key → delete 它 → 触发 memtable flush（数据超 4MB 或 db_close）→ 重新打开 db → get 该 key 返回非 NULL（应该返回 NULL）。过程不满意：compaction.c 有 3 个 unused static function（merge_entry_create、merge_entry_free、key_compare）和 1 个 unused variable（ret）未清理 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 77-c-wal-kv |

---

## 77-c-wal-kv — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | .335769888099319:4d91c1070feece780e1b016da65c106d_69f62ff157c8d33337f23b51.69f67d2157c8d33337f23c61.69f67d20d962c6cda068b6ac:Trae CN.T(2026/5/3 06:39:29) |
| 第一轮Session ID | .335769888099319:6216b3bf58bbd261d142ad8ab631fa3f_69f62ff157c8d33337f23b51.69f62ff357c8d33337f23b54.69f62ff1d962c6cda068b6ab:Trae CN.T(2026/5/3 01:10:11) |
| 轮次 | 2 |
| User Prompt | compaction 不处理 DELETE 记录。PROMPT 要求"遇到 DELETE 记录就删除"，当前 sstable_create_from_memtable 把 tombstone 当普通条目写入（value_len=0），SSTable 格式没有 deleted 标记；sstable_get 硬编码 deleted=false，已删除 key flush 后被当作空值有效 key 返回；compaction 的 load_sstable_to_memtable 用 skiplist_put 回写丢失 deleted 标记。复现场景：put → delete → flush → reopen → get 返回非 NULL（应返回 NULL）。另外 compaction.c 有 3 个 unused static function 和 1 个 unused variable 未清理 |
| 任务类型 | Bug修复 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 |  |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 77-c-wal-kv |

---
