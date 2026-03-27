export function formatMultiSelectDisplay(selected: string[], placeholder = "Todos") {
  if (selected.length === 0) return placeholder;
  if (selected.length <= 2) return selected.join(", ");
  return `${selected.slice(0, 2).join(", ")} (+${selected.length - 2})`;
}
