// Package trajectory 轨迹回放：轨迹文件的解析、存储与沿轨迹插值定位。
package trajectory

import (
	"errors"
	"fmt"
	"math"
)

// Point 轨迹点
type Point struct {
	Lat      float64 `json:"lat"`                 // 纬度（度）
	Lon      float64 `json:"lon"`                 // 经度（度）
	SpeedKmh float64 `json:"speed_kmh,omitempty"` // 文件自带速度（km/h），0 表示未提供
	TimeSec  float64 `json:"time_sec"`            // 相对首点的时间偏移（秒），仅 HasTime 时有效
}

// Route 轨迹（含预计算的累计距离，供插值定位使用）
type Route struct {
	Name       string  `json:"name"`
	Source     string  `json:"source"` // gpx | text
	HasTime    bool    `json:"has_time"`
	TotalDistM float64 `json:"total_distance_m"` // 总距离（米）
	DurationS  float64 `json:"duration_s"`       // 总时长（秒），仅 HasTime 时有效
	Points     []Point `json:"points"`

	cumDist []float64 // 各点累计距离（米），len == len(Points)，不参与序列化
}

// Pos 沿轨迹定位结果
type Pos struct {
	Lat     float64 // 纬度
	Lon     float64 // 经度
	Dir     uint16  // 方向（0-359，正北顺时针）
	SpeedKmh float64 // 轨迹推算/文件自带的速度；0 表示轨迹未提供速度（由回放引擎回退到配置速度）
	Progress float64 // 沿全程进度 0-1
}

// errShortRoute 轨迹点数不足
var errShortRoute = errors.New("轨迹至少需要 2 个点")

// NewRoute 由点序列构建轨迹，计算累计距离/总距离/时长。
// 校验坐标范围与点数，HasTime 要求所有点均有有效且严格递增的时间。
func NewRoute(name, source string, pts []Point) (*Route, error) {
	if len(pts) < 2 {
		return nil, errShortRoute
	}
	if len(pts) > MaxPoints {
		return nil, fmt.Errorf("轨迹点数超过上限 %d", MaxPoints)
	}
	for i, p := range pts {
		if p.Lat < -90 || p.Lat > 90 {
			return nil, fmt.Errorf("第 %d 个点纬度越界: %v", i+1, p.Lat)
		}
		if p.Lon < -180 || p.Lon > 180 {
			return nil, fmt.Errorf("第 %d 个点经度越界: %v", i+1, p.Lon)
		}
	}

	// 时间有效性：全部点均提供（首点 0 为基准）且整体严格递增才启用时间轴
	hasTime := true
	for i := range pts {
		if i > 0 && pts[i].TimeSec <= pts[i-1].TimeSec {
			hasTime = false
			break
		}
	}
	if hasTime && pts[len(pts)-1].TimeSec <= 0 {
		hasTime = false
	}
	if !hasTime {
		for i := range pts {
			pts[i].TimeSec = 0
		}
	}

	r := &Route{
		Name:    name,
		Source:  source,
		HasTime: hasTime,
		Points:  pts,
	}
	r.recompute()
	return r, nil
}

// recompute 重算累计距离、总距离与时长（NewRoute 与从存储加载后调用）
func (r *Route) recompute() {
	n := len(r.Points)
	r.cumDist = make([]float64, n)
	for i := 1; i < n; i++ {
		r.cumDist[i] = r.cumDist[i-1] + HaversineM(r.Points[i-1], r.Points[i])
	}
	r.TotalDistM = r.cumDist[n-1]
	if r.HasTime {
		r.DurationS = r.Points[n-1].TimeSec - r.Points[0].TimeSec
	}
}

// PositionAtDist 按累计距离定位（用于匀速推进模式）。
// d 越界时收敛到端点。速度取文件自带值（段起点），无则返回 0。
func (r *Route) PositionAtDist(d float64) Pos {
	return r.position(r.locateDist(d))
}

// PositionAtTime 按时间定位（用于按时间回放模式），t 为相对首点秒数。
// 速度由相邻点距离/时间差推算；时间差为 0 的段沿用前一段速度。
func (r *Route) PositionAtTime(t float64) Pos {
	i, frac := r.locateTime(t)
	p1, p2 := r.Points[i], r.Points[i+1]
	pos := r.position(i, frac)
	if dt := p2.TimeSec - p1.TimeSec; dt > 0 {
		pos.SpeedKmh = HaversineM(p1, p2) / dt * 3.6
	} else if i > 0 {
		// 零时长段沿用上一段推算速度，避免速度跳变归零
		prev := r.Points[i-1]
		if dtp := p1.TimeSec - prev.TimeSec; dtp > 0 {
			pos.SpeedKmh = HaversineM(prev, p1) / dtp * 3.6
		}
	}
	return pos
}

// locateDist 返回 (段索引, 段内距离比例)；终点返回最后一段比例 1。
func (r *Route) locateDist(d float64) (int, float64) {
	n := len(r.Points)
	last := n - 1
	if d <= 0 {
		return 0, 0
	}
	if d >= r.cumDist[last] {
		return lastSeg(r), 1
	}
	// 二分找 cumDist[i] <= d < cumDist[i+1]
	lo, hi := 0, last
	for lo+1 < hi {
		mid := (lo + hi) / 2
		if r.cumDist[mid] <= d {
			lo = mid
		} else {
			hi = mid
		}
	}
	segLen := r.cumDist[lo+1] - r.cumDist[lo]
	if segLen <= 0 {
		return lo, 0
	}
	return lo, (d - r.cumDist[lo]) / segLen
}

// locateTime 返回 (段索引, 段内时间比例)；终点返回最后一段比例 1。
func (r *Route) locateTime(t float64) (int, float64) {
	n := len(r.Points)
	last := n - 1
	if t <= r.Points[0].TimeSec {
		return 0, 0
	}
	if t >= r.Points[last].TimeSec {
		return lastSeg(r), 1
	}
	lo, hi := 0, last
	for lo+1 < hi {
		mid := (lo + hi) / 2
		if r.Points[mid].TimeSec <= t {
			lo = mid
		} else {
			hi = mid
		}
	}
	dt := r.Points[lo+1].TimeSec - r.Points[lo].TimeSec
	if dt <= 0 {
		return lo, 0
	}
	return lo, (t - r.Points[lo].TimeSec) / dt
}

// lastSeg 返回最后一个长度非零段；全程零长度（所有点重合）时返回段 0。
func lastSeg(r *Route) int {
	for i := len(r.Points) - 2; i >= 0; i-- {
		if r.cumDist[i+1] > r.cumDist[i] {
			return i
		}
	}
	return 0
}

// position 计算段 i 内 frac 比例处的位置
func (r *Route) position(i int, frac float64) Pos {
	p1, p2 := r.Points[i], r.Points[i+1]
	pos := Pos{
		Lat: p1.Lat + (p2.Lat-p1.Lat)*frac,
		Lon: p1.Lon + (p2.Lon-p1.Lon)*frac,
		Dir: Bearing(p1, p2),
	}
	if p1.SpeedKmh > 0 {
		pos.SpeedKmh = p1.SpeedKmh
	}
	pos.Progress = r.cumDist[i] / r.TotalDistM
	if r.TotalDistM > 0 {
		pos.Progress += (r.cumDist[i+1]-r.cumDist[i])/r.TotalDistM*frac
	}
	if pos.Progress > 1 {
		pos.Progress = 1
	}
	return pos
}

// earthRadiusM 地球平均半径（米）
const earthRadiusM = 6371000.0

// HaversineM 两点球面距离（米）
func HaversineM(a, b Point) float64 {
	lat1, lat2 := a.Lat*math.Pi/180, b.Lat*math.Pi/180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLon := (b.Lon - a.Lon) * math.Pi / 180
	sinDLat, sinDLon := math.Sin(dLat/2), math.Sin(dLon/2)
	h := sinDLat*sinDLat + math.Cos(lat1)*math.Cos(lat2)*sinDLon*sinDLon
	return 2 * earthRadiusM * math.Asin(math.Sqrt(h))
}

// Bearing 从 a 指向 b 的方位角（0-359，正北为 0，顺时针）
func Bearing(a, b Point) uint16 {
	lat1, lat2 := a.Lat*math.Pi/180, b.Lat*math.Pi/180
	dLon := (b.Lon - a.Lon) * math.Pi / 180
	y := math.Sin(dLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)
	if x == 0 && y == 0 {
		return 0
	}
	brng := math.Atan2(y, x) * 180 / math.Pi
	if brng < 0 {
		brng += 360
	}
	return uint16(math.Mod(brng, 360))
}
