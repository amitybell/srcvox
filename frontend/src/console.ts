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
  try {
    return JSON.stringify(v)
  } catch {
    return `${v}`
  }
}

function strings(a: unknown[]): any {
  return a.map(stringify)
}

function stack(skip: number): string[] {
  return new Error().stack?.split('\n').slice(skip + 1) ?? []
}

const { log, debug, error, warn, info } = window.console

function logApp(skip: number, level: string, ...a: unknown[]) {
  const msg = strings(a).join(' ')
  if (msg === '' || msg === '{}') {
    return
  }
  app.Log({ level, message: msg, trace: stack(skip + 1) })
}

window.console.log = function (...a: unknown[]) {
  log(...a)
  logApp(1, 'info', ...a)
}

window.console.debug = function (...a: unknown[]) {
  debug(...a)
  logApp(1, 'debug', ...a)
}

window.console.error = function (...a: unknown[]) {
  error(...a)
  logApp(1, 'error', ...a)
}

window.console.warn = function (...a: unknown[]) {
  warn(...a)
  logApp(1, 'warn', ...a)
}

window.console.info = function (...a: unknown[]) {
  info(...a)
  logApp(1, 'info', ...a)
}
