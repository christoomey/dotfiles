import { OAuth } from "@raycast/api";
import { OAuthService } from "@raycast/utils";
import { preferences } from "./preferences";

const client = new OAuth.PKCEClient({
  redirectMethod: OAuth.RedirectMethod.Web,
  providerName: "Spotify",
  providerIcon: "icon.png",
  description: "Connect your Spotify account",
  providerId: "spotify-artist-roulette",
});

export const spotifyOAuth = new OAuthService({
  client,
  clientId: preferences().spotifyClientId,
  scope: "user-top-read user-follow-read user-library-read",
  authorizeUrl: "https://accounts.spotify.com/authorize/",
  tokenUrl: "https://accounts.spotify.com/api/token",
  refreshTokenUrl: "https://accounts.spotify.com/api/token",
  bodyEncoding: "url-encoded",
});
