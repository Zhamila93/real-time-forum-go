// posts.js — лента, новый пост, просмотр поста с комментариями

const Posts = {
  async renderFeed() {
    const tpl = document.getElementById('tpl-feed').content.cloneNode(true);
    const content = document.getElementById('content');
    content.innerHTML = '';
    content.appendChild(tpl);

    // загружаем категории в фильтр
    try {
      const cats = await API.categories();
      const sel = document.getElementById('filter-category');
      (cats || []).forEach(c => {
        const opt = document.createElement('option');
        opt.value = c.name; opt.textContent = c.name;
        sel.appendChild(opt);
      });
      sel.addEventListener('change', () => this.loadPosts(sel.value));
    } catch (e) {}

    await this.loadPosts('');
  },

  async loadPosts(category) {
    const list = document.getElementById('posts-list');
    list.innerHTML = '<p style="color:var(--fg-dim)">Загрузка...</p>';
    try {
      const posts = await API.posts(category);
      if (!posts || posts.length === 0) {
        list.innerHTML = '<p style="color:var(--fg-dim)">Пока нет постов.</p>';
        return;
      }
      list.innerHTML = '';
      posts.forEach(p => {
        const card = document.createElement('div');
        card.className = 'post-card';
        card.innerHTML = `
          <h3>${escapeHTML(p.title)}</h3>
          <div class="post-meta">
            <span>@${escapeHTML(p.nickname)}</span>
            <span>${formatDate(p.created_at)}</span>
            <span>💬 ${p.comment_count}</span>
          </div>
          <div class="post-cats">
            ${(p.categories || []).map(c => `<span class="cat-tag">${escapeHTML(c)}</span>`).join('')}
          </div>
          <div class="post-excerpt">${escapeHTML((p.content || '').slice(0, 160))}${(p.content || '').length > 160 ? '…' : ''}</div>
        `;
        card.addEventListener('click', () => Posts.renderPostView(p.id));
        list.appendChild(card);
      });
    } catch (err) {
      list.innerHTML = `<p class="err">${escapeHTML(err.message)}</p>`;
    }
  },

  async renderNewPost() {
    const tpl = document.getElementById('tpl-newpost').content.cloneNode(true);
    const content = document.getElementById('content');
    content.innerHTML = '';
    content.appendChild(tpl);

    // загружаем категории
    try {
      const cats = await API.categories();
      const box = document.getElementById('categories-checkboxes');
      (cats || []).forEach(c => {
        const lbl = document.createElement('label');
        lbl.innerHTML = `<input type="checkbox" value="${escapeHTML(c.name)}" /> ${escapeHTML(c.name)}`;
        box.appendChild(lbl);
      });
    } catch (e) {}

    document.getElementById('newpost-form').addEventListener('submit', async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      const cats = Array.from(document.querySelectorAll('#categories-checkboxes input:checked'))
                        .map(i => i.value);
      const body = {
        title: fd.get('title'),
        content: fd.get('content'),
        categories: cats,
      };
      try {
        await API.createPost(body);
        App.go('feed');
      } catch (err) {
        document.getElementById('newpost-err').textContent = err.message;
      }
    });
  },

  async renderPostView(postID) {
    const tpl = document.getElementById('tpl-post-view').content.cloneNode(true);
    const content = document.getElementById('content');
    content.innerHTML = '';
    content.appendChild(tpl);

    try {
      const data = await API.postByID(postID);
      const p = data.post;
      const post = document.getElementById('single-post');
      post.innerHTML = `
        <div class="post-meta">
          <span>@${escapeHTML(p.nickname)}</span>
          <span>${formatDate(p.created_at)}</span>
        </div>
        <h2>${escapeHTML(p.title)}</h2>
        <div class="post-cats">
          ${(p.categories || []).map(c => `<span class="cat-tag">${escapeHTML(c)}</span>`).join('')}
        </div>
        <div class="body">${escapeHTML(p.content)}</div>
      `;

      const cList = document.getElementById('comments-list');
      cList.innerHTML = '';
      if (!data.comments || data.comments.length === 0) {
        cList.innerHTML = '<p style="color:var(--fg-dim)">Нет комментариев.</p>';
      } else {
        data.comments.forEach(c => cList.appendChild(this.renderComment(c)));
      }

      document.getElementById('comment-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const ta = e.target.querySelector('textarea');
        const text = ta.value.trim();
        if (!text) return;
        try {
          await API.createComment({ post_id: postID, content: text });
          ta.value = '';
          // перерисуем
          this.renderPostView(postID);
        } catch (err) {
          alert(err.message);
        }
      });
    } catch (err) {
      content.innerHTML = `<p class="err">${escapeHTML(err.message)}</p>`;
    }
  },

  renderComment(c) {
    const d = document.createElement('div');
    d.className = 'comment';
    d.innerHTML = `
      <div class="comment-meta">@${escapeHTML(c.nickname)} • ${formatDate(c.created_at)}</div>
      <div class="comment-text">${escapeHTML(c.content)}</div>
    `;
    return d;
  },
};
