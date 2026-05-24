from engine.optimizer import BudgetOptimizer


class TestBudgetOptimizer:
    def test_high_roi_high_spend_suggests_increase(self):
        opt = BudgetOptimizer()
        result = opt.analyze(
            1,
            {"roi": 3.0, "cost": 900, "budget": 1000},
            [{"roi": 3.2}, {"roi": 2.8}, {"roi": 3.0}],
        )
        assert result is not None
        assert result.action == "update_budget"
        assert result.value > 0

    def test_low_roi_suggests_decrease(self):
        opt = BudgetOptimizer()
        result = opt.analyze(
            2,
            {"roi": 0.5, "cost": 500, "budget": 1000},
            [{"roi": 0.6}, {"roi": 0.4}, {"roi": 0.5}],
        )
        assert result is not None
        assert result.action == "update_budget"
        assert result.value < 0

    def test_normal_roi_no_suggestion(self):
        opt = BudgetOptimizer()
        result = opt.analyze(
            3,
            {"roi": 1.5, "cost": 500, "budget": 1000},
            [{"roi": 1.6}, {"roi": 1.4}, {"roi": 1.5}],
        )
        assert result is None
