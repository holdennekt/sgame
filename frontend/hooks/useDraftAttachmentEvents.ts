import { useEffect, useRef } from "react";
import { Attachment } from "@/types/pack";

// Kept open only while `active`; EventSource auto-reconnects on drops.
export function useDraftAttachmentEvents(
  draftId: string | undefined,
  active: boolean,
  onEvent: (attachment: Attachment) => void
) {
  const handler = useRef(onEvent);
  handler.current = onEvent;

  useEffect(() => {
    if (!draftId || !active) return;

    const source = new EventSource(`/api/packs/drafts/${draftId}/events`);
    const listener = (e: MessageEvent) => {
      handler.current(JSON.parse(e.data) as Attachment);
    };
    source.addEventListener("attachment_updated", listener);
    return () => source.close();
  }, [draftId, active]);
}
