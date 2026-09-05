export type User = {
  id: number;
  username: string;
  avatar_url: string | null;
};

export type Server = {
  id: number;
  name: string;
  owner_id: number;
};

export type Channel = {
  id: number;
  server_id: number;
  name: string;
  type: string;
};

export type Message = {
  id: number;
  channel_id: number;
  user_id: number;
  username: string;
  content: string;
  created_at: string;
};

export type Member = {
  id: number;
  username: string;
  avatar_url: string | null;
  role: "owner" | "admin" | "member";
};


export type VoiceParticipant = {
  id: number;
  channelId: number;
  username: string;
  avatar_url: string | null;
  muted: boolean;
  deafened: boolean;
  speaking: boolean;
};



export type VoiceState = {
  user_id: number;
  channel_id: number;
  muted: boolean;
  deafened: boolean;
};

type UnreadUpdate = {
  server_id: number;
  channel_id: number;
  message_id: number;
  user_id: number;
  unread_count: number;
};

export type RealtimeEvent =
  | {
      type: "voice_state";
      data: VoiceState[];
    }
  | {
      type: "voice_join";
      data: VoiceState;
    }
  | {
      type: "voice_leave";
      data: VoiceState;
    }
  | { type: "voice_update"; data: VoiceState }
  | {
      type: "message_created";
      data: Message;
    }
  | {
      type: "message_updated";
      data: Message;
    }
  | {
      type: "unread_update";
      data: UnreadUpdate;
    }
  | {
      type: "mention";
      data: {
        id: number;
        server_id: number;
        channel_id: number;
        message_id: number;
        from_user_id: number;
      };
    }
  | {
      type: "message_deleted";
      data: {
        id: number;
        channel_id: number;
      };
    }
  | {
      type: "dm_created";
      data: DirectMessage;
    }
  | {
      type: "user_updated";
      data: User;
    };



export type AppNotification = {
  id: number;
  user_id: number;
  type: string;
  server_id: number | null;
  channel_id: number | null;
  message_id: number | null;
  from_user_id: number | null;
  read: boolean;
  created_at: string;
};


export type DirectMessage = {
  id: number;
  conversation_id: number;
  user_id: number;
  username: string;
  avatar_url: string | null;
  content: string;
  created_at: string;
};

export type DMConversation = {
  id: number;
  user_id: number;
  username: string;
  avatar_url: string | null;
  created_at: string;
};