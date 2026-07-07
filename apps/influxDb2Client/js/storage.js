/**
 * 连接信息本地存储管理
 * 使用 localStorage 持久化保存连接列表
 */
(function (global) {
  'use strict';

  var STORAGE_KEY = 'influxdb_client_connections';

  /**
   * 读取全部连接
   * @returns {Array} 连接数组
   */
  function list() {
    try {
      var raw = localStorage.getItem(STORAGE_KEY);
      return raw ? JSON.parse(raw) : [];
    } catch (e) {
      return [];
    }
  }

  /**
   * 保存全部连接
   * @param {Array} conns 连接数组
   */
  function save(conns) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(conns));
  }

  /**
   * 新增或更新连接（存在同 id 则覆盖）
   * @param {Object} conn 连接对象
   * @returns {Object} 保存后的连接
   */
  function upsert(conn) {
    var conns = list();
    if (!conn.id) {
      conn.id = 'c_' + Date.now() + '_' + Math.random().toString(36).slice(2, 7);
      conn.createdAt = Date.now();
      conns.push(conn);
    } else {
      var found = false;
      for (var i = 0; i < conns.length; i++) {
        if (conns[i].id === conn.id) {
          conns[i] = conn;
          found = true;
          break;
        }
      }
      if (!found) conns.push(conn);
    }
    save(conns);
    return conn;
  }

  /**
   * 根据 id 删除连接
   * @param {String} id 连接 id
   */
  function remove(id) {
    var conns = list().filter(function (c) { return c.id !== id; });
    save(conns);
  }

  /**
   * 根据 id 获取连接
   * @param {String} id 连接 id
   * @returns {Object|null}
   */
  function get(id) {
    var conns = list();
    for (var i = 0; i < conns.length; i++) {
      if (conns[i].id === id) return conns[i];
    }
    return null;
  }

  global.Storage = { list: list, save: save, upsert: upsert, remove: remove, get: get };
})(window);
