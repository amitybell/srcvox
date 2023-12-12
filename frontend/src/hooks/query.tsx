import {
  QueryFunction,
  UndefinedInitialDataOptions,
  useQueries,
  useQuery,
  UseQueryResult,
} from '@tanstack/react-query'
import { ReactElement, useId } from 'react'
import Err from '../Error'
import {
  AppError,
  AppState,
  GameInfo,
  InGame,
  SoundInfo,
  Presence,
  coerce,
  Environment,
  ServerInfo,
  Region,
  Profile,
} from '../appstate'
import { app, useAppEvent } from '../api'
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
  if (!data) {
    throw new Error(`data is ${data}`)
  }
  return { type: 'ok', v: data, refetch }
}

export function useData<T, U>(
  key: string | string[],
  queryFn: QueryFunction<U>,
  select: (data: U) => T,
  extraOptions: Omit<UndefinedInitialDataOptions<U>, 'queryKey' | 'queryFn'> = {},
): QResult<T> {
  const opts = {
    ...extraOptions,
    queryKey: typeof key === 'string' ? [key] : key,
    queryFn,
    select,
  }
  return qResult(useQuery(opts))
}

export function useFetchDataURI(
  key: string | string[],
  url: string,
  params: Record<string, string> = {},
  extraOptions: Omit<UndefinedInitialDataOptions<string>, 'queryKey' | 'queryFn'> = {},
): QResult<string> {
  return useData(
    key,
    async ({ signal }) => {
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

export function useInGame(p: { gameID: number; refresh: number }): QResult<InGame> {
  return useData(
    'app.InGame',
    () => app.InGame(p.gameID),
    (p) => new InGame(p),
    {
      refetchInterval: p.refresh > 0 ? p.refresh : false,
    },
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

export function useSounds(): QResult<SoundInfo[]> {
  return useData('app.Sounds', app.Sounds, (p) => coerce([], p).map((q) => new SoundInfo(q)))
}

export function useAppState(): QResult<AppState> {
  return useData('app.State', app.State, (p) => new AppState(p))
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
  useAppEvent('sv.PresenceChange', () => {
    pr.refetch()
  })
  return pr
}

export function useEnv(): QResult<Environment> {
  return useData('app.Env', app.Env, (p) => new Environment(p))
}

export function useServerInfo(region: Region, addr: string, refresh: number): QResult<ServerInfo> {
  return useData(
    `app.ServerInfo(${addr})`,
    () => app.ServerInfo(region, addr),
    (p) => new ServerInfo(p),
    {
      refetchInterval: refresh > 0 ? refresh : false,
    },
  )
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
      select: (p: unknown) => new ServerInfo(p),
      refetchInterval: refresh > 0 ? refresh : undefined,
    })),
  })
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
    queries: players.map(({ id, name }) => ({
      queryKey: ['useProfiles', id],
      queryFn: () => app.Profile(id, name),
      select: (p: unknown) => new Profile(p),
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
  return useProfiles(humanIDs, (a, b) => a.id < b.id).filter(({ avatarURI }) => !!avatarURI)
}
