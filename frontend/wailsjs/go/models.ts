export namespace appstate {
	
	export class AppError {
	    fatal: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new AppError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fatal = source["fatal"];
	        this.message = source["message"];
	    }
	}
	export class Presence {
	    inGame: boolean;
	    error: string;
	    userID: number;
	    avatarURL: string;
	    username: string;
	    clan: string;
	    name: string;
	    gameID: number;
	    gameIconURI: string;
	    gameHeroURI: string;
	    gameDir: string;
	    // Go type: data
	    humans: any;
	    // Go type: data
	    bots: any;
	    server: string;
	    // Go type: time
	    ts: any;
	
	    static createFrom(source: any = {}) {
	        return new Presence(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inGame = source["inGame"];
	        this.error = source["error"];
	        this.userID = source["userID"];
	        this.avatarURL = source["avatarURL"];
	        this.username = source["username"];
	        this.clan = source["clan"];
	        this.name = source["name"];
	        this.gameID = source["gameID"];
	        this.gameIconURI = source["gameIconURI"];
	        this.gameHeroURI = source["gameHeroURI"];
	        this.gameDir = source["gameDir"];
	        this.humans = this.convertValues(source["humans"], null);
	        this.bots = this.convertValues(source["bots"], null);
	        this.server = source["server"];
	        this.ts = this.convertValues(source["ts"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AppState {
	    // Go type: time
	    lastUpdate: any;
	    presence: Presence;
	    error: AppError;
	    netcon: config.ConnInfo;
	    rcon: config.ConnInfo;
	    // Go type: config
	    audioDelay: any;
	    // Go type: config
	    audioLimit: any;
	    // Go type: config
	    audioLimitTTS: any;
	    textLimit: number;
	    includeUsernames: {[key: string]: boolean};
	    excludeUsernames: {[key: string]: boolean};
	    hosts: {[key: string]: boolean};
	    firstVoice: string;
	    logLevel: string;
	    // Go type: config
	    rateLimit: any;
	    // Go type: config
	    serverListMaxAge: any;
	    // Go type: config
	    serverInfoMaxAge: any;
	    minimized?: boolean;
	    demo?: boolean;
	    tnetPort: number;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lastUpdate = this.convertValues(source["lastUpdate"], null);
	        this.presence = this.convertValues(source["presence"], Presence);
	        this.error = this.convertValues(source["error"], AppError);
	        this.netcon = this.convertValues(source["netcon"], config.ConnInfo);
	        this.rcon = this.convertValues(source["rcon"], config.ConnInfo);
	        this.audioDelay = this.convertValues(source["audioDelay"], null);
	        this.audioLimit = this.convertValues(source["audioLimit"], null);
	        this.audioLimitTTS = this.convertValues(source["audioLimitTTS"], null);
	        this.textLimit = source["textLimit"];
	        this.includeUsernames = source["includeUsernames"];
	        this.excludeUsernames = source["excludeUsernames"];
	        this.hosts = source["hosts"];
	        this.firstVoice = source["firstVoice"];
	        this.logLevel = source["logLevel"];
	        this.rateLimit = this.convertValues(source["rateLimit"], null);
	        this.serverListMaxAge = this.convertValues(source["serverListMaxAge"], null);
	        this.serverInfoMaxAge = this.convertValues(source["serverInfoMaxAge"], null);
	        this.minimized = source["minimized"];
	        this.demo = source["demo"];
	        this.tnetPort = source["tnetPort"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace config {
	
	export class ConnInfo {
	    host: string;
	    port: number;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.password = source["password"];
	    }
	}
	export class Config {
	    netcon: ConnInfo;
	    rcon: ConnInfo;
	    // Go type: Dur
	    audioDelay: any;
	    // Go type: Dur
	    audioLimit: any;
	    // Go type: Dur
	    audioLimitTTS: any;
	    textLimit: number;
	    includeUsernames: {[key: string]: boolean};
	    excludeUsernames: {[key: string]: boolean};
	    hosts: {[key: string]: boolean};
	    firstVoice: string;
	    logLevel: string;
	    // Go type: Dur
	    rateLimit: any;
	    // Go type: Dur
	    serverListMaxAge: any;
	    // Go type: Dur
	    serverInfoMaxAge: any;
	    minimized?: boolean;
	    demo?: boolean;
	    tnetPort: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.netcon = this.convertValues(source["netcon"], ConnInfo);
	        this.rcon = this.convertValues(source["rcon"], ConnInfo);
	        this.audioDelay = this.convertValues(source["audioDelay"], null);
	        this.audioLimit = this.convertValues(source["audioLimit"], null);
	        this.audioLimitTTS = this.convertValues(source["audioLimitTTS"], null);
	        this.textLimit = source["textLimit"];
	        this.includeUsernames = source["includeUsernames"];
	        this.excludeUsernames = source["excludeUsernames"];
	        this.hosts = source["hosts"];
	        this.firstVoice = source["firstVoice"];
	        this.logLevel = source["logLevel"];
	        this.rateLimit = this.convertValues(source["rateLimit"], null);
	        this.serverListMaxAge = this.convertValues(source["serverListMaxAge"], null);
	        this.serverInfoMaxAge = this.convertValues(source["serverInfoMaxAge"], null);
	        this.minimized = source["minimized"];
	        this.demo = source["demo"];
	        this.tnetPort = source["tnetPort"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace logs {
	
	export class APILog {
	    level: string;
	    message: string;
	    trace: string[];
	
	    static createFrom(source: any = {}) {
	        return new APILog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.message = source["message"];
	        this.trace = source["trace"];
	    }
	}

}

export namespace sound {
	
	export class SoundInfo {
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new SoundInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	    }
	}

}

export namespace steam {
	
	export class GameInfo {
	    id: number;
	    title: string;
	    dirName: string;
	    iconURI: string;
	    heroURI: string;
	    mapImageURL: string;
	    bgVideoURL: string;
	    mapNames: string[];
	    mapImageURLs: string[];
	
	    static createFrom(source: any = {}) {
	        return new GameInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.dirName = source["dirName"];
	        this.iconURI = source["iconURI"];
	        this.heroURI = source["heroURI"];
	        this.mapImageURL = source["mapImageURL"];
	        this.bgVideoURL = source["bgVideoURL"];
	        this.mapNames = source["mapNames"];
	        this.mapImageURLs = source["mapImageURLs"];
	    }
	}
	export class Profile {
	    userID: number;
	    avatarURI: string;
	    username: string;
	    clan: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.userID = source["userID"];
	        this.avatarURI = source["avatarURI"];
	        this.username = source["username"];
	        this.clan = source["clan"];
	        this.name = source["name"];
	    }
	}
	export class ServerInfo {
	    addr: string;
	    name: string;
	    players: number;
	    bots: number;
	    restricted: boolean;
	    ping: number;
	    map: string;
	    game: string;
	    maxPlayers: number;
	    region: number;
	    country: string;
	    // Go type: time
	    ts: any;
	
	    static createFrom(source: any = {}) {
	        return new ServerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.addr = source["addr"];
	        this.name = source["name"];
	        this.players = source["players"];
	        this.bots = source["bots"];
	        this.restricted = source["restricted"];
	        this.ping = source["ping"];
	        this.map = source["map"];
	        this.game = source["game"];
	        this.maxPlayers = source["maxPlayers"];
	        this.region = source["region"];
	        this.country = source["country"];
	        this.ts = this.convertValues(source["ts"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

