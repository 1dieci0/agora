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
  speaking: boolean;
};



export type VoiceState = {
  user_id: number;
  channel_id: number;
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
  | {
      type: "message_created";
      data: Message;
    }
  | {
      type: "message_updated";
      data: Message;
    }
  | {
      type: "message_deleted";
      data: {
        id: number;
        channel_id: number;
      };
    };
