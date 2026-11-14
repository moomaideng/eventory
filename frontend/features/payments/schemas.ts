import { z } from "zod";

export const teamPaymentReferenceSchema = z
  .string()
  .regex(/^[0-9a-f-]{36}:[A-Za-z0-9]{8}:[0-9a-f-]{36}$/i)
  .refine((value) => {
    const [tournamentId, , attemptId] = value.split(":");
    return (
      z.uuid().safeParse(tournamentId).success &&
      z.uuid().safeParse(attemptId).success
    );
  });

export const paymentIdSchema = z.uuid();

const optionalDate = z.union([z.literal(""), z.iso.date()]);

export const historyDatesSchema = z
  .object({ from: optionalDate, to: optionalDate })
  .refine(({ from, to }) => !from || !to || from <= to, {
    message: "End date must be on or after the start date.",
    path: ["to"],
  });
