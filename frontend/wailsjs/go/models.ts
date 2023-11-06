export namespace main {
	
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
	export class GameInfo {
	    id: number;
	    title: string;
	    dirName: string;
	    iconURI: string;
	    heroURI: string;
	
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
	    }
	}
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
	export class Presence {
	    ok: boolean;
	    error: string;
	    userID: number;
	    username: string;
	    clan: string;
	    name: string;
	    gameID: number;
	    gameIconURI: string;
	    gameHeroURI: string;
	    gameDir: string;
	
	    static createFrom(source: any = {}) {
	        return new Presence(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.error = source["error"];
	        this.userID = source["userID"];
	        this.username = source["username"];
	        this.clan = source["clan"];
	        this.name = source["name"];
	        this.gameID = source["gameID"];
	        this.gameIconURI = source["gameIconURI"];
	        this.gameHeroURI = source["gameHeroURI"];
	        this.gameDir = source["gameDir"];
	    }
	}
	export class AppState {
	    // Go type: time
	    lastUpdate: any;
	    presence: Presence;
	    sounds: SoundInfo[];
	    games: GameInfo[];
	    error: AppError;
	    tnetPort: number;
	    audioDelay: number;
	    audioLimit: number;
	    includeUsernames: {[key: string]: boolean};
	    excludeUsernames: {[key: string]: boolean};
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lastUpdate = this.convertValues(source["lastUpdate"], null);
	        this.presence = this.convertValues(source["presence"], Presence);
	        this.sounds = this.convertValues(source["sounds"], SoundInfo);
	        this.games = this.convertValues(source["games"], GameInfo);
	        this.error = this.convertValues(source["error"], AppError);
	        this.tnetPort = source["tnetPort"];
	        this.audioDelay = source["audioDelay"];
	        this.audioLimit = source["audioLimit"];
	        this.includeUsernames = source["includeUsernames"];
	        this.excludeUsernames = source["excludeUsernames"];
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
	export class Environment {
	    startMinimized: boolean;
	    fakeData: boolean;
	    defaultTab: string;
	
	    static createFrom(source: any = {}) {
	        return new Environment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.startMinimized = source["startMinimized"];
	        this.fakeData = source["fakeData"];
	        this.defaultTab = source["defaultTab"];
	    }
	}
	
	export class InGame {
	    error: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new InGame(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.error = source["error"];
	        this.count = source["count"];
	    }
	}
	
	export class ServerInfo {
	    addr: string;
	    name: string;
	    players: number;
	    bots: number;
	    restricted: boolean;
	    ping: number;
	
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
	    }
	}

}

