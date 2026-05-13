import os
from aiohttp import web
from routes import setup_routes
from broker import MessageBroker

async def on_startup(app):
    await app['broker'].start()

async def on_cleanup(app):
    await app['broker'].stop()

def create_app():
    app = web.Application()
    app['broker'] = MessageBroker()
    app.on_startup.append(on_startup)
    app.on_cleanup.append(on_cleanup)
    setup_routes(app)
    return app

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8800))
    app = create_app()
    web.run_app(app, port=port)
