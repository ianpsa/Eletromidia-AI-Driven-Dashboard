import { useEffect, useMemo, useState } from "react";
import type {
  CompareChartsConfig,
  CompareFilter,
  CompareFilterKey,
  CompareMode,
} from "../types/hexbin";
import { HEXBIN_DISTANCE_DEFAULT } from "../utils/hexbinOptions";

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
  {
    key: "distance",
    label: "Distância",
    enabled: false,
    value: HEXBIN_DISTANCE_DEFAULT,
  },
  { key: "gender", label: "Gênero", enabled: false, value: [] },
  { key: "age", label: "Idade", enabled: false, value: [] },
  { key: "socialClass", label: "Classe social", enabled: false, value: [] },
];

export function cloneCompareFilter(filter: CompareFilter): CompareFilter {
  if (filter.key === "location") {
    return {
      ...filter,
      value: {
        state: [...filter.value.state],
        city: [...filter.value.city],
        address: [...filter.value.address],
      },
    };
  }

  if (Array.isArray(filter.value)) {
    return {
      ...filter,
      value: [...filter.value],
    } as CompareFilter;
  }

  return {
    ...filter,
  };
}

export function getEmptyCompareConfig(): CompareChartsConfig {
  return {
    compareMode: null,
    filters: INITIAL_COMPARE_FILTERS.map(cloneCompareFilter),
  };
}

type UseHexbinCompareModalParams = {
  open: boolean;
  initialConfig?: CompareChartsConfig | null;
};

export function useHexbinCompareModal({
  open,
  initialConfig,
}: UseHexbinCompareModalParams) {
  const [compareMode, setCompareMode] = useState<CompareMode | null>(null);
  const [filters, setFilters] = useState<CompareFilter[]>(
    getEmptyCompareConfig().filters,
  );

  useEffect(() => {
    if (!open) return;

    const source = initialConfig ?? getEmptyCompareConfig();
    setCompareMode(source.compareMode);
    setFilters(source.filters.map(cloneCompareFilter));
  }, [open, initialConfig]);

  const visibleFilters = useMemo(() => {
    if (!compareMode) return filters;
    return filters.filter((filter) => filter.key !== compareMode);
  }, [filters, compareMode]);

  function updateFilter(key: CompareFilterKey, patch: Partial<CompareFilter>) {
    setFilters((current) =>
      current.map((item) =>
        item.key === key ? ({ ...item, ...patch } as CompareFilter) : item,
      ),
    );
  }

  function toggleFilter(key: CompareFilterKey) {
    setFilters((current) =>
      current.map((item) =>
        item.key === key ? { ...item, enabled: !item.enabled } : item,
      ),
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
