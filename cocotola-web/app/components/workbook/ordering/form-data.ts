import { parseTagsFormData } from "../schemas";
import { parseOrderingNotation } from "./notation";

function requiredField(formData: FormData, name: string): string {
  const value = formData.get(name);
  if (typeof value !== "string" || !value.trim()) {
    throw new Response(`${name} is required`, { status: 400 });
  }
  return value;
}

export function parseOrderingFormData(formData: FormData) {
  const sourceText = requiredField(formData, "sourceText");
  const sourceLang = requiredField(formData, "sourceLang");
  const notation = requiredField(formData, "targetText");
  const targetLang = requiredField(formData, "targetLang");

  const parsed = parseOrderingNotation(notation);
  if (!parsed.ok) {
    throw new Response("targetText is invalid", { status: 400 });
  }

  const explanation = formData.get("explanation");
  const content = JSON.stringify({
    source: { text: sourceText, lang: sourceLang },
    target: { lang: targetLang, segments: parsed.segments },
    ...(typeof explanation === "string" && explanation.trim() ? { explanation } : {}),
  });

  return { content, tags: parseTagsFormData(formData) };
}
