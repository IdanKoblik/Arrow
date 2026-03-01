// ─── Alert / API shapes ───────────────────────────────────────────────────────

export interface AlertData {
  id: string;
  cat: string;        // numeric string, e.g. "1", "13"
  title: string;
  data: string[];     // city/location names in Hebrew
  desc?: string;
  alertDate?: string; // ISO-8601, present on history items
  savedAt?: string;   // ISO-8601, fallback timestamp
  demo?: boolean;
}

// ─── Category ─────────────────────────────────────────────────────────────────

export interface Category {
  icon: string;
  label: string;
  color: string;
  chipBg: string;
  chipText: string;
  chipBorder: string;
}

// ─── Location coordinates ─────────────────────────────────────────────────────

export type LatLng = [number, number];
export type LocsMap = Record<string, LatLng>;

// ─── Region ───────────────────────────────────────────────────────────────────

export type RegionId =
  | 'all'
  | 'north'
  | 'haifa'
  | 'valleys'
  | 'sharon'
  | 'gush-dan'
  | 'center'
  | 'jerusalem'
  | 'south'
  | 'otef';

export interface Region {
  id: RegionId;
  label: string;
  cities: ReadonlySet<string>;
}

export type RegionsMap = Record<RegionId, Region>;

// ─── Connection status ────────────────────────────────────────────────────────

export interface ConnectionStatus {
  cls: '' | 'err' | 'conn';
  txt: string;
}

// ─── MapView imperative handle ────────────────────────────────────────────────

export interface MapViewHandle {
  clearMarkers: () => void;
  placeMarkers: (alertData: AlertData) => void;
  focusLoc: (name: string) => void;
  invalidateSize: () => void;
}

// ─── Hook options / return types ──────────────────────────────────────────────

export interface UseAlertsOptions {
  onNewAlert: (data: AlertData | null, breakDemo: boolean) => void;
  isDemo: boolean;
}

export interface UseAlertsReturn {
  status: ConnectionStatus;
  lastUpdate: string;
}

export interface UseSoundReturn {
  soundOn: boolean;
  toggleSound: () => void;
  playAlert: () => void;
}
