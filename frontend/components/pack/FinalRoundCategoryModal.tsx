import React, { useLayoutEffect, useState } from "react";
import { toast } from "react-toastify";
import Modal from "../Modal";
import { FinalRoundCategoryFormData } from "@/types/pack";
import AttachmentEditor from "./AttachmentEditor";

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
  const [answers, setAnswers] = useState(category.question.answers);
  const [comment, setComment] = useState(category.question.comment);
  const [attachment, setAttachment] = useState(category.question.attachment);
  const [editAttachment, setEditAttachment] = useState(
    category.question.attachment,
  );

  const [answerInput, setAnswerInput] = useState("");

  useLayoutEffect(() => {
    setName(category.name);
    setText(category.question.text);
    setAnswerInput("");
    setAnswers(category.question.answers);
    setComment(category.question.comment);
    setAttachment(category.question.attachment);
    setEditAttachment(category.question.attachment);
  }, [category]);

  const onSave = () => {
    try {
      if (!name || name.length > 25)
        throw new Error("Name must be between 1 and 25 charachters long");
      if (!text || text.length > 200)
        throw new Error("Text must be between 1 and 200 charachters long");
      if (
        editAttachment.type === "url" &&
        editAttachment?.url &&
        editAttachment?.url.length > 2000
      )
        throw new Error("Attachment URL is too long");
      if (!answers.length) throw new Error("At least 1 answer is required");
      if (comment && comment.length > 200)
        throw new Error("Comment must be less than 200 charachters long");
      saveCategory({
        name,
        question: {
          text,
          attachment: editAttachment,
          answers,
          comment,
        },
      });
      close();
    } catch (error) {
      if (error instanceof Error)
        toast.error(error.message, { containerId: "editor" });
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={close} closeByClickingOutside>
      <h3 className="text-base/7 font-medium">Edit question</h3>
      <div className="flex flex-col sm:flex-row gap-2 sm:gap-4 mt-2">
        <div className="w-48 flex flex-col gap-2 flex-1">
          <label>
            <p className="text-sm font-medium">Name</p>
            <input
              className="w-full h-8 rounded-md mt-1 p-1 text-black"
              type="text"
              placeholder="Value"
              name="value"
              value={name}
              onChange={e => setName(e.target.value)}
              required
              readOnly={readOnly}
            />
          </label>
          <label>
            <p className="text-sm font-medium">Text</p>
            <textarea
              className="w-full h-8 rounded-md mt-1 p-1 text-black"
              placeholder="Question text"
              name="text"
              value={text}
              onChange={e => setText(e.target.value)}
              maxLength={200}
              required
              readOnly={readOnly}
            />
          </label>
          <div className="flex flex-col gap-2">
            <p className="text-sm font-medium">Answers</p>
            {answers.length > 0 && (
              <ul className="list-inside list-disc">
                {answers.map((answer, index) => (
                  <li
                    className="cursor-pointer"
                    onClick={
                      readOnly ?
                        undefined :
                        () =>
                          setAnswers(answers =>
                            answers.filter((a, i) => index !== i),
                          )
                    }
                    key={index}
                  >
                    {answer}
                  </li>
                ))}
              </ul>
            )}
            {!readOnly && (
              <div className="w-full flex">
                <input
                  className="flex-1 min-w-0 h-8 p-1 rounded-l-lg text-black focus:outline-none"
                  type="text"
                  placeholder="Answer"
                  value={answerInput}
                  onChange={e => setAnswerInput(e.target.value)}
                  onKeyDown={e => {
                    if (e.code !== "Enter") return;
                    setAnswers(answers => [...answers, answerInput]);
                    setAnswerInput("");
                  }}
                />
                <button
                  className="h-full rounded-r-lg px-2 py-1 primary"
                  onClick={() => {
                    setAnswers(answers => [...answers, answerInput]);
                    setAnswerInput("");
                  }}
                >
                  Add
                </button>
              </div>
            )}
          </div>
        </div>
        <div className="w-48 flex flex-col gap-2">
          <label>
            <p className="text-sm font-medium">Comment</p>
            <textarea
              className="w-full h-8 rounded-md mt-1 p-1 text-black"
              placeholder="Comment"
              name="text"
              value={comment ?? ""}
              onChange={e => setComment(e.target.value ?? null)}
              maxLength={100}
              readOnly={readOnly}
            />
          </label>
          <AttachmentEditor
            attachment={attachment}
            editAttachment={editAttachment}
            setEditAttachment={setEditAttachment}
            readOnly={readOnly}
          />
        </div>
      </div>
      {!readOnly && (
        <div className="mt-4 flex flex-row-reverse">
          <button
            className="w-fit h-fit rounded px-2 py-1 primary"
            type="button"
            onClick={onSave}
          >
            Save
          </button>
        </div>
      )}
    </Modal>
  );
}
