import difflib
import math
from typing import List, Tuple
from models import LostItem, FoundItem
from config import settings


def calculate_similarity(text1: str, text2: str) -> float:
    if not text1 or not text2:
        return 0.0
    return difflib.SequenceMatcher(None, text1, text2).ratio()


def calculate_location_similarity(loc1: str, loc2: str) -> float:
    if not loc1 or not loc2:
        return 0.5
    return difflib.SequenceMatcher(None, loc1, loc2).ratio()


def calculate_total_score(similarity: float, location_similarity: float, weight_sim: float = 0.6, weight_loc: float = 0.4) -> float:
    return weight_sim * similarity + weight_loc * location_similarity


def match_items(lost_item: LostItem, found_items: List[FoundItem]) -> List[Tuple[FoundItem, float, float, float]]:
    if not lost_item.description:
        return []
    
    results = []
    for found_item in found_items:
        if found_item.status.value != "待认领":
            continue
        
        sim_score = calculate_similarity(
            f"{lost_item.item_type} {lost_item.description}",
            f"{found_item.item_type} {found_item.description}"
        )
        loc_score = calculate_location_similarity(lost_item.location, found_item.location)
        total_score = calculate_total_score(sim_score, loc_score)
        
        if total_score >= settings.MATCHING_THRESHOLD:
            results.append((found_item, sim_score, loc_score, total_score))
    
    results.sort(key=lambda x: x[3], reverse=True)
    return results
