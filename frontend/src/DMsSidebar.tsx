import type { DMConversation } from "./types";
import { API_URL } from "./api";

type DMsSidebarProps = {
  conversations: DMConversation[];
  selectedConversationId: number | null;
  onSelectConversation: (conversationId: number) => void;
};

function DMsSidebar({
  conversations,
  selectedConversationId,
  onSelectConversation,
}: DMsSidebarProps) {
  return (
    <aside className="dms-sidebar">
      <div className="dms-sidebar-header">
        Direct Messages
      </div>

      <div className="dms-sidebar-list">
        {(conversations ?? []).length === 0 ? (
          <div className="dms-sidebar-empty">
            No direct messages yet.
          </div>
        ) : (
          (conversations ?? []).map((conversation) => (
            <button
              key={conversation.id}
              type="button"
              className={`dm-conversation ${
                selectedConversationId === conversation.id
                  ? "selected"
                  : ""
              }`}
              onClick={() =>
                onSelectConversation(conversation.id)
              }
            >
              <div className="dm-conversation-avatar">
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

              <div className="dm-conversation-name">
                {conversation.username}
              </div>
            </button>
          ))
        )}
      </div>
    </aside>
  );
}

export default DMsSidebar;