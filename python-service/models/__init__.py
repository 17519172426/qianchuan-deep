from pydantic import BaseModel
from typing import Optional


class ConditionEvalResult(BaseModel):
    matched: bool
    metric_name: str
    current_value: float
    threshold: float
    operator: str
    description: str


class ActionResult(BaseModel):
    action_type: str
    value: float
    value_type: str
    target_ad_id: int
    rule_id: int
    reason: str
