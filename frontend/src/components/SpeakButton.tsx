import { useSpeech } from '../hooks/useSpeech'

interface Props {
  text: string
  lang?: string
  size?: 'sm' | 'md'
}

export function SpeakButton({ text, lang = 'it-IT', size = 'sm' }: Props) {
  const { speak, stop, speaking } = useSpeech()

  function handleClick(e: React.MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    speaking ? stop() : speak(text, lang)
  }

  return (
    <button
      className={`speak-btn speak-btn--${size} ${speaking ? 'speak-btn--active' : ''}`}
      onClick={handleClick}
      title={speaking ? 'Stop' : 'Listen'}
      aria-label={speaking ? 'Stop audio' : 'Play in Italian'}
    >
      {speaking ? '■' : '▶'}
    </button>
  )
}
