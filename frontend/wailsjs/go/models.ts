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
	    // Go type: SliceSet[string]
	    humans: any;
	    // Go type: SliceSet[string]
	    bots: any;
	
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
	    tnetPort: number;
	    audioDelay: number;
	    audioLimit: number;
	    audioLimitTTS: number;
	    textLimit: number;
	    includeUsernames: {[key: string]: boolean};
	    excludeUsernames: {[key: string]: boolean};
	    hosts: {[key: string]: boolean};
	    firstVoice: string;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lastUpdate = this.convertValues(source["lastUpdate"], null);
	        this.presence = this.convertValues(source["presence"], Presence);
	        this.error = this.convertValues(source["error"], AppError);
	        this.tnetPort = source["tnetPort"];
	        this.audioDelay = source["audioDelay"];
	        this.audioLimit = source["audioLimit"];
	        this.audioLimitTTS = source["audioLimitTTS"];
	        this.textLimit = source["textLimit"];
	        this.includeUsernames = source["includeUsernames"];
	        this.excludeUsernames = source["excludeUsernames"];
	        this.hosts = source["hosts"];
	        this.firstVoice = source["firstVoice"];
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
	    minimized: boolean;
	    demo: boolean;
	    initTab: string;
	    initSbText: string;
	
	    static createFrom(source: any = {}) {
	        return new Environment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.minimized = source["minimized"];
	        this.demo = source["demo"];
	        this.initTab = source["initTab"];
	        this.initSbText = source["initSbText"];
	    }
	}
	export class GameInfo {
	    id: number;
	    title: string;
	    dirName: string;
	    iconURI: string;
	    heroURI: string;
	    mapImageURL: string;
	    bgVideoURL: string;
	
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
	    map: string;
	    game: string;
	    maxPlayers: number;
	    region: number;
	    country: string;
	
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

}

