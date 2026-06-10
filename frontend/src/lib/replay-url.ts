export type ReplayType = 'youtube' | 'twitch' | null

export function detectReplayType(url: string): ReplayType {
  if (/^(https?:\/\/)?(www\.)?(youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)/.test(url)) return 'youtube'
  if (/^(https?:\/\/)?(www\.|clips\.)?twitch\.tv\//.test(url)) return 'twitch'
  return null
}

export function getYouTubeEmbedUrl(url: string): string | null {
  const match = url.match(/(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([\w-]+)/)
  return match ? `https://www.youtube.com/embed/${match[1]}` : null
}
