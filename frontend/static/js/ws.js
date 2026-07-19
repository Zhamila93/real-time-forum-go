/* ws.js — single WebSocket connection with reconnect */

const WS = {
    socket: null,
    handlers: {},
    shouldReconnect: false,
    reconnectDelay: 2000,
  
    connect() {
      this.shouldReconnect = true;
      const proto = location.protocol === "https:" ? "wss" : "ws";
      this.socket = new WebSocket(`${proto}://${location.host}/ws`);
  
      this.socket.addEventListener("message", (ev) => {
        let msg;
        try { msg = JSON.parse(ev.data); } catch { return; }
        const list = this.handlers[msg.type];
        if (list) list.forEach((fn) => fn(msg));
      });
  
      this.socket.addEventListener("close", () => {
        this.socket = null;
        if (this.shouldReconnect) {
          setTimeout(() => this.connect(), this.reconnectDelay);
        }
      });
  
      this.socket.addEventListener("error", () => {});
    },
  
    disconnect() {
      this.shouldReconnect = false;
      if (this.socket) this.socket.close();
      this.socket = null;
    },
  
    on(type, fn) {
      (this.handlers[type] = this.handlers[type] || []).push(fn);
    },
  
    send(type, payload = {}) {
      if (!this.socket || this.socket.readyState !== WebSocket.OPEN) return false;
      this.socket.send(JSON.stringify({ type, ...payload }));
      return true;
    },
  };