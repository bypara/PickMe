export function showFieldErrors(form, fields) {
  let firstInvalid;
  for (const message of form.querySelectorAll("[data-error-for]")) {
    const name = message.dataset.errorFor;
    const input = form.elements.namedItem(name);
    const text = fields[name] || "";
    message.textContent = text;
    message.hidden = !text;
    if (text) {
      input.setAttribute("aria-invalid", "true");
      firstInvalid ??= input;
    } else {
      input.removeAttribute("aria-invalid");
    }
  }
  firstInvalid?.focus();
}
