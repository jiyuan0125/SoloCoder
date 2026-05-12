#!/bin/bash

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$BASE_DIR"

echo "=========================================="
echo "  体育馆场地预约管理系统"
echo "=========================================="
echo ""

case "$1" in
    install)
        echo "安装依赖..."
        pip install -r requirements.txt
        ;;
        
    init)
        echo "初始化数据库和测试数据..."
        python init_data.py
        ;;
        
    server)
        echo "启动API服务器..."
        python -m uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
        ;;
        
    api)
        echo "启动API服务器..."
        python -m uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
        ;;
        
    cli)
        shift
        python cli.py "$@"
        ;;
        
    scheduler)
        echo "运行定时任务..."
        python cli.py scheduler
        ;;
        
    help|--help|-h)
        echo "用法: $0 [命令] [参数]"
        echo ""
        echo "命令:"
        echo "  install         安装Python依赖"
        echo "  init            初始化数据库和测试数据"
        echo "  server|api      启动FastAPI服务器 (端口8000)"
        echo "  cli             运行命令行客户端"
        echo "  scheduler       手动运行定时任务"
        echo ""
        echo "CLI 命令示例:"
        echo "  $0 cli bookings              查看所有预约"
        echo "  $0 cli bookings --user 2     查看用户ID=2的预约"
        echo "  $0 cli booking 1             查看预约ID=1的详情"
        echo "  $0 cli matches               查看所有赛程"
        echo "  $0 cli matches --tournament 1  查看赛事ID=1的赛程"
        echo "  $0 cli stats 1               查看赛事ID=1的积分榜"
        echo "  $0 cli tournaments           查看所有赛事"
        echo "  $0 cli classes               查看所有培训班"
        echo "  $0 cli venues                查看所有场地"
        echo "  $0 cli users                 查看所有用户"
        echo ""
        ;;
        
    *)
        echo "未知命令: $1"
        echo "使用 '$0 help' 查看帮助"
        ;;
esac
