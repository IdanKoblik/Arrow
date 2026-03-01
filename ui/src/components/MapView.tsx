import { useEffect, useRef, forwardRef, useImperativeHandle } from 'react';
import L from 'leaflet';
import { LOCS, getCat } from '../constants.js';
import type { AlertData, MapViewHandle } from '../types/index.js';

function mkIcon(color: string): L.DivIcon {
  return L.divIcon({
    className: '',
    html: `<div style="width:18px;height:18px;border-radius:50%;background:${color};border:3px solid #fff;animation:ap 1.5s infinite"></div>
           <style>@keyframes ap{0%{box-shadow:0 0 0 2px ${color}88}60%{box-shadow:0 0 0 14px ${color}11}100%{box-shadow:0 0 0 2px ${color}88}}</style>`,
    iconSize: [18, 18],
    iconAnchor: [9, 9],
    popupAnchor: [0, -12],
  });
}

const MapView = forwardRef<MapViewHandle>(function MapView(_props, ref) {
  const containerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<L.Map | null>(null);
  const markersRef = useRef<L.Marker[]>([]);

  useEffect(() => {
    if (!containerRef.current) return;
    const map = L.map(containerRef.current).setView([31.5, 34.9], 8);
    L.tileLayer('https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png', {
      attribution:
        '&copy; <a href="https://openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/attributions">CARTO</a>',
      subdomains: 'abcd',
      maxZoom: 19,
    }).addTo(map);
    mapRef.current = map;
    return () => { map.remove(); mapRef.current = null; };
  }, []);

  useImperativeHandle(ref, () => ({
    clearMarkers() {
      markersRef.current.forEach(m => mapRef.current?.removeLayer(m));
      markersRef.current = [];
    },
    placeMarkers(alertData: AlertData) {
      const c = getCat(alertData.cat);
      for (const name of alertData.data ?? []) {
        const ll = LOCS[name];
        if (!ll) continue;
        const m = L.marker(ll, { icon: mkIcon(c.color) }).addTo(mapRef.current!);
        m.bindPopup(
          `<div dir="rtl" style="text-align:right;min-width:120px">
            <strong style="display:block;font-size:0.93em">${name}</strong>
            <div style="color:${c.color};font-size:0.8em;margin-top:2px">${c.icon} ${alertData.title}</div>
            ${alertData.desc ? `<div style="color:#888;font-size:0.77em;margin-top:1px">${alertData.desc}</div>` : ''}
          </div>`
        );
        markersRef.current.push(m);
      }
      if (markersRef.current.length > 0) {
        try {
          mapRef.current?.fitBounds(
            L.featureGroup(markersRef.current).getBounds().pad(0.35),
            { maxZoom: 12, animate: false }
          );
        } catch { /* ignore */ }
      }
    },
    focusLoc(name: string) {
      const ll = LOCS[name];
      if (!ll || !mapRef.current) return;
      mapRef.current.setView(ll, 14, { animate: true });
      for (const m of markersRef.current) {
        const p = m.getLatLng();
        if (Math.abs(p.lat - ll[0]) < 0.002 && Math.abs(p.lng - ll[1]) < 0.002) {
          m.openPopup();
          break;
        }
      }
    },
    invalidateSize() {
      mapRef.current?.invalidateSize();
    },
  }));

  return <div ref={containerRef} className="map-container" />;
});

export default MapView;
