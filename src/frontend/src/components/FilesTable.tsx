import type { DisplayItem } from "../types/api";
import { formatBytes, formatDate } from "../utils/format";

interface FilesTableProps {
  items: DisplayItem[];
  currentFolder: string;
  sortField: "name" | "size";
  sortAsc: boolean;
  downloadingId: string;
  onSort: (field: "name" | "size") => void;
  onNavigate: (folder: string) => void;
  onNavigateUp: () => void;
  onDownload: (file: { id: string; name: string }) => void;
}

function sortIndicator(
  field: "name" | "size",
  sortField: "name" | "size",
  sortAsc: boolean,
): string {
  if (sortField !== field) return "";
  return sortAsc ? " \u25B2" : " \u25BC";
}

export function FilesTable({
  items,
  currentFolder,
  sortField,
  sortAsc,
  downloadingId,
  onSort,
  onNavigate,
  onNavigateUp,
  onDownload,
}: FilesTableProps) {
  return (
    <section className="files-section">
      <div className="files-table-wrap">
        <table>
          <thead>
            <tr>
              <th className="sortable" onClick={() => onSort("name")}>
                Nome{sortIndicator("name", sortField, sortAsc)}
              </th>
              <th className="sortable" onClick={() => onSort("size")}>
                Tamanho{sortIndicator("size", sortField, sortAsc)}
              </th>
              <th>Tipo</th>
              <th>Atualizado em</th>
              <th aria-label="Acoes" />
            </tr>
          </thead>
          <tbody>
            {currentFolder && (
              <tr className="row-folder row-clickable" onClick={onNavigateUp}>
                <td colSpan={5} className="back-row">
                  <span className="folder-icon" aria-hidden>
                    &#128193;
                  </span>{" "}
                  ..
                </td>
              </tr>
            )}
            {items.map((item) =>
              item.type === "folder" ? (
                <tr
                  key={item.id}
                  className="row-folder row-clickable"
                  onClick={() => onNavigate(item.id)}
                >
                  <td>
                    <span className="folder-icon" aria-hidden>
                      &#128193;
                    </span>{" "}
                    {item.name}
                  </td>
                  <td>-</td>
                  <td>Pasta</td>
                  <td>-</td>
                  <td />
                </tr>
              ) : (
                <tr key={item.id}>
                  <td>
                    <span className="file-icon" aria-hidden>
                      &#128196;
                    </span>{" "}
                    {item.name}
                  </td>
                  <td>{formatBytes(item.size)}</td>
                  <td>{item.content_type || "-"}</td>
                  <td>{formatDate(item.updated_at)}</td>
                  <td>
                    <button
                      type="button"
                      className="download-btn"
                      onClick={() => onDownload(item)}
                      disabled={downloadingId === item.id}
                    >
                      {downloadingId === item.id ? "Baixando..." : "Download"}
                    </button>
                  </td>
                </tr>
              ),
            )}
          </tbody>
        </table>
      </div>
    </section>
  );
}
