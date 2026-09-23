// JT-Simulate Frontend JS v7 - Pages
// Pages: devices, monitor, stress, scenario, checker

function f808(m){switch(m){
    case '0x0001':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+FI('resp_msg_id','应答消息ID','0x8103')+NI('result','结果(0成功)','0')+'</div>';
    case '0x8001':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+FI('resp_msg_id','应答消息ID','0x0100')+NI('result','结果(0成功)','0')+'</div>';
    case '0x0002':return '<p style="color:#888;font-size:13px">心跳消息无需参数</p>';
    case '0x0003':return '<p style="color:#888;font-size:13px">终端注销无需参数</p>';
    case '0x0100':return '<div class="grid-2">'+NI('province_id','省域ID','11')+NI('city_id','市县域ID','100')+FI('manufacturer','制造商(5字节)','SUOTE')+FI('terminal_model','终端型号(20字节)','JT-100')+FI('terminal_id','终端ID(7字节)','1234567')+NI('plate_color','车牌颜色','1')+FI('plate_number','车牌号','京A12345')+'</div>';
    case '0x8100':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+NI('result','结果(0成功)','0')+FI('auth_code','鉴权码','AUTHCODE123')+'</div>';
    case '0x0102':return FI('auth_code','鉴权码','test123')+FI('imei','IMEI(仅2019)','123456789012345');
    case '0x8103':return '<div class="grid-2">'+NI('param_id','参数ID(十进制)','0x0001')+FI('param_value','参数值','1')+'</div>';
    case '0x0103':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+NI('result','结果(0成功)','0')+NI('param_count','参数数量','1')+'</div>';
    case '0x8104':return NI('param_id','参数ID(十进制)','0x0001');
    case '0x0104':return '<div class="grid-2">'+NI('param_id','参数ID(十进制)','0x0001')+FI('param_value','参数值','1')+'</div>';
    case '0x8105':return '<div class="grid-2">'+NI('command','控制命令','1')+FI('param','参数','')+'</div>';
    case '0x8107':return '<p style="color:#888;font-size:13px">终端属性查询无需参数</p>';
    case '0x0107':return '<div class="grid-2">'+NI('terminal_type','终端类型','1')+FI('manufacturer','制造商(5字节)','SUOTE')+FI('terminal_model','终端型号(20字节)','JT-100')+FI('terminal_id','终端ID(7字节)','1234567')+FI('iccid','ICCID(20位)','8986011899100123456')+'</div>';
    case '0x8108':return '<div class="grid-2">'+NI('upgrade_type','升级类型','0')+FI('manufacturer','制造商(5字节)','SUOTE')+FI('model','型号(20字节)','JT-100')+FI('version','版本','1.0')+FI('url','URL','http://127.0.0.1')+'</div>';
    case '0x0108':return '<div class="grid-2">'+NI('upgrade_type','升级类型','0')+NI('result','结果(0成功)','0')+FI('upgrade_msg','升级消息','OK')+'</div>';
    case '0x0200':return '<div class="grid-2">'+NI('latitude','纬度','39.9093')+NI('longitude','经度','116.3974')+NI('speed','速度(km/h)','60')+NI('direction','方向(0-359)','180')+NI('altitude','海拔(m)','5000')+NI('alarm_flag','报警标志','0')+NI('status_flag','状态标志','0')+'</div>';
    case '0x8201':return '<p style="color:#888;font-size:13px">位置查询无需参数</p>';
    case '0x0201':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+NI('latitude','纬度','39.9093')+NI('longitude','经度','116.3974')+NI('speed','速度(km/h)','60')+NI('direction','方向','180')+NI('altitude','海拔(m)','5000')+NI('alarm_flag','报警标志','0')+NI('status_flag','状态标志','0')+'</div>';
    case '0x0704':return '<div class="grid-2">'+NI('latitude','纬度','39.9093')+NI('longitude','经度','116.3974')+NI('speed','速度(km/h)','60')+NI('direction','方向(0-359)','180')+NI('altitude','海拔(m)','5000')+NI('alarm_flag','报警标志','0')+NI('status_flag','状态标志','0')+'</div>';
    case '0x8202':return '<div class="grid-2">'+NI('interval','间隔(秒)','5')+NI('validity','有效时间(秒)','60')+'</div>';
    case '0x0202':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+NI('result','结果(0成功)','0')+'</div>';
    case '0x8203':return '<div class="grid-2">'+NI('seq_num','流水号','1')+NI('alarm_id','报警ID','1')+'</div>';
    case '0x0900':return vehF()+NI('alarm_flag','报警标志','1')+NI('water_level','水位','0');
    case '0x0901':return vehF()+NI('alarm_id','报警ID','1')+NI('attachment_len','附件长度','0');
    case '0x9001':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+NI('result','结果(0成功)','0')+'</div>';
    case '0x0400':return NI('alarm_flag','报警标志','1');
    case '0x0401':return NI('alarm_flag','报警标志','1');
    case '0x8300':return NI('flag','标志','1')+FI('content','文本内容','测试文本');
    case '0x8301':return NI('event_id','事件ID','0')+FI('content','事件内容','测试事件');
    case '0x0301':return '<div class="grid-2">'+NI('seq_num','流水号','1')+NI('event_id','事件ID','0')+'</div>';
    case '0x0302':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+NI('answer_id','答案ID','0')+'</div>';
    case '0x8400':return '<div class="grid-2">'+NI('seq_num','流水号','1')+NI('speed','速度(km/h)','120')+NI('duration','持续时间(秒)','10')+'</div>';
    case '0x8401':return '<div class="grid-2">'+NI('seq_num','流水号','1')+NI('day_max_drive','日最大驾驶(秒)','28800')+NI('day_min_rest','日最小休息(秒)','7200')+NI('max_drive','最大连续驾驶(秒)','14400')+NI('min_rest','最小休息(秒)','1200')+'</div>';
    case '0x8500':return NI('control_flag','控制标志','0');
    case '0x8600':return '<div class="grid-2">'+NI('area_id','区域ID','1')+NI('attr','属性','0')+NI('latitude','中心纬度','39.9093')+NI('longitude','中心经度','116.3974')+NI('radius','半径(m)','500')+FI('start_time','起始时间','000000')+FI('end_time','结束时间','235959')+'</div>';
    case '0x8602':return '<div class="grid-2">'+NI('area_id','区域ID','1')+NI('attr','属性','0')+NI('upper_lat','上纬度','40.0')+NI('upper_lon','上经度','117.0')+NI('lower_lat','下纬度','39.5')+NI('lower_lon','下经度','116.0')+FI('start_time','起始时间','000000')+FI('end_time','结束时间','235959')+'</div>';
    case '0x8604':return '<div class="grid-2">'+NI('area_id','区域ID','1')+NI('attr','属性','0')+NI('lat1','顶点1纬度','39.9')+NI('lon1','顶点1经度','116.3')+FI('start_time','起始时间','000000')+FI('end_time','结束时间','235959')+'</div>';
    case '0x8606':return '<div class="grid-2">'+NI('route_id','路线ID','1')+NI('attr','属性','0')+FI('start_time','起始时间','000000')+FI('end_time','结束时间','235959')+NI('point_id','路段ID','1')+NI('latitude','纬度','39.9')+NI('longitude','经度','116.3')+NI('width','路宽','10')+NI('max_speed','最高速度','60')+NI('max_duration','最长时长','3600')+'</div>';
    case '0x0700':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+NI('menu_type','菜单类型','0')+'</div>';
    case '0x0702':return FI('driver_name','驾驶员姓名','张三')+FI('driver_id','驾驶员证号','110105199001011234')+FI('licence','从业资格证号','1234567890')+FI('certify_org','发证机构','北京交通委');
    case '0x0703':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+NI('result','结果(0成功)','0')+'</div>';
    case '0x0701':return FI('waybill_data','运单数据(JSON)','{"id":"WB001"');
    case '0x0705':return '<div class="grid-2">'+NI('can_time','CAN时间','0')+NI('can_id','CAN ID','1')+'</div>';
    case '0x0801':return '<div class="grid-2">'+NI('multimedia_id','多媒体ID','1')+NI('multimedia_type','类型(0图像)','0')+NI('format','格式(0JPEG)','0')+NI('event_code','事件编码','0')+NI('channel_id','通道ID','1')+'</div>';
    case '0x0802':return '<div class="grid-2">'+NI('multimedia_id','多媒体ID','1')+NI('multimedia_type','类型(0图像)','0')+NI('format','格式','0')+NI('play_time','播放时长','0')+NI('package_size','每包大小','0')+NI('total_packages','总包数','1')+NI('offset','偏移量','0')+'</div>';
    case '0x0803':return '<div class="grid-2">'+NI('seq_num','流水号','1')+NI('logical_ch','逻辑通道','1')+FI('start_time','开始时间','240101000000')+FI('end_time','结束时间','240101010000')+NI('media_type','媒体类型','0')+NI('stream_type','码流类型','0')+'</div>';
    case '0x0804':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+NI('result','结果(0成功)','0')+NI('multimedia_id','多媒体ID','1')+NI('package_size','每包大小','0')+NI('offset','偏移量','0')+'</div>';
    case '0x0805':return '<div class="grid-2">'+NI('resp_seq_num','应答流水号','1')+NI('result','结果(0成功)','0')+NI('multimedia_id','多媒体ID','1')+NI('channel_id','通道ID','1')+NI('package_size','每包大小','0')+'</div>';
    case '0x8801':return '<div class="grid-2">'+NI('seq_num','流水号','1')+NI('channel_id','通道ID','1')+NI('command','命令(0拍照)','0')+NI('interval','间隔(秒)','0')+NI('count','数量','1')+NI('resolution','分辨率','1')+'</div>';
    case '0x0A00':return FI('e_module','RSA模数(hex)','010203')+FI('exponent','RSA指数(hex)','010001');
    case '0x8A00':return FI('e_module','RSA模数(hex)','010203')+FI('exponent','RSA指数(hex)','010001');
    case '0x0B00':return NI('operate_type','操作类型','0')+FI('operate_data','操作数据(hex)','00');
    default:return '';
}}
function f809(m){var vc=NI('vehicle_color','车牌颜色','1'),vp=FI('vehicle_plate','车牌号','京A12345');switch(m){
    case '0x1001':return '<div class="grid-2">'+NI('user_name','用户名','1001')+FI('password','密码','123456')+FI('downlink_ip','下行IP','127.0.0.1')+NI('downlink_port','下行端口','8080')+'</div>';
    case '0x1003':return '<div class="grid-2">'+NI('user_name','用户名','1001')+FI('password','密码','123456')+'</div>';
    case '0x1200':return vc+vp+'<div class="grid-2">'+FI('vin','VIN','LSJA1234567890123')+NI('vehicle_type','车辆类型','1')+'</div>';
    case '0x1202':case '0x1203':case '0x1205':return vc+vp+locF();
    case '0x1300':return vc+vp+'<div class="grid-2">'+NI('alarm_source','报警来源','0')+NI('alarm_type','报警类型','1')+'</div>'+locF();
    case '0x1002':return '<div class="grid-2">'+NI('result','结果(0成功)','0')+NI('verify_code','校验码','12345')+FI('downlink_ip','下行IP','127.0.0.1')+NI('downlink_port','下行端口','8080')+'</div>';
    case '0x1701':return vc+vp+'<div class="grid-2">'+FI('driver_name','驾驶员姓名','张三')+FI('driver_id','驾驶员证号','110105199001011234')+FI('licence','从业资格证号','1234567890')+FI('org_name','机构名称','北京交通委')+'</div>';
    case '0x1801':return vc+vp+'<div class="grid-2">'+FI('waybill_id','运单号','WB20240101001')+NI('hazard_class','危险品类别','3')+FI('hazard_name','危险品名称','汽油')+NI('weight','重量(kg)','10000')+FI('origin','出发地','北京')+FI('destination','目的地','天津')+'</div>';
    case '0x1501':return vc+vp+'<div class="grid-2">'+NI('total_msg','总消息数','100')+NI('loc_count','定位数','80')+NI('alarm_count','报警数','5')+NI('online_duration','在线时长(秒)','3600')+'</div>';
    default:if('0x1004|0x1005|0x1006|0x9003|0x9004|0x9005|0x9006'.indexOf(m)>=0)return '<p style="color:#888;font-size:13px">该消息无需参数</p>';return vc+vp+locF();
}}
function f1078(m){var lc=NI('logical_ch','逻辑通道','1'),dt=NI('data_type','数据类型','0'),st=NI('stream_type','码流类型','0');switch(m){
    case '0x9101':return lc+dt+st;
    case '0x9102':return lc+NI('ctrl_cmd','控制命令','0')+NI('switch_type','切换类型','0');
    case '0x9301':return lc+NI('direction','方向','0')+NI('speed','速度','5');
    case '0x9201':return lc+NI('start_time','开始时间','240101000000')+NI('end_time','结束时间','240101010000')+dt+st;
    case '0x9203':return lc+NI('ctrl_cmd','控制命令','0')+NI('play_speed','播放速度','0')+FI('play_time','播放时间','240101000000');
    case '0x9202':return lc+NI('result','结果(0成功)','0')+NI('start_time','开始时间','240101000000')+NI('end_time','结束时间','240101010000');
    case '0x9501':return lc+dt+st+NI('resolution','分辨率','1')+NI('frame_rate','帧率','25')+NI('bitrate','码率','1024');
    default:return lc+dt+st;
}}
function f32960(m){switch(m){
    case '0x01':return FI('vin','VIN','LSJA1234567890123')+FI('iccid','ICCID','12345678901234567890');
    case '0x02':return locF();
    case '0x03':return FI('vin','VIN','LSJA1234567890123');
    case '0x04':return FI('platform_id','平台ID','10001')+FI('password','密码','123456')+FI('encrypt_key','加密密钥','AESKEY123');
    case '0x06':return NI('cmd_flag','命令标识','1')+NI('result','结果(0成功)','0');
    case '0x07':return '<p style="color:#888;font-size:13px">心跳消息无需参数</p>';
    default:return FI('vin','VIN','LSJA1234567890123');
}}
function f905(m){var vc=NI('vehicle_color','车牌颜色','1'),vp=FI('vehicle_plate','车牌号','京A12345');switch(m){
    case '0x0B01':return vc+vp+locF()+'<div class="grid-2">'+NI('distance','距离(m)','5000')+NI('amount','金额(分)','2000')+'</div>';
    case '0x0B00':return vc+vp+NI('status','状态','0')+locF();
    case '0x8B02':return NI('dispatch_type','调度类型','0')+FI('content','内容','测试调度')+FI('passenger_phone','乘客电话','13800001111')+FI('pickup_addr','上车地址','北京站');
    case '0x8B03':return '<div class="grid-2">'+NI('unit_price','单价(分)','250')+NI('unit_distance','单位距离(m)','1000')+NI('per_km_price','每公里价格(分)','250')+NI('night_surcharge','夜间附加(分)','50')+'</div>';
    case '0x8B04':return NI('ad_type','广告类型','0')+NI('play_times','播放次数','1')+FI('content','广告内容','测试广告');
    default:return vc+vp+locF();
}}
function f1045(m){var vc=NI('vehicle_color','车牌颜色','1'),vp=FI('vehicle_plate','车牌号','京A12345'),loc=locF();switch(m){
    case '0x0901':return vc+vp+'<div class="grid-2">'+NI('alarm_type','报警类型','1')+NI('alarm_level','报警级别','1')+NI('alarm_param','报警参数','100')+NI('altitude','海拔(m)','500')+'</div>'+loc;
    case '0x1205':return vc+vp+'<div class="grid-2">'+NI('alarm_type','报警类型','1')+NI('alarm_level','报警级别','1')+NI('alarm_param','报警参数','100')+'</div>'+loc;
    case '0x120B':return vc+vp+'<div class="grid-2">'+NI('tire_position','轮胎位置','1')+NI('pressure','胎压(kPa)','250')+NI('temp','温度','25')+NI('battery','电量','90')+NI('alarm_flag','报警标志','0')+'</div>';
    case '0x0902':return vc+vp+'<div class="grid-2">'+NI('alarm_type','报警类型','1')+NI('alarm_level','报警级别','1')+FI('vehicle_no','车牌号','京A12345')+NI('vehicle_speed','车速(km/h)','60')+NI('distance','距离(m)','100')+'</div>'+loc;
    default:return vc+vp+'<div class="grid-2">'+NI('alarm_type','报警类型','1')+NI('alarm_level','报警级别','1')+'</div>'+loc;
}}
function f1253(m){var vc=NI('vehicle_color','车牌颜色','1'),vp=FI('vehicle_plate','车牌号','京A12345'),loc=locF();switch(m){
    case '0x0D01':return vc+vp+'<div class="grid-2">'+FI('waybill_id','运单号','WB20240101001')+NI('hazard_class','危险品类别','3')+FI('hazard_name','危险品名称','汽油')+NI('weight','重量(kg)','10000')+FI('origin','出发地','北京')+FI('destination','目的地','天津')+'</div>';
    case '0x0D02':return vc+vp+'<div class="grid-2">'+NI('alarm_type','报警类型','1')+NI('alarm_level','报警级别','1')+'</div>'+loc;
    case '0x0D03':return vc+vp+'<div class="grid-2">'+NI('hazard_class','危险品类别','3')+NI('hazard_state','危险品状态','1')+NI('temp','温度','25')+NI('humidity','湿度','50')+NI('pressure','压力(Pa)','101325')+NI('level','液位(mm)','500')+'</div>'+loc;
    case '0x0D04':return vc+vp+'<div class="grid-2">'+NI('hazard_class','危险品类别','3')+NI('unload_amount','卸载数量(kg)','5000')+FI('unload_place','卸载地点','天津仓库')+'</div>'+loc;
    default:return vc+vp+loc;
}}
// ==================== 设备管理 ====================
function renderDevices(){
    document.getElementById('page-devices').innerHTML=
        '<h2 class="page-title">设备管理</h2>'+
        '<div class="card guide-card"><div style="display:flex;align-items:flex-start;gap:12px"><div style="flex-shrink:0;color:var(--accent)">'+icon('info','card-icon')+'</div><div><div style="color:var(--accent);font-size:13px;font-weight:600;margin-bottom:6px">使用指南</div><div style="color:var(--text-muted);font-size:12px;line-height:1.8">1. 创建设备：填写终端手机号和目标平台地址（默认 <code style="color:var(--accent)">127.0.0.1:7611</code>）<br>2. 启动设备：点击「启动」按钮，设备会自动连接目标平台并发送注册、鉴权、心跳、位置上报<br>3. 如果启动失败：请确认目标平台已启动并监听对应端口<br>4. 启动后可到「消息监控」页面查看实时报文交互</div></div></div></div>'+
        '<div class="grid-2"><div class="card"><h3>'+icon('plus')+'创建设备</h3>'+
        '<div class="form-group"><label>设备ID <span style="color:var(--text-dim);text-transform:none;font-weight:400">（留空自动用手机号）</span></label><input type="text" id="dev-id" placeholder="自动生成"></div>'+
        '<div class="grid-2"><div class="form-group"><label>协议类型</label><select id="dev-protocol">'+pO()+'</select></div><div class="form-group"><label>终端手机号 <span style="color:var(--danger);text-transform:none">*</span></label><input type="text" id="dev-phone" value="013800001234" placeholder="12位数字"></div></div>'+
        '<div class="grid-2"><div class="form-group"><label>车牌号</label><input type="text" id="dev-plate" value="京A12345"></div><div class="form-group"><label>目标地址 <span style="color:var(--text-dim);font-weight:400;text-transform:none">（平台监听地址）</span></label><input type="text" id="dev-target" value="127.0.0.1:7611" placeholder="如 127.0.0.1:7611"></div></div>'+
        '<div class="grid-2"><div class="form-group"><label>上报间隔(秒)</label><input type="number" id="dev-interval" value="30" min="1"></div><div class="form-group"><label>心跳间隔(秒)</label><input type="number" id="dev-heartbeat" value="60" min="1"></div></div>'+
        '<button class="btn btn-primary" onclick="createDevice()">'+btnIcon('plus')+'创建设备</button></div>'+
        '<div class="card"><h3>'+icon('devices')+'设备列表 <button class="btn btn-sm" onclick="loadDevices()" style="margin-left:8px" id="dev-refresh-btn">'+btnIcon('refresh')+'刷新</button></h3><div id="dev-list"><p style="color:#888">加载中...</p></div></div></div>';
    loadDevices();
    devTimer=setInterval(loadDevices,5000);
}
async function createDevice(){
    var cfg={id:document.getElementById('dev-id').value,protocol:document.getElementById('dev-protocol').value,phone:document.getElementById('dev-phone').value,plate:document.getElementById('dev-plate').value,target_addr:document.getElementById('dev-target').value,location_interval:parseInt(document.getElementById('dev-interval').value)||30,heartbeat_interval:parseInt(document.getElementById('dev-heartbeat').value)||60};
    if(!cfg.phone){toast('请填写终端手机号','error');return;}
    try{var r=await fetch(API+'/devices',{method:'POST',headers:{'Content-Type':'application/json; charset=utf-8'},body:JSON.stringify(cfg)});var d=await r.json();if(d.success){toast('设备创建成功','success');loadDevices();}else toast('创建失败: '+(d.error||'未知错误'),'error');}catch(e){toast('请求失败: '+e.message,'error');}
}
async function loadDevices(){
    try{
        var r=await fetch(API+'/devices'),d=await r.json(),list=document.getElementById('dev-list');if(!list)return;
        if(!d.devices||d.devices.length===0){list.innerHTML='<div class="empty-state">'+icon('empty','empty-icon')+'<p>暂无设备，请在左侧创建</p></div>';return;}
        var h='<table><thead><tr><th>ID</th><th>协议</th><th>手机号</th><th>车牌</th><th>目标地址</th><th>上报间隔</th><th>状态</th><th>操作</th></tr></thead><tbody>';
        d.devices.forEach(function(dev){
            var isOnline=dev.state==='online';var isConnecting=dev.state==='connecting';
            var btnStart=(isOnline||isConnecting)?'<button class="btn btn-sm" disabled>'+(isConnecting?'连接中':'已启动')+'</button>':'<button class="btn btn-sm btn-primary" onclick="startDevice(\''+dev.id+'\')">'+btnIcon('play')+'启动</button>';
            var btnStop=(isOnline||isConnecting)?'<button class="btn btn-sm btn-danger" onclick="stopDevice(\''+dev.id+'\')">'+btnIcon('stop')+'停止</button>':'<button class="btn btn-sm" disabled>已停止</button>';
            var btnDetail='<button class="btn btn-sm" onclick="showDeviceDetail(\''+dev.id+'\')" title="查看详情">'+btnIcon('eye')+'</button>';
            var btnDel='<button class="btn btn-sm btn-danger" onclick="confirmDelDevice(\''+dev.id+'\')" title="删除">'+btnIcon('trash')+'</button>';
            h+='<tr><td>'+escH(dev.id)+'</td><td>'+escH(dev.protocol)+'</td><td>'+escH(dev.phone)+'</td><td>'+escH(dev.plate)+'</td><td style="font-size:12px;color:#8b92a8">'+escH(dev.target||'-')+'</td><td>'+(dev.interval||30)+'s</td><td>'+stateBadge(dev.state)+'</td><td style="white-space:nowrap">'+btnStart+' '+btnStop+' '+btnDetail+' '+btnDel+'</td></tr>';
        });
        h+='</tbody></table>';list.innerHTML=h;
    }catch(e){var el=document.getElementById('dev-list');if(el)el.innerHTML='<p style="color:#ef5350">加载失败</p>';}
}
function confirmDelDevice(id){
    var overlay=document.createElement('div');
    overlay.className='modal-overlay';
    overlay.onclick=function(e){if(e.target===overlay)overlay.remove();};
    overlay.innerHTML='<div class="modal-content" style="max-width:400px"><div class="modal-header"><h3>'+icon('warning')+'确认删除</h3></div><div class="modal-body"><p style="color:#ccc;font-size:14px">确定要删除设备 <b style="color:var(--accent)">'+escH(id)+'</b> 吗？此操作不可撤销。</p></div><div class="modal-footer"><button class="btn" onclick="this.closest(\'.modal-overlay\').remove()">取消</button> <button class="btn btn-danger" onclick="delDevice(\''+id+'\');this.closest(\'.modal-overlay\').remove()">确认删除</button></div></div>';
    document.body.appendChild(overlay);
}
async function showDeviceDetail(id){
    try{
        var r=await fetch(API+'/devices/'+id);var d=await r.json();
        var overlay=document.createElement('div');
        overlay.className='modal-overlay';
        overlay.onclick=function(e){if(e.target===overlay)overlay.remove();};
        var html='<div class="modal-content"><div class="modal-header"><h3>'+icon('devices')+'设备详情 - '+escH(id)+'</h3><button class="btn btn-sm" onclick="this.closest(\'.modal-overlay\').remove()">'+btnIcon('close')+'</button></div>';
        var cfg=d.config||{};
        html+='<div class="modal-body">';
        html+='<div class="detail-row"><span class="detail-label">设备ID</span><span class="detail-value">'+escH(d.id||'')+'</span></div>';
        html+='<div class="detail-row"><span class="detail-label">协议</span><span class="detail-value">'+escH(d.protocol||'')+'</span></div>';
        html+='<div class="detail-row"><span class="detail-label">手机号</span><span class="detail-value">'+escH(d.phone||'')+'</span></div>';
        html+='<div class="detail-row"><span class="detail-label">车牌号</span><span class="detail-value">'+escH(d.plate||'')+'</span></div>';
        html+='<div class="detail-row"><span class="detail-label">目标地址</span><span class="detail-value">'+escH(d.target||'')+'</span></div>';
        html+='<div class="detail-row"><span class="detail-label">状态</span><span class="detail-value">'+stateBadge(d.state)+'</span></div>';
        if(cfg.manufacturer)html+='<div class="detail-row"><span class="detail-label">制造商</span><span class="detail-value">'+escH(cfg.manufacturer)+'</span></div>';
        if(cfg.terminal_model)html+='<div class="detail-row"><span class="detail-label">终端型号</span><span class="detail-value">'+escH(cfg.terminal_model)+'</span></div>';
        if(cfg.terminal_id)html+='<div class="detail-row"><span class="detail-label">终端ID</span><span class="detail-value">'+escH(cfg.terminal_id)+'</span></div>';
        if(cfg.auth_code)html+='<div class="detail-row"><span class="detail-label">鉴权码</span><span class="detail-value">'+escH(cfg.auth_code)+'</span></div>';
        if(cfg.heartbeat_interval)html+='<div class="detail-row"><span class="detail-label">心跳间隔</span><span class="detail-value">'+cfg.heartbeat_interval+'秒</span></div>';
        if(cfg.location_interval)html+='<div class="detail-row"><span class="detail-label">位置上报间隔</span><span class="detail-value">'+cfg.location_interval+'秒</span></div>';
        html+='</div>';
        html+='<div class="modal-footer"><button class="btn btn-primary" onclick="this.closest(\'.modal-overlay\').remove()">关闭</button></div>';
        overlay.innerHTML=html+'</div>';
        document.body.appendChild(overlay);
    }catch(e){toast('获取详情失败: '+e.message,'error');}
}
async function startDevice(id){
    try{
        var r=await fetch(API+'/devices/'+id+'/start',{method:'POST'});
        var d=await r.json();
        if(d.success){toast(d.message||'设备已启动','success');}
        else{
            if(d.error_type==='connection_failed'){toast('无法连接到目标平台','error');showConnectionHelp(id);}
            else{toast(d.error||'启动失败','error');}
        }
        loadDevices();
    }catch(e){toast(e.message,'error');}
}
async function stopDevice(id){
    try{var r=await fetch(API+'/devices/'+id+'/stop',{method:'POST'});var d=await r.json();if(d.success)toast('设备已停止','success');else toast(d.error||'停止失败','error');loadDevices();}catch(e){toast(e.message,'error');}
}
async function delDevice(id){
    try{var r=await fetch(API+'/devices/'+id,{method:'DELETE'});var d=await r.json();if(d.success)toast('设备已删除','success');else toast(d.error||'删除失败','error');loadDevices();}catch(e){toast(e.message,'error');}
}
function showConnectionHelp(id){
    var overlay=document.createElement('div');
    overlay.className='modal-overlay';
    overlay.onclick=function(e){if(e.target===overlay)overlay.remove();};
    overlay.innerHTML='<div class="modal-content"><div class="modal-header"><h3>'+icon('info')+'连接失败 - 排查指南</h3><button class="btn btn-sm" onclick="this.closest(\'.modal-overlay\').remove()">'+btnIcon('close')+'</button></div><div class="modal-body"><div style="padding:14px;background:var(--danger-bg);border-radius:var(--radius-sm);border-left:3px solid var(--danger);margin-bottom:16px"><p style="color:var(--danger);font-size:13px;margin:0">设备无法连接到目标平台，请按以下步骤排查：</p></div><ol style="color:#ccc;font-size:13px;line-height:2;padding-left:20px"><li><b>检查目标平台是否已启动</b> - 确认目标平台（如 JTE）正在运行并监听对应端口</li><li><b>检查端口是否正确</b> - 默认端口 7611（JT/T 808），请确认与目标平台配置一致</li><li><b>检查防火墙设置</b> - 确保防火墙未阻止该端口</li><li><b>检查网络连通性</b> - 可用 telnet 127.0.0.1 7611 测试</li></ol><p style="color:#8b92a8;font-size:12px;margin-top:12px">设备ID: '+escH(id)+'</p></div><div class="modal-footer"><button class="btn btn-primary" onclick="this.closest(\'.modal-overlay\').remove()">我知道了</button></div></div>';
    document.body.appendChild(overlay);
}
// ==================== 消息监控 ====================
function renderMonitor(){
    document.getElementById('page-monitor').innerHTML=
        '<h2 class="page-title">消息监控</h2>'+
        '<div class="card guide-card"><div style="display:flex;align-items:flex-start;gap:12px"><div style="flex-shrink:0;color:var(--accent)">'+icon('info','card-icon')+'</div><div><div style="color:var(--accent);font-size:13px;font-weight:600;margin-bottom:6px">使用指南</div><div style="color:var(--text-muted);font-size:12px;line-height:1.8">1. 选择要监控的设备，点击「开始监控」<br>2. 系统会实时显示该设备收发的所有报文<br>3. [UP] 表示终端上报消息，[DOWN] 表示平台下发消息<br>4. 点击报文可查看详细解析</div></div></div></div>'+
        '<div class="card"><div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap"><div class="form-group" style="flex:1;min-width:200px"><label>监控设备</label><select id="monitor-device"><option value="">全部设备</option></select></div><button class="btn btn-primary" id="monitor-start-btn" onclick="startMonitor()">'+btnIcon('play')+'开始监控</button><button class="btn btn-danger hidden" id="monitor-stop-btn" onclick="stopMonitor()">'+btnIcon('stop')+'停止监控</button><button class="btn btn-sm" onclick="clearMonitor()">清空</button><span id="monitor-count" style="color:var(--text-muted);font-size:13px;margin-left:auto">共 0 条</span></div>'+
        '<div id="monitor-list" style="margin-top:16px;max-height:600px;overflow-y:auto;font-family:Consolas,monospace;font-size:12px">'+
        '<p style="color:#888;padding:20px;text-align:center">点击「开始监控」查看实时报文</p></div></div>';
    loadMonitorDevices();
}
async function loadMonitorDevices(){
    try{var r=await fetch(API+'/devices'),d=await r.json();var sel=document.getElementById('monitor-device');if(!sel)return;d.devices.forEach(function(dev){var o=document.createElement('option');o.value=dev.id;o.textContent=dev.id+' ('+dev.protocol+')';sel.appendChild(o);});}catch(e){}
}
function startMonitor(){
    if(ws)ws.close();
    var dev=document.getElementById('monitor-device').value;
    var url='ws://'+location.host+'/api/v1/ws/monitor';
    if(dev)url+='?device='+dev;
    ws=new WebSocket(url);
    ws.onopen=function(){toast('监控已启动','success');document.getElementById('monitor-start-btn').classList.add('hidden');document.getElementById('monitor-stop-btn').classList.remove('hidden');monitorMsgs=[];};
    ws.onmessage=function(e){var msg=JSON.parse(e.data);monitorMsgs.push(msg);renderMonitorMsg(msg);var c=document.getElementById('monitor-count');if(c)c.textContent='共 '+monitorMsgs.length+' 条';};
    ws.onclose=function(){toast('监控已停止','info');document.getElementById('monitor-start-btn').classList.remove('hidden');document.getElementById('monitor-stop-btn').classList.add('hidden');};
    ws.onerror=function(){toast('监控连接失败','error');};
}
function stopMonitor(){if(ws){ws.close();ws=null;}}
function clearMonitor(){monitorMsgs=[];var list=document.getElementById('monitor-list');if(list)list.innerHTML='<p style="color:#888;padding:20px;text-align:center">已清空</p>';var c=document.getElementById('monitor-count');if(c)c.textContent='共 0 条';}
function renderMonitorMsg(msg){
    var list=document.getElementById('monitor-list');if(!list)return;
    var dir=msg.direction==='up'?'<span style="color:#4fc3f7">[UP]</span>':'<span style="color:#81c784">[DOWN]</span>';
    var time=new Date(msg.timestamp*1000).toLocaleTimeString();
    var hex=(msg.hex||'').substring(0,80);
    var div=document.createElement('div');
    div.style.cssText='padding:6px 8px;border-bottom:1px solid rgba(255,255,255,0.05);cursor:pointer';
    div.innerHTML=dir+' <span style="color:#888">'+time+'</span> <span style="color:#ccc">'+escH(msg.device_id||'')+'</span> <span style="color:#aaa">'+escH(msg.protocol||'')+'</span> <span style="color:#66bb6a">'+escH(hex)+'</span>';
    list.appendChild(div);
    list.scrollTop=list.scrollHeight;
}

// ==================== 压力测试 ====================
function renderStress(){
    document.getElementById('page-stress').innerHTML=
        '<h2 class="page-title">压力测试</h2>'+
        '<div class="card guide-card"><div style="display:flex;align-items:flex-start;gap:12px"><div style="flex-shrink:0;color:var(--accent)">'+icon('info','card-icon')+'</div><div><div style="color:var(--accent);font-size:13px;font-weight:600;margin-bottom:6px">使用指南</div><div style="color:var(--text-muted);font-size:12px;line-height:1.8">1. 设置设备数量和测试时长<br>2. 填写目标平台地址（默认 127.0.0.1:7611）<br>3. 点击「开始压测」，系统会模拟大量设备同时连接<br>4. 实时查看发送/接收消息数、成功率等指标</div></div></div></div>'+
        '<div class="grid-2"><div class="card"><h3>'+icon('bolt')+'压测配置</h3>'+
        '<div class="form-group"><label>设备数量</label><input type="number" id="stress-count" value="10" min="1" max="1000"></div>'+
        '<div class="form-group"><label>测试时长(秒)</label><input type="number" id="stress-duration" value="60" min="1" max="3600"></div>'+
        '<div class="form-group"><label>目标地址</label><input type="text" id="stress-target" value="127.0.0.1:7611"></div>'+
        '<div class="form-group"><label>协议</label><select id="stress-protocol">'+pO()+'</select></div>'+
        '<button class="btn btn-primary" id="stress-start-btn" onclick="startStress()">'+btnIcon('bolt')+'开始压测</button></div>'+
        '<div class="card"><h3>'+icon('chart')+'压测结果</h3><div id="stress-result"><p style="color:#888">尚未开始压测</p></div></div></div>';
}
async function startStress(){
    var cfg={device_count:parseInt(document.getElementById('stress-count').value)||10,duration:parseInt(document.getElementById('stress-duration').value)||60,target_addr:document.getElementById('stress-target').value,protocol:document.getElementById('stress-protocol').value};
    var btn=document.getElementById('stress-start-btn');
    btn.disabled=true;btn.textContent='压测中...';
    try{
        var r=await fetch(API+'/stress/run',{method:'POST',headers:{'Content-Type':'application/json; charset=utf-8'},body:JSON.stringify(cfg)});
        var d=await r.json();
        if(d.success){
            toast(d.message||'压测已启动','success');
            stressTimer=setInterval(updateStressStatus,2000);
        }else{toast(d.error||'压测失败','error');btn.disabled=false;btn.innerHTML=btnIcon('bolt')+'开始压测';}
    }catch(e){toast(e.message,'error');btn.disabled=false;btn.innerHTML=btnIcon('bolt')+'开始压测';}
}
async function updateStressStatus(){
    try{
        var r=await fetch(API+'/stress/status'),d=await r.json();
        var el=document.getElementById('stress-result');if(!el)return;
        if(d.running){
            el.innerHTML='<div style="text-align:center;padding:20px"><span class="dot" style="background:#4caf50;animation:pulse 1.5s infinite;display:inline-block;margin-right:8px"></span><span style="color:#4caf50">压测运行中...</span></div>';
        }else{
            if(stressTimer){clearInterval(stressTimer);stressTimer=null;}
            var btn=document.getElementById('stress-start-btn');
            if(btn){btn.disabled=false;btn.innerHTML=btnIcon('bolt')+'开始压测';}
            var res=d.result||{};
            el.innerHTML='<div class="grid-2">'+
                '<div class="stat-item"><div class="stat-label">总设备数</div><div class="stat-value">'+(res.total_devices||0)+'</div></div>'+
                '<div class="stat-item"><div class="stat-label">在线设备</div><div class="stat-value">'+(res.online_devices||0)+'</div></div>'+
                '<div class="stat-item"><div class="stat-label">总消息数</div><div class="stat-value">'+(res.total_messages||0)+'</div></div>'+
                '<div class="stat-item"><div class="stat-label">错误数</div><div class="stat-value" style="color:#ef5350">'+(res.error_count||0)+'</div></div>'+
                '</div>';
            if(!d.running)toast('压测完成','success');
        }
    }catch(e){}
}

// ==================== 场景模拟 ====================
function renderScenario(){
    document.getElementById('page-scenario').innerHTML=
        '<h2 class="page-title">场景模拟</h2>'+
        '<div class="card guide-card"><div style="display:flex;align-items:flex-start;gap:12px"><div style="flex-shrink:0;color:var(--accent)">'+icon('info','card-icon')+'</div><div><div style="color:var(--accent);font-size:13px;font-weight:600;margin-bottom:6px">使用指南</div><div style="color:var(--text-muted);font-size:12px;line-height:1.8">1. 选择预设场景（如车辆上线、持续上报、报警等）<br>2. 配置设备参数和目标平台地址<br>3. 点击「开始模拟」，系统会按场景脚本自动执行<br>4. 可在「消息监控」页面查看实时报文交互</div></div></div></div>'+
        '<div class="card guide-card"><div style="display:flex;align-items:flex-start;gap:12px"><div style="flex-shrink:0;color:var(--warning)">'+icon('link','card-icon')+'</div><div><div style="color:var(--warning);font-size:13px;font-weight:600;margin-bottom:6px">外部平台联调</div><div style="color:var(--text-muted);font-size:12px;line-height:1.8">如需与外部平台（如 JTE）联调测试：<br>1. 启动目标平台（如 <code style="color:var(--accent)">E:\'硕腾网络\'JTE</code>）<br>2. 确认平台监听端口（默认 7611）<br>3. 在设备管理中创建设备，目标地址填写平台地址<br>4. 启动设备后，仿真终端会自动注册并开始上报</div></div></div></div>'+
        '<div class="grid-2"><div class="card"><h3>'+icon('scenario')+'场景选择</h3>'+
        '<div class="form-group"><label>预设场景</label><select id="scenario-name"><option value="vehicle_online">终端上线流程</option><option value="overspeed_alarm">超速报警</option><option value="fatigue_drive">疲劳驾驶</option><option value="emergency_alarm">紧急报警</option><option value="offline_reconnect">离线重连</option></select></div>'+
        '<div class="form-group"><label>设备数量</label><input type="number" id="scenario-count" value="1" min="1" max="100"></div>'+
        '<div class="form-group"><label>目标地址</label><input type="text" id="scenario-target" value="127.0.0.1:7611"></div>'+
        '<div class="form-group"><label>手机号前缀</label><input type="text" id="scenario-phone-prefix" value="01380000"></div>'+
        '<button class="btn btn-primary" id="scenario-start-btn" onclick="runScenario()">'+btnIcon('play')+'开始模拟</button>'+
        '<button class="btn btn-danger hidden" id="scenario-stop-btn" onclick="stopScenario()">'+btnIcon('stop')+'停止模拟</button></div>'+
        '<div class="card"><h3>'+icon('chart')+'场景状态</h3><div id="scenario-status"><p style="color:#888">未开始</p></div></div></div>';
}
async function runScenario(){
    var cfg={name:document.getElementById('scenario-name').value,target:document.getElementById('scenario-target').value};
    try{
        var body=JSON.stringify(cfg);
        var r=await fetch(API+'/scenarios/run',{method:'POST',headers:{'Content-Type':'application/json; charset=utf-8'},body:body});
        var d=await r.json();
        if(d.success){
            toast('场景已启动','success');
            document.getElementById('scenario-start-btn').classList.add('hidden');
            document.getElementById('scenario-stop-btn').classList.remove('hidden');
            scenarioTimer=setInterval(updateScenarioStatus,2000);
        }else{toast(d.error||'启动失败','error');}
    }catch(e){toast(e.message,'error');}
}
async function stopScenario(){
    try{var r=await fetch(API+'/scenarios/stop',{method:'POST'});var d=await r.json();if(d.success){toast('场景已停止','success');}else{toast(d.error||'停止失败','error');}}catch(e){toast(e.message,'error');}
    document.getElementById('scenario-start-btn').classList.remove('hidden');
    document.getElementById('scenario-stop-btn').classList.add('hidden');
    if(scenarioTimer){clearInterval(scenarioTimer);scenarioTimer=null;}
}
async function updateScenarioStatus(){
    try{
        var r=await fetch(API+'/scenarios/status'),d=await r.json();
        var el=document.getElementById('scenario-status');if(!el)return;
        if(d.running){
            el.innerHTML='<div style="display:flex;align-items:center;gap:8px;margin-bottom:12px"><span class="dot" style="background:#4caf50;animation:pulse 1.5s infinite"></span><span style="color:#4caf50;font-weight:600">运行中</span><span style="color:#888">'+escH(d.scenario_name||'')+'</span></div>';
            if(d.progress_percent!==undefined){el.innerHTML+='<div style="background:rgba(255,255,255,0.05);border-radius:4px;height:8px;overflow:hidden;margin-bottom:8px"><div style="background:var(--accent);height:100%;width:'+d.progress_percent+'%;transition:width 0.3s"></div></div>';}
            if(d.current_step!==undefined&&d.total_steps!==undefined){el.innerHTML+='<p style="color:#888;font-size:13px">步骤: '+d.current_step+' / '+d.total_steps+'</p>';}
        }else{
            el.innerHTML='<p style="color:#888">未运行</p>';
            document.getElementById('scenario-start-btn').classList.remove('hidden');
            document.getElementById('scenario-stop-btn').classList.add('hidden');
            if(scenarioTimer){clearInterval(scenarioTimer);scenarioTimer=null;}
        }
    }catch(e){}
}

// ==================== 合规检查 ====================
function renderChecker(){
    document.getElementById('page-checker').innerHTML=
        '<h2 class="page-title">合规检查</h2>'+
        '<div class="card guide-card"><div style="display:flex;align-items:flex-start;gap:12px"><div style="flex-shrink:0;color:var(--accent)">'+icon('info','card-icon')+'</div><div><div style="color:var(--accent);font-size:13px;font-weight:600;margin-bottom:6px">使用指南</div><div style="color:var(--text-muted);font-size:12px;line-height:1.8">1. 填写或粘贴一条 JT/T 808 报文（Hex格式）<br>2. 点击「开始检查」，系统会验证报文是否符合协议规范<br>3. 查看检查结果，包括通过/失败项和详细错误信息<br>4. 可在「报文工坊」生成测试报文后复制到此检查</div></div></div></div>'+
        '<div class="card"><h3>'+icon('check')+'合规检查</h3>'+
        '<div class="form-group"><label>JT/T 808 报文 (Hex)</label><textarea id="check-hex" rows="4" placeholder="7E 02 00 00 1C ..." style="font-family:Consolas,monospace">7E0200001C013800001234000100000000000000000260F7B406F015581388003C00B4260923142409BC7E</textarea></div>'+
        '<button class="btn btn-primary" id="check-start-btn" onclick="runCheck()">'+btnIcon('check')+'开始检查</button></div>'+
        '<div id="check-result" style="margin-top:16px"></div>'+
        '<div class="card" style="margin-top:16px"><h3>'+icon('bolt')+'Fuzz 健壮性测试</h3>'+
        '<p style="color:var(--text-muted);font-size:12px;line-height:1.6;margin-bottom:12px">自动生成各种异常报文（空帧、超长帧、非法字段等）测试协议解析器的健壮性</p>'+
        '<button class="btn btn-primary" id="fuzz-start-btn" onclick="runFuzz()">'+btnIcon('bolt')+'开始 Fuzz 测试</button>'+
        '<div id="fuzz-result" style="margin-top:16px"></div></div>';
}
async function runCheck(){
    var hex=document.getElementById('check-hex').value.trim();
    if(!hex){toast('请输入报文Hex','error');return;}
    var cfg={protocol:'jt808',hex:hex};
    var btn=document.getElementById('check-start-btn');btn.disabled=true;btn.textContent='检查中...';
    try{
        var r=await fetch(API+'/check/compliance',{method:'POST',headers:{'Content-Type':'application/json; charset=utf-8'},body:JSON.stringify(cfg)});
        var d=await r.json();
        var el=document.getElementById('check-result');
        if(d.success&&d.result){
            var res=d.result;var h='<div class="card"><h3>检查结果</h3>';
            h+='<div class="grid-3">';
            h+='<div class="stat-item"><div class="stat-label">总检查项</div><div class="stat-value">'+(res.total||0)+'</div></div>';
            h+='<div class="stat-item"><div class="stat-label">通过</div><div class="stat-value" style="color:#4caf50">'+(res.passed||0)+'</div></div>';
            h+='<div class="stat-item"><div class="stat-label">失败</div><div class="stat-value" style="color:#ef5350">'+(res.failed||0)+'</div></div>';
            h+='</div>';
            if(res.items&&res.items.length>0){
                h+='<table style="margin-top:16px"><thead><tr><th>消息</th><th>结果</th><th>详情</th></tr></thead><tbody>';
                res.items.forEach(function(item){h+='<tr><td>'+escH(item.name||'')+'</td><td>'+(item.passed?'<span style="color:#4caf50">通过</span>':'<span style="color:#ef5350">失败</span>')+'</td><td style="font-size:12px;color:#888">'+escH(item.message||'')+'</td></tr>';});
                h+='</tbody></table>';
            }
            h+='</div>';el.innerHTML=h;
            toast('检查完成','success');
        }else{el.innerHTML='<div class="card"><p style="color:#ef5350">'+escH(d.error||'检查失败')+'</p></div>';toast(d.error||'检查失败','error');}
    }catch(e){toast(e.message,'error');}
    btn.disabled=false;btn.innerHTML=btnIcon('check')+'开始检查';
}

// ==================== Fuzz测试 ====================
async function runFuzz(){
    var btn=document.getElementById('fuzz-start-btn');if(btn){btn.disabled=true;btn.textContent='Fuzz测试中...';}
    try{
        var r=await fetch(API+'/check/fuzz',{method:'POST',headers:{'Content-Type':'application/json; charset=utf-8'}});
        var d=await r.json();
        var el=document.getElementById('fuzz-result');if(!el)return;
        if(d.success&&d.result){
            var res=d.result;
            el.innerHTML='<div class="grid-4">'+
                '<div class="stat-item"><div class="stat-label">总用例</div><div class="stat-value">'+(res.total_cases||0)+'</div></div>'+
                '<div class="stat-item"><div class="stat-label">成功</div><div class="stat-value" style="color:#4caf50">'+(res.success_cases||0)+'</div></div>'+
                '<div class="stat-item"><div class="stat-label">崩溃</div><div class="stat-value" style="color:#ef5350">'+(res.crash_cases||0)+'</div></div>'+
                '<div class="stat-item"><div class="stat-label">错误</div><div class="stat-value" style="color:#ffa726">'+(res.error_cases||0)+'</div></div>'+
                '</div>';
            if(res.cases&&res.cases.length>0){
                var h='<div style="margin-top:16px;max-height:400px;overflow-y:auto">';
                res.cases.forEach(function(c){
                    var color=c.result==='success'?'#4caf50':(c.result==='crash'?'#ef5350':'#ffa726');
                    h+='<div style="padding:8px;border-bottom:1px solid rgba(255,255,255,0.05);font-size:12px"><span style="color:'+color+'">['+c.result.toUpperCase()+']</span> <span style="color:#888">'+escH(c.name||'')+'</span> <span style="color:#66bb6a;font-family:monospace">'+escH(c.hex||'')+'</span>';
                    if(c.error)h+=' <span style="color:#ef5350">'+escH(c.error)+'</span>';
                    h+='</div>';
                });
                el.innerHTML+=h+'</div>';
            }
            toast('Fuzz测试完成','success');
        }else{el.innerHTML='<p style="color:#ef5350">'+escH(d.error||'测试失败')+'</p>';toast(d.error||'测试失败','error');}
    }catch(e){toast(e.message,'error');}
    if(btn){btn.disabled=false;btn.innerHTML=btnIcon('bolt')+'开始 Fuzz 测试';}
}

// ==================== 设置 ====================
function renderSettings(){
    document.getElementById('page-settings').innerHTML=
        '<h2 class="page-title">设置</h2>'+
        '<div class="card"><h3>'+icon('settings')+'系统设置</h3>'+
        '<div class="form-group"><label>默认目标地址</label><input type="text" id="settings-default-target" value="127.0.0.1:7611"></div>'+
        '<div class="form-group"><label>默认上报间隔(秒)</label><input type="number" id="settings-default-interval" value="30" min="1"></div>'+
        '<div class="form-group"><label>默认心跳间隔(秒)</label><input type="number" id="settings-default-heartbeat" value="60" min="1"></div>'+
        '<button class="btn btn-primary" onclick="saveSettings()">'+btnIcon('save')+'保存设置</button></div>'+
        '<div class="card"><h3>'+icon('info')+'关于</h3><p style="color:#888;font-size:13px;line-height:1.8">JT-Simulate v1.0 - 部标协议仿真平台<br>支持协议：JT/T 808, JT/T 809, JT/T 1078, JT/T 905, JT/T 1045, JT/T 1253, GB/T 32960<br>开发者：硕腾网络</p></div>';
}
function saveSettings(){toast('设置已保存','success');}