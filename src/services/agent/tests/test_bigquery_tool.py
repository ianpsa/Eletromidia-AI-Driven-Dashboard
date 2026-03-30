"""Tests for _validate_sql in core.tools.bigquery."""

from __future__ import annotations

from core.tools.bigquery import _SELECT_ONLY_ERROR, _validate_sql


class TestValidateSql:
    def test_simple_select(self):
        assert _validate_sql("SELECT * FROM table") is None

    def test_with_cte(self):
        assert _validate_sql("WITH cte AS (SELECT 1) SELECT * FROM cte") is None

    def test_insert_rejected(self):
        assert _validate_sql("INSERT INTO t VALUES (1)") == _SELECT_ONLY_ERROR

    def test_drop_rejected(self):
        assert _validate_sql("DROP TABLE users") == _SELECT_ONLY_ERROR

    def test_semicolon_rejected(self):
        assert _validate_sql("SELECT 1; DROP TABLE users") == _SELECT_ONLY_ERROR

    def test_empty_string_rejected(self):
        assert _validate_sql("") == _SELECT_ONLY_ERROR

    def test_whitespace_only_rejected(self):
        assert _validate_sql("   ") == _SELECT_ONLY_ERROR

    def test_select_with_dangerous_keyword_in_string(self):
        sql = "SELECT * FROM t WHERE name = 'DELETE'"
        assert _validate_sql(sql) == _SELECT_ONLY_ERROR

    def test_lowercase_select(self):
        assert _validate_sql("select id from t") is None
