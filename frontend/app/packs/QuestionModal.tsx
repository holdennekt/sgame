import React, { useEffect, useLayoutEffect, useRef, useState } from "react";
import Modal from "@/components/Modal";
import { QuestionFormData, QuestionType } from "@/types/pack";
import AttachmentEditor from "./AttachmentEditor";
import { IoIosClose, IoIosArrowDown } from "react-icons/io";

const inputCls =
  "w-full h-9 px-2.5 bg-background border border-border text-on-background rounded-lg text-sm outline-none focus-ring placeholder:text-on-surface-muted transition-[border-color] duration-150";
const textareaCls =
  "w-full px-2.5 py-2 bg-background border border-border text-on-background rounded-lg text-sm outline-none focus-ring placeholder:text-on-surface-muted transition-[border-color] duration-150 resize-none overflow-hidden";
const selectCls =
  "w-full h-9 pl-2.5 pr-8 bg-background border border-border text-on-background rounded-lg text-sm outline-none focus-ring appearance-none";
const labelCls = "block text-xs font-medium text-on-surface-muted";

const autoResize = (el: HTMLTextAreaElement | null) => {
  if (!el) return;
  el.style.height = "auto";
  el.style.height = el.scrollHeight + "px";
};

export default function QuestionModal({
  isOpen,
  close,
  question,
  saveQuestion,
  readOnly = false,
}: {
  isOpen: boolean;
  close: () => void;
  question: QuestionFormData;
  saveQuestion: (question: Omit<QuestionFormData, "index">) => void;
  readOnly?: boolean;
}) {
  const [value, setValue] = useState(question.value);
  const [type, setType] = useState(question.type);
  const [text, setText] = useState(question.text);
  const [attachment, setAttachment] = useState(question.attachment);
  const [answers, setAnswers] = useState(question.answers);
  const [comment, setComment] = useState(question.comment);

  const [answerInput, setAnswerInput] = useState("");
  const [error, setError] = useState<string | null>(null);
  const textRef = useRef<HTMLTextAreaElement>(null);
  const commentRef = useRef<HTMLTextAreaElement>(null);

  useLayoutEffect(() => {
    setValue(question.value);
    setType(question.type);
    setText(question.text);
    setAttachment(question.attachment);
    setAnswers(question.answers);
    setComment(question.comment);

    setAnswerInput("");
    setError(null);
  }, [question]);

  useEffect(() => {
    if (!isOpen) return;
    autoResize(textRef.current);
    autoResize(commentRef.current);
  }, [isOpen, text, comment.text]);

  const onSave = () => {
    const numValue = Number(value);
    if (
      !numValue ||
      numValue < 1 ||
      !Number.isInteger(numValue) ||
      numValue > 10000
    )
      return setError("Value must be a positive integer under 10k");
    if (text.length > 500)
      return setError("Text must be under 500 characters long");
    if (
      !text.length &&
      ((attachment.type === "file" && !attachment.file) ||
        (attachment.type === "url" && !attachment.url))
    )
      return setError("Text is required without attachment");
    if (
      attachment.type === "url" &&
      attachment.url &&
      attachment.url.length > 2000
    )
      return setError("Attachment URL is too long");
    if (!answers.length) return setError("At least 1 answer is required");
    if (answers.some((answer) => answer.length > 200))
      return setError("Answer must be under 200 characters long");
    if (comment.text.length > 400)
      return setError("Comment text must be under 400 characters long");
    if (
      comment.attachment.type === "url" &&
      comment.attachment.url &&
      comment.attachment.url.length > 2000
    )
      return setError("Comment attachment URL is too long");

    saveQuestion({
      value: numValue,
      type,
      text,
      attachment,
      answers,
      comment,
    });
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
        Edit question
      </h3>

      <div className="flex flex-col sm:flex-row gap-4">
        <div className="flex flex-col gap-3 w-52">
          <div className="flex flex-col gap-1.5">
            <span className={labelCls}>Value</span>
            <input
              className={inputCls}
              type="text"
              inputMode="numeric"
              pattern="[0-9]*"
              placeholder="e.g. 100"
              value={value}
              onChange={(e) =>
                setValue(Number(e.target.value.replace(/[^0-9]/g, "")))
              }
              required
              readOnly={readOnly}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <span className={labelCls}>Type</span>
            <div className="relative">
              <select
                className={selectCls}
                value={type}
                onChange={(e) => setType(e.target.value as QuestionType)}
                disabled={readOnly}
              >
                <option value="regular">Regular</option>
                <option value="catInBag">Cat in bag</option>
                <option value="auction">Auction</option>
              </select>
              <div className="pointer-events-none absolute inset-y-0 right-2.5 flex items-center text-on-surface-muted">
                <IoIosArrowDown size={14} />
              </div>
            </div>
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
