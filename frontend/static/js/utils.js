/* utils.js — shared helpers */

const $ = (sel, root = document) => root.querySelector(sel);
const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

function tpl(id) {
  const t = document.getElementById(id);
  if (!t) throw new Error("Template not found: " + id);
  return t.content.cloneNode(true);
}

function escapeHTML(str) {
  return String(str ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

function formatDate(value) {
  const d = new Date(value);
  if (isNaN(d.getTime())) return "";
  return d.toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function throttle(fn, ms) {
  let last = 0;
  let timer = null;
  return function throttled(...args) {
    const now = Date.now();
    const remaining = ms - (now - last);
    if (remaining <= 0) {
      last = now;
      fn.apply(this, args);
    } else if (!timer) {
      timer = setTimeout(() => {
        last = Date.now();
        timer = null;
        fn.apply(this, args);
      }, remaining);
    }
  };
}

function debounce(fn, ms) {
  let timer = null;
  return function debounced(...args) {
    clearTimeout(timer);
    timer = setTimeout(() => fn.apply(this, args), ms);
  };
}

function setError(id, message) {
  const el = document.getElementById(id);
  if (el) el.textContent = message || "";
}