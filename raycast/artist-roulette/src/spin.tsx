import { Action, ActionPanel, Icon, Image, Keyboard, List, closeMainWindow, showHUD } from "@raycast/api";
import { showFailureToast, useCachedPromise, usePromise, withAccessToken } from "@raycast/utils";
import { useRef, useState } from "react";
import { ArtistContext, activeYears, artistContext } from "./lib/context";
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

const TAG_LIMIT = 6;
const SIMILAR_LIMIT = 5;

function detailMarkdown(suggestion: Suggestion, context: ArtistContext | undefined): string {
  const image = suggestion.image ? `![${suggestion.name}](${suggestion.image}?raycast-height=140)` : "";
  const similar = context?.profile?.similar.slice(0, SIMILAR_LIMIT) ?? [];

  return [
    `## ${suggestion.name}`,
    image,
    context?.profile?.bio ?? "",
    similar.length > 0 ? `**Similar:** ${similar.join(", ")}` : "",
  ]
    .filter((part) => part.length > 0)
    .join("\n\n");
}

function ArtistDetail(props: { suggestion: Suggestion; context?: ArtistContext; isLoading: boolean }) {
  const { suggestion, context, isLoading } = props;
  const tags = context?.profile?.tags.length ? context.profile.tags : suggestion.genres;
  const origin = context?.origin;
  const years = activeYears(origin);

  return (
    <List.Item.Detail
      isLoading={isLoading}
      markdown={detailMarkdown(suggestion, context)}
      metadata={
        <List.Item.Detail.Metadata>
          {origin?.type ? (
            <List.Item.Detail.Metadata.Label
              title="Type"
              text={origin.disambiguation ? `${origin.type} (${origin.disambiguation})` : origin.type}
            />
          ) : null}
          {tags.length > 0 ? (
            <List.Item.Detail.Metadata.TagList title="Tags">
              {tags.slice(0, TAG_LIMIT).map((tag) => (
                <List.Item.Detail.Metadata.TagList.Item key={tag} text={tag} />
              ))}
            </List.Item.Detail.Metadata.TagList>
          ) : null}
          <List.Item.Detail.Metadata.Label title="Listeners" text={formatListeners(suggestion.listeners)} />
          {years ? <List.Item.Detail.Metadata.Label title="Active" text={years} /> : null}
          {origin?.from ? <List.Item.Detail.Metadata.Label title="From" text={origin.from} /> : null}
          {context?.profile?.url ? (
            <>
              <List.Item.Detail.Metadata.Separator />
              <List.Item.Detail.Metadata.Link title="Last.fm" target={context.profile.url} text="Artist page" />
            </>
          ) : null}
        </List.Item.Detail.Metadata>
      }
    />
  );
}

function Command() {
  const refreshLibrary = useRef(false);
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const { isLoading, data, revalidate } = usePromise(async () => {
    const shouldRefresh = refreshLibrary.current;
    refreshLibrary.current = false;
    const spin = await createSpin({ refreshLibrary: shouldRefresh });
    await recordSpin(spin);
    return spin;
  });

  const selected = data?.shown.find((suggestion) => suggestion.id === selectedId) ?? data?.shown[0];
  const { data: context, isLoading: contextLoading } = useCachedPromise(artistContext, [selected?.name ?? ""], {
    execute: Boolean(selected),
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
    <List
      isLoading={isLoading}
      isShowingDetail
      searchBarPlaceholder="Filter suggestions"
      onSelectionChange={setSelectedId}
    >
      {SECTIONS.map((section) => (
        <List.Section key={section.bucket} title={section.title} subtitle={section.subtitle}>
          {(data?.shown ?? [])
            .filter((suggestion) => suggestion.bucket === section.bucket)
            .map((suggestion) => (
              <List.Item
                key={suggestion.id}
                id={suggestion.id}
                icon={
                  suggestion.image
                    ? { source: suggestion.image, mask: Image.Mask.Circle }
                    : { source: Icon.Music, tintColor: null }
                }
                title={suggestion.name}
                subtitle={suggestion.genres.slice(0, 3).join(", ")}
                detail={
                  <ArtistDetail
                    suggestion={suggestion}
                    context={suggestion.id === selected?.id ? context : undefined}
                    isLoading={suggestion.id === selected?.id && contextLoading}
                  />
                }
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
