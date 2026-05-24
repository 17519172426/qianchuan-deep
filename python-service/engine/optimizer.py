from dataclasses import dataclass
from typing import Optional


@dataclass
class BudgetRecommendation:
    ad_id: int
    title: str
    description: str
    confidence: float
    action: str
    value: float
    value_type: str


class BudgetOptimizer:
    HIGH_ROI_THRESHOLD = 2.5
    LOW_ROI_THRESHOLD = 0.8
    HIGH_SPEND_RATE = 0.8

    def analyze(
        self, ad_id: int, current_metrics: dict, history_7d: list[dict]
    ) -> Optional[BudgetRecommendation]:
        roi = current_metrics.get("roi", 0)
        cost = current_metrics.get("cost", 0)
        budget = current_metrics.get("budget", cost * 2) if cost > 0 else 1000

        if budget <= 0:
            return None

        spend_rate = cost / budget if budget > 0 else 0
        recent_rois = [h.get("roi", 0) for h in history_7d[-3:]] if history_7d else [roi]
        avg_roi = sum(recent_rois) / len(recent_rois) if recent_rois else roi

        if avg_roi >= self.HIGH_ROI_THRESHOLD and spend_rate >= self.HIGH_SPEND_RATE:
            return BudgetRecommendation(
                ad_id=ad_id,
                title="高 ROI 计划建议追加预算",
                description=f"近 3 天平均 ROI {avg_roi:.2f}，消耗率 {spend_rate:.0%}，建议追加 20% 预算",
                confidence=min(0.9, avg_roi / 4.0),
                action="update_budget",
                value=0.2,
                value_type="percentage",
            )

        if avg_roi < self.LOW_ROI_THRESHOLD and cost > 300:
            return BudgetRecommendation(
                ad_id=ad_id,
                title="低 ROI 计划建议暂停或降预算",
                description=f"近 3 天平均 ROI {avg_roi:.2f}，低于阈值 {self.LOW_ROI_THRESHOLD}，建议降预算 30%",
                confidence=min(0.85, (1 - avg_roi / self.LOW_ROI_THRESHOLD) * 0.8),
                action="update_budget",
                value=-0.3,
                value_type="percentage",
            )

        return None
