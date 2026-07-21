/* ==========================================================================
   chat.js — сведён с internal/transport/websocket/hub.go:
   ← new_message {id, from, from_nick, to, content, created_at}
   ← online_list {online_ids:[...]}
   ← user_online / user_offline {from}
   ← typing {from, from_nick, to}
   → message {to, content} · → typing {to}
   История: GET /api/messages?with=ID&offset=N&limit=10
   ========================================================================== */

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
  
      // История могла прийти новые→старые; рисуем старые→новые
      if (batch.length > 1) {
        const t0 = Date.parse(Chat.norm(batch[0]).created_at || 0);
        const t1 = Date.parse(Chat.norm(batch[batch.length - 1]).created_at || 0);
        if (t0 > t1) batch.reverse();
      }
  
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
  
    // Понимает WS-формат (from/from_nick) и любой формат HTTP-истории
    norm(m) {
      return {
        from: m.from ?? m.sender_id ?? m.SenderID ?? m.From ?? 0,
        to: m.to ?? m.receiver_id ?? m.ReceiverID ?? m.To ?? 0,
        nick: m.from_nick ?? m.sender_nickname ?? m.FromNick ?? "",
        content: m.content ?? m.Content ?? "",
        created_at: m.created_at ?? m.CreatedAt ?? "",
      };
    },
  
    messageEl(raw) {
      const m = Chat.norm(raw);
      const mine = m.from === State.user.id;
      const el = document.createElement("div");
      el.className = "msg " + (mine ? "out" : "in");
      el.innerHTML = `
        <div>${escapeHTML(m.content)}</div>
        <div class="msg-meta">
          ${escapeHTML(mine ? State.user.nickname : (m.nick || (State.chat.peer && State.chat.peer.nickname) || ""))}
          · ${formatDate(m.created_at)}
        </div>
      `;
      return el;
    },
  
    appendMessage(raw) {
      const box = $("#chat-messages");
      const m = Chat.norm(raw);
      const nearBottom =
        box.scrollHeight - box.scrollTop - box.clientHeight < 80;
      box.appendChild(Chat.messageEl(raw));
      if (nearBottom || m.from === State.user.id) {
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
      // своё сообщение придёт эхом new_message от сервера — его и нарисуем
    },
  
    showTyping(nick) {
      const box = $("#chat-messages");
      if (!box) return;
      let hint = document.querySelector(".typing-hint");
      if (!hint) {
        hint = document.createElement("div");
        hint.className = "typing-hint";
        box.after(hint);
      }
      hint.textContent = `${nick || (State.chat.peer && State.chat.peer.nickname) || ""} is typing…`;
      clearTimeout(Chat._typingTimer);
      Chat._typingTimer = setTimeout(() => hint.remove(), 2500);
    },
  
    setOnline(userID, online) {
      const u = State.users.find((x) => x.id === userID);
      if (u) {
        u.online = online;
        Chat.renderUsers();
      } else if (online) {
        Chat.loadUsers(); // зарегистрировался кто-то новый
      }
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
  
      // Подгрузка истории по 10 при скролле вверх — с throttle (требование аудита)
      $("#chat-messages").addEventListener(
        "scroll",
        throttle(function () {
          if (this.scrollTop < 40) Chat.loadOlder(false);
        }, 500)
      );
    },
  };
  
  /* ---------- События ровно в именах твоего Hub ---------- */
  
  // Личное сообщение (сервер шлёт и получателю, и эхо отправителю)
  WS.on("new_message", (raw) => {
    const m = Chat.norm(raw);
    const peer = State.chat.peer;
    const chatOpen =
      peer &&
      !$("#chat-window").classList.contains("hidden") &&
      (m.from === peer.id || m.from === State.user.id) &&
      (m.to === peer.id || m.to === State.user.id);
  
    if (chatOpen) {
      Chat.appendMessage(raw);
    } else if (m.from !== State.user.id) {
      State.unread[m.from] = (State.unread[m.from] || 0) + 1;
    }
  
    const otherID = m.from === State.user.id ? m.to : m.from;
    const u = State.users.find((x) => x.id === otherID);
    if (u) u.last_message_at = m.created_at || new Date().toISOString();
    Chat.renderUsers();
  });
  
  // Снапшот онлайна при подключении: {type:"online_list", online_ids:[...]}
  WS.on("online_list", (msg) => {
    const set = new Set(msg.online_ids || []);
    State.users.forEach((u) => { u.online = set.has(u.id); });
    Chat.renderUsers();
  });
  
  // Кто-то вошёл/вышел: {type:"user_online"/"user_offline", from: userID}
  WS.on("user_online", (msg) => Chat.setOnline(msg.from, true));
  WS.on("user_offline", (msg) => Chat.setOnline(msg.from, false));
  
  // Печатает: {type:"typing", from, from_nick}
  WS.on("typing", (msg) => {
    if (State.chat.peer && msg.from === State.chat.peer.id) {
      Chat.showTyping(msg.from_nick);
    }
  });