"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { Link2 } from "lucide-react";
import { getStats, type StatsResponse } from "@/lib/api";

function AnimatedNumber({ value }: { value: number }) {
  const [display, setDisplay] = useState(0);

  useEffect(() => {
    const duration = 600;
    const start = display;
    const diff = value - start;
    if (diff === 0) return;

    const startTime = Date.now();
    const tick = () => {
      const elapsed = Date.now() - startTime;
      const progress = Math.min(elapsed / duration, 1);
      const eased = 1 - Math.pow(1 - progress, 3);
      setDisplay(Math.floor(start + diff * eased));
      if (progress < 1) requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  }, [value]);

  return <span>{display.toLocaleString("pt-BR")}</span>;
}

const cards = [
  {
    key: "urls_shortened" as const,
    label: "URLs encurtadas",
    icon: Link2,
    gradient: "from-violet-500/20 to-violet-500/5",
    border: "border-violet-500/20",
    iconColor: "text-violet-400",
    numberColor: "text-violet-300",
  },
];

export default function StatsCards() {
  const [stats, setStats] = useState<StatsResponse | null>(null);

  useEffect(() => {
    const fetchStats = () => getStats().then(setStats).catch(() => {});
    fetchStats();
    const interval = setInterval(fetchStats, 3000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="grid grid-cols-1 gap-4 max-w-xs">
      {cards.map((card, i) => {
        const Icon = card.icon;
        const value = stats ? Math.round(stats[card.key]) : 0;

        return (
          <motion.div
            key={card.key}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 * i, duration: 0.4 }}
            className={`relative overflow-hidden rounded-2xl border ${card.border} bg-gradient-to-br ${card.gradient} p-5`}
          >
            <div className="flex items-center gap-2 mb-3">
              <Icon className={`w-4 h-4 ${card.iconColor}`} />
              <span className="text-xs font-medium text-zinc-400 uppercase tracking-wider">
                {card.label}
              </span>
            </div>
            <div className={`text-3xl font-bold ${card.numberColor} font-mono`}>
              <AnimatedNumber value={value} />
            </div>
          </motion.div>
        );
      })}
    </div>
  );
}
