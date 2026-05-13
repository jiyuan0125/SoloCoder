import os
from aiohttp import web

from log_aggregator.web import create_app


def main():
    port = int(os.environ.get("PORT", 9200))
    log_dir = os.environ.get("LOG_DIR", "./logs")
    
    app = create_app(log_dir=log_dir)
    
    print(f"Starting log aggregator on port {port}...")
    print(f"Log directory: {os.path.abspath(log_dir)}")
    print(f"Endpoints:")
    print(f"  POST /logs              - Collect logs")
    print(f"  GET  /logs              - Query logs")
    print(f"  PUT  /services/:name/config  - Update service config")
    print(f"  GET  /services/:name/config  - Get service config")
    print(f"  GET  /services          - List all services")
    print(f"  GET  /health            - Health check")
    
    web.run_app(app, port=port)


if __name__ == "__main__":
    main()
