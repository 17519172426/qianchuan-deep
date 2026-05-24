import json
from strategy_pb import strategy_pb2, strategy_pb2_grpc
from engine.evaluator import RuleEvaluator
from engine.actions import ActionResolver


class StrategyServicer(strategy_pb2_grpc.StrategyServiceServicer):
    def __init__(self):
        self.evaluator = RuleEvaluator()
        self.resolver = ActionResolver()

    async def EvaluateRules(self, request, context):
        response = strategy_pb2.EvaluateResponse()
        for rule in request.rules:
            cond_json = rule.condition_json
            action_json = rule.action_json
            rule_id = rule.id
            cooldown = rule.cooldown

            for ad in request.ads:
                ad_context = {
                    "ad_id": ad.ad_id,
                    "cost": ad.cost,
                    "roi": ad.roi,
                    "ctr": ad.ctr,
                    "conversions": ad.conversions,
                    "impressions": ad.impressions,
                }
                cond_result = self.evaluator.evaluate_condition(cond_json, ad_context, rule_id)
                if not cond_result.matched:
                    continue
                if self.evaluator.is_in_cooldown(rule_id, ad.ad_id, cooldown):
                    continue

                action_result = self.resolver.resolve(action_json, ad_context, rule_id)
                if action_result is not None:
                    action = strategy_pb2.RuleAction(
                        rule_id=rule_id,
                        ad_id=ad.ad_id,
                        action_type=action_result.action_type,
                        value=action_result.value,
                        value_type=action_result.value_type,
                        reason=f"{cond_result.description}; {action_result.reason}"
                    )
                    response.actions.append(action)
                    self.evaluator.mark_triggered(rule_id, ad.ad_id)
        return response

    async def TestRule(self, request, context):
        response = strategy_pb2.TestRuleResponse()
        rule = request.rule
        cond_json = rule.condition_json
        action_json = rule.action_json

        for ad in request.ads:
            ad_context = {
                "ad_id": ad.ad_id,
                "cost": ad.cost,
                "roi": ad.roi,
                "ctr": ad.ctr,
                "conversions": ad.conversions,
                "impressions": ad.impressions,
            }
            cond_result = self.evaluator.evaluate_condition(cond_json, ad_context, rule.id)
            if cond_result.matched:
                response.would_trigger = True
                response.trigger_reason = cond_result.description
                action_result = self.resolver.resolve(action_json, ad_context, rule.id)
                if action_result is not None:
                    action = strategy_pb2.RuleAction(
                        rule_id=rule.id,
                        ad_id=ad.ad_id,
                        action_type=action_result.action_type,
                        value=action_result.value,
                        value_type=action_result.value_type,
                        reason=action_result.reason
                    )
                    response.actions.append(action)
                break
        return response
