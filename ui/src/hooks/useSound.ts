import { useRef, useState } from 'react';
import type { UseSoundReturn } from '../types/index.js';

export function useSound(): UseSoundReturn {
  const ctxRef = useRef<AudioContext | null>(null);
  const [soundOn, setSoundOn] = useState(true);

  function playAlert() {
    if (!soundOn) return;
    try {
      if (!ctxRef.current) {
        ctxRef.current = new (window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext)();
      }
      const ctx = ctxRef.current;
      ([880, 660, 880, 660] as number[]).forEach((hz, i) => {
        const o = ctx.createOscillator();
        const g = ctx.createGain();
        o.connect(g);
        g.connect(ctx.destination);
        o.frequency.value = hz;
        o.type = 'sine';
        const t = ctx.currentTime + i * 0.26;
        g.gain.setValueAtTime(0.28, t);
        g.gain.exponentialRampToValueAtTime(0.001, t + 0.22);
        o.start(t);
        o.stop(t + 0.26);
      });
    } catch { /* ignore */ }
  }

  function toggleSound() {
    setSoundOn(prev => !prev);
  }

  return { soundOn, toggleSound, playAlert };
}
