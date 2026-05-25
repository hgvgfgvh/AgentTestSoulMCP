const API = '/console/api';

async function loadStats() {
  const r = await fetch(API + '/stats');
  const j = await r.json();
  document.getElementById('stats').textContent =
    `data=${j.data_dir} | 日文件=${(j.history_days || []).join(', ')} | person=${j.person_bytes}B map=${j.map_bytes}B cache=${j.cache_bytes}B`;
}

function appendMsg(role, text) {
  const el = document.createElement('div');
  el.className = 'msg ' + role;
  el.textContent = text;
  document.getElementById('messages').appendChild(el);
  el.scrollIntoView({ behavior: 'smooth' });
}

function showOutput(obj) {
  document.getElementById('output').textContent = JSON.stringify(obj, null, 2);
}

async function post(path, body) {
  const r = await fetch(API + path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  return r.json();
}

document.getElementById('btnRetrieve').onclick = async () => {
  const user = document.getElementById('userInput').value.trim();
  if (!user) return alert('请输入用户话');
  appendMsg('user', user);
  appendMsg('system', '调用 soul_retrieve…');
  const j = await post('/retrieve', { query: user });
  showOutput(j);
  const hints = j.hints || j.raw || JSON.stringify(j);
  appendMsg('system', 'retrieve 完成，见右侧面板');
};

document.getElementById('btnStore').onclick = async () => {
  const user = document.getElementById('userInput').value.trim();
  const assistant = document.getElementById('assistantInput').value.trim();
  if (!user || !assistant) return alert('store 需要用户话和助手回复');
  appendMsg('system', '调用 soul_store（等待异步）…');
  const j = await post('/store', {
    user_input: user,
    assistant_reply: assistant,
    turn_id: 'console-' + Date.now(),
    wait_async: true,
  });
  showOutput(j);
  await loadStats();
  appendMsg('system', 'store 已 ACK，异步任务已等待结束');
};

document.getElementById('btnTurn').onclick = async () => {
  const user = document.getElementById('userInput').value.trim();
  const assistant = document.getElementById('assistantInput').value.trim();
  if (!user) return alert('请输入用户话');
  appendMsg('user', user);
  if (assistant) appendMsg('assistant', assistant);
  appendMsg('system', '一轮：retrieve → store（同 Host 顺序的变体）');
  const j = await post('/turn', {
    user_message: user,
    assistant_message: assistant,
    turn_id: 'console-' + Date.now(),
    store_after: !!assistant,
  });
  showOutput(j);
  await loadStats();
};

loadStats().catch(console.error);
