import ShortenForm from "@/components/shorten-form";
import StatsCards from "@/components/stats-cards";
import TechBadge from "@/components/tech-badge";
import { Link2 } from "lucide-react";

export default function Home() {
  return (
    <main className="flex-1 flex flex-col">
      <div className="fixed inset-0 -z-10">
        <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[600px] bg-violet-600/8 rounded-full blur-[120px]" />
        <div className="absolute bottom-0 right-0 w-[400px] h-[400px] bg-emerald-600/5 rounded-full blur-[100px]" />
      </div>

      <header className="border-b border-zinc-800/50 backdrop-blur-sm">
        <div className="max-w-4xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-violet-600 flex items-center justify-center">
              <Link2 className="w-4 h-4 text-white" />
            </div>
            <span className="font-bold text-lg tracking-tight">snip</span>
          </div>
          <div className="flex items-center gap-2">
            <TechBadge name="Go" />
            <TechBadge name="Redis" />
            <TechBadge name="Cassandra" />
          </div>
        </div>
      </header>

      <div className="flex-1 flex flex-col justify-center">
        <div className="max-w-4xl mx-auto px-6 py-16 w-full space-y-16">
          <section className="space-y-6">
            <div className="space-y-3">
              <h1 className="text-5xl sm:text-6xl font-bold tracking-tight">
                Links curtos,{" "}
                <span className="bg-gradient-to-r from-violet-400 to-violet-300 bg-clip-text text-transparent">
                  resultados grandes.
                </span>
              </h1>
              <p className="text-lg text-zinc-400 max-w-xl">
                Encurte suas URLs com alta performance. Arquitetura distribuida
                com Redis para IDs atomicos e Cassandra para escala massiva.
              </p>
            </div>
            <ShortenForm />
          </section>

          <section className="space-y-4">
            <div className="flex items-center gap-2">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
              <h2 className="text-sm font-medium text-zinc-400 uppercase tracking-wider">
                Metricas em tempo real
              </h2>
            </div>
            <StatsCards />
          </section>
        </div>
      </div>

      <footer className="border-t border-zinc-800/50">
        <div className="max-w-4xl mx-auto px-6 py-4 flex items-center justify-between text-xs text-zinc-600">
          <span>System Design inspirado no video &quot;Arquitetando um Encurtador de URL&quot;</span>
          <span>Base62 + HashIDs + CQRS + VSA</span>
        </div>
      </footer>
    </main>
  );
}
