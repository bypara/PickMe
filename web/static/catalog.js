import { api } from "./api.js";
import { fillAnimal } from "./animal-view.js";

const grid = document.querySelector("[data-animal-grid]");
const template = document.getElementById("animal-card-template");
const status = document.querySelector("[data-catalog-status]");
const summary = document.querySelector("[data-catalog-summary]");
const count = document.querySelector("[data-catalog-count]");
const plural = new Intl.PluralRules("ru");
const number = new Intl.NumberFormat("ru");

function renderCatalog(animals) {
  const cards = document.createDocumentFragment();
  for (const animal of animals) {
    const item = template.content.cloneNode(true);
    const link = item.querySelector("a");
    link.href = `/animals/${encodeURIComponent(animal.id)}`;
    link.setAttribute("aria-label", animal.name);
    fillAnimal(item, animal);
    cards.append(item);
  }
  grid.replaceChildren(cards);
  count.textContent = `${number.format(animals.length)} ${plural.select(animals.length) === "one" ? "животное" : "животных"}`;
  summary.hidden = false;
  status.textContent = "В каталоге пока нет животных.";
  status.hidden = animals.length > 0;
  status.classList.toggle("catalog-empty", animals.length === 0);
}

// Обновляем список и при возврате браузера на главную страницу из истории.
window.addEventListener("pageshow", async () => {
  try {
    renderCatalog(await api("/api/animals"));
  } catch (error) {
    status.textContent = error.message;
    status.classList.remove("catalog-empty");
    status.hidden = false;
  }
});
