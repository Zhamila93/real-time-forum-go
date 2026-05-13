// ws.js — клиент WebSocket с автореконнектом

const WS = {
  socket: null,
  reconnectDelay: 1000,
  intentionallyClosed: false,
  handlers: {},

  connect() {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) return;
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = `${proto}//${location.host}/ws`;
    this.intentionallyClosed = false;
    this.socket = new WebSocket(url);

    this.socket.onopen = () => {
      console.log('WS подключён');
      this.reconnectDelay = 1000;
    };

    this.socket.onmessage = (e) => {
      let data;
      try { data = JSON.parse(e.data); } catch (err) {
        console.warn('WS: bad JSON', e.data);
        return;
      }
      const h = this.handlers[data.type];
      if (h) h(data);
      else console.log('WS: нет обработчика для', data.type);
    };

    this.socket.onclose = () => {
      console.log('WS закрыт');
      if (!this.intentionallyClosed) {
        setTimeout(() => this.connect(), this.reconnectDelay);
        this.reconnectDelay = Math.min(this.reconnectDelay * 2, 15000);
      }
    };

    this.socket.onerror = (e) => {
      console.warn('WS ошибка', e);
    };
  },

  close() {
    this.intentionallyClosed = true;
    if (this.socket) this.socket.close();
    this.socket = null;
  },

  send(obj) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(obj));
    } else {
      console.warn('WS не подключён, сообщение не отправлено');
    }
  },

  on(type, handler) {
    this.handlers[type] = handler;
  },
};
