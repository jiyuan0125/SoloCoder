import os
import asyncio
import json
import logging
import hashlib
import time
from datetime import datetime
from typing import Dict, List, Any, Optional, Tuple
from dataclasses import dataclass, field, asdict
from collections import deque

from aiohttp import web
import aiohttp


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@dataclass
class GroupConfig:
    parent: Optional[str] = None
    keys: Dict[str, Any] = field(default_factory=dict)


@dataclass
class Project:
    groups: Dict[str, GroupConfig] = field(default_factory=dict)
    versions: List[Dict] = field(default_factory=list)


@dataclass
class Watcher:
    project: str
    callback_url: str
    last_notified: Optional[datetime] = None
    last_status: Optional[str] = None


class ConfigService:
    def __init__(self):
        self._projects: Dict[str, Project] = {}
        self._watchers: Dict[str, Watcher] = {}
        self._data_file = 'config_data.json'
        self._load()

    def _load(self):
        if os.path.exists(self._data_file):
            try:
                with open(self._data_file, 'r') as f:
                    data = json.load(f)
                for proj_name, proj_data in data.get('projects', {}).items():
                    project = Project()
                    for group_name, group_data in proj_data.get('groups', {}).items():
                        project.groups[group_name] = GroupConfig(
                            parent=group_data.get('parent'),
                            keys=group_data.get('keys', {})
                        )
                    project.versions = proj_data.get('versions', [])
                    self._projects[proj_name] = project
                self._watchers = {k: Watcher(**v) for k, v in data.get('watchers', {}).items()}
                logger.info(f"Loaded {len(self._projects)} projects")
            except Exception as e:
                logger.error(f"Failed to load data: {e}")

    def _save(self):
        data = {
            'projects': {
                name: {
                    'groups': {
                        g: {'parent': gc.parent, 'keys': gc.keys}
                        for g, gc in project.groups.items()
                    },
                    'versions': project.versions
                }
                for name, project in self._projects.items()
            },
            'watchers': {k: asdict(v) for k, v in self._watchers.items()}
        }
        with open(self._data_file, 'w') as f:
            json.dump(data, f, indent=2, default=str)

    def _generate_version(self, project_name: str, changes: List[Dict]) -> int:
        project = self._projects.get(project_name, Project())
        version_number = len(project.versions) + 1
        timestamp = datetime.now().isoformat()
        
        groups_snapshot = {
            g: {'parent': gc.parent, 'keys': dict(gc.keys)}
            for g, gc in project.groups.items()
        }
        
        version_entry = {
            'version': version_number,
            'timestamp': timestamp,
            'changes': changes,
            'groups': groups_snapshot
        }
        project.versions.append(version_entry)
        self._projects[project_name] = project
        return version_number

    def _resolve_group_config(self, project: Project, group_name: str) -> Tuple[Dict[str, Any], List[str]]:
        config = {}
        chain = []
        visited = set()
        
        current = group_name
        while current and current not in visited:
            visited.add(current)
            chain.append(current)
            if current in project.groups:
                current = project.groups[current].parent
            else:
                break
        
        for group_name in reversed(chain):
            if group_name in project.groups:
                config.update(project.groups[group_name].keys)
        
        return config, chain

    def get_config(self, project_name: str, group_name: str) -> Dict[str, Any]:
        if project_name not in self._projects:
            return {}
        project = self._projects[project_name]
        config, _ = self._resolve_group_config(project, group_name)
        return config

    def get_all_groups_config(self, project_name: str) -> Dict[str, Any]:
        if project_name not in self._projects:
            return {}
        project = self._projects[project_name]
        result = {}
        for group_name in project.groups:
            result[group_name] = self.get_config(project_name, group_name)
        return result

    def set_group_parent(self, project_name: str, group_name: str, parent_name: Optional[str]) -> bool:
        project = self._projects.setdefault(project_name, Project())
        if group_name not in project.groups:
            project.groups[group_name] = GroupConfig()
        
        old_config = self.get_config(project_name, group_name)
        project.groups[group_name].parent = parent_name
        new_config = self.get_config(project_name, group_name)
        
        changes = self._compare_configs(old_config, new_config)
        if changes:
            self._generate_version(project_name, changes)
        
        self._save()
        return True

    def set_keys(self, project_name: str, group_name: str, keys: Dict[str, Any]) -> Tuple[int, List[Dict]]:
        project = self._projects.setdefault(project_name, Project())
        if group_name not in project.groups:
            project.groups[group_name] = GroupConfig()
        
        group = project.groups[group_name]
        old_keys = dict(group.keys)
        
        added = []
        modified = []
        deleted = []
        
        for key, value in keys.items():
            if value is None:
                if key in old_keys:
                    deleted.append({'key': key, 'old_value': old_keys[key]})
                    del group.keys[key]
            else:
                if key not in old_keys:
                    added.append({'key': key, 'new_value': value})
                elif old_keys[key] != value:
                    modified.append({'key': key, 'old_value': old_keys[key], 'new_value': value})
                group.keys[key] = value
        
        changes = []
        if added:
            changes.append({'action': 'add', 'items': added})
        if modified:
            changes.append({'action': 'modify', 'items': modified})
        if deleted:
            changes.append({'action': 'delete', 'items': deleted})
        
        version = self._generate_version(project_name, changes) if changes else len(project.versions)
        self._save()
        return version, changes

    def get_versions(self, project_name: str) -> List[Dict]:
        if project_name not in self._projects:
            return []
        return self._projects[project_name].versions

    def _resolve_groups_snapshot(self, groups_snapshot: Dict[str, Any], group_name: str) -> Dict[str, Any]:
        config = {}
        chain = []
        visited = set()
        
        current = group_name
        while current and current not in visited:
            visited.add(current)
            chain.append(current)
            if current in groups_snapshot:
                current = groups_snapshot[current]['parent']
            else:
                break
        
        for g in reversed(chain):
            if g in groups_snapshot:
                config.update(groups_snapshot[g]['keys'])
        
        return config

    def get_version_diff(self, project_name: str, v1: int, v2: int) -> List[Dict]:
        versions = self.get_versions(project_name)
        if not versions:
            return []
        
        min_v = min(v1, v2) - 1
        max_v = max(v1, v2) - 1
        
        if min_v < 0 or max_v >= len(versions):
            return []
        
        groups1 = versions[min_v]['groups']
        groups2 = versions[max_v]['groups']
        
        all_groups = set(groups1.keys()) | set(groups2.keys())
        
        diffs = []
        for group in sorted(all_groups):
            g1_resolved = self._resolve_groups_snapshot(groups1, group)
            g2_resolved = self._resolve_groups_snapshot(groups2, group)
            for key in sorted(set(g1_resolved.keys()) | set(g2_resolved.keys())):
                if g1_resolved.get(key) != g2_resolved.get(key):
                    diffs.append({
                        'group': group,
                        'key': key,
                        'old_value': g1_resolved.get(key),
                        'new_value': g2_resolved.get(key)
                    })
        
        return diffs

    def rollback(self, project_name: str, version_number: int) -> Tuple[bool, List[Dict]]:
        if project_name not in self._projects:
            return False, []
        
        project = self._projects[project_name]
        if version_number < 1 or version_number > len(project.versions):
            return False, []
        
        target_version = project.versions[version_number - 1]
        target_groups = target_version['groups']
        
        old_full_config = self.get_all_groups_config(project_name)
        
        for group_name, group_data in target_groups.items():
            if group_name not in project.groups:
                project.groups[group_name] = GroupConfig()
            project.groups[group_name].parent = group_data['parent']
            project.groups[group_name].keys = dict(group_data['keys'])
        
        for group_name in list(project.groups.keys()):
            if group_name not in target_groups:
                del project.groups[group_name]
        
        new_full_config = self.get_all_groups_config(project_name)
        
        changes = []
        for group in sorted(set(old_full_config.keys()) | set(new_full_config.keys())):
            old_g = old_full_config.get(group, {})
            new_g = new_full_config.get(group, {})
            for key in sorted(set(old_g.keys()) | set(new_g.keys())):
                if old_g.get(key) != new_g.get(key):
                    changes.append({
                        'group': group,
                        'key': key,
                        'old_value': old_g.get(key),
                        'new_value': new_g.get(key)
                    })
        
        self._generate_version(project_name, [{'action': 'rollback', 'items': changes}])
        self._save()
        return True, changes

    def register_watch(self, project_name: str, callback_url: str) -> str:
        watcher_id = hashlib.sha256(f"{project_name}:{callback_url}:{time.time()}".encode()).hexdigest()[:12]
        self._watchers[watcher_id] = Watcher(
            project=project_name,
            callback_url=callback_url
        )
        self._save()
        return watcher_id

    def get_watchers(self, project_name: str) -> List[Dict]:
        return [
            {
                'id': wid,
                'callback_url': w.callback_url,
                'last_notified': w.last_notified.isoformat() if w.last_notified else None,
                'last_status': w.last_status
            }
            for wid, w in self._watchers.items()
            if w.project == project_name
        ]

    async def notify_watchers(self, project_name: str, version: int):
        watchers = [
            (wid, w) for wid, w in self._watchers.items()
            if w.project == project_name
        ]
        
        async def notify_single(wid: str, watcher: Watcher):
            payload = {
                'project': project_name,
                'version': version,
                'timestamp': datetime.now().isoformat()
            }
            
            for attempt in range(2):
                try:
                    async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=10)) as session:
                        async with session.post(watcher.callback_url, json=payload) as response:
                            if response.status >= 200 and response.status < 300:
                                self._watchers[wid].last_notified = datetime.now()
                                self._watchers[wid].last_status = 'success'
                                logger.info(f"Notified watcher {wid} (attempt {attempt+1})")
                                return
                            else:
                                logger.warning(f"Watcher {wid} returned status {response.status}")
                except Exception as e:
                    logger.warning(f"Watcher {wid} notification failed (attempt {attempt+1}): {e}")
                
                if attempt == 0:
                    await asyncio.sleep(1)
            
            self._watchers[wid].last_notified = datetime.now()
            self._watchers[wid].last_status = 'failed'
            logger.error(f"Watcher {wid} notification failed after retry")
        
        if watchers:
            await asyncio.gather(*[notify_single(wid, w) for wid, w in watchers])
            self._save()

    def _compare_configs(self, old_config: Dict[str, Any], new_config: Dict[str, Any]) -> List[Dict]:
        added = []
        modified = []
        deleted = []
        
        for key in set(old_config.keys()) | set(new_config.keys()):
            if key not in old_config:
                added.append({'key': key, 'new_value': new_config[key]})
            elif key not in new_config:
                deleted.append({'key': key, 'old_value': old_config[key]})
            elif old_config[key] != new_config[key]:
                modified.append({'key': key, 'old_value': old_config[key], 'new_value': new_config[key]})
        
        changes = []
        if added:
            changes.append({'action': 'add', 'items': added})
        if modified:
            changes.append({'action': 'modify', 'items': modified})
        if deleted:
            changes.append({'action': 'delete', 'items': deleted})
        
        return changes


config_service = ConfigService()


async def get_keys_handler(request: web.Request):
    project = request.match_info['project']
    group = request.match_info['group']
    config = config_service.get_config(project, group)
    return web.json_response(config)


async def put_keys_handler(request: web.Request):
    project = request.match_info['project']
    group = request.match_info['group']
    
    try:
        body = await request.json()
    except Exception:
        return web.json_response({'error': 'Invalid JSON'}, status=400)
    
    if not isinstance(body, dict):
        return web.json_response({'error': 'Body must be a JSON object'}, status=400)
    
    version, changes = config_service.set_keys(project, group, body)
    
    if changes:
        asyncio.create_task(config_service.notify_watchers(project, version))
    
    return web.json_response({'version': version, 'changes': changes})


async def get_versions_handler(request: web.Request):
    project = request.match_info['project']
    v1 = request.query.get('from')
    v2 = request.query.get('to')
    
    if v1 and v2:
        try:
            v1_int, v2_int = int(v1), int(v2)
            diffs = config_service.get_version_diff(project, v1_int, v2_int)
            return web.json_response({'diffs': diffs})
        except ValueError:
            return web.json_response({'error': 'Invalid version numbers'}, status=400)
    
    versions = config_service.get_versions(project)
    return web.json_response({'versions': versions})


async def rollback_handler(request: web.Request):
    project = request.match_info['project']
    
    try:
        body = await request.json()
    except Exception:
        return web.json_response({'error': 'Invalid JSON'}, status=400)
    
    version = body.get('version')
    if not isinstance(version, int):
        return web.json_response({'error': 'version must be an integer'}, status=400)
    
    success, changes = config_service.rollback(project, version)
    if not success:
        return web.json_response({'error': 'Version not found'}, status=404)
    
    if changes:
        current_version = len(config_service.get_versions(project))
        asyncio.create_task(config_service.notify_watchers(project, current_version))
    
    return web.json_response({'success': True, 'changes': changes})


async def watch_handler(request: web.Request):
    try:
        body = await request.json()
    except Exception:
        return web.json_response({'error': 'Invalid JSON'}, status=400)
    
    project = body.get('project')
    callback_url = body.get('callback_url')
    
    if not project or not callback_url:
        return web.json_response({'error': 'project and callback_url are required'}, status=400)
    
    watcher_id = config_service.register_watch(project, callback_url)
    return web.json_response({'id': watcher_id, 'status': 'registered'})


async def watch_status_handler(request: web.Request):
    project = request.query.get('project')
    if not project:
        return web.json_response({'error': 'project query param is required'}, status=400)
    
    watchers = config_service.get_watchers(project)
    return web.json_response({'watchers': watchers})


async def get_all_groups_handler(request: web.Request):
    project = request.match_info['project']
    all_config = config_service.get_all_groups_config(project)
    return web.json_response(all_config)


async def set_parent_handler(request: web.Request):
    project = request.match_info['project']
    group = request.match_info['group']
    
    try:
        body = await request.json()
    except Exception:
        return web.json_response({'error': 'Invalid JSON'}, status=400)
    
    parent = body.get('parent')
    if parent is not None and not isinstance(parent, str):
        return web.json_response({'error': 'parent must be a string or null'}, status=400)
    
    success = config_service.set_group_parent(project, group, parent)
    return web.json_response({'success': success})


def create_app():
    app = web.Application()
    
    app.router.add_get('/projects/{project}/groups', get_all_groups_handler)
    app.router.add_get('/projects/{project}/groups/{group}/keys', get_keys_handler)
    app.router.add_put('/projects/{project}/groups/{group}/keys', put_keys_handler)
    app.router.add_put('/projects/{project}/groups/{group}/parent', set_parent_handler)
    app.router.add_get('/projects/{project}/versions', get_versions_handler)
    app.router.add_post('/projects/{project}/rollback', rollback_handler)
    app.router.add_post('/watch', watch_handler)
    app.router.add_get('/watch/status', watch_status_handler)
    
    return app


if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8200))
    app = create_app()
    logger.info(f"Starting config center on port {port}")
    web.run_app(app, port=port)
