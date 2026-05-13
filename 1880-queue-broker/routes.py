import json
from aiohttp import web
from typing import Optional


def get_user(request) -> tuple:
    user = request.headers.get('X-User', 'anonymous')
    is_admin = request.headers.get('X-Role') == 'admin'
    return user, is_admin


async def create_topic(request):
    broker = request.app['broker']
    user, _ = get_user(request)
    
    try:
        data = await request.json()
        name = data.get('name')
        max_size = data.get('max_size', 1000)
        
        if not name:
            return web.json_response({'error': 'Topic name is required'}, status=400)
        
        try:
            topic = broker.create_topic(name, user, max_size)
            return web.json_response({
                'name': topic.name,
                'max_size': topic.max_size,
                'creator': topic.creator
            }, status=201)
        except ValueError as e:
            return web.json_response({'error': str(e)}, status=409)
    except json.JSONDecodeError:
        return web.json_response({'error': 'Invalid JSON'}, status=400)


async def list_topics(request):
    broker = request.app['broker']
    topics = broker.list_topics()
    return web.json_response({'topics': topics})


async def publish_message(request):
    broker = request.app['broker']
    topic_name = request.match_info['name']
    
    try:
        data = await request.json()
        content = data.get('content')
        priority = data.get('priority', 0)
        
        if content is None:
            return web.json_response({'error': 'Content is required'}, status=400)
        
        try:
            message = await broker.publish_message(topic_name, str(content), priority)
            return web.json_response({
                'id': message.id,
                'content': message.content,
                'priority': message.priority
            }, status=201)
        except ValueError as e:
            return web.json_response({'error': str(e)}, status=404)
    except json.JSONDecodeError:
        return web.json_response({'error': 'Invalid JSON'}, status=400)


async def subscribe(request):
    broker = request.app['broker']
    topic_name = request.match_info['name']
    group_id = request.match_info['gid']
    
    try:
        data = await request.json()
        filter_expr = data.get('filter_expr')
    except json.JSONDecodeError:
        filter_expr = None
    
    try:
        group = await broker.subscribe(topic_name, group_id, filter_expr)
        return web.json_response({
            'group_id': group.id,
            'filter_expr': group.filter_expr,
            'offset': group.offset
        }, status=200)
    except ValueError as e:
        return web.json_response({'error': str(e)}, status=404)


async def consume(request):
    broker = request.app['broker']
    topic_name = request.match_info['name']
    group_id = request.match_info['gid']
    
    try:
        message = await broker.consume_message(topic_name, group_id)
        if message is None:
            return web.json_response({}, status=204)
        
        return web.json_response({
            'id': message.id,
            'content': message.content,
            'priority': message.priority,
            'retry_count': message.retry_count
        })
    except ValueError as e:
        return web.json_response({'error': str(e)}, status=404)


async def get_offset(request):
    broker = request.app['broker']
    topic_name = request.match_info['name']
    group_id = request.match_info['gid']
    
    try:
        offset = await broker.get_offset(topic_name, group_id)
        return web.json_response({'offset': offset})
    except ValueError as e:
        return web.json_response({'error': str(e)}, status=404)


async def mark_failed(request):
    broker = request.app['broker']
    topic_name = request.match_info['name']
    group_id = request.match_info['gid']
    
    try:
        data = await request.json()
        message_id = data.get('message_id')
        
        if message_id is None:
            return web.json_response({'error': 'Message ID is required'}, status=400)
        
        success = await broker.mark_message_failed(topic_name, message_id)
        if success:
            return web.json_response({'status': 'marked_failed'})
        else:
            return web.json_response({'error': 'Message not found'}, status=404)
    except json.JSONDecodeError:
        return web.json_response({'error': 'Invalid JSON'}, status=400)


async def get_dead_letter(request):
    broker = request.app['broker']
    topic_name = request.match_info['name']
    user, is_admin = get_user(request)
    
    if not broker.is_authorized(topic_name, user, is_admin):
        return web.json_response({'error': 'Forbidden'}, status=403)
    
    try:
        messages = await broker.get_dead_letter(topic_name)
        return web.json_response({
            'messages': [
                {
                    'id': m.id,
                    'content': m.content,
                    'priority': m.priority,
                    'retry_count': m.retry_count
                }
                for m in messages
            ]
        })
    except ValueError as e:
        return web.json_response({'error': str(e)}, status=404)


async def render_message(request):
    broker = request.app['broker']
    topic_name = request.match_info['name']
    message_id = int(request.match_info['mid'])
    user, is_admin = get_user(request)
    
    if not broker.is_authorized(topic_name, user, is_admin):
        return web.Response(text='<html><body><h1>403 Forbidden</h1></body></html>', 
                           status=403, content_type='text/html')
    
    message = broker.get_message(topic_name, message_id)
    if not message:
        return web.Response(text='<html><body><h1>404 Not Found</h1></body></html>',
                           status=404, content_type='text/html')
    
    html_content = f'''
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Message {message_id}</title>
</head>
<body>
    <h1>Message ID: {message_id}</h1>
    <p><strong>Topic:</strong> {topic_name}</p>
    <p><strong>Priority:</strong> {message.priority}</p>
    <p><strong>Retry Count:</strong> {message.retry_count}</p>
    <hr>
    <div style="border: 1px solid #ccc; padding: 10px; margin: 10px 0;">
        <h3>Content:</h3>
        {message.content}
    </div>
</body>
</html>
'''
    return web.Response(text=html_content, content_type='text/html')


async def health_check(request):
    return web.json_response({'status': 'ok'})


def setup_routes(app):
    app.router.add_get('/health', health_check)
    app.router.add_post('/topics', create_topic)
    app.router.add_get('/topics', list_topics)
    app.router.add_post('/topics/{name}/messages', publish_message)
    app.router.add_post('/topics/{name}/groups/{gid}/subscribe', subscribe)
    app.router.add_get('/topics/{name}/groups/{gid}/consume', consume)
    app.router.add_post('/topics/{name}/groups/{gid}/failed', mark_failed)
    app.router.add_get('/topics/{name}/groups/{gid}/offset', get_offset)
    app.router.add_get('/topics/{name}/dead-letter', get_dead_letter)
    app.router.add_get('/topics/{name}/messages/{mid}/render', render_message)
