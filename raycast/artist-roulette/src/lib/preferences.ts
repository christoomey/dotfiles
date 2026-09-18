import { getPreferenceValues } from "@raycast/api";

type RawPreferences = {
  spotifyClientId: string;
  lastfmApiKey: string;
  adjacentMinListeners?: string;
  wildcardMinListeners?: string;
};

export type Preferences = {
  spotifyClientId: string;
  lastfmApiKey: string;
  adjacentMinListeners: number;
  wildcardMinListeners: number;
};

function parseListeners(value: string | undefined, fallback: number): number {
  const parsed = Number.parseInt((value ?? "").replace(/[_,]/g, ""), 10);
  if (Number.isNaN(parsed) || parsed < 0) {
    return fallback;
  } else {
    return parsed;
  }
}

export function preferences(): Preferences {
  const raw = getPreferenceValues<RawPreferences>();
  return {
    spotifyClientId: raw.spotifyClientId.trim(),
    lastfmApiKey: raw.lastfmApiKey.trim(),
    adjacentMinListeners: parseListeners(raw.adjacentMinListeners, 50_000),
    wildcardMinListeners: parseListeners(raw.wildcardMinListeners, 500_000),
  };
}
