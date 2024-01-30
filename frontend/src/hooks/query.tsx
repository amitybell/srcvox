import {
  QueryFunction,
  UndefinedInitialDataOptions,
  useQueries,
  useQuery,
  UseQueryResult,
} from '@tanstack/react-query'
import deepEqual from 'deep-equal'
import { ReactElement, useEffect, useId, useMemo, useRef, useState } from 'react'
import { app, useAppEvent } from '../api'
import {
  AppError,
  AppState,
  coerce,
  Config,
  GameInfo,
  Presence,
  Profile,
  Region,
  ServerInfo,
  SoundInfo,
} from '../appstate'
import Err from '../Error'
import Spinner from '../Spinner'

export type QResult<T> =
  | {
      type: 'loading'
      alt: ReactElement
      refetch: UseQueryResult['refetch']
    }
  | {
      type: 'error'
      error: Error
      alt: ReactElement
      refetch: UseQueryResult['refetch']
    }
  | {
      type: 'ok'
      v: T
      refetch: UseQueryResult['refetch']
    }

function qResult<T>({ isLoading, refetch, error, data }: UseQueryResult<T, Error>): QResult<T> {
  if (isLoading) {
    return {
      type: 'loading',
      alt: (
        <span>
          <Spinner />
        </span>
      ),
      refetch,
    }
  }
  if (error) {
    return { type: 'error', error, alt: <Err noLogo>{error.message}</Err>, refetch }
  }
  if (typeof data === 'undefined') {
    throw new Error(`data is undefined`)
  }
  return { type: 'ok', v: data, refetch }
}

function eq(a: unknown, b: unknown): boolean {
  return deepEqual(a, b, { strict: true })
}

export function useMemoEq<T, U>(create: (prevVal: T | null, prevDeps: U | null) => T, deps: U): T {
  const ref = useRef<{ deps: U | null; val: T | null }>({ deps: null, val: null })
  if (!ref.current.val || !eq(ref.current.deps, deps)) {
    const { deps: prevDeps, val: prevVal } = ref.current
    ref.current.val = create(prevVal, prevDeps)
    ref.current.deps = deps
  }
  return ref.current.val
}

export function useData<T, U>(
  key: string | string[],
  queryFn: QueryFunction<U>,
  select: (data: U) => T,
  extraOptions: Omit<UndefinedInitialDataOptions<U>, 'queryKey' | 'queryFn'> = {},
): QResult<T> {
  const r = useQuery({
    ...extraOptions,
    queryKey: typeof key === 'string' ? [key] : key,
    queryFn,
    select,
  })
  return useMemoEq(() => qResult(r), [key, r])
}

export function useFetchDataURI(
  key: string | string[],
  url: string | undefined | null,
  params: Record<string, string> = {},
  extraOptions: Omit<UndefinedInitialDataOptions<string>, 'queryKey' | 'queryFn'> = {},
): QResult<string> {
  return useData(
    key,
    async ({ signal }) => {
      if (!url) {
        throw new Error('Invalid URL')
      }

      const q = new URLSearchParams()
      for (const [k, v] of Object.entries(params)) {
        q.set(k, v)
      }
      const targ = `${url}${url.includes('?') ? '&' : '?'}${q}`
      const r = await fetch(targ, {
        signal,
        headers: {
          pragma: 'no-cache',
          'cache-control': 'no-cache',
        },
      })
      const b = await r.blob()
      return URL.createObjectURL(b)
    },
    (r) => r,
    extraOptions,
  )
}

export function useSynthesize(text: string): QResult<string> {
  return useFetchDataURI(`app.Synthesize(${text})`, '/app.synthesize', { text })
}

export function useMapImage(gameID: number, mapName: string): QResult<string> {
  return useFetchDataURI(`app.useMapImage(${gameID}, ${mapName})`, '/app.mapimage', {
    id: gameID.toString(),
    map: mapName,
  })
}

export function useAppAddr(): QResult<string> {
  return useData('app.AppAddr', app.AppAddr, (p) => coerce('', p))
}

export function useAppURL(path: string, params: Record<string, string> = {}): QResult<string> {
  const addr = useAppAddr()
  if (addr.type !== 'ok') {
    return addr
  }
  const q = new URLSearchParams(params)
  return { ...addr, v: `http://${addr.v}${path}?${q}` }
}

export function useBgVideo(gameID: number): QResult<string> {
  return useAppURL('/app.bgvideo', { id: gameID.toString() })
}

export function useGames(): QResult<GameInfo[]> {
  return useData('app.Games', app.Games, (p) => coerce([], p).map((q) => new GameInfo(q)))
}

export function useLaunchOptions(userID: number, gameID: number): QResult<string> {
  return useData(
    'app.LaunchOptions',
    () => app.LaunchOptions(userID, gameID),
    (p) => coerce('', p),
  )
}

export function useSounds(): QResult<SoundInfo[]> {
  return useData('app.Sounds', app.Sounds, (p) => coerce([], p).map((q) => new SoundInfo(q)))
}

export function useAppState(): QResult<AppState> {
  return useData('app.State', app.State, (p) => new AppState(p))
}

export function useConfig(): QResult<Config> {
  const cfg = useData('app.Config', app.Config, (p) => new Config(p))
  useAppEvent('sv.ConfigChange', () => {
    cfg.refetch()
  })
  return cfg
}

export function useAppError(): QResult<AppError> {
  const err = useData('app.Error', app.Error, (p) => new AppError(p))
  useAppEvent('sv.ErrorChange', () => {
    err.refetch()
  })
  return err
}

export function usePresence(): QResult<Presence> {
  const pr = useData('app.Presence', app.Presence, (p) => new Presence(p))
  useAppEvent('sv.PresenceChange', pr.refetch)
  return pr
}

export function useServerInfos(
  addrs: Record<string, Region>,
  refresh: number,
  less: (a: ServerInfo, b: ServerInfo) => boolean,
): ServerInfo[] {
  const res = useQueries({
    queries: Object.entries(addrs).map(([addr, reg]) => ({
      queryKey: ['useServerInfos', addr],
      queryFn: () => app.ServerInfo(reg, addr),
      select: (p: Partial<ServerInfo>) => new ServerInfo(p),
      refetchInterval: refresh > 0 ? refresh : undefined,
    })),
  })

  const pr = usePresence()
  useEffect(() => {
    if (pr.type !== 'ok' || !pr.v.server) {
      return
    }
    res.find((p) => p.data?.addr === pr.v.server)?.refetch()
  }, [res, pr])

  return res
    .map((p) => {
      const r = qResult(p)
      if (r.type === 'ok') {
        return r.v
      }
      return null
    })
    .filter((p): p is ServerInfo => p != null)
    .sort((a, b) => (less(a, b) ? -1 : 1))
}

export function useServers(gameID: number, refresh: number): QResult<{ [key: string]: Region }> {
  return useData(
    `app.Servers(${gameID})`,
    () => app.Servers(gameID),
    (p) => coerce({}, p),
    {
      refetchInterval: refresh > 0 ? refresh : false,
    },
  )
}

export function useProfiles(
  players: Profile[],
  less: (a: Profile, b: Profile) => boolean,
): Profile[] {
  const res = useQueries({
    queries: players.map(({ userID, name }) => ({
      queryKey: ['useProfiles', userID],
      queryFn: () => app.Profile(userID, name),
      select: (p: Partial<Profile>) => new Profile(p),
    })),
  })
  return res
    .map((p) => {
      const r = qResult(p)
      return r.type === 'ok' ? r.v : null
    })
    .filter((p): p is Profile => p != null)
    .sort((a, b) => (less(a, b) ? -1 : 1))
}

export function useHumanPlayerProfiles(server: string = ''): Profile[] {
  const pr = usePresence()
  const humanIDs = pr.type === 'ok' && (!server || pr.v.server === server) ? pr.v.humans : []
  return useProfiles(humanIDs, (a, b) => a.userID < b.userID).filter(({ avatarURI }) => !!avatarURI)
}
