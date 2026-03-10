import type { BreadcrumbItem } from "../types/api";

interface BreadcrumbProps {
  currentFolder: string;
  breadcrumbs: BreadcrumbItem[];
  onNavigate: (folder: string) => void;
}

export function Breadcrumb({
  currentFolder,
  breadcrumbs,
  onNavigate,
}: BreadcrumbProps) {
  return (
    <nav className="breadcrumb" aria-label="Caminho">
      <button
        type="button"
        className={
          !currentFolder
            ? "breadcrumb-item breadcrumb-active"
            : "breadcrumb-item"
        }
        onClick={() => onNavigate("")}
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
            onClick={() => onNavigate(crumb.path)}
          >
            {crumb.label}
          </button>
        </span>
      ))}
    </nav>
  );
}
