import { z } from "zod";

// Mirror cocotola-question service/study/study.go MaxSelectedChoiceIDsCount / MaxChoiceIDLength.
export const MAX_SELECTED_CHOICE_IDS_COUNT = 40;
export const MAX_CHOICE_ID_LENGTH = 100;

// Mirror cocotola-question domain/question/content_ordering.go MaxOrderingSegments /
// MaxOrderingTextLength, which service/study/study.go validateOrderingRegions also uses.
export const MAX_ORDERING_SEGMENTS = 40;
export const MAX_ORDERING_TEXT_BYTES = 500;

const textEncoder = new TextEncoder();

const answerPayloadSchema = z.discriminatedUnion("type", [
  z.object({
    type: z.literal("multiple_choice"),
    selectedChoiceIds: z
      .array(z.string().max(MAX_CHOICE_ID_LENGTH))
      .max(MAX_SELECTED_CHOICE_IDS_COUNT),
  }),
  z.object({
    type: z.literal("word_fill"),
    correct: z.boolean(),
  }),
  z.object({
    type: z.literal("ordering"),
    regions: z
      .array(
        z.array(
          z.string().refine((text) => textEncoder.encode(text).length <= MAX_ORDERING_TEXT_BYTES),
        ),
      )
      .max(MAX_ORDERING_SEGMENTS)
      .refine(
        (regions) =>
          regions.reduce((total, region) => total + region.length, 0) <= MAX_ORDERING_SEGMENTS,
      ),
  }),
]);

export type AnswerPayload = z.infer<typeof answerPayloadSchema>;

export function parseAnswerPayload(raw: string): AnswerPayload | null {
  let json: unknown;
  try {
    json = JSON.parse(raw);
  } catch {
    return null;
  }
  const result = answerPayloadSchema.safeParse(json);
  return result.success ? result.data : null;
}
