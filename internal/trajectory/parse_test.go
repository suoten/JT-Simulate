package trajectory

import (
	"strings"
	"testing"
)

const sampleGPX = `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="test" xmlns="http://www.topografix.com/GPX/1/1">
  <trk>
    <name>demo</name>
    <trkseg>
      <trkpt lat="39.9093" lon="116.3974"><ele>43</ele><time>2026-01-01T08:00:00Z</time></trkpt>
      <trkpt lat="39.9103" lon="116.3974"><ele>44</ele><time>2026-01-01T08:01:00Z</time></trkpt>
      <trkpt lat="39.9113" lon="116.3974"><ele>45</ele><time>2026-01-01T08:02:00Z</time></trkpt>
    </trkseg>
  </trk>
</gpx>`

func TestParseGPX(t *testing.T) {
	r, err := Parse("示例", []byte(sampleGPX))
	if err != nil {
		t.Fatalf("解析 GPX 失败: %v", err)
	}
	if r.Source != "gpx" {
		t.Fatalf("Source 应为 gpx，实际 %s", r.Source)
	}
	if len(r.Points) != 3 {
		t.Fatalf("点数应为 3，实际 %d", len(r.Points))
	}
	if !r.HasTime {
		t.Fatal("带时间 GPX 应启用时间轴")
	}
	if !nearlyEq(r.DurationS, 120) {
		t.Fatalf("时长应 120 秒，实际 %f", r.DurationS)
	}
	if !nearlyEq(r.Points[0].Lat, 39.9093) {
		t.Fatalf("首点纬度错误: %v", r.Points[0].Lat)
	}
}

func TestParseGPXMultiTrackAndSegment(t *testing.T) {
	gpx := `<gpx><trk><trkseg>
		<trkpt lat="39.9" lon="116.4"><time>2026-01-01T00:00:00Z</time></trkpt>
		<trkpt lat="39.91" lon="116.4"><time>2026-01-01T00:01:00Z</time></trkpt>
	</trkseg><trkseg>
		<trkpt lat="39.92" lon="116.4"><time>2026-01-01T00:02:00Z</time></trkpt>
		<trkpt lat="39.93" lon="116.4"><time>2026-01-01T00:03:00Z</time></trkpt>
	</trkseg></trk></gpx>`
	r, err := Parse("multi", []byte(gpx))
	if err != nil {
		t.Fatalf("多段 GPX 应可解析: %v", err)
	}
	if len(r.Points) != 4 {
		t.Fatalf("点数应为 4，实际 %d", len(r.Points))
	}
}

func TestParseGPXWithoutTime(t *testing.T) {
	gpx := `<gpx><trk><trkseg>
		<trkpt lat="39.9" lon="116.4"/><trkpt lat="39.91" lon="116.4"/>
	</trkseg></trk></gpx>`
	r, err := Parse("notime", []byte(gpx))
	if err != nil {
		t.Fatalf("无时间 GPX 应可解析: %v", err)
	}
	if r.HasTime {
		t.Fatal("无时间 GPX 不应启用时间轴")
	}
}

func TestParseGPXInvalid(t *testing.T) {
	if _, err := Parse("bad", []byte("<gpx><trk><trkseg></trkseg></trk></gpx>")); err == nil {
		t.Fatal("空轨迹应报错")
	}
	if _, err := Parse("bad", []byte("<gpx><trk><trkseg><trkpt lat=\"91\" lon=\"0\"/><trkpt lat=\"92\" lon=\"0\"/></trkseg></trk></gpx>")); err == nil {
		t.Fatal("纬度越界应报错")
	}
	badTime := `<gpx><trk><trkseg>
		<trkpt lat="39.9" lon="116.4"><time>not-a-time</time></trkpt>
		<trkpt lat="39.91" lon="116.4"><time>also-bad</time></trkpt>
	</trkseg></trk></gpx>`
	if _, err := Parse("bad", []byte(badTime)); err == nil {
		t.Fatal("非法时间应报错")
	}
}

func TestParseTextVariants(t *testing.T) {
	t.Run("基础逗号分隔+表头+注释", func(t *testing.T) {
		content := "# 注释行\nlat,lng,speed,time\n39.9093,116.3974,40,0\n39.9100,116.4000,45,30\n39.9120,116.4050,55,90\n"
		r, err := Parse("基础", []byte(content))
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		if len(r.Points) != 3 {
			t.Fatalf("点数应为 3，实际 %d", len(r.Points))
		}
		if !r.HasTime || !nearlyEq(r.DurationS, 90) {
			t.Fatalf("时间轴异常: has=%v dur=%f", r.HasTime, r.DurationS)
		}
		if r.Points[1].SpeedKmh != 45 {
			t.Fatalf("速度应 45，实际 %f", r.Points[1].SpeedKmh)
		}
	})

	t.Run("仅坐标两列", func(t *testing.T) {
		content := "39.9 116.4\n39.91 116.4\n39.92 116.4\n"
		r, err := Parse("两列", []byte(content))
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		if r.HasTime || r.Points[0].SpeedKmh != 0 {
			t.Fatal("两列格式不应有时间和速度")
		}
	})

	t.Run("HHMMSS时间跨午夜", func(t *testing.T) {
		content := "39.9,116.4,,23:59:00\n39.91,116.4,,00:01:00\n"
		r, err := Parse("跨午夜", []byte(content))
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		if !r.HasTime {
			t.Fatal("HH:MM:SS 应启用时间轴")
		}
		// 23:59:00 → 00:01:00 跨午夜应得 120 秒
		if !nearlyEq(r.DurationS, 120) {
			t.Fatalf("跨午夜时长应 120 秒，实际 %f", r.DurationS)
		}
	})

	t.Run("空格拆开的日期时间", func(t *testing.T) {
		content := "39.9,116.4,2026-01-01 08:00:00\n39.91,116.4,2026-01-01 08:02:00\n"
		r, err := Parse("日期时间", []byte(content))
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		if !r.HasTime || !nearlyEq(r.DurationS, 120) {
			t.Fatalf("日期时间轴异常: has=%v dur=%f", r.HasTime, r.DurationS)
		}
	})

	t.Run("Tab与分号分隔", func(t *testing.T) {
		content := "39.9;\t116.4\t40\n39.91;\t116.4\t50\n"
		r, err := Parse("分隔符", []byte(content))
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		if len(r.Points) != 2 || r.Points[1].SpeedKmh != 50 {
			t.Fatalf("Tab/分号解析异常: %+v", r.Points)
		}
	})

	t.Run("速度加RFC3339时间戳", func(t *testing.T) {
		content := "39.9,116.4,30,2026-01-01T08:00:00Z\n39.91,116.4,60,2026-01-01T08:10:00Z\n"
		r, err := Parse("混合", []byte(content))
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}
		if !r.HasTime || !nearlyEq(r.DurationS, 600) {
			t.Fatalf("RFC3339 时间轴异常: has=%v dur=%f", r.HasTime, r.DurationS)
		}
		if r.Points[1].SpeedKmh != 60 {
			t.Fatalf("速度应 60，实际 %f", r.Points[1].SpeedKmh)
		}
	})
}

func TestParseTextInvalid(t *testing.T) {
	if _, err := Parse("空", []byte("")); err == nil {
		t.Fatal("空文件应报错")
	}
	if _, err := Parse("单点", []byte("39.9,116.4\n")); err == nil {
		t.Fatal("单点应报错")
	}
	if _, err := Parse("坏行", []byte("39.9,116.4\n39.91,abc\n")); err == nil {
		t.Fatal("坐标非法应报错")
	}
	if _, err := Parse("坏时间", []byte("39.9,116.4,25:99:00\n39.91,116.4,08:00:00\n")); err == nil {
		t.Fatal("时间越界应报错")
	}
}

func TestParseAutoDetect(t *testing.T) {
	r, err := Parse("auto", []byte(sampleGPX))
	if err != nil || r.Source != "gpx" {
		t.Fatalf("GPX 自动识别失败: %v", err)
	}
	r, err = Parse("auto", []byte("39.9,116.4\n39.91,116.4\n"))
	if err != nil || r.Source != "text" {
		t.Fatalf("文本自动识别失败: %v", err)
	}
}

func TestSanitizeName(t *testing.T) {
	cases := []struct {
		in   string
		want string
		err  bool
	}{
		{"demo-route_1", "demo-route_1", false},
		{"示例-北京环路", "示例-北京环路", false},
		{"  spaced name  ", "spaced_name", false},
		{"a/b\\c:d", "a_b_c_d", false},
		{"../../etc/passwd", "etc_passwd", false},
		{"***", "", true},
		{"", "", true},
	}
	for _, c := range cases {
		got, err := SanitizeName(c.in)
		if c.err && err == nil {
			t.Fatalf("SanitizeName(%q) 应报错", c.in)
		}
		if !c.err && (err != nil || got != c.want) {
			t.Fatalf("SanitizeName(%q) = %q (err=%v), want %q", c.in, got, err, c.want)
		}
		// 结果必须不含路径分隔符
		if strings.ContainsAny(got, `/\`) {
			t.Fatalf("清洗后的名称不应含路径分隔符: %q", got)
		}
	}
}

func TestStoreSaveLoadDelete(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	r := straightRoute()
	r.Name = "测试:路线/1"
	if err := s.Save(r); err != nil {
		t.Fatalf("保存失败: %v", err)
	}

	got, err := s.Get("测试_路线_1")
	if err != nil {
		t.Fatalf("按清洗名加载失败: %v", err)
	}
	if !nearlyEq(got.TotalDistM, r.TotalDistM) {
		t.Fatalf("往返距离不一致: %f vs %f", got.TotalDistM, r.TotalDistM)
	}
	if !nearlyEq(got.PositionAtDist(100).Lat, r.PositionAtDist(100).Lat) {
		t.Fatal("加载后插值结果应一致")
	}

	list, err := s.List()
	if err != nil || len(list) != 1 {
		t.Fatalf("列表应含 1 条: n=%d err=%v", len(list), err)
	}

	if err := s.Delete("测试:路线/1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	if _, err := s.Get("测试_路线_1"); err == nil {
		t.Fatal("删除后读取应报错")
	}

	// 不存在目录的 List 应返回空而非报错
	empty := NewStore(dir + "/nonexistent")
	l, err := empty.List()
	if err != nil || len(l) != 0 {
		t.Fatalf("不存在目录的 List 应为空: %v", err)
	}
}
