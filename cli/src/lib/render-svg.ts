import { encode } from "html-entities";
import escapeHtml from "escape-html";
import { Colors, Token } from "./parse-theme";
import { Language } from "./languages";

export interface SvgOptions {
  rounded?: boolean;
  monoFontFamily?: string;
  sansSerifFontFamily?: string;
}

const titleFontSize = 11;
const titleLetterSpacing = 0.5;
const tokenFontSize = 12.5;
const tokenLineHeight = 17;
const tokenXOffset = 50;
const tokenYOffset = 58;
const editorLetterSpacing = 0.2;
const badgeFontSize = 6.5;

export const defaultOptions = {
  rounded: true,
  monoFontFamily:
    "'SFMono-Regular',Consolas,'Liberation Mono',Menlo,Courier,monospace",
  sansSerifFontFamily:
    "-apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif",
};

export default function renderSvg(
  displayName: string,
  colors: Colors,
  language: Language,
  tokens: Token[][],
  opts = defaultOptions
) {
  let svg = `<svg viewBox="0 0 460 331" xmlns="http://www.w3.org/2000/svg" font-family="${opts.sansSerifFontFamily}"  letter-spacing="${titleLetterSpacing}">`;

  // Editor
  const rx = opts.rounded ? 'rx="8"' : "";
  svg += `<rect width="460" height="331" ${rx} fill="${colors.editorBackground}" />`;

  // Activity bar
  svg += `<path d="M0 20H36.8V311H0V20Z" fill="${colors.activityBarBackground}" />`;

  if (colors.activityBarBorder) {
    svg += `<path d="M35.8 20H36.8V311H35.8V20Z" fill="${colors.activityBarBorder}" />`;
  }

  // Explorer
  svg += `<path fill-rule="evenodd" clip-rule="evenodd" d="M15.625 28H22.375L25.75 31.375V40.45L24.775 41.5H21.25V44.95L20.2 46H11.125L10 44.95V33.625L11.125 32.5H14.5V29.125L15.625 28ZM21.25 29.125V32.5H24.625V40.375H15.625V29.125H21.25ZM24.175 31.375L22.375 29.575V31.375H24.175ZM14.5 33.625V40.45L15.625 41.5H20.125V44.875H11.125V33.625H14.5Z" fill="${colors.activityBarForeground}" />`;
  // Search
  svg += `<path fill-rule="evenodd" clip-rule="evenodd" d="M25.8159 70.251C25.8159 73.0377 23.5565 75.2971 20.7699 75.2971C17.9079 75.2971 15.6485 73.0377 15.6485 70.251C15.6485 67.3891 17.9079 65.1297 20.7699 65.1297C23.5565 65.1297 25.8159 67.3891 25.8159 70.251ZM26.9456 70.251C26.9456 73.6402 24.159 76.4268 20.7699 76.4268C19.2636 76.4268 17.9833 75.8996 16.9289 75.0711L10.8285 82L10 81.2469L16.1004 74.318C15.1213 73.2636 14.5188 71.7573 14.5188 70.251C14.5188 66.7866 17.3054 64 20.7699 64C24.159 64 26.9456 66.7866 26.9456 70.251V70.251Z" fill="${colors.activityBarInActiveForeground}" />`;
  // Source control
  svg += `<path d="M24.5187 106.199C24.5187 105.676 24.3693 105.154 24.1452 104.705C23.8465 104.257 23.473 103.884 22.9502 103.66C22.5021 103.436 21.9793 103.286 21.4564 103.361C20.9336 103.436 20.4855 103.585 20.0373 103.884C19.6639 104.183 19.2905 104.631 19.1411 105.079C18.917 105.602 18.8423 106.124 18.9917 106.647C19.0664 107.17 19.2905 107.618 19.5892 107.992C19.9627 108.365 20.4108 108.664 20.9336 108.813C20.7095 109.187 20.4855 109.485 20.112 109.71C19.7386 109.934 19.3651 110.083 18.917 110.083H16.6763C15.8548 110.083 15.0332 110.382 14.4357 110.979V105.527C15.1079 105.378 15.7054 105.004 16.1535 104.481C16.527 103.884 16.751 103.212 16.6763 102.539C16.6017 101.793 16.3029 101.195 15.7801 100.747C15.2573 100.224 14.5851 100 13.9129 100C13.166 100 12.4938 100.224 11.971 100.747C11.4481 101.195 11.1494 101.793 11.0747 102.539C11 103.212 11.2241 103.884 11.5975 104.481C12.0456 105.004 12.6432 105.378 13.3154 105.527V112.398C12.6432 112.473 12.0456 112.846 11.5975 113.444C11.1494 113.967 11 114.639 11 115.386C11.0747 116.058 11.3734 116.656 11.8963 117.178C12.3444 117.627 13.0166 117.925 13.6888 118C14.361 118 15.0332 117.776 15.6307 117.402C16.1535 116.954 16.527 116.357 16.6017 115.685C16.751 114.938 16.6017 114.266 16.2282 113.668C15.9295 113.071 15.332 112.622 14.6598 112.473C14.8838 112.1 15.1826 111.801 15.4813 111.577C15.8548 111.353 16.2282 111.203 16.6763 111.203H18.917C19.5892 111.203 20.2614 110.979 20.8589 110.606C21.4564 110.158 21.8299 109.56 22.0539 108.963C22.7261 108.813 23.3983 108.515 23.8465 107.992C24.2946 107.469 24.5187 106.871 24.5187 106.199V106.199ZM12.195 102.838C12.195 102.465 12.2697 102.166 12.4938 101.867C12.6432 101.568 12.9419 101.344 13.2407 101.27C13.5394 101.12 13.9129 101.12 14.2116 101.12C14.5104 101.195 14.8091 101.344 15.0332 101.643C15.332 101.867 15.4813 102.166 15.556 102.465C15.556 102.763 15.556 103.137 15.4066 103.436C15.332 103.734 15.1079 104.033 14.8091 104.183C14.5104 104.407 14.2116 104.481 13.9129 104.481C13.4647 104.481 13.0166 104.332 12.7178 103.959C12.3444 103.66 12.195 103.212 12.195 102.838V102.838ZM15.556 115.162C15.556 115.461 15.4813 115.759 15.2573 116.058C15.1079 116.357 14.8091 116.581 14.5104 116.656C14.2116 116.805 13.8382 116.805 13.5394 116.805C13.2407 116.73 12.9419 116.581 12.7178 116.282C12.4191 116.058 12.2697 115.759 12.195 115.461C12.195 115.162 12.195 114.788 12.3444 114.49C12.4191 114.191 12.6432 113.892 12.9419 113.743C13.2407 113.519 13.5394 113.444 13.9129 113.444C14.2863 113.444 14.7344 113.593 15.0332 113.967C15.4066 114.266 15.556 114.714 15.556 115.162V115.162ZM21.7552 107.842C21.3817 107.842 21.083 107.768 20.7842 107.544C20.4855 107.394 20.2614 107.095 20.1867 106.797C20.0373 106.498 20.0373 106.124 20.0373 105.826C20.112 105.527 20.2614 105.228 20.5602 105.004C20.7842 104.705 21.083 104.556 21.3817 104.481C21.6805 104.481 22.0539 104.481 22.3527 104.631C22.6515 104.705 22.9502 104.929 23.0996 105.228C23.3237 105.527 23.3983 105.826 23.3983 106.199C23.3983 106.573 23.249 107.021 22.8755 107.32C22.5768 107.693 22.1286 107.842 21.7552 107.842V107.842Z" fill="${colors.activityBarInActiveForeground}" />`;
  // Debug
  svg += `<path d="M17.175 146.125L16.2 147.1C16.05 146.5 15.75 145.975 15.225 145.6C14.7 145.225 14.1 145 13.5 145C12.9 145 12.3 145.225 11.775 145.6C11.25 145.975 10.95 146.5 10.8 147.1L9.825 146.125L9 146.95L10.275 148.225L10.125 148.375V149.5H9V150.625H10.125V150.7C10.2 151.075 10.275 151.375 10.425 151.75L9 153.175L9.825 154L11.025 152.8C11.325 153.175 11.7 153.475 12.15 153.625C12.525 153.85 13.05 154 13.5 154C13.95 154 14.475 153.85 14.85 153.625C15.3 153.475 15.675 153.175 15.975 152.8L17.175 154L18 153.175L16.575 151.75C16.725 151.375 16.8 151 16.875 150.7V150.625H18V149.5H16.875V148.375L16.725 148.225L18 146.95L17.175 146.125ZM13.5 146.125C13.95 146.125 14.4 146.275 14.7 146.65C15 146.95 15.225 147.4 15.225 147.85H11.85C11.85 147.4 12 146.95 12.3 146.65C12.6 146.275 13.05 146.125 13.5 146.125ZM15.75 150.625C15.675 151.225 15.45 151.75 15 152.125C14.625 152.575 14.1 152.8 13.5 152.875C12.9 152.8 12.375 152.575 12 152.125C11.55 151.75 11.325 151.225 11.25 150.625V148.975H15.75V150.625ZM26.85 143.2V144.175L19.125 149.05V147.7L25.5 143.65L15.75 137.5V144.625C15.375 144.325 15 144.175 14.625 144.025V136.45L15.45 136L26.85 143.2Z" fill="${colors.activityBarInActiveForeground}" />`;
  // Extensions
  svg += `<path d="M19.125 173.125L20.25 172H25.875L27 173.125V178.75L25.875 179.875H20.25L19.125 178.75V173.125ZM20.25 173.125V178.75H25.875V173.125H20.25ZM9 183.25V176.5L10.125 175.375H15.75L16.875 176.5V182.125H22.5L23.625 183.25V188.875L22.5 190H16.875H15.75H10.125L9 188.875V183.25ZM15.75 182.125V176.5H10.125V182.125H15.75ZM15.75 183.25H10.125V188.875H15.75V183.25ZM16.875 188.875H22.5V183.25H16.875V188.875Z" fill="${colors.activityBarInActiveForeground}" />`;

  // Badge
  svg += `<svg x="12" y="102" width="24" height="24">`;
  svg += `<circle cx="50%" cy="50%" r="6" fill="${colors.activityBarBadgeBackground}" />`;
  svg += `<text x="50%" y="50%" text-anchor="middle" dominant-baseline="middle" fill="${colors.activityBarBadgeForeground}" font-size="${badgeFontSize}">`;
  svg += `<tspan>3</tspan>`;
  svg += "</text>";
  svg += "</svg>";

  // Tabs Container
  svg += `<rect x="116" y="20" width="344" height="28" fill="${colors.tabsContainerBackground}" />`;
  if (colors.tabsContainerBorder) {
    svg += `<rect x="116" y="47" width="344" height="1" fill="${colors.tabsContainerBorder}" />`;
  }

  // Tab
  svg += `<svg x="36.8" y="20" width="116" height="28">`;
  svg += `<rect x="0" y="0" width="116" height="28" fill="${colors.tabActiveBackground}" />`;
  svg += `<text x="50%" y="50%" text-anchor="middle" dominant-baseline="middle" fill="${colors.tabActiveForeground}" font-size="${titleFontSize}">`;
  svg += `<tspan>${language.tabName}</tspan>`;
  svg += "</text>";
  svg += "</svg>";

  if (colors.tabActiveBorder) {
    svg += `<rect x="36.8" y="47" width="116" height="1" fill="${colors.tabActiveBorder}" />`;
  }

  if (colors.tabActiveBorderTop) {
    svg += `<rect x="36.8" y="20" width="116" height="1" fill="${colors.tabActiveBorderTop}" />`;
  }

  svg += `<rect x="152" y="20" width="1" height="28" fill="${colors.tabBorder}" />`;

  // Status bar
  if (opts.rounded) {
    svg += `<path d="M0 311H460V323C460 327.418 456.418 331 452 331H8.00001C3.58173 331 0 327.418 0 323V311Z" fill="${colors.statusBarBackground}" />`;
  } else {
    svg += `<path d="M0 311H460V331H0V311Z" fill="${colors.statusBarBackground}" />`;
  }

  if (colors.statusBarBorder) {
    svg += `<path d="M0 311H460V312H0V311Z" fill="${colors.statusBarBorder}" />`;
  }

  svg += `<path fill-rule="evenodd" clip-rule="evenodd" d="M14.4246 316.011C15.5636 316.082 16.6314 316.651 17.4145 317.434C18.3399 318.431 18.8382 319.641 18.8382 321.065C18.8382 322.204 18.4111 323.272 17.6992 324.197C16.9873 325.051 15.9907 325.692 14.8517 325.906C13.7127 326.119 12.5737 325.977 11.5771 325.407C10.5805 324.838 9.7974 323.984 9.37027 322.916C8.94315 321.848 8.87196 320.638 9.2279 319.57C9.58384 318.431 10.2245 317.506 11.2211 316.865C12.1466 316.224 13.2856 315.939 14.4246 316.011ZM14.7805 325.194C15.706 324.98 16.5602 324.482 17.2009 323.699C17.7704 322.916 18.1263 321.99 18.0551 320.994C18.0551 319.855 17.628 318.716 16.845 317.933C16.1331 317.221 15.2788 316.794 14.2822 316.723C13.3568 316.651 12.3601 316.865 11.5771 317.434C10.794 318.004 10.2245 318.787 9.93977 319.784C9.65502 320.709 9.65502 321.706 10.0821 322.631C10.5093 323.557 11.15 324.268 12.0042 324.767C12.8585 325.265 13.8551 325.407 14.7805 325.194V325.194ZM13.9263 320.638L15.6348 318.858L16.1331 319.356L14.4246 321.136L16.1331 322.916L15.6348 323.414L13.9263 321.634L12.2178 323.414L11.7195 322.916L13.428 321.136L11.7195 319.356L12.2178 318.858L13.9263 320.638V320.638Z" fill="${colors.statusBarForeground}" />`;
  svg += `<text font-size="10" fill="${colors.statusBarForeground}"><tspan x="46.1748" y="325">0</tspan></text>`;
  svg += `<path fill-rule="evenodd" clip-rule="evenodd" d="M37.0462 316H37.7231L42.7538 325.431L42.4154 326H32.3385L32 325.431L37.0462 316ZM37.3846 316.985L32.9846 325.231H41.7692L37.3846 316.985ZM37.8654 324.462V323.692H36.9038V324.462H37.8654ZM36.9038 322.923V319.846H37.8654V322.923H36.9038Z" fill="${colors.statusBarForeground}" />`;
  svg += `<text font-size="10" fill="${colors.statusBarForeground}"><tspan x="21.9787" y="325">0</tspan></text>`;

  svg += `<text x="100%" y="324" transform="translate(-9, 0)" text-anchor="end" font-size="10" fill="${colors.statusBarForeground}">`;
  svg += language.name;
  svg += `</text>`;

  // Title bar

  if (opts.rounded) {
    svg += `<path d="M0 8C0 3.58172 3.58172 0 8 0H452C456.418 0 460 3.58172 460 8V20H0V8Z" fill="${colors.titleBarActiveBackground}" />`;
  } else {
    svg += `<path d="M0 0H460V20H0V0Z" fill="${colors.titleBarActiveBackground}" />`;
  }
  if (colors.titleBarBorder) {
    svg += `<path d="M0 19H460V20H0V19Z" fill="${colors.titleBarBorder}" />`;
  }
  svg += `<text x="50%" y="14" text-anchor="middle" fill="${
    colors.titleBarActiveForeground
  }" font-size="${titleFontSize}">${escapeHtml(displayName)}</text>`;

  svg += `<text font-family="${opts.monoFontFamily}" fill="${colors.editorForeground}" font-size="${tokenFontSize}" letter-spacing="${editorLetterSpacing}">`;

  for (let i = 0; i < tokens.length; i++) {
    const lineTokens = tokens[i];
    if (!lineTokens) continue;

    const y = i * tokenLineHeight + tokenYOffset;
    svg += `<tspan x="${tokenXOffset}" y="${y}" dy="${tokenFontSize}">`;
    for (let j = 0; j < lineTokens.length; j++) {
      const token = lineTokens[j];
      if (!token) continue;
      const text = encode(token.text).replace(/\s/g, "&#160;");
      const style = renderSvgTokenStyle(token.style, colors.editorForeground);
      svg += `<tspan ${style}>${text}</tspan>`;
    }
    svg += "</tspan>";
  }
  svg += "</text>";

  svg += "</svg>";
  return svg;
}

interface LanguageTokenStyle {
  color?: string;
  fontWeight?: string;
  fontStyle?: string;
  textDecoration?: string;
}

function renderSvgTokenStyle(
  style: LanguageTokenStyle,
  editorForeground: string
) {
  let styles: string[] = [];

  if (style.color) {
    // TODO: This is a hack to fix certain tokens being #000000 when they should probably
    // be the editorForeground. Need to figure out a proper fix in the tokenizer.
    let textColor = style.color;
    if (textColor === "#000000") {
      textColor = editorForeground;
    }
    styles.push(`fill="${textColor}"`);
  }

  if (style.fontWeight) {
    styles.push(`font-weight="${style.fontWeight}"`);
  }

  if (style.fontStyle) {
    styles.push(`font-style="${style.fontStyle}"`);
  }

  if (style.textDecoration) {
    styles.push(`text-decoration="${style.textDecoration}"`);
  }

  return styles.join(" ");
}