from typing import Dict, List, Optional

from app.models import RouteRule, ServiceRoutes


class RoutingService:
    def __init__(self):
        self._service_routes: Dict[str, ServiceRoutes] = {}

    def add_service(self, service_name: str) -> ServiceRoutes:
        if service_name not in self._service_routes:
            self._service_routes[service_name] = ServiceRoutes(
                service_name=service_name,
                rules={},
            )
        return self._service_routes[service_name]

    def get_service(self, service_name: str) -> Optional[ServiceRoutes]:
        return self._service_routes.get(service_name)

    def list_services(self) -> List[str]:
        return list(self._service_routes.keys())

    def add_route_rule(self, service_name: str, rule: RouteRule) -> RouteRule:
        service = self.add_service(service_name)
        service.rules[rule.tag] = rule
        return rule

    def remove_route_rule(self, service_name: str, tag: str) -> bool:
        service = self._service_routes.get(service_name)
        if service and tag in service.rules:
            del service.rules[tag]
            return True
        return False

    def get_route_rule(self, service_name: str, tag: str) -> Optional[RouteRule]:
        service = self._service_routes.get(service_name)
        if service:
            return service.rules.get(tag)
        return None

    def get_upstream(self, service_name: str, tag: str) -> Optional[str]:
        rule = self.get_route_rule(service_name, tag)
        if rule:
            return rule.upstream
        default_rule = self.get_route_rule(service_name, "default")
        if default_rule:
            return default_rule.upstream
        return None
