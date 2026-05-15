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
  const onClick: MouseEventHandler<HTMLDialogElement> = ev => {
    const rect = ev.currentTarget.getBoundingClientRect();
    if (!rect) return;
    if (
      ev.clientX < rect.left ||
      ev.clientX > rect.right ||
      ev.clientY < rect.top ||
      ev.clientY > rect.bottom
    ) {
      dialog.current?.close();
    }
  };

  return (
    <dialog
      className="w-fit rounded-lg p-6 surface border backdrop:bg-transparent"
      ref={dialog}
      onClick={closeByClickingOutside ? onClick : undefined}
      onClose={onClose}
    >
      {children}
    </dialog>
  );
}
