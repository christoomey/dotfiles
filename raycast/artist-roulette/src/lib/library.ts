import { LocalStorage } from "@raycast/api";
import { ArtistRef, followedArtists, savedTrackArtists, topArtists } from "./spotify";

const STORAGE_KEY = "liked-artists";
const TTL_MS = 24 * 60 * 60 * 1000;

type CachedLibrary = {
  at: number;
  artists: Array<ArtistRef>;
};

function dedupe(artists: Array<ArtistRef>): Array<ArtistRef> {
  const byId = new Map<string, ArtistRef>();
  for (const artist of artists) {
    if (!byId.has(artist.id)) {
      byId.set(artist.id, artist);
    }
  }
  return [...byId.values()];
}

async function build(): Promise<Array<ArtistRef>> {
  const groups = await Promise.all([
    topArtists("short_term"),
    topArtists("medium_term"),
    topArtists("long_term"),
    followedArtists(),
    savedTrackArtists(),
  ]);
  return dedupe(groups.flat());
}

export async function likedArtists(options: { refresh?: boolean } = {}): Promise<Array<ArtistRef>> {
  if (!options.refresh) {
    const raw = await LocalStorage.getItem<string>(STORAGE_KEY);
    if (raw) {
      const cached = JSON.parse(raw) as CachedLibrary;
      if (Date.now() - cached.at < TTL_MS && cached.artists.length > 0) {
        return cached.artists;
      }
    }
  }

  const artists = await build();
  await LocalStorage.setItem(STORAGE_KEY, JSON.stringify({ at: Date.now(), artists } satisfies CachedLibrary));
  return artists;
}

export async function libraryCachedAt(): Promise<number | undefined> {
  const raw = await LocalStorage.getItem<string>(STORAGE_KEY);
  if (raw) {
    return (JSON.parse(raw) as CachedLibrary).at;
  } else {
    return undefined;
  }
}
