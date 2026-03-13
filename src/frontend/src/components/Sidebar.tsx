type Props = {
  onNavigate?: (route: string) => void;
  active?: string;
};

export function Sidebar({ onNavigate, active = "cloud" }: Props) {
  return (
    <aside className="sidebar">
      <div className="brand">
        <span className="brand-dot" />
        <div>
          <strong>MIH Cloud</strong>
          <small>Storage BFF</small>
        </div>
      </div>

      <nav className="nav-list" aria-label="Navegacao">
        <button
          type="button"
          className={`nav-item ${active === "cloud" ? "nav-item-active" : ""}`}
          onClick={() => onNavigate?.("cloud")}
        >
          Cloud
        </button>

        <button
          type="button"
          className={`nav-item ${active === "agent" ? "nav-item-active" : ""}`}
          onClick={() => onNavigate?.("agent")}
        >
          Agente IA
        </button>
      </nav>
    </aside>
  );
}

export default Sidebar;
