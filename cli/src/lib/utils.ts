import color from "color";

// Returns the error message is error is an Error, otherwise returns the
// string representation of error.
export function unwrapError(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}

// Normalizes a color value to a hex string. If the color has an alpha value,
// it will be mixed with the background color to get a more accurate color
// value that can be used to search for the color in the theme.
export function normalizeColor(
  value?: string | null,
  background?: string
): string | undefined {
  if (!value) {
    return;
  }

  let c = color(value);

  if (c.alpha() < 1 && background) {
    c = c.alpha(1).mix(color.lab(background), 1 - c.alpha());
  }

  return c.hex();
}

// Applies the alpha value to the color and returns the new string.
export const alpha = (value: string, alpha: number) =>
  color ? color(value).alpha(alpha).toString() : undefined;
