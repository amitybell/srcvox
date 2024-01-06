import cls from './Badges.module.css'
import { useRef, useState } from 'react'
import { GameInfo, Presence, ServerInfo } from './appstate'
import { useMapImage } from './hooks/query'
import Flag from 'react-world-flags'
import Avatar from './Avatar'
import { domToBlob } from 'modern-screenshot'
import { notifications } from '@mantine/notifications'
import { Loader } from '@mantine/core'
import {
  PiLockKeyOpenFill as PrivateServerIcon,
  PiCameraDuotone as ScreenshotIcon,
} from 'react-icons/pi'

export async function screenshot<E extends HTMLElement>({
  elem,
  onDone,
}: {
  elem: E | null
  onDone?: (r: Blob | Error) => void
}) {
  function done(r: Blob | Error) {
    if (r instanceof Blob) {
      notifications.show({ color: 'green', message: 'Badge copied to clipbaord' })
    } else {
      notifications.show({ color: 'red', message: r.message })
    }
    onDone?.(r)
  }

  if (!elem) {
    done(new Error('No element'))
    return
  }

  const opts = {
    width: elem.clientWidth,
    height: elem.clientHeight,
  }

  navigator.clipboard
    .write([
      new ClipboardItem({
        'image/png': domToBlob(elem, opts).then((blob) => {
          done(blob)
          return blob
        }),
      }),
    ])
    .catch((e) => {
      const err = new Error(`Failed to copy badge to clipbaord: ${e}`)
      done(err)
      notifications.show({ color: 'red', message: err.message })
    })
}

export interface BadgeProps {
  srv: ServerInfo
  game: GameInfo
  pr: Presence
}

export function Badge({ srv, game, pr }: BadgeProps) {
  const badgeRef = useRef<HTMLDivElement | null>(null)
  const mapImg = useMapImage(game.id, srv.map)

  if (mapImg.type !== 'ok') {
    return null
  }

  const players = srv.players
  const humans = pr.server === srv.addr ? pr.humans : []

  return (
    <div ref={badgeRef} className={cls.badge}>
      <div
        className={`${cls.content} ${cls.bgImg}`}
        style={{ ['--bg-img-url' as string]: `url("${mapImg.v}")` }}
      >
        <div className={cls.head}>
          <span className={cls.tag}>
            {players} {players === 1 ? 'player' : 'players'}
          </span>

          {srv.name.split(/(?:\s*[|]+\s*)|(?:\s+[-]\s+)/).map((s, i) => (
            <span key={i} className={cls.tag}>
              {s}
            </span>
          ))}

          <span className={`${cls.tag} ${cls.compact}`}>
            {srv.restricted ? (
              <PrivateServerIcon className={cls.icon} />
            ) : (
              <Flag className={cls.icon} code={srv.country} />
            )}
          </span>
        </div>
        {humans.length ? (
          <div className={cls.body}>
            {humans.map((p, i) => (
              <Avatar {...p} key={p.id || i} className={cls.avatar} embedded />
            ))}
          </div>
        ) : null}
      </div>
    </div>
  )
}

export interface BadgesProps {
  srvs: ServerInfo[]
  game: GameInfo
  pr: Presence
  onSnap?: (src: Blob) => void
}

export default function Badges({ srvs, game, pr, onSnap }: BadgesProps) {
  const badgesRef = useRef<HTMLDivElement | null>(null)
  const [loading, setLoading] = useState(false)
  const hasPlayers = srvs.some((p) => p.players > 0)

  function onClick() {
    setLoading(true)
    screenshot({
      elem: badgesRef.current,
      onDone: (r) => {
        setLoading(false)
        if (onSnap && r instanceof Blob) {
          onSnap(r)
        }
      },
    })
  }

  return (
    <div className={cls.badgesCtr}>
      {hasPlayers ? (
        <>
          <div>
            <button className={cls.snapBtn} disabled={loading} onClick={onClick}>
              {loading ? <Loader color="#fff" /> : <ScreenshotIcon />}
            </button>
          </div>
          <div ref={badgesRef} className={cls.badges}>
            {srvs.map((srv) => (
              <Badge key={srv.addr} srv={srv} game={game} pr={pr} />
            ))}
          </div>
        </>
      ) : (
        <div className={cls.defaultMsg}>
          <h2>All servers are empty!</h2>
          <p>This view will refresh when players come online.</p>
        </div>
      )}
    </div>
  )
}
