import { renderAsync } from "@resvg/resvg-js";

export default async function renderPng(svg: string): Promise<Buffer> {
  const image = await renderAsync(svg, {
    // https://github.com/yisibl/resvg-js/blob/main/index.d.ts
    shapeRendering: 2, // geometricPrecision
    textRendering: 2, // geometricPrecision
    imageRendering: 0, // optimizeQuality
    fitTo: {
      mode: "width",
      value: 800,
    },
  });
  return image.asPng();
}
