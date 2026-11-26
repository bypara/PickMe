import { api } from "./api.js";
import { fillAnimal } from "./animal-view.js";

const detail = document.querySelector("[data-animal-detail]");
const status = document.querySelector("[data-detail-status]");
const id = location.pathname.split("/").at(-1);

async function loadAnimal() {
  if (!/^[1-9][0-9]*$/.test(id)) {
    status.textContent = "Животное не найдено.";
    return;
  }
  try {
    const animal = await api(`/api/animals/${id}`);
    fillAnimal(detail, animal);
    document.title = `${animal.name} — PickMe`;
    status.hidden = true;
    detail.hidden = false;
  } catch (error) {
    status.textContent = error.message;
  }
}

loadAnimal();
