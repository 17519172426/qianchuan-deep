from dataclasses import dataclass
from typing import Optional


@dataclass
class PredictResult:
    ad_id: int
    predicted_roi_24h: float
    confidence: float
    trend: str


class ROIPredictor:
    def predict(self, ad_id: int, history_7d: list[dict]) -> Optional[PredictResult]:
        roi_values = [h.get("roi", 0) for h in history_7d if "roi" in h]
        if len(roi_values) < 3:
            return None

        n = len(roi_values)
        x_mean = (n - 1) / 2.0
        y_mean = sum(roi_values) / n

        num = sum((i - x_mean) * (roi_values[i] - y_mean) for i in range(n))
        den = sum((i - x_mean) ** 2 for i in range(n))

        if den == 0:
            slope = 0
        else:
            slope = num / den

        intercept = y_mean - slope * x_mean
        predicted = intercept + slope * n

        if predicted < 0:
            predicted = 0.01

        ss_res = sum((roi_values[i] - (intercept + slope * i)) ** 2 for i in range(n))
        ss_tot = sum((v - y_mean) ** 2 for v in roi_values)
        r_squared = 1 - (ss_res / ss_tot) if ss_tot > 0 else 0
        confidence = min(max(r_squared, 0), 1)

        if slope > 0.05:
            trend = "up"
        elif slope < -0.05:
            trend = "down"
        else:
            trend = "stable"

        return PredictResult(
            ad_id=ad_id,
            predicted_roi_24h=round(predicted, 4),
            confidence=round(confidence, 2),
            trend=trend,
        )
