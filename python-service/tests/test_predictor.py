from engine.predictor import ROIPredictor


class TestROIPredictor:
    def test_predict_upward_trend(self):
        predictor = ROIPredictor()
        history = [
            {"roi": 1.0}, {"roi": 1.2}, {"roi": 1.3},
            {"roi": 1.5}, {"roi": 1.6}, {"roi": 1.8}, {"roi": 2.0},
        ]
        result = predictor.predict(1, history)
        assert result is not None
        assert result.trend == "up"
        assert result.predicted_roi_24h > 2.0

    def test_predict_downward_trend(self):
        predictor = ROIPredictor()
        history = [
            {"roi": 2.0}, {"roi": 1.8}, {"roi": 1.6},
            {"roi": 1.5}, {"roi": 1.3}, {"roi": 1.2}, {"roi": 1.0},
        ]
        result = predictor.predict(2, history)
        assert result is not None
        assert result.trend == "down"

    def test_insufficient_data_returns_none(self):
        predictor = ROIPredictor()
        result = predictor.predict(3, [{"roi": 1.0}, {"roi": 1.1}])
        assert result is None
