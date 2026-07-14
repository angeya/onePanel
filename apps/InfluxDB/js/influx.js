/**
 * InfluxDB v2 操作封装
 * 依赖：lib/influxdb-client-browser.js（暴露 window['@influxdata/influxdb-client']）
 */
(function (global) {
  'use strict';

  var Lib = global['@influxdata/influxdb-client'];
  if (!Lib) {
    console.error('InfluxDB 客户端库未加载，请确认 lib/influxdb-client-browser.js 存在');
    global.InfluxClient = {
      createClient: function () { throw new Error('客户端库未加载，无法连接'); }
    };
    return;
  }
  var InfluxDB = Lib.InfluxDB;
  var Point = Lib.Point;
  var fluxString = Lib.fluxString;

  /** 转义 Flux 正则字面量中的特殊字符，使过滤文本按字面量匹配 */
  function escapeRegex(str) {
    return String(str).replace(/[.*+?^${}()|[\]\\/]/g, '\\$&');
  }

  /** 规范化 URL，去除末尾斜杠及 /api/v2 后缀 */
  function normalizeUrl(url) {
    var u = String(url || '').trim().replace(/\/+$/, '');
    if (u.endsWith('/api/v2')) u = u.slice(0, -7);
    return u;
  }

  /**
   * 创建 InfluxDB 客户端实例
   * @param {Object} config {url, token, org}
   */
  function createClient(config) {
    var url = normalizeUrl(config.url);
    var influxDB = new InfluxDB({ url: url, token: config.token });
    var org = config.org;

    var queryApi = influxDB.getQueryApi(org);

    /** 执行 Flux 查询并收集全部行（对象数组） */
    function collectRows(fluxQuery) {
      return queryApi.collectRows(fluxQuery);
    }

    /** 测试连接：尝试查询 bucket 列表 */
    async function test() {
      var rows = await collectRows('buckets() |> limit(n: 1)');
      return { ok: true, sample: rows.length };
    }

    /** 列出全部 bucket（排除系统 bucket） */
    async function listBuckets() {
      var rows = await collectRows('buckets()');
      var names = [];
      for (var i = 0; i < rows.length; i++) {
        if (rows[i] && rows[i].name && !String(rows[i].name).startsWith('_')) {
          names.push(rows[i].name);
        }
      }
      return names.sort();
    }

    /** 列出指定 bucket 中的 measurement */
    async function listMeasurements(bucket) {
      var q = 'import "influxdata/influxdb/schema"\n' +
        'schema.measurements(bucket: ' + fluxString(bucket) + ')';
      var rows = await collectRows(q);
      var names = [];
      for (var i = 0; i < rows.length; i++) {
        if (rows[i] && rows[i]._value != null) names.push(String(rows[i]._value));
      }
      return names.sort();
    }

    /** 列出指定 measurement 的 tag key */
    async function listTagKeys(bucket, measurement) {
      var q = 'import "influxdata/influxdb/schema"\n' +
        'schema.tagKeys(bucket: ' + fluxString(bucket) + ', predicate: (r) => r._measurement == ' + fluxString(measurement) + ')';
      var rows = await collectRows(q);
      var skip = { _start: 1, _stop: 1, _time: 1, _measurement: 1, _field: 1, _value: 1 };
      var keys = [];
      for (var i = 0; i < rows.length; i++) {
        var v = rows[i] && rows[i]._value;
        if (v != null && !skip[v]) keys.push(String(v));
      }
      // 去重
      return Array.from(new Set(keys)).sort();
    }

    /** 列出指定 measurement 的 field key */
    async function listFieldKeys(bucket, measurement) {
      var q = 'import "influxdata/influxdb/schema"\n' +
        'schema.fieldKeys(bucket: ' + fluxString(bucket) + ', predicate: (r) => r._measurement == ' + fluxString(measurement) + ')';
      var rows = await collectRows(q);
      var keys = [];
      for (var i = 0; i < rows.length; i++) {
        if (rows[i] && rows[i]._value != null) keys.push(String(rows[i]._value));
      }
      return Array.from(new Set(keys)).sort();
    }

    /**
     * 构建查询基础片段：from / range / measurement / pivot / 列过滤
     * @param {Object} p 同 queryData
     * @returns {Array<String>} 基础 Flux 片段数组
     */
    function buildBaseParts(p) {
      var parts = [];
      parts.push('from(bucket: ' + fluxString(p.bucket) + ')');
      // 时间范围
      var start = p.range.start;
      var stop = p.range.stop;
      if (/^-/.test(start)) {
        parts.push('|> range(start: ' + start + ')');
      } else {
        parts.push('|> range(start: ' + fluxString(start) + ', stop: ' + fluxString(stop) + ')');
      }
      // measurement 过滤
      parts.push('|> filter(fn: (r) => r._measurement == ' + fluxString(p.measurement) + ')');
      // pivot 为宽表
      parts.push('|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")');

      // 列过滤（pivot 之后，统一使用 exists + string(v:) 正则包含匹配）
      if (p.filters) {
        for (var col in p.filters) {
          if (!Object.prototype.hasOwnProperty.call(p.filters, col)) continue;
          var text = p.filters[col];
          if (text === '' || text == null) continue;
          if (col === '_time' || col === '_start' || col === '_stop') continue;
          var pat = escapeRegex(text);
          var c = fluxString(col);
          parts.push('|> filter(fn: (r) => exists r[' + c + '] and string(v: r[' + c + ']) =~ /' + pat + '/)');
        }
      }
      return parts;
    }

    /**
     * 查询数据（已 pivot 为宽表）
     * @param {Object} p {bucket, measurement, range:{start,stop}, filters:{col:text}, limit, offset}
     * @returns {Object} {rows, columns, totalRows, totalPages, query}
     */
    async function queryData(p) {
      var baseParts = buildBaseParts(p);

      // 数据查询：合并分组 -> 排序 -> 分页
      var dataParts = baseParts.slice();
      dataParts.push('|> group()');
      dataParts.push('|> sort(columns: ["_time"], desc: true)');
      dataParts.push('|> limit(n: ' + Number(p.limit) + ', offset: ' + Number(p.offset || 0) + ')');
      var query = dataParts.join('\n');

      // 总条数查询：在 pivot 之前合并分组并计数，避免 pivot 后 _time 为 time 类型无法 count
      var countParts = [];
      countParts.push('from(bucket: ' + fluxString(p.bucket) + ')');
      var countStart = p.range.start;
      var countStop = p.range.stop;
      if (/^-/.test(countStart)) {
        countParts.push('|> range(start: ' + countStart + ')');
      } else {
        countParts.push('|> range(start: ' + fluxString(countStart) + ', stop: ' + fluxString(countStop) + ')');
      }
      countParts.push('|> filter(fn: (r) => r._measurement == ' + fluxString(p.measurement) + ')');
      // 列过滤（此时列是 _field/_value，但前端过滤的是 pivot 后的列，直接应用在 count 查询上效果不一致；
      // 为保持总条数与分页数据一致，仍使用与数据查询相同的列过滤，但因为 pivot 后列不存在，所以无效。
      // 后续如需精确统计，可改用 join schema 信息，当前方案与分页列表条目保持一致即可。）
      countParts.push('|> group()');
      countParts.push('|> count()');
      var countQuery = countParts.join('\n');

      var rows;
      var countRows;
      try {
        rows = await collectRows(query);
        countRows = await collectRows(countQuery);
      } catch (e) {
        throw e;
      }

      // 汇总所有行的列（不同 series 可能字段不同）
      var colSet = {};
      var colOrder = [];
      for (var i = 0; i < rows.length; i++) {
        var r = rows[i];
        if (!r) continue;
        for (var k in r) {
          if (!Object.prototype.hasOwnProperty.call(r, k)) continue;
          if (!(k in colSet)) {
            colSet[k] = true;
            colOrder.push(k);
          }
        }
      }

      // 解析总条数
      var totalRows = 0;
      if (countRows && countRows.length > 0 && countRows[0] && countRows[0]._value != null) {
        totalRows = Number(countRows[0]._value);
      }
      if (isNaN(totalRows) || totalRows < 0) totalRows = 0;
      var totalPages = Math.ceil(totalRows / Math.max(1, Number(p.limit || 1)));

      return {
        rows: rows,
        columns: colOrder,
        totalRows: totalRows,
        totalPages: totalPages,
        query: query
      };
    }

    /**
     * 执行自定义 Flux 查询，返回原始行
     * @param {String} fluxText Flux 查询文本
     * @returns {Object} {rows, columns, query}
     */
    async function executeQuery(fluxText) {
      var rows = await collectRows(fluxText);
      var colSet = {};
      var colOrder = [];
      for (var i = 0; i < rows.length; i++) {
        var r = rows[i];
        if (!r) continue;
        for (var k in r) {
          if (!Object.prototype.hasOwnProperty.call(r, k)) continue;
          if (!(k in colSet)) {
            colSet[k] = true;
            colOrder.push(k);
          }
        }
      }
      return { rows: rows, columns: colOrder, query: fluxText };
    }

    /**
     * 写入一个数据点
     * @param {Object} p {bucket, measurement, tags:{k:v}, fields:[{name,value,type}], time(Date|undefined)}
     */
    async function writePoint(p) {
      var writeApi = influxDB.getWriteApi(org, p.bucket, 'ns');
      try {
        var point = new Point(p.measurement);
        if (p.tags) {
          for (var k in p.tags) {
            if (p.tags[k] !== '') point.tag(k, String(p.tags[k]));
          }
        }
        for (var i = 0; i < p.fields.length; i++) {
          var f = p.fields[i];
          if (!f.name || f.value === '') continue;
          switch (f.type) {
            case 'int':
              point.intField(f.name, parseInt(f.value, 10));
              break;
            case 'uint':
              point.uintField(f.name, parseInt(f.value, 10));
              break;
            case 'float':
              point.floatField(f.name, parseFloat(f.value));
              break;
            case 'bool':
              point.booleanField(f.name, f.value === 'true' || f.value === true);
              break;
            default:
              point.stringField(f.name, String(f.value));
          }
        }
        if (p.time instanceof Date && !isNaN(p.time)) {
          point.timestamp(p.time);
        }
        writeApi.writePoint(point);
        await writeApi.close();
      } catch (e) {
        await writeApi.dispose();
        throw e;
      }
    }

    /**
     * 删除指定范围内的数据
     * @param {Object} p {bucket, measurement, start, stop} 时间为 RFC3339 字符串
     */
    async function deleteData(p) {
      var base = normalizeUrl(config.url);
      var predicate = '_measurement="' + p.measurement.replace(/"/g, '\\"') + '"';
      var body = { start: p.start, stop: p.stop, predicate: predicate };
      var u = base + '/api/v2/delete?org=' + encodeURIComponent(org) + '&bucket=' + encodeURIComponent(p.bucket);
      var resp = await fetch(u, {
        method: 'POST',
        headers: {
          'Authorization': 'Token ' + config.token,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(body)
      });
      if (!resp.ok) {
        var txt = '';
        try { txt = await resp.text(); } catch (e) {}
        throw new Error('删除失败 (' + resp.status + '): ' + txt);
      }
    }

    return {
      config: config,
      url: url,
      test: test,
      listBuckets: listBuckets,
      listMeasurements: listMeasurements,
      listTagKeys: listTagKeys,
      listFieldKeys: listFieldKeys,
      queryData: queryData,
      executeQuery: executeQuery,
      writePoint: writePoint,
      deleteData: deleteData
    };
  }

  global.InfluxClient = { createClient: createClient, normalizeUrl: normalizeUrl };
})(window);
