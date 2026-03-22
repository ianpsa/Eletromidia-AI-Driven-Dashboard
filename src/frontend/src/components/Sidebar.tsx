type Page = "dashboard" | "cloud";

interface SidebarProps {
  activePage: Page;
  onNavigate: (page: Page) => void;
  onToggleChat: () => void;
  chatOpen: boolean;
  collapsed: boolean;
  onToggleCollapse: () => void;
}

function IconDashboard() {
  return (
    <svg
      className="nav-icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect x="3" y="3" width="7" height="7" rx="1" />
      <rect x="14" y="3" width="7" height="7" rx="1" />
      <rect x="3" y="14" width="7" height="7" rx="1" />
      <rect x="14" y="14" width="7" height="7" rx="1" />
    </svg>
  );
}

function IconCloud() {
  return (
    <svg
      className="nav-icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z" />
    </svg>
  );
}

function IconAgent() {
  return (
    <svg
      className="nav-icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
    </svg>
  );
}

function IconCollapse() {
  return (
    <svg
      className="nav-icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <line x1="3" y1="12" x2="21" y2="12" />
      <line x1="3" y1="6" x2="21" y2="6" />
      <line x1="3" y1="18" x2="21" y2="18" />
    </svg>
  );
}

export function Sidebar({
  activePage,
  onNavigate,
  onToggleChat,
  chatOpen,
  collapsed,
  onToggleCollapse,
}: SidebarProps) {
  return (
    <aside className={`sidebar ${collapsed ? "sidebar-collapsed" : ""}`}>
      <div className="brand">
        <img
          src="/eletromidia-icon.png"
          alt="Eletromidia"
          className="brand-logo"
        />
        {!collapsed && (
          <div>
            <strong>Eletromidia</strong>
          </div>
        )}
      </div>

      <nav className="nav-list" aria-label="Navegacao">
        <button
          type="button"
          className={`nav-item ${activePage === "dashboard" ? "nav-item-active" : ""}`}
          onClick={() => onNavigate("dashboard")}
          title="Dashboard"
        >
          <IconDashboard />
          {!collapsed && <span>Dashboard</span>}
        </button>
        <button
          type="button"
          className={`nav-item ${activePage === "cloud" ? "nav-item-active" : ""}`}
          onClick={() => onNavigate("cloud")}
          title="Cloud"
        >
          <IconCloud />
          {!collapsed && <span>Cloud</span>}
        </button>
      </nav>

      <button
        type="button"
        className={`nav-item chat-toggle-btn ${chatOpen ? "chat-toggle-active" : ""}`}
        onClick={onToggleChat}
        title="Agente IA"
      >
        <IconAgent />
        {!collapsed && <span>Agente IA</span>}
      </button>

      <button
        type="button"
        className="nav-item sidebar-collapse-btn"
        onClick={onToggleCollapse}
        title={collapsed ? "Expandir menu" : "Recolher menu"}
      >
        <IconCollapse />
        {!collapsed && <span>Recolher</span>}
      </button>
    </aside>
  );
}

export default Sidebar;
