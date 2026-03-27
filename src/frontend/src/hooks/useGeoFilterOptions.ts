import { useCallback, useEffect, useState } from "react";
import type { GeodataFilterOptions } from "../types/geodataApi";
import { buildApiUrl } from "../utils/url";
import { buildGeoFilterOptionsQuery, toGeoFilterOptions } from "../utils/geodataFilters";

type UseGeoFilterOptionsParams = {
  selectedState?: string;
  selectedCity?: string;
  fallbackOptions?: GeodataFilterOptions;
};

const DEFAULT_GEO_FILTER_OPTIONS: GeodataFilterOptions = {
  states: [],
  cities: [],
  addresses: [],
  hours: [],
  genders: [],
  ages: [],
  socialClasses: [],
};

export function useGeoFilterOptions({
  selectedState,
  selectedCity,
  fallbackOptions,
}: UseGeoFilterOptionsParams) {
  const effectiveFallback = fallbackOptions ?? DEFAULT_GEO_FILTER_OPTIONS;
  const [options, setOptions] = useState<GeodataFilterOptions>(effectiveFallback);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const loadOptions = useCallback(async () => {
    setLoading(true);
    setError("");

    try {
      const query = buildGeoFilterOptionsQuery(selectedState, selectedCity);
      const response = await fetch(buildApiUrl("/geodata/filter-options", query));
      if (!response.ok) {
        throw new Error(`Falha ao buscar opções de filtros (${response.status}).`);
      }

      const payload = (await response.json()) as {
        estados?: string[];
        cidades?: string[];
        enderecos?: string[];
        horarios?: string[];
        generos?: string[];
        faixas_etarias?: string[];
        classes_sociais?: string[];
      };
      const normalized = toGeoFilterOptions(payload);

      setOptions({
        states:
          normalized.states.length > 0
            ? normalized.states
            : effectiveFallback.states,
        cities: normalized.cities.length > 0 ? normalized.cities : effectiveFallback.cities,
        addresses:
          normalized.addresses.length > 0
            ? normalized.addresses
            : effectiveFallback.addresses,
        hours: normalized.hours.length > 0 ? normalized.hours : effectiveFallback.hours,
        genders:
          normalized.genders.length > 0
            ? normalized.genders
            : effectiveFallback.genders,
        ages: normalized.ages.length > 0 ? normalized.ages : effectiveFallback.ages,
        socialClasses:
          normalized.socialClasses.length > 0
            ? normalized.socialClasses
            : effectiveFallback.socialClasses,
      });
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Erro ao buscar opções de filtro no servidor.";
      setError(message);
      setOptions(effectiveFallback);
    } finally {
      setLoading(false);
    }
  }, [selectedState, selectedCity, effectiveFallback]);

  useEffect(() => {
    void loadOptions();
  }, [loadOptions]);

  return {
    options,
    loading,
    error,
    refresh: () => {
      void loadOptions();
    },
  };
}
