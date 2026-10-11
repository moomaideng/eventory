import { z } from "zod";

export const paymentIdSchema = z.uuid();

const optionalDate = z.union([z.literal(""), z.iso.date()]);

export const historyDatesSchema = z
  .object({ from: optionalDate, to: optionalDate })
  .refine(({ from, to }) => !from || !to || from <= to, {
    message: "End date must be on or after the start date.",
    path: ["to"],
  });
