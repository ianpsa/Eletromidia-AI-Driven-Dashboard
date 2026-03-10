export function Sidebar() {
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
        <button type="button" className="nav-item nav-item-active">
          Cloud
        </button>
      </nav>
    </aside>
  );
}
