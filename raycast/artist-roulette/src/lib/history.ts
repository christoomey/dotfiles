import { LocalStorage } from "@raycast/api";
import { Artist } from "./spotify";

export type Bucket = "known" | "adjacent" | "wildcard";

export type Suggestion = Artist & { bucket: Bucket; listeners?: number };

export type Spin = {
  id: string;
  at: number;
  shown: Array<Suggestion>;
  picked?: { id: string; name: string; bucket: Bucket };
};

const STORAGE_KEY = "spins";
const MAX_SPINS = 200;
const SUPPRESS_SPINS = 20;

export async function loadSpins(): Promise<Array<Spin>> {
  const raw = await LocalStorage.getItem<string>(STORAGE_KEY);
  if (raw) {
    return JSON.parse(raw) as Array<Spin>;
  } else {
    return [];
  }
}

async function save(spins: Array<Spin>): Promise<void> {
  await LocalStorage.setItem(STORAGE_KEY, JSON.stringify(spins.slice(0, MAX_SPINS)));
}

export async function recordSpin(spin: Spin): Promise<void> {
  const spins = await loadSpins();
  await save([spin, ...spins]);
}

export async function recordPick(spinId: string, suggestion: Suggestion): Promise<void> {
  const spins = await loadSpins();
  const updated = spins.map((spin) => {
    if (spin.id === spinId) {
      return { ...spin, picked: { id: suggestion.id, name: suggestion.name, bucket: suggestion.bucket } };
    } else {
      return spin;
    }
  });
  await save(updated);
}

export async function clearSpins(): Promise<void> {
  await LocalStorage.removeItem(STORAGE_KEY);
}

export async function recentlyShownIds(): Promise<Set<string>> {
  const spins = await loadSpins();
  return new Set(spins.slice(0, SUPPRESS_SPINS).flatMap((spin) => spin.shown.map((artist) => artist.id)));
}
