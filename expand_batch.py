#!/usr/bin/env python3
"""
Solo-expand: 扩写 8251-8500 的 PROMPT.txt 到 500-700 字。
节奏 A/B/C/D/E 循环，禁止 AI 腔，自然口语化。
"""
import os, re, glob, random, sys

BASE = os.path.expanduser("~/Work/Private/work01/SoloCoder")
AI_SLOP = [
    "本项目旨在", "核心功能包括", "具体需求如下", "约束条件",
    "验收标准", "需要实现", "技术栈：", "功能点", "提供以下能力",
    "支持以下功能", "以下是", "主要功能如下",
]

# ── 节奏模板 ──────────────────────────────────────────
# 节奏 A: 场景叙事开头，流动叙述
# 节奏 B: 问题/痛点驱动，混合长短段
# 节奏 C: 直接但变化句式，破折号/列表混合
# 节奏 D: 使用流程叙述，编号步骤
# 节奏 E: 反问/对话式，紧凑段落

def rhythm_A(orig, topic_info, extra):
    """场景叙事开头，流动叙述无标签，模块融入叙述"""
    scene = topic_info.get('scene_a', '')
    middle = orig.strip()
    detail = extra.get('detail_a', '')
    ending = topic_info.get('end_a', '')
    parts = [p for p in [scene, middle, detail, ending] if p]
    return '\n\n'.join(parts)

def rhythm_B(orig, topic_info, extra):
    """问题/痛点驱动，混合长短段"""
    problem = topic_info.get('scene_b', '')
    middle = orig.strip()
    detail = extra.get('detail_b', '')
    ending = topic_info.get('end_b', '')
    parts = [p for p in [problem, middle, detail, ending] if p]
    return '\n\n'.join(parts)

def rhythm_C(orig, topic_info, extra):
    """直接但变化句式，破折号/列表混合"""
    opener = topic_info.get('scene_c', '')
    middle = orig.strip()
    detail = extra.get('detail_c', '')
    ending = topic_info.get('end_c', '')
    parts = [p for p in [opener, middle, detail, ending] if p]
    return '\n\n'.join(parts)

def rhythm_D(orig, topic_info, extra):
    """使用流程叙述，编号步骤"""
    opener = topic_info.get('scene_d', '')
    middle = orig.strip()
    detail = extra.get('detail_d', '')
    ending = topic_info.get('end_d', '')
    parts = [p for p in [opener, middle, detail, ending] if p]
    return '\n\n'.join(parts)

def rhythm_E(orig, topic_info, extra):
    """反问/对话式，紧凑段落"""
    opener = topic_info.get('scene_e', '')
    middle = orig.strip()
    detail = extra.get('detail_e', '')
    ending = topic_info.get('end_e', '')
    parts = [p for p in [opener, middle, detail, ending] if p]
    return '\n\n'.join(parts)

RHYTHM_FN = {'A': rhythm_A, 'B': rhythm_B, 'C': rhythm_C, 'D': rhythm_D, 'E': rhythm_E}

# ── 话题识别 + 扩写素材 ─────────────────────────────────

# 按 keyword → 扩写片段映射
TOPIC_DB = {}

def _add(keyword, scenes, extras, endings):
    TOPIC_DB[keyword] = {'scenes': scenes, 'extras': extras, 'endings': endings}

# ─── Rust: 编译器/二进制工具 ─────────────────────
_add("wasm-interpreter", {
    'A': '上个月我们组在做一个边缘计算网关，需要在不部署完整运行时的情况下执行轻量级 wasm 模块。调研了一圈发现现成的 wasm runtime 对嵌入场景太重了，于是决定自己搓一个最小可用的字节码解释器。',
    'B': '想在嵌入式设备上跑 wasm 却发现 wasmtime 编译出来好几 MB？其实很多时候我们只需要执行纯计算逻辑，不需要 WASI 和内存沙箱那些重量级特性。',
    'C': '团队的项目需要内嵌一个 wasm 执行引擎，但对体积有严格限制。所以目标很明确——手写一个精简的解释器，只覆盖最常用的指令子集。',
    'D': '先明确一下场景：后端服务需要动态加载 wasm 插件执行简单计算。不依赖任何外部 wasm runtime，纯手写解释器。整个流程是加载 wasm 二进制 → 解析函数段和代码段 → 解释执行。',
    'E': '为什么不能直接用 wasmtime？因为目标平台资源有限，而且只需要跑纯计算模块。一个能执行基本运算和控制流的 wasm 解释器就够了。',
}, {
    'detail_a': 'wasm 二进制格式按 section 组织，Preamble 魔数 \\\\0asm 后面跟版本号，然后是各种 section。你要解析的至少包括 Type Section（函数签名）、Function Section（函数索引）、Code Section（函数体字节码）。指令编码方面，i32.add 是 0x6a，i32.sub 是 0x6b，local.get 是 0x20，local.set 是 0x21，block 是 0x02，loop 是 0x03，if 是 0x04。操作数栈用 Vec<u32> 维护就行。遇到 0x00（unreachable）或者未识别的 opcode，返回自定义的 RuntimeError 而非 panic。',
    'detail_b': '解释器架构建议用栈式虚拟机。核心是一个操作数栈加上一个局部变量表，每个栈帧对应一个函数调用。控制流指令（block/loop/if）通过标签栈管理跳转目标，br 指令弹栈到对应层级的深度。栈深度追踪要做好，不然 br 指令的参数容易搞错。建议先把 wasm 二进制的 LEB128 编码搞定，整数字面量和 section 长度都用这个编码。',
    'detail_c': '几个要注意的坑：wasm 的 block 类型可以是空或者 valtype，解析时别遗漏；LEB128 有符号和无符号两种变体（i32.const 用有符号的）；线性内存可以用 Vec<u8> 模拟，i32.load/i32.store 暂时可以不实现但结构上预留好。',
    'detail_d': '建议分这几步实现：第一步解析 wasm 二进制的 header 和 section，至少处理 Type、Function、Export、Code 四个 section。第二步实现操作数栈和函数调用帧。第三步逐条解释执行指令，先覆盖 i32 算术和比较指令。第四步实现控制流，block/loop/if 各自的语义要区分清楚——block 的 br 跳到结尾，loop 的 br 跳回开头。',
    'detail_e': 'LEB128 解码是个容易踩的坑，尤其是有符号变体。wasm 规范里每个 section 的 id 和长度都得对上，碰到不认识的 section 跳过就行。i32 溢出的行为要跟 wasm spec 一致——都是按 32 位截断。',
}, {
    'end_a': '最终效果是能加载一个手写或 wasm-tools 生成的 .wasm 文件，调用导出的函数并拿到返回值。测试用例可以包括阶乘计算、斐波那契数列和嵌套循环。',
    'end_b': 'Rust 实现，cargo build 和 cargo test 都能通过。附带 3-5 个 .wasm 测试文件，覆盖算术运算、控制流和函数调用。不用接浏览器，纯命令行运行。',
    'end_c': '纯 Rust，cargo build 通过即可。测试用几个 wasm-tools 生成的 .wasm 文件，验证阶乘、斐波那契、累加和这类纯计算场景。',
    'end_d': 'cargo run -- test.wasm 执行指定模块。输出每个导出函数的执行结果和耗时。cargo test 覆盖解析器和解释器的核心路径。',
    'end_e': 'cargo build 能过就行。命令行执行 wasm 文件，打印函数返回值。附几个测试 wasm 模块。',
})

_add("elf-parser", {
    'A': '之前做内核开发的时候经常需要看 ELF 文件的段布局和符号表，readelf 虽然好用但我想自己写一个理解得更深。正好用 Rust 练手，从二进制格式解析开始。',
    'B': '每次碰到 ELF 格式都要查规范，不如自己写个解析器吃透它。重点是理解 header 结构和段表的组织方式。',
    'C': '项目中要分析 ELF 文件的依赖关系和符号导出情况，先从底层解析器做起。64 位 ELF 足够覆盖 Linux 上的大部分场景了。',
    'D': '目标是做一个简化版 readelf。先读取 ELF header 拿到 e_shoff 和 e_phoff，再按偏移量解析 section header table 和 program header table。',
    'E': '为什么不直接用 elf crate？因为想对 ELF 格式有更深的理解，而且项目里只需要读取基本信息，手写解析器比引入依赖更合适。',
}, {
    'detail_a': 'ELF64 的 Elf64_Ehdr 结构有 64 字节，前四个字节是魔数 0x7f454c46（\\x7fELF），第五个字节标识位数（2=64位），第六个标识字节序（1=小端，2=大端）。e_type 字段说明文件类型（2=可执行，3=共享对象）。e_entry 是程序入口点虚拟地址。Section headers 从 e_shoff 开始，每个 64 字节，e_shstrndx 指向段名字符串表的索引。Program headers 从 e_phoff 开始，描述内存段加载信息。',
    'detail_b': '解析流程建议这样：先读 64 字节的 Ehdr，校验魔数和位数。然后根据 e_shoff 和 e_shnum 读所有 Section Headers，每个 Shdr 包含段名索引、类型、虚拟地址、文件偏移和大小。注意段名要从 e_shstrndx 指向的字符串表里取。碰到 e_shnum 为 0 的边界情况要处理（真实值在第一个 Shdr 的 sh_size 里）。',
    'detail_c': '几个容易忽略的细节：ELF header 的 e_ident 数组前 4 字节是魔数，后面 8 字节是标识信息；section 的 sh_type 决定了如何解读 sh_link 和 sh_info；字符串表段（SHT_STRTAB）的内容是 null-terminated 字符串的拼接。',
    'detail_d': '建议分步实现：先定义 Elf64_Ehdr、Elf64_Shdr、Elf64_Phdr 的结构体（用 #[repr(C)] 保证内存布局）。然后实现 Ehdr 解析和校验。第三步解析 Shdr 数组并输出段名和类型。第四步解析 Phdr 数组。最后支持 --sections、--segments、--header 等子命令分别展示不同信息。',
    'detail_e': 'ELF 的 section header table 可能在文件末尾，e_shnum 可以是 0（表示超过 65535 个段），这些边界情况都遇到过。用 nom 或手写解析都行，重点是错误处理要到位——文件太短、魔数不对、偏移越界都要给清楚的提示。',
}, {
    'end_a': '最终 cargo run -- /bin/ls 能输出 header 摘要、段表列表和 program header 信息。解析错误的文件路径时给出友好提示而不是 panic。',
    'end_b': 'Rust，cargo build 通过。cargo run -- input-elf 解析输出，拿 /bin/ls 和系统库文件做测试。',
    'end_c': '纯 Rust 手写解析，cargo build 通过。至少能正确解析 /bin/ls 和其他常见 ELF 二进制的 header 和段表。',
    'end_d': 'cargo run -- /bin/ls 输出完整的 ELF 分析报告。错误场景（非 ELF 文件、截断文件）有合理的错误消息。cargo test 通过。',
    'end_e': 'cargo build 能过。命令行传入 ELF 文件路径，打印 header、sections、segments 信息。出错不要 panic。',
})

# 通用填充器：当没有精确匹配时使用
GENERIC_SCENES = {
    'A': ['上个季度在重构基础设施的时候发现需要对底层二进制格式有更深的理解，决定自己实现一套工具链来吃透这些格式。这不只是学习，也是为了在排查线上问题时不依赖外部工具。',
          '团队里每次排查二进制相关的问题都要装一堆工具，还不如自己写一个。正好趁这个机会把底层格式吃透，以后遇到问题心里有底。',
          '前阵子排查一个编译器的 bug，发现现有的调试工具不够用。一咬牙决定自己写一个，虽然只覆盖核心场景，但够用就行。',
          '之前给新同事培训的时候发现，大家对底层工具的原理都不太清楚。与其讲 PPT，不如带着一起写一个精简版的工具，理解更深。',
          '做项目的时候碰到一个奇怪的链接错误，排查了两天才发现是符号表解析有 bug。从那以后我就想自己实现一遍相关工具，免得再踩同样的坑。',
    ],
    'B': ['手头有个需求需要解析和处理特定的二进制格式，但现有的工具要么太重要么不够灵活。考虑到长期维护，自己实现一个更靠谱。',
          '每次处理这类问题都要翻文档，与其每次临时抱佛脚不如彻底搞定它。用 Rust 写一个可靠的实现，以后碰到类似问题直接拿过来用。',
          '遇到一个棘手的性能问题，归根到底是对底层实现理解不够深。与其黑盒调试，不如白盒实现一遍，搞清楚每一步在干什么。',
          '线上出了个数据损坏的事故，恢复的时候发现对格式理解不够。事后复盘决定自己实现一套工具，确保下次能快速定位和修复。',
          '第三方依赖的版本升级老是出兼容性问题，维护成本越来越高。核心功能自己掌握更放心，反正量也不大。',
    ],
    'C': ['有个长期项目需要深度处理某种数据格式，纯靠命令行工具拼凑效率太低。用 Python/Rust/Go 写一个专用库，集成到工具链里。',
          '之前用现成库完成任务，但出了问题完全不知道怎么调试。这次从头实现，每一步都清清楚楚。',
          '想对某种算法或数据结构有实战级别的理解，光看论文和博客不够。动手写一个完整实现，边界条件和性能瓶颈自然就暴露出来了。',
          '做技术选型的时候需要对比几种方案，但光看 benchmark 数据不够直观。自己实现一遍，对时间空间复杂度有第一手的体感。',
          '最近在整理团队的技术基建，发现有些关键模块还是黑盒。补齐这些基础能力，后面做业务的时候更踏实。',
    ],
    'D': ['先说下背景：项目需要处理和分析特定格式的数据。整体流程是解析输入 → 验证格式 → 处理转换 → 输出结果。每一步都要有错误处理和进度反馈。',
          '事情的起因是线上一个偶现的 bug，跟底层数据处理有关。排查的过程暴露出工具链的短板，所以决定自己实现一个更可靠的版本。',
          '任务分几步：先搞定基础数据结构和解析逻辑，然后实现核心算法，最后加命令行接口和测试。每一步都要确保正确性。',
          '项目需要一个可嵌入的模块来处理特定操作。不引入重量级依赖，纯手写实现。整体从数据模型开始，然后是算法逻辑，最后暴露接口。',
          '目标很明确：实现一个最小可用的版本，覆盖 80% 的常见场景。先跑通核心路径，再逐步补齐边界情况。',
    ],
    'E': ['为什么不用现成的库？因为想完全掌控实现细节，而且只需要核心功能，引入一整个依赖太重了。',
          '有没有一种方法能既理解原理又有可用的产出？自己从零实现一遍可能是最好的方式。',
          '一直对这类算法的实现细节好奇，但看别人的代码总隔着一层。自己写一遍，每一步为什么这么设计都门儿清。',
          '之前踩过一个坑：依赖库的某个行为跟文档不一致，排查半天才发现是 bug。自研的核心模块至少出了问题能自己修。',
          '团队在讨论技术方案的时候总觉得缺少底层的判断依据。纸上得来终觉浅，撸一个实际实现出来，讨论的时候更有底气。',
    ],
}

GENERIC_EXTRAS = {
    'A': '实现的时候注意错误处理要具体，不要笼统地返回一个 "invalid format"。不同阶段的错误要区分开：输入为空、格式不对、数据截断、值超出范围，分别给不同的错误消息。有条件的话加一些日志，方便后续排查问题。性能方面先不用刻意优化，正确性优先，等跑起来了再 profile 看瓶颈在哪。',
    'B': '几个容易忽略的点：边界输入（空数据、单元素数据、超大数据）都要测到；错误路径不要直接 panic，用 Result 返回具体原因；资源释放要干净，特别是涉及文件句柄和内存映射的场景。测试用例至少覆盖正常路径、空输入、格式错误、极端值四种场景。',
    'C': '关于数据格式——输入输出的编码要明确，字符串用 UTF-8，数字的字节序统一用小端。内存管理方面，大文件建议分块读取，不要一次性全部 load 到内存里。接口设计尽量简洁，对外暴露最少的 API。',
    'D': '测试策略建议分层：单元测试覆盖每个解析函数和核心算法，集成测试覆盖完整的处理流程。可以在项目根目录放一个 tests/ 文件夹，里面放测试数据和预期输出。性能测试用 go test -bench 或 cargo bench 都行，主要是看有没有明显的退化。',
    'E': '另外一个容易忽视的点：并发安全性。如果接口会被多个线程同时调用，内部的共享状态要用 Mutex 或 RwLock 保护。但不建议过度加锁——读写分离的场景用 RwLock 比 Mutex 性能好很多。实在拿不准就先写单线程版本，后面加并发的时候再改。',
}

GENERIC_ENDINGS = {
    'A': '整体实现保持简洁，不要过度设计。跑通之后如果发现性能有瓶颈，再针对性地优化。正确性第一，性能第二。',
    'B': '最终交付的是一个能编译通过的完整项目，附带基础测试和 README。不需要特别花哨，但核心功能要扎实。',
    'C': '编译能过，测试能跑。命令行接口保持简洁，输入输出格式明确。有问题的地方通过错误消息自解释。',
    'D': 'go build / cargo build / python 能通过。核心路径有测试覆盖。命令行用法写在注释或帮助信息里就行。',
    'E': '实现质量比代码量重要。能跑、能测、能维护。花里胡哨的功能以后再加，先把核心做对。',
}


def get_topic_key(dirname):
    """从目录名提取 topic keyword"""
    # 8251-rust-wasm-interpreter -> wasm-interpreter
    parts = dirname.split('-', 2)
    if len(parts) >= 3:
        return parts[2]
    return ""

def get_lang(dirname):
    """从目录名提取语言"""
    parts = dirname.split('-')
    if len(parts) >= 2:
        return parts[1]
    return ""

def get_scene_n(topic_key, rhythm, idx):
    """获取场景开头，优先用精确匹配，fallback 到通用"""
    if topic_key in TOPIC_DB:
        db = TOPIC_DB[topic_key]
        s = db['scenes'].get(rhythm.lower(), '')
        if s:
            return s
    # fallback 到通用
    pool = GENERIC_SCENES.get(rhythm, GENERIC_SCENES['A'])
    return pool[idx % len(pool)]

def get_extra(topic_key, rhythm, idx):
    """获取中间扩写段落"""
    if topic_key in TOPIC_DB:
        db = TOPIC_DB[topic_key]
        e = db['extras'].get(rhythm.lower(), '')
        if e:
            return e
    pool = GENERIC_EXTRAS
    return pool.get(rhythm, pool['A'])

def get_ending(topic_key, rhythm, idx):
    """获取结尾段落"""
    if topic_key in TOPIC_DB:
        db = TOPIC_DB[topic_key]
        e = db['endings'].get(rhythm.lower(), '')
        if e:
            return e
    pool = GENERIC_ENDINGS
    return pool.get(rhythm, pool['A'])


def char_count(text):
    """计算字符数（去掉空行后的实际内容字数）"""
    # 去掉纯空白的行，然后计算所有非空白字符
    return len(text.replace('\n', '').replace(' ', ''))


def check_ai_slop(text):
    """检查是否包含 AI 腔用语"""
    hits = []
    for phrase in AI_SLOP:
        if phrase in text:
            hits.append(phrase)
    return hits


def expand_prompt(dirpath, dirname, idx):
    """扩写单个 PROMPT.txt"""
    prompt_path = os.path.join(dirpath, "PROMPT.txt")
    if not os.path.exists(prompt_path):
        return None

    with open(prompt_path, 'r', encoding='utf-8') as f:
        orig = f.read().strip()

    orig_chars = char_count(orig)
    if orig_chars >= 480:
        # 已经够长了，只做微调
        return orig, orig_chars, orig_chars, "skip"

    # 确定节奏
    rhythms = ['A', 'B', 'C', 'D', 'E']
    rhythm = rhythms[idx % 5]

    topic_key = get_topic_key(dirname)
    lang = get_lang(dirname)

    # 构建扩写
    scene = get_scene_n(topic_key, rhythm, idx)
    extra = get_extra(topic_key, rhythm, idx)
    ending = get_ending(topic_key, rhythm, idx)

    # 根据节奏组合
    if rhythm == 'A':
        # 场景叙事，流动叙述
        parts = []
        if scene:
            parts.append(scene)
        parts.append(orig)
        if extra and char_count('\n'.join(parts)) < 550:
            parts.append(extra)
        if ending and char_count('\n'.join(parts)) < 620:
            parts.append(ending)
        result = '\n\n'.join(parts)

    elif rhythm == 'B':
        # 问题驱动，混合长短段
        parts = []
        if scene:
            parts.append(scene)
        parts.append(orig)
        if extra and char_count('\n'.join(parts)) < 550:
            parts.append(extra)
        if ending and char_count('\n'.join(parts)) < 620:
            parts.append(ending)
        result = '\n\n'.join(parts)

    elif rhythm == 'C':
        # 直接变化句式
        parts = []
        parts.append(orig)
        if scene:
            # 把场景插在中间作为补充段落
            mid_parts = [orig, scene]
        else:
            mid_parts = [orig]
        if extra and char_count('\n'.join(mid_parts)) < 550:
            mid_parts.append(extra)
        if ending and char_count('\n'.join(mid_parts)) < 620:
            mid_parts.append(ending)
        result = '\n\n'.join(mid_parts)

    elif rhythm == 'D':
        # 流程叙述
        parts = []
        if scene:
            parts.append(scene)
        parts.append(orig)
        if extra and char_count('\n'.join(parts)) < 550:
            parts.append(extra)
        if ending and char_count('\n'.join(parts)) < 620:
            parts.append(ending)
        result = '\n\n'.join(parts)

    else:  # E
        parts = []
        if scene:
            parts.append(scene)
        parts.append(orig)
        if extra and char_count('\n'.join(parts)) < 550:
            parts.append(extra)
        if ending and char_count('\n'.join(parts)) < 620:
            parts.append(ending)
        result = '\n\n'.join(parts)

    # 确保 build 指令保留
    build_patterns = ['cargo build', 'go build', 'python', 'python3', 'cargo test',
                      'cargo run', 'go test', 'PORT']
    orig_has_build = any(p in orig for p in build_patterns)
    result_has_build = any(p in result for p in build_patterns)
    if orig_has_build and not result_has_build:
        result += '\n\n' + orig.split('\n')[-1] if orig.split('\n')[-1].strip() else ''

    new_chars = char_count(result)
    return result, orig_chars, new_chars, rhythm


def main():
    # 收集所有 8251-8500 项目
    projects = []
    for i in range(8251, 8501):
        pattern = os.path.join(BASE, f"{i}-*")
        matches = glob.glob(pattern)
        if matches:
            projects.append((i, matches[0]))

    print(f"找到 {len(projects)} 个项目")

    stats = {
        'total': 0, 'expanded': 0, 'skipped': 0,
        'chars_min': 99999, 'chars_max': 0,
        'rhythms': {'A': 0, 'B': 0, 'C': 0, 'D': 0, 'E': 0},
        'ai_slop_hits': 0,
        'errors': [],
    }

    for idx, (num, dirpath) in enumerate(projects):
        dirname = os.path.basename(dirpath)
        prompt_path = os.path.join(dirpath, "PROMPT.txt")

        if not os.path.exists(prompt_path):
            stats['errors'].append(f"{num}: PROMPT.txt 不存在")
            continue

        result = expand_prompt(dirpath, dirname, idx)
        if result is None:
            stats['errors'].append(f"{num}: 扩写失败")
            continue

        expanded, orig_chars, new_chars, rhythm = result
        stats['total'] += 1

        if rhythm == 'skip':
            stats['skipped'] += 1
            print(f"  [{num}] 跳过（{orig_chars}字 ≥ 480）")
            continue

        # AI 腔检查
        hits = check_ai_slop(expanded)
        if hits:
            stats['ai_slop_hits'] += 1
            stats['errors'].append(f"{num}: AI腔命中 {hits}")

        # 写入
        with open(prompt_path, 'w', encoding='utf-8') as f:
            f.write(expanded + '\n')

        stats['expanded'] += 1
        stats['rhythms'][rhythm] += 1
        stats['chars_min'] = min(stats['chars_min'], new_chars)
        stats['chars_max'] = max(stats['chars_max'], new_chars)

        if (idx + 1) % 50 == 0:
            print(f"  已处理 {idx + 1}/{len(projects)}")

    # 收集最终字符数统计
    all_chars = []
    for idx, (num, dirpath) in enumerate(projects):
        prompt_path = os.path.join(dirpath, "PROMPT.txt")
        if os.path.exists(prompt_path):
            with open(prompt_path, 'r', encoding='utf-8') as f:
                all_chars.append(char_count(f.read()))

    avg = sum(all_chars) / len(all_chars) if all_chars else 0
    sorted_chars = sorted(all_chars)
    median = sorted_chars[len(sorted_chars) // 2] if sorted_chars else 0

    print(f"\n📊 扩写统计")
    print(f"范围：8251-8500（{stats['total']} 个项目）")
    print(f"扩写：{stats['expanded']} 个 | 跳过：{stats['skipped']} 个")
    print(f"字数：min={min(all_chars)} max={max(all_chars)} avg={avg:.0f} median={median}")
    print(f"节奏分布：A={stats['rhythms']['A']} B={stats['rhythms']['B']} C={stats['rhythms']['C']} D={stats['rhythms']['D']} E={stats['rhythms']['E']}")
    print(f"AI腔命中：{stats['ai_slop_hits']}")
    if stats['errors']:
        print(f"错误/警告（{len(stats['errors'])}）：")
        for e in stats['errors'][:20]:
            print(f"  {e}")


if __name__ == '__main__':
    main()
