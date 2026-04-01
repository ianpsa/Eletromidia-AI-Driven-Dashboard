import { useCallback, useEffect, useMemo, useState } from "react";
import { useAuth } from "../AuthContext";
import type { ObjectItem } from "../types/api";
import { fileDisplayName, parseDownloadFilename } from "../utils/filename";
import { buildApiUrl } from "../utils/url";

export function useBucketFolder() {
  const [bucketName, setBucketName] = useState("-");
  const [folders, setFolders] = useState<string[]>([]);
  const [files, setFiles] = useState<ObjectItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [currentFolder, setCurrentFolder] = useState("");
  const [downloadingId, setDownloadingId] = useState("");
  const [downloadError, setDownloadError] = useState("");
  const { getToken } = useAuth();

  const breadcrumbs = useMemo(() => {
    if (!currentFolder) return [];
    const parts = currentFolder.replace(/\/+$/, "").split("/");
    return parts.map((part, i) => ({
      label: part,
      path: `${parts.slice(0, i + 1).join("/")}/`,
    }));
  }, [currentFolder]);

  const fetchFolder = useCallback(
    async (folder: string) => {
      setLoading(true);
      setError("");

      try {
        const token = await getToken();
        const response = await fetch(
          buildApiUrl("/bucket/items/by-folder", { folder }),
          {
            headers: { Authorization: `Bearer ${token}` },
          },
        );
        if (!response.ok) {
          throw new Error(`Falha ao buscar itens (${response.status}).`);
        }

        const payload = (await response.json()) as {
          bucket?: string;
          folders?: string[];
          items?: ObjectItem[];
        };
        setBucketName(payload.bucket ?? "-");
        setFolders(Array.isArray(payload.folders) ? payload.folders : []);
        setFiles(Array.isArray(payload.items) ? payload.items : []);
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Erro ao consultar o servidor.";
        setError(message);
        setFolders([]);
        setFiles([]);
      } finally {
        setLoading(false);
      }
    },
    [getToken],
  );

  useEffect(() => {
    void fetchFolder(currentFolder);
  }, [currentFolder, fetchFolder]);

  function navigateTo(folder: string) {
    setCurrentFolder(folder);
  }

  function navigateUp() {
    const trimmed = currentFolder.replace(/\/+$/, "");
    const lastSlash = trimmed.lastIndexOf("/");
    navigateTo(lastSlash === -1 ? "" : trimmed.slice(0, lastSlash + 1));
  }

  async function handleDownload(file: { id: string; name: string }) {
    setDownloadingId(file.id);
    setDownloadError("");

    try {
      const token = await getToken();
      const response = await fetch(
        buildApiUrl("/bucket/items/file", { id: file.id }),
        {
          headers: { Authorization: `Bearer ${token}` },
        },
      );
      if (!response.ok) {
        throw new Error(`Falha no download (${response.status}).`);
      }

      const blob = await response.blob();
      const contentDisposition = response.headers.get("content-disposition");
      const filename = parseDownloadFilename(
        contentDisposition,
        file.name || fileDisplayName(file.id),
      );
      const href = window.URL.createObjectURL(blob);
      const anchor = document.createElement("a");
      anchor.href = href;
      anchor.download = filename;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      window.URL.revokeObjectURL(href);
    } catch (err) {
      const message =
        err instanceof Error
          ? err.message
          : "Nao foi possivel baixar o arquivo.";
      setDownloadError(message);
    } finally {
      setDownloadingId("");
    }
  }

  return {
    bucketName,
    folders,
    files,
    loading,
    error,
    currentFolder,
    downloadingId,
    downloadError,
    breadcrumbs,
    navigateTo,
    navigateUp,
    refresh: () => {
      void fetchFolder(currentFolder);
    },
    handleDownload,
  };
}
