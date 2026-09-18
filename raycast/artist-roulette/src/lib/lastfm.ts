import { preferences } from "./preferences";

const BASE_URL = "https://ws.audioscrobbler.com/2.0/";

export type Candidate = {
  name: string;
  listeners?: number;
};

type LastfmError = { error?: number; message?: string };

async function call<T>(params: Record<string, string>): Promise<T> {
  const query = new URLSearchParams({ ...params, api_key: preferences().lastfmApiKey, format: "json" });
  const response = await fetch(`${BASE_URL}?${query.toString()}`);

  if (!response.ok) {
    throw new Error(`Last.fm ${response.status}: ${await response.text()}`);
  }

  const data = (await response.json()) as T & LastfmError;
  if (data.error) {
    throw new Error(`Last.fm error ${data.error}: ${data.message ?? "unknown"}`);
  } else {
    return data;
  }
}

function toCount(value: string | undefined): number | undefined {
  const parsed = Number.parseInt(value ?? "", 10);
  if (Number.isNaN(parsed)) {
    return undefined;
  } else {
    return parsed;
  }
}

export async function similarArtists(name: string, limit = 100): Promise<Array<Candidate>> {
  const data = await call<{ similarartists?: { artist?: Array<{ name: string }> } }>({
    method: "artist.getsimilar",
    artist: name,
    limit: String(limit),
    autocorrect: "1",
  });
  return (data.similarartists?.artist ?? []).map((artist) => ({ name: artist.name }));
}

export async function chartTopArtists(page: number, limit = 50): Promise<Array<Candidate>> {
  const data = await call<{ artists?: { artist?: Array<{ name: string; listeners?: string }> } }>({
    method: "chart.gettopartists",
    page: String(page),
    limit: String(limit),
  });
  return (data.artists?.artist ?? []).map((artist) => ({ name: artist.name, listeners: toCount(artist.listeners) }));
}

export async function artistListeners(name: string): Promise<number | undefined> {
  const data = await call<{ artist?: { stats?: { listeners?: string } } }>({
    method: "artist.getinfo",
    artist: name,
    autocorrect: "1",
  });
  return toCount(data.artist?.stats?.listeners);
}

export function formatListeners(listeners: number | undefined): string {
  if (listeners === undefined) {
    return "unknown";
  } else if (listeners >= 1_000_000) {
    return `${(listeners / 1_000_000).toFixed(1)}M listeners`;
  } else if (listeners >= 1_000) {
    return `${Math.round(listeners / 1_000)}K listeners`;
  } else {
    return `${listeners} listeners`;
  }
}
