# JT-Simulate | 部标协议仿真平台

> 做部标协议开发、车联网开发的人都该用的仿真工具。

**JT-Simulate** 是一个开箱即用的部标协议仿真平台。Windows 双击 exe 直接弹出桌面窗口，Linux/macOS 一行命令启动 Web 服务。支持 JT/T 808、JT/T 809、JT/T 1078、JT/T 905、JT/T 1045、JT/T 1253、GB/T 32960 等 7 种协议的报文构造、解析、设备仿真、压力测试和合规检查——你开发过程中需要的调试工具，这里全有。

---

## 能干什么？

| 你在做什么 | JT-Simulate 怎么帮你 |
|------------|---------------------|
| 开发 JT/T 808 终端或平台 | 报文工坊一键生成/解析报文，不用手算校验码 |
| 调试设备注册/鉴权/位置上报流程 | 创建仿真设备，一键启动自动走完整个上线流程 |
| 测试平台并发能力 | 压力测试模拟几十上百台设备同时连接 |
| 验证报文是否符合协议规范 | 合规检查自动验证字段范围、格式等 |
| 测试协议解析器健壮性 | Fuzz 测试自动生成各种异常报文 |
| 模拟超速/疲劳/紧急报警等业务场景 | 场景模拟一键运行预置脚本 |

---

## 支持协议

| 协议 | 版本 | 说明 |
|------|------|------|
| **JT/T 808** | 2011 / 2013 / 2019 | 道路运输车辆卫星定位系统终端通信协议 |
| **JT/T 809** | 2011 / 2019 | 道路运输车辆卫星定位系统平台数据交换 |
| **JT/T 1078** | 2016 / 2022 | 道路运输车辆卫星定位系统视频通信协议 |
| **JT/T 905** | 2014 | 出租汽车服务管理信息系统协议 |
| **JT/T 1045** | 2018 | 道路运输车辆卫星定位系统信息安全协议 |
| **JT/T 1253** | 2019 | 危险货物道路运输安全协议 |
| **GB/T 32960** | 2016 | 电动汽车远程服务与管理系统技术规范 |

---

## 快速开始（三种方式，总有一种适合你）

### 方式一：直接下载发布包（推荐，零门槛）

1. 到 [Releases 页面](../../releases) 下载对应你系统的 zip 包
2. 解压
3. 运行：

**Windows（桌面应用，双击即用）：**
```
双击 jt-simulate.exe
```
> Windows 版是原生桌面应用，双击就会弹出软件窗口，不需要浏览器、不需要命令行。基于 WebView2 技术，Win10/Win11 系统自带运行时。

**Linux / macOS（Web 服务模式）：**
```bash
chmod +x jt-simulate
./jt-simulate serve -c configs/config.yaml
```
浏览器打开 `http://localhost:8095`

> 就这么简单，不需要安装任何其他东西。单个可执行文件，前端已经打包进去了。

### 方式二：从源码编译

**前置条件：** 安装 [Go 1.22+](https://go.dev/dl/)

```bash
git clone <repo-url>
cd JT-Simulate

# 编译（Web 服务模式，适用所有平台）
go build -o jt-simulate .

# 运行
./jt-simulate serve -c configs/config.yaml
```

浏览器打开 `http://localhost:8095` 即可使用。

**Windows 桌面应用编译（需要 Wails CLI）：**
```powershell
# 安装 Wails CLI（仅需一次）
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 编译桌面应用
wails build -platform windows/amd64

# 产物在 build/bin/jt-simulate.exe，双击即可运行
```

### 方式三：打包发布（开发者用）

如果你要自己打包发布版本，用 `make-release.ps1` 脚本，详见下方[打包脚本](#打包脚本-make-releaseps1)章节。

---

## 使用指南（手把手教你用）

### 1. 仪表盘 — 看全局

打开浏览器进入首页就是仪表盘，一眼看到：
- **设备总数** / **在线设备** / **消息总数** — 实时刷新
- **支持的协议和版本** — 一览无余
- **快速操作按钮** — 直接跳到报文工坊、设备管理、压力测试、合规检查

### 2. 报文工坊 — 构造和解析报文

这是最常用的功能。左边构造报文，右边解析报文。

**怎么用：**
1. 选择协议（如 JT/T 808-2019）
2. 选择消息ID（如 0x0200 位置上报）
3. 填写字段（终端手机号、经纬度、速度等）
4. 点「生成报文」→ 左边显示 Hex 报文
5. 报文自动同步到右边分析区 → 点「分析」看解析结果
6. 点「复制」按钮可以复制 Hex 报文

> **小白提示：** 左边生成的报文会自动填到右边分析区，编解码是否一致一目了然。

### 3. 设备管理 — 创建仿真设备

**怎么用：**
1. 填写终端手机号（12位数字，如 `013800001234`）
2. 填写目标平台地址（如 `127.0.0.1:7611`，这是你的 JT/T 808 平台监听地址）
3. 点「创建设备」
4. 设备列表里点「启动」→ 仿真设备自动连接平台，发送注册→鉴权→心跳→位置上报

> **注意：** 启动设备前，请确认目标平台已经在对应端口监听。如果连接失败，会弹出排查指南帮你定位问题。

### 4. 消息监控 — 实时看报文交互

**怎么用：**
1. 选择要监控的设备（或选「全部设备」）
2. 点「开始监控」
3. 页面实时显示设备收发的每一条报文
4. `[UP]` 表示终端上报，`[DOWN]` 表示平台下发
5. 点「停止监控」结束

### 5. 压力测试 — 批量设备并发测试

**怎么用：**
1. 设置设备数量（如 100 台）
2. 设置测试时长（如 60 秒）
3. 填写目标平台地址
4. 点「开始压测」
5. 实时查看：总设备数、在线设备数、总消息数、错误数

### 6. 场景模拟 — 一键运行业务场景

预置了 5 个场景，点一下就跑：

| 场景 | 说明 |
|------|------|
| 终端上线流程 | 注册→鉴权→心跳→位置上报完整流程 |
| 超速报警 | 从正常速度逐渐加速到超速触发报警 |
| 疲劳驾驶 | 连续驾驶超过4小时触发疲劳报警 |
| 紧急报警 | 触发紧急按钮报警 |
| 离线重连 | 断线后重连并补报历史位置 |

**怎么用：**
1. 选择场景
2. 填写目标地址
3. 点「开始模拟」→ 进度条实时显示执行进度
4. 点「停止模拟」可以随时中断

### 7. 合规检查 + Fuzz 测试

**合规检查：** 验证报文是否符合协议规范
1. 填写 JT/T 808 报文 Hex
2. 点「开始检查」
3. 查看每项检查结果（纬度范围、经度范围、速度范围、方向范围、时间格式等）

**Fuzz 测试：** 自动生成异常报文测试解析器健壮性
1. 点「开始 Fuzz 测试」
2. 自动生成 19 个测试用例（空帧、超长帧、非法字段、随机突变等）
3. 查看每个用例的结果（成功/错误/崩溃）

---

## 配置说明

配置文件在 `configs/config.yaml`：

```yaml
server:
  host: "0.0.0.0"       # 监听地址，0.0.0.0 表示所有网卡
  port: 8095             # 监听端口
  mode: release          # release 模式不输出调试日志

simulator:
  default_protocol: jt808
  heartbeat_interval: 60   # 心跳间隔（秒）
  location_interval: 30    # 位置上报间隔（秒）
  reconnect_interval: 15   # 重连间隔（秒）

targets:
  - name: default
    protocol: jt808
    address: "127.0.0.1:7611"  # 默认目标平台地址
    active: true
```

> **改端口？** 把 `port: 8095` 改成你想要的端口，重启即可。

---

## 打包脚本 make-release.ps1

如果你需要自己编译打包发布版本（比如修改了代码后要发布），用这个脚本。

### 前置条件

- Windows + PowerShell
- 安装 [Go 1.22+](https://go.dev/dl/)
- 安装 Wails CLI：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- 确保能联网（首次编译需要下载 Go 依赖）

### 用法

打开 PowerShell，进入项目根目录：

```powershell
# 基本用法（版本号固定 1.0.0）
powershell -ExecutionPolicy Bypass -File make-release.ps1

# 指定版本号
powershell -ExecutionPolicy Bypass -File make-release.ps1 -Version 1.0.0
```

### 脚本做了什么

1. **编译 5 个平台的二进制文件**：
   - Windows amd64：用 `wails build` 编译为桌面应用（双击弹出原生窗口）
   - Linux amd64 / arm64：用 `go build` 交叉编译（Web 服务模式）
   - macOS amd64 / arm64：用 `go build` 交叉编译（Web 服务模式）

2. **打包 zip**，每个包包含：
   - 可执行文件
   - 默认配置文件 (`configs/config.yaml`)
   - README.md

3. **生成 SHA256 校验文件** (`release/checksums.txt`)

### 产出

```
release/
├── jt-simulate-1.0.0-windows-amd64.zip
├── jt-simulate-1.0.0-linux-amd64.zip
├── jt-simulate-1.0.0-linux-arm64.zip
├── jt-simulate-1.0.0-darwin-amd64.zip
├── jt-simulate-1.0.0-darwin-arm64.zip
└── checksums.txt
```

### 验证下载包完整性

下载方可以用 `checksums.txt` 验证 zip 包是否完整：

```powershell
# Windows
Get-FileHash jt-simulate-1.0.0-windows-amd64.zip -Algorithm SHA256
# 对比 checksums.txt 中的值
```

```bash
# Linux / macOS
shasum -a 256 jt-simulate-1.0.0-linux-amd64.zip
# 对比 checksums.txt 中的值
```

---

## 命令行用法

JT-Simulate 支持两种运行模式：

### 桌面模式（默认，Windows 推荐）

```bash
# 直接运行（Windows 双击 exe 也是这个模式）
jt-simulate
```
弹出原生桌面窗口，不需要浏览器。

### Web 服务模式（Linux / macOS / 服务器部署）

```bash
# 启动 Web 服务
jt-simulate serve -c configs/config.yaml
```
浏览器访问 `http://localhost:8095`。

### 其他命令

```bash
# 分析一条报文
jt-simulate analyze 7E02000026013800001234000100...

# 查看版本
jt-simulate version
```

---

## API 接口

所有功能都提供 RESTful API，方便集成到你的工具链：

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/workshop/encode` | 编码报文 |
| POST | `/api/v1/workshop/analyze` | 解析报文 |
| GET | `/api/v1/devices` | 获取设备列表 |
| POST | `/api/v1/devices` | 创建设备 |
| GET | `/api/v1/devices/:id` | 获取设备详情 |
| POST | `/api/v1/devices/:id/start` | 启动设备 |
| POST | `/api/v1/devices/:id/stop` | 停止设备 |
| DELETE | `/api/v1/devices/:id` | 删除设备 |
| POST | `/api/v1/check/compliance` | 合规检查 |
| POST | `/api/v1/check/fuzz` | Fuzz 模糊测试 |
| GET | `/api/v1/scenarios` | 获取场景列表 |
| POST | `/api/v1/scenarios/run` | 运行场景 |
| POST | `/api/v1/scenarios/stop` | 停止场景 |
| GET | `/api/v1/scenarios/status` | 场景状态 |
| POST | `/api/v1/stress/run` | 启动压测 |
| POST | `/api/v1/stress/stop` | 停止压测 |
| GET | `/api/v1/stress/status` | 压测状态 |
| GET | `/api/v1/stats` | 引擎统计 |
| GET | `/api/v1/ws/monitor` | WebSocket 实时监控 |

---

## 与外部平台联调

JT-Simulate 创建的仿真设备可以连接任何标准的 JT/T 808 平台进行联调测试：

1. 启动你的目标平台（如 JTE），确认监听端口（默认 7611）
2. 在「设备管理」中创建设备，目标地址填写平台地址
3. 启动设备，仿真终端会自动注册并开始上报
4. 在「消息监控」查看实时报文交互

> **小白提示：** 如果不确定平台端口，一般 JT/T 808 默认就是 7611。

---

## 项目结构

```
JT-Simulate/
├── main.go                 # 程序入口（桌面模式 + Web 模式）
├── desktop.go              # Wails 桌面应用逻辑
├── cmd/jt-simulate/
│   └── frontend/dist/      # 前端静态文件（编译时嵌入二进制）
├── internal/
│   ├── api/                # API 服务器与路由
│   │   ├── handler/        # HTTP 处理器
│   │   └── websocket/      # WebSocket Hub
│   ├── checker/            # 合规检查器
│   ├── config/             # 配置管理
│   ├── engine/             # 仿真引擎（设备管理、压测、场景）
│   ├── simulator/          # 设备仿真器
│   │   ├── base/           # 仿真器基类
│   │   └── jt808/          # JT/T 808 仿真器
│   ├── storage/            # 存储层
│   └── workshop/           # 报文工坊（编解码）
├── pkg/
│   ├── codec/              # 各协议编解码器
│   │   ├── gbt32960/
│   │   ├── jt808/
│   │   ├── jt809/
│   │   ├── jt1078/
│   │   ├── jt905/
│   │   ├── jt1045/
│   │   ├── jt1253/
│   │   └── regional/
│   ├── simulate/           # 仿真工具
│   └── types/              # 公共类型
├── configs/                # 配置文件
├── wails.json              # Wails 配置
├── make-release.ps1        # 打包发布脚本
└── README.md
```

---

## 常见问题

**Q：Windows 双击 exe 没反应？**
A：JT-Simulate Windows 版是桌面应用，需要 WebView2 运行时。Win10 1803+ 和 Win11 默认已安装。如果没有，从微软官网下载安装即可。

**Q：启动设备后报"无法连接到目标平台"？**
A：说明目标平台没有启动或者端口不对。确认你的平台在运行，端口默认 7611。

**Q：浏览器打开是白屏？**
A：确认服务已经启动。命令行里应该显示 `服务模式启动: http://0.0.0.0:8095`。

**Q：修改了端口不生效？**
A：修改 `configs/config.yaml` 后需要重启服务。

**Q：支持 Linux 服务器部署吗？**
A：支持。下载 Linux 版本，`chmod +x jt-simulate` 然后 `./jt-simulate serve -c configs/config.yaml` 即可。

**Q：数据存在哪里？**
A：默认使用内存存储，重启后清空。适合开发调试用途。

---

## License

MIT License
