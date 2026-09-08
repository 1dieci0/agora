import type { Channel, Message, Server, User, Member, AppNotification, DirectMessage, DMConversation } from "./types";

export const API_URL = "http://localhost:8080";

type LoginResponse = {
  message: string;
  user: User;
};

export type ChannelUnread = {
  server_id: number;
  channel_id: number;
  unread_count: number;
  last_message_id: number;
  last_read_message_id: number;
};

export type UnreadResponse = {
  channels: ChannelUnread[];
};

export async function login(username: string, password: string): Promise<LoginResponse> {
  const response = await fetch(`${API_URL}/api/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
    body: JSON.stringify({
      username,
      password,
    }),
  });

  if (!response.ok) {
    throw new Error("Invalid username or password");
  }

  return response.json();
}

export async function getMe(): Promise<User> {
  const response = await fetch(`${API_URL}/api/me`, {
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error("Not authenticated");
  }

  return response.json();
}


export async function getServers(): Promise<Server[]> {
  const response = await fetch(`${API_URL}/api/servers`, {
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error("Could not load servers");
  }

  const data = await response.json();

  return data.servers;
}

export async function getChannels(serverID: number): Promise<Channel[]> {
  const response = await fetch(
    `${API_URL}/api/servers/${serverID}/channels`,
    {
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not load channels");
  }

  const data = await response.json();

  return data.channels;
}


export async function getMessages(
  channelID: number,
): Promise<Message[]> {
  const response = await fetch(
    `${API_URL}/api/channels/${channelID}/messages`,
    {
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not load messages");
  }

  const data = await response.json();

  return data.messages;
}

export async function sendMessage(
  channelID: number,
  content: string,
): Promise<Message> {
  const response = await fetch(
    `${API_URL}/api/channels/${channelID}/messages`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        content,
      }),
    },
  );

  if (!response.ok) {
    throw new Error("Could not send message");
  }

  const data = await response.json();

  return data.message;
}


export async function updateMessage(
  messageID: number,
  content: string,
): Promise<Message> {
  const response = await fetch(
    `${API_URL}/api/messages/${messageID}`,
    {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        content,
      }),
    },
  );

  if (!response.ok) {
    throw new Error("Could not update message");
  }

  const data = await response.json();

  return data.message;
}

export async function deleteMessage(messageID: number): Promise<void> {
  const response = await fetch(
    `${API_URL}/api/messages/${messageID}`,
    {
      method: "DELETE",
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not delete message");
  }
}


export async function createServer(name: string): Promise<Server> {
  const response = await fetch(`${API_URL}/api/servers`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
    body: JSON.stringify({
      name,
    }),
  });

  if (!response.ok) {
    throw new Error("Could not create server");
  }

  const data = await response.json();

  return data.server;
}


export async function createChannel(
  serverID: number,
  name: string,
  type: "text" | "voice",
): Promise<Channel> {
  const response = await fetch(
    `${API_URL}/api/servers/${serverID}/channels`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        name,
        type,
      }),
    },
  );

  if (!response.ok) {
    throw new Error("Could not create channel");
  }

  const data = await response.json();

  return data.channel;
}


export async function logout(): Promise<void> {
  const response = await fetch(`${API_URL}/api/logout`, {
    method: "POST",
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error("Could not log out");
  }
}


export async function createInvite(
  serverID: number,
): Promise<string> {
  const response = await fetch(
    `${API_URL}/api/servers/${serverID}/invites`,
    {
      method: "POST",
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not create invite");
  }

  const data = await response.json();

  return data.code;
}

export async function joinServer(
  code: string,
): Promise<Server> {
  const response = await fetch(
    `${API_URL}/api/invites/${encodeURIComponent(code)}/join`,
    {
      method: "POST",
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not join server");
  }

  const data = await response.json();

  return data.server;
}


export async function getMembers(
  serverID: number,
): Promise<Member[]> {
  const response = await fetch(
    `${API_URL}/api/servers/${serverID}/members`,
    {
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Could not load members");
  }

  const data = await response.json();

  return data.members;
}


export async function register(
  username: string,
  password: string,
): Promise<User> {
  const response = await fetch(`${API_URL}/api/register`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
    body: JSON.stringify({
      username,
      password,
    }),
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || "Could not create account");
  }

  const data = await response.json();

  return data.user;
}

export async function uploadAvatar(file: File): Promise<User> {
  const formData = new FormData();

  formData.append("avatar", file);

  const response = await fetch(
    `${API_URL}/api/me/avatar`,
    {
      method: "PUT",
      credentials: "include",
      body: formData,
    },
  );

  if (!response.ok) {
    const message = await response.text();

    throw new Error(
      message || "Could not upload avatar",
    );
  }

  const data = await response.json();

  return data.user;
}


export async function getVoiceToken(
  channelId: number,
): Promise<string> {
  const response = await fetch(
    `${API_URL}/api/voice/token?channel_id=${channelId}`,
    {
      method: "POST",
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error(
      `Failed to get voice token: ${response.status}`,
    );
  }

  const data = await response.json();

  if (typeof data.token !== "string") {
    throw new Error("Voice token missing from response");
  }

  return data.token;
}



export async function getUnread(): Promise<UnreadResponse> {
  const response = await fetch(`${API_URL}/api/unread`, {
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error("Could not load unread messages");
  }

  return response.json();
}

export async function markChannelRead(
  channelID: number,
  messageID: number,
): Promise<void> {
  const response = await fetch(
    `${API_URL}/api/channels/${channelID}/read`,
    {
      method: "PUT",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        message_id: messageID,
      }),
    },
  );

  if (!response.ok) {
    const body = await response.text();

    console.error(
      "markChannelRead failed:",
      response.status,
      body,
    );

    throw new Error(
      `Could not mark channel as read (${response.status})`,
    );
  }
}

export async function getNotifications(): Promise<
  AppNotification[]
> {
  const response = await fetch(
    `${API_URL}/api/notifications`,
    {
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error(
      `HTTP ${response.status}`,
    );
  }

  return response.json();
}

export async function markNotificationRead(
  notificationID: number,
): Promise<void> {
  const response = await fetch(
    `${API_URL}/api/notifications/${notificationID}/read`,
    {
      method: "POST",
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error(
      `HTTP ${response.status}`,
    );
  }
}

export async function markChannelNotificationsRead(
  channelID: number,
): Promise<void> {
  const response = await fetch(
    `${API_URL}/api/notifications/channel/${channelID}/read`,
    {
      method: "POST",
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }
}



export async function createDM(
  userID: number,
): Promise<{ conversation_id: number }> {
  const response = await fetch(`${API_URL}/api/dms`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      user_id: userID,
    }),
  });

  if (!response.ok) {
    throw new Error("Failed to create DM");
  }

  return response.json();
}

export async function getDMs(): Promise<DMConversation[]> {
  const response = await fetch(`${API_URL}/api/dms`, {
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error("Failed to get DMs");
  }

  return response.json();
}

export async function getDMMessages(
  conversationID: number,
): Promise<DirectMessage[]> {
  const response = await fetch(
    `${API_URL}/api/dms/${conversationID}/messages`,
    {
      credentials: "include",
    },
  );

  if (!response.ok) {
    throw new Error("Failed to get DM messages");
  }

  return response.json();
}

export async function sendDM(
  conversationID: number,
  content: string,
): Promise<DirectMessage> {
  const response = await fetch(
    `${API_URL}/api/dms/${conversationID}/messages`,
    {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        content,
      }),
    },
  );

  if (!response.ok) {
    throw new Error("Failed to send DM");
  }

  return response.json();
}


export async function markDMConversationRead(
  conversationId: number,
  messageId: number,
) {
  const response = await fetch(
    `${API_URL}/api/dms/${conversationId}/read`,
    {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        message_id: messageId,
      }),
    },
  );

  if (!response.ok) {
    throw new Error(
      `Could not mark DM as read: ${response.status}`,
    );
  }
}