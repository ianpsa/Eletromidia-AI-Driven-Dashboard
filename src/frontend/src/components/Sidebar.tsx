type Page = "dashboard" | "cloud" | "agent";

interface SidebarProps {
  activePage: Page;
  onNavigate: (page: Page) => void;
}

export function Sidebar({ activePage, onNavigate }: SidebarProps) {
  return (
    <aside className="sidebar">
      <div className="brand">
        <span className="brand-dot" />
        <div>
          <strong>Eletromidia</strong>
        </div>
      </div>

      <nav className="nav-list" aria-label="Navegacao">
        <button
          type="button"
          className={`nav-item ${activePage === "dashboard" ? "nav-item-active" : ""}`}
          onClick={() => onNavigate("dashboard")}
        >
          Dashboard
        </button>
        <button
          type="button"
          className={`nav-item ${activePage === "cloud" ? "nav-item-active" : ""}`}
          onClick={() => onNavigate("cloud")}
        >
          Cloud
        </button>

        <button
          type="button"
          className={`nav-item ${activePage === "agent" ? "nav-item-active" : ""}`}
          onClick={() => onNavigate("agent")}
        >
          Agente IA
        </button>
      </nav>
    </aside>
  );
}

export default Sidebar;
