package telemetry

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/nobonobo/easportswrc/packet"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var DefaultTimeout = 3 * time.Second

type Event struct {
	Mode         string  `json:"mode"`
	Shakedown    bool    `json:"shakedown"`
	Recording    bool    `json:"recording"`
	Result       string  `json:"result"`
	Location     string  `json:"location"`
	Route        string  `json:"route"`
	Class        string  `json:"class"`
	Manufacturer string  `json:"manufacturer"`
	Vehicle      string  `json:"vehicle"`
	Length       float64 `json:"length"`
	Current      float64 `json:"current"`
	Progress     float32 `json:"progress"`
	Time         float32 `json:"time"`
	Penalty      float32 `json:"penalty"`
}

type TyreState struct {
	ForwardLeft   string `json:"forwardLeft"`
	ForwardRight  string `json:"forwardRight"`
	BackwordLeft  string `json:"backwordLeft"`
	BackwordRight string `json:"backwordRight"`
}

type Service struct {
	addr       *net.UDPAddr
	app        *application.App
	lastPacket *packet.Packet
	timeout    *time.Timer
	countDown  int
}

func Start(ctx context.Context, addr string) (*Service, error) {
	ua, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	s := &Service{
		addr: ua,
		app:  application.Get(),
	}
	s.timeout = time.AfterFunc(DefaultTimeout, func() {
		if s.lastPacket == nil {
			return
		}
		s.app.EmitEvent("recording", Event{
			Recording:    false,
			Mode:         s.lastPacket.GameModeString(),
			Shakedown:    s.lastPacket.StageShakedown,
			Result:       s.lastPacket.StageResultStatusString(),
			Location:     s.lastPacket.Location(),
			Route:        s.lastPacket.Route(),
			Class:        s.lastPacket.VehicleClass(),
			Manufacturer: s.lastPacket.VehicleManufacturer(),
			Vehicle:      s.lastPacket.Vehicle(),
			Length:       s.lastPacket.StageLength,
			Current:      s.lastPacket.StageCurrentDistance,
			Progress:     s.lastPacket.StageProgress,
			Time:         s.lastPacket.StageResultTime,
			Penalty:      s.lastPacket.StageResultTimePenalty,
		},
		)
		s.lastPacket = nil
		//log.Printf("recording-stop: %#v", s.lastPacket)
	})
	s.timeout.Stop()
	go s.run(ctx)
	return s, nil
}

func (s *Service) run(ctx context.Context) {
	conn, err := net.ListenUDP("udp", s.addr)
	if err != nil {
		log.Print(err)
		return
	}
	log.Println("udp listening start:", s.addr)
	defer log.Println("udp listener terminated:", s.addr)
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	buf := make([]byte, 4096)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			log.Print(err)
			return
		}
		s.app.EmitEvent("raw", buf[:n])
		pkt := packet.New()
		if err := pkt.UnmarshalBinary(buf[:n]); err != nil {
			log.Print(err)
			continue
		}
		s.app.EmitEvent("packet", pkt)
		s.handler(pkt)
	}
}

func (s *Service) handler(pkt *packet.Packet) {
	defer func() {
		s.timeout.Reset(DefaultTimeout)
		s.lastPacket = pkt
	}()
	orig := s.countDown
	if s.countDown > 0 {
		s.countDown--
	}
	if orig > 0 && s.countDown == 0 {
		s.app.EmitEvent("recording", Event{
			Recording:    true,
			Shakedown:    pkt.StageShakedown,
			Mode:         pkt.GameModeString(),
			Result:       pkt.StageResultStatusString(),
			Location:     pkt.Location(),
			Route:        pkt.Route(),
			Class:        pkt.VehicleClass(),
			Manufacturer: pkt.VehicleManufacturer(),
			Vehicle:      pkt.Vehicle(),
			Length:       pkt.StageLength,
			Current:      pkt.StageCurrentDistance,
			Progress:     pkt.StageProgress,
		})
	}
	if s.lastPacket == nil || (pkt.StageCurrentDistance == 0 && s.lastPacket.StageCurrentDistance != 0) {
		s.countDown = 60
		return
	}
	if s.lastPacket == nil || pkt.StageResultStatus != s.lastPacket.StageResultStatus {
		switch pkt.StageResultStatus {
		default: // unknown
			log.Print("unknown stage result status:", pkt.StageResultStatus)
			return
		case 0: // not_finished
		case 1, 2, 3, 4, 5: // finished
			s.app.EmitEvent("finished", Event{
				Mode:         pkt.GameModeString(),
				Result:       pkt.StageResultStatusString(),
				Location:     pkt.Location(),
				Route:        pkt.Route(),
				Class:        pkt.VehicleClass(),
				Manufacturer: pkt.VehicleManufacturer(),
				Vehicle:      pkt.Vehicle(),
				Length:       pkt.StageLength,
				Current:      pkt.StageCurrentDistance,
				Progress:     pkt.StageProgress,
				Time:         pkt.StageResultTime,
				Penalty:      pkt.StageResultTimePenalty,
			})
		}
	}
	if s.lastPacket == nil || pkt.VehicleTyreStateBl != s.lastPacket.VehicleTyreStateBl ||
		pkt.VehicleTyreStateBr != s.lastPacket.VehicleTyreStateBr ||
		pkt.VehicleTyreStateFl != s.lastPacket.VehicleTyreStateFl ||
		pkt.VehicleTyreStateFr != s.lastPacket.VehicleTyreStateFr {
		s.app.EmitEvent("tyre-state", TyreState{
			ForwardLeft:   pkt.VehicleTyreState(packet.ForwardLeft),
			ForwardRight:  pkt.VehicleTyreState(packet.ForwardRight),
			BackwordLeft:  pkt.VehicleTyreState(packet.BackwordLeft),
			BackwordRight: pkt.VehicleTyreState(packet.BackwordRight),
		})
	}
}
