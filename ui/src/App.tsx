import { useCallback, useEffect, useRef, useState } from 'react';
import Header from './components/Header.js';
import MapView from './components/MapView.js';
import Sidebar from './components/Sidebar.js';
import { useAlerts, fetchHistory } from './hooks/useAlerts.js';
import { useSound } from './hooks/useSound.js';
import { DEMO_ALERTS, REGIONS, cityInRegion } from './constants.js';
import type { AlertData, MapViewHandle, RegionId } from './types/index.js';

export default function App() {
  const [alert, setAlert] = useState<AlertData | null>(null);
  const [history, setHistory] = useState<AlertData[]>([]);
  const [isDemo, setIsDemo] = useState(false);
  const [selectedRegion, setSelectedRegion] = useState<RegionId>('all');

  const mapRef = useRef<MapViewHandle | null>(null);
  const isDemoRef = useRef(false);
  const demoIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const demoIndexRef = useRef(0);

  const { soundOn, toggleSound, playAlert } = useSound();

  const notify = useCallback((data: AlertData) => {
    if (Notification.permission === 'granted') {
      new Notification(`🚨 ${data.title}`, { body: (data.data ?? []).slice(0, 3).join(', ') });
    }
  }, []);

  // Applies an alert to state (does NOT place markers — that's handled by the effect below)
  const applyAlert = useCallback((data: AlertData | null) => {
    setAlert(data);
    if (data) {
      playAlert();
      notify(data);
    }
  }, [playAlert, notify]);

  const loadHistoryData = useCallback(async () => {
    const items = await fetchHistory();
    setHistory(items);
  }, []);

  const handleNewAlert = useCallback((data: AlertData | null, breakDemo: boolean) => {
    if (isDemoRef.current) {
      if (breakDemo && data) {
        isDemoRef.current = false;
        setIsDemo(false);
        if (demoIntervalRef.current) clearInterval(demoIntervalRef.current);
        applyAlert(data);
        void loadHistoryData();
      }
      return;
    }
    applyAlert(data);
    if (data) void loadHistoryData();
  }, [applyAlert, loadHistoryData]);

  const { status, lastUpdate } = useAlerts({ onNewAlert: handleNewAlert, isDemo });

  // Reactively re-render map markers whenever alert or region filter changes
  useEffect(() => {
    mapRef.current?.clearMarkers();
    if (!alert) return;
    const filteredData: AlertData = {
      ...alert,
      data: selectedRegion === 'all'
        ? alert.data
        : alert.data.filter(c => REGIONS[selectedRegion].cities.has(c)),
    };
    if (filteredData.data.length > 0) {
      mapRef.current?.placeMarkers(filteredData);
    }
  }, [alert, selectedRegion]);

  useEffect(() => {
    void loadHistoryData();
    const iv = setInterval(() => void loadHistoryData(), 60_000);
    return () => clearInterval(iv);
  }, [loadHistoryData]);

  useEffect(() => {
    if ('Notification' in window && Notification.permission === 'default') {
      void Notification.requestPermission();
    }
  }, []);

  function showDemoAlert(data: AlertData, index: number) {
    demoIndexRef.current = index;
    applyAlert(data);
  }

  function loadDemo() {
    isDemoRef.current = true;
    setIsDemo(true);
    demoIndexRef.current = 0;
    showDemoAlert(DEMO_ALERTS[0], 0);

    if (demoIntervalRef.current) clearInterval(demoIntervalRef.current);
    demoIntervalRef.current = setInterval(() => {
      const next = (demoIndexRef.current + 1) % DEMO_ALERTS.length;
      showDemoAlert(DEMO_ALERTS[next], next);
    }, 4000);
  }

  function handleLocClick(name: string) {
    mapRef.current?.focusLoc(name);
  }

  function handleRegionChange(id: RegionId) {
    setSelectedRegion(id);
  }

  const activeCount = alert
    ? (selectedRegion === 'all'
        ? alert.data?.length ?? 0
        : alert.data?.filter(c => cityInRegion(c, selectedRegion)).length ?? 0)
    : 0;

  return (
    <div className="app">
      <Header
        status={status}
        lastUpdate={lastUpdate}
        soundOn={soundOn}
        onToggleSound={toggleSound}
        onLoadDemo={loadDemo}
        activeCount={activeCount}
        onToggleSidebar={() => {}}
      />

      {isDemo && (
        <div className="demo-banner">⚠️ מצב דמו — הנתונים אינם אמיתיים</div>
      )}

      <div className="body">
        <MapView ref={mapRef} />
        <Sidebar
          alert={alert}
          history={history}
          isDemo={isDemo}
          onLocClick={handleLocClick}
          selectedRegion={selectedRegion}
          onRegionChange={handleRegionChange}
        />
      </div>
    </div>
  );
}
