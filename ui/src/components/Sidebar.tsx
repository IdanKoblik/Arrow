import AlertCard from './AlertCard.js';
import HistoryList from './HistoryList.js';
import RegionSelector from './RegionSelector.js';
import { alertMatchesRegion } from '../constants.js';
import type { AlertData, RegionId } from '../types/index.js';

interface SidebarProps {
  alert: AlertData | null;
  history: AlertData[];
  isDemo: boolean;
  onLocClick: (name: string) => void;
  selectedRegion: RegionId;
  onRegionChange: (id: RegionId) => void;
}

export default function Sidebar({
  alert, history, isDemo, onLocClick, selectedRegion, onRegionChange,
}: SidebarProps) {
  const filteredHistory = history.filter(h =>
    alertMatchesRegion(h.data ?? [], selectedRegion)
  );

  const activeCount = alert
    ? (selectedRegion === 'all'
        ? alert.data?.length ?? 0
        : alert.data?.filter(c => alertMatchesRegion([c], selectedRegion)).length ?? 0)
    : 0;

  return (
    <aside className="sidebar">

        <RegionSelector selected={selectedRegion} onChange={onRegionChange} />

        <div className="active-wrap">
          <div className="section-header">
            <span>אזעקות פעילות</span>
            {activeCount > 0 && <span className="badge">{activeCount}</span>}
          </div>
          <div className="active-body">
            <AlertCard
              alert={alert}
              isDemo={isDemo}
              onLocClick={onLocClick}
              selectedRegion={selectedRegion}
            />
          </div>
        </div>

        <div className="hist-wrap">
          <div className="section-header">
            <span>היסטוריה</span>
            {filteredHistory.length > 0 && (
              <span className="hist-count">{filteredHistory.length}</span>
            )}
          </div>
          <div className="hist-body">
            <HistoryList items={filteredHistory} />
          </div>
        </div>
    </aside>
  );
}
