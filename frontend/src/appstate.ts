import { main } from '../wailsjs/go/models'

export type Region = number

type Dur = string

export class Profile implements main.Profile {
  id: number
  username: string
  clan: string
  name: string
  avatarURI: string
  avatarAlt: string

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.id = coerce(0, p.id)
    this.username = coerce('', p.username)
    this.clan = coerce('', p.clan)
    this.name = coerce('', p.name)
    this.avatarURI = coerce('', p.avatarURI)
    this.avatarAlt = /(\w)/.exec(this.name)?.[1] ?? '?'
  }

  static from(source?: unknown): Profile {
    return new Profile(source)
  }
}

export class ServerInfo implements Omit<main.ServerInfo, 'convertValues'> {
  addr: string
  name: string
  players: number
  bots: number
  restricted: boolean
  ping: number
  map: string
  game: string
  maxPlayers: number
  region: number
  country: string
  ts: Date

  sortName: string

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.addr = coerce('', p.addr)
    this.name = coerce('', p.name)
    this.players = coerce(0, p.players)
    this.bots = coerce(0, p.bots)
    this.restricted = coerce(false, p.restricted)
    this.ping = coerce(0, p.ping)
    this.map = coerce('', p.map)
    this.game = coerce('', p.game)
    this.maxPlayers = coerce(0, p.maxPlayers)
    this.region = coerce(0xff, p.region)
    this.country = coerce('', p.country)
    this.ts = new Date(coerce('', p.ts) || '0000')
    this.sortName =
      // remove prefixes [abc] | (abc) | \W+
      naturallySortable(
        this.name.replace(/^(?:\[[^\]]+\]|\([^)]+\)|[^[\]()\w]+)+/, '') || this.name,
      )
  }

  less(that: ServerInfo): boolean {
    return this.sortName < that.sortName
  }

  static orderByPlayers(a: ServerInfo, b: ServerInfo): boolean {
    if (a.players !== b.players) {
      return b.players < a.players
    }
    if (a.restricted || b.restricted) {
      return !a.restricted
    }
    return a.sortName < b.sortName
  }
}

export class Environment implements main.Environment {
  minimized: boolean
  demo: boolean
  initTab: string
  initSbText: string
  tnetPort: number

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.minimized = coerce(false, p.minimized)
    this.demo = coerce(false, p.demo)
    this.initTab = coerce('', p.initTab)
    this.initSbText = coerce('', p.initSbText)
    this.tnetPort = coerce(0, p.tnetPort)
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
}

export class AppError implements main.AppError {
  fatal: boolean
  message: string

  constructor(source?: unknown) {
    if (typeof source === 'string' || source instanceof Error) {
      this.fatal = false
      this.message = `${source}`
      return
    }
    const p = sourceObject(source)
    this.fatal = coerce(false, p.fatal)
    this.message = coerce('', p.message)
  }

  toString(): string {
    if (this.fatal) {
      return `Fatal Error: ${this.message}`
    }
    return `Error: ${this.message}`
  }

  toLogString(): string {
    return this.toString()
  }
}

export class GameInfo implements main.GameInfo {
  id: number
  title: string
  dirName: string
  iconURI: string
  heroURI: string
  bgVideoURL: string
  mapImageURL: string
  mapImageURLs: string[]
  mapNames: string[]

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.id = coerce(0, p.id)
    this.title = coerce('', p.title)
    this.dirName = coerce('', p.dirName)
    this.iconURI = coerce('', p.iconURI)
    this.heroURI = coerce('', p.heroURI)
    this.bgVideoURL = coerce('', p.bgVideoURL)
    this.mapImageURL = coerce('', p.mapImageURL)
    this.mapImageURLs = coerce([], p.mapImageURLs)
    this.mapNames = coerce([], p.mapNames)
  }
}

export class SoundInfo implements main.SoundInfo {
  name: string
  sortName: string

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.name = coerce('', p.name)
    this.sortName = this.name.toLocaleLowerCase()
  }

  less(that: SoundInfo): boolean {
    return this.sortName < that.sortName
  }

  lessPat(that: SoundInfo, pat: string): boolean {
    if (this.sortName === pat.toLowerCase()) {
      return true
    }
    if (this.sortName.startsWith(pat)) {
      if (that.sortName.startsWith(pat)) {
        return this.sortName < that.sortName
      }
      return true
    }
    if (that.sortName.startsWith(pat)) {
      return false
    }
    return this.sortName < that.sortName
  }
}

export class Presence implements Omit<main.Presence, 'convertValues'> {
  inGame: boolean
  error: string
  userID: number
  avatarURL: string
  gameID: number
  gameDir: string
  gameIconURI: string
  gameHeroURI: string
  username: string
  clan: string
  name: string
  humans: Profile[]
  bots: Profile[]
  server: string
  ts: Date

  constructor(source?: unknown) {
    const p = sourceObject(source)
    this.inGame = coerce(false, p.inGame)
    this.error = coerce('', p.error)
    this.userID = coerce(0, p.userID)
    this.avatarURL = coerce('', p.avatarURL)
    this.gameID = coerce(0, p.gameID)
    this.gameIconURI = coerce('', p.gameIconURI)
    this.gameHeroURI = coerce('', p.gameHeroURI)
    this.gameDir = coerce('', p.gameDir)
    this.username = coerce('', p.username)
    this.clan = coerce('', p.clan)
    this.name = coerce('', p.name)
    this.humans = coerce([], p.humans).map((p) => new Profile(p))
    this.bots = coerce([], p.bots).map((p) => new Profile(p))
    this.server = coerce('', p.server)
    this.ts = new Date(coerce('', p.ts) || '0000')
  }
}

export class Config implements Omit<main.Config, 'convertValues'> {
  tnetPort: number
  audioDelay: Dur
  audioLimit: Dur
  audioLimitTTS: Dur
  textLimit: number
  includeUsernames: { [key: string]: boolean }
  excludeUsernames: { [key: string]: boolean }
  hosts: { [key: string]: boolean }
  firstVoice: string
  logLevel: string
  rateLimit: Dur
  serverListMaxAge: Dur
  serverInfoMaxAge: Dur

  constructor(source?: Partial<main.Config>) {
    const p = sourceObject(source)
    this.tnetPort = coerce(0, p.tnetPort)
    this.audioDelay = coerce('', p.audioDelay)
    this.audioLimit = coerce('', p.audioLimit)
    this.audioLimitTTS = coerce('', p.audioLimitTTS)
    this.textLimit = coerce(0, p.textLimit)
    this.includeUsernames = coerce({}, p.includeUsernames)
    this.excludeUsernames = coerce({}, p.excludeUsernames)
    this.hosts = coerce({}, p.hosts)
    this.firstVoice = coerce('', p.firstVoice)
    this.logLevel = coerce('', p.logLevel)
    this.rateLimit = coerce('', p.rateLimit)
    this.serverListMaxAge = coerce('', p.serverListMaxAge)
    this.serverInfoMaxAge = coerce('', p.serverInfoMaxAge)
  }
}

export class AppState extends Config implements Omit<main.AppState, 'convertValues' | 'presence'> {
  lastUpdate: string
  addr: string
  presence: Presence
  error: AppError

  constructor(source?: Partial<main.AppState>) {
    super(source)

    const p = sourceObject(source)
    this.lastUpdate = coerce('', p.lastUpdate)
    this.addr = coerce('', p.addr)
    this.presence = new Presence(p.presence)
    this.error = new AppError(p.error)
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

function naturallySortable(s: string): string {
  // convert numbers to 00xy so 10 sorts _natural_ly after 2
  return s.toLocaleLowerCase().replaceAll(/([\d.]+)/g, (_, m) => {
    if (typeof m !== 'string') {
      return m
    }
    if (m.includes('.')) {
      return m
    }
    return parseInt(m).toString().padStart(4, '0')
  })
}

export interface Lesser<T> {
  less: (that: T) => boolean
}

export interface LesserPat<T> {
  lessPat: (that: T, pat: string) => boolean
}

export function cmpLess<T extends Lesser<T>>(a: T, b: T): number {
  return a.less(b) ? -1 : 1
}

export function cmpLessPat<T extends LesserPat<T>>(pat: string): (a: T, b: T) => number {
  return (a, b) => (a.lessPat(b, pat) ? -1 : 1)
}
