// Modifed from https://github.com/microsoft/vscode/blob/94c9ea46838a9a619aeafb7e8afd1170c967bb55/src/vs/editor/common/modes.ts#L148

export const enum MetadataConsts {
  LANGUAGEID_MASK = 0b00000000000000000000000011111111,
  TOKEN_TYPE_MASK = 0b00000000000000000000011100000000,
  FONT_STYLE_MASK = 0b00000000000000000011100000000000,
  FOREGROUND_MASK = 0b00000000011111111100000000000000,
  BACKGROUND_MASK = 0b11111111100000000000000000000000,

  ITALIC_MASK = 0b00000000000000000000100000000000,
  BOLD_MASK = 0b00000000000000000001000000000000,
  UNDERLINE_MASK = 0b00000000000000000010000000000000,

  SEMANTIC_USE_ITALIC = 0b00000000000000000000000000000001,
  SEMANTIC_USE_BOLD = 0b00000000000000000000000000000010,
  SEMANTIC_USE_UNDERLINE = 0b00000000000000000000000000000100,
  SEMANTIC_USE_FOREGROUND = 0b00000000000000000000000000001000,
  SEMANTIC_USE_BACKGROUND = 0b00000000000000000000000000010000,

  LANGUAGEID_OFFSET = 0,
  TOKEN_TYPE_OFFSET = 8,
  FONT_STYLE_OFFSET = 11,
  FOREGROUND_OFFSET = 14,
  BACKGROUND_OFFSET = 23,
}

export const enum LanguageId {
  Null = 0,
  PlainText = 1,
}

export const enum StandardTokenType {
  Other = 0,
  Comment = 1,
  String = 2,
  RegEx = 4,
}

export const enum FontStyle {
  NotSet = -1,
  None = 0,
  Italic = 1,
  Bold = 2,
  Underline = 4,
}

export const enum ColorId {
  None = 0,
  DefaultForeground = 1,
  DefaultBackground = 2,
}

export interface Style {
  color?: string;
  fontWeight?: string;
  fontStyle?: string;
  textDecoration?: string;
}

export default class TokenMetadata {
  public static getLanguageId(metadata: number): LanguageId {
    return (
      (metadata & MetadataConsts.LANGUAGEID_MASK) >>>
      MetadataConsts.LANGUAGEID_OFFSET
    );
  }

  public static getTokenType(metadata: number): StandardTokenType {
    return (
      (metadata & MetadataConsts.TOKEN_TYPE_MASK) >>>
      MetadataConsts.TOKEN_TYPE_OFFSET
    );
  }

  public static getFontStyle(metadata: number): FontStyle {
    return (
      (metadata & MetadataConsts.FONT_STYLE_MASK) >>>
      MetadataConsts.FONT_STYLE_OFFSET
    );
  }

  public static getForeground(metadata: number): ColorId {
    return (
      (metadata & MetadataConsts.FOREGROUND_MASK) >>>
      MetadataConsts.FOREGROUND_OFFSET
    );
  }

  public static getBackground(metadata: number): ColorId {
    return (
      (metadata & MetadataConsts.BACKGROUND_MASK) >>>
      MetadataConsts.BACKGROUND_OFFSET
    );
  }

  public static getClassNameFromMetadata(metadata: number): string {
    let foreground = this.getForeground(metadata);
    let className = "mtk" + foreground;

    let fontStyle = this.getFontStyle(metadata);
    if (fontStyle & FontStyle.Italic) {
      className += " mtki";
    }
    if (fontStyle & FontStyle.Bold) {
      className += " mtkb";
    }
    if (fontStyle & FontStyle.Underline) {
      className += " mtku";
    }

    return className;
  }

  public static getInlineStyleFromMetadata(
    metadata: number,
    colorMap: string[]
  ): string {
    const foreground = this.getForeground(metadata);
    const fontStyle = this.getFontStyle(metadata);

    let result = `color: ${colorMap[foreground]};`;
    if (fontStyle & FontStyle.Italic) {
      result += "font-style: italic;";
    }
    if (fontStyle & FontStyle.Bold) {
      result += "font-weight: bold;";
    }
    if (fontStyle & FontStyle.Underline) {
      result += "text-decoration: underline;";
    }
    return result;
  }

  static getStyleObject(metadata: number, colorMap: string[]): Style {
    const foreground = TokenMetadata.getForeground(metadata);
    const fontStyle = TokenMetadata.getFontStyle(metadata);

    const style: Style = {};

    if (colorMap[foreground]) {
      style.color = colorMap[foreground];
    }
    if (fontStyle & FontStyle.Italic) {
      style.fontStyle = "italic";
    }
    if (fontStyle & FontStyle.Bold) {
      style.fontWeight = "bold";
    }
    if (fontStyle & FontStyle.Underline) {
      style.textDecoration = "underline";
    }

    return style;
  }
}
