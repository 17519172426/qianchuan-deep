from engine.creative_analyzer import CreativeAnalyzer


class TestCreativeAnalyzer:
    def test_low_ctr_creative_flagged(self):
        analyzer = CreativeAnalyzer()
        creatives = [
            {"id": 1, "name": "素材A", "ctr": 0.03, "conversions": 10},
            {"id": 2, "name": "素材B", "ctr": 0.005, "conversions": 2},
            {"id": 3, "name": "素材C", "ctr": 0.025, "conversions": 8},
        ]
        results = analyzer.analyze(1, creatives)
        assert len(results) >= 1
        flagged_ids = [r.creative_id for r in results]
        assert 2 in flagged_ids

    def test_single_creative_no_analysis(self):
        analyzer = CreativeAnalyzer()
        results = analyzer.analyze(
            1, [{"id": 1, "name": "A", "ctr": 0.03, "conversions": 5}]
        )
        assert len(results) == 0

    def test_all_performing_well_no_flags(self):
        analyzer = CreativeAnalyzer()
        creatives = [
            {"id": 1, "name": "A", "ctr": 0.04, "conversions": 15},
            {"id": 2, "name": "B", "ctr": 0.035, "conversions": 12},
        ]
        results = analyzer.analyze(1, creatives)
        assert len(results) == 0
