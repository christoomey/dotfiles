import { Action, ActionPanel, Alert, Color, Icon, List, closeMainWindow, confirmAlert, showHUD } from "@raycast/api";
import { showFailureToast, usePromise } from "@raycast/utils";
import { Bucket, Spin, clearSpins, loadSpins } from "./lib/history";
import { formatListeners } from "./lib/lastfm";
import { startArtistRadio } from "./lib/play";

const BUCKET_LABELS: Record<Bucket, string> = {
  known: "rotation",
  adjacent: "adjacent",
  wildcard: "wild card",
};

function detail(spin: Spin): string {
  const rows = spin.shown.map((suggestion) => {
    const marker = spin.picked?.id === suggestion.id ? "▶︎" : "·";
    const genres = suggestion.genres.slice(0, 3).join(", ");
    return `| ${marker} | ${suggestion.name} | ${BUCKET_LABELS[suggestion.bucket]} | ${formatListeners(suggestion.listeners)} | ${genres} |`;
  });

  return [
    `**${new Date(spin.at).toLocaleString()}**`,
    "",
    "| | Artist | Bucket | Listeners | Genres |",
    "| - | - | - | - | - |",
    ...rows,
  ].join("\n");
}

async function replay(spin: Spin) {
  if (spin.picked) {
    try {
      await startArtistRadio(spin.picked.id);
      await closeMainWindow();
      await showHUD(`Playing ${spin.picked.name} radio`);
    } catch (error) {
      await showFailureToast(error, { title: "Could not start radio" });
    }
  }
}

export default function Command() {
  const { isLoading, data, revalidate } = usePromise(loadSpins);

  return (
    <List isLoading={isLoading} isShowingDetail searchBarPlaceholder="Filter spins">
      {(data ?? []).map((spin) => (
        <List.Item
          key={spin.id}
          icon={spin.picked ? { source: Icon.Play, tintColor: Color.Green } : { source: Icon.Minus }}
          title={spin.picked?.name ?? "No pick"}
          subtitle={new Date(spin.at).toLocaleString()}
          detail={<List.Item.Detail markdown={detail(spin)} />}
          actions={
            <ActionPanel>
              {spin.picked ? <Action title="Play This Again" icon={Icon.Play} onAction={() => replay(spin)} /> : null}
              <Action
                title="Clear History"
                icon={Icon.Trash}
                style={Action.Style.Destructive}
                onAction={async () => {
                  const confirmed = await confirmAlert({
                    title: "Clear spin history?",
                    message: "This removes every recorded spin and pick.",
                    primaryAction: { title: "Clear", style: Alert.ActionStyle.Destructive },
                  });
                  if (confirmed) {
                    await clearSpins();
                    revalidate();
                  }
                }}
              />
            </ActionPanel>
          }
        />
      ))}
      <List.EmptyView icon={Icon.Clock} title="No spins yet" description="Run Spin Artist Roulette to get started." />
    </List>
  );
}
