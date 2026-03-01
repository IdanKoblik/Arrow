import { getCat, cityInRegion } from '../constants.js';
import type { AlertData, RegionId } from '../types/index.js';

interface AlertCardProps {
  alert: AlertData | null;
  isDemo: boolean;
  onLocClick: (name: string) => void;
  selectedRegion: RegionId;
}

export default function AlertCard({ alert, isDemo, onLocClick, selectedRegion }: AlertCardProps) {
  if (!alert) {
    return (
      <div className="ok">
        <div className="ok-icon">✓</div>
        אין אזעקות פעילות כרגע
      </div>
    );
  }

  const c = getCat(alert.cat);
  const allLocs = alert.data ?? [];
  const locs = allLocs.filter(loc => cityInRegion(loc, selectedRegion));
  const hiddenCount = allLocs.length - locs.length;

  return (
    <div className="card" style={{ '--cat-color': c.color } as React.CSSProperties}>
      <div className="card-head" style={{ background: c.color }}>
        <span className="cat-icon">{c.icon}</span>
        <div className="card-head-text">
          <div className="card-title">
            {isDemo && <span className="demo-tag">דמו</span>}
            {alert.title}
          </div>
          {alert.desc && <div className="card-desc">{alert.desc}</div>}
        </div>
        <span className="card-count">{locs.length} יישובים</span>
      </div>
      <div className="card-locs">
        {locs.length === 0 ? (
          <span className="no-region-match">
            {hiddenCount > 0
              ? `אין יישובים מאזור זה (${hiddenCount} ממוסננים)`
              : 'אין יישובים בהתרעה'}
          </span>
        ) : (
          locs.map(loc => (
            <button
              key={loc}
              className="chip"
              style={{ background: c.chipBg, color: c.chipText, borderColor: c.chipBorder }}
              onClick={() => onLocClick(loc)}
            >
              {loc}
            </button>
          ))
        )}
        {locs.length > 0 && hiddenCount > 0 && (
          <span className="hidden-count">+{hiddenCount} באזורים אחרים</span>
        )}
      </div>
    </div>
  );
}
