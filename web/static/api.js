export class APIError extends Error {
  constructor(message, { status = 0, code = "", fields = {} } = {}) {
    super(message);
    this.name = "APIError";
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

// Для FormData браузер сам задаёт Content-Type и границу частей.
export async function api(path, { method = "GET", body } = {}) {
  const headers = { Accept: "application/json" };
  let requestBody = body;
  if (body !== undefined && !(body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
    requestBody = JSON.stringify(body);
  }

  let response;
  try {
    response = await fetch(path, { method, headers, body: requestBody, credentials: "same-origin" });
  } catch {
    throw new APIError("Не удалось подключиться. Проверьте соединение и попробуйте ещё раз.");
  }

  if (response.status === 204) return null;
  const contentType = response.headers.get("content-type") || "";
  const data = contentType.includes("application/json") ? await response.json().catch(() => null) : null;
  if (!response.ok) {
    throw new APIError(data?.error?.message || "Не удалось выполнить запрос. Попробуйте позже.", {
      status: response.status,
      code: data?.error?.code,
      fields: data?.error?.fields,
    });
  }
  if (data === null) {
    throw new APIError("Сервер вернул неожиданный ответ. Попробуйте позже.", { status: response.status });
  }
  return data;
}
