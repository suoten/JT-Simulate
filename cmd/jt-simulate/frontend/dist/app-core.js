// JT-Simulate Frontend JS v7 - Core
var API='/api/v1',ws=null,monitorMsgs=[],dashTimer=null,devTimer=null,stressTimer=null,scenarioTimer=null;
function copyToClipboard(text){var ta=document.createElement('textarea');ta.value=text;ta.style.position='fixed';ta.style.opacity='0';document.body.appendChild(ta);ta.select();try{document.execCommand('copy');toast('已复制到剪贴板','success');}catch(e){toast('复制失败','error');}document.body.removeChild(ta);}
function copyEncodeOutput(){var el=document.getElementById('encode-output');if(el)copyToClipboard(el.textContent);}

function toast(msg,type){
    var t=document.createElement('div');
    t.className='toast toast-'+(type||'info');
    t.innerHTML='<span>'+(type==='success'?'✓':type==='error'?'✕':'ℹ')+'</span> '+msg;
    document.body.appendChild(t);
    setTimeout(function(){t.classList.add('hide');setTimeout(function(){t.remove();},300);},3000);
}

var ICO={
    dashboard:'<svg viewBox="0 0 24 24"><rect x="3" y="3" width="7" height="9" rx="1"/><rect x="14" y="3" width="7" height="5" rx="1"/><rect x="14" y="12" width="7" height="9" rx="1"/><rect x="3" y="16" width="7" height="5" rx="1"/></svg>',
    workshop:'<svg viewBox="0 0 24 24"><path d="M14.7 6.3a4 4 0 0 1-5.4 5.4L4 17v3h3l5.3-5.3a4 4 0 0 1 5.4-5.4l-3 3L13 10.7l1.7-4.4z"/></svg>',
    devices:'<svg viewBox="0 0 24 24"><rect x="5" y="2" width="14" height="20" rx="2"/><line x1="12" y1="18" x2="12" y2="18.01"/></svg>',
    monitor:'<svg viewBox="0 0 24 24"><path d="M2 12h2m4 0h2m4 0h2m4 0h2"/><path d="M6 8a4 4 0 0 1 0 8m12-8a4 4 0 0 0 0 8m-9-9a8 8 0 0 1 6 0m-6 10a8 8 0 0 0 6 0"/></svg>',
    stress:'<svg viewBox="0 0 24 24"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>',
    scenario:'<svg viewBox="0 0 24 24"><path d="M4 4h16v16H4z"/><path d="M4 9h16M9 4v16"/></svg>',
    checker:'<svg viewBox="0 0 24 24"><path d="M9 11l2 2 4-4"/><circle cx="12" cy="12" r="9"/></svg>',
    settings:'<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"/><path d="M12 2v3m0 14v3M2 12h3m14 0h3M5 5l2 2m10 10l2 2M5 19l2-2m10-10l2-2"/></svg>',
    play:'<svg viewBox="0 0 24 24"><polygon points="5,3 19,12 5,21"/></svg>',
    stop:'<svg viewBox="0 0 24 24"><rect x="6" y="6" width="12" height="12" rx="1"/></svg>',
    plus:'<svg viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>',
    trash:'<svg viewBox="0 0 24 24"><path d="M3 6h18M8 6V4a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1v2m2 0v14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V6"/></svg>',
    send:'<svg viewBox="0 0 24 24"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/></svg>',
    search:'<svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="7"/><path d="M21 21l-4.3-4.3"/></svg>',
    shield:'<svg viewBox="0 0 24 24"><path d="M12 2l9 4v6c0 5-3.5 8-9 10-5.5-2-9-5-9-10V6l9-4z"/></svg>',
    info:'<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="9"/><path d="M12 8v.01M11 12h1v4h1"/></svg>',
    cube:'<svg viewBox="0 0 24 24"><path d="M12 2l9 5v10l-9 5-9-5V7l9-5z"/><path d="M12 2v10m0 0l9-5m-9 5l-9-5"/></svg>',
    activity:'<svg viewBox="0 0 24 24"><path d="M3 12h4l2-7 4 14 2-7h6"/></svg>',
    server:'<svg viewBox="0 0 24 24"><rect x="3" y="4" width="18" height="6" rx="1"/><rect x="3" y="14" width="18" height="6" rx="1"/><circle cx="7" cy="7" r="0.5" fill="currentColor"/><circle cx="7" cy="17" r="0.5" fill="currentColor"/></svg>',
    network:'<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18"/></svg>',
    message:'<svg viewBox="0 0 24 24"><path d="M3 5h18v12H8l-5 4V5z"/></svg>',
    empty:'<svg viewBox="0 0 24 24"><path d="M12 2l9 5v10l-9 5-9-5V7l9-5z"/><path d="M12 7v5"/><circle cx="12" cy="15" r="0.5"/></svg>',
    refresh:'<svg viewBox="0 0 24 24"><path d="M3 12a9 9 0 0 1 15-6.7L21 8M21 8V3M21 8h-5M21 12a9 9 0 0 1-15 6.7L3 16M3 16v5M3 16h5"/></svg>',
    eye:'<svg viewBox="0 0 24 24"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>',
    edit:'<svg viewBox="0 0 24 24"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>',
    close:'<svg viewBox="0 0 24 24"><path d="M18 6L6 18M6 6l12 12"/></svg>',
    copy:'<svg viewBox="0 0 24 24"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>',
    bolt:'<svg viewBox="0 0 24 24"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>',
    chart:'<svg viewBox="0 0 24 24"><path d="M3 3v18h18M7 16l4-4 4 2 5-7"/></svg>',
    check:'<svg viewBox="0 0 24 24"><path d="M9 11l2 2 4-4"/><circle cx="12" cy="12" r="9"/></svg>',
    warning:'<svg viewBox="0 0 24 24"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><path d="M12 9v4M12 17h.01"/></svg>',
    link:'<svg viewBox="0 0 24 24"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>',
    save:'<svg viewBox="0 0 24 24"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><path d="M17 21v-8H7v8M7 3v5h8"/></svg>',
};
function icon(name,cls){return '<span class="'+(cls||'card-icon')+'">'+(ICO[name]||ICO.empty)+'</span>';}
function btnIcon(name){return '<span class="btn-icon">'+(ICO[name]||ICO.empty)+'</span>';}

function FI(id,l,v){return '<div class="form-group"><label>'+l+'</label><input type="text" id="f-'+id+'" value="'+(v||'')+'"></div>';}
function NI(id,l,v){return '<div class="form-group"><label>'+l+'</label><input type="number" id="f-'+id+'" value="'+(v||'0')+'"></div>';}
function pO(){return PROTOCOLS.map(function(p){return '<option value="'+p.id+'">'+p.name+'</option>';}).join('');}
function escH(s){return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');}
function stateBadge(state){
    var colors={online:'badge-green',connecting:'badge-yellow',offline:'badge-gray',closing:'badge-gray',unknown:'badge-gray'};
    var labels={online:'在线',connecting:'连接中',offline:'离线',closing:'关闭中',unknown:'未知'};
    var c=colors[state]||'badge-gray';
    var l=labels[state]||state||'未知';
    var dot=state==='online'?'<span class="state-dot dot-online"></span>':state==='connecting'?'<span class="state-dot dot-connecting"></span>':'<span class="state-dot dot-offline"></span>';
    return dot+'<span class="badge '+c+'">'+l+'</span>';
}
function locF(){return '<div class="grid-2">'+NI('latitude','纬度','39.9093')+NI('longitude','经度','116.3974')+NI('speed','速度(km/h)','60')+NI('direction','方向(0-359)','180')+'</div>';}
function vehF(){return NI('vehicle_color','车牌颜色(0-9)','1')+FI('vehicle_plate','车牌号','京A12345');}

// 导航
document.querySelectorAll('.nav-item').forEach(function(el){el.addEventListener('click',function(){showPage(this.dataset.page,this);});});
function showPage(n,el){
    if(dashTimer){clearInterval(dashTimer);dashTimer=null;}
    if(devTimer){clearInterval(devTimer);devTimer=null;}
    if(stressTimer){clearInterval(stressTimer);stressTimer=null;}
    if(typeof scenarioTimer!=='undefined'&&scenarioTimer){clearInterval(scenarioTimer);scenarioTimer=null;}
    if(n!=='monitor'&&ws){ws.close();ws=null;}
    document.querySelectorAll('.page').forEach(function(e){e.classList.add('hidden');});
    document.getElementById('page-'+n).classList.remove('hidden');
    document.querySelectorAll('.nav-item').forEach(function(e){e.classList.remove('active');});
    if(el)el.classList.add('active');
    var fn={dashboard:renderDashboard,workshop:renderWorkshop,devices:renderDevices,monitor:renderMonitor,stress:renderStress,scenario:renderScenario,checker:renderChecker,settings:renderSettings};
    if(fn[n])fn[n]();
}

// 仪表盘
function renderDashboard(){
    var b=PROTOCOLS.map(function(p){var vs=p.vs.map(function(v){return '<span class="badge badge-blue" style="margin:2px">'+v.l+'</span>';}).join('');return '<div style="margin-bottom:8px"><span style="color:#4fc3f7;margin-right:8px">'+p.name+':</span>'+vs+'</div>';}).join('');
    document.getElementById('page-dashboard').innerHTML=
        '<h2 class="page-title">仪表盘</h2>'+
        '<div class="stat-grid">'+
            '<div class="stat-card"><div class="stat-icon-bar" style="color:#4fc3f7;margin-bottom:8px">'+icon('devices','card-icon')+'</div><div class="stat-value" id="stat-total">-</div><div class="stat-label">设备总数</div></div>'+
            '<div class="stat-card"><div class="stat-icon-bar" style="color:#66bb6a;margin-bottom:8px">'+icon('server','card-icon')+'</div><div class="stat-value" id="stat-online" style="color:#66bb6a">-</div><div class="stat-label">在线设备</div></div>'+
            '<div class="stat-card"><div class="stat-icon-bar" style="color:#64b5f6;margin-bottom:8px">'+icon('message','card-icon')+'</div><div class="stat-value" id="stat-messages" style="color:#64b5f6">-</div><div class="stat-label">消息总数</div></div>'+
            '<div class="stat-card"><div class="stat-icon-bar" style="color:#66bb6a;margin-bottom:8px">'+icon('activity','card-icon')+'</div><div class="stat-value" id="stat-status" style="color:#66bb6a">运行中</div><div class="stat-label">系统状态</div></div>'+
        '</div>'+
        '<div class="card guide-card"><div style="display:flex;align-items:flex-start;gap:12px"><div style="flex-shrink:0;color:var(--accent)">'+icon('info','card-icon')+'</div><div><div style="color:var(--accent);font-size:13px;font-weight:600;margin-bottom:8px">快速入门</div><div style="color:var(--text-muted);font-size:12px;line-height:2">'+
        '<div style="margin-bottom:6px"><strong style="color:var(--accent)">第1步：</strong>到「报文工坊」构造和验证报文编解码</div>'+
        '<div style="margin-bottom:6px"><strong style="color:var(--accent)">第2步：</strong>到「设备管理」创建仿真设备，填入目标平台地址（如 <code style="color:var(--accent)">127.0.0.1:7611</code>）</div>'+
        '<div style="margin-bottom:6px"><strong style="color:var(--accent)">第3步：</strong>启动设备后到「消息监控」查看实时报文交互</div>'+
        '<div style="margin-bottom:6px"><strong style="color:var(--accent)">第4步：</strong>到「场景模拟」运行预置场景测试完整业务流程</div>'+
        '<div><strong style="color:var(--accent)">第5步：</strong>到「合规检查」验证报文是否符合协议规范</div>'+
        '</div></div></div></div>'+
        '<div class="card"><h3>'+icon('cube')+'支持协议与版本</h3>'+b+'</div>'+
        '<div class="card"><h3>'+icon('activity')+'快速操作</h3>'+
            '<button class="btn btn-primary" onclick="showPage(\'workshop\',document.querySelectorAll(\'.nav-item\')[1])">'+btnIcon('send')+'构造报文</button> '+
            '<button class="btn btn-primary" onclick="showPage(\'devices\',document.querySelectorAll(\'.nav-item\')[2])">'+btnIcon('plus')+'创建设备</button> '+
            '<button class="btn btn-primary" onclick="showPage(\'stress\',document.querySelectorAll(\'.nav-item\')[4])">'+btnIcon('stress')+'压力测试</button> '+
            '<button class="btn btn-primary" onclick="showPage(\'checker\',document.querySelectorAll(\'.nav-item\')[6])">'+btnIcon('shield')+'合规检查</button>'+
        '</div>';
    loadStats();
    dashTimer=setInterval(loadStats,3000);
}
async function loadStats(){
    try{
        var r=await fetch(API+'/stats');var d=await r.json();var el;
        if((el=document.getElementById('stat-total')))el.textContent=d.total_devices||0;
        if((el=document.getElementById('stat-online'))){var online=d.online_devices||0;el.textContent=online;el.style.color=online>0?'#66bb6a':'#888';}
        if((el=document.getElementById('stat-messages')))el.textContent=d.total_messages||0;
    }catch(e){console.error(e);}
}

// 报文工坊
function renderWorkshop(){
    document.getElementById('page-workshop').innerHTML=
        '<h2 class="page-title">报文工坊</h2>'+
        '<div class="card guide-card"><div style="display:flex;align-items:flex-start;gap:12px"><div style="flex-shrink:0;color:var(--accent)">'+icon('info','card-icon')+'</div><div><div style="color:var(--accent);font-size:13px;font-weight:600;margin-bottom:6px">使用指南</div><div style="color:var(--text-muted);font-size:12px;line-height:1.8">1. 选择协议类型和版本（如 JT/T 808-2019）<br>2. 选择消息ID（如 0x0200 位置上报）<br>3. 填写字段参数，点击「生成报文」<br>4. 生成的报文会自动填入右侧分析框，点击「分析报文」验证编解码一致性<br>5. 可将报文复制到其他工具或发送给目标平台测试</div></div></div></div>'+
        '<div class="grid-2"><div class="card"><h3>'+icon('workshop')+'报文构造</h3><div class="form-group"><label>协议类型</label><select id="encode-protocol" onchange="onPC()">'+pO()+'</select></div><div class="form-group"><label>协议版本</label><select id="encode-version" onchange="updateMID()"></select></div><div class="form-group"><label>消息ID</label><select id="encode-msgid" onchange="updateF()"></select></div><div class="form-group"><label>终端手机号/VIN</label><input type="text" id="encode-phone" value="013800001234"></div><div id="encode-fields"></div><button class="btn btn-primary" onclick="encodeMsg()">'+btnIcon('send')+'生成报文</button><div style="margin-top:12px"><div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:4px"><label style="font-size:12px;color:#aaa;display:block;margin:0">报文Hex</label><button class="btn btn-sm" id="encode-copy-btn" onclick="copyEncodeOutput()" style="display:none">'+btnIcon('copy')+'复制</button></div><div class="hex-output" id="encode-output">点击生成报文</div></div></div><div class="card"><h3>'+icon('search')+'报文分析</h3><div class="form-group"><label>协议</label><select id="analyze-protocol">'+pO()+'</select></div><div class="form-group"><label>Hex报文</label><textarea id="analyze-hex" rows="5" placeholder="7E 02 00 00 1C ..." style="font-family:Consolas,monospace"></textarea></div><button class="btn btn-primary" onclick="analyzeMsg()">'+btnIcon('search')+'分析报文</button><div id="analyze-result" style="margin-top:12px"></div></div></div>';
    onPC();
}
function onPC(){var p=document.getElementById('encode-protocol').value;var pr=PROTOCOLS.find(function(x){return x.id===p;});var vs=document.getElementById('encode-version');vs.innerHTML='';(pr?pr.vs:[]).forEach(function(v){var o=document.createElement('option');o.value=v.v;o.textContent=v.l;vs.appendChild(o);});updateMID();}
function updateMID(){var p=document.getElementById('encode-protocol').value;var v=document.getElementById('encode-version').value;var s=document.getElementById('encode-msgid');s.innerHTML='';fM(p,v).forEach(function(m){var o=document.createElement('option');o.value=m.i;o.textContent=m.i+' '+m.n+' ['+m.d+']';s.appendChild(o);});updateF();}
function updateF(){
    var p=document.getElementById('encode-protocol').value,mid=document.getElementById('encode-msgid').value,c=document.getElementById('encode-fields'),h='';
    if(p==='jt808')h=f808(mid);else if(p==='jt809')h=f809(mid);else if(p==='jt1078')h=f1078(mid);else if(p==='gbt32960')h=f32960(mid);else if(p==='jt905')h=f905(mid);else if(p==='jt1045')h=f1045(mid);else if(p==='jt1253')h=f1253(mid);
    if(!h)h='<p style="color:#888;font-size:13px">该消息使用默认参数构造</p>';
    c.innerHTML=h;
}

async function encodeMsg(){
    var p=document.getElementById('encode-protocol').value,v=document.getElementById('encode-version').value,mid=document.getElementById('encode-msgid').value,phone=document.getElementById('encode-phone').value,out=document.getElementById('encode-output');
    out.style.color='#66bb6a';out.textContent='生成中...';
    var fields={};
    document.querySelectorAll('#encode-fields input[id^="f-"]').forEach(function(el){var key=el.id.substring(2);fields[key]=el.type==='number'?(parseFloat(el.value)||0):el.value;});
    try{
        var r=await fetch(API+'/workshop/encode',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({protocol:p,version:v,msg_id:mid,phone:phone,fields:fields})});
        var d=await r.json();
        if(d.success){out.textContent=d.hex;var ah=document.getElementById('analyze-hex'),ap=document.getElementById('analyze-protocol'),cb=document.getElementById('encode-copy-btn');if(ah)ah.value=d.hex;if(ap)ap.value=p;if(cb)cb.style.display='inline-flex';}
        else{out.style.color='#ef5350';out.textContent='错误: '+(d.error||'未知错误');}
    }catch(e){out.style.color='#ef5350';out.textContent='请求失败: '+e.message;}
}
async function analyzeMsg(){
    var p=document.getElementById('analyze-protocol').value,hex=document.getElementById('analyze-hex').value,result=document.getElementById('analyze-result');
    result.innerHTML='<p style="color:#888">分析中...</p>';
    try{
        var r=await fetch(API+'/workshop/analyze',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({protocol:p,hex:hex})});
        var d=await r.json();
        if(d.success){
            var iss='';
            if(d.issues&&d.issues.length>0){iss='<div style="margin-top:8px">';d.issues.forEach(function(i){var c=i.level==='error'?'#ef5350':'#ff9800';iss+='<div style="color:'+c+';font-size:13px;margin:2px 0">['+i.level.toUpperCase()+'] '+i.field+': '+i.message+'</div>';});iss+='</div>';}
            result.innerHTML='<div style="margin-bottom:8px"><span style="color:#888;font-size:13px">消息ID: </span><span style="color:#4fc3f7">'+d.msg_id+'</span> <span style="color:#888;font-size:13px;margin-left:12px">名称: </span><span style="color:#66bb6a">'+d.msg_name+'</span></div><div style="margin-bottom:8px"><span style="color:#888;font-size:13px">终端: </span><span>'+d.phone+'</span> <span style="color:#888;font-size:13px;margin-left:12px">流水号: </span><span>'+d.seq_num+'</span></div>'+iss+'<div style="margin-top:12px"><label style="font-size:12px;color:#aaa;display:block;margin-bottom:4px">解析结果</label><pre class="result-pre">'+escH(JSON.stringify(d.parsed,null,2))+'</pre></div>';
        }else{result.innerHTML='<div style="color:#ef5350">分析失败: '+(d.error||'未知错误')+'</div>';}
    }catch(e){result.innerHTML='<div style="color:#ef5350">请求失败: '+e.message+'</div>';}
}

// 设置
function renderSettings(){
    document.getElementById('page-settings').innerHTML='<h2 class="page-title">设置</h2><div class="card"><h3>'+icon('info')+'系统信息</h3><div class="info-row"><span class="info-label">版本</span><span class="info-value">JT-Simulate v1.0</span></div><div class="info-row"><span class="info-label">支持协议</span><span class="info-value">JT/T 808, JT/T 809, JT/T 1078, JT/T 905, JT/T 1045, JT/T 1253, GB/T 32960</span></div><div class="info-row"><span class="info-label">协议版本</span><span class="info-value">808(2011/2013/2019), 809(2011/2019), 1078(2016/2022), 905(2014), 1045(2018), 1253(2019), 32960(2016)</span></div></div><div class="card"><h3>'+icon('network')+'API信息</h3><div class="info-row"><span class="info-label">Base URL</span><span class="info-value"><code>'+API+'</code></span></div><div class="info-row"><span class="info-label">编码</span><span class="info-value"><code>POST /workshop/encode</code></span></div><div class="info-row"><span class="info-label">分析</span><span class="info-value"><code>POST /workshop/analyze</code></span></div><div class="info-row"><span class="info-label">设备</span><span class="info-value"><code>GET/POST /devices</code></span></div><div class="info-row"><span class="info-label">合规</span><span class="info-value"><code>POST /check/compliance</code></span></div><div class="info-row"><span class="info-label">压测</span><span class="info-value"><code>POST /stress/run</code></span></div></div>';
}

// 初始化
renderDashboard();
