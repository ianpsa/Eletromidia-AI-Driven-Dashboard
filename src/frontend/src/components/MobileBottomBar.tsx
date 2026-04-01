import { IconAgent, IconAnalytics, IconCloud, IconDashboard } from "./NavIcons";

type Page = "dashboard" | "analytics" | "cloud";

interface MobileBottomBarProps {
  activePage: Page;
  onNavigate: (page: Page) => void;
  onToggleChat: () => void;
  chatOpen: boolean;
}

const items: readonly { page: Page; label: string; Icon: React.FC }[] = [
  { page: "dashboard", label: "Dashboard", Icon: IconDashboard },
  { page: "analytics", label: "Analytics", Icon: IconAnalytics },
  { page: "cloud", label: "Cloud", Icon: IconCloud },
];

export function MobileBottomBar({
  activePage,
  onNavigate,
  onToggleChat,
  chatOpen,
}: MobileBottomBarProps) {
  return (
    <nav className="mobile-bottom-bar" aria-label="Navegacao mobile">
      {items.map(({ page, label, Icon }) => (
        <button
          key={page}
          type="button"
          className={`mobile-bottom-bar__item ${activePage === page ? "mobile-bottom-bar__item--active" : ""}`}
          onClick={() => onNavigate(page)}
        >
          <Icon />
          <span>{label}</span>
        </button>
      ))}
      <button
        type="button"
        className={`mobile-bottom-bar__item ${chatOpen ? "mobile-bottom-bar__item--active" : ""}`}
        onClick={onToggleChat}
      >
        <IconAgent />
        <span>Chat</span>
      </button>
    </nav>
  );
}
