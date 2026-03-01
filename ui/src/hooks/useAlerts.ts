import { useEffect, useRef, useState } from 'react';
import type { AlertData, ConnectionStatus, UseAlertsOptions, UseAlertsReturn } from '../types/index.js';

const ALERTS_URL = '/api/alerts';
const HISTORY_URL = '/api/history';
const POLL_MS = 1000;

export function useAlerts({ onNewAlert, isDemo }: UseAlertsOptions): UseAlertsReturn {
  const [status, setStatus] = useState<ConnectionStatus>({ cls: 'conn', txt: 'מתחבר...' });
  const [lastUpdate, setLastUpdate] = useState('');
  const lastIdRef = useRef<string | null>(null);
  const pollingRef = useRef(false);
  const isDemoRef = useRef(isDemo);

  useEffect(() => {
    isDemoRef.current = isDemo;
  }, [isDemo]);

  async function poll() {
    if (pollingRef.current) return;
    pollingRef.current = true;
    try {
      const res = await fetch(ALERTS_URL, { cache: 'no-store' });
      const text = (await res.text()).trim().replace(/^\uFEFF/, '');
      const data: AlertData | null =
        text && text !== 'null' && text !== '\r\n' ? (JSON.parse(text) as AlertData) : null;

      setStatus({ cls: '', txt: 'מחובר' });
      setLastUpdate(new Date().toLocaleTimeString('he-IL'));

      const id = data?.id ?? null;

      if (isDemoRef.current) {
        if (data) onNewAlert(data, true);
        return;
      }

      if (id !== lastIdRef.current) {
        lastIdRef.current = id;
        onNewAlert(data, false);
      }
    } catch {
      setStatus({ cls: 'err', txt: 'שגיאת חיבור' });
    } finally {
      pollingRef.current = false;
    }
  }

  useEffect(() => {
    poll();
    const iv = setInterval(poll, POLL_MS);
    return () => clearInterval(iv);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return { status, lastUpdate };
}

export async function fetchHistory(): Promise<AlertData[]> {
  try {
    const res = await fetch(HISTORY_URL, { cache: 'no-store' });
    const data = (await res.json()) as AlertData[];
    return Array.isArray(data) ? data : [];
  } catch {
    return [];
  }
}
