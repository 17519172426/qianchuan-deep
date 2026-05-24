from engine.anomaly import AnomalyDetector

class TestAnomalyDetector:
    def test_detect_roi_drop(self):
        detector = AnomalyDetector()
        current = [{"ad_id": 1, "roi": 0.5, "cpa": 50, "cost": 500, "ctr": 0.03}]
        history = [[
            {"ad_id": 1, "roi": 2.0, "cpa": 45, "cost": 480, "ctr": 0.032},
            {"ad_id": 1, "roi": 2.1, "cpa": 48, "cost": 490, "ctr": 0.031},
            {"ad_id": 1, "roi": 1.9, "cpa": 47, "cost": 470, "ctr": 0.033},
            {"ad_id": 1, "roi": 2.2, "cpa": 44, "cost": 500, "ctr": 0.030},
            {"ad_id": 1, "roi": 2.0, "cpa": 46, "cost": 485, "ctr": 0.031},
            {"ad_id": 1, "roi": 1.8, "cpa": 49, "cost": 495, "ctr": 0.029},
            {"ad_id": 1, "roi": 2.1, "cpa": 43, "cost": 475, "ctr": 0.032},
        ]]
        results = detector.detect(current, history)
        assert len(results) >= 1
        roi_anomaly = [r for r in results if r.metric_name == "roi"]
        assert len(roi_anomaly) > 0
        assert roi_anomaly[0].severity in ("medium", "high")

    def test_no_anomaly_when_stable(self):
        detector = AnomalyDetector()
        current = [{"ad_id": 1, "roi": 2.0, "cpa": 45, "cost": 500, "ctr": 0.03}]
        history = [[
            {"ad_id": 1, "roi": 2.0, "cpa": 46, "cost": 495, "ctr": 0.031},
            {"ad_id": 1, "roi": 2.1, "cpa": 44, "cost": 505, "ctr": 0.029},
            {"ad_id": 1, "roi": 1.9, "cpa": 45, "cost": 500, "ctr": 0.030},
        ]]
        results = detector.detect(current, history)
        assert len(results) == 0

    def test_insufficient_history_skipped(self):
        detector = AnomalyDetector()
        current = [{"ad_id": 1, "roi": 0.1, "cpa": 999, "cost": 1, "ctr": 0}]
        history = [[{"ad_id": 1, "roi": 2.0}]]
        results = detector.detect(current, history)
        assert len(results) == 0
