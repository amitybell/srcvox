import React, { useEffect } from 'react'
import * as ap from '../wailsjs/go/main/API'
import { config } from '../wailsjs/go/models'
import { BrowserOpenURL, EventsOn } from '../wailsjs/runtime'

export type DataURI = string

export interface FetchDataResult {
  promise: Promise<DataURI>
  cancel: (reason?: string) => void
  done: boolean | Error
}

export function fetchData(path: string, params: Record<string, string>): FetchDataResult {
  const ctl = new AbortController()
  const url = `${path}?${new URLSearchParams(params)}`
  const opts = {
    sinal: ctl.signal,
    headers: {
      pragma: 'no-cache',
      'cache-control': 'no-cache',
    },
  }
  const res: FetchDataResult = {
    promise: new Promise<DataURI>((resolve, reject) => {
      fetch(url, opts)
        .then((r) => {
          r.blob()
            .then((b) => {
              res.done = true
              resolve(URL.createObjectURL(b))
            })
            .catch((e) => {
              res.done = new Error(`error: ${e}`)
              reject(e)
            })
        })
        .catch((e) => {
          res.done = new Error(`error: ${e}`)
          reject(e)
        })
    }),
    done: false,
    cancel: (reason?: string) => {
      if (res.done) {
        return
      }
      res.done = new Error(reason || 'canceled')
      ctl.abort(res.done)
    },
  }
  return res
}

export function fetchSound(text: string): FetchDataResult {
  return fetchData('/app.sound', { text })
}

export function fetchGameIcon(id: number): FetchDataResult {
  return fetchData('/app.gameicon', { id: id.toString(10) })
}

export function useAppEvent(name: string, hdl: () => void) {
  useEffect(() => {
    // the server might send data in the event, but our API doesn't account for that
    const cb = () => hdl()
    return EventsOn(name, cb)
  }, [name, hdl])
}

export function openBrowser(url: string, evt?: React.MouseEvent<unknown>) {
  if (evt) {
    evt.preventDefault()
  }
  BrowserOpenURL(url)
  return false
}

export function UpdateConfig(cfg: Partial<config.Config>): Promise<void> {
  return ap.UpdateConfig(cfg as config.Config)
}

export const app = {
  ...ap,
  UpdateConfig,
}
