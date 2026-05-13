// utils.js — общие функции

// Throttle: вызывает fn не чаще, чем раз в delay ms
function throttle(fn, delay) {
  let lastCall = 0;
  let timeoutId = null;
  return function (...args) {
    const now = Date.now();
    const remaining = delay - (now - lastCall);
    if (remaining <= 0) {
      lastCall = now;
      fn.apply(this, args);
    } else if (!timeoutId) {
      timeoutId = setTimeout(() => {
        lastCall = Date.now();
        timeoutId = null;
        fn.apply(this, args);
      }, remaining);
    }
  };
}

// Debounce: вызывает fn только спустя delay ms после последнего вызова
function debounce(fn, delay) {
  let timer = null;
  return function (...args) {
    clearTimeout(timer);
    timer = setTimeout(() => fn.apply(this, args), delay);
  };
}

// Безопасный вывод текста
function escapeHTML(str) {
  if (str == null) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

// Форматирование даты для отображения
function formatDate(iso) {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return '';
  const now = new Date();
  const sameDay = d.toDateString() === now.toDateString();
  const pad = (n) => String(n).padStart(2, '0');
  const time = `${pad(d.getHours())}:${pad(d.getMinutes())}`;
  if (sameDay) return time;
  return `${pad(d.getDate())}.${pad(d.getMonth() + 1)}.${d.getFullYear()} ${time}`;
}

// Глобальное состояние приложения
const State = {
  me: null,           // { id, nickname }
  users: [],          // список пользователей
  onlineIDs: new Set(),
  unreadFrom: new Set(), // ID пользователей, от которых есть непрочитанные сообщения
  activeChat: null,   // { id, nickname } собеседник
  chatOffset: 0,
  chatHasMore: true,
  chatLoading: false,
  view: 'login',      // текущая страница
};
