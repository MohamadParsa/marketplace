import { useEffect, useRef, useState } from "react";
import { EditorContent, useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import DOMPurify from "dompurify";
const clean = (html: string) =>
  DOMPurify.sanitize(html, {
    ALLOWED_TAGS: [
      "p",
      "br",
      "h2",
      "h3",
      "strong",
      "b",
      "em",
      "i",
      "u",
      "s",
      "ul",
      "ol",
      "li",
      "blockquote",
      "a",
      "hr",
      "code",
      "pre",
    ],
    ALLOWED_ATTR: ["href", "rel"],
  });
export function RichProductEditor({
  value,
  onChange,
}: {
  value: string;
  onChange: (html: string) => void;
}) {
  const changed = useRef(onChange);
  changed.current = onChange;
  const [href, setHref] = useState(""),
    [error, setError] = useState("");
  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        heading: { levels: [2, 3] },
        link: { openOnClick: false },
      }),
    ],
    content: clean(value),
    shouldRerenderOnTransaction: true,
    editorProps: {
      attributes: {
        role: "textbox",
        "aria-label": "توضیح کامل",
        "aria-multiline": "true",
        dir: "rtl",
      },
      transformPastedHTML: clean,
    },
    onUpdate: ({ editor }) => changed.current(editor.getHTML()),
  });
  useEffect(() => {
    if (editor && editor.getHTML() !== value)
      editor.commands.setContent(clean(value), { emitUpdate: false });
  }, [value, editor]);
  if (!editor) return <p role="status">در حال آماده‌سازی ویرایشگر…</p>;
  return (
    <div className="rich-product-editor">
      <span id="rich-product-label">توضیح کامل</span>
      <div
        className="rich-product-toolbar"
        role="toolbar"
        aria-label="قالب‌بندی توضیح کامل"
      >
        <button
          type="button"
          aria-pressed={editor.isActive("bold")}
          onClick={() => editor.chain().focus().toggleBold().run()}
        >
          پررنگ
        </button>
        <button
          type="button"
          aria-pressed={editor.isActive("italic")}
          onClick={() => editor.chain().focus().toggleItalic().run()}
        >
          مورب
        </button>
        <button
          type="button"
          aria-pressed={editor.isActive("heading", { level: 2 })}
          onClick={() =>
            editor.chain().focus().toggleHeading({ level: 2 }).run()
          }
        >
          عنوان
        </button>
        <button
          type="button"
          aria-pressed={editor.isActive("bulletList")}
          onClick={() => editor.chain().focus().toggleBulletList().run()}
        >
          فهرست
        </button>
        <button
          type="button"
          aria-pressed={editor.isActive("orderedList")}
          onClick={() => editor.chain().focus().toggleOrderedList().run()}
        >
          فهرست شماره‌دار
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().undo().run()}
          disabled={!editor.can().undo()}
        >
          واگرد
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().redo().run()}
          disabled={!editor.can().redo()}
        >
          بازانجام
        </button>
      </div>
      <EditorContent editor={editor} />
      <div className="rich-product-link">
        <label>
          نشانی لینک
          <input
            type="url"
            value={href}
            dir="ltr"
            placeholder="https://"
            onChange={(e) => {
              setHref(e.target.value);
              setError("");
            }}
          />
        </label>
        <button
          type="button"
          onClick={() => {
            try {
              const u = new URL(href);
              if (!["http:", "https:"].includes(u.protocol)) throw Error();
              editor
                .chain()
                .focus()
                .extendMarkRange("link")
                .setLink({ href: u.href })
                .run();
              setHref("");
              setError("");
            } catch {
              setError("لینک باید یک نشانی معتبر http یا https باشد.");
            }
          }}
        >
          درج لینک
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().unsetLink().run()}
        >
          حذف لینک
        </button>
      </div>
      {error && <p role="alert">{error}</p>}
      <small>{value.length.toLocaleString("fa-IR")} از۲۰٬۰۰۰ نویسه HTML</small>
    </div>
  );
}
