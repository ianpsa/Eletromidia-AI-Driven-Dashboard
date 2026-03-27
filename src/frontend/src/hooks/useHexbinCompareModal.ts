import { useMemo, useState } from "react";
import type { CompareFilter, CompareFilterKey, CompareMode } from "../types/hexbin";

export const INITIAL_COMPARE_FILTERS: CompareFilter[] = [
  {
    key: "location",
    label: "Localização",
    enabled: false,
    value: {
      state: [],
      city: [],
      address: [],
    },
  },
  { key: "hour", label: "Hora", enabled: false, value: [] },
  { key: "distance", label: "Distância", enabled: false, value: 10 },
  { key: "gender", label: "Gênero", enabled: false, value: [] },
  { key: "age", label: "Idade", enabled: false, value: [] },
  { key: "socialClass", label: "Classe social", enabled: false, value: [] },
];

export function useHexbinCompareModal() {
  const [compareMode, setCompareMode] = useState<CompareMode | null>(null);
  const [filters, setFilters] = useState<CompareFilter[]>(INITIAL_COMPARE_FILTERS);

  const visibleFilters = useMemo(() => {
    if (!compareMode) return filters;
    return filters.filter((filter) => filter.key !== compareMode);
  }, [filters, compareMode]);

  function updateFilter(key: CompareFilterKey, patch: Partial<CompareFilter>) {
    setFilters((current) =>
      current.map((item) => (item.key === key ? ({ ...item, ...patch } as CompareFilter) : item)),
    );
  }

  function toggleFilter(key: CompareFilterKey) {
    setFilters((current) =>
      current.map((item) => (item.key === key ? { ...item, enabled: !item.enabled } : item)),
    );
  }

  return {
    compareMode,
    setCompareMode,
    filters,
    visibleFilters,
    updateFilter,
    toggleFilter,
  };
}
