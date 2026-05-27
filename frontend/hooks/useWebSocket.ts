import { useEffect, useRef } from "react";
import { toast } from "react-toastify";

export type WsMessage = { event: string; payload: unknown };
export type WsMessageHandler = (payload: unknown) => void;

export function useWebSocket(path: string, toastContainerId: string) {
  const wsConn = useRef<WebSocket | null>(null);
  const handlersRef = useRef(new Map<string, WsMessageHandler>());

  useEffect(() => {
    let isMounted = true;

    const connect = () => {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      wsConn.current = new WebSocket(`${protocol}//${window.location.host}${path}`);

      wsConn.current.onmessage = (ev: MessageEvent<string>) => {
        const message: WsMessage = JSON.parse(ev.data);
        console.log("incoming message", message);
        handlersRef.current.get(message.event)?.(message.payload);
      };

      wsConn.current.onclose = () => {
        toast.error("Disconnected from server. Trying to reconnect in 3s", {
          containerId: toastContainerId,
        });
        if (isMounted) setTimeout(connect, 3000);
      };
    };

    connect();
    return () => {
      isMounted = false;
      wsConn.current?.close();
    };
  }, [path]);

  return { wsConn, handlers: handlersRef.current };
}
