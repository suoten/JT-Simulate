package trajectory

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MaxPoints 单条轨迹点数上限
const MaxPoints = 100000

// Parse 解析轨迹文件内容，按内容自动识别 GPX 或文本格式。
// name 为轨迹名称（来源文件名或用户指定）。
func Parse(name string, data []byte) (*Route, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("文件内容为空")
	}
	head := string(data[:min(len(data), 512)])
	lower := strings.ToLower(head)
	if strings.Contains(lower, "<gpx") || strings.Contains(lower, "<trkpt") || strings.Contains(lower, "<?xml") {
		return parseGPX(name, data)
	}
	return parseText(name, data)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================== GPX ==============================

type gpxFile struct {
	Tracks []gpxTrack `xml:"trk"`
}

type gpxTrack struct {
	Segments []gpxSegment `xml:"trkseg"`
}

type gpxSegment struct {
	Points []gpxPoint `xml:"trkpt"`
}

type gpxPoint struct {
	Lat  float64 `xml:"lat,attr"`
	Lon  float64 `xml:"lon,attr"`
	Ele  float64 `xml:"ele"`
	Time string  `xml:"time"`
}

// gpxTimeLayouts GPX 时间戳兼容格式（标准为 RFC3339，部分设备缺时区或毫秒）
var gpxTimeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05.999",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
}

func parseGPX(name string, data []byte) (*Route, error) {
	var f gpxFile
	if err := xml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("GPX 解析失败: %w", err)
	}

	var raw []gpxPoint
	for _, trk := range f.Tracks {
		for _, seg := range trk.Segments {
			raw = append(raw, seg.Points...)
		}
	}
	if len(raw) < 2 {
		return nil, fmt.Errorf("GPX 中有效轨迹点少于 2 个（需 <trkpt> 位于 <trk><trkseg> 内）")
	}

	pts := make([]Point, 0, len(raw))
	times := make([]*time.Time, 0, len(raw))
	allTimed := true
	for _, g := range raw {
		pts = append(pts, Point{Lat: g.Lat, Lon: g.Lon})
		if g.Time == "" {
			times = append(times, nil)
			allTimed = false
			continue
		}
		var t time.Time
		var ok bool
		for _, layout := range gpxTimeLayouts {
			if t, ok = tryParseTime(layout, g.Time); ok {
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("无法解析 GPX 时间: %s", g.Time)
		}
		times = append(times, &t)
	}

	// 全部点均有时间才启用时间轴；首点时间作为基准
	if allTimed {
		base := *times[0]
		for i := range pts {
			pts[i].TimeSec = times[i].Sub(base).Seconds()
		}
	}

	r, err := NewRoute(name, "gpx", pts)
	if err != nil {
		return nil, fmt.Errorf("GPX 轨迹无效: %w", err)
	}
	return r, nil
}

func tryParseTime(layout, val string) (time.Time, bool) {
	t, err := time.Parse(layout, val)
	return t, err == nil
}

// ============================== 文本 ==============================

// fieldSplitter 逗号 / 分号 / Tab / 连续空格分隔
var fieldSplitter = regexp.MustCompile(`[\s,;]+`)

// timeHintRe 含冒号或日期分隔符的字段视为时间（而非速度/秒偏移）
var timeHintRe = regexp.MustCompile(`^\d{1,2}:\d{2}(:\d{2}(\.\d+)?)?$` + // HH:MM[:SS]
	`|^\d{4}-\d{1,2}-\d{1,2}([T ]\d{1,2}:\d{2}(:\d{2}(\.\d+)?)?Z?)?$`) // 日期[ 时间]（含 RFC3339）

// spaceDateRe 修复被空白拆开的日期时间（如 "2026-01-01 08:00:00" 分成两列）
var spaceDateRe = regexp.MustCompile(`(\d{4}-\d{1,2}-\d{1,2})[ ]+(\d{1,2}:\d{2}(:\d{2}(\.\d+)?)?)$`)

// bomPrefix UTF-8 BOM
var bomPrefix = string(rune(0xFEFF))

// parseText 解析文本轨迹：每行 `纬度,经度[,速度][,时间]`。
// 支持 # 注释、自动跳过表头；时间支持秒偏移量、HH:MM:SS、YYYY-MM-DD HH:MM:SS。
func parseText(name string, data []byte) (*Route, error) {
	text := strings.TrimPrefix(string(data), bomPrefix)
	lines := strings.Split(text, "\n")

	var pts []Point
	var secs []float64 // 每行原始时间值（秒偏移 / 当日秒数 / 绝对时间戳）
	var timed []bool   // 每行是否带时间

	lineNo := 0
	for _, line := range lines {
		lineNo++
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := fieldSplitter.Split(line, -1)
		if len(fields) < 2 {
			continue // 单列杂项行忽略
		}

		lat, err1 := strconv.ParseFloat(fields[0], 64)
		lon, err2 := strconv.ParseFloat(fields[1], 64)
		if err1 != nil || err2 != nil {
			if len(pts) == 0 {
				continue // 首个有效点之前的非数据行（如表头）忽略
			}
			return nil, fmt.Errorf("第 %d 行坐标无法解析: %q", lineNo, line)
		}

		p := Point{Lat: lat, Lon: lon}

		// 坐标后的剩余字段：先合并，修复被空格拆开的日期时间，再拆回至多两列（速度? 时间?）
		rest := spaceDateRe.ReplaceAllString(strings.TrimSpace(strings.Join(fields[2:], " ")), "$1T$2")
		tokens := fieldSplitter.Split(rest, -1)

		tGiven := false
		var tVal float64
		for _, f := range tokens {
			if f == "" {
				continue
			}
			switch {
			case timeHintRe.MatchString(f):
				sec, err := parseWallClock(f)
				if err != nil {
					return nil, fmt.Errorf("第 %d 行时间无法解析: %s", lineNo, f)
				}
				tGiven, tVal = true, sec
			default:
				v, err := strconv.ParseFloat(f, 64)
				if err != nil {
					return nil, fmt.Errorf("第 %d 行存在无法识别的字段 %q（格式：纬度,经度[,速度][,时间]）", lineNo, f)
				}
				if !tGiven && p.SpeedKmh == 0 {
					// 第一个纯数字字段解释为速度（km/h）
					p.SpeedKmh = v
				} else if !tGiven {
					// 第二个纯数字字段解释为时间秒偏移
					tGiven, tVal = true, v
				} else {
					return nil, fmt.Errorf("第 %d 行字段过多（格式：纬度,经度[,速度][,时间]）", lineNo)
				}
			}
		}

		pts = append(pts, p)
		secs = append(secs, tVal)
		timed = append(timed, tGiven)
	}

	if len(pts) < 2 {
		return nil, fmt.Errorf("文本轨迹有效点少于 2 个（每行格式：纬度,经度[,速度][,时间]）")
	}

	// 全部行均带时间才启用时间轴
	allTimed := true
	for _, t := range timed {
		if !t {
			allTimed = false
			break
		}
	}
	if allTimed {
		normalizeTextTimes(pts, secs)
	}

	r, err := NewRoute(name, "text", pts)
	if err != nil {
		return nil, fmt.Errorf("文本轨迹无效: %w", err)
	}
	return r, nil
}

// normalizeTextTimes 将每行的时间归一化为相对首点的秒数，写入 pts[i].TimeSec。
// HH:MM:SS（当日时刻，全部落在 [0,86400)）跨午夜回绕时补一天；绝对时间戳直接作差。
func normalizeTextTimes(pts []Point, secs []float64) {
	if isDailyClock(secs) {
		for i := 1; i < len(secs); i++ {
			if secs[i] < secs[i-1] {
				for j := i; j < len(secs); j++ {
					secs[j] += 86400
				}
			}
		}
	}
	base := secs[0]
	for i := range pts {
		pts[i].TimeSec = secs[i] - base
	}
}

// parseWallClock 解析 HH:MM:SS（返回当日秒数）或 YYYY-MM-DD[THH:MM:SS]（返回绝对时间戳秒）
func parseWallClock(s string) (float64, error) {
	if m := regexp.MustCompile(`^(\d{1,2}):(\d{2})(:\d{2}(\.\d+)?)?$`).FindStringSubmatch(s); m != nil {
		h, _ := strconv.Atoi(m[1])
		mi, _ := strconv.Atoi(m[2])
		sec := 0
		if m[3] != "" {
			sec, _ = strconv.Atoi(strings.TrimPrefix(m[3], ":"))
		}
		if h > 23 || mi > 59 || sec > 59 {
			return 0, fmt.Errorf("时间超出范围")
		}
		return float64(h*3600 + mi*60 + sec), nil
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return float64(t.Unix()), nil
		}
	}
	return 0, fmt.Errorf("无法识别的时间格式")
}

// isDailyClock 判断序列是否为当日时刻（HH:MM:SS，全部落在 [0,86400)）
func isDailyClock(secs []float64) bool {
	for _, s := range secs {
		if s < 0 || s >= 86400 {
			return false
		}
	}
	return true
}
