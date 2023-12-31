// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {main} from '../models';

export function AppAddr():Promise<string>;

export function Env():Promise<main.Environment>;

export function Error():Promise<main.AppError>;

export function Games():Promise<Array<main.GameInfo>>;

export function InGame(arg1:number):Promise<main.InGame>;

export function Log(arg1:main.APILog):Promise<void>;

export function Presence():Promise<main.Presence>;

export function Profile(arg1:number,arg2:string):Promise<main.Profile>;

export function ServerInfo(arg1:main.Region,arg2:string):Promise<main.ServerInfo>;

export function Servers(arg1:number):Promise<{[key: string]: main.Region}>;

export function Sounds():Promise<Array<main.SoundInfo>>;

export function State():Promise<main.AppState>;
