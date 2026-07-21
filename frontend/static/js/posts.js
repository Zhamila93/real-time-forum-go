/* ==========================================================================
   posts.js — лента, создание поста, статья, реакции, комментарии
   Обложки: свой цвет у каждой категории. Категории — кнопки-теги.
   ========================================================================== */

   function strHash(s) {
    let h = 0;
    for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) | 0;
    return Math.abs(h);
  }
  
  function coverFor(p) {
    const cat = ((p.categories || [])[0] || {}).name || "The Forum";
    const hue = strHash(cat.toLowerCase()) % 360;
    const bg = `hsl(${hue}, 22%, 88%)`;
    const inkTint = `hsl(${hue}, 30%, 24%)`;
    const letter = escapeHTML(((p.title || "F").trim().charAt(0) || "F").toUpperCase());
    const catLabel = escapeHTML(cat.toUpperCase());
    const pid = "hatch-" + strHash(cat) + "-" + (p.id || 0);
  
    return `
      <div class="post-cover">
        <svg viewBox="0 0 600 400" preserveAspectRatio="xMidYMid slice" xmlns="http://www.w3.org/2000/svg">
          <defs>
            <pattern id="${pid}" width="8" height="8" patternTransform="rotate(45)" patternUnits="userSpaceOnUse">
              <rect width="8" height="8" fill="${bg}"/>
              <line x1="0" y1="0" x2="0" y2="8" stroke="rgba(18,18,18,0.05)" stroke-width="2"/>
            </pattern>
          </defs>
          <rect width="600" height="400" fill="url(#${pid})"/>
          <text x="300" y="238" text-anchor="middle"
                font-family="Libre Baskerville, Georgia, serif"
                font-size="220" font-weight="700"
                fill="${inkTint}">${letter}</text>
          <rect x="150" y="318" width="300" height="44" fill="rgba(18,18,18,0.85)"/>
          <text x="300" y="347" text-anchor="middle"
                font-family="Inter, Arial, sans-serif"
                font-size="18" letter-spacing="5"
                fill="#ffffff">${catLabel}</text>
          <rect x="7" y="7" width="586" height="386" fill="none"
                stroke="rgba(18,18,18,0.25)" stroke-width="2"/>
        </svg>
      </div>`;
  }
  
  const Posts = {
    /* ---------- Лента ---------- */
  
    async renderFeed() {
      const content = $("#content");
      content.innerHTML = "";
      content.appendChild(tpl("tpl-feed"));
  
      await Posts.fillCategorySelect();
      $("#filter-category").addEventListener("change", (e) => {
        Posts.loadFeed(e.target.value);
      });
  
      await Posts.loadFeed("");
    },
  
    async fillCategorySelect() {
      if (!State.categories.length) {
        try { State.categories = (await API.categories()) || []; }
        catch { State.categories = []; }
      }
      const sel = $("#filter-category");
      State.categories.forEach((c) => {
        const opt = document.createElement("option");
        opt.value = c.id;
        opt.textContent = c.name;
        sel.appendChild(opt);
      });
    },
  
    async loadFeed(categoryID) {
      const list = $("#posts-list");
      if (!list) return;
      list.innerHTML = "";
      let posts = [];
      try { posts = (await API.posts(categoryID)) || []; }
      catch (err) {
        list.innerHTML = `<p class="feed-empty">${escapeHTML(err.message)}</p>`;
        return;
      }
      if (!posts.length) {
        list.innerHTML = `<p class="feed-empty">Пока нет ни одной истории. Напишите первую!</p>`;
        return;
      }
      posts.forEach((p) => list.appendChild(Posts.postCard(p)));
    },
  
    postCard(p) {
      const card = document.createElement("article");
      card.className = "post-card";
      card.tabIndex = 0;
  
      const cats = (p.categories || [])
        .map((c) => `<span class="cat-tag">${escapeHTML(c.name || c)}</span>`)
        .join("");
  
      const excerpt =
        (p.content || "").length > 220
          ? p.content.slice(0, 220).trimEnd() + "…"
          : p.content || "";
  
      card.innerHTML = `
        ${coverFor(p)}
        <h3 class="post-card__title">${escapeHTML(p.title)}</h3>
        <div class="post-meta">
          <span class="post-meta__author">${escapeHTML(p.author || p.author_nickname || "")}</span>
          <span>${formatDate(p.created_at)}</span>
        </div>
        <div class="post-cats">${cats}</div>
        <p class="post-excerpt">${escapeHTML(excerpt)}</p>
      `;
  
      const open = () => App.go("post", p.id);
      card.addEventListener("click", open);
      card.addEventListener("keydown", (e) => {
        if (e.key === "Enter") open();
      });
      return card;
    },
  
    /* ---------- Новый пост ---------- */
  
    async renderNewPost() {
      const content = $("#content");
      content.innerHTML = "";
      content.appendChild(tpl("tpl-newpost"));
  
      if (!State.categories.length) {
        try { State.categories = (await API.categories()) || []; } catch {}
      }
  
      // Категории — кнопки-теги: клик выбирает/снимает
      const box = $("#categories-checkboxes");
      box.innerHTML = "";
      State.categories.forEach((c) => {
        const pill = document.createElement("button");
        pill.type = "button";
        pill.className = "cat-pill";
        pill.dataset.id = c.id;
        pill.textContent = c.name;
        pill.addEventListener("click", () => pill.classList.toggle("selected"));
        box.appendChild(pill);
      });
  
      $("#newpost-form").addEventListener("submit", async (e) => {
        e.preventDefault();
        setError("newpost-err", "");
        const f = e.target;
  
        // считаем и выбранные кнопки, и чекбоксы — на всякий случай
        const ids = [
          ...$$("#categories-checkboxes .cat-pill.selected").map((el) => Number(el.dataset.id)),
          ...$$("#categories-checkboxes input[type='checkbox']:checked").map((el) => Number(el.value)),
        ];
  
        if (!ids.length) {
          setError("newpost-err", "Выберите хотя бы одну категорию — нажмите на тег ниже.");
          return;
        }
        try {
          const created = await API.createPost(
            f.title.value.trim(),
            f.content.value.trim(),
            ids
          );
          if (created && created.id) App.go("post", created.id);
          else App.go("feed");
        } catch (err) {
          setError("newpost-err", err.message);
        }
      });
    },
  
    /* ---------- Одна статья ---------- */
  
    async renderPost(id) {
      State.currentPostID = id;
      const content = $("#content");
      content.innerHTML = "";
      content.appendChild(tpl("tpl-post-view"));
  
      try {
        const p = await API.postByID(id);
        Posts.fillArticle(p);
      } catch (err) {
        $("#single-post").innerHTML = `<p class="feed-empty">${escapeHTML(err.message)}</p>`;
        return;
      }
  
      await Posts.loadComments(id);
  
      $("#comment-form").addEventListener("submit", async (e) => {
        e.preventDefault();
        const input = $("#comment-input");
        const text = input.value.trim();
        if (!text) return;
        try {
          await API.addComment(id, text);
          input.value = "";
          await Posts.loadComments(id);
        } catch (err) {
          alert(err.message);
        }
      });
    },
  
    fillArticle(p) {
      const cats = (p.categories || [])
        .map((c) => `<span class="cat-tag">${escapeHTML(c.name || c)}</span>`)
        .join("");
  
      $("#single-post").innerHTML = `
        ${coverFor(p)}
        <div class="post-cats">${cats}</div>
        <h2 class="post-view__title">${escapeHTML(p.title)}</h2>
        <div class="post-meta">
          <span class="post-meta__author">${escapeHTML(p.author || p.author_nickname || "")}</span>
          <span>${formatDate(p.created_at)}</span>
        </div>
        <div class="body post-view__body">${escapeHTML(p.content)}</div>
        <div class="reactions-bar">
          <button type="button" class="reaction-btn ${p.my_reaction === 1 ? "active" : ""}" data-react="1">
            ▲ Like <span class="like-count">${p.likes ?? 0}</span>
          </button>
          <button type="button" class="reaction-btn ${p.my_reaction === -1 ? "active" : ""}" data-react="-1">
            ▼ Dislike <span class="dislike-count">${p.dislikes ?? 0}</span>
          </button>
        </div>
      `;
  
      $$("#single-post [data-react]").forEach((btn) => {
        btn.addEventListener("click", async () => {
          try {
            await API.react(p.id, Number(btn.dataset.react));
            const fresh = await API.postByID(p.id);
            Posts.fillArticle(fresh);
          } catch (err) {
            alert(err.message);
          }
        });
      });
    },
  
    /* ---------- Комментарии ---------- */
  
    async loadComments(postID) {
      const list = $("#comments-list");
      if (!list) return;
      list.innerHTML = "";
      let comments = [];
      try { comments = (await API.comments(postID)) || []; }
      catch { comments = []; }
  
      if (!comments.length) {
        list.innerHTML = `<p class="comments-empty">Комментариев пока нет.</p>`;
        return;
      }
      comments.forEach((c) => list.appendChild(Posts.commentEl(c)));
    },
  
    commentEl(c) {
      const el = document.createElement("div");
      el.className = "comment";
      el.innerHTML = `
        <div class="comment-meta">
          ${escapeHTML(c.author || c.author_nickname || "")} · ${formatDate(c.created_at)}
        </div>
        <div class="comment-text">${escapeHTML(c.content)}</div>
      `;
      return el;
    },
  };
  
  /* ---- Realtime-обновления форума ---- */
  WS.on("post", () => {
    if ($("#posts-list")) Posts.loadFeed($("#filter-category")?.value || "");
  });
  
  WS.on("comment", (msg) => {
    const c = msg.comment || msg;
    if (State.currentPostID && Number(c.post_id) === Number(State.currentPostID)) {
      const list = $("#comments-list");
      if (list) {
        const empty = list.querySelector(".comments-empty");
        if (empty) empty.remove();
        list.appendChild(Posts.commentEl(c));
      }
    }
  });/* ---- События форума в формате бэкенда ---- */
  WS.on("post_created", () => {
    if ($("#posts-list")) Posts.loadFeed($("#filter-category")?.value || "");
  });
  
  WS.on("comment_created", (msg) => {
    const c = msg.comment || {};
    if (State.currentPostID && Number(msg.post_id) === Number(State.currentPostID)) {
      const list = $("#comments-list");
      if (list) {
        const empty = list.querySelector(".comments-empty");
        if (empty) empty.remove();
        list.appendChild(Posts.commentEl(c));
      }
    }
  });
  
  WS.on("reaction_updated", (msg) => {
    if (State.currentPostID && Number(msg.post_id) === Number(State.currentPostID)) {
      const likeEl = document.querySelector("#single-post .like-count");
      const disEl = document.querySelector("#single-post .dislike-count");
      if (likeEl) likeEl.textContent = msg.like_count ?? likeEl.textContent;
      if (disEl) disEl.textContent = msg.dislike_count ?? disEl.textContent;
    }
  });