import { api } from "./api.js";
import { showFieldErrors } from "./form-errors.js";

const form = document.querySelector("[data-auth-form]");
const mode = form.dataset.authForm;
const submit = form.querySelector('[type="submit"]');
const formError = form.querySelector("[data-form-error]");
const registrationNotice = document.querySelector("[data-registration-notice]");
if (registrationNotice && new URLSearchParams(location.search).get("registered") === "1") {
  registrationNotice.hidden = false;
}

for (const button of form.querySelectorAll("[data-password-toggle]")) {
  const input = document.getElementById(button.getAttribute("aria-controls"));
  button.addEventListener("click", () => {
    const show = input.type === "password";
    input.type = show ? "text" : "password";
    button.toggleAttribute("data-visible", show);
    button.setAttribute("aria-label", show ? "Скрыть пароль" : "Показать пароль");
  });
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (submit.disabled) return;
  formError.hidden = true;
  showFieldErrors(form, {});

  const fields = new FormData(form);
  const body = { email: fields.get("email"), password: fields.get("password") };
  if (mode === "register") {
    if (body.password !== fields.get("password_confirmation")) {
      showFieldErrors(form, { password_confirmation: "Пароли не совпадают." });
      return;
    }
    body.full_name = fields.get("full_name");
  }

  submit.disabled = true;
  try {
    await api(`/api/${mode}`, { method: "POST", body });
    location.replace(mode === "register" ? "/login?registered=1" : "/");
  } catch (error) {
    if (error.code === "already_authenticated") {
      location.replace("/");
      return;
    }
    if (error.code === "email_taken") {
      showFieldErrors(form, { email: error.message });
    } else if (error.fields && Object.keys(error.fields).length) {
      showFieldErrors(form, error.fields);
    } else {
      formError.textContent = error.message;
      formError.hidden = false;
    }
    submit.disabled = false;
  }
});

// pageshow срабатывает и при возврате браузера к странице из истории.
window.addEventListener("pageshow", async () => {
  try {
    if ((await api("/api/me")).user) location.replace("/");
  } catch {
    // Форма покажет ошибку запроса при отправке, даже если фоновая проверка сессии недоступна.
  }
});

submit.disabled = false;
