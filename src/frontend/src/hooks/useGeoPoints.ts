import { useCallback, useEffect, useState } from "react";
import type { HexbinPoint } from "../types/geo";
import type { GeodataPointsResponse } from "../types/geodataApi";
import type { HexbinFiltersState } from "../types/hexbin";
import { buildGeoPointsQuery, toHexbinPoints } from "../utils/geodataFilters";
import { buildApiUrl } from "../utils/url";

type UseGeoPointsParams = {
  filters: HexbinFiltersState;
  limit?: number;
};

export function useGeoPoints({ filters, limit }: UseGeoPointsParams) {
  const [points, setPoints] = useState<HexbinPoint[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const loadPoints = useCallback(async () => {
    setLoading(true);
    setError("");

    try {
  const query = buildGeoPointsQuery(filters, limit);
      const response = await fetch(buildApiUrl("/geodata/points", query));

      if (!response.ok) {
        throw new Error(`Falha ao buscar pontos do mapa (${response.status}).`);
      }

      const payload = (await response.json()) as GeodataPointsResponse;
      setPoints(toHexbinPoints(payload));
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Erro ao buscar dados de densidade.";
      setError(message);
      setPoints([]);
    } finally {
      setLoading(false);
    }
  }, [filters, limit]);

  useEffect(() => {
    void loadPoints();
  }, [loadPoints]);

  return {
    points,
    loading,
    error,
    refresh: () => {
      void loadPoints();
    },
  };
}
