import {
  QueryFunction,
  UndefinedInitialDataOptions,
  useQuery,
  UseQueryResult,
} from '@tanstack/react-query'
import { ReactElement } from 'react'
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

export function useData<T, U>(
  key: string | string[],
  query: QueryFunction<U>,
  consruct: (data: U) => T,
  extraOptions: Omit<UndefinedInitialDataOptions<U>, 'queryKey' | 'queryFn'> = {},
): QResult<T> {
  const opts = {
    ...extraOptions,
    queryKey: typeof key === 'string' ? [key] : key,
    queryFn: query,
  }
  const { isLoading, error, data, refetch } = useQuery<U>(opts)
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
    return { type: 'error', error, alt: <Err>{`${error}`}</Err>, refetch }
  }
  if (!data) {
    throw new Error(`data is ${data}`)
  }
  return { type: 'ok', v: consruct(data), refetch }
}

export function useFetch<T, U>(
  key: string | string[],
  url: string,
  params: Record<string, string> = {},
  consruct: (data: U) => T,
  extraOptions: Omit<UndefinedInitialDataOptions<U>, 'queryKey' | 'queryFn'> = {},
): QResult<T> {
  return useData(
    key,
    async ({ signal }) => {
      const u = new URL(url)
      for (const [k, v] of Object.entries(params)) {
        u.searchParams.set(k, v)
      }
      const r = await fetch(u, {
        signal,
        headers: {
          pragma: 'no-cache',
          'cache-control': 'no-cache',
        },
      })
      return await r.json()
    },
    consruct,
    extraOptions,
  )
}

export function useFetchString(
  key: string | string[],
  url: string,
  params: Record<string, string> = {},
  extraOptions: Omit<UndefinedInitialDataOptions<string>, 'queryKey' | 'queryFn'> = {},
): QResult<string> {
  return useFetch<string, string>(key, url, params, (s) => s, extraOptions)
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
  return useFetchString(`app.Synthesize(${text})`, '/app.synthesize', { text })
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

export function useServerInfos(gameID: number, refresh: number): QResult<ServerInfo[]> {
  return useData(
    'app.ServerInfos',
    () => app.ServerInfos(gameID),
    (p) => coerce([], p).map((q) => new ServerInfo(q)),
    {
      refetchInterval: refresh > 0 ? refresh : false,
    },
  )
}
