import os
from unittest.mock import MagicMock, patch

import numpy as np
import pandas as pd
import pytest
from fastapi.testclient import TestClient

from api import CampaignRequest, app, convert_types


client = TestClient(app)


# ── convert_types ────────────────────────────────────────────────────────────


class TestConvertTypes:
    def test_numpy_int(self):
        assert convert_types(np.int64(42)) == 42
        assert isinstance(convert_types(np.int64(42)), int)

    def test_numpy_float(self):
        assert convert_types(np.float64(3.14)) == 3.14
        assert isinstance(convert_types(np.float64(3.14)), float)

    def test_string_inalterada(self):
        assert convert_types("hello") == "hello"

    def test_int_python_inalterado(self):
        assert convert_types(42) == 42

    def test_none_inalterado(self):
        assert convert_types(None) is None


# ── CampaignRequest ──────────────────────────────────────────────────────────


class TestCampaignRequest:
    def test_prompt_obrigatorio(self):
        with pytest.raises(Exception):
            CampaignRequest()

    def test_limit_default(self):
        req = CampaignRequest(prompt="teste")
        assert req.limit == 5

    def test_limit_customizado(self):
        req = CampaignRequest(prompt="teste", limit=10)
        assert req.limit == 10


# ── Endpoint /analyze ────────────────────────────────────────────────────────


class TestAnalyzeEndpoint:
    @patch.dict(os.environ, {}, clear=True)
    def test_sem_groq_api_key_retorna_500(self):
        """Sem GROQ_API_KEY definida deve retornar HTTP 500."""
        # Garantir que a variável não existe
        os.environ.pop("GROQ_API_KEY", None)
        response = client.post("/analyze", json={"prompt": "teste"})
        assert response.status_code == 500
        assert "GROQ_API_KEY" in response.json()["detail"]

    @patch("api.generate_final_answer", return_value="Análise completa.")
    @patch("api.build_report", return_value=([], (None, None)))
    @patch("api.parse_prompt", return_value={"gender": "female"})
    @patch("api.pd.read_csv")
    @patch.dict(os.environ, {"GROQ_API_KEY": "fake-key"})
    def test_sem_resultados_retorna_success_false(
        self, mock_csv, mock_parse, mock_report, mock_answer
    ):
        """Quando não há dados para os filtros, retorna success=false."""
        mock_csv.return_value = pd.DataFrame()

        response = client.post("/analyze", json={"prompt": "campanha impossível"})
        assert response.status_code == 200
        body = response.json()
        assert body["success"] is False
        assert "Nenhum dado encontrado" in body["message"]

    @patch("api.generate_final_answer", return_value="Análise completa.")
    @patch("api.build_report")
    @patch("api.parse_prompt", return_value={"gender": "female", "city": "Manaus"})
    @patch("api.pd.read_csv")
    @patch.dict(os.environ, {"GROQ_API_KEY": "fake-key"})
    def test_city_fallback(self, mock_csv, mock_parse, mock_report, mock_answer):
        """Quando cidade não tem dados, remove filtro de cidade e retenta."""
        mock_csv.return_value = pd.DataFrame()

        # Primeira chamada (com city): sem resultados; segunda (sem city): com resultados
        ranking = [{"point": "Av. X, 1", "total_target": 10.0, "total_flow": 100.0, "affinity": 10.0}]
        mock_report.side_effect = [
            ([], (None, None)),
            (ranking, (18, 25)),
        ]

        response = client.post("/analyze", json={"prompt": "campanha em Manaus"})
        assert response.status_code == 200
        body = response.json()
        assert body["success"] is True
        assert "top_points" in body

        # Verificar que generate_final_answer foi chamada com city_fallback=True
        mock_answer.assert_called_once()
        call_kwargs = mock_answer.call_args
        assert call_kwargs.kwargs.get("city_fallback") is True or (
            len(call_kwargs.args) > 4 and call_kwargs.args[4] is True
        )

    @patch("api.generate_final_answer", return_value="Recomendo os pontos A e B.")
    @patch("api.build_report")
    @patch("api.parse_prompt", return_value={"gender": "female"})
    @patch("api.pd.read_csv")
    @patch.dict(os.environ, {"GROQ_API_KEY": "fake-key"})
    def test_sucesso_retorna_analise_e_top_points(
        self, mock_csv, mock_parse, mock_report, mock_answer
    ):
        """Resposta de sucesso deve conter analysis e top_points."""
        mock_csv.return_value = pd.DataFrame()
        ranking = [
            {"point": "Av. Paulista, 1000", "total_target": 100.0, "total_flow": 500.0, "affinity": 20.0},
            {"point": "Rua Augusta, 500", "total_target": 30.0, "total_flow": 200.0, "affinity": 15.0},
        ]
        mock_report.return_value = (ranking, (18, 25))

        response = client.post("/analyze", json={"prompt": "campanha para mulheres"})
        assert response.status_code == 200
        body = response.json()
        assert body["success"] is True
        assert body["analysis"] == "Recomendo os pontos A e B."
        assert len(body["top_points"]) == 2
        assert body["top_points"][0]["point"] == "Av. Paulista, 1000"

    @patch("api.parse_prompt", side_effect=RuntimeError("erro inesperado"))
    @patch("api.pd.read_csv")
    @patch.dict(os.environ, {"GROQ_API_KEY": "fake-key"})
    def test_excecao_generica_retorna_500(self, mock_csv, mock_parse):
        """Exceção genérica durante processamento deve retornar HTTP 500."""
        response = client.post("/analyze", json={"prompt": "teste"})
        assert response.status_code == 500
        assert "erro inesperado" in response.json()["detail"]

    @patch("api.generate_final_answer", return_value="Análise.")
    @patch("api.build_report")
    @patch("api.parse_prompt", return_value={"gender": "male"})
    @patch("api.pd.read_csv")
    @patch.dict(os.environ, {"GROQ_API_KEY": "fake-key"})
    def test_limit_respeita_parametro(
        self, mock_csv, mock_parse, mock_report, mock_answer
    ):
        """O parâmetro limit deve limitar a quantidade de top_points."""
        mock_csv.return_value = pd.DataFrame()
        ranking = [
            {"point": f"Ponto {i}", "total_target": 10.0, "total_flow": 100.0, "affinity": 10.0}
            for i in range(10)
        ]
        mock_report.return_value = (ranking, (None, None))

        response = client.post("/analyze", json={"prompt": "teste", "limit": 3})
        assert response.status_code == 200
        body = response.json()
        assert len(body["top_points"]) == 3
