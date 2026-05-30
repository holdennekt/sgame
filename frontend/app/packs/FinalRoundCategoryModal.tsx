import React, { useEffect, useLayoutEffect, useRef, useState } from "react";
import Modal from "@/components/Modal";
import { FinalRoundCategoryFormData } from "@/types/pack";
import AttachmentEditor from "./AttachmentEditor";
import { IoIosClose } from "react-icons/io";

const inputCls =
  "w-full h-9 px-2.5 bg-background border border-border text-on-background rounded-lg text-sm outline-none focus-ring placeholder:text-on-surface-muted transition-[border-color] duration-150";
const textareaCls =
  "w-full px-2.5 py-2 bg-background border border-border text-on-background rounded-lg text-sm outline-none focus-ring placeholder:text-on-surface-muted transition-[border-color] duration-150 resize-none overflow-hidden";
const labelCls = "block text-xs font-medium text-on-surface-muted";

const autoResize = (el: HTMLTextAreaElement | null) => {
  if (!el) return;
  el.style.height = "auto";
  el.style.height = el.scrollHeight + "px";
};

export default function FinalRoundCategoryModal({
  isOpen,
  close,
  category,
  saveCategory,
  readOnly = false,
}: {
  isOpen: boolean;
  close: () => void;
  category: FinalRoundCategoryFormData;
  saveCategory: (category: FinalRoundCategoryFormData) => void;
  readOnly?: boolean;
}) {
  const [name, setName] = useState(category.name);
  const [text, setText] = useState(category.question.text);
  const [attachment, setAttachment] = useState(category.question.attachment);
  const [answers, setAnswers] = useState(category.question.answers);
  const [comment, setComment] = useState(category.question.comment);

  const [answerInput, setAnswerInput] = useState("");
  const [error, setError] = useState<string | null>(null);
  const textRef = useRef<HTMLTextAreaElement>(null);
  const commentRef = useRef<HTMLTextAreaElement>(null);

  useLayoutEffect(() => {
    setName(category.name);
    setText(category.question.text);
    setAttachment(category.question.attachment);
    setAnswers(category.question.answers);
    setComment(category.question.comment);
    setAnswerInput("");
    setError(null);
  }, [category]);

  useEffect(() => {
    if (!isOpen) return;
    autoResize(textRef.current);
    autoResize(commentRef.current);
  }, [isOpen, text, comment.text]);

  const onSave = () => {
    if (!name || name.length > 25)
      return setError("Name must be between 1 and 25 characters long");
    if (!text || text.length > 200)
      return setError("Text must be between 1 and 200 characters long");
    if (
      attachment.type === "url" &&
      attachment.url &&
      attachment.url.length > 2000
    )
      return setError("Attachment URL is too long");
    if (!answers.length) return setError("At least 1 answer is required");
    if (answers.some((a) => a.length > 200))
      return setError("Answer must be under 200 characters long");
    if (comment.text.length > 400)
      return setError("Comment text must be under 400 characters long");
    if (
      comment.attachment.type === "url" &&
      comment.attachment.url &&
      comment.attachment.url.length > 2000
    )
      return setError("Comment attachment URL is too long");

    saveCategory({ name, question: { text, attachment, answers, comment } });
    close();
  };

  const addAnswer = () => {
    if (!answerInput.trim()) return;
    setAnswers((answers) => [...answers, answerInput.trim()]);
    setAnswerInput("");
  };

  return (
    <Modal isOpen={isOpen} onClose={close} closeByClickingOutside={readOnly}>
      <h3 className="text-base font-semibold text-on-surface mb-4">
        Edit final round category
      </h3>

      <div className="flex flex-col sm:flex-row gap-4">
        <div className="flex flex-col gap-3 w-52">
          <div className="flex flex-col gap-1.5">
            <span className={labelCls}>Category name</span>
            <input
              className={inputCls}
              type="text"
              placeholder="Name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              readOnly={readOnly}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <span className={labelCls}>Question text</span>
            <textarea
              ref={textRef}
              className={textareaCls}
              placeholder="Enter question text..."
              value={text}
              onChange={(e) => setText(e.target.value)}
              maxLength={200}
              required
              readOnly={readOnly}
            />
          </div>

          <AttachmentEditor
            attachment={attachment}
            saveAttachment={setAttachment}
            readOnly={readOnly}
          />
        </div>

        <div className="flex flex-col gap-3 w-52">
          <div className="flex flex-col gap-1.5">
            <span className={labelCls}>Answers</span>
            {answers.length > 0 && (
              <div className="flex flex-wrap gap-1.5">
                {answers.map((answer, index) => (
                  <span
                    key={index}
                    className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-surface-raised text-on-surface text-xs border border-border"
                  >
                    {answer}
                    {!readOnly && (
                      <button
                        type="button"
                        className="w-4 h-4 inline-flex items-center justify-center rounded-full text-on-surface-muted hover:bg-primary hover:text-on-primary transition-colors duration-150"
                        onClick={() =>
                          setAnswers((answers) =>
                            answers.filter((_, i) => i !== index)
                          )
                        }
                      >
                        <IoIosClose size={14} />
                      </button>
                    )}
                  </span>
                ))}
              </div>
            )}
            {!readOnly && (
              <div className="flex">
                <input
                  className="flex-1 min-w-0 h-9 px-2.5 bg-background border border-border border-r-0 text-on-background rounded-l-md text-sm outline-none focus-ring placeholder:text-on-surface-muted transition-[border-color] duration-150"
                  type="text"
                  placeholder="Add answer..."
                  value={answerInput}
                  onChange={(e) => setAnswerInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      addAnswer();
                    }
                  }}
                />
                <button
                  type="button"
                  className="h-9 px-3 rounded-r-md text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 shrink-0"
                  onClick={addAnswer}
                >
                  Add
                </button>
              </div>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <span className={labelCls}>Comment</span>
            <div className="rounded-lg border border-border overflow-hidden">
              <div className="px-3 py-2.5 flex flex-col gap-2.5">
                <div className="flex flex-col gap-1">
                  <span className={labelCls}>Text</span>
                  <textarea
                    ref={commentRef}
                    className={textareaCls}
                    placeholder="Explanation (optional)"
                    value={comment.text}
                    onChange={(e) =>
                      setComment((prev) => ({ ...prev, text: e.target.value }))
                    }
                    maxLength={400}
                    readOnly={readOnly}
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <AttachmentEditor
                    attachment={comment.attachment}
                    saveAttachment={(attachment) =>
                      setComment((prev) => ({ ...prev, attachment }))
                    }
                    readOnly={readOnly}
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {!readOnly && (
        <div className="mt-4 flex items-center justify-between gap-4">
          <button
            type="button"
            className="inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium border border-border text-on-surface-muted hover:bg-surface-raised hover:text-on-surface transition-colors duration-150"
            onClick={close}
          >
            Cancel
          </button>
          <div className="flex items-center gap-2">
            {error && <p className="text-xs text-danger">{error}</p>}
            <button
              type="button"
              className="inline-flex items-center justify-center px-3.5 py-1.5 rounded-lg text-sm font-medium bg-primary text-on-primary hover:bg-primary-hover transition-colors duration-150 shrink-0"
              onClick={onSave}
            >
              Save
            </button>
          </div>
        </div>
      )}
    </Modal>
  );
}
