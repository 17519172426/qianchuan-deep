import pytest
from engine.evaluator import RuleEvaluator
from engine.actions import ActionResolver

class TestRuleEvaluator:
    def test_roi_lt_triggers(self):
        evaluator = RuleEvaluator()
        cond = '{"metric":"roi","operator":"lt","threshold":1.5}'
        result = evaluator.evaluate_condition(cond, {"roi": 1.2}, 1)
        assert result.matched is True

    def test_roi_gt_not_triggered(self):
        evaluator = RuleEvaluator()
        cond = '{"metric":"roi","operator":"gt","threshold":3.0}'
        result = evaluator.evaluate_condition(cond, {"roi": 2.5}, 1)
        assert result.matched is False

    def test_cost_gte_triggers(self):
        evaluator = RuleEvaluator()
        cond = '{"metric":"cost","operator":"gte","threshold":1000}'
        result = evaluator.evaluate_condition(cond, {"cost": 1000}, 1)
        assert result.matched is True

    def test_ctr_lt_triggers(self):
        evaluator = RuleEvaluator()
        cond = '{"metric":"ctr","operator":"lt","threshold":0.05}'
        result = evaluator.evaluate_condition(cond, {"ctr": 0.03}, 1)
        assert result.matched is True

    def test_unsupported_metric(self):
        evaluator = RuleEvaluator()
        result = evaluator.evaluate_condition(
            '{"metric":"unknown","operator":"gt","threshold":10}', {"unknown": 5}, 1
        )
        assert result.matched is False

    def test_unsupported_operator(self):
        evaluator = RuleEvaluator()
        result = evaluator.evaluate_condition(
            '{"metric":"roi","operator":"ne","threshold":1}', {"roi": 1}, 1
        )
        assert result.matched is False

    def test_cooldown_blocks(self):
        evaluator = RuleEvaluator()
        evaluator.mark_triggered(1, 100)
        assert evaluator.is_in_cooldown(1, 100, "30m") is True

    def test_cooldown_expired(self):
        evaluator = RuleEvaluator()
        evaluator.mark_triggered(1, 100)
        evaluator._cooldowns["1:100"] = evaluator._cooldowns["1:100"].replace(hour=0)
        assert evaluator.is_in_cooldown(1, 100, "30m") is False


class TestActionResolver:
    def test_pause_ad(self):
        resolver = ActionResolver()
        result = resolver.resolve('{"type":"pause_ad"}', {"ad_id": 1}, 1)
        assert result is not None
        assert result.action_type == "pause_ad"

    def test_unsupported_action(self):
        resolver = ActionResolver()
        result = resolver.resolve('{"type":"delete_ad"}', {"ad_id": 1}, 1)
        assert result is None

    def test_update_budget_with_percentage(self):
        resolver = ActionResolver()
        result = resolver.resolve(
            '{"type":"update_budget","value":-0.2,"value_type":"percentage"}',
            {"ad_id": 2}, 1
        )
        assert result is not None
        assert result.value == -0.2
        assert result.value_type == "percentage"
