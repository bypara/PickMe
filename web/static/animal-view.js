export function showAnimalPhoto(container, animal) {
  if (!animal.photo_url) return;

  const image = container.querySelector("[data-animal-photo]");
  const placeholder = container.querySelector("[data-photo-placeholder]");
  image.addEventListener("error", () => {
    image.hidden = true;
    placeholder.hidden = false;
  }, { once: true });

  image.alt = animal.name;
  image.src = animal.photo_url;
  image.hidden = false;
  placeholder.hidden = true;
}

export function fillAnimal(container, animal) {
  container.querySelector("[data-animal-name]").textContent = animal.name;
  container.querySelector("[data-animal-breed]").textContent = animal.breed;
  container.querySelector("[data-animal-location]").textContent = animal.found_location;
  showAnimalPhoto(container, animal);
}
