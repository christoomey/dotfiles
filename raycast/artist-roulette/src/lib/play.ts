import { runAppleScript } from "@raycast/utils";

export async function startArtistRadio(artistId: string): Promise<void> {
  if (!/^[A-Za-z0-9]+$/.test(artistId)) {
    throw new Error(`Unexpected Spotify artist id: ${artistId}`);
  }

  await runAppleScript(`
    tell application "Spotify"
      if it is not running then
        launch
        delay 2
      end if
      play track "spotify:station:artist:${artistId}"
    end tell
  `);
}
