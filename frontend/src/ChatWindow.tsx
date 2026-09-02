import { useEffect,
  useRef,
  useState,
  type SyntheticEvent,
  type ChangeEvent,
} from "react";
import {
  API_URL,
  deleteMessage,
  getMessages,
  sendMessage,
  updateMessage,
} from "./api";
import type { Channel, Message, User , RealtimeEvent, Member} from "./types";


type ChatWindowProps = {
  user: User;
  channel: Channel | null;
  members: Member[];
  realtimeEvent: RealtimeEvent | null;
  onLatestMessage: (channelID: number, messageID: number) => void;
};

function ChatWindow({
  user,
  channel,
  members,
  realtimeEvent,
  onLatestMessage,
}: ChatWindowProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(false);

  const [editingMessageID, setEditingMessageID] = useState<number | null>(
  null,
  );

  const [editingContent, setEditingContent] = useState("");

  const messagesEndRef = useRef<HTMLDivElement | null>(null);
  const messageInputRef = useRef<HTMLTextAreaElement | null>(null);


  useEffect(() => {
    if (!realtimeEvent || !channel) return;

    switch (realtimeEvent.type) {
      case "message_created": {
        const message = realtimeEvent.data;

        if (message.channel_id !== channel.id) return;

        setMessages((current) => {
          if (current.some((m) => m.id === message.id)) {
            return current;
          }

          return [...current, message];
        });

        onLatestMessage(channel.id, message.id);

        break;
      }

      case "message_updated": {
        const message = realtimeEvent.data;

        if (message.channel_id !== channel.id) return;

        setMessages((current) =>
          current.map((m) =>
            m.id === message.id ? message : m
          )
        );

        break;
      }

      case "message_deleted": {
        const { id, channel_id } = realtimeEvent.data;

        if (channel_id !== channel.id) return;

        setMessages((current) =>
          current.filter((m) => m.id !== id)
        );

        break;
      }
    }
  }, [realtimeEvent, channel]);


  useEffect(() => {
    if (channel === null) {
      setMessages([]);
      return;
    }

    const channelID = channel.id;

    async function loadMessages() {
      try {
        const messages = await getMessages(channelID);

        const orderedMessages = [...messages].reverse();

        setMessages(orderedMessages);

        if (orderedMessages.length > 0) {
          const latestMessage =
            orderedMessages[orderedMessages.length - 1];

          onLatestMessage(channelID, latestMessage.id);
        }
      } catch (error) {
        console.error(error);
        setMessages([]);
      }
    }

    loadMessages();
  }, [channel]);

  useEffect(() => {
    if (channel !== null) {
      requestAnimationFrame(() => {
        messageInputRef.current?.focus();
      });
    }
  }, [channel]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({
      behavior: "smooth",
    });
  }, [messages]);


  function handleContentChange(
    event: ChangeEvent<HTMLTextAreaElement>,
  ) {
    const textarea = event.currentTarget;

    textarea.style.height = "auto";

    const maxHeight = 400;

    textarea.style.height = `${Math.min(
      textarea.scrollHeight,
      maxHeight,
    )}px`;

    setContent(textarea.value);
  }


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

      if (messageInputRef.current) {
        messageInputRef.current.style.height = "auto";
      }
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (!loading) {
      messageInputRef.current?.focus();
    }
  }, [loading]);

  async function handleDelete(messageID: number) {
  try {
    await deleteMessage(messageID);
  } catch (error) {
    console.error(error);
  }
}

  function startEditing(message: Message) {
    setEditingMessageID(message.id);
    setEditingContent(message.content);
  }

  function cancelEditing() {
    setEditingMessageID(null);
    setEditingContent("");
  }

  async function handleEdit(messageID: number) {
    const trimmed = editingContent.trim();

    if (!trimmed) {
      return;
    }

    try {
      await updateMessage(messageID, trimmed);

      cancelEditing();
    } catch (error) {
      console.error(error);
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
        {messages.map((message) => {

          const member = members.find(
            (member) => member.id === message.user_id,
          );

          const isOwnMessage = message.user_id === user.id;
          const isEditing = editingMessageID === message.id;

          return (
            <div className="message" key={message.id}>
               <div className="message-avatar">
                  {member?.avatar_url ? (
                    <img
                      src={`${API_URL}${member.avatar_url}`}
                      alt=""
                    />
                  ) : (
                    message.username.charAt(0).toUpperCase()
                  )}
                </div>

                <div className="message-body">
                  <div className="message-header">
                    <span className="message-author">
                      {message.username}
                    </span>

                    <span className="message-time">
                      {new Date(message.created_at).toLocaleTimeString([], {
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </span>

                    {isOwnMessage && !isEditing && (
                      <div className="message-actions">
                        <button
                          type="button"
                          onClick={() => startEditing(message)}
                        >
                          Edit
                        </button>

                        <button
                          type="button"
                          onClick={() => handleDelete(message.id)}
                        >
                          Delete
                        </button>
                      </div>
                    )}
                  </div>

                  {isEditing ? (
                    <div className="message-edit">
                      <input
                        value={editingContent}
                        onChange={(event) =>
                          setEditingContent(event.target.value)
                        }
                        onKeyDown={(event) => {
                          if (event.key === "Enter") {
                            handleEdit(message.id);
                          }

                          if (event.key === "Escape") {
                            cancelEditing();
                          }
                        }}
                      />

                      <div className="edit-actions">
                        <button
                          type="button"
                          onClick={() => handleEdit(message.id)}
                        >
                          Save
                        </button>

                        <button
                          type="button"
                          onClick={cancelEditing}
                        >
                          Cancel
                        </button>
                      </div>
                    </div>
                  ) : (
                    <div className="message-content">
                      {message.content}
                    </div>
                  )}
                </div>
            </div>
          );
        })}

      <div ref={messagesEndRef} />

      </div>

      <form className="message-form" onSubmit={handleSubmit}>

        <textarea
          ref={messageInputRef}
          placeholder={`Message #${channel.name}`}
          value={content}
          onChange={handleContentChange}
          disabled={loading}
          rows={1}
          onKeyDown={(event) => {
            if (event.key === "Enter" && !event.shiftKey) {
              event.preventDefault();

              if (!content.trim() || loading) {
                return;
              }

              event.currentTarget.form?.requestSubmit();
            }
          }}
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
