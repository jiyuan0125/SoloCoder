import json
import os
import time
import uuid
from collections import OrderedDict
from typing import Dict, List, Optional, Any


class ConfigStore:
    def __init__(self, data_dir: str = "data"):
        self.data_dir = data_dir
        self.projects_path = os.path.join(data_dir, "projects.json")
        self.versions_path = os.path.join(data_dir, "versions.json")
        self.watches_path = os.path.join(data_dir, "watches.json")
        
        os.makedirs(data_dir, exist_ok=True)
        self._init_data()

    def _init_data(self):
        if not os.path.exists(self.projects_path):
            self._save_json(self.projects_path, {})
        if not os.path.exists(self.versions_path):
            self._save_json(self.versions_path, {})
        if not os.path.exists(self.watches_path):
            self._save_json(self.watches_path, [])

    def _save_json(self, path: str, data: Any):
        with open(path, "w", encoding="utf-8") as f:
            json.dump(data, f, ensure_ascii=False, indent=2)

    def _load_json(self, path: str) -> Any:
        with open(path, "r", encoding="utf-8") as f:
            return json.load(f)

    def _get_projects(self) -> Dict:
        return self._load_json(self.projects_path)

    def _save_projects(self, projects: Dict):
        self._save_json(self.projects_path, projects)

    def _get_versions(self) -> Dict:
        return self._load_json(self.versions_path)

    def _save_versions(self, versions: Dict):
        self._save_json(self.versions_path, versions)

    def _get_watches(self) -> List:
        return self._load_json(self.watches_path)

    def _save_watches(self, watches: List):
        self._save_json(self.watches_path, watches)

    def get_config(self, project: str, group: str) -> Dict[str, Any]:
        projects = self._get_projects()
        project_data = projects.get(project, {"groups": {}})
        
        all_keys = {}
        group_chain = self._get_group_chain(project_data.get("groups", {}), group)
        
        for ancestor in reversed(group_chain):
            group_data = project_data["groups"].get(ancestor, {})
            all_keys.update(group_data.get("keys", {}))
        
        return all_keys

    def _get_group_chain(self, groups: Dict, group: str) -> List[str]:
        chain = [group]
        visited = set([group])
        current = group
        
        while True:
            group_data = groups.get(current, {})
            parent = group_data.get("parent")
            if not parent or parent in visited:
                break
            visited.add(parent)
            chain.append(parent)
            current = parent
        
        return chain

    def get_all_configs(self, project: str) -> Dict[str, Dict[str, Any]]:
        projects = self._get_projects()
        project_data = projects.get(project, {"groups": {}})
        groups = project_data.get("groups", {})
        
        result = {}
        for group_name in groups:
            result[group_name] = self.get_config(project, group_name)
        
        return result

    def put_config(self, project: str, group: str, keys: Dict[str, Any], parent: Optional[str] = None) -> Dict:
        projects = self._get_projects()
        versions = self._get_versions()
        
        if project not in projects:
            projects[project] = {"groups": {}}
        
        project_data = projects[project]
        if "groups" not in project_data:
            project_data["groups"] = {}
        
        old_group_data = project_data["groups"].get(group, {"keys": {}})
        old_keys = old_group_data.get("keys", {})
        
        if parent is not None:
            old_group_data["parent"] = parent
        
        new_group_data = {
            "keys": {**old_keys, **keys},
            "parent": old_group_data.get("parent")
        }
        
        project_data["groups"][group] = new_group_data
        
        version_id = str(uuid.uuid4())
        version_info = {
            "version_id": version_id,
            "timestamp": time.time(),
            "project": project,
            "group": group,
            "changes": self._calculate_changes(old_keys, keys)
        }
        
        if project not in versions:
            versions[project] = []
        versions[project].append(version_info)
        
        self._save_projects(projects)
        self._save_versions(versions)
        
        return version_info

    def _calculate_changes(self, old_keys: Dict, new_keys: Dict) -> List[Dict]:
        changes = []
        all_keys = set(old_keys.keys()) | set(new_keys.keys())
        
        for key in all_keys:
            if key not in old_keys:
                changes.append({
                    "key": key,
                    "action": "added",
                    "new_value": new_keys[key]
                })
            elif key not in new_keys:
                changes.append({
                    "key": key,
                    "action": "deleted",
                    "old_value": old_keys[key]
                })
            elif old_keys[key] != new_keys[key]:
                changes.append({
                    "key": key,
                    "action": "modified",
                    "old_value": old_keys[key],
                    "new_value": new_keys[key]
                })
        
        return changes

    def get_versions(self, project: str) -> List[Dict]:
        versions = self._get_versions()
        return versions.get(project, [])

    def rollback(self, project: str, version_id: str) -> Optional[Dict]:
        projects = self._get_projects()
        versions = self._get_versions()
        
        if project not in versions:
            return None
        
        target_index = None
        for i, v in enumerate(versions[project]):
            if v["version_id"] == version_id:
                target_index = i
                break
        
        if target_index is None:
            return None
        
        if project not in projects:
            projects[project] = {"groups": {}}
        
        for i in range(len(versions[project]) - 1, target_index, -1):
            version = versions[project][i]
            group = version["group"]
            
            if group in projects[project].get("groups", {}):
                group_data = projects[project]["groups"][group]
                
                for change in reversed(version["changes"]):
                    action = change["action"]
                    key = change["key"]
                    
                    if action == "added":
                        if key in group_data["keys"]:
                            del group_data["keys"][key]
                    elif action == "deleted":
                        group_data["keys"][key] = change["old_value"]
                    elif action == "modified":
                        group_data["keys"][key] = change["old_value"]
        
        versions[project] = versions[project][:target_index + 1]
        
        self._save_projects(projects)
        self._save_versions(versions)
        
        return {"success": True, "version_id": version_id}

    def add_watch(self, project: str, callback_url: str) -> Dict:
        watches = self._get_watches()
        
        watch_info = {
            "watch_id": str(uuid.uuid4()),
            "project": project,
            "callback_url": callback_url,
            "last_notified_at": None,
            "last_notified_status": None,
            "created_at": time.time()
        }
        
        watches.append(watch_info)
        self._save_watches(watches)
        
        return watch_info

    def get_watches(self, project: str) -> List[Dict]:
        watches = self._get_watches()
        return [w for w in watches if w["project"] == project]

    def update_watch_status(self, watch_id: str, status: str, timestamp: float):
        watches = self._get_watches()
        for watch in watches:
            if watch["watch_id"] == watch_id:
                watch["last_notified_at"] = timestamp
                watch["last_notified_status"] = status
                break
        self._save_watches(watches)
