import { Profile, artistProfile } from "./lastfm";
import { Origin, artistOrigin } from "./musicbrainz";

export type ArtistContext = {
  profile?: Profile;
  origin?: Origin;
};

export async function artistContext(name: string): Promise<ArtistContext> {
  const [profile, origin] = await Promise.all([
    artistProfile(name).catch(() => undefined),
    artistOrigin(name).catch(() => undefined),
  ]);
  return { profile, origin };
}

export function activeYears(origin: Origin | undefined): string | undefined {
  if (!origin?.began) {
    return undefined;
  } else if (origin.ended) {
    return `${origin.began}-${origin.ended}`;
  } else {
    return `since ${origin.began}`;
  }
}
