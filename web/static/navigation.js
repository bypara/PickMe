import { api } from "./api.js";

const profile = document.querySelector("[data-user-profile]");
const name = document.querySelector("[data-user-name]");
const initials = document.querySelector("[data-user-initials]");
const login = document.querySelector("[data-login-link]");
const logout = document.querySelector("[data-logout]");
const addAnimal = document.querySelector("[data-add-animal]");
const message = document.querySelector("[data-navigation-error]");
let navigationCheck = 0;

function renderUser(user) {
  login.hidden = Boolean(user);
  logout.hidden = !user;
  addAnimal.hidden = !user;
  profile.hidden = !user;
  if (user) {
    const displayName = user.full_name || user.email;
    const words = user.full_name ? user.full_name.split(/\s+/, 2) : [displayName];
    name.textContent = displayName;
    initials.textContent = words.map((word) => Array.from(word)[0]).join("").toLocaleUpperCase("ru");
  } else {
    name.textContent = "";
    initials.textContent = "";
  }
}

window.addEventListener("pageshow", async () => {
  if (logout.disabled) return;
  const check = ++navigationCheck;
  try {
    const { user } = await api("/api/me");
    if (check === navigationCheck) renderUser(user);
  } catch {
    if (check === navigationCheck) renderUser(null);
  }
});

logout.addEventListener("click", async () => {
  if (logout.disabled) return;
  // После выхода учитываем только запросы состояния, запущенные позже.
  navigationCheck++;
  logout.disabled = true;
  message.hidden = true;
  try {
    await api("/api/logout", { method: "POST" });
    renderUser(null);
    login.focus();
  } catch (error) {
    message.textContent = error.message;
    message.hidden = false;
  } finally {
    logout.disabled = false;
  }
});
