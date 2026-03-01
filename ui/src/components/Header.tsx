import type { ConnectionStatus } from '../types/index.js';

interface HeaderProps {
  status: ConnectionStatus;
  lastUpdate: string;
  soundOn: boolean;
  onToggleSound: () => void;
  onLoadDemo: () => void;
  activeCount: number;
  onToggleSidebar: () => void;
}

export default function Header({
  status, lastUpdate, soundOn, onToggleSound, onLoadDemo, activeCount, onToggleSidebar,
}: HeaderProps) {
  return (
    <header className="header">
      <div className="brand">
        <div className="logo">🚨</div>
        <div className="brand-text">
          <div className="brand-t1">מפת אזעקות חיות</div>
          <div className="brand-t2">עדכון בזמן אמת · פיקוד העורף</div>
        </div>
      </div>

      <div className="header-right">
        <span className="upd">{lastUpdate}</span>
        <div className="status-indicator">
          <div className={`sdot ${status.cls}`} />
          <span className="stxt">{status.txt}</span>
        </div>
        <button className="btn btn-sound" onClick={onToggleSound}>
          {soundOn ? '🔊' : '🔇'}<span className="btn-label"> קול</span>
        </button>
        <button className="btn btn-demo" onClick={onLoadDemo}>
          <span className="btn-label">דמו</span>
        </button>
        <button
          className="btn btn-sidebar-toggle"
          onClick={onToggleSidebar}
          aria-label="הצג/הסתר התרעות"
        >
          {activeCount > 0
            ? <span className="alert-badge-btn">{activeCount}</span>
            : '☰'}
        </button>
      </div>
    </header>
  );
}
