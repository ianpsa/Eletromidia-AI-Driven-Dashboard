"""Tests for _validate_sql and query_bigquery in core.tools.bigquery."""

from __future__ import annotations

from unittest.mock import patch

from core.tools.bigquery import _SELECT_ONLY_ERROR, _validate_sql, query_bigquery


class TestValidateSql:
    def test_simple_select(self):
        assert _validate_sql("SELECT * FROM table") is None

    def test_with_cte(self):
        assert _validate_sql("WITH cte AS (SELECT 1) SELECT * FROM cte") is None

    def test_insert_rejected(self):
        assert _validate_sql("INSERT INTO t VALUES (1)") == _SELECT_ONLY_ERROR

    def test_drop_rejected(self):
        assert _validate_sql("DROP TABLE users") == _SELECT_ONLY_ERROR

    def test_update_rejected(self):
        assert _validate_sql("UPDATE t SET col = 1 WHERE id = 1") == _SELECT_ONLY_ERROR

    def test_delete_rejected(self):
        assert _validate_sql("DELETE FROM t WHERE id = 1") == _SELECT_ONLY_ERROR

    def test_alter_rejected(self):
        assert _validate_sql("ALTER TABLE t ADD COLUMN x INT") == _SELECT_ONLY_ERROR

    def test_truncate_rejected(self):
        assert _validate_sql("TRUNCATE TABLE t") == _SELECT_ONLY_ERROR

    def test_merge_rejected(self):
        sql = "MERGE t USING s ON t.id = s.id WHEN MATCHED THEN UPDATE SET t.x = s.x"
        assert _validate_sql(sql) == _SELECT_ONLY_ERROR

    def test_grant_rejected(self):
        assert _validate_sql("GRANT SELECT ON t TO user") == _SELECT_ONLY_ERROR

    def test_create_rejected(self):
        sql = "CREATE TABLE new_table AS SELECT * FROM t"
        assert _validate_sql(sql) == _SELECT_ONLY_ERROR

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

    def test_select_with_subquery(self):
        sql = "SELECT * FROM (SELECT id FROM t WHERE id > 1) sub"
        assert _validate_sql(sql) is None

    def test_select_with_join(self):
        sql = "SELECT a.id, b.name FROM a JOIN b ON a.id = b.id"
        assert _validate_sql(sql) is None


# ── query_bigquery (tool function) ────────────────────────────────────

_DS = "proj.dataset"


def _run_query_bigquery(sql: str):
    """Helper to invoke query_bigquery with mocked dependencies."""
    with (
        patch("core.tools.bigquery.get_dataset_ref", return_value=_DS),
        patch("core.tools.bigquery.run_query") as mock_run,
    ):
        yield mock_run, query_bigquery.invoke({"sql_query": sql})


class TestQueryBigquery:
    def test_dangerous_sql_returns_error(self):
        with (
            patch("core.tools.bigquery.get_dataset_ref", return_value=_DS),
            patch("core.tools.bigquery.run_query") as mock_run,
        ):
            result = query_bigquery.invoke({"sql_query": "DROP TABLE t"})
        assert result == _SELECT_ONLY_ERROR
        mock_run.assert_not_called()

    def test_table_names_are_qualified(self):
        rows = [{"id": 1, "cidade": "São Paulo"}]
        with (
            patch("core.tools.bigquery.get_dataset_ref", return_value=_DS),
            patch("core.tools.bigquery.run_query", return_value=rows) as mock_run,
        ):
            query_bigquery.invoke({"sql_query": "SELECT * FROM geodata"})
        called_sql = mock_run.call_args[0][0]
        assert f"`{_DS}.geodata`" in called_sql
        assert "FROM geodata" not in called_sql

    def test_limit_added_when_missing(self):
        rows = [{"id": 1}]
        with (
            patch("core.tools.bigquery.get_dataset_ref", return_value=_DS),
            patch("core.tools.bigquery.run_query", return_value=rows) as mock_run,
        ):
            query_bigquery.invoke({"sql_query": "SELECT * FROM enriched_screens"})
        called_sql = mock_run.call_args[0][0]
        assert "LIMIT 100" in called_sql

    def test_existing_limit_not_doubled(self):
        rows = [{"id": 1}]
        with (
            patch("core.tools.bigquery.get_dataset_ref", return_value=_DS),
            patch("core.tools.bigquery.run_query", return_value=rows) as mock_run,
        ):
            query_bigquery.invoke(
                {"sql_query": "SELECT * FROM enriched_screens LIMIT 10"}
            )
        called_sql = mock_run.call_args[0][0]
        assert called_sql.count("LIMIT") == 1

    def test_empty_result_returns_message(self):
        with (
            patch("core.tools.bigquery.get_dataset_ref", return_value=_DS),
            patch("core.tools.bigquery.run_query", return_value=[]),
        ):
            result = query_bigquery.invoke({"sql_query": "SELECT * FROM geodata"})
        assert result == "Nenhum resultado encontrado."

    def test_exception_returns_error_message(self):
        with (
            patch("core.tools.bigquery.get_dataset_ref", return_value=_DS),
            patch(
                "core.tools.bigquery.run_query",
                side_effect=Exception("connection error"),
            ),
        ):
            result = query_bigquery.invoke({"sql_query": "SELECT * FROM geodata"})
        assert "Erro" in result

    def test_results_formatted_as_table(self):
        rows = [
            {"cidade": "São Paulo", "uniques": 1000},
            {"cidade": "Rio de Janeiro", "uniques": 800},
        ]
        with (
            patch("core.tools.bigquery.get_dataset_ref", return_value=_DS),
            patch("core.tools.bigquery.run_query", return_value=rows),
        ):
            result = query_bigquery.invoke(
                {"sql_query": "SELECT * FROM enriched_screens"}
            )
        assert "2 linhas" in result
        assert "cidade" in result
        assert "São Paulo" in result
        assert "Rio de Janeiro" in result

    def test_trailing_semicolon_stripped(self):
        rows = [{"id": 1}]
        with (
            patch("core.tools.bigquery.get_dataset_ref", return_value=_DS),
            patch("core.tools.bigquery.run_query", return_value=rows) as mock_run,
        ):
            query_bigquery.invoke({"sql_query": "SELECT * FROM geodata;"})
        called_sql = mock_run.call_args[0][0]
        assert not called_sql.rstrip().endswith(";")
