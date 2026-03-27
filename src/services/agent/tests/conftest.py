import pandas as pd
import pytest


@pytest.fixture
def sample_df():
    """DataFrame de exemplo simulando a estrutura do CSV real."""
    data = {
        "gender_label": [
            "Feminino", "Masculino", "Feminino", "Masculino",
            "Feminino", "Masculino", "Feminino",
        ],
        "class_label": ["A", "B", "A", "C", "B", "A", "A"],
        "cidade": [
            "São Paulo", "São Paulo", "Rio de Janeiro", "São Paulo",
            "Curitiba", "Rio de Janeiro", "São Paulo",
        ],
        "age_label": ["18-25", "26-35", "18-25", "36-45", "26-35", "18-25", "46-55"],
        "endereco": [
            "Av. Paulista", "Av. Paulista", "Av. Atlântica", "Rua Augusta",
            "Rua XV", "Av. Atlântica", "Av. Paulista",
        ],
        "numero": [1000, 1000, 200, 500, 100, 200, 1000],
        "joint_count": [100.0, 50.0, 80.0, 30.0, 60.0, 40.0, 70.0],
        "uniques": [500.0, 500.0, 400.0, 200.0, 300.0, 400.0, 500.0],
    }
    return pd.DataFrame(data)
