interface NavigationItem {
    id: string;
    label: string;
    hint: string;
    icon: string;
}

interface NavigationRailProps {
    activeId: string;
    items: NavigationItem[];
    onNavigate: (id: string) => void;
}

export function NavigationRail({activeId, items, onNavigate}: NavigationRailProps) {
    return (
        <aside className="nav-rail">
            <div className="nav-brand">
                <span className="nav-brand-mark">MS</span>
                <div>
                    <p className="nav-brand-title">maliang swarm</p>
                    <p className="nav-brand-subtitle">多智能体工作台</p>
                </div>
            </div>

            <nav className="nav-menu" aria-label="主导航">
                {items.map((item) => (
                    <button
                        key={item.id}
                        className={`nav-button ${activeId === item.id ? "active" : ""}`}
                        onClick={() => onNavigate(item.id)}
                        type="button"
                    >
                        <span className="nav-button-icon" aria-hidden="true">{item.icon}</span>
                        <span className="nav-button-label">{item.label}</span>
                        <span className="nav-button-hint">{item.hint}</span>
                    </button>
                ))}
            </nav>
        </aside>
    );
}
