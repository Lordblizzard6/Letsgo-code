import DOMPurify from "dompurify";

const smd = await import("streaming-markdown");
const { parser, default_renderer, parser_write, parser_end } = smd;

const sanitize = (html: string): string =>
  DOMPurify.sanitize(html, {
    ALLOWED_TAGS: [
      "h1", "h2", "h3", "h4", "h5", "h6",
      "b", "i", "u", "strong", "em", "s", "del",
      "mark", "code", "cite", "abbr", "sub", "sup",
      "blockquote",
      "hr", "br",
      "ul", "ol", "li", "dl", "dt", "dd",
      "table", "thead", "tbody", "tfoot", "tr", "th", "td",
      "a",
      "img",
      "div", "span"
    ],
    ALLOWED_ATTR: ["href", "src", "alt", "title", "class", "id"]
  });

export function renderMarkdown(raw: string): string {
  if (!raw) return "";
  const host = document.createElement("div");
  const renderer = default_renderer(host);
  const p = parser(renderer);

  parser_write(p, raw);
  parser_end(p);

  return sanitize(host.innerHTML);
}

export type TokenInfo = {
  type: string;
  id?: string;
  text?: string;
};

export function tokenInfo(token: any): TokenInfo {
  return {
    type: token.type,
    id: token.id,
    text: token.text
  };
}