/* chat.js — members rail + private chat with 10-message pagination */

const Chat = {
    PAGE: 10,
  
    async loadUsers() {
      try { State.users = (await API.users()) || []; }
      catch { State.users = []; }
      Chat.renderUsers();
    },
  
    sortUsers() {
      State.users.sort((a, b) => {
        const ta = a.last_message_at ? Date.parse(a.last_message_at) : 0;
        const tb = b.last_message_at ? Date.parse(b.last_message_at) : 0;
        if (ta !== tb) return tb - ta;
        return (a.nickname || "").localeCompare(b.nickname || "");
      });
    },
  
    renderUsers() {
      const box = $("#users-list");
      if (!box || !State.user) return;
      Chat.sortUsers();
      box.innerHTML = "";
  
      State.users
        .filter((u) => u.id !== State.user.id)
        .forEach((u) => {
          const item = document.createElement("div");
          const unread = State.unread[u.id] || 0;
          item.className = "user-item" + (unread ? " unread" : "");
          item.innerHTML = `
            <span class="dot ${u.online ? "online" : ""}"></span>
            <span class="user-name">${escapeHTML(u.nickname)}</span>
            ${unread ? `<span class="user-badge">${unread}</span>` : ""}
          `;
          item.addEventListener("click", () => Chat.open(u));
          box.appendChild(item);
        });
    },
  
    async open(user) {
      State.chat.peer = user;
      State.chat.offset = 0;
      State.chat.done = false;
      State.chat.loading = false;
      delete State.unread[user.id];
      Chat.renderUsers();
  
      $("#chat-with").textContent = user.nickname;
      $("#chat-messages").innerHTML = "";
      $("#chat-window").classList.remove("hidden");
      $("#chat-input").focus();
  
      await Chat.loadOlder(true);
    },
  
    close() {
      State.chat.peer = null;
      $("#chat-window").classList.add("hidden");
    },
  
    async loadOlder(initial = false) {
      const st = State.chat;
      if (!st.peer || st.loading || st.done) return;
      st.loading = true;
  
      const box = $("#chat-messages");
      const spinner = document.createElement("div");
      spinner.className = "loading-more";
      spinner.textContent = "Loading…";
      box.prepend(spinner);
  
      let batch = [];
      try {
        batch = (await API.messages(st.peer.id, st.offset, Chat.PAGE)) || [];
      } catch { batch = []; }
  
      spinner.remove();
  
      if (batch.length < Chat.PAGE) st.done = true;
      st.offset += batch.length;
  
      const prevHeight = box.scrollHeight;
      const frag = document.createDocumentFragment();
      batch.forEach((m) => frag.appendChild(Chat.messageEl(m)));
      box.prepend(frag);
  
      if (initial) {
        box.scrollTop = box.scrollHeight;
      } else {
        box.scrollTop = box.scrollHeight - prevHeight;
      }
      st.loading = false;
    },
  
    messageEl(m) {
      const mine = m.sender_id === State.user.id;
      const el = document.createElement("div");
      el.className = "msg " + (mine ? "out" : "in");
      el.innerHTML = `
        <div>${escapeHTML(m.content)}</div>
        <div class="msg-meta">
          ${escapeHTML(mine ? State.user.nickname : (m.sender_nickname || (State.chat.peer && State.chat.peer.nickname) || ""))}
          · ${formatDate(m.created_at)}
        </div>
      `;
      return el;
    },
  
    appendMessage(m) {
      const box = $("#chat-messages");
      const nearBottom =
        box.scrollHeight - box.scrollTop - box.clientHeight < 80;
      box.appendChild(Chat.messageEl(m));
      if (nearBottom || m.sender_id === State.user.id) {
        box.scrollTop = box.scrollHeight;
      }
      State.chat.offset += 1;
    },
  
    send() {
      const st = State.chat;
      const input = $("#chat-input");
      const text = input.value.trim();
      if (!text || !st.peer) return;
      const ok = WS.send("message", { to: st.peer.id, content: text });
      if (!ok) return;
      input.value = "";
    },
  
    showTyping() {
      const box = $("#chat-messages");
      if (!box) return;
      let hint = document.querySelector(".typing-hint");
      if (!hint) {
        hint = document.createElement("div");
        hint.className = "typing-hint";
        box.after(hint);
      }
      hint.textContent = `${(State.chat.peer && State.chat.peer.nickname) || ""} is typing…`;
      clearTimeout(Chat._typingTimer);
      Chat._typingTimer = setTimeout(() => hint.remove(), 2500);
    },
  
    init() {
      $("#chat-close").addEventListener("click", () => Chat.close());
  
      $("#chat-form").addEventListener("submit", (e) => {
        e.preventDefault();
        Chat.send();
      });
  
      const sendTyping = debounce(() => {
        if (State.chat.peer) WS.send("typing", { to: State.chat.peer.id });
      }, 300);
      $("#chat-input").addEventListener("input", sendTyping);
  
      $("#chat-messages").addEventListener(
        "scroll",
        throttle(function () {
          if (this.scrollTop < 40) Chat.loadOlder(false);
        }, 500)
      );
    },
  };
  
  WS.on("message", (m) => {
    const peer = State.chat.peer;
    const chatOpen =
      peer &&
      !$("#chat-window").classList.contains("hidden") &&
      (m.sender_id === peer.id || m.sender_id === State.user.id) &&
      (m.receiver_id === peer.id || m.receiver_id === State.user.id);
  
    if (chatOpen) {
      Chat.appendMessage(m);
    } else if (m.sender_id !== State.user.id) {
      State.unread[m.sender_id] = (State.unread[m.sender_id] || 0) + 1;
    }
  
    const otherID = m.sender_id === State.user.id ? m.receiver_id : m.sender_id;
    const u = State.users.find((x) => x.id === otherID);
    if (u) u.last_message_at = m.created_at || new Date().toISOString();
    Chat.renderUsers();
  });
  
  WS.on("users", (msg) => {
    if (Array.isArray(msg.users)) {
      State.users = msg.users;
      Chat.renderUsers();
    }
  });
  
  WS.on("user_status", (msg) => {
    const u = State.users.find((x) => x.id === msg.user_id);
    if (u) {
      u.online = !!msg.online;
      Chat.renderUsers();
    } else {
      Chat.loadUsers();
    }
  });
  
  WS.on("typing", (msg) => {
    if (State.chat.peer && msg.from === State.chat.peer.id) Chat.showTyping();
  });