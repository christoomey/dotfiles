import { Bucket, Spin, Suggestion, recentlyShownIds } from "./history";
import { Candidate, artistListeners, chartTopArtists, similarArtists } from "./lastfm";
import { likedArtists } from "./library";
import { preferences } from "./preferences";
import { ArtistRef, hydrateArtists, searchArtist } from "./spotify";
import { chunk, normalize, sample, shuffle } from "./text";

const KNOWN_COUNT = 3;
const ADJACENT_COUNT = 2;
const WILDCARD_COUNT = 1;
const SIMILARITY_SEEDS = 4;
const CHART_PAGES = 40;
const RESOLVE_LIMIT = 30;
const RESOLVE_BATCH = 5;

type ResolveOptions = {
  bucket: Bucket;
  count: number;
  minListeners: number;
  taken: Set<string>;
  knownNames: Set<string>;
};

async function withListeners(candidate: Candidate): Promise<Candidate> {
  if (candidate.listeners === undefined) {
    const listeners = await artistListeners(candidate.name).catch(() => undefined);
    return { name: candidate.name, listeners };
  } else {
    return candidate;
  }
}

async function resolveCandidates(candidates: Array<Candidate>, options: ResolveOptions): Promise<Array<Suggestion>> {
  const found: Array<Suggestion> = [];

  for (const batch of chunk(candidates.slice(0, RESOLVE_LIMIT), RESOLVE_BATCH)) {
    if (found.length >= options.count) {
      break;
    }

    const measured = await Promise.all(batch.map(withListeners));
    const popular = measured.filter((candidate) => (candidate.listeners ?? 0) >= options.minListeners);

    const resolved = await Promise.all(
      popular.map(async (candidate) => {
        const artist = await searchArtist(candidate.name).catch(() => undefined);
        if (artist) {
          return { ...artist, listeners: candidate.listeners, bucket: options.bucket };
        } else {
          return undefined;
        }
      }),
    );

    for (const suggestion of resolved) {
      if (
        suggestion &&
        found.length < options.count &&
        !options.taken.has(suggestion.id) &&
        !options.knownNames.has(normalize(suggestion.name))
      ) {
        options.taken.add(suggestion.id);
        found.push(suggestion);
      }
    }
  }

  return found;
}

function pickKnown(pool: Array<ArtistRef>, suppressed: Set<string>): Array<ArtistRef> {
  const fresh = pool.filter((artist) => !suppressed.has(artist.id));
  if (fresh.length >= KNOWN_COUNT) {
    return sample(fresh, KNOWN_COUNT);
  } else {
    return sample(pool, KNOWN_COUNT);
  }
}

async function knownSuggestions(refs: Array<ArtistRef>): Promise<Array<Suggestion>> {
  const artists = await hydrateArtists(refs.map((artist) => artist.id));
  return Promise.all(
    artists.map(async (artist) => ({
      ...artist,
      listeners: await artistListeners(artist.name).catch(() => undefined),
      bucket: "known" as const,
    })),
  );
}

async function adjacentCandidates(pool: Array<ArtistRef>, knownNames: Set<string>): Promise<Array<Candidate>> {
  const seeds = sample(pool, SIMILARITY_SEEDS);
  const lists = await Promise.all(seeds.map((seed) => similarArtists(seed.name).catch(() => [])));
  const unique = new Map<string, Candidate>();

  for (const candidate of lists.flat()) {
    const key = normalize(candidate.name);
    if (!knownNames.has(key) && !unique.has(key)) {
      unique.set(key, candidate);
    }
  }

  return shuffle([...unique.values()]);
}

async function wildcardCandidates(knownNames: Set<string>): Promise<Array<Candidate>> {
  const page = 1 + Math.floor(Math.random() * CHART_PAGES);
  const candidates = await chartTopArtists(page);
  return shuffle(candidates.filter((candidate) => !knownNames.has(normalize(candidate.name))));
}

export async function createSpin(options: { refreshLibrary?: boolean } = {}): Promise<Spin> {
  const { adjacentMinListeners, wildcardMinListeners } = preferences();
  const pool = await likedArtists({ refresh: options.refreshLibrary });

  if (pool.length === 0) {
    throw new Error("No liked artists found. Listen to or follow some artists on Spotify first.");
  }

  const suppressed = await recentlyShownIds();
  const knownNames = new Set(pool.map((artist) => normalize(artist.name)));
  const taken = new Set(pool.map((artist) => artist.id));

  const [known, adjacentPool, wildcardPool] = await Promise.all([
    knownSuggestions(pickKnown(pool, suppressed)),
    adjacentCandidates(pool, knownNames),
    wildcardCandidates(knownNames),
  ]);

  const adjacent = await resolveCandidates(adjacentPool, {
    bucket: "adjacent",
    count: ADJACENT_COUNT,
    minListeners: adjacentMinListeners,
    taken,
    knownNames,
  });

  const wildcard = await resolveCandidates(wildcardPool, {
    bucket: "wildcard",
    count: WILDCARD_COUNT,
    minListeners: wildcardMinListeners,
    taken,
    knownNames,
  });

  return {
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    at: Date.now(),
    shown: [...known, ...adjacent, ...wildcard],
  };
}
