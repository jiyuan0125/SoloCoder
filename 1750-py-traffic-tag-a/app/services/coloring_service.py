import ipaddress
from typing import Dict, List, Optional

from app.models import ColoringRule, CookieMatch, HeaderMatch, IPRange


class ColoringService:
    def __init__(self):
        self._rules: Dict[str, ColoringRule] = {}

    def add_rule(self, rule: ColoringRule) -> None:
        self._rules[rule.id] = rule

    def remove_rule(self, rule_id: str) -> bool:
        if rule_id in self._rules:
            del self._rules[rule_id]
            return True
        return False

    def get_rule(self, rule_id: str) -> Optional[ColoringRule]:
        return self._rules.get(rule_id)

    def list_rules(self) -> List[ColoringRule]:
        return sorted(self._rules.values(), key=lambda r: (-r.priority, r.id))

    def get_tag(
        self,
        headers: Dict[str, str],
        cookies: Dict[str, str],
        client_ip: str,
    ) -> str:
        sorted_rules = sorted(
            self._rules.values(),
            key=lambda r: (-r.priority, r.id),
        )

        for rule in sorted_rules:
            if self._match_rule(rule, headers, cookies, client_ip):
                return rule.tag

        return "default"

    def _match_rule(
        self,
        rule: ColoringRule,
        headers: Dict[str, str],
        cookies: Dict[str, str],
        client_ip: str,
    ) -> bool:
        has_condition = False

        if rule.headers:
            has_condition = True
            if not self._match_headers(rule.headers, headers):
                return False

        if rule.cookies:
            has_condition = True
            if not self._match_cookies(rule.cookies, cookies):
                return False

        if rule.ip_ranges:
            has_condition = True
            if not self._match_ip_ranges(rule.ip_ranges, client_ip):
                return False

        if not has_condition:
            return True

        return True

    def _match_headers(
        self,
        header_matches: List[HeaderMatch],
        headers: Dict[str, str],
    ) -> bool:
        for hm in header_matches:
            actual_value = headers.get(hm.name.lower(), "")
            if hm.exact:
                if actual_value != hm.value:
                    return False
            else:
                if hm.value not in actual_value:
                    return False
        return True

    def _match_cookies(
        self,
        cookie_matches: List[CookieMatch],
        cookies: Dict[str, str],
    ) -> bool:
        for cm in cookie_matches:
            actual_value = cookies.get(cm.name, "")
            if cm.exact:
                if actual_value != cm.value:
                    return False
            else:
                if cm.value not in actual_value:
                    return False
        return True

    def _match_ip_ranges(
        self,
        ip_ranges: List[IPRange],
        client_ip: str,
    ) -> bool:
        try:
            ip = ipaddress.ip_address(client_ip)
        except ValueError:
            return False

        for ip_range in ip_ranges:
            try:
                network = ipaddress.ip_network(ip_range.cidr, strict=False)
                if ip in network:
                    return True
            except ValueError:
                continue
        return False
