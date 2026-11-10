package trajectory

import (
	"math"
	"testing"
)

func nearlyEq(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

// 北京 天安门(39.9087,116.3975) → 正北 1110m 处纬度约 +0.01 度
func TestHaversineM(t *testing.T) {
	a := Point{Lat: 39.9087, Lon: 116.3975}
	b := Point{Lat: 39.9187, Lon: 116.3975}
	d := HaversineM(a, b)
	if d < 1100 || d > 1125 {
		t.Fatalf("南北向 0.01 度距离应在 1100-1125 米之间，实际 %f", d)
	}
	if HaversineM(a, a) != 0 {
		t.Fatal("同一点距离应为 0")
	}
}

func TestBearing(t *testing.T) {
	// 注意：同纬度两点的正东/正西大圆初始方位角并非严格 90°/270°（赤道除外），
	// 因此东西向允许 ±1° 容差；南北向严格相等。
	cases := []struct {
		name   string
		a, b   Point
		want   uint16
		tolDeg uint16
	}{
		{"正北", Point{Lat: 39.9, Lon: 116.4}, Point{Lat: 39.92, Lon: 116.4}, 0, 0},
		{"正东", Point{Lat: 39.9, Lon: 116.4}, Point{Lat: 39.9, Lon: 116.45}, 90, 1},
		{"正南", Point{Lat: 39.92, Lon: 116.4}, Point{Lat: 39.9, Lon: 116.4}, 180, 0},
		{"正西", Point{Lat: 39.9, Lon: 116.45}, Point{Lat: 39.9, Lon: 116.4}, 270, 1},
		{"重合", Point{Lat: 39.9, Lon: 116.4}, Point{Lat: 39.9, Lon: 116.4}, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Bearing(c.a, c.b)
			if got == c.want {
				return
			}
			// 环绕差值：|got-want| 与 360-|got-want| 取小（用 int 避免 uint16 下溢）
			diff := int(got) - int(c.want)
			if diff < 0 {
				diff = -diff
			}
			if diff > 180 {
				diff = 360 - diff
			}
			if diff > int(c.tolDeg) {
				t.Fatalf("bearing = %d, want %d (±%d)", got, c.want, c.tolDeg)
			}
		})
	}
}

func straightRoute() *Route {
	// 南北向直线 0.02 度 ≈ 2220 米，带速度标记
	pts := []Point{
		{Lat: 39.90, Lon: 116.4, SpeedKmh: 40},
		{Lat: 39.91, Lon: 116.4, SpeedKmh: 60},
		{Lat: 39.92, Lon: 116.4},
	}
	r, err := NewRoute("直线", "text", pts)
	if err != nil {
		panic(err)
	}
	return r
}

func TestRouteBasics(t *testing.T) {
	r := straightRoute()
	if r.HasTime {
		t.Fatal("无时间轨迹 HasTime 应为 false")
	}
	if r.TotalDistM < 2200 || r.TotalDistM > 2230 {
		t.Fatalf("总距离异常: %f", r.TotalDistM)
	}
}

func TestPositionAtDist(t *testing.T) {
	r := straightRoute()

	// 起点
	p := r.PositionAtDist(0)
	if !nearlyEq(p.Lat, 39.90) || !nearlyEq(p.Lon, 116.4) {
		t.Fatalf("起点位置错误: %v, %v", p.Lat, p.Lon)
	}
	if p.Dir != 0 {
		t.Fatalf("向北方向应为 0，实际 %d", p.Dir)
	}
	if p.SpeedKmh != 40 {
		t.Fatalf("起点速度应取文件值 40，实际 %f", p.SpeedKmh)
	}

	// 中点（约在第二段中部，第二段有速度标记 60）
	p = r.PositionAtDist(r.TotalDistM / 2)
	if !nearlyEq(p.Lat, 39.91) {
		t.Fatalf("中点纬度应约 39.91，实际 %v", p.Lat)
	}
	if p.Dir != 0 {
		t.Fatalf("中点方向应为 0，实际 %d", p.Dir)
	}

	// 终点
	p = r.PositionAtDist(1e9)
	if !nearlyEq(p.Lat, 39.92) {
		t.Fatalf("越界应收敛到终点，实际 %v", p.Lat)
	}
	if p.Progress != 1 {
		t.Fatalf("终点进度应为 1，实际 %f", p.Progress)
	}
	if p.Dir != 0 {
		t.Fatalf("终点方向应沿用最后一段，实际 %d", p.Dir)
	}
}

func timedRoute() *Route {
	pts := []Point{
		{Lat: 39.90, Lon: 116.4, TimeSec: 0},
		{Lat: 39.91, Lon: 116.4, TimeSec: 60},  // 1110m/60s ≈ 66.6 km/h
		{Lat: 39.915, Lon: 116.4, TimeSec: 240}, // 555m/180s ≈ 11.1 km/h
	}
	r, err := NewRoute("时间线", "gpx", pts)
	if err != nil {
		panic(err)
	}
	return r
}

func TestPositionAtTime(t *testing.T) {
	r := timedRoute()
	if !r.HasTime {
		t.Fatal("时间轨迹 HasTime 应为 true")
	}
	if !nearlyEq(r.DurationS, 240) {
		t.Fatalf("时长应 240 秒，实际 %f", r.DurationS)
	}

	// t=30 在第一段中点
	p := r.PositionAtTime(30)
	if !nearlyEq(p.Lat, 39.905) {
		t.Fatalf("t=30 纬度应约 39.905，实际 %v", p.Lat)
	}
	if p.SpeedKmh < 60 || p.SpeedKmh > 70 {
		t.Fatalf("第一段速度应约 66.6 km/h，实际 %f", p.SpeedKmh)
	}

	// t=150 在第二段中部（速度约 11.1）
	p = r.PositionAtTime(150)
	if p.SpeedKmh < 10 || p.SpeedKmh > 13 {
		t.Fatalf("第二段速度应约 11.1 km/h，实际 %f", p.SpeedKmh)
	}

	// 越界收敛到终点
	p = r.PositionAtTime(99999)
	if !nearlyEq(p.Lat, 39.915) {
		t.Fatalf("越界时间应收敛到终点，实际 %v", p.Lat)
	}
}

func TestNewRouteValidation(t *testing.T) {
	if _, err := NewRoute("x", "text", []Point{{Lat: 1, Lon: 1}}); err == nil {
		t.Fatal("单点轨迹应报错")
	}
	if _, err := NewRoute("x", "text", []Point{{Lat: 91, Lon: 1}, {Lat: 2, Lon: 2}}); err == nil {
		t.Fatal("纬度越界应报错")
	}
	if _, err := NewRoute("x", "text", []Point{{Lat: 1, Lon: 181}, {Lat: 2, Lon: 2}}); err == nil {
		t.Fatal("经度越界应报错")
	}
	// 时间非严格递增 → 时间轴失效而非报错
	r, err := NewRoute("x", "text", []Point{
		{Lat: 1, Lon: 1, TimeSec: 10},
		{Lat: 2, Lon: 2, TimeSec: 10},
		{Lat: 3, Lon: 3, TimeSec: 30},
	})
	if err != nil {
		t.Fatalf("时间重复不应报错: %v", err)
	}
	if r.HasTime {
		t.Fatal("时间非严格递增时 HasTime 应为 false")
	}
}

func TestZeroLengthSegments(t *testing.T) {
	// 重复点（零长度段）不应导致定位异常
	r, err := NewRoute("重合", "text", []Point{
		{Lat: 39.9, Lon: 116.4},
		{Lat: 39.9, Lon: 116.4},
		{Lat: 39.91, Lon: 116.4},
		{Lat: 39.91, Lon: 116.4},
	})
	if err != nil {
		t.Fatalf("含重复点的轨迹应合法: %v", err)
	}
	p := r.PositionAtDist(r.TotalDistM / 2)
	if !nearlyEq(p.Lat, 39.905) {
		t.Fatalf("跳过零长度段后中点应约 39.905，实际 %v", p.Lat)
	}
	// 全部重合的退化轨迹
	r2, err := NewRoute("全重合", "text", []Point{
		{Lat: 39.9, Lon: 116.4},
		{Lat: 39.9, Lon: 116.4},
	})
	if err != nil {
		t.Fatalf("全重合轨迹应合法: %v", err)
	}
	_ = r2.PositionAtDist(0) // 不应 panic
	_ = r2.PositionAtTime(0)
}
