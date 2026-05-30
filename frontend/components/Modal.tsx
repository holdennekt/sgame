import { MouseEventHandler, ReactNode, useRef } from "react";

export default function Modal({
  isOpen,
  onClose,
  closeByClickingOutside,
  children,
}: {
  isOpen: boolean;
  onClose: () => void;
  closeByClickingOutside: boolean;
  children?: ReactNode;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  if (isOpen) {
    if (!dialog.current?.open) dialog.current?.showModal();
  } else {
    dialog.current?.close();
  }

  const onClick: MouseEventHandler<HTMLDialogElement> = (ev) => {
    if (ev.target !== ev.currentTarget) return;
    const rect = dialog.current!.getBoundingClientRect();
    const outside =
      ev.clientX < rect.left ||
      ev.clientX > rect.right ||
      ev.clientY < rect.top ||
      ev.clientY > rect.bottom;
    if (outside) dialog.current?.close();
  };

  return (
    <dialog
      className="w-fit max-w-[95vw] rounded-lg p-6 bg-surface text-on-surface border border-border shadow-lg backdrop:bg-black/50 backdrop:backdrop-blur-sm outline-none"
      ref={dialog}
      tabIndex={-1}
      onClick={closeByClickingOutside ? onClick : undefined}
      onClose={onClose}
    >
      {children}
    </dialog>
  );
}
