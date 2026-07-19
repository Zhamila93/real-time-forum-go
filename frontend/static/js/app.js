/* app.js — application state + single-page router */

const State = {
    user: null,
    users: [],
    categories: [],
    currentPostID: null,
    unread: {},
    chat: { peer: null, offset: 0, done: false, loading: false },
  
    reset() {
      this.user = null;
      this.users = [];
      this.categories = [];
      this.currentPostID = null;
      this.unread = {};
      this.chat = { peer: null, offset: 0, done: false, loading: false };
    },
  };
  
  const App = {
    shellRendered: false,
  
    go(view, arg) {
      switch (view) {
        case "login":
          App.shellRendered = false;
          Auth.renderLogin();
          break;
        case "register":
          App.shellRendered = false;
          Auth.renderRegister();
          break;
        case "feed":
          App.ensureShell();
          State.currentPostID = null;
          Posts.renderFeed();
          break;
        case "newpost":
          App.ensureShell();
          State.currentPostID = null;
          Posts.renderNewPost();
          break;
        case "post":
          App.ensureShell();
          Posts.renderPost(arg);
          break;
        default:
          App.go(State.user ? "feed" : "login");
      }
    },
  
    ensureShell() {
      if (App.shellRendered) return;
      const app = $("#app");
      app.innerHTML = "";
      app.appendChild(tpl("tpl-main"));
      App.shellRendered = true;
  
      $("#me-nick").textContent = State.user ? State.user.nickname : "";
      $("#logout-btn").addEventListener("click", () => Auth.logout());
  
      Chat.init();
      Chat.loadUsers();
    },
  
    async start() {
      State.user = await API.me();
      App.shellRendered = false;
      App.ensureShell();
      WS.connect();
      App.go("feed");
    },
  
    async init() {
      document.body.addEventListener("click", (e) => {
        const go = e.target.closest("[data-go]");
        if (!go) return;
        e.preventDefault();
        App.go(go.dataset.go);
      });
  
      try {
        await App.start();
      } catch {
        App.go("login");
      }
    },
  };
  
  document.addEventListener("DOMContentLoaded", () => App.init());