import * as app from '../wailsjs/go/main/API'

function stringify(v: unknown): unknown {
  if (typeof v === 'string') {
    return v
  }
  if (Array.isArray(v)) {
    return v.map(stringify)
  }
  const o = v as { toLogString?: () => string }
  if (o !== null && typeof o.toLogString === 'function') {
    return o.toLogString()
  }
  return JSON.stringify(v)
}

function strings(a: unknown[]): any {
  return a.map(stringify)
}

const { log, debug, error, warn, info } = window.console

window.console.log = function (...a: unknown[]) {
  log(...a)
  app.Log(strings(a))
}

window.console.debug = function (...a: unknown[]) {
  debug(...a)
  app.Log(['DEBUG:', ...strings(a)])
}

window.console.error = function (...a: unknown[]) {
  error(...a)
  app.Log(['ERROR:', ...strings(a)])
}

window.console.warn = function (...a: unknown[]) {
  warn(...a)
  app.Log(['WARN:', ...strings(a)])
}

window.console.info = function (...a: unknown[]) {
  info(...a)
  app.Log(['INFO:', ...strings(a)])
}
