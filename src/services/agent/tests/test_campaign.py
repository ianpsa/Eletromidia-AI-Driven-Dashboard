"""Tests for pure functions in core.tools.campaign."""

from __future__ import annotations

from unittest.mock import patch

from core.tools.campaign import (
    _build_sql,
    _class_columns,
    _overlapping_age_columns,
)

# ── _overlapping_age_columns ──────────────────────────────────────────


class TestOverlappingAgeColumns:
    def test_exact_bucket(self):
        assert _overlapping_age_columns(18, 19) == ["x18_19"]

    def test_all_buckets(self):
        result = _overlapping_age_columns(18, 120)
        assert len(result) == 8
        assert result[0] == "x18_19"
        assert result[-1] == "x80_plus"

    def test_partial_range(self):
        assert _overlapping_age_columns(25, 45) == ["x20_29", "x30_39", "x40_49"]

    def test_boundary_overlap(self):
        assert _overlapping_age_columns(19, 20) == ["x18_19", "x20_29"]

    def test_min_greater_than_max(self):
        assert _overlapping_age_columns(50, 10) == []

    def test_range_outside_buckets(self):
        assert _overlapping_age_columns(0, 10) == []


# ── _class_columns ────────────────────────────────────────────────────


class TestClassColumns:
    def test_valid_classes(self):
        assert _class_columns(["A", "B1"]) == ["a_class", "b1_class"]

    def test_all_classes(self):
        result = _class_columns(["A", "B1", "B2", "C1", "C2", "DE"])
        assert result == [
            "a_class",
            "b1_class",
            "b2_class",
            "c1_class",
            "c2_class",
            "de_class",
        ]

    def test_invalid_class(self):
        assert _class_columns(["X"]) == []

    def test_mix_valid_invalid(self):
        assert _class_columns(["A", "Z", "DE"]) == ["a_class", "de_class"]

    def test_case_insensitive(self):
        assert _class_columns(["a", "b1"]) == ["a_class", "b1_class"]


# ── _build_sql ────────────────────────────────────────────────────────

_DS = "proj.ds"


def _call_build_sql(**overrides):
    defaults = dict(
        gender=None,
        age_min=None,
        age_max=None,
        classes=None,
        city=None,
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
        assert "1.0" in sql
        assert any(p.name == "result_limit" and p.value == 5 for p in params)

    def test_gender_female(self):
        sql, _ = _call_build_sql(gender="female")
        assert "gnd.feminine" in sql
        assert "gnd.masculine" not in sql

    def test_gender_male(self):
        sql, _ = _call_build_sql(gender="male")
        assert "gnd.masculine" in sql
        assert "gnd.feminine" not in sql

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
        assert "LOWER(g.cidade) = LOWER(@city)" in sql
        city_param = next(p for p in params if p.name == "city")
        assert city_param.value == "São Paulo"

    def test_combined_filters(self):
        sql, params = _call_build_sql(
            latitude=-23.5,
            longitude=-46.6,
            city="São Paulo",
        )
        assert "WHERE" in sql
        assert " AND " in sql
        assert "ST_DISTANCE" in sql
        assert "LOWER(g.cidade)" in sql
