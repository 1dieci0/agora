import type { Message } from "./types";

const WS_URL = "ws://localhost:8080";

export type MessageEvent = {
  type: "message_created" | "message_updated" | "message_deleted";
  data: Message | { id: number };
};

export function connectToChannel(
  channelID: number,
  onEvent: (event: MessageEvent) => void,
): WebSocket {
  const socket = new WebSocket(
    `${WS_URL}/ws/channels/${channelID}`,
  );

  socket.onmessage = (event) => {
    try {
      const data: MessageEvent = JSON.parse(event.data);

      onEvent(data);
    } catch (error) {
      console.error("Invalid WebSocket message:", error);
    }
  };

  socket.onerror = (error) => {
    console.error("WebSocket error:", error);
  };

  return socket;
}
