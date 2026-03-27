import type { CompareChartsConfig, CompareFilterKey } from "../types/hexbin";

export type CompareTagItem =
  | {
      id: CompareFilterKey;
      kind: "filter";
      label: string;
      filterKey: CompareFilterKey;
    }
  | {
      id: "compareMode";
      kind: "mode";
      label: string;
    };

const COMPARE_MODE_LABEL: Record<
  NonNullable<CompareChartsConfig["compareMode"]>,
  string
> = {
  gender: "Comparar por gênero",
  age: "Comparar por idade",
  socialClass: "Comparar por classe social",
};

export function buildCompareTags(config: CompareChartsConfig) {
  const filterTags: CompareTagItem[] = config.filters
    .filter((filter) => filter.enabled)
    .map((filter) => ({
      id: filter.key,
      kind: "filter",
      label: filter.label,
      filterKey: filter.key,
    }));

  const compareTag: CompareTagItem | null = config.compareMode
    ? {
        id: "compareMode",
        kind: "mode",
        label: COMPARE_MODE_LABEL[config.compareMode],
      }
    : null;

  return {
    filterTags,
    compareTag,
  };
}
