import {
  IconAgent,
  IconAnalytics,
  IconCloud,
  IconCollapse,
  IconDashboard,
} from "./NavIcons";
import { useAuth } from "../AuthContext";

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
  const { signOut, user } = useAuth();

  return (
    <aside className={`sidebar ${collapsed ? "sidebar-collapsed" : ""}`}>
      {/* Brand */}
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

      {/* Navigation */}
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

      {/* Agent AI — pushed to bottom via margin-top: auto no CSS */}
      <button
        type="button"
        className={`nav-item chat-toggle-btn ${chatOpen ? "chat-toggle-active" : ""}`}
        onClick={onToggleChat}
        title="Agente IA"
      >
        <IconAgent />
        {!collapsed && <span>Agente IA</span>}
      </button>

      {/* Footer: user + collapse + logout */}
      <div className="sidebar-footer">
        <div className="sidebar-footer-top">
          {/* Avatar + email */}
          <div className="sidebar-user-info" title={user?.email ?? ""}>
            <div className="sidebar-user-avatar">
              {user?.email?.[0].toUpperCase() ?? "?"}
            </div>
            {!collapsed && (
              <span className="sidebar-user-email">{user?.email}</span>
            )}
          </div>
        </div>

        <div className="sidebar-footer-actions">
          <button
            type="button"
            className="nav-item sidebar-collapse-btn"
            onClick={onToggleCollapse}
            title={collapsed ? "Expandir menu" : "Recolher menu"}
          >
            <IconCollapse />
            {!collapsed && <span>Recolher</span>}
          </button>

          <button
            type="button"
            className="nav-item sidebar-logout-btn"
            onClick={() => void signOut()}
            title="Sair da conta"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="nav-icon"
              aria-hidden="true"
            >
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
              <polyline points="16 17 21 12 16 7" />
              <line x1="21" y1="12" x2="9" y2="12" />
            </svg>
            {!collapsed && <span>Sair</span>}
          </button>
        </div>
      </div>
    </aside>
  );
}

export default Sidebar;