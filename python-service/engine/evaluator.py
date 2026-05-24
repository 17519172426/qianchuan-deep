import json
from datetime import datetime, timedelta
from typing import Optional
from models import ConditionEvalResult

SUPPORTED_METRICS = {"cost", "roi", "ctr", "conversions", "impressions", "cpa"}
SUPPORTED_OPERATORS = {"gt", "lt", "gte", "lte", "eq"}

class RuleEvaluator:
    def __init__(self, cooldown_store: Optional[dict] = None):
        self._cooldowns = cooldown_store or {}

    def evaluate_condition(self, condition_json: str, ad_context: dict, rule_id: int) -> ConditionEvalResult:
        condition = json.loads(condition_json) if isinstance(condition_json, str) else condition_json
        metric = condition.get("metric", "")
        operator = condition.get("operator", "")
        threshold = condition.get("threshold", 0)
        duration = condition.get("duration", "30m")

        if metric not in SUPPORTED_METRICS:
            return ConditionEvalResult(matched=False, metric_name=metric, current_value=0,
                                       threshold=threshold, operator=operator,
                                       description=f"unsupported metric: {metric}")

        if operator not in SUPPORTED_OPERATORS:
            return ConditionEvalResult(matched=False, metric_name=metric, current_value=0,
                                       threshold=threshold, operator=operator,
                                       description=f"unsupported operator: {operator}")

        current_value = ad_context.get(metric, 0)
        matched = self._compare(current_value, operator, threshold)

        return ConditionEvalResult(
            matched=matched,
            metric_name=metric,
            current_value=current_value,
            threshold=threshold,
            operator=operator,
            description=f"{metric}={current_value} {operator} {threshold}  {matched}"
        )

    def is_in_cooldown(self, rule_id: int, ad_id: int, cooldown_str: str) -> bool:
        key = f"{rule_id}:{ad_id}"
        last_triggered = self._cooldowns.get(key)
        if last_triggered is None:
            return False

        cooldown = self._parse_cooldown(cooldown_str)
        return datetime.utcnow() < last_triggered + cooldown

    def mark_triggered(self, rule_id: int, ad_id: int):
        self._cooldowns[f"{rule_id}:{ad_id}"] = datetime.utcnow()

    def _compare(self, actual: float, operator: str, threshold: float) -> bool:
        ops = {
            "gt": lambda a, t: a > t,
            "lt": lambda a, t: a < t,
            "gte": lambda a, t: a >= t,
            "lte": lambda a, t: a <= t,
            "eq": lambda a, t: abs(a - t) < 0.0001,
        }
        return ops.get(operator, lambda a, t: False)(actual, threshold)

    def _parse_cooldown(self, cooldown_str: str) -> timedelta:
        cooldown_str = cooldown_str.strip().lower()
        if cooldown_str.endswith("m"):
            return timedelta(minutes=int(cooldown_str[:-1]))
        if cooldown_str.endswith("h"):
            return timedelta(hours=int(cooldown_str[:-1]))
        return timedelta(hours=1)
