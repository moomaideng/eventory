import { z } from "zod";

export const teamPaymentReferenceSchema = z
  .string()
  .regex(/^[0-9a-f-]{36}:[0-9a-f-]{36}$/i)
  .refine((value) =>
    value.split(":").every((id) => z.uuid().safeParse(id).success)
  );

export const paymentIdSchema = z.uuid();

const optionalDate = z.union([z.literal(""), z.iso.date()]);

export const historyDatesSchema = z
  .object({ from: optionalDate, to: optionalDate })
  .refine(({ from, to }) => !from || !to || from <= to, {
    message: "End date must be on or after the start date.",
    path: ["to"],
  });
