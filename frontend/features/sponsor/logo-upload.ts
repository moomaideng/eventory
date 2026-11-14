import { createClient } from "@/lib/client";

export const SPONSOR_LOGO_BUCKET = "sponsor-logos";
export const SPONSOR_LOGO_MAX_BYTES = 5 * 1024 * 1024;

const acceptedLogoTypes = new Set(["image/png", "image/jpeg", "image/webp"]);

export function validateSponsorLogo(file: File): string | undefined {
  if (!acceptedLogoTypes.has(file.type)) {
    return "Choose a PNG, JPEG, or WebP image.";
  }
  if (file.size > SPONSOR_LOGO_MAX_BYTES) {
    return "Logo image must be 5 MB or smaller.";
  }
}

export async function uploadSponsorLogo(userID: string, file: File) {
  const validationError = validateSponsorLogo(file);
  if (validationError) throw new Error(validationError);

  const path = `${userID}/logo`;
  const supabase = createClient();
  const { error } = await supabase.storage
    .from(SPONSOR_LOGO_BUCKET)
    .upload(path, file, {
      contentType: file.type,
      upsert: true,
    });
  if (error) throw new Error(`Logo upload failed: ${error.message}`);

  const { data } = supabase.storage
    .from(SPONSOR_LOGO_BUCKET)
    .getPublicUrl(path);
  return `${data.publicUrl}?v=${crypto.randomUUID()}`;
}
