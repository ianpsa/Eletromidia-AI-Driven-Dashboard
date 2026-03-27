import { useMemo, useState } from "react";
import type { HexbinFiltersState } from "../types/hexbin";
import {
  areHexbinFiltersEqual,
  cloneHexbinFilters,
} from "../utils/hexbinFilters";
import { HEXBIN_DISTANCE_DEFAULT } from "../utils/hexbinOptions";

export const DEFAULT_HEXBIN_FILTERS_STATE: HexbinFiltersState = {
  states: [],
  cities: [],
  addresses: [],
  hours: [],
  genders: [],
  ages: [],
  socialClasses: [],
  maxDistance: HEXBIN_DISTANCE_DEFAULT,
};

export function useHexbinFilters(initialFilters?: Partial<HexbinFiltersState>) {
  const initialState: HexbinFiltersState = {
    ...DEFAULT_HEXBIN_FILTERS_STATE,
    ...initialFilters,
  };

  const [filters, setFilters] = useState<HexbinFiltersState>(initialState);
  const [appliedFilters, setAppliedFilters] = useState<HexbinFiltersState>(
    cloneHexbinFilters(initialState),
  );

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

  const setField = <K extends keyof HexbinFiltersState>(
    key: K,
    value: HexbinFiltersState[K],
  ) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
  };

  const clearFilters = () => {
    const cleared = cloneHexbinFilters(DEFAULT_HEXBIN_FILTERS_STATE);
    setFilters(cleared);
    return cleared;
  };

  const hasChangesSinceLastApply = useMemo(
    () => !areHexbinFiltersEqual(filters, appliedFilters),
    [filters, appliedFilters],
  );

  const markFiltersAsApplied = () => {
    setAppliedFilters(cloneHexbinFilters(filters));
  };

  return {
    filters,
    setField,
    hasAnySelection,
    clearFilters,
    hasChangesSinceLastApply,
    markFiltersAsApplied,
  };
}
