package main

import (
	"bufio"
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
	serverListCacheMaxAge  = 24 * time.Hour
	serverInfoCacheMaxAge  = 5 * time.Minute
)

var (
	masterServerFirstIP = [4]byte{255, 255, 255, 255}
	masterServerLastIP  = [4]byte{0, 0, 0, 0}
	ErrNoServers        = errors.New("No servers found")
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

func (mr *MasterReply) Decode(r io.Reader) error {
	if _, err := io.ReadFull(r, mr.ip[:]); err != nil {
		return err
	}
	port := [2]byte{}
	if _, err := io.ReadFull(r, port[:]); err != nil {
		return err
	}
	mr.port = binary.BigEndian.Uint16(port[:])
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
		return nil, err
	}
	return c.(*net.UDPConn), nil
}

func readMsgUDP(conn *net.UDPConn, buf *memio.File) error {
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	const minBfSize = 2 << 10
	s := buf.Bytes()
	if len(s) < minBfSize {
		s = make([]byte, minBfSize)
	}
	n, _, err := conn.ReadFromUDP(s)
	if err != nil {
		return err
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
		Filter: fmt.Sprintf(`\dedicated\\appid\%d`, gameID),
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
	Challenge uint32

	buf memio.File
}

func (sq *ServerQuery) Encode(w io.Writer) error {
	r := &sq.buf
	r.Seek(0, 0)
	r.WriteByte(0xFF)
	r.WriteByte(0xFF)
	r.WriteByte(0xFF)
	r.WriteByte(0xFF)
	r.WriteByte(sq.Header)
	r.WriteString(sq.Payload)
	r.WriteByte(0)
	if sq.Challenge != 0 {
		r.WriteUint32(binary.LittleEndian, sq.Challenge)
	}
	r.Seek(0, 0)
	_, err := io.Copy(w, r)
	return err
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
	ID uint16

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

	Ping time.Duration

	br bufio.Reader
}

func (sr *ServerReply) Decode(r io.Reader) error {
	br := &sr.br
	br.Reset(r)

	pfx := [4]byte{}
	if _, err := io.ReadFull(br, pfx[:]); err != nil {
		return err
	}

	var err error
	sr.Header, err = br.ReadByte()
	if err != nil {
		return err
	}

	sr.Protocol, err = br.ReadByte()
	if err != nil {
		return err
	}

	sr.Name, err = br.ReadString(0)
	if err != nil {
		return err
	}

	sr.Map, err = br.ReadString(0)
	if err != nil {
		return err
	}

	sr.Folder, err = br.ReadString(0)
	if err != nil {
		return err
	}

	sr.Game, err = br.ReadString(0)
	if err != nil {
		return err
	}

	id := [2]byte{}
	if _, err := io.ReadFull(br, id[:]); err != nil {
		return err
	}
	sr.ID = binary.LittleEndian.Uint16(id[:])

	sr.Players, err = br.ReadByte()
	if err != nil {
		return err
	}

	sr.MaxPlayers, err = br.ReadByte()
	if err != nil {
		return err
	}

	sr.Bots, err = br.ReadByte()
	if err != nil {
		return err
	}

	sr.ServerType, err = br.ReadByte()
	if err != nil {
		return err
	}

	sr.Environment, err = br.ReadByte()
	if err != nil {
		return err
	}

	sr.Visibility, err = br.ReadByte()
	if err != nil {
		return err
	}

	sr.VAC, err = br.ReadByte()
	if err != nil {
		return err
	}

	return nil
}

func queryServerInfo(addr string) (*ServerReply, error) {
	conn, err := dialUDP(addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	query := &ServerQuery{Header: 'T', Payload: "Source Engine Query"}

	buf := memio.NewFile(nil)
	reply := &ServerReply{}

	pingStart := time.Now()
	if err := query.Encode(conn); err != nil {
		return nil, err
	}
	if err := readMsgUDP(conn, buf); err != nil {
		return nil, err
	}
	reply.Ping = time.Since(pingStart)

	if err := reply.Decode(buf); err != nil {
		return nil, err
	}

	return reply, nil
}

func serverInfo(db *DB, addr string) (*ServerReply, error) {
	key := "/serverInfo/" + addr
	return Cache(db, serverInfoCacheMaxAge, key, func() (*ServerReply, error) {
		return queryServerInfo(addr)
	})
}

func ServerInfos(db *DB, gameID uint64) ([]ServerInfo, error) {
	addrs, err := serverList(db, 1012110)
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
