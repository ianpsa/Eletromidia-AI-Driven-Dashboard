import type { DisplayItem, ObjectItem } from "../types/api";
import { fileDisplayName, folderDisplayName } from "./filename";

export function computeSortedItems(
  folders: string[],
  files: ObjectItem[],
  query: string,
  sortField: "name" | "size",
  sortAsc: boolean,
): DisplayItem[] {
  const folderItems: DisplayItem[] = folders.map((f) => ({
    type: "folder",
    id: f,
    name: folderDisplayName(f),
    size: -1,
    content_type: undefined,
    updated_at: "",
  }));

  const fileItems: DisplayItem[] = files.map((f) => ({
    type: "file",
    id: f.id,
    name: fileDisplayName(f.id),
    size: f.size,
    content_type: f.content_type,
    updated_at: f.updated_at,
  }));

  const all = [...folderItems, ...fileItems];

  const normalizedQuery = query.trim().toLowerCase();
  const filtered = normalizedQuery
    ? all.filter((item) => item.name.toLowerCase().includes(normalizedQuery))
    : all;

  return filtered.sort((a, b) => {
    if (a.type !== b.type) return a.type === "folder" ? -1 : 1;

    let cmp = 0;
    if (sortField === "name") {
      cmp = a.name.localeCompare(b.name, "pt-BR", { sensitivity: "base" });
    } else {
      cmp = a.size - b.size;
    }
    return sortAsc ? cmp : -cmp;
  });
}
