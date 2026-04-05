const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface ShortenResponse {
  short_url: string;
}

export interface StatsResponse {
  urls_shortened: number;
  urls_redirected: number;
  cache_hits: number;
  cache_misses: number;
  cache_hit_rate: number;
  requests_shorten: number;
  requests_redirect: number;
}

export async function shortenUrl(url: string): Promise<ShortenResponse> {
  const res = await fetch(`${API_URL}/shorten`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ url }),
  });

  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error || "Erro ao encurtar URL");
  }

  return res.json();
}

export async function getStats(): Promise<StatsResponse> {
  const res = await fetch(`${API_URL}/stats`, { cache: "no-store" });
  if (!res.ok) throw new Error("Erro ao buscar stats");
  return res.json();
}
