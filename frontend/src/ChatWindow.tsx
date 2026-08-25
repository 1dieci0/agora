import { useEffect, useState, type SyntheticEvent } from "react";
import { getMessages, sendMessage } from "./api";
import { connectToChannel } from "./websocket";
import type { Channel, Message } from "./types";

type ChatWindowProps = {
  channel: Channel | null;
};

function ChatWindow({ channel }: ChatWindowProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (channel === null) {
      setMessages([]);
      return;
    }

    const channelID = channel.id;

    async function loadMessages() {
      try {
        const messages = await getMessages(channelID);

        setMessages([...messages].reverse());
      } catch (error) {
        console.error(error);
        setMessages([]);
      }
    }

    loadMessages();
  }, [channel]);

  useEffect(() => {
    if (channel === null) {
      return;
    }

    const socket = connectToChannel(channel.id, (event) => {
      if (event.type === "message_created") {
        const message = event.data as Message;

        setMessages((current) => {
          if (current.some((item) => item.id === message.id)) {
            return current;
          }

          return [...current, message];
        });
      }

      if (event.type === "message_updated") {
        const message = event.data as Message;

        setMessages((current) =>
          current.map((item) =>
            item.id === message.id ? message : item,
          ),
        );
      }

      if (event.type === "message_deleted") {
        const deleted = event.data as { id: number };

        setMessages((current) =>
          current.filter((item) => item.id !== deleted.id),
        );
      }
    });

    return () => {
      socket.close();
    };
  }, [channel]);

  async function handleSubmit(
    event: SyntheticEvent<HTMLFormElement>,
  ) {
    event.preventDefault();

    if (channel === null) {
      return;
    }

    const trimmed = content.trim();

    if (!trimmed || loading) {
      return;
    }

    setLoading(true);

    try {
      await sendMessage(channel.id, trimmed);

      setContent("");
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  }

  if (channel === null) {
    return (
      <main className="chat-window empty">
        <p>Select a channel to start chatting.</p>
      </main>
    );
  }

  return (
    <main className="chat-window">
      <header className="chat-header">
        <span>#</span>
        <h2>{channel.name}</h2>
      </header>

      <div className="message-list">
        {messages.map((message) => (
          <div className="message" key={message.id}>
            <div className="message-author">
              {message.username}
            </div>

            <div className="message-content">
              {message.content}
            </div>
          </div>
        ))}
      </div>

      <form className="message-form" onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder={`Message #${channel.name}`}
          value={content}
          onChange={(event) => setContent(event.target.value)}
          disabled={loading}
        />

        <button
          type="submit"
          disabled={loading || !content.trim()}
        >
          Send
        </button>
      </form>
    </main>
  );
}

export default ChatWindow;
