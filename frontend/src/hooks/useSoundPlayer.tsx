import { ReactNode, useReducer } from 'react'
import { useAppURL } from './query'

interface State {
  text: string
  status: SoundPlayerStatus
  position?: number
  duration?: number
}

type Action =
  | { type: 'toggle'; text: string }
  | { type: 'stop' }
  | { type: 'loading' }
  | { type: 'playing'; pos: number; dur: number }
  | { type: 'progress'; pos: number; dur: number }
  | { type: 'status'; status: SoundPlayerStatus }

function reducer(s: State, a: Action): State {
  switch (a.type) {
    case 'toggle':
      if (a.text !== s.text) {
        // the sounds load very quickly to just skip to the playing state, to avoid a flicker
        return { text: a.text, status: 'playing' }
      }
      return { text: '', status: 'stopped' }
    case 'stop':
      return { text: '', status: 'stopped' }
    case 'loading':
      return { ...s, status: 'loading' }
    case 'playing':
      return { ...s, duration: a.dur }
    case 'progress':
      return { ...s, position: a.pos, duration: a.dur }
    case 'status':
      return { ...s, status: a.status }
  }
}

export type SoundPlayerStatus = 'stopped' | 'loading' | 'playing'

export interface SoundPlayer {
  embed: ReactNode
  toggle: (text: string) => void
  status: SoundPlayerStatus
  text: string
  position?: number
  duration?: number
}

export interface SoundPlayerProps {
  withProgress?: boolean
}

export function useSoundPlayer({ withProgress }: SoundPlayerProps = {}): SoundPlayer {
  const [state, dispatch] = useReducer(reducer, {
    text: '',
    status: 'stopped',
  })
  const { text, status, position, duration } = state
  const src = useAppURL('/app.sound', { text })

  const embed =
    text && src.type === 'ok' ? (
      <audio
        controls={false}
        autoPlay={true}
        hidden={true}
        src={src.v}
        onPlaying={({ currentTarget: el }) =>
          dispatch({ type: 'playing', pos: el.currentTime, dur: el.duration })
        }
        onTimeUpdate={
          withProgress
            ? ({ currentTarget: el }) => {
                const pos = el.currentTime
                const dur = el.duration
                if (dur > 0) {
                  dispatch({ type: 'progress', pos, dur })
                }
              }
            : undefined
        }
        onEnded={() => dispatch({ type: 'stop' })}
        onError={() => dispatch({ type: 'stop' })}
      />
    ) : null

  const toggle = (text: string) => {
    dispatch({ type: 'toggle', text })
  }

  return {
    embed,
    toggle,
    text,
    status,
    position,
    duration,
  }
}
