import { REGIONS, REGION_ORDER } from '../constants.js';
import type { RegionId } from '../types/index.js';

interface RegionSelectorProps {
  selected: RegionId;
  onChange: (id: RegionId) => void;
}

export default function RegionSelector({ selected, onChange }: RegionSelectorProps) {
  return (
    <div className="region-selector" role="group" aria-label="סנן לפי אזור">
      {REGION_ORDER.map(id => (
        <button
          key={id}
          className={`region-chip${selected === id ? ' region-chip--active' : ''}`}
          onClick={() => onChange(id)}
          aria-pressed={selected === id}
        >
          {REGIONS[id].label}
        </button>
      ))}
    </div>
  );
}
