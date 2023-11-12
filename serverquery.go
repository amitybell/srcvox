package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/amitybell/memio"
	"io"
	"net"
	"net/netip"
	"sort"
	"sync"
	"time"
)

const (
	SourceMasterServerAddr = "hl2master.steampowered.com:27011"
	serverListCacheMaxAge  = 6 * time.Hour
	serverInfoCacheMaxAge  = 2 * time.Minute
)

var (
	masterServerFirstIP = [4]byte{255, 255, 255, 255}
	masterServerLastIP  = [4]byte{0, 0, 0, 0}
	ErrNoServers        = errors.New("No servers found")
	ErrChallenge        = errors.New("Challenge")
	ErrUnknownPrefix    = errors.New("Unknown Prefix")
	ErrUnknownHeader    = errors.New("Unknown Header")
)

const (
	USEastCoast    = 0x00
	USWestCoast    = 0x01
	SouthAmerica   = 0x02
	Europe         = 0x03
	Asia           = 0x04
	Australia      = 0x05
	MiddleEast     = 0x06
	Africa         = 0x07
	RestOfTheworld = 0xFF
)

type ServerInfo struct {
	Addr       string `json:"addr"`
	Name       string `json:"name"`
	Players    int    `json:"players"`
	Bots       int    `json:"bots"`
	Restricted bool   `json:"restricted"`
	PingMs     int    `json:"ping"`
}

type MasterQuery struct {
	Header byte
	Region byte
	Addr   string
	Filter string

	buf memio.File
}

func (mq *MasterQuery) Encode(w io.Writer) error {
	r := &mq.buf
	r.Seek(0, 0)
	r.WriteByte(mq.Header)
	r.WriteByte(mq.Region)
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
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

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

func queryRegionServerList(gameID uint64, region byte) ([]*MasterReply, error) {
	conn, err := dialUDP(SourceMasterServerAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	mq := MasterQuery{
		Header: '1',
		Addr:   "0.0.0.0:0",
		Region: region,
		Filter: fmt.Sprintf(`\appid\%d`, gameID),
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
			reply := &MasterReply{}
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

func queryServerList(gameID uint64) ([][]*MasterReply, error) {
	regions := []byte{
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
	replies := make([][]*MasterReply, len(regions))

	wg := sync.WaitGroup{}
	for i, _ := range regions {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			replies[i], _ = queryRegionServerList(gameID, regions[i])
		}(i)
	}
	wg.Wait()

	return replies, nil
}

func serverList(db *DB, gameID uint64) ([]string, error) {
	key := fmt.Sprintf("/serverList/%d", gameID)
	return Cache(db, serverListCacheMaxAge, key, func() ([]string, error) {
		replies, err := queryServerList(gameID)
		if err != nil {
			return nil, fmt.Errorf("serverList: %s", err)
		}

		seen := map[string]bool{}
		var addrs []string
		for _, l := range replies {
			for _, r := range l {
				addr := r.String()
				if seen[addr] {
					continue
				}
				seen[addr] = true
				addrs = append(addrs, addr)
			}
		}
		if len(addrs) == 0 {
			return nil, fmt.Errorf("serverList: %s", ErrNoServers)
		}
		return addrs, nil
	})
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

func queryServerInfo(addr string) (_r *ServerReply, _err error) {
	conn, err := dialUDP(addr)
	if err != nil {
		return nil, fmt.Errorf("queryServerInfo(%s): %w", addr, err)
	}
	defer conn.Close()

	buf := memio.NewFile(nil)

	var challenge int32
	for {
		reply := &ServerReply{}
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

func serverInfo(db *DB, addr string) (*ServerReply, error) {
	key := "/serverInfo/" + addr
	return Cache(db, serverInfoCacheMaxAge, key, func() (*ServerReply, error) {
		return queryServerInfo(addr)
	})
}

func ServerInfos(db *DB, gameID uint64) ([]ServerInfo, error) {
	addrs, err := serverList(db, gameID)
	// the info is still useful even if some requests fail
	if len(addrs) == 0 && err != nil {
		return nil, fmt.Errorf("ServerInfos: %w", err)
	}

	queue := make(chan string, len(addrs))
	for _, s := range addrs {
		queue <- s
	}
	close(queue)

	results := make(chan ServerInfo)
	workers := sync.WaitGroup{}
	for i := 0; i < 32; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()

			for addr := range queue {
				inf, err := serverInfo(db, addr)
				if err != nil {
					continue
				}
				results <- ServerInfo{
					Addr:       addr,
					Name:       inf.Name,
					Players:    int(inf.Players),
					Bots:       int(inf.Bots),
					Restricted: inf.Visibility != 0,
					PingMs:     int(inf.Ping / time.Millisecond),
				}
			}
		}()
	}
	go func() {
		workers.Wait()
		close(results)
	}()

	infos := make([]ServerInfo, 0, len(addrs))
	for inf := range results {
		infos = append(infos, inf)
	}
	sort.Slice(infos, func(i, j int) bool {
		p, q := infos[i], infos[j]
		if p.Players != q.Players {
			return q.Players < p.Players
		}
		return p.Name < q.Name
	})
	return infos, nil
}
