import { useMemo, useState } from "react";
import type { HexbinFiltersState } from "../types/hexbin";

export const DEFAULT_HEXBIN_FILTERS_STATE: HexbinFiltersState = {
  states: [],
  cities: [],
  addresses: [],
  hours: [],
  genders: [],
  ages: [],
  socialClasses: [],
  maxDistance: 10,
};

export function useHexbinFilters(initialFilters?: Partial<HexbinFiltersState>) {
  const [filters, setFilters] = useState<HexbinFiltersState>({
    ...DEFAULT_HEXBIN_FILTERS_STATE,
    ...initialFilters,
  });

  const hasAnySelection = useMemo(
    () =>
      filters.states.length > 0 ||
      filters.cities.length > 0 ||
      filters.addresses.length > 0 ||
      filters.hours.length > 0 ||
      filters.genders.length > 0 ||
      filters.ages.length > 0 ||
      filters.socialClasses.length > 0 ||
      filters.maxDistance !== DEFAULT_HEXBIN_FILTERS_STATE.maxDistance,
    [filters],
  );

  const setField = <K extends keyof HexbinFiltersState>(key: K, value: HexbinFiltersState[K]) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
  };

  const clearFilters = () => {
    const cleared = { ...DEFAULT_HEXBIN_FILTERS_STATE };
    setFilters(cleared);
    return cleared;
  };

  return {
    filters,
    setField,
    hasAnySelection,
    clearFilters,
  };
}
