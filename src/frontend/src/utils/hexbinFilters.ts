import type { HexbinFiltersState } from "../types/hexbin";

export function cloneHexbinFilters(filters: HexbinFiltersState): HexbinFiltersState {
  return {
    states: [...filters.states],
    cities: [...filters.cities],
    addresses: [...filters.addresses],
    hours: [...filters.hours],
    genders: [...filters.genders],
    ages: [...filters.ages],
    socialClasses: [...filters.socialClasses],
    maxDistance: filters.maxDistance,
  };
}

function sameStringArray(left: string[], right: string[]) {
  return left.length === right.length && left.every((value, index) => value === right[index]);
}

export function areHexbinFiltersEqual(left: HexbinFiltersState, right: HexbinFiltersState) {
  return (
    sameStringArray(left.states, right.states) &&
    sameStringArray(left.cities, right.cities) &&
    sameStringArray(left.addresses, right.addresses) &&
    sameStringArray(left.hours, right.hours) &&
    sameStringArray(left.genders, right.genders) &&
    sameStringArray(left.ages, right.ages) &&
    sameStringArray(left.socialClasses, right.socialClasses) &&
    left.maxDistance === right.maxDistance
  );
}
