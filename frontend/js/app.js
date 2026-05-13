// app.js — точка входа, роутинг

const App = {
  async init() {
    // делегирование кликов с data-go
    document.addEventListener('click', (e) => {
      const a = e.target.closest('[data-go]');
      if (a) {
        e.preventDefault();
        this.go(a.dataset.go);
      }
    });

    // проверяем, авторизованы ли мы
    try {
      const me = await API.me();
      State.me = me;
      await this.enterApp();
    } catch (e) {
      this.go('login');
    }
  },

  go(view) {
    State.view = view;
    if (view === 'login') {
      Auth.renderLogin();
      return;
    }
    if (view === 'register') {
      Auth.renderRegister();
      return;
    }

    // Авторизованные виды — нужна layout
    if (!document.querySelector('.layout')) {
      this.renderLayout();
    }

    if (view === 'feed') {
      Posts.renderFeed();
    } else if (view === 'newpost') {
      Posts.renderNewPost();
    }
  },

  renderLayout() {
    const tpl = document.getElementById('tpl-main').content.cloneNode(true);
    const app = document.getElementById('app');
    app.innerHTML = '';
    app.appendChild(tpl);

    document.getElementById('me-nick').textContent = '@' + State.me.nickname;
    document.getElementById('logout-btn').addEventListener('click', () => Auth.logout());

    document.getElementById('chat-close').addEventListener('click', () => Chat.closeChat());

    Chat.setupSendForm();
    Chat.setupScrollListener();
  },

  async enterApp() {
    // Подготовить WebSocket-обработчики
    WS.on('online_list', (data) => {
      State.onlineIDs = new Set(data.online_ids || []);
      Chat.renderUserList();
    });
    WS.on('user_online', (data) => {
      State.onlineIDs.add(data.from);
      Chat.renderUserList();
    });
    WS.on('user_offline', (data) => {
      State.onlineIDs.delete(data.from);
      Chat.renderUserList();
    });
    WS.on('new_message', (data) => {
      Chat.onNewMessage(data);
    });
    WS.on('typing', (data) => {
      // здесь можно показывать индикатор "печатает"
    });

    WS.connect();

    // показать ленту по умолчанию
    this.go('feed');

    // подгрузить пользователей в сайдбар
    await Chat.loadUsers();
  },
};

document.addEventListener('DOMContentLoaded', () => App.init());
