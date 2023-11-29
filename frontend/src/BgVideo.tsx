import { useEffect, useRef } from 'react'
import './BgVideo.css'

export default function BgVideo({ src }: { src: string }) {
  const ref = useRef<HTMLVideoElement | null>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) {
      return
    }
    if (document.hasFocus()) {
      el.play()
    }
    window.addEventListener('focus', () => el.play())
    window.addEventListener('blur', () => el.pause())
    el.addEventListener('ended', () => {
      el.currentTime = 0
    })
    el.addEventListener('timeupdate', () => {
      // https://wails.io/docs/next/guides/linux/#video-tag-doesnt-fire-ended-event
      if (el.duration && el.duration - el.currentTime < 1) {
        el.dispatchEvent(new Event('ended'))
      }
    })
  }, [src])

  return (
    <video
      className="bg-video"
      ref={ref}
      loop={false /* video appears to crash the browser on Linux when looped */}
      muted
      preload="auto"
      autoPlay={false}
      src={src}
    />
  )
}
