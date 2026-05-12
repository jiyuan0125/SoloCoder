#!/bin/bash

PORT=${PORT:-8000}

echo "Starting server on port $PORT"
exec python3 main.py
