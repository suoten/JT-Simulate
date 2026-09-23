package workshop

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/suoten/jt-simulate/pkg/codec/gbt32960"
	"github.com/suoten/jt-simulate/pkg/codec/jt808"
	"github.com/suoten/jt-simulate/pkg/codec/jt809"
	"github.com/suoten/jt-simulate/pkg/codec/jt1078"
	"github.com/suoten/jt-simulate/pkg/codec/jt905"
	"github.com/suoten/jt-simulate/pkg/codec/jt1045"
	"github.com/suoten/jt-simulate/pkg/codec/jt1253"
	"github.com/suoten/jt-simulate/pkg/types"
)

// Workshop 报文工坊
type Workshop struct {
	jt808Codec *jt808.JT808Codec
	jt809Codec *jt809.JT809Codec
}

// New 创建报文工坊
func New() *Workshop {
	return &Workshop{
		jt808Codec: jt808.NewCodec(),
		jt809Codec: jt809.NewCodec(),
	}
}

// EncodeRequest 编码请求
type EncodeRequest struct {
	Protocol string                 `json:"protocol"`
	Version  string                 `json:"version"`
	MsgID    string                 `json:"msg_id"`
	Phone    string                 `json:"phone"`
	Fields   map[string]interface{} `json:"fields"`
}

// EncodeResponse 编码响应
type EncodeResponse struct {
	Hex     string `json:"hex"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Encode 编码报文
func (w *Workshop) Encode(req *EncodeRequest) *EncodeResponse {
	msgID, err := parseMsgID(req.MsgID)
	if err != nil {
		return &EncodeResponse{Success: false, Error: err.Error()}
	}
	switch req.Protocol {
	case "jt808":
		return w.encodeJT808(msgID, req)
	case "jt809":
		return w.encodeJT809(msgID, req)
	case "jt1078":
		return w.encodeJT1078(msgID, req)
	case "gbt32960":
		return w.encodeGBT32960(msgID, req)
	case "jt905":
		return w.encodeJT905(msgID, req)
	case "jt1045":
		return w.encodeJT1045(msgID, req)
	case "jt1253":
		return w.encodeJT1253(msgID, req)
	default:
		return &EncodeResponse{Success: false, Error: fmt.Sprintf("unsupported protocol: %s", req.Protocol)}
	}
}

func isVer2019(ver string) bool { return ver == "" || ver == "2019" }

func (w *Workshop) encodeJT808(msgID uint16, req *EncodeRequest) *EncodeResponse {
	ver2019 := isVer2019(req.Version)
	header := &types.MessageHeader{
		MsgID: msgID, Phone: req.Phone, SeqNum: 1,
		Version2019: ver2019,
	}
	if ver2019 {
		header.ProtocolVer = 1
	}
	body, err := w.buildJT808Body(msgID, req.Fields, ver2019)
	if err != nil {
		return &EncodeResponse{Success: false, Error: err.Error()}
	}
	data, err := w.jt808Codec.Encode(header, body)
	if err != nil {
		return &EncodeResponse{Success: false, Error: err.Error()}
	}
	return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
}

func (w *Workshop) buildJT808Body(msgID uint16, f map[string]interface{}, ver2019 bool) (types.MessageBody, error) {
	switch msgID {
	// ==================== 通用 ====================
	case jt808.MsgIDTerminalGeneralResp:
		return &jt808.TerminalGeneralRespMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")), RespMsgID: uint16(getInt(f, "resp_msg_id")),
			Result: getByte(f, "result"),
		}, nil
	case 0x8001: // 平台通用应答
		return &jt808.PlatformGeneralRespMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")), RespMsgID: uint16(getInt(f, "resp_msg_id")),
			Result: getByte(f, "result"),
		}, nil
	case jt808.MsgIDHeartbeat:
		return &jt808.HeartbeatMessage{}, nil
	case jt808.MsgIDTerminalCancel:
		return &jt808.RawMessage{ID: msgID}, nil

	// ==================== 注册与鉴权 ====================
	case jt808.MsgIDRegister:
		return &jt808.RegisterMessage{
			ProvinceID: uint16(getInt(f, "province_id")), CityID: uint16(getInt(f, "city_id")),
			Manufacturer: getString(f, "manufacturer"), TerminalModel: getString(f, "terminal_model"),
			TerminalID: getString(f, "terminal_id"), PlateColor: getByte(f, "plate_color"),
			PlateNumber: getString(f, "plate_number"),
		}, nil
	case jt808.MsgIDRegisterResp:
		return &jt808.RegisterRespMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")), Result: getByte(f, "result"),
			AuthCode: getString(f, "auth_code"),
		}, nil
	case jt808.MsgIDAuth:
		auth := &jt808.AuthMessage{AuthCode: getString(f, "auth_code")}
		if ver2019 {
			auth.IMEI = getString(f, "imei")
		}
		return auth, nil

	// ==================== 参数设置/查询 ====================
	case jt808.MsgIDCommand:
		return &jt808.CommandMessage{
			Params: []jt808.ParamItem{{
				ParamID: uint32(getInt(f, "param_id")),
				ParamLen: byte(len(getString(f, "param_value"))),
				ParamValue: []byte(getString(f, "param_value")),
			}},
		}, nil
	case jt808.MsgIDCommandResp:
		return &jt808.CommandRespMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")), Result: getByte(f, "result"),
			ParamCount: getByte(f, "param_count"),
		}, nil
	case jt808.MsgIDParamQuery:
		return &jt808.ParamQueryMessage{
			ParamIDs: []uint32{uint32(getInt(f, "param_id"))},
		}, nil
	case jt808.MsgIDParamResp:
		return &jt808.ParamRespMessage{
			Params: []jt808.ParamItem{{
				ParamID: uint32(getInt(f, "param_id")),
				ParamLen: byte(len(getString(f, "param_value"))),
				ParamValue: []byte(getString(f, "param_value")),
			}},
		}, nil
	case jt808.MsgIDTerminalCtrl:
		return &jt808.TerminalCtrlMessage{Command: getByte(f, "command"), Param: getString(f, "param")}, nil
	case jt808.MsgIDTerminalPropQuery:
		return &jt808.TerminalPropQueryMessage{}, nil
	case jt808.MsgIDTerminalPropResp:
		return &jt808.TerminalPropRespMessage{
			TerminalType:   uint16(getInt(f, "terminal_type")),
			ManufacturerID: getString(f, "manufacturer"), TerminalModel: getString(f, "terminal_model"),
			TerminalID: getString(f, "terminal_id"), ICCID: getString(f, "iccid"),
		}, nil
	case jt808.MsgIDTerminalUpgrade:
		return &jt808.TerminalUpgradeMessage{
			UpgradeType: getByte(f, "upgrade_type"), Manufacturer: getString(f, "manufacturer"),
			Model: getString(f, "model"), Version: getString(f, "version"),
			URL: getString(f, "url"),
		}, nil
	case jt808.MsgIDTerminalUpgradeResp:
		return &jt808.TerminalUpgradeRespMessage{
			UpgradeType: getByte(f, "upgrade_type"), Result: getByte(f, "result"),
			UpgradeMsg: getString(f, "upgrade_msg"),
		}, nil

	// ==================== 位置相关 ====================
	case jt808.MsgIDLocation:
		t := getString(f, "time")
		if t == "" {
			t = time.Now().Format("060102150405")
		}
		return &jt808.LocationMessage{
			AlarmFlag: getUint32(f, "alarm_flag"), StatusFlag: getUint32(f, "status_flag"),
			Latitude: getFloat64(f, "latitude"), Longitude: getFloat64(f, "longitude"),
			Altitude: uint16(getInt(f, "altitude")), Speed: uint16(getInt(f, "speed")),
			Direction: uint16(getInt(f, "direction")), Time: t,
		}, nil
	case jt808.MsgIDLocationQuery:
		return &jt808.LocationQueryMessage{}, nil
	case jt808.MsgIDLocationQueryResp:
		return &jt808.LocationQueryRespMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")),
			LocationMessage: jt808.LocationMessage{
				AlarmFlag: getUint32(f, "alarm_flag"), StatusFlag: getUint32(f, "status_flag"),
				Latitude: getFloat64(f, "latitude"), Longitude: getFloat64(f, "longitude"),
				Altitude: uint16(getInt(f, "altitude")), Speed: uint16(getInt(f, "speed")),
				Direction: uint16(getInt(f, "direction")), Time: time.Now().Format("060102150405"),
			},
		}, nil
	case jt808.MsgIDLocationBatch:
		loc := &jt808.LocationMessage{
			AlarmFlag: getUint32(f, "alarm_flag"), StatusFlag: getUint32(f, "status_flag"),
			Latitude: getFloat64(f, "latitude"), Longitude: getFloat64(f, "longitude"),
			Altitude: uint16(getInt(f, "altitude")), Speed: uint16(getInt(f, "speed")),
			Direction: uint16(getInt(f, "direction")), Time: time.Now().Format("060102150405"),
		}
		return &jt808.LocationBatchMessage{Type: 1, Items: []*jt808.LocationMessage{loc}}, nil
	case jt808.MsgIDTempLocationTrack:
		return &jt808.TempLocationTrackMessage{
			Interval: uint16(getInt(f, "interval")), Validity: uint16(getInt(f, "validity")),
		}, nil
	case jt808.MsgIDTempLocationTrackResp:
		return &jt808.TempLocationTrackRespMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")), Result: getByte(f, "result"),
		}, nil
	case jt808.MsgIDManualAlarmConfirm:
		return &jt808.ManualAlarmConfirmMessage{
			SeqNum: uint16(getInt(f, "seq_num")), AlarmID: uint16(getInt(f, "alarm_id")),
		}, nil

	// ==================== 报警 ====================
	case jt808.MsgIDAlarm:
		return &jt808.AlarmMessage{
			VehicleColor: getByte(f, "vehicle_color"), VehiclePlate: getString(f, "vehicle_plate"),
			AlarmFlag: uint16(getInt(f, "alarm_flag")), WaterLevel: getByte(f, "water_level"),
		}, nil
	case jt808.MsgIDAlarmAttachment:
		return &jt808.AlarmAttachmentMessage{
			VehicleColor: getByte(f, "vehicle_color"), VehiclePlate: getString(f, "vehicle_plate"),
			AlarmID: uint16(getInt(f, "alarm_id")),
			AttachmentTime: time.Now().Format("060102150405"),
			AttachmentLen: uint16(getInt(f, "attachment_len")),
			AttachmentData: []byte{},
		}, nil
	case jt808.MsgIDAlarmAttachmentResp:
		return &jt808.AlarmAttachmentRespMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")), Result: getByte(f, "result"),
		}, nil
	case jt808.MsgIDOverspeedAlarm:
		return &jt808.OverspeedAlarmMessage{
			AlarmFlag: uint16(getInt(f, "alarm_flag")),
		}, nil
	case jt808.MsgIDFatigueDriveAlarm:
		return &jt808.FatigueDriveAlarmMessage{
			AlarmFlag: uint16(getInt(f, "alarm_flag")),
		}, nil

	// ==================== 文本与事件 ====================
	case jt808.MsgIDTextSend:
		return &jt808.TextSendMessage{Flag: getByte(f, "flag"), Content: getString(f, "content")}, nil
	case jt808.MsgIDEventSet:
		return &jt808.EventSetMessage{
			Items: []jt808.EventItem{{EventID: getByte(f, "event_id"), Content: getString(f, "content")}},
		}, nil
	case jt808.MsgIDEventResp:
		return &jt808.EventRespMessage{
			SeqNum: uint16(getInt(f, "seq_num")), EventID: getByte(f, "event_id"),
		}, nil
	case jt808.MsgIDQuestionResp:
		return &jt808.QuestionRespMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")), AnswerID: getByte(f, "answer_id"),
		}, nil

	// ==================== 报警设置 ====================
	case jt808.MsgIDOverspeedSet:
		return &jt808.OverspeedSetMessage{
			SeqNum: uint16(getInt(f, "seq_num")), Speed: uint16(getInt(f, "speed")),
			Duration: uint16(getInt(f, "duration")),
		}, nil
	case jt808.MsgIDFatigueDriveSet:
		return &jt808.FatigueDriveSetMessage{
			SeqNum: uint16(getInt(f, "seq_num")), DayMaxDrive: uint16(getInt(f, "day_max_drive")),
			DayMinRest: uint16(getInt(f, "day_min_rest")), MaxDrive: uint16(getInt(f, "max_drive")),
			MinRest: uint16(getInt(f, "min_rest")),
		}, nil

	// ==================== 车辆控制 ====================
	case jt808.MsgIDVehicleControl:
		return &jt808.VehicleControlMessage{ControlFlag: getByte(f, "control_flag")}, nil

	// ==================== 区域设置 ====================
	case jt808.MsgIDCircularAreaSet:
		return &jt808.CircularAreaSetMessage{
			AreaID: uint32(getInt(f, "area_id")), Attr: uint16(getInt(f, "attr")),
			CenterLat: getFloat64(f, "latitude"), CenterLon: getFloat64(f, "longitude"),
			Radius: uint32(getInt(f, "radius")),
			StartTime: getString(f, "start_time"), EndTime: getString(f, "end_time"),
		}, nil
	case jt808.MsgIDRectAreaSet:
		return &jt808.RectAreaSetMessage{
			AreaID: uint32(getInt(f, "area_id")), Attr: uint16(getInt(f, "attr")),
			UpperLat: getFloat64(f, "upper_lat"), UpperLon: getFloat64(f, "upper_lon"),
			LowerLat: getFloat64(f, "lower_lat"), LowerLon: getFloat64(f, "lower_lon"),
			StartTime: getString(f, "start_time"), EndTime: getString(f, "end_time"),
		}, nil
	case jt808.MsgIDPolygonAreaSet:
		return &jt808.PolygonAreaSetMessage{
			AreaID: uint32(getInt(f, "area_id")), Attr: uint16(getInt(f, "attr")),
			Points: []jt808.PolygonPoint{{Lat: getFloat64(f, "lat1"), Lon: getFloat64(f, "lon1")}},
			StartTime: getString(f, "start_time"), EndTime: getString(f, "end_time"),
		}, nil
	case jt808.MsgIDRouteSet:
		return &jt808.RouteSetMessage{
			RouteID: uint32(getInt(f, "route_id")), Attr: uint16(getInt(f, "attr")),
			StartTime: getString(f, "start_time"), EndTime: getString(f, "end_time"),
			Points: []jt808.RoutePoint{{
				PointID: uint32(getInt(f, "point_id")),
				Lat: getFloat64(f, "latitude"), Lon: getFloat64(f, "longitude"),
				Width: getByte(f, "width"), Attr: getByte(f, "point_attr"),
				MaxSpeed: uint16(getInt(f, "max_speed")), MaxDuration: uint16(getInt(f, "max_duration")),
			}},
		}, nil

	// ==================== 信息与通讯 ====================
	case jt808.MsgIDDriverID:
		return &jt808.DriverIDMessage{
			DriverName: getString(f, "driver_name"), DriverID: getString(f, "driver_id"),
			Licence: getString(f, "licence"), CertifyOrg: getString(f, "certify_org"),
		}, nil
	case jt808.MsgIDInfoMenuResp:
		return &jt808.InfoMenuRespMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")), MenuType: getByte(f, "menu_type"),
		}, nil
	case jt808.MsgIDSMSForwardResp:
		return &jt808.SMSForwardRespMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")), Result: getByte(f, "result"),
		}, nil
	case jt808.MsgIDElectronicWaybill:
		return &jt808.ElectronicWaybillMessage{
			WaybillData: []byte(getString(f, "waybill_data")),
		}, nil

	// ==================== 多媒体 ====================
	case jt808.MsgIDMultimedia:
		return &jt808.MultimediaMessage{
			MultimediaID: uint32(getInt(f, "multimedia_id")), MultimediaType: getByte(f, "multimedia_type"),
			Format: getByte(f, "format"), EventCode: getByte(f, "event_code"),
			ChannelID: getByte(f, "channel_id"),
		}, nil
	case jt808.MsgIDMultimediaUpload:
		return &jt808.MultimediaUploadMessage{
			MultimediaID: uint32(getInt(f, "multimedia_id")),
			MultimediaType: getByte(f, "multimedia_type"), Format: getByte(f, "format"),
			PlayTime: uint16(getInt(f, "play_time")), PackageSize: getByte(f, "package_size"),
			TotalPackages: uint16(getInt(f, "total_packages")), Offset: uint16(getInt(f, "offset")),
			Data: []byte{},
		}, nil
	case jt808.MsgIDStorageMediaSearch:
		return &jt808.StorageMediaSearchMessage{
			SeqNum: uint16(getInt(f, "seq_num")), LogicalCh: getByte(f, "logical_ch"),
			StartTime: getString(f, "start_time"), EndTime: getString(f, "end_time"),
			AlarmFlag: getByte(f, "alarm_flag"), MediaType: getByte(f, "media_type"),
			StreamType: getByte(f, "stream_type"),
		}, nil
	case jt808.MsgIDStorageMediaUpload:
		return &jt808.StorageMediaUploadMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")), Result: getByte(f, "result"),
			MultimediaID: uint32(getInt(f, "multimedia_id")),
			PackageSize: getByte(f, "package_size"), Offset: uint16(getInt(f, "offset")),
		}, nil
	case jt808.MsgIDPhotoCommand:
		return &jt808.PhotoCommandMessage{
			SeqNum: uint16(getInt(f, "seq_num")), ChannelID: getByte(f, "channel_id"),
			Command: getByte(f, "command"), Interval: uint16(getInt(f, "interval")),
			Count: getByte(f, "count"), Resolution: getByte(f, "resolution"),
		}, nil
	case jt808.MsgIDPhotoCommandResp:
		return &jt808.PhotoCommandRespMessage{
			RespSeqNum: uint16(getInt(f, "resp_seq_num")), Result: getByte(f, "result"),
			MultimediaID: uint32(getInt(f, "multimedia_id")),
			ChannelID: getByte(f, "channel_id"), PackageSize: uint16(getInt(f, "package_size")),
		}, nil

	// ==================== CAN数据 ====================
	case jt808.MsgIDCanData:
		return &jt808.CanDataMessage{
			CanItems: []jt808.CanItem{{
				CanTime: uint32(getInt(f, "can_time")), CanID: uint32(getInt(f, "can_id")),
				CanData: []byte{},
			}},
		}, nil

	// ==================== RSA与计价器 ====================
	case jt808.MsgIDRSAPublicKey:
		return &jt808.RSAPublicKeyMessage{
			EModule: []byte(getString(f, "e_module")), Exponent: []byte(getString(f, "exponent")),
		}, nil
	case jt808.MsgIDRSADistribute:
		return &jt808.RSADistributeMessage{
			EModule: []byte(getString(f, "e_module")), Exponent: []byte(getString(f, "exponent")),
		}, nil
	case jt808.MsgIDBillOperate:
		return &jt808.BillOperateMessage{
			OperateType: getByte(f, "operate_type"),
			OperateData: []byte(getString(f, "operate_data")),
		}, nil

	default:
		return &jt808.RawMessage{ID: msgID}, nil
	}
}

func (w *Workshop) encodeJT809(msgID uint16, req *EncodeRequest) *EncodeResponse {
	lat := getFloat64(req.Fields, "latitude")
	lon := getFloat64(req.Fields, "longitude")
	sp := uint16(getInt(req.Fields, "speed"))
	dir := uint16(getInt(req.Fields, "direction"))
	vc := getByte(req.Fields, "vehicle_color")
	vp := getString(req.Fields, "vehicle_plate")
	now := time.Now().Format("060102150405")

	var bodyBytes []byte
	switch msgID {
	case jt809.MsgIDLocationMsg:
		loc := &jt809.LocationMessage{
			VehicleColor: vc, VehiclePlate: vp, Lat: lat, Lon: lon,
			Speed: sp, Direction: dir, Time: now,
		}
		bodyBytes, _ = loc.Marshal()
	case jt809.MsgIDUpConnectReq:
		body := &jt809.ConnectReqMessage{
			UserName: uint32(getInt(req.Fields, "user_name")), Password: getString(req.Fields, "password"),
			DownLinkIP: getString(req.Fields, "downlink_ip"), DownLinkPort: uint32(getInt(req.Fields, "downlink_port")),
		}
		bodyBytes, _ = body.Marshal()
	case jt809.MsgIDUpDisconnectReq:
		body := &jt809.DisconnectReqMessage{
			UserName: uint32(getInt(req.Fields, "user_name")), Password: getString(req.Fields, "password"),
		}
		bodyBytes, _ = body.Marshal()
	case jt809.MsgIDUpLinktestReq, jt809.MsgIDUpLinktestRsp,
		jt809.MsgIDDnDisconnectReq, jt809.MsgIDDnDisconnectRsp,
		jt809.MsgIDDnLinktestReq, jt809.MsgIDDnLinktestRsp,
		jt809.MsgIDUpDisconnectRsp:
		bodyBytes = nil
	case jt809.MsgIDStartupMsg:
		body := &jt809.StartupMessage{
			VehicleColor: vc, VehiclePlate: vp, StartupTime: now,
			Lat: lat, Lon: lon, Speed: sp, Direction: dir,
		}
		bodyBytes, _ = body.Marshal()
	case jt809.MsgIDShutdownMsg:
		body := &jt809.ShutdownMessage{
			VehicleColor: vc, VehiclePlate: vp, ShutdownTime: now,
			Lat: lat, Lon: lon, Speed: sp, Direction: dir,
		}
		bodyBytes, _ = body.Marshal()
	case jt809.MsgIDUploadCarMsg:
		body := &jt809.VehicleInfoMessage{
			VehicleColor: vc, VehiclePlate: vp,
			VIN: getString(req.Fields, "vin"), VehicleType: uint16(getInt(req.Fields, "vehicle_type")),
		}
		bodyBytes, _ = body.Marshal()
	case jt809.MsgIDAlarmMsg:
		body := &jt809.AlarmMessage{
			VehicleColor: vc, VehiclePlate: vp,
			AlarmSource: getByte(req.Fields, "alarm_source"),
			AlarmType:   uint16(getInt(req.Fields, "alarm_type")),
			AlarmTime:   now, Lat: lat, Lon: lon, Speed: sp, Direction: dir,
		}
		bodyBytes, _ = body.Marshal()
	case jt809.MsgIDUpConnectRsp:
		body := &jt809.ConnectRspMessage{
			Result: getByte(req.Fields, "result"), VerifyCode: uint32(getInt(req.Fields, "verify_code")),
			DownLinkIP: getString(req.Fields, "downlink_ip"), DownLinkPort: uint32(getInt(req.Fields, "downlink_port")),
		}
		bodyBytes, _ = body.Marshal()
	case jt809.MsgIDDriverInfoMsg:
		body := &jt809.DriverInfoUpMessage{
			VehicleColor: vc, VehiclePlate: vp,
			DriverName: getString(req.Fields, "driver_name"), DriverID: getString(req.Fields, "driver_id"),
			Licence: getString(req.Fields, "licence"), OrgName: getString(req.Fields, "org_name"),
			UploadTime: now,
		}
		bodyBytes, _ = body.Marshal()
	case jt809.MsgIDEWaybillUpMsg:
		body := &jt809.EWaybillUpMessage{
			VehicleColor: vc, VehiclePlate: vp,
			WaybillID: getString(req.Fields, "waybill_id"), HazardClass: getByte(req.Fields, "hazard_class"),
			HazardName: getString(req.Fields, "hazard_name"), Weight: uint32(getInt(req.Fields, "weight")),
			Origin: getString(req.Fields, "origin"), Destination: getString(req.Fields, "destination"),
			LoadTime: now, UnloadTime: now,
		}
		bodyBytes, _ = body.Marshal()
	case jt809.MsgIDStatisticsMsg:
		body := &jt809.StatisticsMessage{
			VehicleColor: vc, VehiclePlate: vp,
			TotalMsgCount: uint32(getInt(req.Fields, "total_msg")),
			LocationCount: uint32(getInt(req.Fields, "loc_count")),
			AlarmCount:    uint32(getInt(req.Fields, "alarm_count")),
			OnlineDuration: uint32(getInt(req.Fields, "online_duration")),
		}
		bodyBytes, _ = body.Marshal()
	default:
		bodyBytes = nil
	}

	h := &jt809.Header{MsgSN: 1, MsgID: msgID, GNSSCenterID: 1001, VersionFlag: 1}
	data, err := w.jt809Codec.Encode(h, bodyBytes)
	if err != nil {
		return &EncodeResponse{Success: false, Error: err.Error()}
	}
	return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
}

func (w *Workshop) encodeJT1078(msgID uint16, req *EncodeRequest) *EncodeResponse {
	header := &types.MessageHeader{
		MsgID: msgID, Phone: req.Phone, SeqNum: 1,
		Version2019: true, ProtocolVer: 1,
	}
	lc := getByte(req.Fields, "logical_ch")
	dt := getByte(req.Fields, "data_type")
	st := getByte(req.Fields, "stream_type")

	var body types.MessageBody
	switch msgID {
	case jt1078.MsgIDRealtimeAVReq:
		body = &jt1078.RealtimeAVReqMessage{SeqNum: 1, LogicalCh: lc, DataType: dt, StreamType: st}
	case jt1078.MsgIDRealtimeAVCtrl:
		body = &jt1078.RealtimeAVCtrlMessage{LogicalCh: lc, CtrlCmd: dt, SwitchType: st}
	case jt1078.MsgIDPTZControl:
		body = &jt1078.PTZControlMessage{LogicalCh: lc, Direction: dt, Speed: st}
	case jt1078.MsgIDPlaybackReq:
		body = &jt1078.PlaybackReqMessage{
			SeqNum: 1, LogicalCh: lc,
			StartTime: getString(req.Fields, "start_time"), EndTime: getString(req.Fields, "end_time"),
			DataType: dt, StreamType: st,
		}
	case jt1078.MsgIDPlaybackCtrl:
		body = &jt1078.PlaybackCtrlMessage{
			LogicalCh: lc, CtrlCmd: dt, PlaySpeed: st,
			PlayTime: getString(req.Fields, "play_time"),
		}
	case jt1078.MsgIDPlaybackResp:
		body = &jt1078.PlaybackRespMessage{
			RespSeqNum: 1, LogicalCh: lc, Result: dt,
			StartTime: getString(req.Fields, "start_time"), EndTime: getString(req.Fields, "end_time"),
		}
	case jt1078.MsgIDTermAVReq:
		body = &jt1078.TermAVReqMessage{LogicalCh: lc, DataType: dt, StreamType: st}
	case jt1078.MsgIDTermAVResp:
		body = &jt1078.TermAVRespMessage{RespSeqNum: 1, LogicalCh: lc, Result: dt}
	case jt1078.MsgIDAVParamSet:
		body = &jt1078.AVParamSetMessage{
			LogicalCh: lc, AudioFormat: dt, VideoFormat: st,
			Resolution: getByte(req.Fields, "resolution"),
			FrameRate:  getByte(req.Fields, "frame_rate"),
			Bitrate:    uint32(getInt(req.Fields, "bitrate")),
		}
	case jt1078.MsgIDAVParamResp:
		body = &jt1078.AVParamMessage{SeqNum: 1, LogicalCh: lc}
	default:
		body = &jt808.RawMessage{ID: msgID}
	}

	data, err := w.jt808Codec.Encode(header, body)
	if err != nil {
		return &EncodeResponse{Success: false, Error: err.Error()}
	}
	return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
}

func (w *Workshop) encodeGBT32960(msgID uint16, req *EncodeRequest) *EncodeResponse {
	vin := getString(req.Fields, "vin")
	if vin == "" {
		vin = req.Phone
	}
	lat := getFloat64(req.Fields, "latitude")
	lon := getFloat64(req.Fields, "longitude")
	sp := uint16(getInt(req.Fields, "speed"))
	now := time.Now().Format("060102150405")

	switch byte(msgID) {
	case gbt32960.MsgTypeVehicleLogin:
		iccid := getString(req.Fields, "iccid")
		if iccid == "" {
			iccid = "12345678901234567890"
		}
		data := gbt32960.EncodeVehicleLogin(vin, now, iccid, 1)
		return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
	case gbt32960.MsgTypeHeartbeat:
		data := gbt32960.EncodeHeartbeat(vin)
		return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
	case gbt32960.MsgTypeVehicleLogout:
		data := gbt32960.EncodeVehicleLogout(vin, now, 1)
		return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
	case gbt32960.MsgTypeRealtimeInfo:
		vehData := &gbt32960.VehicleData{
			Status: 0x01, ChargeStatus: 0x01, Mode: 0x01, Speed: sp,
			TotalMileage: 100000, Voltage: 35000, Current: 5000, SOC: 80,
			DCDCStatus: 0x01, Shift: 0x02, Resistance: 500,
		}
		posData := &gbt32960.PositionData{
			ChargeStatus: 0x01, Lat: int32(lat * 1000000), Lon: int32(lon * 1000000),
			Altitude: 500, Direction: uint16(getInt(req.Fields, "direction")), Speed: sp,
		}
		items := []gbt32960.RealtimeDataItem{
			{Type: gbt32960.DataItemVehicle, Data: gbt32960.EncodeVehicleData(vehData)},
			{Type: gbt32960.DataItemPosition, Data: gbt32960.EncodePositionData(posData)},
		}
		data := gbt32960.EncodeRealtimeInfo(vin, now, 1, items)
		return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
	case gbt32960.MsgTypePlatformLogin:
		data := gbt32960.EncodePlatformLogin(
			getString(req.Fields, "platform_id"), now,
			getString(req.Fields, "password"), getString(req.Fields, "encrypt_key"),
		)
		return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
	case gbt32960.MsgTypeCheckResp:
		data := gbt32960.EncodeCheckResp(vin, byte(getInt(req.Fields, "cmd_flag")), now, 1, getByte(req.Fields, "result"))
		return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
	default:
		h := &gbt32960.Header{StartChar: 0x23, CmdFlag: byte(msgID), RespFlag: 0xFE, VIN: vin, EncryptType: 0x01}
		data := gbt32960.Encode(h, nil)
		return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
	}
}

func (w *Workshop) encodeJT905(msgID uint16, req *EncodeRequest) *EncodeResponse {
	header := &types.MessageHeader{MsgID: msgID, Phone: req.Phone, SeqNum: 1, Version2019: true, ProtocolVer: 1}
	lat := getFloat64(req.Fields, "latitude")
	lon := getFloat64(req.Fields, "longitude")
	vc := getByte(req.Fields, "vehicle_color")
	vp := getString(req.Fields, "vehicle_plate")
	now := time.Now().Format("060102150405")

	var body types.MessageBody
	switch msgID {
	case jt905.MsgIDTaxiOperate:
		body = &jt905.TaxiOperateMessage{
			VehicleColor: vc, VehiclePlate: vp, OperateType: getByte(req.Fields, "operate_type"),
			GetOnTime: now, GetOffTime: now,
			GetOnLat: lat, GetOnLon: lon, GetOffLat: lat, GetOffLon: lon,
			Distance: uint32(getInt(req.Fields, "distance")), Amount: uint32(getInt(req.Fields, "amount")),
		}
	case jt905.MsgIDTaxiStatus:
		body = &jt905.TaxiStatusMessage{
			VehicleColor: vc, VehiclePlate: vp, Status: getByte(req.Fields, "status"),
			ChangeTime: now, Lat: lat, Lon: lon,
		}
	case jt905.MsgIDTaxiDispatch:
		body = &jt905.TaxiDispatchMessage{
			DispatchType: getByte(req.Fields, "dispatch_type"), Content: getString(req.Fields, "content"),
			PassengerPhone: getString(req.Fields, "passenger_phone"), PickupAddr: getString(req.Fields, "pickup_addr"),
		}
	case jt905.MsgIDTaxiPrice:
		body = &jt905.TaxiPriceMessage{
			StartTime: now, UnitPrice: uint16(getInt(req.Fields, "unit_price")),
			UnitDistance: uint16(getInt(req.Fields, "unit_distance")),
			PerKmPrice: uint16(getInt(req.Fields, "per_km_price")),
			NightSurcharge: uint16(getInt(req.Fields, "night_surcharge")),
		}
	case jt905.MsgIDTaxiAd:
		body = &jt905.TaxiAdMessage{
			AdType: getByte(req.Fields, "ad_type"), PlayTimes: getByte(req.Fields, "play_times"),
			Content: getString(req.Fields, "content"),
		}
	default:
		body = &jt808.RawMessage{ID: msgID}
	}
	data, err := w.jt808Codec.Encode(header, body)
	if err != nil {
		return &EncodeResponse{Success: false, Error: err.Error()}
	}
	return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
}

func (w *Workshop) encodeJT1045(msgID uint16, req *EncodeRequest) *EncodeResponse {
	header := &types.MessageHeader{MsgID: msgID, Phone: req.Phone, SeqNum: 1, Version2019: true, ProtocolVer: 1}
	lat := getFloat64(req.Fields, "latitude")
	lon := getFloat64(req.Fields, "longitude")
	vc := getByte(req.Fields, "vehicle_color")
	vp := getString(req.Fields, "vehicle_plate")
	now := time.Now().Format("060102150405")

	var body types.MessageBody
	switch msgID {
	case jt1045.MsgIDADASAlarm:
		body = &jt1045.ADASAlarmMessage{
			VehicleColor: vc, VehiclePlate: vp, AlarmTime: now,
			AlarmType: uint16(getInt(req.Fields, "alarm_type")), AlarmLevel: getByte(req.Fields, "alarm_level"),
			AlarmParam: uint16(getInt(req.Fields, "alarm_param")),
			Lat: lat, Lon: lon, Altitude: uint16(getInt(req.Fields, "altitude")),
			Speed: uint16(getInt(req.Fields, "speed")), Direction: uint16(getInt(req.Fields, "direction")),
		}
	case jt1045.MsgIDDSMAlarm:
		body = &jt1045.DSMAlarmMessage{
			VehicleColor: vc, VehiclePlate: vp, AlarmTime: now,
			AlarmType: uint16(getInt(req.Fields, "alarm_type")), AlarmLevel: getByte(req.Fields, "alarm_level"),
			AlarmParam: uint16(getInt(req.Fields, "alarm_param")),
			Lat: lat, Lon: lon,
			Speed: uint16(getInt(req.Fields, "speed")), Direction: uint16(getInt(req.Fields, "direction")),
		}
	case jt1045.MsgIDTireAlarm:
		body = &jt1045.TireAlarmMessage{
			VehicleColor: vc, VehiclePlate: vp, AlarmTime: now,
			TirePosition: getByte(req.Fields, "tire_position"),
			Pressure: uint16(getInt(req.Fields, "pressure")),
			Temp: int16(getInt(req.Fields, "temp")),
			Battery: getByte(req.Fields, "battery"), AlarmFlag: getByte(req.Fields, "alarm_flag"),
		}
	case jt1045.MsgIDADASData:
		body = &jt1045.ADASDataMessage{
			VehicleColor: vc, VehiclePlate: vp, AlarmTime: now,
			AlarmType: uint16(getInt(req.Fields, "alarm_type")), AlarmLevel: getByte(req.Fields, "alarm_level"),
			Lat: lat, Lon: lon,
			Speed: uint16(getInt(req.Fields, "speed")), Direction: uint16(getInt(req.Fields, "direction")),
			VehicleNo: getString(req.Fields, "vehicle_no"),
			VehicleSpeed: uint16(getInt(req.Fields, "vehicle_speed")),
			Distance: uint16(getInt(req.Fields, "distance")),
		}
	default:
		body = &jt808.RawMessage{ID: msgID}
	}
	data, err := w.jt808Codec.Encode(header, body)
	if err != nil {
		return &EncodeResponse{Success: false, Error: err.Error()}
	}
	return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
}

func (w *Workshop) encodeJT1253(msgID uint16, req *EncodeRequest) *EncodeResponse {
	header := &types.MessageHeader{MsgID: msgID, Phone: req.Phone, SeqNum: 1, Version2019: true, ProtocolVer: 1}
	lat := getFloat64(req.Fields, "latitude")
	lon := getFloat64(req.Fields, "longitude")
	vc := getByte(req.Fields, "vehicle_color")
	vp := getString(req.Fields, "vehicle_plate")
	now := time.Now().Format("060102150405")

	var body types.MessageBody
	switch msgID {
	case jt1253.MsgIDEWaybill:
		body = &jt1253.EWaybillMessage{
			VehicleColor: vc, VehiclePlate: vp,
			WaybillID: getString(req.Fields, "waybill_id"), HazardClass: getByte(req.Fields, "hazard_class"),
			HazardName: getString(req.Fields, "hazard_name"), Weight: uint32(getInt(req.Fields, "weight")),
			Origin: getString(req.Fields, "origin"), Destination: getString(req.Fields, "destination"),
			LoadTime: now, UnloadTime: now,
		}
	case jt1253.MsgIDHazardAlarm:
		body = &jt1253.HazardAlarmMessage{
			VehicleColor: vc, VehiclePlate: vp,
			AlarmType: uint16(getInt(req.Fields, "alarm_type")), AlarmLevel: getByte(req.Fields, "alarm_level"),
			AlarmTime: now, Lat: lat, Lon: lon,
			Speed: uint16(getInt(req.Fields, "speed")), Direction: uint16(getInt(req.Fields, "direction")),
		}
	case jt1253.MsgIDHazardStatus:
		body = &jt1253.HazardStatusMessage{
			VehicleColor: vc, VehiclePlate: vp, StatusTime: now,
			HazardClass: getByte(req.Fields, "hazard_class"), HazardState: getByte(req.Fields, "hazard_state"),
			Lat: lat, Lon: lon,
			Speed: uint16(getInt(req.Fields, "speed")), Direction: uint16(getInt(req.Fields, "direction")),
			Temp: int16(getInt(req.Fields, "temp")), Humidity: uint16(getInt(req.Fields, "humidity")),
			Pressure: uint32(getInt(req.Fields, "pressure")), Level: uint16(getInt(req.Fields, "level")),
		}
	case jt1253.MsgIDUnloadReport:
		body = &jt1253.UnloadReportMessage{
			VehicleColor: vc, VehiclePlate: vp, UnloadTime: now,
			HazardClass: getByte(req.Fields, "hazard_class"),
			UnloadAmount: uint32(getInt(req.Fields, "unload_amount")),
			UnloadPlace: getString(req.Fields, "unload_place"),
			Lat: lat, Lon: lon,
		}
	default:
		body = &jt808.RawMessage{ID: msgID}
	}
	data, err := w.jt808Codec.Encode(header, body)
	if err != nil {
		return &EncodeResponse{Success: false, Error: err.Error()}
	}
	return &EncodeResponse{Hex: strings.ToUpper(hex.EncodeToString(data)), Success: true}
}

// AnalyzeRequest 分析请求
type AnalyzeRequest struct {
	Protocol string `json:"protocol"`
	Hex      string `json:"hex"`
}

// AnalyzeResponse 分析响应
type AnalyzeResponse struct {
	Success bool          `json:"success"`
	Error   string        `json:"error,omitempty"`
	MsgID   string        `json:"msg_id"`
	MsgName string        `json:"msg_name"`
	Phone   string        `json:"phone"`
	SeqNum  uint16        `json:"seq_num"`
	Parsed  interface{}   `json:"parsed"`
	Issues  []types.Issue `json:"issues,omitempty"`
}

// Analyze 分析报文
func (w *Workshop) Analyze(req *AnalyzeRequest) *AnalyzeResponse {
	hexStr := strings.TrimSpace(req.Hex)
	hexStr = strings.ReplaceAll(hexStr, " ", "")
	hexStr = strings.ReplaceAll(hexStr, "0x", "")
	hexStr = strings.ReplaceAll(hexStr, "0X", "")

	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return &AnalyzeResponse{Success: false, Error: fmt.Sprintf("invalid hex: %v", err)}
	}

	switch req.Protocol {
	case "jt808", "jt1078", "jt905", "jt1045", "jt1253":
		return w.analyzeJT808(data, req.Protocol)
	case "jt809":
		return w.analyzeJT809(data)
	case "gbt32960":
		return w.analyzeGBT32960(data)
	default:
		return w.analyzeJT808(data, "jt808")
	}
}

func (w *Workshop) analyzeJT808(data []byte, protocol string) *AnalyzeResponse {
	if len(data) < 2 || data[0] != 0x7E || data[len(data)-1] != 0x7E {
		return &AnalyzeResponse{Success: false, Error: "missing 0x7E delimiter"}
	}
	msg, err := w.jt808Codec.Decode(data)
	if err != nil {
		return &AnalyzeResponse{Success: false, Error: err.Error()}
	}

	// Determine the correct message name based on protocol
	msgName := jt808.MsgName(msg.Header.MsgID)
	switch protocol {
	case "jt1078":
		msgName = jt1078.MsgName(msg.Header.MsgID)
	case "jt905":
		msgName = jt905.MsgName(msg.Header.MsgID)
	case "jt1045":
		msgName = jt1045.MsgName(msg.Header.MsgID)
	case "jt1253":
		msgName = jt1253.MsgName(msg.Header.MsgID)
	}

	resp := &AnalyzeResponse{
		Success: true,
		MsgID:   fmt.Sprintf("0x%04X", msg.Header.MsgID),
		MsgName: msgName,
		Phone:   msg.Header.Phone,
		SeqNum:  msg.Header.SeqNum,
		Parsed:  msg.Body,
	}
	resp.Issues = w.detectIssues(msg)
	return resp
}

func (w *Workshop) analyzeJT809(data []byte) *AnalyzeResponse {
	if len(data) < 2 || data[0] != 0x5B || data[len(data)-1] != 0x5E {
		return &AnalyzeResponse{Success: false, Error: "missing 0x5B/0x5E delimiter"}
	}
	h, bodyData, err := w.jt809Codec.Decode(data)
	if err != nil {
		return &AnalyzeResponse{Success: false, Error: err.Error()}
	}
	var parsed interface{}
	parsed = fmt.Sprintf("body(%d bytes): %X", len(bodyData), bodyData)
	switch h.MsgID {
	case jt809.MsgIDLocationMsg:
		loc := &jt809.LocationMessage{}
		if err := loc.Unmarshal(bodyData); err == nil {
			parsed = loc
		}
	case jt809.MsgIDAlarmMsg:
		alarm := &jt809.AlarmMessage{}
		if err := alarm.Unmarshal(bodyData); err == nil {
			parsed = alarm
		}
	case jt809.MsgIDUpConnectReq:
		conn := &jt809.ConnectReqMessage{}
		if err := conn.Unmarshal(bodyData); err == nil {
			parsed = conn
		}
	case jt809.MsgIDUpConnectRsp:
		rsp := &jt809.ConnectRspMessage{}
		if err := rsp.Unmarshal(bodyData); err == nil {
			parsed = rsp
		}
	}
	return &AnalyzeResponse{
		Success: true, MsgID: fmt.Sprintf("0x%04X", h.MsgID),
		MsgName: jt809.MsgName(h.MsgID),
		Phone:   fmt.Sprintf("SN=%d", h.MsgSN),
		SeqNum:  uint16(h.MsgSN), Parsed: parsed,
	}
}

func (w *Workshop) analyzeGBT32960(data []byte) *AnalyzeResponse {
	pkt, err := gbt32960.Decode(data)
	if err != nil {
		return &AnalyzeResponse{Success: false, Error: err.Error()}
	}
	return &AnalyzeResponse{
		Success: true,
		MsgID:   fmt.Sprintf("0x%02X", pkt.Header.CmdFlag),
		MsgName: gbt32960.MsgTypeString(pkt.Header.CmdFlag),
		Phone:   pkt.Header.VIN, SeqNum: 0,
		Parsed: map[string]interface{}{
			"cmd_flag":     fmt.Sprintf("0x%02X", pkt.Header.CmdFlag),
			"resp_flag":    fmt.Sprintf("0x%02X", pkt.Header.RespFlag),
			"vin":          pkt.Header.VIN,
			"encrypt_type": pkt.Header.EncryptType,
			"body_len":     pkt.Header.BodyLen,
			"body_hex":     fmt.Sprintf("%X", pkt.Body),
		},
	}
}

func (w *Workshop) detectIssues(msg *types.Message) []types.Issue {
	var issues []types.Issue
	if msg.Header.MsgID == 0 {
		issues = append(issues, types.Issue{Level: "error", Field: "msg_id", Message: "消息ID为0，不合法"})
	}
	if msg.Header.Phone == "" || msg.Header.Phone == "000000000000" {
		issues = append(issues, types.Issue{Level: "warning", Field: "phone", Message: "终端手机号为空或全零"})
	}
	if loc, ok := msg.Body.(*jt808.LocationMessage); ok {
		if loc.Latitude < -90 || loc.Latitude > 90 {
			issues = append(issues, types.Issue{Level: "error", Field: "latitude", Message: fmt.Sprintf("纬度超出范围: %.6f", loc.Latitude)})
		}
		if loc.Longitude < -180 || loc.Longitude > 180 {
			issues = append(issues, types.Issue{Level: "error", Field: "longitude", Message: fmt.Sprintf("经度超出范围: %.6f", loc.Longitude)})
		}
		if loc.Speed > 3000 {
			issues = append(issues, types.Issue{Level: "warning", Field: "speed", Message: fmt.Sprintf("速度异常: %.1f km/h", float64(loc.Speed)/10)})
		}
	}
	return issues
}

func parseMsgID(s string) (uint16, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	var id uint16
	_, err := fmt.Sscanf(s, "%x", &id)
	if err != nil {
		return 0, fmt.Errorf("invalid msg_id: %s", s)
	}
	return id, nil
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

func getUint32(m map[string]interface{}, key string) uint32 { return uint32(getInt(m, key)) }
func getByte(m map[string]interface{}, key string) byte     { return byte(getInt(m, key)) }

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}