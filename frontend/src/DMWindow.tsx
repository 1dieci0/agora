import { useEffect, useRef, useState } from "react";

import {
  API_URL,
  getDMMessages,
  sendDM,
} from "./api";

import type {
  DMConversation,
  DirectMessage,
  RealtimeEvent,
  User,
} from "./types";

type DMWindowProps = {
  user: User | null;
  conversation: DMConversation | null;
  conversationId: number | null;
  realtimeDMEvent: RealtimeEvent | null;
};

function DMWindow({
  user,
  conversation,
  conversationId,
  realtimeDMEvent,
}: DMWindowProps) {
  const [messages, setMessages] = useState<DirectMessage[]>(
    [],
  );

  const [content, setContent] = useState("");

  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);

  const messagesEndRef = useRef<HTMLDivElement | null>(
    null,
  );

    useEffect(() => {
    if (conversationId === null) {
        setMessages([]);
        return;
    }

    const currentConversationId = conversationId;

    async function loadMessages() {
      setLoading(true);

      try {
        const messages = await getDMMessages(
        currentConversationId,
        );

        setMessages(messages ?? []);
      } catch (error) {
        console.error( 
          "Could not load DM messages:",
          error,
        );
        setMessages([]);
      } finally {
        setLoading(false);
      }
    }

    loadMessages();
  }, [conversationId]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({
      behavior: "smooth",
    });
  }, [messages]);

    async function handleSendMessage() {
    if (
        conversationId === null ||
        sending ||
        content.trim() === ""
    ) {
        return;
    }

    const currentConversationId = conversationId;
    const messageContent = content.trim();

    setContent("");
    setSending(true);

    try {
        const message = await sendDM(
        currentConversationId,
        messageContent,
        );

      setMessages((current) => [
        ...current,
        message,
      ]);
    } catch (error) {
      console.error(
        "Could not send DM:",
        error,
      );

      setContent(messageContent);
    } finally {
      setSending(false);
    }
  }
  
  useEffect(() => {
  if (
    !realtimeDMEvent ||
    realtimeDMEvent.type !== "dm_created"
  ) {
    return;
  }

  const message = realtimeDMEvent.data;

  if (message.conversation_id !== conversationId) {
    return;
  }

  setMessages((current) => {
    if (
      current.some(
        (existing) => existing.id === message.id,
      )
    ) {
      return current;
    }

    return [...current, message];
  });
}, [realtimeDMEvent, conversationId]);

  function handleKeyDown(
    event: React.KeyboardEvent<HTMLTextAreaElement>,
  ) {
    if (
      event.key === "Enter" &&
      !event.shiftKey
    ) {
      event.preventDefault();
      handleSendMessage();
    }
  }

  if (!conversation || conversationId === null) {
    return (
      <main className="dm-window dm-window-empty">
        <div className="dm-window-empty-content">
          <div className="dm-window-empty-title">
            Direct Messages
          </div>

          <div className="dm-window-empty-text">
            Select a conversation to start chatting.
          </div>
        </div>
      </main>
    );
  }

  return (
    <main className="dm-window">
      <header className="dm-window-header">
        <div className="dm-window-header-avatar">
          {conversation.avatar_url ? (
            <img
              src={`${API_URL}${conversation.avatar_url}`}
              alt=""
            />
          ) : (
            conversation.username
              .charAt(0)
              .toUpperCase()
          )}
        </div>

        <div className="dm-window-header-name">
          {conversation.username}
        </div>
      </header>

      <div className="dm-window-messages">
        {loading ? (
          <div className="dm-window-loading">
            Loading messages...
          </div>
        ) : messages.length === 0 ? (
          <div className="dm-window-no-messages">
            <div className="dm-window-no-messages-avatar">
              {conversation.avatar_url ? (
                <img
                  src={`${API_URL}${conversation.avatar_url}`}
                  alt=""
                />
              ) : (
                conversation.username
                  .charAt(0)
                  .toUpperCase()
              )}
            </div>

            <div className="dm-window-no-messages-title">
              {conversation.username}
            </div>

            <div className="dm-window-no-messages-text">
              This is the beginning of your direct
              message history with{" "}
              {conversation.username}.
            </div>
          </div>
        ) : (
          messages.map((message) => (
            <div
              key={message.id}
              className={[
                "dm-message",
                message.user_id === user?.id
                  ? "own"
                  : "",
              ]
                .filter(Boolean)
                .join(" ")}
            >
              <div className="dm-message-avatar">
                {message.avatar_url ? (
                  <img
                    src={`${API_URL}${message.avatar_url}`}
                    alt=""
                  />
                ) : (
                  message.username
                    .charAt(0)
                    .toUpperCase()
                )}
              </div>

              <div className="dm-message-body">
                <div className="dm-message-meta">
                  <span className="dm-message-username">
                    {message.username}
                  </span>

                  <span className="dm-message-time">
                    {new Date(
                      message.created_at,
                    ).toLocaleString()}
                  </span>
                </div>

                <div className="dm-message-content">
                  {message.content}
                </div>
              </div>
            </div>
          ))
        )}

        <div ref={messagesEndRef} />
      </div>

      <div className="dm-window-input-area">
        <textarea
          className="dm-window-input"
          value={content}
          onChange={(event) =>
            setContent(event.target.value)
          }
          onKeyDown={handleKeyDown}
          placeholder={`Message ${conversation.username}`}
          rows={1}
          disabled={sending}
        />

        <button
          type="button"
          className="dm-window-send"
          onClick={handleSendMessage}
          disabled={
            sending || content.trim() === ""
          }
        >
          Send
        </button>
      </div>
    </main>
  );
}

export default DMWindow;