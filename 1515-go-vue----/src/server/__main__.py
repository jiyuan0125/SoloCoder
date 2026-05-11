import os
import uvicorn
import sys

def main():
    port = int(os.getenv("PORT", "8000"))
    print(f"启动宠物寄养中心管理系统服务端，端口: {port}")
    uvicorn.run("server.app:app", host="0.0.0.0", port=port, reload=False)

if __name__ == "__main__":
    main()
