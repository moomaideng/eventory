import { z } from "zod";
import type { components } from "@/lib/api/schema";

export type RegistrationForm = components["schemas"]["RegistrationForm"];
export type RegistrationQuestion =
  components["schemas"]["RegistrationQuestion"];
export type RegistrationSubmission =
  components["schemas"]["RegistrationSubmission"];

export const defaultConsentNotice =
  "I agree to share my registration answers and uploaded files with the tournament organizer for participant review and tournament administration.";

const questionConfigSchema = z
  .object({
    id: z.string().regex(/^[a-zA-Z0-9_-]{1,64}$/),
    label: z.string().trim().min(1, "Enter a question.").max(160),
    type: z.enum(["TEXT", "DROPDOWN", "FILE"]),
    required: z.boolean(),
    options: z.array(z.string()).max(100),
    pattern: z.string().max(256),
    helpText: z.string().max(500),
    allowedTypes: z.array(
      z.enum(["application/pdf", "image/png", "image/jpeg"])
    ),
    maxFileBytes: z.number().int(),
  })
  .superRefine((q, ctx) => {
    if (
      q.type === "DROPDOWN" &&
      (!q.options.length ||
        q.options.some((o) => !o.trim()) ||
        new Set(q.options).size !== q.options.length)
    ) {
      ctx.addIssue({
        code: "custom",
        path: ["options"],
        message: "Add unique options, one per line.",
      });
    }
    if (q.type === "FILE" && !q.allowedTypes.length) {
      ctx.addIssue({
        code: "custom",
        path: ["allowedTypes"],
        message: "Choose at least one file type.",
      });
    }
    if (
      q.type === "FILE" &&
      (q.maxFileBytes < 1 || q.maxFileBytes > 5 * 1024 * 1024)
    ) {
      ctx.addIssue({
        code: "custom",
        path: ["maxFileBytes"],
        message: "Choose a size limit between 1 byte and 5 MB.",
      });
    }
  });

export const registrationConfigSchema = z.object({
  questions: z.array(questionConfigSchema).max(30),
  consentNotice: z.string().trim().min(1, "Enter a consent notice.").max(2000),
});
export type RegistrationConfig = z.infer<typeof registrationConfigSchema>;

export function newRegistrationQuestion(): RegistrationConfig["questions"][number] {
  return {
    id: crypto.randomUUID(),
    label: "",
    type: "TEXT",
    required: true,
    options: [],
    pattern: "",
    helpText: "",
    allowedTypes: [],
    maxFileBytes: 0,
  };
}

export function answerSchema(questions: RegistrationQuestion[]) {
  return z
    .object({
      values: z.record(z.string(), z.string()),
      files: z.record(z.string(), z.instanceof(File)),
      consent: z.boolean(),
    })
    .superRefine((data, ctx) => {
      for (const q of questions) {
        const value = data.values[q.id]?.trim() ?? "";
        const file = data.files[q.id];
        let message = "";
        if (q.type === "FILE") {
          if (q.required && !file) message = "Choose a file.";
          else if (file && (file.size === 0 || file.size > q.maxFileBytes))
            message = "The file exceeds the allowed size or is empty.";
          else if (file && !(q.allowedTypes ?? []).includes(file.type))
            message = "Choose an allowed file type.";
        } else if (q.required && !value) message = "This answer is required.";
        else if (
          value &&
          q.type === "DROPDOWN" &&
          !(q.options ?? []).includes(data.values[q.id])
        )
          message = "Choose an available option.";
        else if (value.length > 2000) message = "Use at most 2000 characters.";
        // The backend checks organizer patterns using Go's RE2 syntax.
        if (message) ctx.addIssue({ code: "custom", path: [q.id], message });
      }
      if (
        Object.values(data.files).reduce(
          (total, file) => total + file.size,
          0
        ) >
        10 * 1024 * 1024
      ) {
        ctx.addIssue({
          code: "custom",
          path: ["files"],
          message: "Upload at most 10 MB in total.",
        });
      }
      if (!data.consent)
        ctx.addIssue({
          code: "custom",
          path: ["consent"],
          message: "Accept the participant-data notice to continue.",
        });
    });
}

export async function readRegistrationFile(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result).split(",")[1]);
    reader.onerror = () =>
      reject(
        new Error("Could not read the selected file. Please choose it again.")
      );
    reader.readAsDataURL(file);
  });
}
