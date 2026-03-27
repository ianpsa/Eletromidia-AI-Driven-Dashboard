export const DEFAULT_MULTISELECT_MAX_VISIBLE_OPTIONS = 6;
const OPTION_ROW_HEIGHT = 42;
const DROPDOWN_VERTICAL_PADDING = 16;

export function formatMultiSelectDisplay(selected: string[], placeholder = "Todos") {
  if (selected.length === 0) return placeholder;
  if (selected.length <= 2) return selected.join(", ");
  return `${selected.slice(0, 2).join(", ")} (+${selected.length - 2})`;
}

export function shouldEnableMultiSelectScroll(
  optionCount: number,
  maxVisibleOptions = DEFAULT_MULTISELECT_MAX_VISIBLE_OPTIONS,
) {
  return optionCount > maxVisibleOptions;
}

export function getMultiSelectDropdownMaxHeight(
  optionCount: number,
  maxVisibleOptions = DEFAULT_MULTISELECT_MAX_VISIBLE_OPTIONS,
) {
  const visibleRows = Math.min(optionCount, maxVisibleOptions);
  return visibleRows * OPTION_ROW_HEIGHT + DROPDOWN_VERTICAL_PADDING;
}
