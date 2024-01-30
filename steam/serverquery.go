package steam

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/amitybell/ip2country"
	"github.com/amitybell/memio"
	"github.com/amitybell/srcvox/demo"
	"github.com/amitybell/srcvox/rng"
	"github.com/amitybell/srcvox/store"
)

const (
	SourceMasterServerAddr = "hl2master.steampowered.com:27011"
)

var (
	masterServerFirstIP = [4]byte{255, 255, 255, 255}
	masterServerLastIP  = [4]byte{0, 0, 0, 0}
	ErrNoServers        = errors.New("No servers found")
	ErrChallenge        = errors.New("Challenge")
	ErrUnknownPrefix    = errors.New("Unknown Prefix")
	ErrUnknownHeader    = errors.New("Unknown Header")

	regions = []Region{
		USEastCoast,
		USWestCoast,
		SouthAmerica,
		Europe,
		Asia,
		Australia,
		MiddleEast,
		Africa,
		RestOfTheworld,
	}
)

type Region byte

const (
	USEastCoast    Region = 0x00
	USWestCoast    Region = 0x01
	SouthAmerica   Region = 0x02
	Europe         Region = 0x03
	Asia           Region = 0x04
	Australia      Region = 0x05
	MiddleEast     Region = 0x06
	Africa         Region = 0x07
	RestOfTheworld Region = 0xFF
)

func (r Region) String() string {
	switch r {
	case USEastCoast:
		return "us-east"
	case USWestCoast:
		return "us-west"
	case SouthAmerica:
		return "south-america"
	case Europe:
		return "europe"
	case Asia:
		return "asia"
	case Australia:
		return "australia"
	case MiddleEast:
		return "middle-east"
	case Africa:
		return "africa"
	default:
		return "rest-of-world"
	}
}

type ServerInfo struct {
	Addr       string    `json:"addr"`
	Name       string    `json:"name"`
	Players    int       `json:"players"`
	Bots       int       `json:"bots"`
	Restricted bool      `json:"restricted"`
	PingMs     int       `json:"ping"`
	Map        string    `json:"map"`
	Game       string    `json:"game"`
	MaxPlayers int       `json:"maxPlayers"`
	Region     Region    `json:"region"`
	Country    string    `json:"country"`
	Ts         time.Time `json:"ts"`
}

type MasterQuery struct {
	Header byte
	Region Region
	Addr   string
	Filter string

	buf memio.File
}

func (mq *MasterQuery) Encode(w io.Writer) error {
	r := &mq.buf
	r.Seek(0, 0)
	r.WriteByte(mq.Header)
	r.WriteByte(byte(mq.Region))
	r.WriteString(mq.Addr)
	r.WriteByte(0)
	r.WriteString(mq.Filter)
	r.WriteByte(0)
	r.Seek(0, 0)
	_, err := io.Copy(w, r)
	return err
}

type MasterReply struct {
	ip   [4]byte
	port uint16

	Region Region
}

func (mr *MasterReply) Decode(r *memio.File) error {
	_, err := r.ReadFull(mr.ip[:])
	if err != nil {
		return err
	}
	mr.port, err = r.ReadUint16(binary.BigEndian)
	if err != nil {
		return err
	}
	return nil
}

func (mr *MasterReply) ID() netip.Addr {
	return netip.AddrFrom4(mr.ip)
}

func (mr *MasterReply) Port() int {
	return int(mr.port)
}

func (mr *MasterReply) String() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", mr.ip[0], mr.ip[1], mr.ip[2], mr.ip[3], mr.port)
}

func dialUDP(addr string) (*net.UDPConn, error) {
	c, err := net.Dial("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("dialUDP: %w", err)
	}
	return c.(*net.UDPConn), nil
}

func readMsgUDP(conn *net.UDPConn, buf *memio.File) error {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	const minBfSize = 2 << 10
	s := buf.Bytes()
	if len(s) < minBfSize {
		s = make([]byte, minBfSize)
	}
	n, _, err := conn.ReadFromUDP(s)
	if err != nil {
		return fmt.Errorf("readMsgUDP: %w", err)
	}
	buf.Reset(s[:n])
	return nil
}

func queryRegionServerList(gameID ID, region Region) ([]*MasterReply, error) {
	conn, err := dialUDP(SourceMasterServerAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	mq := MasterQuery{
		Header: '1',
		Addr:   "0.0.0.0:0",
		Region: region,
		Filter: fmt.Sprintf(`\appid\%d`, gameID.To32()),
	}

	if err := mq.Encode(conn); err != nil {
		return nil, err
	}

	var replies []*MasterReply

	buf := memio.NewFile(nil)
	for {
		if err := readMsgUDP(conn, buf); err != nil {
			if errors.Is(err, io.EOF) {
				return replies, nil
			}
			return replies, err
		}

		var last *MasterReply
		for {
			reply := &MasterReply{Region: region}
			if err := reply.Decode(buf); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return replies, err
			}
			if reply.ip == masterServerLastIP {
				return replies, nil
			}
			if reply.ip != masterServerFirstIP {
				replies = append(replies, reply)
				last = reply
			}
		}
		if last == nil {
			return replies, nil
		}
		mq.Addr = last.String()
		if err := mq.Encode(conn); err != nil {
			return replies, err
		}
	}
}

func queryServerList(gameID ID) ([][]*MasterReply, error) {
	replies := make([][]*MasterReply, len(regions))

	wg := sync.WaitGroup{}
	for i := range regions {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			replies[i], _ = queryRegionServerList(gameID, regions[i])
		}(i)
	}
	wg.Wait()

	return replies, nil
}

func QueryServerList(db *store.DB, maxAge time.Duration, gameID ID) (map[string]Region, error) {
	if demo.Enabled {
		maxAge = -1
	}
	key := fmt.Sprintf("/serverList/addr-region/%d", gameID.To32())
	m, err := store.CacheTTL(db, maxAge, key, 2, func() (map[string]Region, error) {
		replies, err := queryServerList(gameID)
		if err != nil {
			return nil, fmt.Errorf("serverList: %s", err)
		}

		addrs := map[string]Region{}
		for _, l := range replies {
			for _, r := range l {
				addrs[r.String()] = r.Region
			}
		}
		if len(addrs) == 0 {
			return nil, fmt.Errorf("serverList: %s", ErrNoServers)
		}
		return addrs, nil
	})
	if err != nil && !errors.Is(err, store.ErrStale) {
		return m, err
	}
	return m, nil
}

type ServerQuery struct {
	Header    byte
	Payload   string
	Challenge int32

	buf memio.File
}

func (sq *ServerQuery) Encode(w io.Writer) error {
	r := &sq.buf
	r.Seek(0, 0)
	r.WriteInt32(binary.LittleEndian, -1)
	r.WriteByte(sq.Header)
	r.WriteString(sq.Payload)
	r.WriteByte(0)
	if sq.Challenge != 0 {
		r.WriteInt32(binary.LittleEndian, sq.Challenge)
	}
	r.Seek(0, 0)
	_, err := io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("ServerQuery.Encode: %w", err)
	}
	return nil
}

type ServerReply struct {
	// 	Always equal to 'I' (0x49)
	Header byte

	// 	Protocol version used by the server.
	Protocol byte

	// Name of the server.
	Name string

	// Map the server has currently loaded.
	Map string

	// Name of the folder containing the game files.
	Folder string

	// Full name of the game.
	Game string

	// Steam Application ID of game.
	ID int16

	// 	Number of players on the server.
	Players byte

	// 	Maximum number of players the server reports it can hold.
	MaxPlayers byte

	// Number of bots on the server.
	Bots byte

	// Indicates the type of server:
	// 'd' for a dedicated server
	// 'l' for a non-dedicated server
	// 'p' for a SourceTV relay (proxy)
	ServerType byte

	// Indicates the operating system of the server:
	// 'l' for Linux
	// 'w' for Windows
	// 'm' or 'o' for Mac (the code changed after L4D1)
	Environment byte

	// Indicates whether the server requires a password:
	// 0 for public
	// 1 for private
	Visibility byte

	// Specifies whether the server uses VAC:
	// 0 for unsecured
	// 1 for secured
	VAC byte

	Ping      time.Duration
	Challenge int32

	Ts time.Time
}

func (sr *ServerReply) decodeChallenge(r *memio.File) error {
	var err error
	sr.Challenge, err = r.ReadInt32(binary.LittleEndian)
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Challenge: %w", err)
	}
	return ErrChallenge
}

func (sr *ServerReply) decodeInfo(r *memio.File) error {
	var err error
	sr.Protocol, err = r.ReadByte()
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Protocol: %w", err)
	}

	sr.Name, err = r.ReadString(0)
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Name: %w", err)
	}

	sr.Map, err = r.ReadString(0)
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Map: %w", err)
	}

	sr.Folder, err = r.ReadString(0)
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Folder: %w", err)
	}

	sr.Game, err = r.ReadString(0)
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Game: %w", err)
	}

	sr.ID, err = r.ReadInt16(binary.LittleEndian)
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: ID: %w", err)
	}

	sr.Players, err = r.ReadByte()
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Players: %w", err)
	}

	sr.MaxPlayers, err = r.ReadByte()
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: MaxPlayers: %w", err)
	}

	sr.Bots, err = r.ReadByte()
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Bots: %w", err)
	}

	sr.ServerType, err = r.ReadByte()
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: ServerType: %w", err)
	}

	sr.Environment, err = r.ReadByte()
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Environment: %w", err)
	}

	sr.Visibility, err = r.ReadByte()
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Visibility: %w", err)
	}

	sr.VAC, err = r.ReadByte()
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: VAC: %w", err)
	}

	return nil
}

func (sr *ServerReply) Decode(r *memio.File) error {
	pfx, err := r.ReadInt32(binary.LittleEndian)
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Pfx: %w", err)
	}
	if pfx != -1 {
		return fmt.Errorf("ServerReply.Decode: %w: %v", ErrUnknownPrefix, pfx)
	}

	sr.Header, err = r.ReadByte()
	if err != nil {
		return fmt.Errorf("ServerReply.Decode: Header: %w", err)
	}
	switch sr.Header {
	case 'I':
		return sr.decodeInfo(r)
	case 'A':
		return sr.decodeChallenge(r)
	default:
		return fmt.Errorf("ServerReply.Decode: %w: %c", ErrUnknownHeader, sr.Header)
	}
}

func queryConn(conn *net.UDPConn, query interface{ Encode(io.Writer) error }, buf *memio.File, reply interface{ Decode(*memio.File) error }) (time.Duration, error) {
	pingStart := time.Now()
	if err := query.Encode(conn); err != nil {
		return 0, fmt.Errorf("queryConn: Encode: %w", err)
	}
	if err := readMsgUDP(conn, buf); err != nil {
		return 0, fmt.Errorf("queryConn: Read: %w", err)
	}
	if err := reply.Decode(buf); err != nil {
		return 0, fmt.Errorf("queryConn: Decode: %w", err)
	}
	return time.Since(pingStart), nil
}

func queryServerInfo(region Region, addr string) (*ServerReply, error) {
	conn, err := dialUDP(addr)
	if err != nil {
		return nil, fmt.Errorf("queryServerInfo(%s): %w", addr, err)
	}
	defer conn.Close()

	buf := memio.NewFile(nil)

	var challenge int32
	for {
		reply := &ServerReply{Ts: time.Now()}
		query := &ServerQuery{Header: 'T', Payload: "Source Engine Query", Challenge: challenge}
		ping, err := queryConn(conn, query, buf, reply)
		switch {
		case err == nil:
			reply.Ping = ping
			return reply, nil
		case errors.Is(err, ErrChallenge):
			challenge = reply.Challenge
		default:
			return nil, fmt.Errorf("queryServerInfo(%s): %w", addr, err)
		}
	}
}

func QueryServerInfo(db *store.DB, maxAge time.Duration, region Region, addr string) (ServerInfo, *ServerReply, error) {
	if demo.Enabled {
		maxAge = -1
	}

	key := fmt.Sprintf("/serverInfo/%s", addr)
	rep, err := store.CacheTTL(db, maxAge, key, 2, func() (*ServerReply, error) {
		return queryServerInfo(region, addr)
	})
	if err != nil && (!errors.Is(err, store.ErrStale) || rep == nil) {
		return ServerInfo{}, nil, err
	}
	ip, _, _ := net.SplitHostPort(addr)
	cc, _ := ip2country.LookupString(ip)
	inf := ServerInfo{
		Addr:       addr,
		Name:       rep.Name,
		Players:    int(rep.Players),
		Bots:       int(rep.Bots),
		Restricted: rep.Visibility != 0,
		PingMs:     int(rep.Ping / time.Millisecond),
		Map:        rep.Map,
		Game:       rep.Game,
		MaxPlayers: int(rep.MaxPlayers),
		Region:     region,
		Country:    cc,
		Ts:         rep.Ts,
	}
	if demo.Enabled {
		inf.MaxPlayers = 32
		inf.Players = inf.MaxPlayers - rng.Intn(4)
		inf.Bots = inf.MaxPlayers - inf.Players
	}
	return inf, rep, nil
}
