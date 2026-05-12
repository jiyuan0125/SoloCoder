from typing import Dict, Any, List
from sqlalchemy.orm import Session
from .models import RedactionRuleModel
from .schemas import RedactionType


def mask_value(value: Any) -> str:
    if value is None:
        return "***"
    
    str_val = str(value)
    if len(str_val) <= 4:
        return "*" * len(str_val)
    
    if len(str_val) == 11 and str_val.isdigit():
        return f"{str_val[:3]}****{str_val[7:]}"
    
    if len(str_val) == 18:
        return f"{str_val[:6]}********{str_val[14:]}"
    
    mask_len = max(1, len(str_val) - 6)
    return f"{str_val[:3]}{'*' * mask_len}{str_val[-3:]}"


def get_redaction_rules(db: Session) -> Dict[str, Dict[str, Any]]:
    rules = db.query(RedactionRuleModel).all()
    result = {}
    for rule in rules:
        result[rule.tag_key] = {
            "redaction_type": RedactionType(rule.redaction_type),
            "replacement": rule.replacement
        }
    return result


def redact_tags(tags: Dict[str, Any], rules: Dict[str, Dict[str, Any]]) -> Dict[str, Any]:
    if not tags:
        return tags
    
    result = {}
    for key, value in tags.items():
        if key in rules:
            rule = rules[key]
            redaction_type = rule["redaction_type"]
            
            if redaction_type == RedactionType.mask:
                result[key] = mask_value(value)
            elif redaction_type == RedactionType.replace:
                result[key] = rule["replacement"] or "***"
        else:
            result[key] = value
    
    return result
