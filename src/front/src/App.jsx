import { useCallback, useEffect, useMemo, useState } from "react";

const API_BASE = (import.meta.env.VITE_BFF_BASE_URL || "/api").replace(/\/$/, "");

function buildApiUrl(endpoint, query = {}) {
  const basePath = endpoint.startsWith("/") ? endpoint : `/${endpoint}`;
  const url = new URL(`${API_BASE}${basePath}`, window.location.origin);

  Object.entries(query).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      url.searchParams.set(key, value);
    }
  });

  return url.toString();
}

function formatBytes(value) {
  const size = Number(value || 0);
  if (size === 0) return "0 B";

  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.min(Math.floor(Math.log(size) / Math.log(1024)), units.length - 1);
  const formatted = size / Math.pow(1024, i);

  return `${formatted >= 10 || i === 0 ? formatted.toFixed(0) : formatted.toFixed(1)} ${units[i]}`;
}

function formatDate(value) {
  if (!value) return "-";

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";

  return new Intl.DateTimeFormat("pt-BR", {
    dateStyle: "short",
    timeStyle: "short"
  }).format(date);
}

function parseDownloadFilename(contentDisposition, fallback) {
  if (!contentDisposition) return fallback;

  const utf8Match = /filename\*=UTF-8''([^;]+)/i.exec(contentDisposition);
  if (utf8Match?.[1]) {
    try {
      return decodeURIComponent(utf8Match[1]);
    } catch {
      return utf8Match[1];
    }
  }

  const asciiMatch = /filename="?([^";]+)"?/i.exec(contentDisposition);
  return asciiMatch?.[1] || fallback;
}

function folderDisplayName(fullPath) {
  const trimmed = fullPath.replace(/\/+$/, "");
  const lastSlash = trimmed.lastIndexOf("/");
  return lastSlash === -1 ? trimmed : trimmed.slice(lastSlash + 1);
}

function fileDisplayName(id) {
  const lastSlash = id.lastIndexOf("/");
  return lastSlash === -1 ? id : id.slice(lastSlash + 1);
}

function App() {
  const [bucketName, setBucketName] = useState("-");
  const [folders, setFolders] = useState([]);
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");
  const [downloadingId, setDownloadingId] = useState("");
  const [currentFolder, setCurrentFolder] = useState("");
  const [sortField, setSortField] = useState("name");
  const [sortAsc, setSortAsc] = useState(true);

  const fetchFolder = useCallback(async (folder) => {
    setLoading(true);
    setError("");

    try {
      const response = await fetch(
        buildApiUrl("/bucket/items/by-folder", { folder })
      );
      if (!response.ok) {
        throw new Error(`Falha ao buscar itens (${response.status}).`);
      }

      const payload = await response.json();
      setBucketName(payload.bucket || "-");
      setFolders(Array.isArray(payload.folders) ? payload.folders : []);
      setFiles(Array.isArray(payload.items) ? payload.items : []);
    } catch (requestError) {
      setError(requestError.message || "Erro ao consultar o BFF.");
      setFolders([]);
      setFiles([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchFolder(currentFolder);
  }, [currentFolder, fetchFolder]);

  function navigateTo(folder) {
    setQuery("");
    setCurrentFolder(folder);
  }

  function navigateUp() {
    const trimmed = currentFolder.replace(/\/+$/, "");
    const lastSlash = trimmed.lastIndexOf("/");
    if (lastSlash === -1) {
      navigateTo("");
    } else {
      navigateTo(trimmed.slice(0, lastSlash + 1));
    }
  }

  const breadcrumbs = useMemo(() => {
    if (!currentFolder) return [];
    const parts = currentFolder.replace(/\/+$/, "").split("/");
    return parts.map((part, i) => ({
      label: part,
      path: parts.slice(0, i + 1).join("/") + "/"
    }));
  }, [currentFolder]);

  function handleSort(field) {
    if (sortField === field) {
      setSortAsc(!sortAsc);
    } else {
      setSortField(field);
      setSortAsc(true);
    }
  }

  const sortIndicator = (field) => {
    if (sortField !== field) return "";
    return sortAsc ? " \u25B2" : " \u25BC";
  };

  const sortedItems = useMemo(() => {
    const folderItems = folders.map((f) => ({
      type: "folder",
      id: f,
      name: folderDisplayName(f),
      size: -1,
      content_type: "",
      updated_at: ""
    }));

    const fileItems = files.map((f) => ({
      type: "file",
      id: f.id,
      name: fileDisplayName(f.id),
      size: f.size,
      content_type: f.content_type,
      updated_at: f.updated_at
    }));

    const all = [...folderItems, ...fileItems];

    const normalizedQuery = query.trim().toLowerCase();
    const filtered = normalizedQuery
      ? all.filter((item) => item.name.toLowerCase().includes(normalizedQuery))
      : all;

    filtered.sort((a, b) => {
      // Folders always come first
      if (a.type !== b.type) return a.type === "folder" ? -1 : 1;

      let cmp = 0;
      if (sortField === "name") {
        cmp = a.name.localeCompare(b.name, "pt-BR", { sensitivity: "base" });
      } else if (sortField === "size") {
        cmp = (a.size || 0) - (b.size || 0);
      }
      return sortAsc ? cmp : -cmp;
    });

    return filtered;
  }, [folders, files, query, sortField, sortAsc]);

  async function handleDownload(file) {
    setDownloadingId(file.id);

    try {
      const response = await fetch(buildApiUrl("/bucket/items/file", { id: file.id }));
      if (!response.ok) {
        throw new Error(`Falha no download (${response.status}).`);
      }

      const blob = await response.blob();
      const contentDisposition = response.headers.get("content-disposition");
      const filename = parseDownloadFilename(contentDisposition, file.name);
      const href = window.URL.createObjectURL(blob);
      const anchor = document.createElement("a");
      anchor.href = href;
      anchor.download = filename;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      window.URL.revokeObjectURL(href);
    } catch (downloadError) {
      window.alert(downloadError.message || "Nao foi possivel baixar o arquivo.");
    } finally {
      setDownloadingId("");
    }
  }

  const hasNoResults = !loading && !error && sortedItems.length === 0;

  return (
    <div className="drive-shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="brand-dot" />
          <div>
            <strong>MIH Cloud</strong>
            <small>Storage BFF</small>
          </div>
        </div>

        <nav className="nav-list" aria-label="Navegacao">
          <button type="button" className="nav-item nav-item-active">
            Cloud
          </button>
        </nav>
      </aside>

      <main className="content">
        <header className="topbar">
          <div>
            <h1>Cloud Storage</h1>
            <p>
              Bucket: <strong>{bucketName}</strong>
            </p>
          </div>

          <div className="topbar-controls">
            <input
              type="search"
              placeholder="Buscar nesta pasta"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              aria-label="Buscar nesta pasta"
            />
            <button type="button" onClick={() => fetchFolder(currentFolder)}>
              Atualizar
            </button>
          </div>
        </header>

        <nav className="breadcrumb" aria-label="Caminho">
          <button
            type="button"
            className={!currentFolder ? "breadcrumb-item breadcrumb-active" : "breadcrumb-item"}
            onClick={() => navigateTo("")}
          >
            Raiz
          </button>
          {breadcrumbs.map((crumb, i) => (
            <span key={crumb.path}>
              <span className="breadcrumb-sep">/</span>
              <button
                type="button"
                className={
                  i === breadcrumbs.length - 1
                    ? "breadcrumb-item breadcrumb-active"
                    : "breadcrumb-item"
                }
                onClick={() => navigateTo(crumb.path)}
              >
                {crumb.label}
              </button>
            </span>
          ))}
        </nav>

        <section className="summary-cards" aria-label="Resumo">
          <article>
            <span>Pastas</span>
            <strong>{folders.length}</strong>
          </article>
          <article>
            <span>Arquivos</span>
            <strong>{files.length}</strong>
          </article>
          <article>
            <span>Tamanho total</span>
            <strong>{formatBytes(files.reduce((sum, f) => sum + Number(f.size || 0), 0))}</strong>
          </article>
        </section>

        {loading && <p className="status">Carregando...</p>}
        {error && <p className="status status-error">{error}</p>}
        {hasNoResults && <p className="status">Nenhum item encontrado.</p>}

        {!loading && !error && sortedItems.length > 0 && (
          <section className="files-section">
            <div className="files-table-wrap">
              <table>
                <thead>
                  <tr>
                    <th
                      className="sortable"
                      onClick={() => handleSort("name")}
                    >
                      Nome{sortIndicator("name")}
                    </th>
                    <th
                      className="sortable"
                      onClick={() => handleSort("size")}
                    >
                      Tamanho{sortIndicator("size")}
                    </th>
                    <th>Tipo</th>
                    <th>Atualizado em</th>
                    <th aria-label="Acoes" />
                  </tr>
                </thead>
                <tbody>
                  {currentFolder && (
                    <tr className="row-folder" onClick={navigateUp} style={{ cursor: "pointer" }}>
                      <td colSpan={5} className="back-row">
                        <span className="folder-icon" aria-hidden>&#128193;</span> ..
                      </td>
                    </tr>
                  )}
                  {sortedItems.map((item) =>
                    item.type === "folder" ? (
                      <tr
                        key={item.id}
                        className="row-folder"
                        onClick={() => navigateTo(item.id)}
                        style={{ cursor: "pointer" }}
                      >
                        <td>
                          <span className="folder-icon" aria-hidden>&#128193;</span>{" "}
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
                          <span className="file-icon" aria-hidden>&#128196;</span>{" "}
                          {item.name}
                        </td>
                        <td>{formatBytes(item.size)}</td>
                        <td>{item.content_type || "-"}</td>
                        <td>{formatDate(item.updated_at)}</td>
                        <td>
                          <button
                            type="button"
                            className="download-btn"
                            onClick={() => handleDownload(item)}
                            disabled={downloadingId === item.id}
                          >
                            {downloadingId === item.id ? "Baixando..." : "Download"}
                          </button>
                        </td>
                      </tr>
                    )
                  )}
                </tbody>
              </table>
            </div>
          </section>
        )}
      </main>
    </div>
  );
}

export default App;
