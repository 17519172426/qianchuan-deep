import math
from dataclasses import dataclass

METRIC_THRESHOLDS = {
    "roi": 2.0,
    "cpa": 2.0,
    "cost": 2.5,
    "ctr": 2.0,
}

@dataclass
class AnomalyResult:
    ad_id: int
    metric_name: str
    current_value: float
    mean_value: float
    std_value: float
    z_score: float
    severity: str

class AnomalyDetector:
    def detect(self, current_metrics: list[dict], history_windows: list[list[dict]]) -> list[AnomalyResult]:
        results = []
        for metric_entry in current_metrics:
            ad_id = metric_entry.get("ad_id", 0)
            history = self._get_history_for_ad(ad_id, history_windows)
            if not history or len(history) < 3:
                continue

            for metric_name in ["roi", "cpa", "cost", "ctr"]:
                current_val = metric_entry.get(metric_name, 0)
                historical_vals = [h.get(metric_name, 0) for h in history if metric_name in h]
                if len(historical_vals) < 3:
                    continue

                mean_val = sum(historical_vals) / len(historical_vals)
                variance = sum((x - mean_val) ** 2 for x in historical_vals) / len(historical_vals)
                std_val = math.sqrt(variance) if variance > 0 else 0.001

                z_score = (current_val - mean_val) / std_val if std_val > 0 else 0
                threshold = METRIC_THRESHOLDS.get(metric_name, 2.0)

                if abs(z_score) >= threshold:
                    severity = "high" if abs(z_score) >= threshold * 1.5 else "medium"
                    results.append(AnomalyResult(
                        ad_id=ad_id,
                        metric_name=metric_name,
                        current_value=current_val,
                        mean_value=round(mean_val, 4),
                        std_value=round(std_val, 4),
                        z_score=round(z_score, 2),
                        severity=severity,
                    ))
        return results

    def _get_history_for_ad(self, ad_id: int, history_windows: list[list[dict]]) -> list[dict]:
        all_points = []
        for window in history_windows:
            for entry in window:
                if entry.get("ad_id") == ad_id:
                    all_points.append(entry)
        return all_points
