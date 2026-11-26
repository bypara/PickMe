import { api } from "./api.js";
import { showFieldErrors } from "./form-errors.js";

const dialog = document.querySelector("[data-add-dialog]");
const form = dialog.querySelector("[data-animal-form]");
const submit = form.querySelector('[type="submit"]');
const photoInput = form.elements.namedItem("photo");
const photoName = form.querySelector("[data-photo-name]");
const formError = form.querySelector("[data-animal-form-error]");
const navigationError = document.querySelector("[data-navigation-error]");
const maxPhotoSize = 5 * 1024 * 1024;

window.addEventListener("pageshow", async () => {
  if (location.pathname !== "/animals/new") return;
  try {
    const { user } = await api("/api/me");
    if (!user) {
      location.replace("/login");
      return;
    }
    if (!dialog.open) dialog.showModal();
  } catch (error) {
    navigationError.textContent = error.message;
    navigationError.hidden = false;
  }
});

dialog.addEventListener("cancel", (event) => {
  event.preventDefault();
  location.replace("/");
});

photoInput.addEventListener("change", () => {
  photoName.textContent = photoInput.files[0]?.name || "Выбрать файл";
  const message = form.querySelector("[data-error-for=photo]");
  message.hidden = true;
  photoInput.removeAttribute("aria-invalid");
});

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (submit.disabled) return;

  formError.hidden = true;
  showFieldErrors(form, {});
  const photo = photoInput.files[0];
  if (photo?.size > maxPhotoSize) {
    showFieldErrors(form, { photo: "Размер фотографии превышает 5 МиБ." });
    return;
  }

  const body = new FormData(form);
  // Пустой выбор файла может сериализоваться как часть с пустым именем.
  if (!photo) body.delete("photo");
  submit.disabled = true;
  try {
    const animal = await api("/api/animals", { method: "POST", body });
    location.replace(`/animals/${encodeURIComponent(animal.id)}`);
  } catch (error) {
    if (error.status === 401) {
      location.replace("/login");
      return;
    }
    if (error.fields && Object.keys(error.fields).length) {
      showFieldErrors(form, error.fields);
    } else {
      formError.textContent = error.message;
      formError.hidden = false;
    }
    submit.disabled = false;
  }
});
