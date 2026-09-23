package types

// ProtocolVersion 协议版本
type ProtocolVersion string

const (
	// JT/T 808 版本
	VersionJT808_2011 ProtocolVersion = "2011"
	VersionJT808_2013 ProtocolVersion = "2013"
	VersionJT808_2019 ProtocolVersion = "2019"

	// JT/T 809 版本
	VersionJT809_2011 ProtocolVersion = "2011"
	VersionJT809_2019 ProtocolVersion = "2019"

	// JT/T 1078 版本
	VersionJT1078_2016 ProtocolVersion = "2016"
	VersionJT1078_2022 ProtocolVersion = "2022"

	// JT/T 905 版本
	VersionJT905_2014 ProtocolVersion = "2014"

	// JT/T 1045 版本
	VersionJT1045_2018 ProtocolVersion = "2018"

	// JT/T 1253 版本
	VersionJT1253_2019 ProtocolVersion = "2019"

	// GB/T 32960 版本
	VersionGBT32960_2016 ProtocolVersion = "2016"
)

// ProtocolVersions 每个协议支持的版本列表
var ProtocolVersions = map[ProtocolType][]ProtocolVersion{
	ProtocolJT808:   {VersionJT808_2011, VersionJT808_2013, VersionJT808_2019},
	ProtocolJT809:   {VersionJT809_2011, VersionJT809_2019},
	ProtocolJT1078:  {VersionJT1078_2016, VersionJT1078_2022},
	ProtocolJT905:   {VersionJT905_2014},
	ProtocolJT1045:  {VersionJT1045_2018},
	ProtocolJT1253:  {VersionJT1253_2019},
	ProtocolGBT32960: {VersionGBT32960_2016},
}

// VersionLabel 版本显示名称
var VersionLabel = map[ProtocolType]map[ProtocolVersion]string{
	ProtocolJT808: {
		VersionJT808_2011: "JT/T 808-2011",
		VersionJT808_2013: "JT/T 808-2013",
		VersionJT808_2019: "JT/T 808-2019",
	},
	ProtocolJT809: {
		VersionJT809_2011: "JT/T 809-2011",
		VersionJT809_2019: "JT/T 809-2019",
	},
	ProtocolJT1078: {
		VersionJT1078_2016: "JT/T 1078-2016",
		VersionJT1078_2022: "JT/T 1078-2022",
	},
	ProtocolJT905: {
		VersionJT905_2014: "JT/T 905-2014",
	},
	ProtocolJT1045: {
		VersionJT1045_2018: "JT/T 1045-2018",
	},
	ProtocolJT1253: {
		VersionJT1253_2019: "JT/T 1253-2019",
	},
	ProtocolGBT32960: {
		VersionGBT32960_2016: "GB/T 32960.3-2016",
	},
}

// DefaultVersion 每个协议的默认版本
var DefaultVersion = map[ProtocolType]ProtocolVersion{
	ProtocolJT808:    VersionJT808_2019,
	ProtocolJT809:    VersionJT809_2019,
	ProtocolJT1078:   VersionJT1078_2016,
	ProtocolJT905:    VersionJT905_2014,
	ProtocolJT1045:   VersionJT1045_2018,
	ProtocolJT1253:   VersionJT1253_2019,
	ProtocolGBT32960: VersionGBT32960_2016,
}

// IsVersion2019 判断808是否为2019版本
func IsVersion2019(ver ProtocolVersion) bool {
	return ver == VersionJT808_2019
}
