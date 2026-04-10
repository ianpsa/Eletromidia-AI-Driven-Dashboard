"""Tests for pure functions in core.tools.campaign."""

from __future__ import annotations

from unittest.mock import patch

from core.tools.campaign import _build_sql, _class_columns, _overlapping_age_columns

# ── _overlapping_age_columns ──────────────────────────────────────────


class TestOverlappingAgeColumns:
    def test_exact_bucket(self):
        assert _overlapping_age_columns(18, 19) == ["age_18_19_count"]

    def test_all_buckets(self):
        result = _overlapping_age_columns(18, 120)
        assert len(result) == 8
        assert result[0] == "age_18_19_count"
        assert result[-1] == "age_80_plus_count"

    def test_partial_range(self):
        expected = ["age_20_29_count", "age_30_39_count", "age_40_49_count"]
        assert _overlapping_age_columns(25, 45) == expected

    def test_boundary_overlap(self):
        expected = ["age_18_19_count", "age_20_29_count"]
        assert _overlapping_age_columns(19, 20) == expected

    def test_min_greater_than_max(self):
        assert _overlapping_age_columns(50, 10) == []

    def test_range_outside_buckets(self):
        assert _overlapping_age_columns(0, 10) == []


# ── _class_columns ────────────────────────────────────────────────────


class TestClassColumns:
    def test_valid_classes(self):
        assert _class_columns(["A", "B1"]) == ["class_a_count", "class_b1_count"]

    def test_all_classes(self):
        result = _class_columns(["A", "B1", "B2", "C1", "C2", "DE"])
        assert result == [
            "class_a_count",
            "class_b1_count",
            "class_b2_count",
            "class_c1_count",
            "class_c2_count",
            "class_de_count",
        ]

    def test_invalid_class(self):
        assert _class_columns(["X"]) == []

    def test_mix_valid_invalid(self):
        assert _class_columns(["A", "Z", "DE"]) == ["class_a_count", "class_de_count"]

    def test_case_insensitive(self):
        assert _class_columns(["a", "b1"]) == ["class_a_count", "class_b1_count"]


# ── _build_sql ────────────────────────────────────────────────────────

_DS = "proj.ds"


def _call_build_sql(**overrides):
    defaults = dict(
        gender=None,
        age_min=None,
        age_max=None,
        classes=None,
        city=None,
        endereco=None,
        latitude=None,
        longitude=None,
        radius_km=2.0,
        limit=5,
    )
    defaults.update(overrides)
    with patch("core.tools.campaign.get_dataset_ref", return_value=_DS):
        return _build_sql(**defaults)


class TestBuildSql:
    def test_no_filters(self):
        sql, params = _call_build_sql()
        assert "WHERE" not in sql
        assert "SUM(s.uniques)" in sql
        assert any(p.name == "result_limit" and p.value == 5 for p in params)

    def test_gender_female(self):
        sql, _ = _call_build_sql(gender="female")
        assert "s.feminine_count" in sql
        assert "s.masculine_count" not in sql

    def test_gender_male(self):
        sql, _ = _call_build_sql(gender="male")
        assert "s.masculine_count" in sql
        assert "s.feminine_count" not in sql

    def test_geo_filter(self):
        sql, params = _call_build_sql(latitude=-23.5, longitude=-46.6, radius_km=3.0)
        assert "ST_DISTANCE" in sql
        assert "WHERE" in sql
        param_names = {p.name for p in params}
        assert {"lng", "lat", "radius_m"} <= param_names
        radius_param = next(p for p in params if p.name == "radius_m")
        assert radius_param.value == 3000.0

    def test_city_filter(self):
        sql, params = _call_build_sql(city="São Paulo")
        assert "LOWER(s.cidade) = LOWER(@city)" in sql
        city_param = next(p for p in params if p.name == "city")
        assert city_param.value == "São Paulo"

    def test_endereco_filter(self):
        sql, params = _call_build_sql(endereco="FARIA LIMA")
        assert "WHERE" in sql
        assert "LOWER(s.endereco) LIKE" in sql
        endereco_param = next(p for p in params if p.name == "endereco")
        assert endereco_param.value == "FARIA LIMA"

    def test_age_filter(self):
        sql, _ = _call_build_sql(age_min=25, age_max=45)
        assert "age_20_29_count" in sql
        assert "age_30_39_count" in sql
        assert "age_40_49_count" in sql
        assert "age_18_19_count" not in sql

    def test_class_filter(self):
        sql, _ = _call_build_sql(classes=["A", "B1"])
        assert "class_a_count" in sql
        assert "class_b1_count" in sql
        assert "class_c1_count" not in sql

    def test_combined_filters(self):
        sql, params = _call_build_sql(
            latitude=-23.5,
            longitude=-46.6,
            city="São Paulo",
        )
        assert "WHERE" in sql
        assert " AND " in sql
        assert "ST_DISTANCE" in sql
        assert "LOWER(s.cidade)" in sql

    def test_gender_none_uses_uniques(self):
        sql, _ = _call_build_sql(gender=None)
        assert "SUM(s.uniques)" in sql
        assert "s.feminine_count" not in sql
        assert "s.masculine_count" not in sql

    def test_invalid_classes_fallback_to_uniques(self):
        sql, _ = _call_build_sql(classes=["X", "Y"])
        # No valid class columns → fallback to SUM(s.uniques)
        assert "class_" not in sql

    def test_empty_classes_uses_uniques(self):
        sql, _ = _call_build_sql(classes=[])
        assert "class_" not in sql

    def test_age_min_only_none_uses_uniques(self):
        # Only age_max set, age_min is None → no age filter applied
        sql, _ = _call_build_sql(age_min=None, age_max=30)
        # Both must be non-None to apply age filter
        assert "age_" not in sql

    def test_age_max_only_none_uses_uniques(self):
        sql, _ = _call_build_sql(age_min=25, age_max=None)
        assert "age_" not in sql

    def test_result_limit_param(self):
        _, params = _call_build_sql(limit=25)
        limit_param = next(p for p in params if p.name == "result_limit")
        assert limit_param.value == 25

    def test_geo_adds_coordinate_validity_checks(self):
        sql, _ = _call_build_sql(latitude=-23.5, longitude=-46.6, radius_km=1.0)
        assert "s.latitude IS NOT NULL" in sql
        assert "BETWEEN -90 AND 90" in sql
        assert "BETWEEN -180 AND 180" in sql

    def test_dataset_ref_used_in_from(self):
        sql, _ = _call_build_sql()
        assert f"`{_DS}.vw_geodata_enriched`" in sql

    def test_all_demographic_filters_combined(self):
        sql, params = _call_build_sql(
            gender="female",
            age_min=25,
            age_max=35,
            classes=["A", "B1"],
        )
        assert "s.feminine_count" in sql
        assert "age_20_29_count" in sql
        assert "age_30_39_count" in sql
        assert "class_a_count" in sql
        assert "class_b1_count" in sql


# ── analyze_campaign (tool function) ─────────────────────────────────


def _make_row(**kwargs):
    defaults = {
        "location_id": "loc1",
        "endereco_ref": "Av. Paulista, 1000 — São Paulo",
        "cidade": "São Paulo",
        "latitude": -23.5505,
        "longitude": -46.6333,
        "total_flow": 5000,
        "affinity": 85.5,
        "target_audience": 3000,
    }
    defaults.update(kwargs)
    return defaults


class TestAnalyzeCampaign:
    from core.tools.campaign import analyze_campaign

    def test_returns_ranked_results(self):
        from core.tools.campaign import analyze_campaign

        rows = [_make_row(), _make_row(location_id="loc2", affinity=70.0)]
        with (
            patch("core.tools.campaign.get_dataset_ref", return_value=_DS),
            patch("core.tools.campaign.run_query_with_params", return_value=rows),
        ):
            result = analyze_campaign.invoke({"gender": "female", "city": "São Paulo"})
        assert "2 pontos" in result
        assert "Ranking:" in result
        assert "85.5%" in result

    def test_no_results_returns_message(self):
        from core.tools.campaign import analyze_campaign

        with (
            patch("core.tools.campaign.get_dataset_ref", return_value=_DS),
            patch("core.tools.campaign.run_query_with_params", return_value=[]),
        ):
            result = analyze_campaign.invoke({})
        assert "Nenhum ponto" in result

    def test_exception_returns_error_message(self):
        from core.tools.campaign import analyze_campaign

        with (
            patch("core.tools.campaign.get_dataset_ref", return_value=_DS),
            patch(
                "core.tools.campaign.run_query_with_params",
                side_effect=Exception("connection error"),
            ),
        ):
            result = analyze_campaign.invoke({"city": "São Paulo"})
        assert "Erro" in result

    def test_filters_description_included(self):
        from core.tools.campaign import analyze_campaign

        rows = [_make_row()]
        with (
            patch("core.tools.campaign.get_dataset_ref", return_value=_DS),
            patch("core.tools.campaign.run_query_with_params", return_value=rows),
        ):
            result = analyze_campaign.invoke(
                {"gender": "female", "age_min": 25, "age_max": 35, "city": "São Paulo"}
            )
        assert "gênero=female" in result
        assert "idade=25-35" in result
        assert "cidade=São Paulo" in result

    def test_coords_in_output(self):
        from core.tools.campaign import analyze_campaign

        rows = [_make_row(latitude=-23.5505, longitude=-46.6333)]
        with (
            patch("core.tools.campaign.get_dataset_ref", return_value=_DS),
            patch("core.tools.campaign.run_query_with_params", return_value=rows),
        ):
            result = analyze_campaign.invoke({})
        assert "coords:" in result

    def test_no_coords_when_none(self):
        from core.tools.campaign import analyze_campaign

        rows = [_make_row(latitude=None, longitude=None)]
        with (
            patch("core.tools.campaign.get_dataset_ref", return_value=_DS),
            patch("core.tools.campaign.run_query_with_params", return_value=rows),
        ):
            result = analyze_campaign.invoke({})
        assert "coords:" not in result
