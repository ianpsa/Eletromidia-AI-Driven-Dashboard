"""Spatial join between Eletromidia screen locations and Claro audience data.

Three-tier matching:
  1. Exact   — Claro points within ``radius_m`` (default 500 m), weighted by uniques.
  2. Fallback — Claro points within ``fallback_radius_m``
     (default 5 km), inverse-distance weighted.
  3. Global   — Uniques-weighted average of all Claro data (ensures 100 % coverage).

Usage::

    python scripts/enrich.py \\
        --claro  ../../data/fluxo_claro.csv \\
        --eletro ../../data/eletromidia_pontos.csv \\
        --output ../../data/eletromidia_enriched.csv
"""

from __future__ import annotations

import argparse
import ast
import sys
from pathlib import Path

import numpy as np
import pandas as pd

PROP_COLS: list[str] = [
    "p_18_19",
    "p_20_29",
    "p_30_39",
    "p_40_49",
    "p_50_59",
    "p_60_69",
    "p_70_79",
    "p_80_plus",
    "p_f",
    "p_m",
    "p_a",
    "p_b1",
    "p_b2",
    "p_c1",
    "p_c2",
    "p_de",
]


# ---------------------------------------------------------------------------
# Haversine
# ---------------------------------------------------------------------------


def haversine_matrix(
    lat1: np.ndarray,
    lon1: np.ndarray,
    lat2: np.ndarray,
    lon2: np.ndarray,
) -> np.ndarray:
    """Vectorised haversine — returns distance matrix in **metres**.

    Parameters are 1-D arrays; the result is shape ``(len(lat1), len(lat2))``.
    """
    R = 6_371_000.0
    lat1_r = np.radians(lat1[:, None])
    lon1_r = np.radians(lon1[:, None])
    lat2_r = np.radians(lat2[None, :])
    lon2_r = np.radians(lon2[None, :])

    dlat = lat2_r - lat1_r
    dlon = lon2_r - lon1_r

    a = np.sin(dlat / 2) ** 2 + np.cos(lat1_r) * np.cos(lat2_r) * np.sin(dlon / 2) ** 2
    return 2 * R * np.arcsin(np.sqrt(a))


# ---------------------------------------------------------------------------
# Claro data loading
# ---------------------------------------------------------------------------


def parse_target(s: str) -> dict:
    """Parse the stringified ``target`` dict into flat proportion columns."""
    if not isinstance(s, str) or not s.strip():
        return {c: 0.0 for c in PROP_COLS}
    d = ast.literal_eval(s)
    idade = d.get("idade", {})
    genero = d.get("genero", {})
    classe = d.get("classe_social", {})
    return {
        "p_18_19": idade.get("18-19", 0.0),
        "p_20_29": idade.get("20-29", 0.0),
        "p_30_39": idade.get("30-39", 0.0),
        "p_40_49": idade.get("40-49", 0.0),
        "p_50_59": idade.get("50-59", 0.0),
        "p_60_69": idade.get("60-69", 0.0),
        "p_70_79": idade.get("70-79", 0.0),
        "p_80_plus": idade.get("80+", 0.0),
        "p_f": genero.get("F", 0.0),
        "p_m": genero.get("M", 0.0),
        "p_a": classe.get("A", 0.0),
        "p_b1": classe.get("B1", 0.0),
        "p_b2": classe.get("B2", 0.0),
        "p_c1": classe.get("C1", 0.0),
        "p_c2": classe.get("C2", 0.0),
        "p_de": classe.get("DE", 0.0),
    }


def load_claro(path: str | Path) -> pd.DataFrame:
    """Load Claro CSV, parse targets, and aggregate by ``location_id``.

    Returns one row per location with uniques-weighted average proportions.
    """
    df = pd.read_csv(path)
    parsed = df["target"].apply(parse_target).apply(pd.Series)
    df = pd.concat([df, parsed], axis=1)

    grouped = df.groupby("location_id")
    result = grouped.agg(
        {
            "uniques": "sum",
            "latitude": "first",
            "longitude": "first",
            "cidade": "first",
            "endereco": "first",
            "numero": "first",
        },
    ).reset_index()

    props_wa = []
    for _lid, grp in grouped:
        w = grp["uniques"].values
        total_w = w.sum()
        if total_w == 0:
            props_wa.append({c: 0.0 for c in PROP_COLS})
            continue
        row = {}
        for c in PROP_COLS:
            row[c] = float(np.average(grp[c].values, weights=w))
        props_wa.append(row)

    result = pd.concat(
        [result.reset_index(drop=True), pd.DataFrame(props_wa)],
        axis=1,
    )
    return result


# ---------------------------------------------------------------------------
# Global average fallback
# ---------------------------------------------------------------------------


def _global_averages(claro: pd.DataFrame, prop_cols: list[str]) -> dict:
    """Compute uniques-weighted global average of all Claro proportions."""
    w = claro["uniques"].values
    props = claro[prop_cols].values
    total_w = w.sum()
    avg_props = (props * w[:, None]).sum(axis=0) / total_w
    avg_uniques = w.mean()
    cidade = claro["cidade"].mode().iloc[0]
    return {
        "props": avg_props,
        "uniques": float(avg_uniques),
        "cidade": cidade,
        "prop_cols": prop_cols,
    }


# ---------------------------------------------------------------------------
# Row builder
# ---------------------------------------------------------------------------


def _build_row(
    eletro_row: pd.Series,
    claro_uniques: np.ndarray,
    claro_props: np.ndarray,
    claro_cidades: np.ndarray,
    claro_enderecos: np.ndarray,
    claro_numeros: np.ndarray,
    prop_cols: list[str],
    mask: np.ndarray,
    distances: np.ndarray,
    is_fallback: bool,
) -> dict:
    """Build one enriched row for an Eletromidia screen."""
    dists = distances[mask]
    uniques = claro_uniques[mask]

    if is_fallback:
        safe_dists = np.maximum(dists, 1.0)
        w = uniques / safe_dists
    else:
        w = uniques

    total_w = w.sum()
    if total_w == 0:
        return {}

    props = claro_props[mask]
    avg_props = (props * w[:, None]).sum(axis=0) / total_w
    total_uniques = float(uniques.sum())

    nearest_idx = np.argmin(dists)
    idxs = np.where(mask)[0]
    nearest_global = idxs[nearest_idx]

    row = {
        "cod_predio": eletro_row["cod_predio"],
        "latitude": eletro_row["latitude"],
        "longitude": eletro_row["longitude"],
        "vertical": eletro_row["vertical"],
        "ambiente": eletro_row["ambiente"],
        "cidade": str(claro_cidades[nearest_global]),
        "endereco_ref": (
            f"{claro_enderecos[nearest_global]}, {claro_numeros[nearest_global]}"
        ),
        "uniques": total_uniques,
        "match_type": "fallback" if is_fallback else "exact",
    }
    for j, c in enumerate(prop_cols):
        row[c] = float(avg_props[j])
    return row


# ---------------------------------------------------------------------------
# Main enrichment
# ---------------------------------------------------------------------------


def enrich(
    claro: pd.DataFrame,
    eletro: pd.DataFrame,
    prop_cols: list[str],
    radius_m: float,
    fallback_radius_m: float,
) -> pd.DataFrame:
    """Perform three-tier spatial join and return enriched DataFrame."""
    dist = haversine_matrix(
        eletro["latitude"].values,
        eletro["longitude"].values,
        claro["latitude"].values,
        claro["longitude"].values,
    )

    claro_uniques = claro["uniques"].values
    claro_props = claro[prop_cols].values
    claro_cidades = claro["cidade"].values
    claro_enderecos = claro["endereco"].values
    claro_numeros = claro["numero"].values

    global_avg = _global_averages(claro, prop_cols)

    rows: list[dict] = []
    n = len(eletro)

    for i in range(n):
        if (i + 1) % 5000 == 0 or i == n - 1:
            print(f"  [{i + 1}/{n}]", file=sys.stderr)

        eletro_row = eletro.iloc[i]
        dists_i = dist[i]

        primary_mask = dists_i <= radius_m
        if primary_mask.any():
            row = _build_row(
                eletro_row,
                claro_uniques,
                claro_props,
                claro_cidades,
                claro_enderecos,
                claro_numeros,
                prop_cols,
                primary_mask,
                dists_i,
                is_fallback=False,
            )
            if row:
                rows.append(row)
                continue

        fallback_mask = dists_i <= fallback_radius_m
        if fallback_mask.any():
            row = _build_row(
                eletro_row,
                claro_uniques,
                claro_props,
                claro_cidades,
                claro_enderecos,
                claro_numeros,
                prop_cols,
                fallback_mask,
                dists_i,
                is_fallback=True,
            )
            if row:
                rows.append(row)
                continue

        row = {
            "cod_predio": eletro_row["cod_predio"],
            "latitude": eletro_row["latitude"],
            "longitude": eletro_row["longitude"],
            "vertical": eletro_row["vertical"],
            "ambiente": eletro_row["ambiente"],
            "cidade": global_avg["cidade"],
            "endereco_ref": str(eletro_row["cod_predio"]),
            "uniques": global_avg["uniques"],
            "match_type": "global_avg",
        }
        for j, c in enumerate(prop_cols):
            row[c] = float(global_avg["props"][j])
        rows.append(row)

    return pd.DataFrame(rows)


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Enrich Eletromidia screens with Claro audience data",
    )
    parser.add_argument("--claro", required=True, help="Path to Claro CSV")
    parser.add_argument("--eletro", required=True, help="Path to Eletromidia CSV")
    parser.add_argument("--output", required=True, help="Output path for enriched CSV")
    parser.add_argument(
        "--radius",
        type=float,
        default=500,
        help="Primary match radius in metres (default 500)",
    )
    parser.add_argument(
        "--fallback-radius",
        type=float,
        default=5000,
        help="Fallback radius in metres (default 5000)",
    )
    args = parser.parse_args()

    print(f"Loading Claro data from {args.claro} …", file=sys.stderr)
    claro = load_claro(args.claro)
    print(f"  {len(claro)} unique locations", file=sys.stderr)

    print(f"Loading Eletromidia data from {args.eletro} …", file=sys.stderr)
    eletro = pd.read_csv(args.eletro)
    print(f"  {len(eletro)} screens", file=sys.stderr)

    print(
        f"Enriching (radius={args.radius} m, fallback={args.fallback_radius} m) …",
        file=sys.stderr,
    )
    result = enrich(claro, eletro, PROP_COLS, args.radius, args.fallback_radius)

    counts = result["match_type"].value_counts()
    print(f"Match types:\n{counts.to_string()}", file=sys.stderr)
    print(
        f"Total: {len(result)} rows ({len(result) / len(eletro) * 100:.1f}% coverage)",
        file=sys.stderr,
    )

    result.to_csv(args.output, index=False)
    print(f"Saved to {args.output}", file=sys.stderr)


if __name__ == "__main__":
    main()
