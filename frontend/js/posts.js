// posts.js — лента, новый пост, просмотр поста с комментариями

const Posts = {
  feedCategory: '',

  async renderFeed() {
    const tpl = document.getElementById('tpl-feed').content.cloneNode(true);
    const content = document.getElementById('content');
    content.innerHTML = '';
    content.appendChild(tpl);
    State.viewingPostId = null;

    try {
      const cats = await API.categories();
      const sel = document.getElementById('filter-category');
      (cats || []).forEach((c) => {
        const opt = document.createElement('option');
        opt.value = c.name;
        opt.textContent = c.name;
        sel.appendChild(opt);
      });
      sel.addEventListener('change', () => {
        this.feedCategory = sel.value;
        this.loadPosts(this.feedCategory);
      });
    } catch (e) {}

    await this.loadPosts(this.feedCategory);
  },

  async loadPosts(category) {
    this.feedCategory = category || '';
    const list = document.getElementById('posts-list');
    if (!list) return;

    list.innerHTML = '<p class="feed-empty">Загрузка...</p>';
    try {
      const posts = await API.posts(category);
      if (!posts || posts.length === 0) {
        list.innerHTML = '<p class="feed-empty">Пока нет постов.</p>';
        return;
      }
      list.innerHTML = '';
      posts.forEach((p) => list.appendChild(this.buildPostCard(p)));
    } catch (err) {
      list.innerHTML = `<p class="err">${escapeHTML(err.message)}</p>`;
    }
  },

  postMatchesFeedFilter(post) {
    if (!this.feedCategory) return true;
    return (post.categories || []).includes(this.feedCategory);
  },

  buildPostCard(p) {
    const card = document.createElement('article');
    card.className = 'post-card';
    card.dataset.postId = p.id;
    card.innerHTML = `
      <h3 class="post-card__title">${escapeHTML(p.title)}</h3>
      <div class="post-meta">
        <span class="post-meta__author">@${escapeHTML(p.nickname)}</span>
        <span class="post-meta__date">${formatDate(p.created_at)}</span>
        <span class="post-meta__likes">👍 ${p.like_count || 0}</span>
        <span class="post-meta__dislikes">👎 ${p.dislike_count || 0}</span>
        <span class="post-meta__comments">💬 ${p.comment_count || 0}</span>
      </div>
      <div class="post-cats">
        ${(p.categories || []).map((c) => `<span class="cat-tag">${escapeHTML(c)}</span>`).join('')}
      </div>
      <p class="post-excerpt">${escapeHTML((p.content || '').slice(0, 160))}${(p.content || '').length > 160 ? '…' : ''}</p>
    `;
    card.addEventListener('click', () => Posts.renderPostView(p.id));
    return card;
  },

  prependPost(post) {
    const list = document.getElementById('posts-list');
    if (!list || !this.postMatchesFeedFilter(post)) return;
    if (list.querySelector(`[data-post-id="${post.id}"]`)) return;

    const empty = list.querySelector('.feed-empty');
    if (empty) empty.remove();

    list.prepend(this.buildPostCard(post));
  },

  onPostCreated(data) {
    const post = data.post;
    if (!post) return;
    this.prependPost(post);
  },

  bumpCommentCount(postId) {
    const card = document.querySelector(`#posts-list .post-card[data-post-id="${postId}"]`);
    if (!card) return;
    const el = card.querySelector('.post-meta__comments');
    if (!el) return;
    const n = parseInt(el.textContent.replace(/\D/g, ''), 10) || 0;
    el.textContent = `💬 ${n + 1}`;
  },

  onCommentCreated(data) {
    const comment = data.comment;
    if (!comment) return;

    this.bumpCommentCount(comment.post_id);

    if (Number(State.viewingPostId) !== Number(comment.post_id)) return;

    const cList = document.getElementById('comments-list');
    if (!cList) return;
    if (cList.querySelector(`[data-comment-id="${comment.id}"]`)) return;

    const empty = cList.querySelector('.comments-empty');
    if (empty) empty.remove();

    cList.appendChild(this.renderComment(comment));
  },

  async renderNewPost() {
    const tpl = document.getElementById('tpl-newpost').content.cloneNode(true);
    const content = document.getElementById('content');
    content.innerHTML = '';
    content.appendChild(tpl);
    State.viewingPostId = null;

    try {
      const cats = await API.categories();
      const box = document.getElementById('categories-checkboxes');
      (cats || []).forEach((c) => {
        const lbl = document.createElement('label');
        lbl.innerHTML = `<input type="checkbox" value="${escapeHTML(c.name)}" /> ${escapeHTML(c.name)}`;
        box.appendChild(lbl);
      });
    } catch (e) {}

    document.getElementById('newpost-form').addEventListener('submit', async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      const cats = Array.from(document.querySelectorAll('#categories-checkboxes input:checked')).map(
        (i) => i.value,
      );
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
    State.viewingPostId = postID;
    const tpl = document.getElementById('tpl-post-view').content.cloneNode(true);
    const content = document.getElementById('content');
    content.innerHTML = '';
    content.appendChild(tpl);

    try {
      const data = await API.postByID(postID);
      const p = data.post;
      const post = document.getElementById('single-post');
      post.dataset.postId = postID;
      post.innerHTML = `
        <div class="post-meta">
          <span class="post-meta__author">@${escapeHTML(p.nickname)}</span>
          <span class="post-meta__date">${formatDate(p.created_at)}</span>
        </div>
        <h2 class="post-view__title">${escapeHTML(p.title)}</h2>
        <div class="post-cats">
          ${(p.categories || []).map((c) => `<span class="cat-tag">${escapeHTML(c)}</span>`).join('')}
        </div>
        <div class="post-view__body">${escapeHTML(p.content)}</div>
        <div class="reactions-bar" data-post-id="${postID}">
          <button type="button" class="reaction-btn like-btn ${p.user_reaction === 'like' ? 'active' : ''}" data-reaction="like">👍 <span class="like-count">${p.like_count || 0}</span></button>
          <button type="button" class="reaction-btn dislike-btn ${p.user_reaction === 'dislike' ? 'active' : ''}" data-reaction="dislike">👎 <span class="dislike-count">${p.dislike_count || 0}</span></button>
        </div>
      `;
      post.querySelectorAll('.reaction-btn').forEach((btn) => {
        btn.addEventListener('click', async (e) => {
          e.stopPropagation();
          try {
            const updated = await API.setReaction(postID, btn.dataset.reaction);
            Posts.applyReactionUpdate(updated);
          } catch (err) {
            alert(err.message);
          }
        });
      });

      const cList = document.getElementById('comments-list');
      cList.innerHTML = '';
      if (!data.comments || data.comments.length === 0) {
        cList.innerHTML = '<p class="comments-empty">Нет комментариев.</p>';
      } else {
        data.comments.forEach((c) => cList.appendChild(this.renderComment(c)));
      }

      document.getElementById('comment-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const ta = e.target.querySelector('textarea');
        const text = ta.value.trim();
        if (!text) return;
        try {
          await API.createComment({ post_id: postID, content: text });
          ta.value = '';
        } catch (err) {
          alert(err.message);
        }
      });
    } catch (err) {
      content.innerHTML = `<p class="err">${escapeHTML(err.message)}</p>`;
    }
  },

  applyReactionUpdate(data) {
    const bar = document.querySelector(`.reactions-bar[data-post-id="${data.post_id}"]`);
    if (bar) {
      bar.querySelector('.like-count').textContent = data.like_count;
      bar.querySelector('.dislike-count').textContent = data.dislike_count;
      bar.querySelector('.like-btn').classList.toggle('active', data.reaction === 'like');
      bar.querySelector('.dislike-btn').classList.toggle('active', data.reaction === 'dislike');
    }
    const card = document.querySelector(`#posts-list .post-card[data-post-id="${data.post_id}"]`);
    if (card) {
      const meta = card.querySelector('.post-meta');
      if (meta) {
        const likes = meta.querySelector('.post-meta__likes');
        const dislikes = meta.querySelector('.post-meta__dislikes');
        if (likes) likes.textContent = `👍 ${data.like_count}`;
        if (dislikes) dislikes.textContent = `👎 ${data.dislike_count}`;
      }
    }
  },

  renderComment(c) {
    const d = document.createElement('article');
    d.className = 'comment';
    d.dataset.commentId = c.id;
    d.innerHTML = `
      <div class="comment-meta">@${escapeHTML(c.nickname)} • ${formatDate(c.created_at)}</div>
      <div class="comment-text">${escapeHTML(c.content)}</div>
    `;
    return d;
  },
};
