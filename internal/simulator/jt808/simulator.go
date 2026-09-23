package jt808

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/suoten/jt-simulate/internal/simulator/base"
	"github.com/suoten/jt-simulate/pkg/codec/jt808"
	"github.com/suoten/jt-simulate/pkg/types"
)

// Simulator JT808终端仿真器
type Simulator struct {
	*base.Simulator
	codec *jt808.JT808Codec
}

// New 创建JT808仿真器
func New(cfg *base.DeviceConfig) *Simulator {
	s := &Simulator{
		Simulator: base.NewSimulator(cfg),
		codec:     jt808.NewCodec(),
	}
	// 设置原始数据接收处理器——自动应答平台下发消息
	s.SetOnRawRecv(s.handlePlatformMessage)
	return s
}

// Online 上线（注册→鉴权→心跳→位置上报）
func (s *Simulator) Online(ctx context.Context) error {
	if err := s.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	// 发送注册
	if err := s.sendRegister(); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	// 等待注册应答
	time.Sleep(500 * time.Millisecond)

	// 发送鉴权
	if err := s.sendAuth(); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	// 启动心跳
	s.StartHeartbeat(ctx, s.sendHeartbeat)

	// 启动定时位置上报
	s.StartLocationReport(ctx, s.sendLocation)

	return nil
}

// Offline 下线
func (s *Simulator) Offline() {
	s.Disconnect()
}

// handlePlatformMessage 处理平台下发的消息并自动应答
func (s *Simulator) handlePlatformMessage(data []byte) {
	// 按分隔符切分可能的多帧
	frames := jt808.SplitByDelimiter(data)
	if len(frames) == 0 {
		frames = [][]byte{data}
	}

	for _, frame := range frames {
		msg, err := s.codec.Decode(frame)
		if err != nil {
			continue // 忽略无法解析的帧
		}
		s.autoRespond(msg)
	}
}

// autoRespond 根据平台消息自动应答
func (s *Simulator) autoRespond(msg *types.Message) {
	cfg := s.Config()
	respHeader := &types.MessageHeader{
		Phone:       cfg.Phone,
		SeqNum:      s.NextSeqNum(),
		Version2019: true,
		ProtocolVer: 1,
	}

	switch msg.Header.MsgID {
	case 0x8001: // 平台通用应答——不需要回复
		return

	case 0x8100: // 终端注册应答——记录鉴权码
		if regResp, ok := msg.Body.(*jt808.RegisterRespMessage); ok {
			if regResp.Result == 0 && regResp.AuthCode != "" {
				cfg.AuthCode = regResp.AuthCode
				log.Printf("[%s] 注册成功，鉴权码: %s", cfg.Phone, regResp.AuthCode)
			}
		}
		return

	case jt808.MsgIDLocationQuery: // 0x8201 查询位置——回复当前位置
		respHeader.MsgID = jt808.MsgIDLocation
		now := time.Now()
		body := &jt808.LocationMessage{
			Latitude:  cfg.StartLat,
			Longitude: cfg.StartLon,
			Altitude:  5000,
			Speed:     600,
			Direction: 180,
			Time:      now.Format("060102150405"),
		}
		s.sendResp(respHeader, body)

	case jt808.MsgIDCommand: // 0x8103 设置参数——回复设置参数应答
		respHeader.MsgID = jt808.MsgIDCommandResp
		if cmd, ok := msg.Body.(*jt808.CommandMessage); ok {
			var paramIDs []uint32
			for _, p := range cmd.Params {
				paramIDs = append(paramIDs, p.ParamID)
			}
			body := &jt808.CommandRespMessage{
				RespSeqNum: msg.Header.SeqNum,
				Result:     0, // 成功
				ParamCount: byte(len(paramIDs)),
				Params:     paramIDs,
			}
			s.sendResp(respHeader, body)
		}

	case jt808.MsgIDParamQuery: // 0x8104 查询参数——回复参数查询应答
		respHeader.MsgID = jt808.MsgIDParamResp
		body := &jt808.ParamRespMessage{
			Params: []jt808.ParamItem{},
		}
		s.sendResp(respHeader, body)

	case jt808.MsgIDTerminalCtrl: // 0x8105 终端控制——回复通用应答
		respHeader.MsgID = jt808.MsgIDTerminalGeneralResp
		body := &jt808.TerminalGeneralRespMessage{
			RespSeqNum: msg.Header.SeqNum,
			RespMsgID:  msg.Header.MsgID,
			Result:     0, // 成功
		}
		s.sendResp(respHeader, body)

	case jt808.MsgIDTextSend: // 0x8300 文本下发——回复通用应答
		respHeader.MsgID = jt808.MsgIDTerminalGeneralResp
		body := &jt808.TerminalGeneralRespMessage{
			RespSeqNum: msg.Header.SeqNum,
			RespMsgID:  msg.Header.MsgID,
			Result:     0,
		}
		s.sendResp(respHeader, body)

	case jt808.MsgIDVehicleControl: // 0x8500 车辆控制——回复通用应答
		respHeader.MsgID = jt808.MsgIDTerminalGeneralResp
		body := &jt808.TerminalGeneralRespMessage{
			RespSeqNum: msg.Header.SeqNum,
			RespMsgID:  msg.Header.MsgID,
			Result:     0,
		}
		s.sendResp(respHeader, body)

	case jt808.MsgIDPhotoCommand: // 0x8801 摄像头拍摄——回复通用应答
		respHeader.MsgID = jt808.MsgIDTerminalGeneralResp
		body := &jt808.TerminalGeneralRespMessage{
			RespSeqNum: msg.Header.SeqNum,
			RespMsgID:  msg.Header.MsgID,
			Result:     0,
		}
		s.sendResp(respHeader, body)

	case jt808.MsgIDTempLocationTrack: // 0x8202 临时位置跟踪——回复通用应答并开始跟踪
		respHeader.MsgID = jt808.MsgIDTerminalGeneralResp
		body := &jt808.TerminalGeneralRespMessage{
			RespSeqNum: msg.Header.SeqNum,
			RespMsgID:  msg.Header.MsgID,
			Result:     0,
		}
		s.sendResp(respHeader, body)

	default:
		// 其他平台下发消息——回复通用应答
		respHeader.MsgID = jt808.MsgIDTerminalGeneralResp
		body := &jt808.TerminalGeneralRespMessage{
			RespSeqNum: msg.Header.SeqNum,
			RespMsgID:  msg.Header.MsgID,
			Result:     0,
		}
		s.sendResp(respHeader, body)
	}
}

// sendResp 发送应答消息
func (s *Simulator) sendResp(header *types.MessageHeader, body types.MessageBody) {
	data, err := s.codec.Encode(header, body)
	if err != nil {
		log.Printf("[%s] 编码应答失败: %v", s.Config().Phone, err)
		return
	}
	if err := s.Send(data); err != nil {
		log.Printf("[%s] 发送应答失败: %v", s.Config().Phone, err)
	}
}

// sendRegister 发送注册消息
func (s *Simulator) sendRegister() error {
	cfg := s.Config()
	header := &types.MessageHeader{
		MsgID:       jt808.MsgIDRegister,
		Phone:       cfg.Phone,
		SeqNum:      s.NextSeqNum(),
		Version2019: true,
		ProtocolVer: 1,
	}

	body := &jt808.RegisterMessage{
		ProvinceID:    uint16(cfg.ProvinceID),
		CityID:        uint16(cfg.CityID),
		Manufacturer:  cfg.Manufacturer,
		TerminalModel: cfg.TerminalModel,
		TerminalID:    cfg.TerminalID,
		PlateColor:    cfg.PlateColor,
		PlateNumber:   cfg.Plate,
	}

	data, err := s.codec.Encode(header, body)
	if err != nil {
		return err
	}

	return s.Send(data)
}

// sendAuth 发送鉴权消息
func (s *Simulator) sendAuth() error {
	cfg := s.Config()
	header := &types.MessageHeader{
		MsgID:       jt808.MsgIDAuth,
		Phone:       cfg.Phone,
		SeqNum:      s.NextSeqNum(),
		Version2019: true,
		ProtocolVer: 1,
	}

	body := &jt808.AuthMessage{
		AuthCode: cfg.AuthCode,
		IMEI:     cfg.IMEI,
	}

	data, err := s.codec.Encode(header, body)
	if err != nil {
		return err
	}

	return s.Send(data)
}

// sendHeartbeat 发送心跳
func (s *Simulator) sendHeartbeat() error {
	cfg := s.Config()
	header := &types.MessageHeader{
		MsgID:       jt808.MsgIDHeartbeat,
		Phone:       cfg.Phone,
		SeqNum:      s.NextSeqNum(),
		Version2019: true,
		ProtocolVer: 1,
	}

	body := &jt808.HeartbeatMessage{}

	data, err := s.codec.Encode(header, body)
	if err != nil {
		return err
	}

	return s.Send(data)
}

// sendLocation 发送位置上报
func (s *Simulator) sendLocation() error {
	cfg := s.Config()
	header := &types.MessageHeader{
		MsgID:       jt808.MsgIDLocation,
		Phone:       cfg.Phone,
		SeqNum:      s.NextSeqNum(),
		Version2019: true,
		ProtocolVer: 1,
	}

	now := time.Now()
	timeStr := now.Format("060102150405")

	body := &jt808.LocationMessage{
		Latitude:  cfg.StartLat,
		Longitude: cfg.StartLon,
		Altitude:  5000,
		Speed:     600, // 60.0 km/h (单位0.1km/h)
		Direction: 180,
		Time:      timeStr,
	}

	data, err := s.codec.Encode(header, body)
	if err != nil {
		return err
	}

	return s.Send(data)
}

// SendLocation 发送自定义位置
func (s *Simulator) SendLocation(lat, lon float64, speed, direction uint16) error {
	cfg := s.Config()
	header := &types.MessageHeader{
		MsgID:       jt808.MsgIDLocation,
		Phone:       cfg.Phone,
		SeqNum:      s.NextSeqNum(),
		Version2019: true,
		ProtocolVer: 1,
	}

	now := time.Now()
	body := &jt808.LocationMessage{
		Latitude:  lat,
		Longitude: lon,
		Altitude:  5000,
		Speed:     speed,
		Direction: direction,
		Time:      now.Format("060102150405"),
	}

	data, err := s.codec.Encode(header, body)
	if err != nil {
		return err
	}

	return s.Send(data)
}

// SendAlarm 发送报警（通过0x0200位置上报的AlarmFlag传递报警信息）
func (s *Simulator) SendAlarm(alarmFlag uint16) error {
	cfg := s.Config()
	header := &types.MessageHeader{
		MsgID:       jt808.MsgIDLocation,
		Phone:       cfg.Phone,
		SeqNum:      s.NextSeqNum(),
		Version2019: true,
		ProtocolVer: 1,
	}

	now := time.Now()
	body := &jt808.LocationMessage{
		AlarmFlag:  uint32(alarmFlag),
		Latitude:   cfg.StartLat,
		Longitude:  cfg.StartLon,
		Altitude:   5000,
		Speed:      0,
		Direction:  0,
		Time:       now.Format("060102150405"),
	}

	data, err := s.codec.Encode(header, body)
	if err != nil {
		return err
	}

	return s.Send(data)
}

// EncodeMessage 编码消息
func (s *Simulator) EncodeMessage(msgID uint16, body types.MessageBody) ([]byte, error) {
	cfg := s.Config()
	header := &types.MessageHeader{
		MsgID:       msgID,
		Phone:       cfg.Phone,
		SeqNum:      s.NextSeqNum(),
		Version2019: true,
		ProtocolVer: 1,
	}
	return s.codec.Encode(header, body)
}

// Codec 获取编解码器
func (s *Simulator) Codec() *jt808.JT808Codec {
	return s.codec
}
