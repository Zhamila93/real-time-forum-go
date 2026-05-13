// auth.js — обработчики форм логина и регистрации

const Auth = {
  renderLogin() {
    const tpl = document.getElementById('tpl-login').content.cloneNode(true);
    const app = document.getElementById('app');
    app.innerHTML = '';
    app.appendChild(tpl);

    document.getElementById('login-form').addEventListener('submit', async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      const body = {
        identifier: fd.get('identifier'),
        password: fd.get('password'),
      };
      try {
        const data = await API.login(body);
        State.me = data;
        await App.enterApp();
      } catch (err) {
        document.getElementById('login-err').textContent = err.message;
      }
    });
  },

  renderRegister() {
    const tpl = document.getElementById('tpl-register').content.cloneNode(true);
    const app = document.getElementById('app');
    app.innerHTML = '';
    app.appendChild(tpl);

    document.getElementById('register-form').addEventListener('submit', async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      const body = {
        nickname: fd.get('nickname'),
        age: parseInt(fd.get('age'), 10),
        gender: fd.get('gender'),
        first_name: fd.get('first_name'),
        last_name: fd.get('last_name'),
        email: fd.get('email'),
        password: fd.get('password'),
      };
      try {
        const data = await API.register(body);
        State.me = data;
        await App.enterApp();
      } catch (err) {
        document.getElementById('register-err').textContent = err.message;
      }
    });
  },

  async logout() {
    try { await API.logout(); } catch (e) {}
    WS.close();
    State.me = null;
    State.onlineIDs.clear();
    State.unreadFrom.clear();
    State.activeChat = null;
    App.go('login');
  },
};
