import { Action, ActionPanel, Icon, Image, Keyboard, List, closeMainWindow, showHUD } from "@raycast/api";
import { showFailureToast, usePromise, withAccessToken } from "@raycast/utils";
import { useRef } from "react";
import { Bucket, Spin, Suggestion, recordPick, recordSpin } from "./lib/history";
import { formatListeners } from "./lib/lastfm";
import { spotifyOAuth } from "./lib/oauth";
import { startArtistRadio } from "./lib/play";
import { createSpin } from "./lib/roulette";

const SECTIONS: Array<{ bucket: Bucket; title: string; subtitle: string }> = [
  { bucket: "known", title: "From Your Rotation", subtitle: "Artists you already listen to" },
  { bucket: "adjacent", title: "Adjacent", subtitle: "Similar to your taste, not in your library" },
  { bucket: "wildcard", title: "Wild Card", subtitle: "Popular, and nothing to do with you" },
];

async function play(spin: Spin, suggestion: Suggestion) {
  try {
    await startArtistRadio(suggestion.id);
    await recordPick(spin.id, suggestion);
    await closeMainWindow();
    await showHUD(`Playing ${suggestion.name} radio`);
  } catch (error) {
    await showFailureToast(error, { title: "Could not start radio" });
  }
}

function Command() {
  const refreshLibrary = useRef(false);

  const { isLoading, data, revalidate } = usePromise(async () => {
    const shouldRefresh = refreshLibrary.current;
    refreshLibrary.current = false;
    const spin = await createSpin({ refreshLibrary: shouldRefresh });
    await recordSpin(spin);
    return spin;
  });

  function actions(suggestion: Suggestion) {
    return (
      <ActionPanel>
        <Action title="Start Radio" icon={Icon.Play} onAction={() => data && play(data, suggestion)} />
        <Action.OpenInBrowser
          title="Open Artist in Spotify"
          icon={Icon.Music}
          url={`spotify:artist:${suggestion.id}`}
        />
        <Action
          title="Spin Again"
          icon={Icon.ArrowClockwise}
          shortcut={Keyboard.Shortcut.Common.Refresh}
          onAction={revalidate}
        />
        <Action
          title="Rebuild Library Cache"
          icon={Icon.Download}
          shortcut={{ modifiers: ["cmd", "shift"], key: "r" }}
          onAction={() => {
            refreshLibrary.current = true;
            revalidate();
          }}
        />
      </ActionPanel>
    );
  }

  return (
    <List isLoading={isLoading} searchBarPlaceholder="Filter suggestions">
      {SECTIONS.map((section) => (
        <List.Section key={section.bucket} title={section.title} subtitle={section.subtitle}>
          {(data?.shown ?? [])
            .filter((suggestion) => suggestion.bucket === section.bucket)
            .map((suggestion) => (
              <List.Item
                key={suggestion.id}
                icon={
                  suggestion.image
                    ? { source: suggestion.image, mask: Image.Mask.Circle }
                    : { source: Icon.Music, tintColor: null }
                }
                title={suggestion.name}
                subtitle={suggestion.genres.slice(0, 3).join(", ")}
                accessories={[{ icon: Icon.TwoPeople, text: formatListeners(suggestion.listeners) }]}
                actions={actions(suggestion)}
              />
            ))}
        </List.Section>
      ))}
      <List.EmptyView
        icon={Icon.Music}
        title={isLoading ? "Spinning…" : "Nothing to suggest"}
        description={isLoading ? undefined : "Try rebuilding the library cache from the action menu."}
      />
    </List>
  );
}

export default withAccessToken(spotifyOAuth)(Command);
