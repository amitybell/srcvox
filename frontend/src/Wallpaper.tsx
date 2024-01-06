import { ReactNode, useEffect, useState } from 'react'
import cls from './Wallpaper.module.css'
import { useInterval } from '@mantine/hooks'

export interface WallpaperProps {
  children?: ReactNode
  urls: string[]
  interval: number
}

export default function Wallpaper({ children, urls, interval }: WallpaperProps) {
  const [rand, setRand] = useState(Math.random())
  const intv = useInterval(() => setRand(Math.random()), interval)
  useEffect(() => {
    intv.start()
    return intv.stop
  }, [intv])

  const idx = Math.floor(rand * urls.length)
  const url = urls[idx]

  return (
    <div
      className={cls.root}
      style={url ? { ['--wp-bg' as string]: `url("${encodeURI(url)}")` } : undefined}
    >
      {children}
    </div>
  )
}
