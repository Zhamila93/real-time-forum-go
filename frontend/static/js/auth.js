/* auth.js — login / register / logout */

const Auth = {
    renderLogin() {
      const app = $("#app");
      app.innerHTML = "";
      app.appendChild(tpl("tpl-login"));
  
      $("#login-form").addEventListener("submit", async (e) => {
        e.preventDefault();
        setError("login-err", "");
        const f = e.target;
        try {
          await API.login(f.identifier.value.trim(), f.password.value);
          await App.start();
        } catch (err) {
          setError("login-err", err.message);
        }
      });
    },
  
    renderRegister() {
      const app = $("#app");
      app.innerHTML = "";
      app.appendChild(tpl("tpl-register"));
  
      $("#register-form").addEventListener("submit", async (e) => {
        e.preventDefault();
        setError("register-err", "");
        const f = e.target;
        const payload = {
          nickname:   f.nickname.value.trim(),
          age:        Number(f.age.value),
          gender:     f.gender.value,
          first_name: f.first_name.value.trim(),
          last_name:  f.last_name.value.trim(),
          email:      f.email.value.trim(),
          password:   f.password.value,
        };
        try {
          await API.register(payload);
          await API.login(payload.nickname, payload.password);
          await App.start();
        } catch (err) {
          setError("register-err", err.message);
        }
      });
    },
  
    async logout() {
      try { await API.logout(); } catch {}
      WS.disconnect();
      State.reset();
      App.go("login");
    },
  };