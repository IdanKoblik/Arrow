import { getCat } from '../constants.js';
import type { AlertData } from '../types/index.js';

function relTime(iso?: string): string {
  if (!iso) return '';
  try {
    const d = Date.now() - new Date(iso).getTime();
    const m = Math.floor(d / 60000);
    if (m < 1) return 'עכשיו';
    if (m < 60) return `לפני ${m} דק'`;
    if (m < 1440) return `לפני ${Math.floor(m / 60)} ש'`;
    return new Date(iso).toLocaleDateString('he-IL', { day: 'numeric', month: 'short' });
  } catch {
    return '';
  }
}

interface HistoryListProps {
  items: AlertData[];
}

export default function HistoryList({ items }: HistoryListProps) {
  if (!items.length) {
    return <div className="hist-empty">ממתין לאזעקות...</div>;
  }

  return (
    <div>
      {items.map((h, i) => {
        const c = getCat(h.cat);
        const shown = (h.data ?? []).slice(0, 4).join('، ');
        const more = (h.data ?? []).length > 4 ? ` ועוד ${h.data!.length - 4}` : '';
        return (
          <div key={h.id ?? i} className="hist-item">
            <div className="hist-dot" style={{ background: c.color }} />
            <div className="hist-info">
              <div className="hist-title">
                {h.demo && <span className="hdemo">דמו</span>}
                {c.icon} {h.title}
              </div>
              <div className="hist-locs">{shown}{more}</div>
            </div>
            <div className="hist-time">{relTime(h.alertDate ?? h.savedAt)}</div>
          </div>
        );
      })}
    </div>
  );
}
