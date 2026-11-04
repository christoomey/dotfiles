import { normalize } from "./text";

const BASE_URL = "https://musicbrainz.org/ws/2";
const USER_AGENT = "ArtistRoulette/1.0 ( https://github.com/christoomey/dotfiles )";
const MIN_INTERVAL_MS = 1100;
const MIN_SCORE = 90;

export type Origin = {
  type?: string;
  began?: string;
  ended?: string;
  from?: string;
  disambiguation?: string;
};

type ArtistResult = {
  name: string;
  score: number;
  type?: string;
  country?: string;
  disambiguation?: string;
  area?: { name?: string };
  "begin-area"?: { name?: string };
  "life-span"?: { begin?: string; end?: string };
};

let nextSlot = 0;

async function throttle(): Promise<void> {
  const now = Date.now();
  const wait = Math.max(0, nextSlot - now);
  nextSlot = Math.max(now, nextSlot) + MIN_INTERVAL_MS;
  if (wait > 0) {
    await new Promise((resolve) => setTimeout(resolve, wait));
  }
}

function year(date: string | undefined): string | undefined {
  return date?.slice(0, 4);
}

function from(artist: ArtistResult): string | undefined {
  const places = [artist["begin-area"]?.name, artist.area?.name ?? artist.country].filter((place): place is string =>
    Boolean(place),
  );
  const unique = [...new Set(places)];
  if (unique.length > 0) {
    return unique.join(", ");
  } else {
    return undefined;
  }
}

export async function artistOrigin(name: string): Promise<Origin | undefined> {
  await throttle();
  const query = new URLSearchParams({ query: `artist:"${name.replace(/"/g, "")}"`, fmt: "json", limit: "5" });
  const response = await fetch(`${BASE_URL}/artist/?${query.toString()}`, {
    headers: { "User-Agent": USER_AGENT, Accept: "application/json" },
  });

  if (!response.ok) {
    throw new Error(`MusicBrainz ${response.status}`);
  }

  const data = (await response.json()) as { artists?: Array<ArtistResult> };
  const match = (data.artists ?? []).find(
    (artist) => artist.score >= MIN_SCORE && normalize(artist.name) === normalize(name),
  );

  if (match) {
    return {
      type: match.type,
      began: year(match["life-span"]?.begin),
      ended: year(match["life-span"]?.end),
      from: from(match),
      disambiguation: match.disambiguation || undefined,
    };
  } else {
    return undefined;
  }
}
