/* api.js — all HTTP calls. Endpoint paths live in API.routes. */

const API = {
    routes: {
      register:   "/api/register",
      login:      "/api/login",
      logout:     "/api/logout",
      me:         "/api/me",
      categories: "/api/categories",
      posts:      "/api/posts",
      post:       (id) => `/api/posts/${id}`,
      comments:   (id) => `/api/posts/${id}/comments`,
      react:      (id) => `/api/posts/${id}/reaction`,
      users:      "/api/users",
      messages:   "/api/messages",
    },
  
    async request(method, url, body) {
      const opts = {
        method,
        credentials: "include",
        headers: {},
      };
      if (body !== undefined) {
        opts.headers["Content-Type"] = "application/json";
        opts.body = JSON.stringify(body);
      }
  
      const res = await fetch(url, opts);
  
      let data = null;
      const text = await res.text();
      if (text) {
        try { data = JSON.parse(text); } catch { data = text; }
      }
  
      if (!res.ok) {
        const msg =
          (data && (data.error || data.message)) ||
          `Request failed (${res.status})`;
        const err = new Error(msg);
        err.status = res.status;
        throw err;
      }
      return data;
    },
  
    get(url)        { return this.request("GET", url); },
    post(url, body) { return this.request("POST", url, body); },
  
    register(payload) { return this.post(this.routes.register, payload); },
  
    login(identifier, password) {
      return this.post(this.routes.login, { identifier, password });
    },
  
    logout() { return this.post(this.routes.logout); },
  
    me() { return this.get(this.routes.me); },
  
    categories() { return this.get(this.routes.categories); },
  
    posts(categoryID) {
      const q = categoryID ? `?category=${encodeURIComponent(categoryID)}` : "";
      return this.get(this.routes.posts + q);
    },
  
    createPost(title, content, categoryIDs) {
      // Бэкенд ждёт НАЗВАНИЯ категорий строками — конвертируем id → имена
      const names = categoryIDs.map((id) => {
        const c = State.categories.find((x) => Number(x.id) === Number(id));
        return c ? String(c.name) : String(id);
      });
      return this.post(this.routes.posts, {
        title,
        content,
        categories: names,
        category_ids: categoryIDs,
      });
    },
  
    postByID(id) { return this.get(this.routes.post(id)); },
  
    comments(id) { return this.get(this.routes.comments(id)); },
  
    addComment(id, content) {
      return this.post("/api/comments", { post_id: Number(id), content });
    },
  
    react(id, value) { return this.post(this.routes.react(id), { value }); },
  
    users() { return this.get(this.routes.users); },
  
    messages(withID, offset = 0, limit = 10) {
      const q = `?with=${withID}&offset=${offset}&limit=${limit}`;
      return this.get(this.routes.messages + q);
    },
  };