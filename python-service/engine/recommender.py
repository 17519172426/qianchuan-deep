from typing import Optional

from engine.anomaly import AnomalyDetector
from engine.predictor import ROIPredictor
from engine.optimizer import BudgetOptimizer
from engine.creative_analyzer import CreativeAnalyzer

class RecommendationEngine:
    def __init__(self):
        self.anomaly_detector = AnomalyDetector()
        self.roi_predictor = ROIPredictor()
        self.budget_optimizer = BudgetOptimizer()
        self.creative_analyzer = CreativeAnalyzer()

    def detect_anomalies(self, current_metrics: list[dict], history_windows: list[list[dict]]) -> list[dict]:
        results = self.anomaly_detector.detect(current_metrics, history_windows)
        return [
            {
                "ad_id": r.ad_id,
                "metric_name": r.metric_name,
                "current_value": r.current_value,
                "mean_value": r.mean_value,
                "std_value": r.std_value,
                "z_score": r.z_score,
                "severity": r.severity,
            }
            for r in results
        ]

    def predict_roi(self, ad_id: int, history_7d: list[dict]) -> Optional[dict]:
        result = self.roi_predictor.predict(ad_id, history_7d)
        if result is None:
            return None
        return {
            "ad_id": result.ad_id,
            "predicted_roi_24h": result.predicted_roi_24h,
            "confidence": result.confidence,
            "trend": result.trend,
        }

    def generate_recommendations(self, ad_ids: list[int], current_metrics: list[dict], history_7d: list[list[dict]]) -> list[dict]:
        results = []
        metric_by_ad = {m.get("ad_id", 0): m for m in current_metrics}
        history_by_ad = {}
        for i, window in enumerate(history_7d):
            for entry in window:
                ad_id = entry.get("ad_id", 0)
                if ad_id not in history_by_ad:
                    history_by_ad[ad_id] = []
                history_by_ad[ad_id].append(entry)

        for ad_id in ad_ids:
            metrics = metric_by_ad.get(ad_id, {})
            history = history_by_ad.get(ad_id, [])

            budget_rec = self.budget_optimizer.analyze(ad_id, metrics, history)
            if budget_rec:
                results.append({
                    "ad_id": ad_id,
                    "type": "budget_opt",
                    "title": budget_rec.title,
                    "description": budget_rec.description,
                    "confidence": budget_rec.confidence,
                    "suggested_action_json": f'{{"type":"{budget_rec.action}","value":{budget_rec.value},"value_type":"{budget_rec.value_type}"}}',
                })

            anomaly_results = self.anomaly_detector.detect([metrics], [history])
            for a in anomaly_results:
                results.append({
                    "ad_id": ad_id,
                    "type": "anomaly",
                    "title": f"{a.metric_name.upper()} 异常检测",
                    "description": f"当前 {a.metric_name}={a.current_value}，历史均值 {a.mean_value}±{a.std_value}，Z-score={a.z_score}",
                    "confidence": min(0.95, abs(a.z_score) / 4.0),
                    "suggested_action_json": '{"type":"notify"}',
                })

        return results
