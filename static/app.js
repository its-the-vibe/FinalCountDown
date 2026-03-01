'use strict';

const tabsEl = document.getElementById('event-tabs');
const displayEl = document.getElementById('countdown-display');
const eventNameEl = document.getElementById('event-name');
const daysEl = document.getElementById('days');
const hoursEl = document.getElementById('hours');
const minutesEl = document.getElementById('minutes');
const secondsEl = document.getElementById('seconds');
const hoursUnitEl = hoursEl.closest('.unit');
const minutesUnitEl = minutesEl.closest('.unit');
const secondsUnitEl = secondsEl.closest('.unit');
const loadingEl = document.getElementById('loading');
const errorEl = document.getElementById('error');

let events = [];
let activeIndex = -1;
let timerID = null;

function pad(n) {
  return String(Math.floor(n)).padStart(2, '0');
}

function updateCountdown() {
  if (activeIndex < 0 || activeIndex >= events.length) return;

  const now = Date.now();
  const target = events[activeIndex].targetMs;
  const diff = target - now;

  if (diff <= 0) {
    daysEl.textContent = '0';
    hoursEl.textContent = '00';
    minutesEl.textContent = '00';
    secondsEl.textContent = '00';
    return;
  }

  const totalSeconds = diff / 1000;
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = Math.floor(totalSeconds % 60);

  daysEl.textContent = days;
  hoursEl.textContent = pad(hours);
  minutesEl.textContent = pad(minutes);
  secondsEl.textContent = pad(seconds);
}

function selectEvent(index) {
  if (activeIndex === index) return;

  // Update active tab button
  const buttons = tabsEl.querySelectorAll('.tab-btn');
  buttons.forEach((btn, i) => {
    btn.classList.toggle('active', i === index);
    btn.setAttribute('aria-selected', i === index ? 'true' : 'false');
  });

  activeIndex = index;
  eventNameEl.textContent = events[index].name;
  displayEl.classList.add('visible');

  const hasTime = events[index].hasTime;
  hoursUnitEl.hidden = !hasTime;
  minutesUnitEl.hidden = !hasTime;
  secondsUnitEl.hidden = !hasTime;

  updateCountdown();

  if (timerID !== null) clearInterval(timerID);
  timerID = setInterval(updateCountdown, 1000);
}

function renderTabs() {
  tabsEl.innerHTML = '';
  events.forEach((event, i) => {
    const btn = document.createElement('button');
    btn.className = 'tab-btn';
    btn.textContent = event.name;
    btn.setAttribute('role', 'tab');
    btn.setAttribute('aria-selected', 'false');
    btn.addEventListener('click', () => selectEvent(i));
    tabsEl.appendChild(btn);
  });
}

async function fetchEvents() {
  try {
    const res = await fetch('/api/events');
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();

    events = data.map(e => ({
      name: e.name,
      targetMs: new Date(e.target).getTime(),
      hasTime: e.has_time,
    }));

    loadingEl.hidden = true;

    if (events.length === 0) {
      errorEl.textContent = 'No events configured.';
      errorEl.hidden = false;
      return;
    }

    renderTabs();
    selectEvent(0);
  } catch (err) {
    loadingEl.hidden = true;
    errorEl.textContent = `Failed to load events: ${err.message}`;
    errorEl.hidden = false;
  }
}

fetchEvents();
