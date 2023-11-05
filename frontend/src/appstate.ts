import { main } from '../wailsjs/go/models'

export class Environment implements main.Environment {
  startMinimized: boolean
  fakeData: boolean
  defaultTab: string

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.startMinimized = coerce(false, p.startMinimized)
    this.fakeData = coerce(false, p.fakeData)
    this.defaultTab = coerce('', p.defaultTab)
  }
}

export class InGame implements main.InGame {
  error: string
  count: number

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.error = coerce('', p.error)
    this.count = coerce(0, p.count)
  }

  equal(that: InGame): boolean {
    return this.error === that.error && this.count === that.count
  }
}

export class AppError implements main.AppError {
  fatal: boolean
  message: string

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.fatal = coerce(false, p.fatal)
    this.message = coerce('', p.message)
  }

  equal(that: AppError): boolean {
    return this.fatal === that.fatal && this.message === that.message
  }
}

export class GameInfo implements main.GameInfo {
  id: number
  title: string
  dirName: string
  iconURI: string

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.id = coerce(0, p.id)
    this.title = coerce('', p.title)
    this.dirName = coerce('', p.dirName)
    this.iconURI = coerce('', p.iconURI)
  }

  equal(that: GameInfo): boolean {
    return (
      this.id === that.id &&
      this.title === that.title &&
      this.dirName === that.dirName &&
      this.iconURI === that.iconURI
    )
  }
}

export class SoundInfo implements main.SoundInfo {
  name: string

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.name = coerce('', p.name)
  }

  equal(that: SoundInfo): boolean {
    return this.name === that.name
  }
}

export class Presence implements main.Presence {
  ok: boolean
  error: string
  userID: number
  gameID: number
  gameDir: string
  iconURI: string
  username: string
  clan: string
  name: string

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.ok = coerce(false, p.ok)
    this.error = coerce('', p.error)
    this.userID = coerce(0, p.userID)
    this.gameID = coerce(0, p.gameID)
    this.iconURI = coerce('', p.iconURI)
    this.gameDir = coerce('', p.gameDir)
    this.username = coerce('', p.username)
    this.clan = coerce('', p.clan)
    this.name = coerce('', p.name)
  }

  equal(that: Presence): boolean {
    return (
      this.ok === that.ok &&
      this.error === that.error &&
      this.userID === that.userID &&
      this.gameID === that.gameID &&
      this.gameDir === that.gameDir &&
      this.iconURI === that.iconURI &&
      this.username === that.username &&
      this.clan === that.clan &&
      this.name === that.name
    )
  }
}

export class AppState implements Omit<main.AppState, 'convertValues'> {
  lastUpdate: string
  presence: Presence
  games: GameInfo[]
  sounds: SoundInfo[]
  username: string
  clan: string
  name: string
  error: AppError
  inGame: InGame
  tnetPort: number
  audioDelay: number
  audioLimit: number
  includeUsernames: { [key: string]: boolean }
  excludeUsernames: { [key: string]: boolean }

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.lastUpdate = coerce('', p.lastUpdate)
    this.presence = new Presence(p.presence)
    this.games = coerce([], p.games).map((p) => new GameInfo(p))
    this.sounds = coerce([], p.sounds).map((p) => new SoundInfo(p))
    this.username = coerce('', p.username)
    this.clan = coerce('', p.clan)
    this.name = coerce('', p.name)
    this.error = new AppError(p.error)
    this.inGame = new InGame(p.inGame)
    this.tnetPort = coerce(0, p.tnetPort)
    this.audioDelay = coerce(0, p.audioDelay)
    this.audioLimit = coerce(0, p.audioLimit)
    this.includeUsernames = coerce({}, p.includeUsernames)
    this.excludeUsernames = coerce({}, p.excludeUsernames)
  }

  equal(that: AppState): boolean {
    return (
      eqScalar(this.lastUpdate, that.lastUpdate) &&
      eqObject(this.presence, that.presence) &&
      eqArray(this.games, that.games) &&
      eqArray(this.sounds, that.sounds) &&
      eqScalar(this.username, that.username) &&
      eqScalar(this.clan, that.clan) &&
      eqScalar(this.name, that.name) &&
      eqObject(this.error, that.error) &&
      eqObject(this.inGame, that.inGame) &&
      eqScalar(this.tnetPort, that.tnetPort) &&
      eqScalar(this.audioDelay, that.audioDelay) &&
      eqScalar(this.audioLimit, that.audioLimit) &&
      eqRecord(this.includeUsernames, that.includeUsernames) &&
      eqRecord(this.excludeUsernames, that.excludeUsernames)
    )
  }

  merge(source: unknown): AppState {
    const p = this
    const q = new AppState(source)
    if (eqScalar(p.lastUpdate, q.lastUpdate)) {
      q.lastUpdate = p.lastUpdate
    }
    if (eqObject(p.presence, q.presence)) {
      q.presence = p.presence
    }
    if (eqArray(p.games, q.games)) {
      q.games = p.games
    }
    if (eqScalar(p.username, q.username)) {
      q.username = p.username
    }
    if (eqScalar(p.clan, q.clan)) {
      q.clan = p.clan
    }
    if (eqScalar(p.name, q.name)) {
      q.name = p.name
    }
    if (eqObject(p.error, q.error)) {
      q.error = p.error
    }
    if (eqObject(p.inGame, q.inGame)) {
      q.inGame = p.inGame
    }
    if (eqScalar(p.tnetPort, q.tnetPort)) {
      q.tnetPort = p.tnetPort
    }
    if (eqScalar(p.audioDelay, q.audioDelay)) {
      q.audioDelay = p.audioDelay
    }
    if (eqScalar(p.audioLimit, q.audioLimit)) {
      q.audioLimit = p.audioLimit
    }
    if (eqRecord(p.includeUsernames, q.includeUsernames)) {
      q.includeUsernames = p.includeUsernames
    }
    if (eqRecord(p.excludeUsernames, q.excludeUsernames)) {
      q.excludeUsernames = p.excludeUsernames
    }
    if (p.equal(q)) {
      return p
    }
    return q
  }
}

export function coerce<T extends string | number | object | boolean | Array<unknown>>(
  def: T,
  v: unknown | undefined | null,
): T {
  if (v == null) {
    return def
  }
  if (typeof v !== typeof def) {
    return def
  }
  if (Array.isArray(def) && !Array.isArray(v)) {
    return def
  }
  return v as T
}

function sourceObject(source: unknown): Record<string, unknown> {
  if (source == null) {
    return {}
  }
  if (typeof source === 'object') {
    return source as Record<string, unknown>
  }
  if (typeof source === 'string') {
    return sourceObject(JSON.parse(source))
  }
  return {}
}

function eqScalar<T extends string | boolean | number>(p: T, q: T): boolean {
  return p === q
}

function eqArray<T extends { equal: (that: T) => boolean }>(p: T[], q: T[]): boolean {
  if (p === q) {
    return true
  }
  if (p.length !== q.length) {
    return false
  }
  return p.every((v, i) => v.equal(q[i]))
}

function eqObject<T extends { equal: (that: T) => boolean }>(p: T, q: T): boolean {
  if (p === q) {
    return true
  }
  return p.equal(q)
}

function eqRecord<T extends Record<string, U>, U extends boolean | string | number>(
  p: T,
  q: T,
): boolean {
  if (p === q) {
    return true
  }
  const pKeys = Object.keys(p)
  const qKeys = Object.keys(q)
  if (pKeys.length !== qKeys.length) {
    return false
  }
  return pKeys.every((k) => p[k] === q[k])
}
