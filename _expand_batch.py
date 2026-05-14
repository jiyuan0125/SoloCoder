#!/usr/bin/env python3
"""
Solo-expand batch script for projects 9001-9250.
Expands short PROMPT.txt files to 500-700 chars (Chinese chars count as 1).
Rhythm pattern: A/B/C/D/E cycling, adjacent 5 must differ.
"""
import os, sys, re, json, subprocess

BASE = os.path.expanduser("~/Work/Private/work01/SoloCoder")

RHYTHMS = ['A', 'B', 'C', 'D', 'E']

def char_count(text):
    return len(text)

def get_projects():
    """Get all projects in 9001-9250 range with their PROMPT.txt."""
    projects = []
    for d in sorted(os.listdir(BASE)):
        if not re.match(r'^9[0-2]\d\d-', d):
            continue
        num = int(d.split('-')[0])
        if num < 9001 or num > 9250:
            continue
        ppath = os.path.join(BASE, d, 'PROMPT.txt')
        if os.path.exists(ppath):
            with open(ppath, 'r') as f:
                content = f.read().strip()
            projects.append((num, d, ppath, content))
    return projects

def detect_lang(dirname):
    """Detect language from directory name."""
    if '-go-' in dirname or dirname.endswith('-go') or dirname.split('-')[1] == 'go':
        return 'go'
    if '-python-' in dirname or dirname.split('-')[1] == 'python':
        return 'python'
    if '-java-' in dirname:
        return 'java'
    if '-rust-' in dirname:
        return 'rust'
    # Default for this range
    parts = dirname.split('-')
    if len(parts) >= 2:
        return parts[1]
    return 'python'

def detect_topic(dirname):
    """Detect topic from directory name."""
    parts = dirname.split('-', 1)
    if len(parts) >= 2:
        return parts[1]
    return dirname

def detect_framework(lang, content):
    """Detect framework from content."""
    c = content.lower()
    if lang == 'python':
        if 'fastapi' in c: return 'FastAPI'
        if 'flask' in c: return 'Flask'
        if 'aiohttp' in c: return 'aiohttp'
        if 'django' in c: return 'Django'
        return '标准库 http.server / asyncio'
    if lang == 'go':
        if 'gin' in c: return 'Gin'
        if 'fiber' in c: return 'Fiber'
        return 'Go 标准库 net/http'
    if lang == 'java':
        if 'spring' in c: return 'Spring Boot'
        if 'netty' in c: return 'Netty'
        return 'Java 标准库'
    if lang == 'rust':
        if 'actix' in c: return 'actix-web'
        if 'axum' in c: return 'axum'
        return 'tokio'
    return '标准库'

def check_ai_slop(text):
    """Check for AI-slop patterns."""
    blacklist = [
        '本项目旨在', '核心功能包括', '具体需求如下', '约束条件',
        '验收标准', '需要实现', '技术栈：', '功能点',
        '提供以下能力', '支持以下功能', '至少包含'
    ]
    hits = []
    for b in blacklist:
        if b in text:
            hits.append(b)
    return hits

# Template-free expansion functions for each rhythm type
# Each adds meaningful technical details from different angles

def expand_rhythm_A(dirname, lang, topic, content, num):
    """Rhythm A: 场景叙事开头 → 流动叙述 → 模块融入叙述"""
    # Already has scene narrative? Keep and extend it
    # Add: interface details, data model, error handling woven into narrative
    pass

def expand_rhythm_B(dirname, lang, topic, content, num):
    """Rhythm B: 问题/痛点驱动 → 混合长短段有编号 → 架构描述结尾"""
    pass

def expand_rhythm_C(dirname, lang, topic, content, num):
    """Rhythm C: 直接但变化句式 → 破折号/列表混合 → 技术约束结尾"""
    pass

def expand_rhythm_D(dirname, lang, topic, content, num):
    """Rhythm D: 使用流程叙述 → 编号步骤 → 简洁模块列表"""
    pass

def expand_rhythm_E(dirname, lang, topic, content, num):
    """Rhythm E: 反问/对话式 → 紧凑段落 → 开放式结尾"""
    pass


# Topic-specific detail banks for security tools
SECURITY_DETAILS = {
    'sqli': {
        'interfaces': 'POST /api/detect 接收 {"input": "..."} 返回 {"injection": true, "type": "union_based", "risk": "high", "matches": [...]}',
        'patterns': ['union select', 'or 1=1', "' or '", 'sleep(', 'benchmark(', '; drop table', 'information_schema'],
        'errors': '400 输入为空, 413 输入超长(>64KB)',
    },
    'xss': {
        'interfaces': 'POST /api/filter 接收 {"html": "..."} 返回净化后的 HTML 和被过滤的片段列表',
        'patterns': ['<script>', 'javascript:', 'onerror=', 'onload=', 'data:text/html'],
        'errors': '400 输入为空',
    },
    'csrf': {
        'interfaces': 'POST /api/token 生成 CSRF token, POST /api/verify 验证 token 有效性',
        'errors': '401 token 无效/过期, 410 token 已使用(一次性模式)',
    },
    'jwt': {
        'interfaces': 'POST /api/validate 验证 JWT 签名和 claims, POST /api/decode 解码不验证签名',
        'errors': '401 签名无效, 401 token 过期, 400 格式错误',
    },
    'oauth2': {
        'interfaces': 'GET /authorize 授权端点, POST /token 换取 token, POST /token/introspect 验证',
        'errors': '401 client 认证失败, 400 redirect_uri 不匹配, 403 scope 不足',
    },
    'crypto': {
        'interfaces': 'POST /api/encrypt 加密, POST /api/decrypt 解密, GET /api/keygen 生成密钥',
        'errors': '400 参数缺失, 401 密钥无效, 500 加密失败',
    },
    'scan': {
        'interfaces': 'POST /api/scan 启动扫描, GET /api/scan/{id}/status 查询进度, GET /api/scan/{id}/result 获取结果',
        'errors': '404 扫描任务不存在, 409 扫描进行中',
    },
}

# Topic-specific detail banks for audio/video/multimedia
AV_DETAILS = {
    'wav': {
        'chunks': 'RIFF header, fmt chunk (PCM format tag, sample rate, bits per sample, channels), data chunk, LIST/INFO chunk, fact chunk',
        'fields': '采样率(sample rate), 位深度(bits per sample), 声道数(channels), 音频格式(PCM/IEEE float/ADPCM), 数据大小',
        'errors': '不是 RIFF 文件, fmt chunk 缺失, 不支持的格式标签, 数据截断',
    },
    'mp3': {
        'tags': 'ID3v1 (128 bytes tail), ID3v2 (header with frames: TIT2 title, TPE1 artist, TALB album, TRCK track, TYER year)',
        'frames': 'MPEG frame header: version(1/2), layer(I/II/III), bitrate index, sample rate index, channel mode',
        'errors': '非 MP3 文件, ID3 标签损坏, 帧同步丢失',
    },
    'flac': {
        'blocks': 'STREAMINFO (sample rate, channels, bits per sample, total samples), VORBIS_COMMENT (vendor string, key=value pairs), CUESHEET, PICTURE',
        'fields': '采样率(1-655350Hz), 声道数(1-8), 位深度(4-32bit), MD5 校验',
        'errors': '不是 FLAC 文件(magic: fLaC), STREAMINFO 缺失, block 大小异常',
    },
    'mp4': {
        'boxes': 'ftyp(文件类型), moov(容器), mvhd(影片头), tkhd(轨道头), stts/stsc/stsz/stco(采样表), mdat(媒体数据)',
        'fields': '创建时间, 修改时间, 时长, timescale, 轨道类型(video/audio/subtitle), 编解码器',
        'errors': '不是 MP4 文件, ftyp box 缺失, moov box 损坏, 嵌套层级过深',
    },
    'subtitle': {
        'formats': 'SRT(序号+时间轴+文本), ASS(样式定义+事件行), VTT(WEBVTT 头+cue块), SUB(MicroDVD帧号)',
        'fields': '序号, 开始时间, 结束时间, 文本内容, 样式(字体/颜色/位置/特效)',
        'errors': '时间格式错误, 编码问题(非 UTF-8), 序号不连续',
    },
}


def main():
    projects = get_projects()
    print(f"Found {len(projects)} projects in range 9001-9250")
    
    # Filter to those needing expansion (< 500 chars)
    need_expand = [(num, d, pp, c) for num, d, pp, c in projects if char_count(c) < 500]
    skip = [(num, d, pp, c) for num, d, pp, c in projects if char_count(c) >= 500]
    
    print(f"Need expansion: {len(need_expand)}, Already OK: {len(skip)}")
    print(f"Skipped (already >= 500 chars): {[d for _, d, _, _ in skip]}")
    
    # Output the list for processing
    for num, d, pp, c in need_expand:
        print(f"EXPAND|{num}|{d}|{char_count(c)}")


if __name__ == '__main__':
    main()
