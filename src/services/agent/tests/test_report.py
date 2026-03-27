import math

import pandas as pd
import pytest

from core.report import build_report, normalize_gender, parse_age_range, print_table


# ── parse_age_range ──────────────────────────────────────────────────────────


@pytest.mark.parametrize(
    "label, expected",
    [
        ("18-25", (18, 25)),
        ("0-10", (0, 10)),
        ("100-200", (100, 200)),
        (float("nan"), (None, None)),
        (None, (None, None)),
        ("abc", (None, None)),
        ("18", (None, None)),
        ("", (None, None)),
        ("18-25-30", (None, None)),
    ],
    ids=[
        "faixa normal",
        "faixa começando em zero",
        "faixa com valores altos",
        "NaN",
        "None",
        "string inválida",
        "número isolado",
        "string vazia",
        "três partes",
    ],
)
def test_parse_age_range(label, expected):
    assert parse_age_range(label) == expected


# ── normalize_gender ─────────────────────────────────────────────────────────


@pytest.mark.parametrize(
    "value, expected",
    [
        ("Feminino", "female"),
        ("feminino", "female"),
        ("FEMININO", "female"),
        ("FEM", "female"),
        ("Masculino", "male"),
        ("masculino", "male"),
        ("MASCULINO", "male"),
        ("masc", "male"),
        ("Outro", "outro"),
        ("unknown", "unknown"),
        (123, "123"),
    ],
    ids=[
        "Feminino capitalizado",
        "feminino minúsculo",
        "FEMININO maiúsculo",
        "FEM abreviado",
        "Masculino capitalizado",
        "masculino minúsculo",
        "MASCULINO maiúsculo",
        "masc abreviado",
        "Outro valor",
        "unknown",
        "inteiro",
    ],
)
def test_normalize_gender(value, expected):
    assert normalize_gender(value) == expected


# ── build_report ─────────────────────────────────────────────────────────────


class TestBuildReportFilters:
    """Testes de filtragem do build_report."""

    def test_sem_filtros_retorna_tudo_agregado(self, sample_df):
        rows, used_range = build_report(sample_df, {})
        assert len(rows) > 0
        assert used_range == (None, None)

    def test_filtro_genero_feminino(self, sample_df):
        rows, _ = build_report(sample_df, {"gender": "female"})
        assert len(rows) > 0
        # Av. Paulista 1000 deve ter joint_count de Feminino: 100 + 70 = 170
        paulista = [r for r in rows if "Paulista" in r["point"] and "1000" in r["point"]]
        assert len(paulista) == 1
        assert paulista[0]["total_target"] == 170.0

    def test_filtro_genero_masculino(self, sample_df):
        rows, _ = build_report(sample_df, {"gender": "male"})
        assert len(rows) > 0
        # Nenhuma linha feminina deve aparecer nos pontos somente masculinos
        for r in rows:
            assert r["total_target"] > 0

    def test_filtro_classe_social(self, sample_df):
        rows, _ = build_report(sample_df, {"classes": ["A"]})
        assert len(rows) > 0
        # Todas as linhas de classe A somadas por ponto

    def test_filtro_classe_social_multipla(self, sample_df):
        rows_ab, _ = build_report(sample_df, {"classes": ["A", "B"]})
        rows_a, _ = build_report(sample_df, {"classes": ["A"]})
        # A+B deve ter >= resultados que só A
        total_ab = sum(r["total_target"] for r in rows_ab)
        total_a = sum(r["total_target"] for r in rows_a)
        assert total_ab >= total_a

    def test_filtro_cidade_case_insensitive(self, sample_df):
        rows_lower, _ = build_report(sample_df, {"city": "são paulo"})
        rows_upper, _ = build_report(sample_df, {"city": "São Paulo"})
        assert len(rows_lower) == len(rows_upper)
        for r1, r2 in zip(rows_lower, rows_upper):
            assert r1["point"] == r2["point"]
            assert r1["total_target"] == r2["total_target"]

    def test_filtro_idade_overlap(self, sample_df):
        # Pedir 20-30 deve casar com "18-25" (overlap 20-25) e "26-35" (overlap 26-30)
        rows, (used_min, used_max) = build_report(
            sample_df, {"age_min": 20, "age_max": 30}
        )
        assert len(rows) > 0
        assert used_min == 18  # mínimo das faixas que casaram
        assert used_max == 35  # máximo das faixas que casaram

    def test_filtro_idade_sem_overlap(self, sample_df):
        # Faixa que não existe nos dados
        rows, (used_min, used_max) = build_report(
            sample_df, {"age_min": 90, "age_max": 99}
        )
        assert rows == []

    def test_filtro_idade_somente_min(self, sample_df):
        rows, _ = build_report(sample_df, {"age_min": 30})
        assert len(rows) > 0
        # Deve casar com faixas onde end >= 30: "26-35", "36-45", "46-55"

    def test_filtro_idade_somente_max(self, sample_df):
        rows, _ = build_report(sample_df, {"age_max": 20})
        assert len(rows) > 0
        # Deve casar com faixas onde start <= 20: "18-25"

    def test_filtros_que_nao_casam_nada(self, sample_df):
        rows, _ = build_report(sample_df, {"city": "Manaus"})
        assert rows == []

    def test_filtros_combinados(self, sample_df):
        rows, _ = build_report(
            sample_df,
            {"gender": "female", "classes": ["A"], "city": "São Paulo"},
        )
        assert len(rows) > 0
        # Av. Paulista 1000: Feminino + classe A + SP → linhas 0 e 6 (100 + 70)
        paulista = [r for r in rows if "Paulista" in r["point"]]
        assert len(paulista) == 1
        assert paulista[0]["total_target"] == 170.0


class TestBuildReportAggregation:
    """Testes de agregação e formato de saída."""

    def test_affinity_calculado_corretamente(self, sample_df):
        rows, _ = build_report(sample_df, {"gender": "female", "city": "Rio de Janeiro"})
        # Av. Atlântica 200: joint_count=80, uniques=400 → affinity=20.0
        assert len(rows) == 1
        assert rows[0]["point"] == "Av. Atlântica, 200"
        assert rows[0]["total_target"] == 80.0
        assert rows[0]["total_flow"] == 400.0
        assert rows[0]["affinity"] == 20.0

    def test_ordenado_por_affinity_descendente(self, sample_df):
        rows, _ = build_report(sample_df, {})
        affinities = [r["affinity"] for r in rows]
        assert affinities == sorted(affinities, reverse=True)

    def test_formato_saida(self, sample_df):
        rows, _ = build_report(sample_df, {})
        assert len(rows) > 0
        for r in rows:
            assert set(r.keys()) == {"point", "total_target", "total_flow", "affinity"}
            assert isinstance(r["point"], str)
            assert isinstance(r["total_target"], float)
            assert isinstance(r["total_flow"], float)
            assert isinstance(r["affinity"], float)

    def test_groupby_endereco_numero(self, sample_df):
        rows, _ = build_report(sample_df, {})
        points = [r["point"] for r in rows]
        # Não deve ter pontos duplicados
        assert len(points) == len(set(points))

    def test_aggregation_sum_joint_max_uniques(self, sample_df):
        # Av. Paulista 1000: linhas 0,1,6 → joint_count sum=220, uniques max=500
        rows, _ = build_report(sample_df, {})
        paulista = [r for r in rows if r["point"] == "Av. Paulista, 1000"]
        assert len(paulista) == 1
        assert paulista[0]["total_target"] == 220.0
        assert paulista[0]["total_flow"] == 500.0
        expected_affinity = round((220.0 / 500.0) * 100, 2)
        assert paulista[0]["affinity"] == expected_affinity


# ── print_table ──────────────────────────────────────────────────────────────


class TestPrintTable:
    def test_lista_vazia(self, capsys):
        print_table([], 5)
        captured = capsys.readouterr()
        assert "No matching audience found" in captured.out

    def test_respeita_limit(self, capsys):
        rows = [
            {"point": f"Ponto {i}", "total_target": 10.0, "total_flow": 100.0, "affinity": 10.0}
            for i in range(10)
        ]
        print_table(rows, 3)
        captured = capsys.readouterr()
        lines = [l for l in captured.out.strip().split("\n") if l and "---" not in l and "point" not in l]
        assert len(lines) == 3

    def test_formato_output(self, capsys):
        rows = [
            {"point": "Av. Test, 100", "total_target": 50.0, "total_flow": 200.0, "affinity": 25.0}
        ]
        print_table(rows, 5)
        captured = capsys.readouterr()
        assert "Av. Test, 100" in captured.out
        assert "50.0" in captured.out
        assert "200.0" in captured.out
        assert "25.0" in captured.out
