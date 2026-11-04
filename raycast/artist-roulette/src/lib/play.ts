import { runAppleScript } from "@raycast/utils";

const SCRIPT = `
ObjC.import("AppKit");
ObjC.import("CoreGraphics");

function spotifyProcess() {
  return $.NSRunningApplication.runningApplicationsWithBundleIdentifier("com.spotify.client").firstObject;
}

function hasVisibleWindow() {
  const options = $.kCGWindowListOptionOnScreenOnly | $.kCGWindowListExcludeDesktopElements;
  const windows = ObjC.deepUnwrap(ObjC.castRefToObject($.CGWindowListCopyWindowInfo(options, 0))) || [];
  return windows.some((w) => w.kCGWindowOwnerName === "Spotify" && w.kCGWindowLayer === 0);
}

function run(argv) {
  const uri = argv[0];
  const spotify = Application("Spotify");
  const wasRunning = spotify.running();
  const wasVisible = wasRunning && hasVisibleWindow();
  const previous = $.NSWorkspace.sharedWorkspace.frontmostApplication;

  if (!wasRunning) {
    spotify.launch();
    delay(2);
  }
  spotify.playTrack(uri);

  for (let i = 0; i < 20 && !spotifyProcess().active; i++) {
    delay(0.1);
  }

  if (!wasVisible) {
    spotifyProcess().hide;
  } else if (previous && !previous.isEqual(spotifyProcess())) {
    previous.activateWithOptions(0);
  }
}
`;

export async function startArtistRadio(artistId: string): Promise<void> {
  if (!/^[A-Za-z0-9]+$/.test(artistId)) {
    throw new Error(`Unexpected Spotify artist id: ${artistId}`);
  }

  await runAppleScript(SCRIPT, [`spotify:station:artist:${artistId}`], { language: "JavaScript" });
}
