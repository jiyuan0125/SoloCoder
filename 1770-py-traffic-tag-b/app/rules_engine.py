import hashlib
import random
import threading
from typing import Optional, Dict, List
from fastapi import Request
from app.models import GrayRule, GrayStrategy, TrafficType


class GrayRuleEngine:
    def __init__(self):
        self._rules: Dict[str, GrayRule] = {}
        self._lock = threading.RLock()
        self._current_version: Optional[str] = None
        self._transition_enabled: bool = False
        self._transition_completion_callback = None

    def add_or_update_rule(self, rule: GrayRule) -> None:
        with self._lock:
            self._rules[rule.rule_id] = rule

    def remove_rule(self, rule_id: str) -> bool:
        with self._lock:
            if rule_id in self._rules:
                del self._rules[rule_id]
                return True
            return False

    def get_rule(self, rule_id: str) -> Optional[GrayRule]:
        with self._lock:
            return self._rules.get(rule_id)

    def get_all_rules(self) -> List[GrayRule]:
        with self._lock:
            return list(self._rules.values())

    def enable_transition(self, completion_callback=None):
        with self._lock:
            self._transition_enabled = True
            self._transition_completion_callback = completion_callback

    def disable_transition(self):
        with self._lock:
            self._transition_enabled = False
            if self._transition_completion_callback:
                self._transition_completion_callback()
                self._transition_completion_callback = None

    def _get_matching_rules(self, request: Request) -> List[GrayRule]:
        path = request.url.path
        rules = [
            rule for rule in self._rules.values()
            if rule.enabled and (rule.paths is None or path in rule.paths)
        ]
        rules.sort(key=lambda r: r.priority, reverse=True)
        return rules

    def _get_user_id(self, request: Request) -> Optional[str]:
        user_id = request.query_params.get("user_id")
        if user_id:
            return user_id
        user_id = request.headers.get("X-User-ID")
        if user_id:
            return user_id
        return None

    def _get_client_ip(self, request: Request) -> str:
        x_forwarded_for = request.headers.get("X-Forwarded-For")
        if x_forwarded_for:
            return x_forwarded_for.split(",")[0].strip()
        x_real_ip = request.headers.get("X-Real-IP")
        if x_real_ip:
            return x_real_ip
        client = request.client
        return client.host if client else "127.0.0.1"

    def _hash_to_percent(self, key: str, salt: str = "") -> float:
        h = hashlib.md5(f"{key}{salt}".encode()).hexdigest()
        return (int(h[:8], 16) % 10000) / 100.0

    def _check_rule(self, request: Request, rule: GrayRule) -> bool:
        strategy = rule.strategy

        if strategy == GrayStrategy.USER_ID:
            user_id = self._get_user_id(request)
            if user_id and rule.user_ids and user_id in rule.user_ids:
                return True
            return False

        if strategy == GrayStrategy.IP:
            client_ip = self._get_client_ip(request)
            if rule.ip_addresses and client_ip in rule.ip_addresses:
                return True
            return False

        if strategy == GrayStrategy.PARAMETER:
            if rule.parameter_name:
                value = request.query_params.get(rule.parameter_name)
                if value and rule.parameter_values and value in rule.parameter_values:
                    return True
            return False

        if strategy == GrayStrategy.HEADER:
            if rule.header_name:
                value = request.headers.get(rule.header_name)
                if value and rule.header_values and value in rule.header_values:
                    return True
            return False

        if strategy == GrayStrategy.COOKIE:
            if rule.cookie_name:
                value = request.cookies.get(rule.cookie_name)
                if value and rule.cookie_values and value in rule.cookie_values:
                    return True
            return False

        if strategy == GrayStrategy.WEIGHT:
            if rule.weight is not None and 0 <= rule.weight <= 100:
                user_id = self._get_user_id(request) or self._get_client_ip(request)
                percent = self._hash_to_percent(user_id, rule.rule_id)
                return percent < rule.weight
            return False

        return False

    def determine_traffic_type(self, request: Request) -> TrafficType:
        with self._lock:
            matching_rules = self._get_matching_rules(request)

            for rule in matching_rules:
                if self._check_rule(request, rule):
                    return TrafficType.GRAY

            return TrafficType.NORMAL


rule_engine = GrayRuleEngine()
