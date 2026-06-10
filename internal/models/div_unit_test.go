package models

import "testing"

func TestDiv_IsValidReplayURL_ValidYouTubeWatch(t *testing.T) {
	if !IsValidReplayURL("https://www.youtube.com/watch?v=dQw4w9WgXcQ") {
		t.Error("expected true for valid YouTube watch URL")
	}
}

func TestDiv_IsValidReplayURL_ValidYouTubeShort(t *testing.T) {
	if !IsValidReplayURL("https://youtu.be/dQw4w9WgXcQ") {
		t.Error("expected true for valid YouTube short URL")
	}
}

func TestDiv_IsValidReplayURL_ValidYouTubeEmbed(t *testing.T) {
	if !IsValidReplayURL("https://www.youtube.com/embed/dQw4w9WgXcQ") {
		t.Error("expected true for valid YouTube embed URL")
	}
}

func TestDiv_IsValidReplayURL_YouTubeWithoutScheme(t *testing.T) {
	if !IsValidReplayURL("youtube.com/watch?v=dQw4w9WgXcQ") {
		t.Error("expected true for YouTube URL without scheme")
	}
}

func TestDiv_IsValidReplayURL_ValidTwitchVideos(t *testing.T) {
	if !IsValidReplayURL("https://www.twitch.tv/videos/123456789") {
		t.Error("expected true for valid Twitch videos URL")
	}
}

func TestDiv_IsValidReplayURL_ValidTwitchClips(t *testing.T) {
	if !IsValidReplayURL("https://clips.twitch.tv/CoolClipName_123") {
		t.Error("expected true for valid Twitch clips URL")
	}
}

func TestDiv_IsValidReplayURL_ValidTwitchChannel(t *testing.T) {
	if !IsValidReplayURL("https://www.twitch.tv/monchannel") {
		t.Error("expected true for valid Twitch channel URL")
	}
}

func TestDiv_IsValidReplayURL_InvalidSite(t *testing.T) {
	if IsValidReplayURL("https://vimeo.com/123456789") {
		t.Error("expected false for invalid site URL")
	}
}

func TestDiv_IsValidReplayURL_InvalidRandomText(t *testing.T) {
	if IsValidReplayURL("pas une url") {
		t.Error("expected false for random text")
	}
}

func TestDiv_IsValidReplayURL_EmptyString(t *testing.T) {
	if !IsValidReplayURL("") {
		t.Error("expected true for empty string")
	}
}

func TestDiv_IsValidReplayURL_YouTubeWithoutVideoID(t *testing.T) {
	if IsValidReplayURL("https://www.youtube.com/watch") {
		t.Error("expected false for YouTube URL without video ID")
	}
}

func TestDiv_IsValidReplayURL_YouTubeWithExtraParams(t *testing.T) {
	if IsValidReplayURL("https://www.youtube.com/watch?v=test123&t=30") {
		t.Error("expected false for YouTube URL with extra params")
	}
}

func TestDiv_IsValidReplayURL_YouTubeWithPlaylist(t *testing.T) {
	if IsValidReplayURL("https://www.youtube.com/watch?v=test123&list=PLabc") {
		t.Error("expected false for YouTube URL with playlist")
	}
}

func TestDiv_IsValidReplayURL_TwitchWithHTTP(t *testing.T) {
	if !IsValidReplayURL("http://www.twitch.tv/videos/987654321") {
		t.Error("expected true for Twitch URL with http")
	}
}

func TestDiv_IsValidReplayURL_YouTubeWithExtraPath(t *testing.T) {
	if IsValidReplayURL("https://www.youtube.com/watch?v=test123/extra") {
		t.Error("expected false for YouTube URL with extra path")
	}
}

func TestDiv_IsValidReplayURL_TwitchWithParams(t *testing.T) {
	if IsValidReplayURL("https://www.twitch.tv/videos/123456?t=1h") {
		t.Error("expected false for Twitch URL with params")
	}
}
