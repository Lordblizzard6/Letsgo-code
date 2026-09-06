import ReactMarkdown from "react-markdown";
import rehypeHighlight from "rehype-highlight";
import remarkGfm from "remark-gfm";

// Cargado con React.lazy desde App.tsx: react-markdown + rehype-highlight
// (highlight.js) son el grueso del bundle (~550KB). Al dividirlo en su propio
// chunk solo se descarga cuando hay mensajes que renderizar, reduciendo el
// JS inicial y acelerando el arranque de la GUI.
// remark-gfm habilita tablas, autolinks y tachado de GFM (el CSS de tablas ya
// existe en App.css).
export default function Markdown({children}: {children: string}) {
    return (
        <ReactMarkdown rehypePlugins={[rehypeHighlight]} remarkPlugins={[remarkGfm]}>
            {children}
        </ReactMarkdown>
    );
}
