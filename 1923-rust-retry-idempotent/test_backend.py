import http.server
import socketserver
import json

class MyHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        print(f"GET {self.path}")
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        response = {
            "message": "Hello from backend",
            "path": self.path,
            "method": "GET"
        }
        self.wfile.write(json.dumps(response).encode())
    
    def do_POST(self):
        print(f"POST {self.path}")
        content_length = int(self.headers.get('Content-Length', 0))
        body = self.rfile.read(content_length) if content_length > 0 else b""
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        response = {
            "message": "Hello from backend",
            "path": self.path,
            "method": "POST",
            "body_length": len(body)
        }
        self.wfile.write(json.dumps(response).encode())

PORT = 8080
with socketserver.TCPServer(("", PORT), MyHandler) as httpd:
    print(f"Backend server running on port {PORT}")
    httpd.serve_forever()
