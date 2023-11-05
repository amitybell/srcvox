import { ReactNode, useEffect, useReducer, useRef } from 'react'
import { FetchDataResult, fetchSound } from '../api'

interface State {
  text: string
  src: string
}

type Action = { type: 'toggle'; text: string } | { type: 'play'; src: string } | { type: 'stop' }

function reducer(s: State, a: Action): State {
  switch (a.type) {
    case 'toggle':
      return { ...s, text: s.text === a.text ? '' : a.text }
    case 'stop':
      return { ...s, text: '', src: '' }
    case 'play':
      return { ...s, src: a.src }
  }
}

export type SoundStatus = 'stopped' | 'loading' | 'playing'

export interface SoundPlayer {
  embed: ReactNode
  toggle: (text: string) => void
  status: SoundStatus
  text: string
}

export function useSoundPlayer(): SoundPlayer {
  const loading = useRef<FetchDataResult | null>(null)
  const [state, dispatch] = useReducer(reducer, { text: '', src: '' })
  const { text, src } = state

  useEffect(() => {
    loading.current?.cancel()
    loading.current = null

    if (!text) {
      dispatch({ type: 'stop' })
      return
    }

    const fdr = fetchSound(text)
    loading.current = fdr

    fdr.promise
      .then((src) => {
        if (loading.current !== fdr) {
          return
        }
        dispatch({ type: 'play', src })
      })
      .catch((e) => {})
      .finally(() => {
        if (loading.current !== fdr) {
          return
        }
        loading.current = null
      })

    return fdr.cancel
  }, [text])

  const embed = src ? (
    <audio
      controls={false}
      autoPlay={true}
      hidden={true}
      src={src}
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
    status: loading.current ? 'loading' : text ? 'playing' : 'stopped',
  }
}
