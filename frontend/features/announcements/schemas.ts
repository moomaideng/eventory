import { z } from "zod";

// Same limits as the API, which counts characters (so Thai text is not penalised).
export const ANNOUNCEMENT_TITLE_MAX = 150;
export const ANNOUNCEMENT_BODY_MAX = 2000;

export const announcementSchema = z.object({
  title: z
    .string({ message: "Title is required." })
    .trim()
    .min(1, "Enter a title.")
    .max(
      ANNOUNCEMENT_TITLE_MAX,
      `Title must be ${ANNOUNCEMENT_TITLE_MAX} characters or fewer.`
    ),
  body: z
    .string({ message: "Message is required." })
    .trim()
    .min(1, "Enter a message.")
    .max(
      ANNOUNCEMENT_BODY_MAX,
      `Message must be ${ANNOUNCEMENT_BODY_MAX} characters or fewer.`
    ),
});

export type AnnouncementInput = z.infer<typeof announcementSchema>;
