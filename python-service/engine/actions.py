import json
from typing import Optional
from models import ActionResult

SUPPORTED_ACTIONS = {
    "update_budget", "update_roi_goal", "pause_ad",
    "resume_ad", "raise_ad", "notify"
}

class ActionResolver:
    def resolve(self, action_json: str, ad_context: dict, rule_id: int) -> Optional[ActionResult]:
        action = json.loads(action_json) if isinstance(action_json, str) else action_json
        action_type = action.get("type", "")
        if action_type not in SUPPORTED_ACTIONS:
            return None

        value = action.get("value", 0)
        value_type = action.get("value_type", "absolute")
        reason = f"rule_{rule_id}: {action_type} triggered"

        if value_type == "percentage":
            # For update_budget: apply percentage to current cost (via local context, not implemented here)
            pass

        return ActionResult(
            action_type=action_type,
            value=value,
            value_type=value_type,
            target_ad_id=ad_context.get("ad_id", 0),
            rule_id=rule_id,
            reason=reason
        )
