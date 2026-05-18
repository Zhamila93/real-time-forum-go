// chat.js — приватные сообщения, список пользователей, реалтайм

const Chat = {
  typingTimer: null,

  msgKey(m) {
    if (m.id != null && m.id !== '') return 'id:' + m.id;
    const t = m.created_at || '';
    const sid = userId(m.sender_id != null ? m.sender_id : m.from);
    return 'tmp:' + sid + ':' + t + ':' + (m.content || '');
  },

  hasMessageInBox(msgsBox, key) {
    return Array.from(msgsBox.querySelectorAll('[data-msg-key]')).some((el) => el.dataset.msgKey === key);
  },

  // ===== Список пользователей в сайдбаре =====
  async loadUsers() {
    try {
      const users = await API.users();
      State.users = users || [];
      this.renderUserList();
    } catch (err) {
      console.warn('users load err:', err);
    }
  },

  renderUserList() {
    const box = document.getElementById('users-list');
    if (!box) return;

    // обновим онлайн-статусы и пересортируем
    const users = State.users.map(u => ({
      ...u,
      online: State.onlineIDs.has(userId(u.id)),
    }));

    // сортировка: сначала те с last_message по убыванию даты, потом по алфавиту
    const withMsg = users.filter(u => u.last_message).sort((a, b) =>
      new Date(b.last_message) - new Date(a.last_message)
    );
    const withoutMsg = users.filter(u => !u.last_message).sort((a, b) =>
      a.nickname.localeCompare(b.nickname)
    );
    const sorted = [...withMsg, ...withoutMsg];

    box.innerHTML = '';
    if (sorted.length === 0) {
      box.innerHTML = '<p style="color:var(--fg-dim); font-size:13px;">Нет других пользователей.</p>';
      return;
    }

    sorted.forEach(u => {
      const it = document.createElement('div');
      it.className = 'user-item' + (State.unreadFrom.has(userId(u.id)) ? ' unread' : '');
      it.innerHTML = `
        <span class="dot ${u.online ? 'online' : ''}"></span>
        <span class="user-name">${escapeHTML(u.nickname)}</span>
        ${State.unreadFrom.has(userId(u.id)) ? '<span class="user-badge">!</span>' : ''}
      `;
      it.addEventListener('click', () => Chat.openChat(u));
      box.appendChild(it);
    });
  },

  // ===== Открытие чата с собеседником =====
  async openChat(user) {
    State.activeChat = user;
    State.chatOffset = 0;
    State.chatHasMore = true;
    State.chatLoading = false;
    State.unreadFrom.delete(userId(user.id));
    this.renderUserList();

    const win = document.getElementById('chat-window');
    win.classList.remove('hidden');
    document.getElementById('chat-with').textContent = '💬 ' + user.nickname;

    const msgsBox = document.getElementById('chat-messages');
    msgsBox.innerHTML = '';

    await this.loadMessages(true);
  },

  closeChat() {
    State.activeChat = null;
    this.showTyping('', false);
    if (this.typingTimer) clearTimeout(this.typingTimer);
    document.getElementById('chat-window').classList.add('hidden');
  },

  // ===== Подгрузка истории (с пагинацией) =====
  async loadMessages(initial = false) {
    if (!State.activeChat) return;
    if (State.chatLoading || !State.chatHasMore) return;

    State.chatLoading = true;
    const msgsBox = document.getElementById('chat-messages');

    let loader;
    if (!initial) {
      loader = document.createElement('div');
      loader.className = 'loading-more';
      loader.textContent = 'Загрузка...';
      msgsBox.prepend(loader);
    }

    try {
      const data = await API.messages(State.activeChat.id, State.chatOffset, 10);
      const msgs = data || [];

      // запомним позицию прокрутки, чтобы не "прыгало"
      const oldScrollHeight = msgsBox.scrollHeight;

      if (loader) loader.remove();

      if (msgs.length < 10) State.chatHasMore = false;
      State.chatOffset += msgs.length;

      // msgs идут в порядке возрастания времени — вставляем в начало
      const frag = document.createDocumentFragment();
      msgs.forEach(m => frag.appendChild(this.renderMsg(m)));

      if (initial) {
        msgsBox.appendChild(frag);
        msgsBox.scrollTop = msgsBox.scrollHeight;
      } else {
        msgsBox.prepend(frag);
        // удерживаем визуальное положение
        msgsBox.scrollTop = msgsBox.scrollHeight - oldScrollHeight;
      }
    } catch (err) {
      console.warn('load messages err:', err);
    } finally {
      State.chatLoading = false;
    }
  },

  renderMsg(m) {
    const d = document.createElement('div');
    const meId = userId(State.me && State.me.id);
    const isOut = sameUserId(m.sender_id, meId) || sameUserId(m.from, meId);
    d.className = 'msg ' + (isOut ? 'out' : 'in') + (m.pending ? ' pending' : '');
    d.dataset.msgKey = this.msgKey(m);
    if (m.id != null && m.id !== '') d.dataset.msgId = String(m.id);
    const sender = m.sender || m.from_nick || (isOut ? State.me.nickname : (State.activeChat && State.activeChat.nickname) || '');
    const created = m.created_at;
    d.innerHTML = `
      <div>${escapeHTML(m.content)}</div>
      <div class="msg-meta">${escapeHTML(sender)} • ${formatDate(created)}</div>
    `;
    return d;
  },

  appendToActiveChat(m) {
    const msgsBox = document.getElementById('chat-messages');
    if (!msgsBox) return false;
    if (this.hasMessageInBox(msgsBox, this.msgKey(m))) return false;
    const wasAtBottom = msgsBox.scrollHeight - msgsBox.scrollTop - msgsBox.clientHeight < 60;
    msgsBox.appendChild(this.renderMsg(m));
    if (wasAtBottom) msgsBox.scrollTop = msgsBox.scrollHeight;
    return true;
  },

  showTyping(fromNick, show) {
    const msgsBox = document.getElementById('chat-messages');
    if (!msgsBox) return;
    let el = msgsBox.querySelector('.typing-hint');
    if (!show) {
      if (el) el.remove();
      return;
    }
    if (!el) {
      el = document.createElement('div');
      el.className = 'typing-hint';
      msgsBox.appendChild(el);
    }
    el.textContent = (fromNick || 'Собеседник') + ' печатает…';
    msgsBox.scrollTop = msgsBox.scrollHeight;
  },

  onTyping(data) {
    if (!State.activeChat || !State.me) return;
    const fromId = userId(data.from);
    if (userId(State.activeChat.id) !== fromId) return;
    this.showTyping(data.from_nick, true);
    if (this.typingTimer) clearTimeout(this.typingTimer);
    this.typingTimer = setTimeout(() => this.showTyping('', false), 2500);
  },

  // ===== Скролл-листенер (throttle!) =====
  setupScrollListener() {
    const msgsBox = document.getElementById('chat-messages');
    if (!msgsBox) return;

    const onScroll = throttle(() => {
      if (msgsBox.scrollTop < 60 && State.chatHasMore && !State.chatLoading) {
        this.loadMessages(false);
      }
    }, 300);

    msgsBox.addEventListener('scroll', onScroll);
  },

  // ===== Отправка =====
  setupSendForm() {
    const form = document.getElementById('chat-form');
    if (!form) return;
    form.addEventListener('submit', (e) => {
      e.preventDefault();
      if (!State.activeChat) return;
      const input = document.getElementById('chat-input');
      const text = input.value.trim();
      if (!text) return;
      const peerId = userId(State.activeChat.id);
      if (Number.isNaN(peerId)) return;

      const now = new Date().toISOString();
      this.appendToActiveChat({
        sender_id: State.me.id,
        sender: State.me.nickname,
        content: text,
        created_at: now,
        pending: true,
      });
      const listUser = State.users.find(x => userId(x.id) === peerId);
      if (listUser) {
        listUser.last_message = now;
        this.renderUserList();
      }

      WS.send({
        type: 'message',
        to: peerId,
        content: text,
      });
      input.value = '';
      this.showTyping('', false);
    });

    // debounce: индикатор "печатает" (можно расширить — здесь как демонстрация)
    const input = document.getElementById('chat-input');
    const sendTyping = debounce(() => {
      if (State.activeChat) {
        const id = userId(State.activeChat.id);
        if (!Number.isNaN(id)) {
          WS.send({ type: 'typing', to: id });
        }
      }
    }, 500);
    input.addEventListener('input', sendTyping);
  },

  // ===== Реалтайм-приём сообщения =====
  onNewMessage(data) {
    if (!State.me) return;
    // data: { id, from, from_nick, to, content, created_at }
    // обновим список users — last_message
    const meId = userId(State.me.id);
    const fromId = userId(data.from);
    const toId = userId(data.to);
    const otherID = fromId === meId ? toId : fromId;
    const u = State.users.find(x => userId(x.id) === otherID);
    if (u) {
      u.last_message = data.created_at;
    }

    // если открыт чат с этим пользователем — добавим в окно (без дубликатов)
    if (State.activeChat && userId(State.activeChat.id) === otherID) {
      const added = this.appendToActiveChat({
        id: data.id,
        sender_id: data.from,
        sender: data.from_nick,
        content: data.content,
        created_at: data.created_at,
      });
      if (added) State.chatOffset += 1;
      this.showTyping('', false);
    } else if (!sameUserId(data.from, meId) && !Number.isNaN(fromId)) {
      // непрочитанное от другого
      State.unreadFrom.add(fromId);
    }

    this.renderUserList();
  },
};
