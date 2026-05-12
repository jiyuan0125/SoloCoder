import asyncio
from typing import Dict, List, Optional, Set, Tuple
from dataclasses import dataclass, field
from collections import defaultdict
from .config import RouteConfig, MatchType

@dataclass
class RouteState:
    route: RouteConfig
    active_requests: int = 0
    pending_delete: bool = False
    lock: asyncio.Lock = field(default_factory=asyncio.Lock)

class RouteMatcher:
    def __init__(self):
        self._exact_routes: Dict[str, RouteState] = {}
        self._prefix_routes: Dict[str, RouteState] = {}
        self._default_route: Optional[RouteState] = None
        self._lock: asyncio.Lock = asyncio.Lock()
        self._route_ids: Set[str] = set()

    def _normalize_path(self, path: str) -> str:
        if not path.startswith("/"):
            path = "/" + path
        if len(path) > 1 and path.endswith("/"):
            path = path[:-1]
        return path

    def _matches_prefix(self, request_path: str, prefix: str) -> bool:
        if request_path == prefix:
            return True
        if prefix == "/":
            return True
        return request_path.startswith(prefix + "/")

    async def add_route(self, route: RouteConfig) -> bool:
        async with self._lock:
            if route.id in self._route_ids:
                return False
            
            normalized_path = self._normalize_path(route.path)
            route_state = RouteState(route=route)
            
            if route.match_type == MatchType.EXACT:
                self._exact_routes[normalized_path] = route_state
            elif route.match_type == MatchType.PREFIX:
                self._prefix_routes[normalized_path] = route_state
            elif route.match_type == MatchType.DEFAULT:
                self._default_route = route_state
            
            self._route_ids.add(route.id)
            return True

    async def update_route(self, route: RouteConfig) -> bool:
        async with self._lock:
            if route.id not in self._route_ids:
                return False
            
            old_exact = None
            old_prefix = None
            
            for path, state in self._exact_routes.items():
                if state.route.id == route.id:
                    old_exact = path
                    break
            
            if old_exact:
                del self._exact_routes[old_exact]
            else:
                for path, state in self._prefix_routes.items():
                    if state.route.id == route.id:
                        old_prefix = path
                        break
                
                if old_prefix:
                    del self._prefix_routes[old_prefix]
                elif self._default_route and self._default_route.route.id == route.id:
                    self._default_route = None
            
            normalized_path = self._normalize_path(route.path)
            route_state = RouteState(route=route)
            
            if route.match_type == MatchType.EXACT:
                self._exact_routes[normalized_path] = route_state
            elif route.match_type == MatchType.PREFIX:
                self._prefix_routes[normalized_path] = route_state
            elif route.match_type == MatchType.DEFAULT:
                self._default_route = route_state
            
            return True

    async def remove_route(self, route_id: str) -> bool:
        async with self._lock:
            if route_id not in self._route_ids:
                return False
            
            route_state: Optional[RouteState] = None
            path_to_remove: Optional[str] = None
            is_default = False
            
            for path, state in self._exact_routes.items():
                if state.route.id == route_id:
                    path_to_remove = path
                    route_state = state
                    break
            
            if not route_state:
                for path, state in self._prefix_routes.items():
                    if state.route.id == route_id:
                        path_to_remove = path
                        route_state = state
                        break
            
            if not route_state and self._default_route and self._default_route.route.id == route_id:
                route_state = self._default_route
                is_default = True
            
            if not route_state:
                return False
            
            route_state.pending_delete = True
            
            async def _cleanup():
                async with route_state.lock:
                    while route_state.active_requests > 0:
                        await asyncio.sleep(0.1)
                    
                    async with self._lock:
                        if path_to_remove in self._exact_routes:
                            del self._exact_routes[path_to_remove]
                        elif path_to_remove in self._prefix_routes:
                            del self._prefix_routes[path_to_remove]
                        elif is_default:
                            self._default_route = None
                        self._route_ids.discard(route_id)
            
            asyncio.create_task(_cleanup())
            return True

    async def match_route(self, request_path: str) -> Tuple[Optional[RouteConfig], Optional[RouteState]]:
        normalized_path = self._normalize_path(request_path)
        
        async with self._lock:
            if normalized_path in self._exact_routes:
                state = self._exact_routes[normalized_path]
                if not state.pending_delete:
                    async with state.lock:
                        state.active_requests += 1
                    return state.route, state
            
            matching_prefixes: List[Tuple[str, RouteState]] = []
            for prefix, state in self._prefix_routes.items():
                if self._matches_prefix(normalized_path, prefix) and not state.pending_delete:
                    matching_prefixes.append((prefix, state))
            
            if matching_prefixes:
                matching_prefixes.sort(key=lambda x: len(x[0]), reverse=True)
                longest_prefix, state = matching_prefixes[0]
                async with state.lock:
                    state.active_requests += 1
                return state.route, state
            
            if self._default_route and not self._default_route.pending_delete:
                state = self._default_route
                async with state.lock:
                    state.active_requests += 1
                return state.route, state
        
        return None, None

    async def release_route(self, state: RouteState):
        async with state.lock:
            state.active_requests -= 1

    async def get_all_routes(self) -> List[RouteConfig]:
        async with self._lock:
            routes = []
            for state in self._exact_routes.values():
                routes.append(state.route)
            for state in self._prefix_routes.values():
                routes.append(state.route)
            if self._default_route:
                routes.append(self._default_route.route)
            return routes

    async def get_route(self, route_id: str) -> Optional[RouteConfig]:
        async with self._lock:
            for state in self._exact_routes.values():
                if state.route.id == route_id:
                    return state.route
            for state in self._prefix_routes.values():
                if state.route.id == route_id:
                    return state.route
            if self._default_route and self._default_route.route.id == route_id:
                return self._default_route.route
            return None
