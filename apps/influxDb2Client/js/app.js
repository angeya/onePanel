/**
 * InfluxDB v2 客户端 - 主应用逻辑
 */
(function (global) {
  'use strict';

  /* ========== 工具函数 ========== */
  function $(id) { return document.getElementById(id); }
  function el(tag, cls, text) {
    var e = document.createElement(tag);
    if (cls) e.className = cls;
    if (text != null) e.textContent = text;
    return e;
  }
  function escapeHtml(s) {
    if (s == null) return '';
    return String(s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }

  var toastTimer = null;
  function toast(msg, type) {
    var t = $('toast');
    t.textContent = msg;
    t.className = 'toast show ' + (type || '');
    clearTimeout(toastTimer);
    toastTimer = setTimeout(function () { t.className = 'toast'; }, 3000);
  }
  function setStatus(msg, extra) {
    $('statusText').textContent = msg;
    $('statusExtra').textContent = extra || '';
  }

  /* ========== 应用状态 ========== */
  var state = {
    client: null,
    currentConn: null,
    currentBucket: null,
    currentMeasurement: null,
    tagKeys: [],
    fieldKeys: [],
    filters: {},
    page: 1,
    pageSize: 100,
    lastRowCount: 0,
    bucketExpanded: {},
    bucketMeasurements: {},
    editingConnId: null,
    filterTimer: null,
    lastQuery: '',
    customMode: false
  };

  /* ========== 连接列表 ========== */
  function renderConnList() {
    var list = Storage.list();
    var box = $('connList');
    box.innerHTML = '';
    if (!list.length) {
      box.appendChild(el('div', 'empty-hint', '暂无保存的连接，点击 + 新建'));
      return;
    }
    list.forEach(function (c) {
      var item = el('div', 'conn-item');
      item.setAttribute('data-id', c.id);
      if (state.currentConn && state.currentConn.id === c.id) item.className += ' active';

      var dot = el('span', 'conn-status');
      if (state.currentConn && state.currentConn.id === c.id) dot.className += ' online';
      item.appendChild(dot);

      var textWrap = el('div', 'conn-item-text');
      textWrap.appendChild(el('div', 'conn-name', c.name));
      textWrap.appendChild(el('div', 'conn-url', c.url));
      item.appendChild(textWrap);

      var actions = el('div', 'conn-actions');
      var btnConn = el('button', null, '连接');
      btnConn.title = '连接到此服务器';
      btnConn.onclick = function (e) { e.stopPropagation(); connectTo(c); };
      var btnEdit = el('button', null, '编辑');
      btnEdit.title = '编辑连接';
      btnEdit.onclick = function (e) { e.stopPropagation(); openConnModal(c); };
      var btnDel = el('button', null, '删除');
      btnDel.title = '删除连接';
      btnDel.onclick = function (e) { e.stopPropagation(); deleteConn(c); };
      actions.appendChild(btnConn);
      actions.appendChild(btnEdit);
      actions.appendChild(btnDel);
      item.appendChild(actions);

      // 双击连接
      item.ondblclick = function () { connectTo(c); };
      box.appendChild(item);
    });
  }

  /* ========== 连接弹窗 ========== */
  function openConnModal(conn) {
    state.editingConnId = conn ? conn.id : null;
    $('connModalTitle').textContent = conn ? '编辑连接' : '新建连接';
    $('connName').value = conn ? conn.name : '';
    $('connUrl').value = conn ? conn.url : '';
    $('connToken').value = conn ? conn.token : '';
    $('connOrg').value = conn ? conn.org : '';
    $('connBucket').value = conn ? (conn.defaultBucket || '') : '';
    $('connModal').style.display = 'flex';
  }
  function closeConnModal() { $('connModal').style.display = 'none'; }

  function readConnForm() {
    var name = $('connName').value.trim();
    var url = $('connUrl').value.trim();
    var token = $('connToken').value.trim();
    var org = $('connOrg').value.trim();
    var defaultBucket = $('connBucket').value.trim();
    if (!name || !url || !token || !org) {
      toast('请填写名称、URL、Token、Organization', 'error');
      return null;
    }
    return { id: state.editingConnId, name: name, url: url, token: token, org: org, defaultBucket: defaultBucket || '' };
  }

  function saveConn() {
    var data = readConnForm();
    if (!data) return;
    Storage.upsert(data);
    renderConnList();
    closeConnModal();
    toast('连接已保存', 'success');
  }

  function testConn() {
    var data = readConnForm();
    if (!data) return;
    setStatus('正在测试连接...');
    var client = InfluxClient.createClient(data);
    client.test().then(function () {
      toast('连接成功', 'success');
      setStatus('连接测试成功');
    }).catch(function (err) {
      toast('连接失败: ' + (err.message || err), 'error');
      setStatus('连接测试失败');
    });
  }

  function deleteConn(conn) {
    showConfirm('删除确认', '确定删除连接「' + conn.name + '」吗？', function () {
      Storage.remove(conn.id);
      if (state.currentConn && state.currentConn.id === conn.id) {
        disconnect();
      }
      renderConnList();
      toast('连接已删除', 'success');
    });
  }

  /* ========== 连接 / 断开 ========== */
  function connectTo(conn) {
    setStatus('正在连接 ' + conn.name + ' ...');
    var client = InfluxClient.createClient(conn);
    client.test().then(function () {
      state.client = client;
      state.currentConn = conn;
      state.currentBucket = null;
      state.currentMeasurement = null;
      state.bucketExpanded = {};
      state.bucketMeasurements = {};
      $('connInfo').textContent = conn.name + ' · ' + conn.url + ' · org: ' + conn.org;
      $('connInfo').className = 'conn-info connected';
      $('btnDisconnect').disabled = false;
      $('btnRefresh').disabled = false;
      $('btnToggleQuery').disabled = true;
      $('btnToggleQuery').textContent = '显示查询';
      $('queryPanel').style.display = 'none';
      $('queryInput').value = '';
      $('queryResult').innerHTML = '<div class="query-result-placeholder">执行结果将显示在此处</div>';
      updateQueryDisplay('');
      state.lastQuery = '';
      state.customMode = false;
      renderConnList();
      switchTab('tree');
      setStatus('已连接：' + conn.name);
      toast('连接成功', 'success');
      loadTree();
    }).catch(function (err) {
      var msg = err && err.message ? err.message : String(err);
      if (/Failed to fetch|NetworkError|CORS/i.test(msg)) {
        msg += '（可能需要开启 CORS，详见说明）';
      }
      toast('连接失败: ' + msg, 'error');
      setStatus('连接失败');
    });
  }

  function disconnect() {
    state.client = null;
    state.currentConn = null;
    state.currentBucket = null;
    state.currentMeasurement = null;
    state.tagKeys = [];
    state.fieldKeys = [];
    state.filters = {};
    $('connInfo').textContent = '未连接';
    $('connInfo').className = 'conn-info';
    $('btnDisconnect').disabled = true;
    $('btnRefresh').disabled = true;
    $('btnQuery').disabled = true;
    $('btnInsert').disabled = true;
    $('btnDeleteM').disabled = true;
    $('btnToggleQuery').disabled = true;
    $('queryPanel').style.display = 'none';
    $('btnToggleQuery').textContent = '显示查询';
    $('queryInput').value = '';
    $('queryResult').innerHTML = '<div class="query-result-placeholder">执行结果将显示在此处</div>';
    updateQueryDisplay('');
    state.lastQuery = '';
    state.customMode = false;
    $('tree').innerHTML = '<div class="empty-hint">请先连接服务器</div>';
    $('tableContainer').style.display = 'none';
    $('tablePlaceholder').style.display = 'flex';
    $('currentPath').textContent = '未选择 measurement';
    renderConnList();
    switchTab('connections');
    setStatus('已断开连接');
  }

  /* ========== 侧边栏 Tab ========== */
  function switchTab(tab) {
    var tabs = document.querySelectorAll('.sidebar-tabs .tab');
    tabs.forEach(function (t) { t.classList.toggle('active', t.getAttribute('data-tab') === tab); });
    $('connListPanel').style.display = tab === 'connections' ? 'block' : 'none';
    $('treePanel').style.display = tab === 'tree' ? 'block' : 'none';
  }

  /* ========== 数据浏览树 ========== */
  function loadTree() {
    if (!state.client) return;
    var tree = $('tree');
    tree.innerHTML = '<div class="tree-loading">加载 bucket 列表中...</div>';
    state.client.listBuckets().then(function (buckets) {
      renderTree(buckets);
    }).catch(function (err) {
      tree.innerHTML = '<div class="empty-hint">加载失败: ' + escapeHtml(err.message || err) + '</div>';
    });
  }

  function renderTree(buckets) {
    var tree = $('tree');
    tree.innerHTML = '';
    if (!buckets.length) {
      tree.appendChild(el('div', 'empty-hint', '没有可访问的 bucket'));
      return;
    }
    buckets.forEach(function (b) {
      var node = el('div', 'tree-node');
      node.setAttribute('data-bucket', b);

      var row = el('div', 'tree-row');
      var toggle = el('span', 'tree-toggle', '▸');
      var icon = el('span', 'tree-icon bucket', '▣');
      var label = el('span', 'tree-label', b);
      row.appendChild(toggle);
      row.appendChild(icon);
      row.appendChild(label);
      node.appendChild(row);

      var children = el('div', 'tree-children');
      node.appendChild(children);

      row.onclick = function () {
        var expanded = node.classList.toggle('expanded');
        toggle.textContent = expanded ? '▾' : '▸';
        if (expanded && !state.bucketExpanded[b]) {
          state.bucketExpanded[b] = true;
          loadMeasurements(b, children);
        }
      };

      // 默认 bucket 自动展开
      if (state.currentConn && state.currentConn.defaultBucket === b) {
        row.onclick();
      }
      tree.appendChild(node);
    });
  }

  function loadMeasurements(bucket, container) {
    container.innerHTML = '<div class="tree-loading">加载中...</div>';
    state.client.listMeasurements(bucket).then(function (ms) {
      state.bucketMeasurements[bucket] = ms;
      container.innerHTML = '';
      if (!ms.length) {
        container.appendChild(el('div', 'tree-loading', '（无 measurement）'));
        return;
      }
      ms.forEach(function (m) {
        var mNode = el('div', 'tree-node');
        var mRow = el('div', 'tree-row');
        mRow.appendChild(el('span', 'tree-toggle', ''));
        mRow.appendChild(el('span', 'tree-icon measurement', ' ◷'));
        mRow.appendChild(el('span', 'tree-label', m));
        mRow.onclick = function () {
          document.querySelectorAll('.tree-row.selected').forEach(function (r) { r.classList.remove('selected'); });
          mRow.classList.add('selected');
          selectMeasurement(bucket, m);
        };
        mNode.appendChild(mRow);
        container.appendChild(mNode);
      });
    }).catch(function (err) {
      container.innerHTML = '<div class="tree-loading">加载失败: ' + escapeHtml(err.message || err) + '</div>';
    });
  }

  /* ========== 选择 measurement ========== */
  function selectMeasurement(bucket, measurement) {
    state.currentBucket = bucket;
    state.currentMeasurement = measurement;
    state.filters = {};
    state.page = 1;
    $('currentPath').textContent = bucket + ' / ' + measurement;
    $('btnQuery').disabled = false;
    $('btnInsert').disabled = false;
    $('btnDeleteM').disabled = false;
    $('btnToggleQuery').disabled = false;
    setStatus('加载 ' + measurement + ' 的结构...');
    // 先加载 schema（tag/field），再查询数据
    loadSchema().then(function () {
      queryData();
    }).catch(function (err) {
      toast('加载结构失败: ' + (err.message || err), 'error');
      // 结构加载失败仍尝试查询
      queryData();
    });
  }

  function loadSchema() {
    return Promise.all([
      state.client.listTagKeys(state.currentBucket, state.currentMeasurement),
      state.client.listFieldKeys(state.currentBucket, state.currentMeasurement)
    ]).then(function (res) {
      state.tagKeys = res[0];
      state.fieldKeys = res[1];
    });
  }

  /* ========== 时间范围 ========== */
  function getTimeRange() {
    var sel = $('timeRange').value;
    if (sel === 'custom') {
      var start = $('rangeStart').value;
      var stop = $('rangeStop').value;
      if (!start || !stop) {
        toast('请填写自定义时间范围', 'error');
        return null;
      }
      return { start: new Date(start).toISOString(), stop: new Date(stop).toISOString() };
    }
    return { start: sel, stop: null };
  }

  /* ========== 查询数据 ========== */
  function queryData() {
    if (!state.client || !state.currentMeasurement) return;
    var range = getTimeRange();
    if (!range) return;
    var offset = (state.page - 1) * state.pageSize;
    setStatus('查询数据中...');
    $('tableContainer').style.display = 'none';
    $('tablePlaceholder').style.display = 'flex';
    $('tablePlaceholder').innerHTML = '<p>查询中...</p>';

    state.client.queryData({
      bucket: state.currentBucket,
      measurement: state.currentMeasurement,
      range: range,
      filters: state.filters,
      limit: state.pageSize,
      offset: offset
    }).then(function (result) {
      state.lastRowCount = result.rows.length;
      state.totalRows = result.totalRows || 0;
      state.totalPages = result.totalPages || 0;
      state.lastQuery = result.query || '';
      state.customMode = false;
      renderTable(result);
      updatePagination();
      updateQueryDisplay(state.lastQuery);
      setStatus('已加载 ' + result.rows.length + ' 条数据', state.currentBucket + '/' + state.currentMeasurement);
    }).catch(function (err) {
      $('tablePlaceholder').innerHTML = '<p style="color:#f0a0a0">查询失败: ' + escapeHtml(err.message || err) + '</p>';
      setStatus('查询失败');
    });
  }

  /* ========== 渲染表格 ========== */
  function orderColumns(columns) {
    var ordered = [];
    var rest = columns.slice();
    function pick(name) {
      var i = rest.indexOf(name);
      if (i >= 0) { ordered.push(name); rest.splice(i, 1); }
    }
    pick('_time');
    pick('_measurement');
    state.tagKeys.forEach(pick);
    state.fieldKeys.forEach(pick);
    // 剩余列（去除隐藏列）
    rest = rest.filter(function (c) { return c !== '_start' && c !== '_stop'; });
    return ordered.concat(rest);
  }

  function isNumericCol(col) {
    return state.fieldKeys.indexOf(col) >= 0;
  }

  function renderTable(result) {
    var container = $('tableContainer');
    container.innerHTML = '';
    if (!result.columns.length || !result.rows.length) {
      $('tableContainer').style.display = 'none';
      $('tablePlaceholder').style.display = 'flex';
      $('tablePlaceholder').innerHTML = '<div class="placeholder-icon">∅</div><p>没有查询到数据</p>';
      return;
    }
    var cols = orderColumns(result.columns);

    var scroll = el('div', 'table-scroll');
    var table = el('table', 'grid');
    var thead = el('thead');
    var tr = el('tr');
    cols.forEach(function (c) {
      var th = el('th');
      var content = el('div', 'th-content', c);
      th.appendChild(content);
      // 除时间列外均提供过滤输入框（自定义查询模式下不显示过滤）
      if (c !== '_time' && c !== '_start' && c !== '_stop' && !state.customMode) {
        var filterWrap = el('div', 'th-filter');
        var input = el('input');
        input.type = 'text';
        input.value = state.filters[c] || '';
        input.placeholder = '过滤...';
        input.setAttribute('data-col', c);
        input.oninput = function () { onFilterChange(c, input.value); };
        filterWrap.appendChild(input);
        th.appendChild(filterWrap);
      }
      tr.appendChild(th);
    });
    thead.appendChild(tr);
    table.appendChild(thead);

    var tbody = el('tbody');
    result.rows.forEach(function (row) {
      var trow = el('tr');
      cols.forEach(function (c) {
        var val = row[c];
        var td = el('td');
        if (c === '_time') td.className = 'time-col';
        else if (isNumericCol(c)) td.className = 'num-col';
        else if (state.tagKeys.indexOf(c) >= 0 || c === '_measurement') td.className = 'tag-col';
        td.textContent = val == null ? '' : String(val);
        td.title = val == null ? '' : String(val);
        trow.appendChild(td);
      });
      tbody.appendChild(trow);
    });
    table.appendChild(tbody);
    scroll.appendChild(table);
    container.appendChild(scroll);

    $('tablePlaceholder').style.display = 'none';
    $('tableContainer').style.display = 'block';
  }

  function onFilterChange(col, value) {
    state.filters[col] = value;
    state.page = 1;
    clearTimeout(state.filterTimer);
    state.filterTimer = setTimeout(function () { queryData(); }, 500);
  }

  /* ========== 状态栏查询语句显示与复制 ========== */
  function updateQueryDisplay(query) {
    var display = $('queryDisplay');
    var text = $('queryText');
    if (!query) {
      display.style.display = 'none';
      return;
    }
    text.textContent = query;
    text.title = query;
    display.style.display = 'flex';
  }

  function copyQuery() {
    var query = state.lastQuery;
    if (!query) return;
    var btn = $('btnCopyQuery');
    /** 兼容主流浏览器的复制实现 */
    function onSuccess() {
      btn.textContent = '已复制';
      btn.classList.add('copied');
      setTimeout(function () { btn.textContent = '复制'; btn.classList.remove('copied'); }, 1500);
    }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(query).then(onSuccess).catch(function () {
        fallbackCopy(query, onSuccess);
      });
    } else {
      fallbackCopy(query, onSuccess);
    }
  }

  /** execCommand 兜底复制方案 */
  function fallbackCopy(text, done) {
    var ta = document.createElement('textarea');
    ta.value = text;
    ta.style.position = 'fixed';
    ta.style.opacity = '0';
    document.body.appendChild(ta);
    ta.select();
    try { document.execCommand('copy'); done(); } catch (e) { toast('复制失败，请手动复制', 'error'); }
    document.body.removeChild(ta);
  }

  /* ========== 自定义查询面板 ========== */
  function toggleQueryPanel() {
    var panel = $('queryPanel');
    var btn = $('btnToggleQuery');
    var shown = panel.style.display !== 'none';
    if (shown) {
      panel.style.display = 'none';
      btn.textContent = '显示查询';
    } else {
      panel.style.display = 'flex';
      btn.textContent = '关闭查询';
    }
  }

  function renderQueryResult(type, msg) {
    var box = $('queryResult');
    box.innerHTML = '';
    var div = el('div', 'query-result-msg ' + type, msg);
    box.appendChild(div);
  }

  function executeCustomQuery() {
    if (!state.client) return;
    var fluxText = $('queryInput').value.trim();
    if (!fluxText) { toast('请输入查询语句', 'error'); return; }
    var btn = $('btnExecuteQuery');
    btn.disabled = true;
    btn.textContent = '执行中...';
    renderQueryResult('info', '正在执行查询...');
    setStatus('执行自定义查询中...');

    state.client.executeQuery(fluxText).then(function (result) {
      state.lastQuery = result.query;
      updateQueryDisplay(result.query);
      if (result.rows.length > 0) {
        // 有数据返回，加载到左侧表格
        state.lastRowCount = result.rows.length;
        state.customMode = true;
        state.filters = {};
        renderTable(result);
        updatePagination();
        renderQueryResult('success', '查询返回 ' + result.rows.length + ' 条数据，已加载到左侧表格。');
        setStatus('自定义查询返回 ' + result.rows.length + ' 条数据');
      } else {
        // 无数据返回，视为非查询语句执行结果
        renderQueryResult('info', '执行完成，无数据返回。');
        setStatus('自定义查询执行完成，无返回数据');
      }
    }).catch(function (err) {
      var msg = err && err.message ? err.message : String(err);
      renderQueryResult('error', '执行失败: ' + msg);
      setStatus('自定义查询执行失败');
    }).finally(function () {
      btn.disabled = false;
      btn.textContent = '执行';
    });
  }

  /* ========== 分页 ========== */
  function updatePagination() {
    var pageNo = $('pageNo');
    var info = $('pageInfo');
    if (state.customMode) {
      // 自定义查询不支持分页
      pageNo.textContent = '自定义查询';
      info.textContent = '共 ' + state.lastRowCount + ' 条';
      $('btnFirst').disabled = true;
      $('btnPrev').disabled = true;
      $('btnNext').disabled = true;
      return;
    }
    pageNo.textContent = '第 ' + state.page + ' / ' + state.totalPages + ' 页';
    info.innerHTML = '<span class="page-stat">共 ' + state.totalRows + ' 条</span><span class="page-stat">共 ' + state.totalPages + ' 页</span><span>本页 ' + state.lastRowCount + ' 条</span>';
    $('btnFirst').disabled = state.page <= 1;
    $('btnPrev').disabled = state.page <= 1;
    // 基于总条数判断是否还有下一页
    $('btnNext').disabled = state.page >= state.totalPages || state.totalPages === 0;
  }

  /* ========== 写入数据弹窗 ========== */
  function openInsertModal() {
    if (!state.client) return;
    $('insBucket').value = state.currentBucket || '';
    $('insMeasurement').value = state.currentMeasurement || '';
    $('insTimestamp').value = '';
    $('tagRows').innerHTML = '';
    $('fieldRows').innerHTML = '';
    // 预填已知 tag key
    if (state.tagKeys.length) {
      state.tagKeys.forEach(function (k) { addTagRow(k, ''); });
    } else {
      addTagRow('', '');
    }
    addFieldRow('', '', 'float');
    $('insertModal').style.display = 'flex';
  }

  function addTagRow(key, val) {
    var row = el('div', 'kv-row');
    var kInput = el('input', 'kv-key');
    kInput.value = key;
    kInput.placeholder = 'tag 名';
    var vInput = el('input', 'kv-val');
    vInput.value = val;
    vInput.placeholder = 'tag 值';
    var del = el('button', 'kv-del', '×');
    del.onclick = function () { row.remove(); };
    row.appendChild(kInput);
    row.appendChild(vInput);
    row.appendChild(del);
    $('tagRows').appendChild(row);
  }

  function addFieldRow(key, val, type) {
    var row = el('div', 'kv-row');
    var kInput = el('input', 'kv-key');
    kInput.value = key;
    kInput.placeholder = 'field 名';
    var vInput = el('input', 'kv-val');
    vInput.value = val;
    vInput.placeholder = 'field 值';
    var sel = el('select');
    [['string', '字符串'], ['float', '浮点'], ['int', '整数'], ['uint', '无符号'], ['bool', '布尔']].forEach(function (o) {
      var opt = el('option');
      opt.value = o[0];
      opt.textContent = o[1];
      sel.appendChild(opt);
    });
    sel.value = type || 'float';
    var del = el('button', 'kv-del', '×');
    del.onclick = function () { row.remove(); };
    row.appendChild(kInput);
    row.appendChild(vInput);
    row.appendChild(sel);
    row.appendChild(del);
    $('fieldRows').appendChild(row);
  }

  function writeData() {
    var bucket = $('insBucket').value.trim();
    var measurement = $('insMeasurement').value.trim();
    if (!bucket || !measurement) { toast('请填写 bucket 和 measurement', 'error'); return; }
    // 收集 tag
    var tags = {};
    var tagRows = $('tagRows').querySelectorAll('.kv-row');
    tagRows.forEach(function (r) {
      var k = r.querySelector('.kv-key').value.trim();
      var v = r.querySelector('.kv-val').value.trim();
      if (k) tags[k] = v;
    });
    // 收集 field
    var fields = [];
    var fieldRows = $('fieldRows').querySelectorAll('.kv-row');
    var hasField = false;
    fieldRows.forEach(function (r) {
      var k = r.querySelector('.kv-key').value.trim();
      var v = r.querySelector('.kv-val').value;
      var t = r.querySelector('select').value;
      if (k && v !== '') { fields.push({ name: k, value: v, type: t }); hasField = true; }
    });
    if (!hasField) { toast('请至少填写一个 field', 'error'); return; }
    // 时间戳
    var time = null;
    var tsVal = $('insTimestamp').value;
    if (tsVal) { time = new Date(tsVal); }

    setStatus('写入数据中...');
    state.client.writePoint({
      bucket: bucket,
      measurement: measurement,
      tags: tags,
      fields: fields,
      time: time
    }).then(function () {
      toast('写入成功', 'success');
      $('insertModal').style.display = 'none';
      setStatus('写入成功');
      queryData();
    }).catch(function (err) {
      toast('写入失败: ' + (err.message || err), 'error');
      setStatus('写入失败');
    });
  }

  /* ========== 删除数据弹窗 ========== */
  function openDeleteModal() {
    if (!state.client || !state.currentMeasurement) return;
    $('delBucket').value = state.currentBucket;
    $('delMeasurement').value = state.currentMeasurement;
    // 默认范围：当前查询范围或最近1天
    var now = new Date();
    var start = new Date(now.getTime() - 24 * 3600 * 1000);
    $('delStart').value = toLocalInputValue(start);
    $('delStop').value = toLocalInputValue(now);
    $('deleteModal').style.display = 'flex';
  }

  function toLocalInputValue(d) {
    var pad = function (n) { return n < 10 ? '0' + n : '' + n; };
    return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
      'T' + pad(d.getHours()) + ':' + pad(d.getMinutes());
  }

  function confirmDelete() {
    var bucket = $('delBucket').value;
    var measurement = $('delMeasurement').value;
    var start = $('delStart').value;
    var stop = $('delStop').value;
    if (!start || !stop) { toast('请选择时间范围', 'error'); return; }
    showConfirm('删除确认', '确认删除 ' + measurement + ' 在该时间范围内的全部数据？此操作不可恢复！', function () {
      setStatus('删除数据中...');
      state.client.deleteData({
        bucket: bucket,
        measurement: measurement,
        start: new Date(start).toISOString(),
        stop: new Date(stop).toISOString()
      }).then(function () {
        toast('删除成功', 'success');
        $('deleteModal').style.display = 'none';
        setStatus('删除成功');
        queryData();
      }).catch(function (err) {
        toast('删除失败: ' + (err.message || err), 'error');
        setStatus('删除失败');
      });
    });
  }

  /* ========== 侧边栏拖拽 ========== */
  function initResizer() {
    var resizer = $('resizer');
    var sidebar = $('sidebar');
    var dragging = false;
    resizer.addEventListener('mousedown', function (e) {
      dragging = true;
      document.body.style.cursor = 'col-resize';
      e.preventDefault();
    });
    document.addEventListener('mousemove', function (e) {
      if (!dragging) return;
      var w = e.clientX;
      if (w < 180) w = 180;
      if (w > 500) w = 500;
      sidebar.style.width = w + 'px';
    });
    document.addEventListener('mouseup', function () {
      if (dragging) { dragging = false; document.body.style.cursor = ''; }
    });
  }

  /* ========== 自定义确认框（替代浏览器原生 confirm，兼容 webview） ========== */
  function showConfirm(title, message, onOk, onCancel) {
    var modal = $('confirmModal');
    $('confirmTitle').textContent = title || '请确认';
    $('confirmMessage').textContent = message || '';
    modal.style.display = 'flex';

    var okBtn = $('btnConfirmOk');
    var cancelBtn = $('btnConfirmCancel');

    function cleanup() {
      okBtn.onclick = null;
      cancelBtn.onclick = null;
      modal.style.display = 'none';
    }

    okBtn.onclick = function () {
      cleanup();
      if (typeof onOk === 'function') onOk();
    };
    cancelBtn.onclick = function () {
      cleanup();
      if (typeof onCancel === 'function') onCancel();
    };
  }

  /* ========== 初始化 ========== */
  function init() {
    // 顶部按钮
    $('btnNewConn').onclick = function () { openConnModal(null); };
    $('btnRefresh').onclick = function () { queryData(); };
    $('btnDisconnect').onclick = disconnect;

    // 连接弹窗
    $('btnSaveConn').onclick = saveConn;
    $('btnTestConn').onclick = testConn;

    // 写入弹窗
    $('btnInsert').onclick = openInsertModal;
    $('btnAddTag').onclick = function () { addTagRow('', ''); };
    $('btnAddField').onclick = function () { addFieldRow('', '', 'float'); };
    $('btnWriteData').onclick = writeData;

    // 删除弹窗
    $('btnDeleteM').onclick = openDeleteModal;
    $('btnConfirmDelete').onclick = confirmDelete;

    // 自定义查询面板
    $('btnToggleQuery').onclick = toggleQueryPanel;
    $('btnCloseQueryPanel').onclick = toggleQueryPanel;
    $('btnExecuteQuery').onclick = executeCustomQuery;
    $('btnClearQuery').onclick = function () {
      $('queryInput').value = '';
      $('queryInput').focus();
    };
    $('btnCopyQuery').onclick = copyQuery;
    $('queryInput').addEventListener('keydown', function (e) {
      if (e.ctrlKey && e.key === 'Enter') { executeCustomQuery(); }
    });

    // 查询与分页
    $('btnQuery').onclick = function () { state.page = 1; queryData(); };
    $('btnFirst').onclick = function () { if (state.page > 1) { state.page = 1; queryData(); } };
    $('btnPrev').onclick = function () { if (state.page > 1) { state.page--; queryData(); } };
    $('btnNext').onclick = function () { if (state.lastRowCount >= state.pageSize) { state.page++; queryData(); } };
    $('pageSize').onchange = function () {
      state.pageSize = parseInt($('pageSize').value, 10);
      state.page = 1;
      queryData();
    };

    // 时间范围
    $('timeRange').onchange = function () {
      $('customRange').style.display = $('timeRange').value === 'custom' ? 'flex' : 'none';
    };

    // Tab 切换
    document.querySelectorAll('.sidebar-tabs .tab').forEach(function (t) {
      t.onclick = function () { switchTab(t.getAttribute('data-tab')); };
    });

    // 弹窗关闭
    document.querySelectorAll('[data-close]').forEach(function (btn) {
      btn.onclick = function () { $(btn.getAttribute('data-close')).style.display = 'none'; };
    });
    document.querySelectorAll('.modal-mask').forEach(function (mask) {
      mask.addEventListener('click', function (e) {
        if (e.target === mask) mask.style.display = 'none';
      });
    });

    initResizer();
    renderConnList();
    setStatus('就绪');
  }

  document.addEventListener('DOMContentLoaded', init);
})(window);
