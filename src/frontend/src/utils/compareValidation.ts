import type { CompareFilter, CompareMode } from "../types/hexbin";

function hasLocationValue(filter: Extract<CompareFilter, { key: "location" }>) {
  return (
    filter.value.state.length > 0 ||
    filter.value.city.length > 0 ||
    filter.value.address.length > 0
  );
}

export function hasCompareExtraFilterValue(filter: CompareFilter) {
  if (!filter.enabled) return true;

  switch (filter.key) {
    case "location":
      return hasLocationValue(filter);
    case "hour":
    case "gender":
    case "age":
    case "socialClass":
      return filter.value.length > 0;
    case "distance":
      return Number.isFinite(filter.value);
    default:
      return false;
  }
}

export function canConfirmCompareConfig(
  compareMode: CompareMode | null,
  filtersToValidate: CompareFilter[],
) {
  if (!compareMode) return false;
  return filtersToValidate.every(hasCompareExtraFilterValue);
}
