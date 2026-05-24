from dataclasses import dataclass


@dataclass
class CreativeInsight:
    ad_id: int
    creative_id: int
    creative_name: str
    title: str
    description: str
    confidence: float
    is_low_performing: bool


class CreativeAnalyzer:
    LOW_CTR_THRESHOLD = 0.01

    def analyze(self, ad_id: int, creatives: list[dict]) -> list[CreativeInsight]:
        if len(creatives) < 2:
            return []

        ctrs = [c.get("ctr", 0) for c in creatives]
        conversions_list = [c.get("conversions", 0) for c in creatives]
        avg_ctr = sum(ctrs) / len(ctrs) if ctrs else 0
        avg_conversions = (
            sum(conversions_list) / len(conversions_list) if conversions_list else 0
        )

        results = []
        for c in creatives:
            ctr = c.get("ctr", 0)
            conversions = c.get("conversions", 0)
            is_low = ctr < self.LOW_CTR_THRESHOLD or (
                avg_ctr > 0 and ctr < avg_ctr * 0.5
            )

            if is_low:
                results.append(
                    CreativeInsight(
                        ad_id=ad_id,
                        creative_id=c.get("id", 0),
                        creative_name=c.get("name", "未知素材"),
                        title=f"素材 '{c.get('name', '未知')}' 表现不佳",
                        description=f"CTR {ctr:.4f}（平均 {avg_ctr:.4f}），转化 {conversions}（平均 {avg_conversions:.1f}）",
                        confidence=0.75 if ctr < self.LOW_CTR_THRESHOLD else 0.6,
                        is_low_performing=True,
                    )
                )

        return results
