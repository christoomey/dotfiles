import { getAccessToken } from "@raycast/utils";
import { normalize } from "./text";

export type ArtistRef = {
  id: string;
  name: string;
};

export type Artist = ArtistRef & {
  genres: Array<string>;
  image?: string;
};

type ArtistObject = {
  id: string;
  name: string;
  genres?: Array<string>;
  images?: Array<{ url: string }>;
};

type Paged<T> = { items: Array<T> };

const BASE_URL = "https://api.spotify.com/v1";
const SAVED_TRACK_PAGES = 6;
const PAGE_SIZE = 50;

async function get<T>(path: string): Promise<T> {
  const { token } = getAccessToken();
  const response = await fetch(`${BASE_URL}${path}`, {
    headers: { Authorization: `Bearer ${token}` },
  });

  if (response.ok) {
    return (await response.json()) as T;
  } else {
    throw new Error(`Spotify API ${response.status}: ${await response.text()}`);
  }
}

function toArtist(artist: ArtistObject): Artist {
  return {
    id: artist.id,
    name: artist.name,
    genres: artist.genres ?? [],
    image: artist.images?.[0]?.url,
  };
}

function toRef(artist: ArtistObject): ArtistRef {
  return { id: artist.id, name: artist.name };
}

export async function topArtists(timeRange: "short_term" | "medium_term" | "long_term"): Promise<Array<ArtistRef>> {
  const data = await get<Paged<ArtistObject>>(`/me/top/artists?time_range=${timeRange}&limit=${PAGE_SIZE}`);
  return data.items.map(toRef);
}

export async function followedArtists(): Promise<Array<ArtistRef>> {
  const collected: Array<ArtistRef> = [];
  let after: string | undefined = undefined;

  do {
    const cursor: string = after ? `&after=${after}` : "";
    const data: { artists: Paged<ArtistObject> & { cursors?: { after?: string } } } = await get(
      `/me/following?type=artist&limit=${PAGE_SIZE}${cursor}`,
    );
    collected.push(...data.artists.items.map(toRef));
    after = data.artists.cursors?.after;
  } while (after);

  return collected;
}

export async function savedTrackArtists(): Promise<Array<ArtistRef>> {
  const offsets = Array.from({ length: SAVED_TRACK_PAGES }, (_, page) => page * PAGE_SIZE);
  const pages = await Promise.all(
    offsets.map((offset) =>
      get<Paged<{ track?: { artists?: Array<ArtistObject> } }>>(`/me/tracks?limit=${PAGE_SIZE}&offset=${offset}`),
    ),
  );

  return pages.flatMap((page) => page.items.flatMap((item) => (item.track?.artists ?? []).map(toRef)));
}

export async function hydrateArtists(ids: Array<string>): Promise<Array<Artist>> {
  const artists = await Promise.all(ids.map((id) => get<ArtistObject>(`/artists/${id}`).catch(() => undefined)));
  return artists.filter((artist): artist is ArtistObject => artist !== undefined).map(toArtist);
}

export async function searchArtist(name: string): Promise<Artist | undefined> {
  const data = await get<{ artists: Paged<ArtistObject> }>(`/search?q=${encodeURIComponent(name)}&type=artist&limit=5`);
  const match = data.artists.items.find((artist) => normalize(artist.name) === normalize(name));
  return match ? toArtist(match) : undefined;
}
