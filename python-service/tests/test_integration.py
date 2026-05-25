import pytest
from server.grpc_server import StrategyServicer
from strategy_pb import strategy_pb2


@pytest.fixture
def servicer():
    return StrategyServicer()


class TestGRPCServer:
    def test_evaluate_rules_triggers_pause(self, servicer):
        """ROI below threshold should trigger pause action."""
        import asyncio
        rules = [strategy_pb2.RuleDef(
            id=1, name="低ROI暂停", account_id=1,
            condition_json='{"metric":"roi","operator":"lt","threshold":1.5}',
            action_json='{"type":"pause_ad"}',
            cooldown="30m"
        )]
        ads = [strategy_pb2.AdContext(ad_id=1, roi=1.2)]
        req = strategy_pb2.EvaluateRequest(rules=rules, ads=ads)
        resp = asyncio.new_event_loop().run_until_complete(servicer.EvaluateRules(req, None))
        assert len(resp.actions) == 1
        assert resp.actions[0].action_type == "pause_ad"
        assert resp.actions[0].ad_id == 1

    def test_evaluate_rules_no_trigger_when_above_threshold(self, servicer):
        """ROI above threshold should not trigger."""
        import asyncio
        rules = [strategy_pb2.RuleDef(
            id=1, name="高ROI OK", account_id=1,
            condition_json='{"metric":"roi","operator":"lt","threshold":1.0}',
            action_json='{"type":"pause_ad"}',
            cooldown="30m"
        )]
        ads = [strategy_pb2.AdContext(ad_id=1, roi=2.5)]
        req = strategy_pb2.EvaluateRequest(rules=rules, ads=ads)
        resp = asyncio.new_event_loop().run_until_complete(servicer.EvaluateRules(req, None))
        assert len(resp.actions) == 0

    def test_multiple_ads_one_triggers(self, servicer):
        """Only ads matching condition should get actions."""
        import asyncio
        rules = [strategy_pb2.RuleDef(
            id=2, name="高消耗检查", account_id=1,
            condition_json='{"metric":"cost","operator":"gt","threshold":500}',
            action_json='{"type":"update_budget","value":-0.2,"value_type":"percentage"}',
            cooldown="15m"
        )]
        ads = [
            strategy_pb2.AdContext(ad_id=1, cost=300),
            strategy_pb2.AdContext(ad_id=2, cost=800),
        ]
        req = strategy_pb2.EvaluateRequest(rules=rules, ads=ads)
        resp = asyncio.new_event_loop().run_until_complete(servicer.EvaluateRules(req, None))
        assert len(resp.actions) == 1
        assert resp.actions[0].ad_id == 2

    def test_cooldown_prevents_second_trigger(self, servicer):
        """After marking triggered, cooldown should block re-trigger."""
        import asyncio
        servicer.evaluator.mark_triggered(1, 10)
        rules = [strategy_pb2.RuleDef(
            id=1, name="test", account_id=1,
            condition_json='{"metric":"roi","operator":"lt","threshold":2.0}',
            action_json='{"type":"pause_ad"}',
            cooldown="30m"
        )]
        ads = [strategy_pb2.AdContext(ad_id=10, roi=0.5)]
        req = strategy_pb2.EvaluateRequest(rules=rules, ads=ads)
        resp = asyncio.new_event_loop().run_until_complete(servicer.EvaluateRules(req, None))
        assert len(resp.actions) == 0
