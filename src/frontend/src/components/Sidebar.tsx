import {
  IconAgent,
  IconAnalytics,
  IconCloud,
  IconCollapse,
  IconDashboard,
} from "./NavIcons";

type Page = "dashboard" | "analytics" | "cloud";

interface SidebarProps {
  activePage: Page;
  onNavigate: (page: Page) => void;
  onToggleChat: () => void;
  chatOpen: boolean;
  collapsed: boolean;
  onToggleCollapse: () => void;
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
          className={`nav-item ${activePage === "analytics" ? "nav-item-active" : ""}`}
          onClick={() => onNavigate("analytics")}
          title="Analytics"
        >
          <IconAnalytics />
          {!collapsed && <span>Analytics</span>}
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
