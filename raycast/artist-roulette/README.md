# Artist Roulette

Suggests six artists and starts a real Spotify artist radio station from whichever one you pick.

- 3 from your rotation (top artists, followed artists, saved tracks)
- 2 adjacent (Last.fm similarity, filtered to artists not in your library)
- 1 wild card (Last.fm global chart, filtered to artists not in your library)

Picks are suppressed for the next 20 spins so you don't see the same six twice in a week, and
every spin is recorded with what you were shown and what you chose.

## Why not the Spotify recommendations API

Spotify deprecated `/recommendations` and `/artists/{id}/related-artists` on 2024-11-27 for any app
registered after that date, then in February 2026 removed `GET /artists` (Get Several Artists) and
dropped the `popularity` and `followers` fields from the artist object entirely. None of it was
reverted in the March, May, or July 2026 changelogs.

So similarity *and* the "is this artist mainstream" signal both come from Last.fm: `artist.getSimilar`
for adjacency, `artist.getInfo` and `chart.getTopArtists` for listener counts. Artists are hydrated
one at a time through `GET /artists/{id}`, which survives.

Playback does not use the Web API at all. `spotify:station:artist:<id>` handed to the desktop app
over AppleScript starts a genuine Spotify artist radio, with the drift you get from the "Go to
artist radio" menu item. That means no Premium-only player scopes and no hand-built track queue.

## Setup

### 1. Spotify app

1. Go to https://developer.spotify.com/dashboard and create an app.
2. Name/description can be anything. Choose **Web API** as the API.
3. Set the redirect URI to exactly:

   ```
   https://raycast.com/redirect?packageName=Extension
   ```

4. Copy the **Client ID**. No client secret is needed (PKCE flow).

Development mode is fine. It caps the app at 25 users, which is 24 more than you need.

### 2. Last.fm API key

1. Go to https://www.last.fm/api/account/create
2. Copy the **API key**. The shared secret is not needed; all calls are unauthenticated reads.

### 3. Install

```sh
cd ~/code/dotfiles/raycast/artist-roulette
npm install
npm run dev
```

Raycast picks it up as a local development extension. Run **Spin Artist Roulette**, paste the
Client ID and API key into the preferences prompt, and authorize Spotify.

Leave `npm run dev` running only while you're editing. The extension stays installed after you
stop it, as long as this directory stays on disk.

## Preferences

| Preference | Default | Notes |
| - | - | - |
| Spotify Client ID | required | From the Spotify dashboard |
| Last.fm API Key | required | From last.fm/api |
| Adjacent Minimum Listeners | 50000 | Last.fm listeners an adjacent pick must clear. Lower means more obscure |
| Wild Card Minimum Listeners | 500000 | Raise for bigger names, lower for deeper cuts |

## Commands

- **Spin Artist Roulette** - the six suggestions. `⌘R` respins, `⌘⇧R` rebuilds the library cache.
- **Artist Roulette History** - past spins, what was shown, what you picked, replay any pick.

The library cache (top + followed + saved-track artists) refreshes automatically every 24 hours.

## Scopes

`user-top-read`, `user-follow-read`, `user-library-read`. Read-only. Playback happens through
AppleScript against the local Spotify app, so nothing here can modify your account.
