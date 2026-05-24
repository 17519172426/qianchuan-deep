import json
from strategy_pb import strategy_pb2, strategy_pb2_grpc
from engine.evaluator import RuleEvaluator
from engine.actions import ActionResolver
from engine.recommender import RecommendationEngine


class StrategyServicer(strategy_pb2_grpc.StrategyServiceServicer):
    def __init__(self):
        self.evaluator = RuleEvaluator()
        self.resolver = ActionResolver()
        self.recommender = RecommendationEngine()

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
                if not self.evaluator.try_trigger(rule_id, ad.ad_id, cooldown):
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

    async def DetectAnomalies(self, request, context):
        response = strategy_pb2.AnomalyResponse()
        current = []
        for m in request.current:
            current.append({
                "ad_id": m.ad_id, "cost": m.cost, "roi": m.roi,
                "ctr": m.ctr, "conversions": m.conversions,
                "impressions": m.impressions, "cpa": m.cpa,
            })
        history = []
        for w in request.history:
            window = []
            for m in w.hourly:
                window.append({
                    "ad_id": m.ad_id, "cost": m.cost, "roi": m.roi,
                    "ctr": m.ctr, "conversions": m.conversions,
                    "impressions": m.impressions, "cpa": m.cpa,
                })
            history.append(window)

        anomalies = self.recommender.detect_anomalies(current, history)
        for a in anomalies:
            response.anomalies.append(strategy_pb2.Anomaly(
                ad_id=a["ad_id"], metric_name=a["metric_name"],
                current_value=a["current_value"], mean_value=a["mean_value"],
                std_value=a["std_value"], z_score=a["z_score"],
                severity=a["severity"],
            ))
        return response

    async def PredictROI(self, request, context):
        response = strategy_pb2.PredictResponse()
        history = []
        for m in request.history_7d:
            history.append({
                "ad_id": m.ad_id, "cost": m.cost, "roi": m.roi,
                "ctr": m.ctr, "conversions": m.conversions,
                "impressions": m.impressions, "cpa": m.cpa,
            })
        result = self.recommender.predict_roi(request.ad_id, history)
        if result:
            response.predictions.append(strategy_pb2.ROIPrediction(
                ad_id=result["ad_id"],
                predicted_roi_24h=result["predicted_roi_24h"],
                confidence=result["confidence"],
                trend=result["trend"],
            ))
        return response

    async def GenerateRecommendations(self, request, context):
        response = strategy_pb2.RecResponse()
        current = []
        for m in request.current_metrics:
            current.append({
                "ad_id": m.ad_id, "cost": m.cost, "roi": m.roi,
                "ctr": m.ctr, "conversions": m.conversions,
                "impressions": m.impressions, "cpa": m.cpa,
            })
        history = []
        for w in request.history_7d:
            window = []
            for m in w.hourly:
                window.append({
                    "ad_id": m.ad_id, "cost": m.cost, "roi": m.roi,
                    "ctr": m.ctr, "conversions": m.conversions,
                    "impressions": m.impressions, "cpa": m.cpa,
                })
            history.append(window)

        recs = self.recommender.generate_recommendations(
            list(request.ad_ids), current, history)
        for r in recs:
            response.recommendations.append(strategy_pb2.Recommendation(
                ad_id=r["ad_id"], type=r["type"], title=r["title"],
                description=r["description"], confidence=r["confidence"],
                suggested_action_json=r["suggested_action_json"],
            ))
        return response
